package services

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

// These aliases are shared by the enqueue, skip and delivery queries. Only the
// pharmacy currently receiving the sequential alert can receive its email.
const pharmacistEmailActive = `ma.status='pending' AND ma.expires_at>now()
 AND n.sent_at+interval '30 seconds'>now()
 AND u.user_role='pharmacist' AND u.managed_pharmacy_id=n.pharmacy_id
 AND NOT EXISTS(SELECT 1 FROM pharmacist_responses pr WHERE pr.alert_id=ma.id AND pr.pharmacy_id=n.pharmacy_id)
 AND NOT EXISTS(SELECT 1 FROM pharmacy_alert_notifications newer WHERE newer.alert_id=n.alert_id AND newer.id>n.id)`

func (s *AlertService) RunPharmacistAlertEmailWorker(ctx context.Context) {
	if os.Getenv("EMAIL_REMINDERS_ENABLED") != "true" || os.Getenv("PHARMACIST_ALERT_EMAILS_ENABLED") == "false" {
		log.Print("Pharmacist alert emails disabled")
		return
	}
	sender, err := newSMTPReminderSender()
	if err != nil {
		log.Printf("Pharmacist alert emails not enabled: %v", err)
		return
	}
	log.Print("Pharmacist alert emails enabled")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		workCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		if err = s.ProcessPharmacistAlertEmails(workCtx, sender); err != nil {
			log.Print("Pharmacist email processing failed; pending jobs will retry")
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *AlertService) ProcessPharmacistAlertEmails(ctx context.Context, sender ReminderMailSender) error {
	_, err := s.db.Exec(ctx, `INSERT INTO pharmacist_alert_emails(notification_id,user_id)
 SELECT n.id,u.id FROM pharmacy_alert_notifications n JOIN medication_alerts ma ON ma.id=n.alert_id
 JOIN users u ON u.managed_pharmacy_id=n.pharmacy_id WHERE `+pharmacistEmailActive+`
 ON CONFLICT(notification_id,user_id) DO NOTHING`)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `UPDATE pharmacist_alert_emails e SET status='skipped' WHERE e.status='pending' AND NOT EXISTS(
 SELECT 1 FROM pharmacy_alert_notifications n JOIN medication_alerts ma ON ma.id=n.alert_id JOIN users u ON u.id=e.user_id
 WHERE n.id=e.notification_id AND `+pharmacistEmailActive+`)`)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var job int64
	var notification, user, alert, pharmacyID, quantity int32
	var to, pharmacy, medication, dosage, form, ingredient, patient, phone string
	var created, expires time.Time
	err = tx.QueryRow(ctx, `SELECT e.id,n.id,u.id,ma.id,n.pharmacy_id,u.email,p.name,COALESCE(m.speciality,''),
 COALESCE(m.dosage,''),COALESCE(m.form,''),COALESCE(m.active_substance,''),COALESCE(ma.requested_quantity,1),
 patient.name,COALESCE(patient.phone,''),ma.created_at,LEAST(ma.expires_at,n.sent_at+interval '30 seconds')
 FROM pharmacist_alert_emails e JOIN pharmacy_alert_notifications n ON n.id=e.notification_id
 JOIN medication_alerts ma ON ma.id=n.alert_id JOIN users u ON u.id=e.user_id
 JOIN users patient ON patient.id=ma.customer_id JOIN medications m ON m.id=ma.medication_id JOIN pharmacies p ON p.id=n.pharmacy_id
 WHERE e.status='pending' AND e.next_attempt_at<=now() AND `+pharmacistEmailActive+`
 ORDER BY e.id LIMIT 1 FOR UPDATE OF e SKIP LOCKED`).Scan(&job, &notification, &user, &alert, &pharmacyID, &to, &pharmacy, &medication, &dosage, &form, &ingredient, &quantity, &patient, &phone, &created, &expires)
	if err == pgx.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err = plainEmailAddress(to); err != nil {
		if _, err = tx.Exec(ctx, `UPDATE pharmacist_alert_emails SET status='skipped' WHERE id=$1`, job); err != nil {
			return err
		}
		log.Printf("Pharmacist email job %d skipped: account needs a valid email address", job)
		return tx.Commit(ctx)
	}
	body := pharmacistAlertEmailBody(alert, pharmacyID, pharmacy, medication, dosage, form, ingredient, quantity, patient, phone, created, expires)
	subject := fmt.Sprintf("Qayen - new medication alert #%d", alert)
	if err = sender.Send(ctx, to, fmt.Sprintf("pharmacy-alert-%d-user-%d@qayen", notification, user), subject, body); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if _, updateErr := tx.Exec(ctx, `UPDATE pharmacist_alert_emails SET attempts=attempts+1,next_attempt_at=now()+interval '5 seconds' WHERE id=$1`, job); updateErr != nil {
			return updateErr
		}
		if updateErr := tx.Commit(ctx); updateErr != nil {
			return updateErr
		}
		log.Printf("Pharmacist email job %d failed: %v; retrying while alert remains assigned", job, err)
		return nil
	}
	if _, err = tx.Exec(ctx, `UPDATE pharmacist_alert_emails SET status='sent',attempts=attempts+1,sent_at=now() WHERE id=$1`, job); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func pharmacistAlertEmailBody(alert, pharmacyID int32, pharmacy, medication, dosage, form, ingredient string, quantity int32, patient, phone string, created, expires time.Time) string {
	if zone, err := time.LoadLocation("Africa/Casablanca"); err == nil {
		created = created.In(zone)
		expires = expires.In(zone)
	}
	if strings.TrimSpace(phone) == "" {
		phone = "Not provided"
	}
	dashboard := "Open Qayen and sign in to your pharmacist dashboard to respond."
	if base, err := url.Parse(strings.TrimSpace(os.Getenv("QAYEN_PUBLIC_URL"))); err == nil && (base.Scheme == "https" || base.Scheme == "http") && base.Host != "" && base.User == nil {
		base.Path = "/pharmacy"
		base.RawPath = ""
		base.RawQuery = ""
		base.Fragment = ""
		dashboard += "\n" + base.String()
	}
	return fmt.Sprintf("Hello %s,\n\nA patient nearby has requested medication from your pharmacy.\n\nAlert: #%d\nPharmacy reference: %d\nMedication: %s\nDosage: %s\nForm: %s\nActive ingredient: %s\nRequested quantity: %d\nPatient: %s\nPatient phone: %s\nRequested at (Morocco): %s\nResponse window ends (Morocco): %s\n\n%s\n\nPlease check the live dashboard: availability may already have changed. This email does not reserve or change your stock.\n\nQayen\n", pharmacy, alert, pharmacyID, medication, dosage, form, ingredient, quantity, patient, phone, created.Format("02 Jan 2006 15:04:05 -0700"), expires.Format("02 Jan 2006 15:04:05 -0700"), dashboard)
}
