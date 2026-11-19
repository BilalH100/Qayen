package services

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
	"strings"
	"time"
	_ "time/tzdata"
)

// Durable unique jobs survive restarts and keep the two email reminders separate
// from in-app reminder flags. No SMTP configuration means no outgoing mail.
func (s *AlertService) RunPickupEmailWorker(ctx context.Context) {
	if os.Getenv("EMAIL_REMINDERS_ENABLED") != "true" {
		log.Print("Pickup emails disabled; set EMAIL_REMINDERS_ENABLED=true after configuring SMTP")
		return
	}
	sender, err := newSMTPReminderSender()
	if err != nil {
		log.Printf("Pickup emails not enabled: %v", err)
		return
	}
	log.Print("Pickup email reminders enabled")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		workCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		if err = s.ProcessPickupEmails(workCtx, sender); err != nil {
			log.Print("Pickup email processing failed; pending jobs will retry")
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (s *AlertService) ProcessPickupEmails(ctx context.Context, sender ReminderMailSender) error {
	_, err := s.db.Exec(ctx, `INSERT INTO pickup_email_reminders(reservation_id,stage,scheduled_at)
 SELECT r.id,stage,r.created_at+(r.expires_at-r.created_at)*(CASE stage WHEN 1 THEN 0.6 ELSE 0.8 END)
 FROM pickup_reservations r CROSS JOIN generate_series(1,2) stage
 WHERE r.status='awaiting' AND r.expires_at>now() ON CONFLICT(reservation_id,stage) DO NOTHING`)
	if err != nil {
		return err
	}
	// Suppress expired/collected jobs and obsolete first reminders after downtime.
	_, err = s.db.Exec(ctx, `UPDATE pickup_email_reminders e SET status='skipped' FROM pickup_reservations r WHERE e.reservation_id=r.id AND e.status='pending' AND (r.status<>'awaiting' OR r.expires_at<=now() OR (e.stage=1 AND now()>=r.created_at+(r.expires_at-r.created_at)*0.8))`)
	if err != nil {
		return err
	}
	// Process one message per pass so slow SMTP never blocks stock expiry work.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var job int64
	var reservation, stage, attempts int
	var to, name, pharmacy, address string
	var expires time.Time
	err = tx.QueryRow(ctx, `SELECT e.id,r.id,e.stage,e.attempts,u.email,u.name,p.name,COALESCE(p.address,''),r.expires_at
 FROM pickup_email_reminders e JOIN pickup_reservations r ON r.id=e.reservation_id
 JOIN users u ON u.id=r.customer_id JOIN pharmacies p ON p.id=r.pharmacy_id
 WHERE e.status='pending' AND e.scheduled_at<=now() AND e.next_attempt_at<=now() AND r.status='awaiting' AND r.expires_at>now()
 ORDER BY e.scheduled_at,e.id LIMIT 1 FOR UPDATE OF e SKIP LOCKED`).Scan(&job, &reservation, &stage, &attempts, &to, &name, &pharmacy, &address, &expires)
	if err == pgx.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err = plainEmailAddress(to); err != nil {
		_, err = tx.Exec(ctx, `UPDATE pickup_email_reminders SET status='skipped' WHERE id=$1`, job)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	if err = sender.Send(ctx, to, fmt.Sprintf("pickup-%d-reminder-%d@qayen", reservation, stage), "Qayen - medication pickup reminder", pickupEmailBody(name, pharmacy, address, expires, stage)); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		_, updateErr := tx.Exec(ctx, `UPDATE pickup_email_reminders SET attempts=attempts+1,next_attempt_at=now()+interval '15 seconds' WHERE id=$1`, job)
		if updateErr != nil {
			return updateErr
		}
		if updateErr = tx.Commit(ctx); updateErr != nil {
			return updateErr
		}
		log.Printf("Pickup email job %d failed: %v; retrying while reservation remains active", job, err)
		return nil
	}
	_, err = tx.Exec(ctx, `UPDATE pickup_email_reminders SET status='sent',attempts=attempts+1,sent_at=now() WHERE id=$1`, job)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func pickupEmailBody(name, pharmacy, address string, expires time.Time, stage int) string {
	location, err := time.LoadLocation("Africa/Casablanca")
	if err == nil {
		expires = expires.In(location)
	}
	intro := "Please remember to collect your reserved medication."
	if stage == 2 {
		intro = "Your medication reservation is about to expire. Please collect it before the deadline."
	}
	return fmt.Sprintf("Hello %s,\n\n%s\n\nPharmacy: %s\nAddress: %s\nPickup deadline (Morocco): %s\n\nIf it is not collected before this deadline, the reservation will be canceled and the reserved quantity returned to stock.\nThe pharmacist confirms collection in Qayen.\n\nQayen\n", strings.TrimSpace(name), intro, pharmacy, address, expires.Format("02 Jan 2006 15:04:05 -0700"))
}
