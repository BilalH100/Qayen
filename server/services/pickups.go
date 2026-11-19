package services

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
	"strconv"
	"time"
)

type PickupReservation struct {
	ID             int32     `json:"id"`
	AlertID        int32     `json:"alert_id"`
	PharmacyID     int32     `json:"pharmacy_id"`
	Quantity       int32     `json:"quantity"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	ReminderStage  int       `json:"reminder_stage"`
	MedicationName string    `json:"medication_name"`
	PharmacyName   string    `json:"pharmacy_name"`
	CustomerName   string    `json:"customer_name"`
}

func pickupDuration() time.Duration {
	if n, err := strconv.Atoi(os.Getenv("QAYEN_PICKUP_HOLD_SECONDS")); err == nil && n >= 5 && n <= 604800 {
		return time.Duration(n) * time.Second
	}
	return 5 * time.Minute
}

// Reservation and inventory changes commit together, guarded by the alert row lock.
func (s *AlertService) reservePickup(ctx context.Context, req ConfirmPickupRequest) (*ConfirmPickupResult, error) {
	const op = "ReservePickup"
	if req.AlertID <= 0 || req.PharmacyID <= 0 || req.Quantity <= 0 {
		return nil, NewValidationError(op, "Invalid reservation", "")
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var customer, med int32
	var status string
	var expires time.Time
	err = tx.QueryRow(ctx, `SELECT customer_id,medication_id,status,expires_at FROM medication_alerts WHERE id=$1 FOR UPDATE`, req.AlertID).Scan(&customer, &med, &status, &expires)
	if err != nil {
		return nil, NewNotFoundError(op, "Alert not found")
	}
	if req.CustomerID != customer {
		return nil, NewBusinessLogicError(op, "This alert belongs to another patient", "")
	}
	// Retried confirmation returns the original hold without decrementing again.
	var original ConfirmPickupResult
	var pharmacy, qty int32
	err = tx.QueryRow(ctx, `SELECT id,pharmacy_id,quantity,status,expires_at FROM pickup_reservations WHERE alert_id=$1`, req.AlertID).Scan(&original.ReservationID, &pharmacy, &qty, &original.Status, &original.ExpiresAt)
	if err == nil {
		if pharmacy != req.PharmacyID || qty != req.Quantity {
			return nil, NewBusinessLogicError(op, "This alert already has a reservation", "")
		}
		original.AlertID = req.AlertID
		original.PharmacyID = pharmacy
		original.QuantityTaken = qty
		return &original, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}
	if status != "pending" || !expires.After(time.Now()) {
		return nil, NewBusinessLogicError(op, "Alert is no longer active", "")
	}
	var available bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pharmacist_responses WHERE alert_id=$1 AND pharmacy_id=$2 AND response_type='available' AND expires_at>now())`, req.AlertID, req.PharmacyID).Scan(&available)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, NewBusinessLogicError(op, "This pharmacy has not confirmed availability", "")
	}
	var stockID int32
	err = tx.QueryRow(ctx, `SELECT id FROM `+s.alertStockTable()+` WHERE pharmacy_id=$1 AND medication_id=$2`, req.PharmacyID, med).Scan(&stockID)
	if err != nil {
		return nil, NewBusinessLogicError(op, "Stock is unavailable", "")
	}
	table := "stock"
	id := stockID
	if id < 0 {
		table = "catalog_test_stock"
		id = -id
	}
	var remaining int32
	err = tx.QueryRow(ctx, `UPDATE `+table+` SET quantity=quantity-$2,updated_at=now() WHERE id=$1 AND quantity>=$2 RETURNING quantity`, id, req.Quantity).Scan(&remaining)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, NewBusinessLogicError(op, "Not enough stock", "")
		}
		return nil, err
	}
	result := &ConfirmPickupResult{AlertID: req.AlertID, PharmacyID: req.PharmacyID, QuantityTaken: req.Quantity, RemainingStock: remaining, Status: "awaiting"}
	err = tx.QueryRow(ctx, `INSERT INTO pickup_reservations(alert_id,customer_id,pharmacy_id,medication_id,stock_id,quantity,expires_at) VALUES($1,$2,$3,$4,$5,$6,now()+$7*interval '1 second') RETURNING id,expires_at`, req.AlertID, customer, req.PharmacyID, med, stockID, req.Quantity, pickupDuration().Seconds()).Scan(&result.ReservationID, &result.ExpiresAt)
	if err != nil {
		return nil, err
	}
	// Close the search; the independent reservation tracks collection and expiry.
	_, err = tx.Exec(ctx, `UPDATE medication_alerts SET status='completed' WHERE id=$1`, req.AlertID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
func (s *AlertService) ListPickups(ctx context.Context, userID, pharmacyID int32) ([]PickupReservation, error) {
	filter := "r.customer_id=$1"
	id := userID
	if pharmacyID > 0 {
		var permitted bool
		if err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND managed_pharmacy_id=$2)`, userID, pharmacyID).Scan(&permitted); err != nil {
			return nil, err
		}
		if !permitted {
			return nil, NewBusinessLogicError("ListPickups", "You do not manage this pharmacy", "")
		}
		filter = "r.pharmacy_id=$1"
		id = pharmacyID
	}
	rows, err := s.db.Query(ctx, `SELECT r.id,r.alert_id,r.pharmacy_id,r.quantity,r.status,r.created_at,r.expires_at,r.reminder_stage,COALESCE(m.speciality,''),p.name,u.name FROM pickup_reservations r JOIN medications m ON m.id=r.medication_id JOIN pharmacies p ON p.id=r.pharmacy_id JOIN users u ON u.id=r.customer_id WHERE `+filter+` ORDER BY r.created_at DESC LIMIT 200`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PickupReservation{}
	for rows.Next() {
		var r PickupReservation
		if err = rows.Scan(&r.ID, &r.AlertID, &r.PharmacyID, &r.Quantity, &r.Status, &r.CreatedAt, &r.ExpiresAt, &r.ReminderStage, &r.MedicationName, &r.PharmacyName, &r.CustomerName); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, rows.Err()
}

// Called by the pharmacist at handover. The deadline is enforced by the DB.
func (s *AlertService) CollectPickup(ctx context.Context, userID, id int32) error {
	tag, err := s.db.Exec(ctx, `UPDATE pickup_reservations r SET status='picked_up',resolved_at=now() WHERE r.id=$1 AND r.status='awaiting' AND r.expires_at>now() AND EXISTS(SELECT 1 FROM users u WHERE u.id=$2 AND u.managed_pharmacy_id=r.pharmacy_id)`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return NewBusinessLogicError("CollectPickup", "Request expired, already resolved, or not managed by you", "")
	}
	return nil
}

// Persistent deadlines survive browser closure and backend restarts.
func (s *AlertService) ProcessPickups(ctx context.Context) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `SELECT id,alert_id,stock_id,quantity,pharmacy_id,medication_id FROM pickup_reservations WHERE status='awaiting' AND expires_at<=now() FOR UPDATE SKIP LOCKED`)
	if err != nil {
		return err
	}
	type due struct{ id, alert, stock, quantity, pharmacy, medication int32 }
	items := []due{}
	for rows.Next() {
		var d due
		if err = rows.Scan(&d.id, &d.alert, &d.stock, &d.quantity, &d.pharmacy, &d.medication); err != nil {
			rows.Close()
			return err
		}
		items = append(items, d)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, d := range items {
		table := "stock"
		if d.stock < 0 {
			table = "catalog_test_stock"
		}
		_, e := tx.Exec(ctx, `INSERT INTO `+table+` (pharmacy_id,medication_id,quantity) VALUES($1,$2,$3) ON CONFLICT(pharmacy_id,medication_id) DO UPDATE SET quantity=`+table+`.quantity+EXCLUDED.quantity,updated_at=now()`, d.pharmacy, d.medication, d.quantity)
		if e != nil {
			return e
		}
		if _, err = tx.Exec(ctx, `UPDATE pickup_reservations SET status='canceled',resolved_at=now() WHERE id=$1`, d.id); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `UPDATE medication_alerts SET status='cancelled' WHERE id=$1`, d.alert); err != nil {
			return err
		}
	}
	// In test mode these thresholds are exactly 180s and 240s after creation.
	_, err = tx.Exec(ctx, `UPDATE pickup_reservations SET reminder_stage=CASE WHEN now()>=created_at+(expires_at-created_at)*0.8 THEN 2 ELSE 1 END WHERE status='awaiting' AND now()>=created_at+(expires_at-created_at)*0.6 AND reminder_stage<CASE WHEN now()>=created_at+(expires_at-created_at)*0.8 THEN 2 ELSE 1 END`)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *AlertService) RunPickupWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if err := s.ProcessPickups(ctx); err != nil {
			log.Printf("Pickup processing failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
