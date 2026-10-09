package services

import (
	"context"
	"fmt"
	sqlc "kayena/server/database/generated"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AlertService struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
	sseHub  *SSEHub
}

func NewAlertService(db *pgxpool.Pool, sseHub *SSEHub) *AlertService {
	return &AlertService{
		db:      db,
		queries: sqlc.New(db),
		sseHub:  sseHub,
	}
}

// AlertRequest represents a customer's medication search request
type AlertRequest struct {
	CustomerID      int32   `json:"customer_id"`
	MedicationID    int32   `json:"medication_id"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	SearchRadius    float64 `json:"search_radius_km"`
	MaxResponseTime int32   `json:"max_response_time_minutes"`
	// RequestedQuantity is how many units the patient needs. Only pharmacies
	// with at least this much stock are alerted. Defaults to 1.
	RequestedQuantity int32 `json:"requested_quantity"`
}

// How long one pharmacy gets to answer before we move on to the next nearest.
const pharmacyResponseTimeout = 30 * time.Second

// PharmacistResponse represents a pharmacist's response to an alert
type PharmacistResponseRequest struct {
	AlertID                int32  `json:"alert_id"`
	PharmacyID             int32  `json:"pharmacy_id"`
	ResponseType           string `json:"response_type"` // available, unavailable, substitute
	SubstituteMedicationID *int32 `json:"substitute_medication_id,omitempty"`
	SubstituteBrand        string `json:"substitute_brand,omitempty"`
	SubstituteNotes        string `json:"substitute_notes,omitempty"`
	ResponseTimeSeconds    int32  `json:"response_time_seconds"`
}

// AlertResult represents the final result returned to customer
type AlertResult struct {
	AlertID           int32                  `json:"alert_id"`
	MedicationID      int32                  `json:"medication_id"`
	Status            string                 `json:"status"`
	RequestedQuantity int32                  `json:"requested_quantity"`
	Pharmacies        []PharmacyAvailability `json:"pharmacies"`
	CreatedAt         time.Time              `json:"created_at"`
	ExpiresAt         time.Time              `json:"expires_at"`
}

type PharmacyAvailability struct {
	PharmacyID        int32   `json:"pharmacy_id"`
	PharmacyName      string  `json:"pharmacy_name"`
	PharmacyAddress   string  `json:"pharmacy_address"`
	PharmacyPhone     string  `json:"pharmacy_phone"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	DistanceKM        float64 `json:"distance_km"`
	ResponseType      string  `json:"response_type"`
	SubstituteBrand   string  `json:"substitute_brand,omitempty"`
	SubstituteNotes   string  `json:"substitute_notes,omitempty"`
	ResponseTime      int32   `json:"response_time_seconds"`
	AvailableQuantity int32   `json:"available_quantity"`
}

// ConfirmPickupRequest represents a patient confirming they're taking a
// given quantity of the medication from a pharmacy that responded "available".
type ConfirmPickupRequest struct {
	AlertID    int32 `json:"alert_id"`
	PharmacyID int32 `json:"pharmacy_id"`
	Quantity   int32 `json:"quantity"`
}

// ConfirmPickupResult is returned after a successful confirmation.
type ConfirmPickupResult struct {
	AlertID        int32 `json:"alert_id"`
	PharmacyID     int32 `json:"pharmacy_id"`
	QuantityTaken  int32 `json:"quantity_taken"`
	RemainingStock int32 `json:"remaining_stock"`
}

// Custom error types for better error handling and tracing
type AlertError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Operation string `json:"operation"`
	Err       error  `json:"-"`
}

func (e *AlertError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %s (details: %s)", e.Code, e.Operation, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Operation, e.Message)
}

func (e *AlertError) Unwrap() error {
	return e.Err
}

// Error constructors
func NewValidationError(operation, message string, details ...string) *AlertError {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}
	return &AlertError{
		Code:      "VALIDATION_ERROR",
		Operation: operation,
		Message:   message,
		Details:   detail,
	}
}

func NewDatabaseError(operation, message string, err error) *AlertError {
	return &AlertError{
		Code:      "DATABASE_ERROR",
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

func NewBusinessLogicError(operation, message string, details ...string) *AlertError {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}
	return &AlertError{
		Code:      "BUSINESS_LOGIC_ERROR",
		Operation: operation,
		Message:   message,
		Details:   detail,
	}
}

func NewNotFoundError(operation, message string) *AlertError {
	return &AlertError{
		Code:      "NOT_FOUND",
		Operation: operation,
		Message:   message,
	}
}

func NewInternalError(operation, message string, err error) *AlertError {
	return &AlertError{
		Code:      "INTERNAL_ERROR",
		Operation: operation,
		Message:   message,
		Err:       err,
	}
}

// Helper function to log errors with structured format
func (s *AlertService) logError(ctx context.Context, err *AlertError, additionalFields ...interface{}) {
	logMsg := fmt.Sprintf("Alert Service Error - Code: %s, Operation: %s, Message: %s",
		err.Code, err.Operation, err.Message)

	if err.Details != "" {
		logMsg += fmt.Sprintf(", Details: %s", err.Details)
	}

	if err.Err != nil {
		logMsg += fmt.Sprintf(", Underlying Error: %s", err.Err.Error())
	}

	if len(additionalFields) > 0 {
		logMsg += fmt.Sprintf(", Additional: %+v", additionalFields)
	}

	log.Print(logMsg)
}

// CreateMedicationAlert creates a new medication alert and notifies nearby pharmacies
func (s *AlertService) CreateMedicationAlert(ctx context.Context, req AlertRequest) (*AlertResult, error) {
	const operation = "CreateMedicationAlert"

	// Enhanced validation
	if req.CustomerID <= 0 {
		err := NewValidationError(operation, "Invalid customer ID", fmt.Sprintf("CustomerID: %d", req.CustomerID))
		s.logError(ctx, err)
		return nil, err
	}

	if req.MedicationID <= 0 {
		err := NewValidationError(operation, "Invalid medication ID", fmt.Sprintf("MedicationID: %d", req.MedicationID))
		s.logError(ctx, err)
		return nil, err
	}

	if req.Latitude < -90 || req.Latitude > 90 {
		err := NewValidationError(operation, "Invalid latitude", fmt.Sprintf("Latitude: %f", req.Latitude))
		s.logError(ctx, err)
		return nil, err
	}

	if req.Longitude < -180 || req.Longitude > 180 {
		err := NewValidationError(operation, "Invalid longitude", fmt.Sprintf("Longitude: %f", req.Longitude))
		s.logError(ctx, err)
		return nil, err
	}

	if req.SearchRadius <= 0 || req.SearchRadius > 100 {
		err := NewValidationError(operation, "Invalid search radius", fmt.Sprintf("SearchRadius: %f (must be between 0.1 and 100 km)", req.SearchRadius))
		s.logError(ctx, err)
		return nil, err
	}

	if req.RequestedQuantity <= 0 {
		req.RequestedQuantity = 1
	}

	log.Printf("Creating medication alert - CustomerID: %d, MedicationID: %d, Location: (%.6f, %.6f), Radius: %.1fkm",
		req.CustomerID, req.MedicationID, req.Latitude, req.Longitude, req.SearchRadius)

	// Start transaction with better error context
	tx, err := s.db.Begin(ctx)
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to begin transaction", err)
		s.logError(ctx, dbErr)
		return nil, dbErr
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			rollbackErr := NewDatabaseError(operation, "Failed to rollback transaction", err)
			s.logError(ctx, rollbackErr)
		}
	}()

	qtx := s.queries.WithTx(tx)

	// Create the alert with detailed error context
	alert, err := qtx.CreateMedicationAlert(ctx, sqlc.CreateMedicationAlertParams{
		CustomerID:             pgtype.Int4{Int32: req.CustomerID, Valid: true},
		MedicationID:           pgtype.Int4{Int32: req.MedicationID, Valid: true},
		CustomerLatitude:       req.Latitude,
		CustomerLongitude:      req.Longitude,
		SearchRadiusKm:         pgtype.Float8{Float64: req.SearchRadius, Valid: true},
		MaxResponseTimeMinutes: pgtype.Int4{Int32: req.MaxResponseTime, Valid: true},
	})
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to create medication alert", err)
		s.logError(ctx, dbErr, "CustomerID", req.CustomerID, "MedicationID", req.MedicationID)
		return nil, dbErr
	}

	log.Printf("Created medication alert with ID: %d", alert.ID)

	// Remember how many units the patient needs.
	if _, err = tx.Exec(ctx, `UPDATE medication_alerts SET requested_quantity = $2 WHERE id = $1`, alert.ID, req.RequestedQuantity); err != nil {
		dbErr := NewDatabaseError(operation, "Failed to save requested quantity", err)
		s.logError(ctx, dbErr, "AlertID", alert.ID)
		return nil, dbErr
	}

	// Alert pharmacies one-by-one, starting with the nearest pharmacy that
	// actually carries the medication. Quantity is checked when the pharmacy
	// responds: a pharmacy with stock below the requested quantity is treated
	// as unavailable and the alert advances to the next nearest pharmacy.
	firstPharmacyID, found, err := s.notifyNextPharmacy(ctx, tx, alert.ID)
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to find a nearby pharmacy", err)
		s.logError(ctx, dbErr, "AlertID", alert.ID, "SearchRadius", req.SearchRadius)
		return nil, dbErr
	}
	if !found {
		businessErr := NewBusinessLogicError(operation,
			"No nearby pharmacy has this medication in stock",
			fmt.Sprintf("Try a different medication or a larger search radius (current: %.0f km)", req.SearchRadius))
		s.logError(ctx, businessErr, "AlertID", alert.ID)
		return nil, businessErr
	}
	notifiedCount := 1

	// Update alert with notification count
	updatedAlert, err := qtx.UpdateMedicationAlertStatus(ctx, sqlc.UpdateMedicationAlertStatusParams{
		ID:                     alert.ID,
		Status:                 "pending",
		TotalResponsesReceived: pgtype.Int4{Int32: 0, Valid: true},
	})
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to update alert status", err)
		s.logError(ctx, dbErr, "AlertID", alert.ID, "NotifiedCount", notifiedCount)
		return nil, dbErr
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		dbErr := NewDatabaseError(operation, "Failed to commit transaction", err)
		s.logError(ctx, dbErr, "AlertID", alert.ID)
		return nil, dbErr
	}

	log.Printf("Successfully created and processed alert %d - Notified %d pharmacies", alert.ID, notifiedCount)

	// Dispatch only after commit. Pharmacies linked to a user account get the
	// normal dashboard/SSE flow. Pharmacies without an account are evaluated
	// automatically from their stock and can cause an immediate advance.
	s.dispatchPharmacyAlert(ctx, firstPharmacyID, alert.ID, req.MedicationID)

	result := &AlertResult{
		AlertID:           updatedAlert.ID,
		MedicationID:      updatedAlert.MedicationID.Int32,
		Status:            updatedAlert.Status,
		RequestedQuantity: req.RequestedQuantity,
		CreatedAt:         updatedAlert.CreatedAt.Time,
		ExpiresAt:         updatedAlert.ExpiresAt.Time,
		Pharmacies:        []PharmacyAvailability{}, // Will be populated as responses come in
	}

	return result, nil
}

// SubmitPharmacistResponse handles a pharmacist's response to an alert
func (s *AlertService) SubmitPharmacistResponse(ctx context.Context, req PharmacistResponseRequest) error {
	const operation = "SubmitPharmacistResponse"

	// Enhanced validation
	if req.AlertID <= 0 {
		err := NewValidationError(operation, "Invalid alert ID", fmt.Sprintf("AlertID: %d", req.AlertID))
		s.logError(ctx, err)
		return err
	}

	if req.PharmacyID <= 0 {
		err := NewValidationError(operation, "Invalid pharmacy ID", fmt.Sprintf("PharmacyID: %d", req.PharmacyID))
		s.logError(ctx, err)
		return err
	}

	validResponseTypes := map[string]bool{"available": true, "unavailable": true, "substitute": true}
	if !validResponseTypes[req.ResponseType] {
		err := NewValidationError(operation, "Invalid response type",
			fmt.Sprintf("ResponseType: %s (valid: available, unavailable, substitute)", req.ResponseType))
		s.logError(ctx, err)
		return err
	}

	log.Printf("Processing pharmacist response - AlertID: %d, PharmacyID: %d, ResponseType: %s",
		req.AlertID, req.PharmacyID, req.ResponseType)

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to begin transaction", err)
		s.logError(ctx, dbErr)
		return dbErr
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			rollbackErr := NewDatabaseError(operation, "Failed to rollback transaction", err)
			s.logError(ctx, rollbackErr)
		}
	}()

	qtx := s.queries.WithTx(tx)

	// Verify alert is still active with detailed error context
	alert, err := qtx.GetMedicationAlert(ctx, req.AlertID)
	if err != nil {
		if err == pgx.ErrNoRows {
			notFoundErr := NewNotFoundError(operation, "Alert not found")
			s.logError(ctx, notFoundErr, "AlertID", req.AlertID)
			return notFoundErr
		}
		dbErr := NewDatabaseError(operation, "Failed to fetch alert", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID)
		return dbErr
	}

	if alert.Status != "pending" {
		businessErr := NewBusinessLogicError(operation, "Alert is no longer pending",
			fmt.Sprintf("Current status: %s", alert.Status))
		s.logError(ctx, businessErr, "AlertID", req.AlertID, "Status", alert.Status)
		return businessErr
	}

	if alert.ExpiresAt.Time.Before(time.Now()) {
		businessErr := NewBusinessLogicError(operation, "Alert has expired",
			fmt.Sprintf("Expired at: %s", alert.ExpiresAt.Time.Format(time.RFC3339)))
		s.logError(ctx, businessErr, "AlertID", req.AlertID, "ExpiresAt", alert.ExpiresAt.Time)
		return businessErr
	}

	// Only the pharmacy currently being tried may answer. This prevents a
	// previously-notified pharmacy from answering after the alert has already
	// moved on to another nearby pharmacy.
	var isCurrentPharmacy bool
	if err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pharmacy_alert_notifications n
			WHERE n.alert_id = $1
			  AND n.pharmacy_id = $2
			  AND n.id = (
				SELECT id
				FROM pharmacy_alert_notifications
				WHERE alert_id = $1
				ORDER BY id DESC
				LIMIT 1
			  )
		)`, req.AlertID, req.PharmacyID).Scan(&isCurrentPharmacy); err != nil {
		dbErr := NewDatabaseError(operation, "Failed to validate pharmacy alert assignment", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID, "PharmacyID", req.PharmacyID)
		return dbErr
	}
	if !isCurrentPharmacy {
		businessErr := NewBusinessLogicError(operation, "This pharmacy is not the current pharmacy being contacted")
		s.logError(ctx, businessErr, "AlertID", req.AlertID, "PharmacyID", req.PharmacyID)
		return businessErr
	}

	// Enforce the requested quantity from the real stock table. A pharmacy
	// may carry the medication but still be unavailable when it does not have
	// enough units for this specific request.
	var requestedQuantity, availableQuantity int32
	if err = tx.QueryRow(ctx, `
		SELECT COALESCE(requested_quantity, 1)
		FROM medication_alerts
		WHERE id = $1`, req.AlertID).Scan(&requestedQuantity); err != nil {
		dbErr := NewDatabaseError(operation, "Failed to fetch requested quantity", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID)
		return dbErr
	}
	if err = tx.QueryRow(ctx, `
		SELECT COALESCE((
			SELECT s.quantity
			FROM stock s
			WHERE s.pharmacy_id = $1 AND s.medication_id = $2
		), 0)`, req.PharmacyID, alert.MedicationID.Int32).Scan(&availableQuantity); err != nil {
		dbErr := NewDatabaseError(operation, "Failed to fetch pharmacy stock", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID, "PharmacyID", req.PharmacyID)
		return dbErr
	}
	if req.ResponseType == "available" && availableQuantity < requestedQuantity {
		log.Printf("Pharmacy %d cannot fulfill alert %d: stock=%d requested=%d; converting response to unavailable",
			req.PharmacyID, req.AlertID, availableQuantity, requestedQuantity)
		req.ResponseType = "unavailable"
	}

	// Create pharmacist response with enhanced validation
	var substituteMedID pgtype.Int4
	if req.SubstituteMedicationID != nil {
		if *req.SubstituteMedicationID <= 0 {
			err := NewValidationError(operation, "Invalid substitute medication ID",
				fmt.Sprintf("SubstituteMedicationID: %d", *req.SubstituteMedicationID))
			s.logError(ctx, err, "AlertID", req.AlertID, "PharmacyID", req.PharmacyID)
			return err
		}
		substituteMedID = pgtype.Int4{Int32: *req.SubstituteMedicationID, Valid: true}
	}

	// Validate substitute response has required fields
	if req.ResponseType == "substitute" && req.SubstituteBrand == "" {
		err := NewValidationError(operation, "Substitute brand is required for substitute responses", "")
		s.logError(ctx, err, "AlertID", req.AlertID, "PharmacyID", req.PharmacyID)
		return err
	}

	_, err = qtx.CreatePharmacistResponse(ctx, sqlc.CreatePharmacistResponseParams{
		AlertID:                pgtype.Int4{Int32: req.AlertID, Valid: true},
		PharmacyID:             pgtype.Int4{Int32: req.PharmacyID, Valid: true},
		ResponseType:           req.ResponseType,
		SubstituteMedicationID: substituteMedID,
		SubstituteBrand:        req.SubstituteBrand,
		SubstituteNotes:        req.SubstituteNotes,
		ResponseTimeSeconds:    pgtype.Int4{Int32: req.ResponseTimeSeconds, Valid: true},
	})
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to create pharmacist response", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID, "PharmacyID", req.PharmacyID, "ResponseType", req.ResponseType)
		return dbErr
	}

	log.Printf("Created pharmacist response - AlertID: %d, PharmacyID: %d, ResponseType: %s",
		req.AlertID, req.PharmacyID, req.ResponseType)

	// Update analytics with error handling
	availabilityRate := 0.0
	if req.ResponseType == "available" || req.ResponseType == "substitute" {
		availabilityRate = 1.0
	}

	_, analyticsErr := qtx.UpdatePharmacyAnalytics(ctx, sqlc.UpdatePharmacyAnalyticsParams{
		PharmacyID:             pgtype.Int4{Int32: req.PharmacyID, Valid: true},
		Date:                   pgtype.Date{Time: time.Now(), Valid: true},
		TotalAlertsReceived:    pgtype.Int4{Int32: 1, Valid: true},
		TotalResponsesSent:     pgtype.Int4{Int32: 1, Valid: true},
		AvgResponseTimeSeconds: pgtype.Float8{Float64: float64(req.ResponseTimeSeconds), Valid: true},
		AvailabilityRate:       pgtype.Float8{Float64: availabilityRate, Valid: true},
	})
	if analyticsErr != nil {
		// Log analytics error but don't fail the entire operation
		internalErr := NewInternalError(operation, "Failed to update analytics (non-critical)", analyticsErr)
		s.logError(ctx, internalErr, "PharmacyID", req.PharmacyID, "AlertID", req.AlertID)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		dbErr := NewDatabaseError(operation, "Failed to commit transaction", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID, "PharmacyID", req.PharmacyID)
		return dbErr
	}

	log.Printf("Successfully processed pharmacist response - AlertID: %d, PharmacyID: %d", req.AlertID, req.PharmacyID)

	// Send real-time update to customer (non-critical operation)
	customerNotifyErr := s.notifyCustomerOfResponse(ctx, alert.CustomerID.Int32, req.AlertID)
	if customerNotifyErr != nil {
		internalErr := NewInternalError(operation, "Failed to notify customer (non-critical)", customerNotifyErr)
		s.logError(ctx, internalErr, "CustomerID", alert.CustomerID.Int32, "AlertID", req.AlertID)
		// Don't return error as the main operation succeeded
	}

	// "Unavailable" -> immediately try the next nearest pharmacy.
	if req.ResponseType == "unavailable" {
		go func(alertID, pharmacyID int32) {
			bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			s.advanceAlert(bg, alertID, pharmacyID)
		}(req.AlertID, req.PharmacyID)
	}

	return nil
}

// ConfirmPickup is called when a patient confirms they're taking a given
// quantity of medication from a pharmacy that responded "available". It
// atomically decrements that pharmacy's stock and marks the alert
// completed, so other pharmacies stop treating it as still pending.
func (s *AlertService) ConfirmPickup(ctx context.Context, req ConfirmPickupRequest) (*ConfirmPickupResult, error) {
	const operation = "ConfirmPickup"

	if req.AlertID <= 0 {
		err := NewValidationError(operation, "Invalid alert ID", fmt.Sprintf("AlertID: %d", req.AlertID))
		s.logError(ctx, err)
		return nil, err
	}

	if req.PharmacyID <= 0 {
		err := NewValidationError(operation, "Invalid pharmacy ID", fmt.Sprintf("PharmacyID: %d", req.PharmacyID))
		s.logError(ctx, err)
		return nil, err
	}

	if req.Quantity <= 0 {
		err := NewValidationError(operation, "Quantity must be greater than zero", fmt.Sprintf("Quantity: %d", req.Quantity))
		s.logError(ctx, err)
		return nil, err
	}

	log.Printf("Processing pickup confirmation - AlertID: %d, PharmacyID: %d, Quantity: %d",
		req.AlertID, req.PharmacyID, req.Quantity)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to begin transaction", err)
		s.logError(ctx, dbErr)
		return nil, dbErr
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			rollbackErr := NewDatabaseError(operation, "Failed to rollback transaction", err)
			s.logError(ctx, rollbackErr)
		}
	}()

	qtx := s.queries.WithTx(tx)

	// Verify the alert exists and is still pending, and get its medication_id
	// (we never trust a medication_id from the client for this decrement).
	alert, err := qtx.GetMedicationAlert(ctx, req.AlertID)
	if err != nil {
		if err == pgx.ErrNoRows {
			notFoundErr := NewNotFoundError(operation, "Alert not found")
			s.logError(ctx, notFoundErr, "AlertID", req.AlertID)
			return nil, notFoundErr
		}
		dbErr := NewDatabaseError(operation, "Failed to fetch alert", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID)
		return nil, dbErr
	}

	if alert.Status != "pending" {
		businessErr := NewBusinessLogicError(operation, "Alert is no longer active",
			fmt.Sprintf("Current status: %s", alert.Status))
		s.logError(ctx, businessErr, "AlertID", req.AlertID, "Status", alert.Status)
		return nil, businessErr
	}

	// Atomic decrement: fails (no row) if there isn't enough stock.
	stock, err := qtx.DecrementStockQuantity(ctx, sqlc.DecrementStockQuantityParams{
		PharmacyID:   pgtype.Int4{Int32: req.PharmacyID, Valid: true},
		MedicationID: alert.MedicationID,
		Quantity:     req.Quantity,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			businessErr := NewBusinessLogicError(operation, "Not enough stock to fulfill this quantity", "")
			s.logError(ctx, businessErr, "AlertID", req.AlertID, "PharmacyID", req.PharmacyID, "Quantity", req.Quantity)
			return nil, businessErr
		}
		dbErr := NewDatabaseError(operation, "Failed to decrement stock", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID, "PharmacyID", req.PharmacyID)
		return nil, dbErr
	}

	// Mark the alert completed, guarded so a race with expiry/another
	// confirmation is caught instead of silently overwritten.
	if _, err := qtx.CompleteMedicationAlert(ctx, req.AlertID); err != nil {
		if err == pgx.ErrNoRows {
			businessErr := NewBusinessLogicError(operation, "Alert was already resolved", "")
			s.logError(ctx, businessErr, "AlertID", req.AlertID)
			return nil, businessErr
		}
		dbErr := NewDatabaseError(operation, "Failed to complete alert", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID)
		return nil, dbErr
	}

	if err = tx.Commit(ctx); err != nil {
		dbErr := NewDatabaseError(operation, "Failed to commit transaction", err)
		s.logError(ctx, dbErr, "AlertID", req.AlertID)
		return nil, dbErr
	}

	log.Printf("Pickup confirmed - AlertID: %d, PharmacyID: %d, Quantity: %d, RemainingStock: %d",
		req.AlertID, req.PharmacyID, req.Quantity, stock.Quantity)

	// Let the customer's other open tabs (if any) know the alert is resolved.
	notifyErr := s.notifyCustomerOfResponse(ctx, alert.CustomerID.Int32, req.AlertID)
	if notifyErr != nil {
		internalErr := NewInternalError(operation, "Failed to notify customer (non-critical)", notifyErr)
		s.logError(ctx, internalErr, "CustomerID", alert.CustomerID.Int32, "AlertID", req.AlertID)
	}

	return &ConfirmPickupResult{
		AlertID:        req.AlertID,
		PharmacyID:     req.PharmacyID,
		QuantityTaken:  req.Quantity,
		RemainingStock: stock.Quantity,
	}, nil
}

// GetAlertResults retrieves all current responses for an alert
func (s *AlertService) GetAlertResults(ctx context.Context, alertID int32, customerLat, customerLng float64) (*AlertResult, error) {
	const operation = "GetAlertResults"

	// Validation
	if alertID <= 0 {
		err := NewValidationError(operation, "Invalid alert ID", fmt.Sprintf("AlertID: %d", alertID))
		s.logError(ctx, err)
		return nil, err
	}

	if customerLat < -90 || customerLat > 90 {
		err := NewValidationError(operation, "Invalid customer latitude", fmt.Sprintf("Latitude: %f", customerLat))
		s.logError(ctx, err)
		return nil, err
	}

	if customerLng < -180 || customerLng > 180 {
		err := NewValidationError(operation, "Invalid customer longitude", fmt.Sprintf("Longitude: %f", customerLng))
		s.logError(ctx, err)
		return nil, err
	}

	log.Printf("Fetching alert results - AlertID: %d, CustomerLocation: (%.6f, %.6f)",
		alertID, customerLat, customerLng)

	// Get alert details with detailed error handling
	alert, err := s.queries.GetMedicationAlert(ctx, alertID)
	if err != nil {
		if err == pgx.ErrNoRows {
			notFoundErr := NewNotFoundError(operation, "Alert not found")
			s.logError(ctx, notFoundErr, "AlertID", alertID)
			return nil, notFoundErr
		}
		dbErr := NewDatabaseError(operation, "Failed to fetch alert", err)
		s.logError(ctx, dbErr, "AlertID", alertID)
		return nil, dbErr
	}

	// Get all responses with error handling
	responses, err := s.queries.GetPharmacistResponsesForAlert(ctx, sqlc.GetPharmacistResponsesForAlertParams{
		AlertID:   pgtype.Int4{Int32: alertID, Valid: true},
		Radians:   customerLat,
		Radians_2: customerLng,
	})
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to fetch pharmacist responses", err)
		s.logError(ctx, dbErr, "AlertID", alertID)
		return nil, dbErr
	}

	log.Printf("Found %d responses for alert %d", len(responses), alertID)

	pharmacies := make([]PharmacyAvailability, len(responses))
	for i, resp := range responses {
		pharmacies[i] = PharmacyAvailability{
			PharmacyID:        resp.PharmacyID.Int32,
			PharmacyName:      resp.PharmacyName,
			PharmacyAddress:   resp.PharmacyAddress,
			PharmacyPhone:     resp.PharmacyPhone,
			Latitude:          resp.PharmacyLatitude.Float64,
			Longitude:         resp.PharmacyLongitude.Float64,
			DistanceKM:        float64(resp.DistanceKm),
			ResponseType:      resp.ResponseType,
			SubstituteBrand:   resp.SubstituteBrand.String,
			SubstituteNotes:   resp.SubstituteNotes.String,
			ResponseTime:      resp.ResponseTimeSeconds.Int32,
			AvailableQuantity: resp.AvailableQuantity.Int32, // 0 when no stock row exists
		}
	}

	var requestedQty int32 = 1
	_ = s.db.QueryRow(ctx, `SELECT COALESCE(requested_quantity, 1) FROM medication_alerts WHERE id = $1`, alertID).Scan(&requestedQty)

	result := &AlertResult{
		AlertID:           alert.ID,
		MedicationID:      alert.MedicationID.Int32,
		Status:            alert.Status,
		RequestedQuantity: requestedQty,
		Pharmacies:        pharmacies,
		CreatedAt:         alert.CreatedAt.Time,
		ExpiresAt:         alert.ExpiresAt.Time,
	}

	log.Printf("Successfully retrieved alert results - AlertID: %d, Responses: %d", alertID, len(pharmacies))
	return result, nil
}

// CleanupExpiredAlerts marks expired alerts and cleans up old data
func (s *AlertService) CleanupExpiredAlerts(ctx context.Context) error {
	const operation = "CleanupExpiredAlerts"

	log.Printf("Starting cleanup of expired alerts")
	startTime := time.Now()

	// Mark expired alerts with improved error handling
	result, err := s.db.Exec(ctx, `
		UPDATE medication_alerts 
		SET status = 'expired' 
		WHERE status = 'pending' AND expires_at <= now()
	`)
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to expire alerts", err)
		s.logError(ctx, dbErr)
		return dbErr
	}

	expiredCount := result.RowsAffected()
	log.Printf("Marked %d alerts as expired", expiredCount)

	// Clean up expired responses with error handling
	cleanupErr := s.queries.CleanupExpiredResponses(ctx)
	if cleanupErr != nil {
		dbErr := NewDatabaseError(operation, "Failed to cleanup expired responses", cleanupErr)
		s.logError(ctx, dbErr)
		return dbErr
	}

	duration := time.Since(startTime)
	log.Printf("Cleanup completed successfully in %v - Expired alerts: %d", duration, expiredCount)
	return nil
}

// PharmacyDashboardAlert represents a pending alert shown on a pharmacy's dashboard
type PharmacyDashboardAlert struct {
	ID                 int32     `json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	ExpiresAt          time.Time `json:"expires_at"`
	MedicationName     string    `json:"medication_name"`
	ActiveSubstance    string    `json:"active_substance"`
	CustomerName       string    `json:"customer_name"`
	CustomerDistanceKm float64   `json:"customer_distance_km"`
	AlreadyResponded   bool      `json:"already_responded"`
	RequestedQuantity  int32     `json:"requested_quantity"`
}

// GetPharmacyDashboardAlerts returns pending medication alerts for a given pharmacy
func (s *AlertService) GetPharmacyDashboardAlerts(ctx context.Context, pharmacyID int32) ([]PharmacyDashboardAlert, error) {
	const operation = "GetPharmacyDashboardAlerts"

	if pharmacyID <= 0 {
		err := NewValidationError(operation, "Invalid pharmacy ID", fmt.Sprintf("PharmacyID: %d", pharmacyID))
		s.logError(ctx, err)
		return nil, err
	}

	rows, err := s.queries.GetPharmacyDashboardAlerts(ctx, pgtype.Int4{Int32: pharmacyID, Valid: true})
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to fetch pharmacy dashboard alerts", err)
		s.logError(ctx, dbErr, "PharmacyID", pharmacyID)
		return nil, dbErr
	}

	ids := make([]int32, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	quantities := map[int32]int32{}
	if len(ids) > 0 {
		if qrows, qerr := s.db.Query(ctx, `SELECT id, COALESCE(requested_quantity, 1) FROM medication_alerts WHERE id = ANY($1)`, ids); qerr == nil {
			for qrows.Next() {
				var id, q int32
				if scanErr := qrows.Scan(&id, &q); scanErr == nil {
					quantities[id] = q
				}
			}
			qrows.Close()
		}
	}

	alerts := make([]PharmacyDashboardAlert, len(rows))
	for i, row := range rows {
		alerts[i] = PharmacyDashboardAlert{
			ID:                 row.ID,
			CreatedAt:          row.CreatedAt.Time,
			ExpiresAt:          row.ExpiresAt.Time,
			MedicationName:     row.MedicationName,
			ActiveSubstance:    row.ActiveSubstance,
			CustomerName:       row.CustomerName,
			CustomerDistanceKm: row.CustomerDistanceKm,
			AlreadyResponded:   row.AlreadyResponded,
			RequestedQuantity:  max(quantities[row.ID], 1),
		}
	}

	return alerts, nil
}

// sendNotificationToPharmacy pushes a real-time "new alert" event to a pharmacy
// over its open SSE connection(s), if any are currently listening.
func (s *AlertService) sendNotificationToPharmacy(ctx context.Context, pharmacyID, alertID, medicationID int32) error {
	log.Printf("Sending notification to pharmacy %d for alert %d", pharmacyID, alertID)

	if s.sseHub != nil {
		s.sseHub.BroadcastToPharmacy(pharmacyID, PharmacyAlertEvent{
			Type:    "new_alert",
			AlertID: alertID,
		})
	}

	return nil
}

// notifyCustomerOfResponse pushes a real-time "new response" event to a customer
// watching an alert over its open SSE connection(s), if any are currently listening.
func (s *AlertService) notifyCustomerOfResponse(ctx context.Context, customerID, alertID int32) error {
	log.Printf("Notifying customer %d of response to alert %d", customerID, alertID)

	if s.sseHub != nil {
		s.sseHub.BroadcastToAlert(alertID, CustomerAlertEvent{
			Type:    "new_response",
			AlertID: alertID,
		})
	}

	return nil
}

// dbQuerier is satisfied by both *pgxpool.Pool and pgx.Tx.
type dbQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// notifyNextPharmacy picks the nearest pharmacy (within the alert's radius)
// that carries the medication with a positive stock quantity and has not been
// alerted yet. The requested quantity is intentionally NOT part of selection:
// a pharmacy with one unit can still be tried and then answer unavailable when
// the patient needs more than one. It does NOT send the SSE push; the caller
// does that after committing.
func (s *AlertService) notifyNextPharmacy(ctx context.Context, q dbQuerier, alertID int32) (int32, bool, error) {
	var medicationID int32
	var lat, lng, radius float64
	err := q.QueryRow(ctx, `
		SELECT medication_id, customer_latitude, customer_longitude,
		       COALESCE(search_radius_km, 10)
		FROM medication_alerts WHERE id = $1`, alertID).Scan(&medicationID, &lat, &lng, &radius)
	if err != nil {
		return 0, false, err
	}

	var pharmacyID int32
	err = q.QueryRow(ctx, `
		SELECT id FROM (
			SELECT p.id,
			  6371 * acos(LEAST(1.0, GREATEST(-1.0,
			    cos(radians($1::float8)) * cos(radians(p.latitude)) *
			    cos(radians(p.longitude) - radians($2::float8)) +
			    sin(radians($1::float8)) * sin(radians(p.latitude))
			  ))) AS dist
			FROM pharmacies p
			JOIN stock st ON st.pharmacy_id = p.id AND st.medication_id = $4 AND st.quantity > 0
			WHERE p.latitude IS NOT NULL AND p.longitude IS NOT NULL
			  AND NOT EXISTS (
			    SELECT 1 FROM pharmacy_alert_notifications n
			    WHERE n.alert_id = $5 AND n.pharmacy_id = p.id)
		) t
		WHERE dist <= $3
		ORDER BY dist ASC
		LIMIT 1`, lat, lng, radius, medicationID, alertID).Scan(&pharmacyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}

	if _, err = q.Exec(ctx, `
		INSERT INTO pharmacy_alert_notifications (alert_id, pharmacy_id)
		VALUES ($1, $2) ON CONFLICT (alert_id, pharmacy_id) DO NOTHING`, alertID, pharmacyID); err != nil {
		return 0, false, err
	}
	if _, err = q.Exec(ctx, `
		UPDATE medication_alerts
		SET expires_at = now() + INTERVAL '2 minutes',
		    total_pharmacies_notified = COALESCE(total_pharmacies_notified, 0) + 1
		WHERE id = $1`, alertID); err != nil {
		return 0, false, err
	}
	return pharmacyID, true, nil
}

// dispatchPharmacyAlert decides how the currently-selected pharmacy is handled.
// Pharmacies linked to a user account follow the normal dashboard/SSE workflow.
// Pharmacies without an account are evaluated automatically from the current
// stock table so they can still participate in the nearby one-by-one search.
func (s *AlertService) dispatchPharmacyAlert(ctx context.Context, pharmacyID, alertID, medicationID int32) {
	var hasManagedAccount bool
	if err := s.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM users WHERE managed_pharmacy_id = $1
		)`, pharmacyID).Scan(&hasManagedAccount); err != nil {
		s.logError(ctx, NewDatabaseError("dispatchPharmacyAlert", "Failed to determine pharmacy account status", err),
			"AlertID", alertID, "PharmacyID", pharmacyID)
		return
	}

	if hasManagedAccount {
		_ = s.sendNotificationToPharmacy(ctx, pharmacyID, alertID, medicationID)
		s.scheduleEscalation(alertID, pharmacyID)
		return
	}

	if err := s.processUnmanagedPharmacy(ctx, pharmacyID, alertID, medicationID); err != nil {
		s.logError(ctx, NewInternalError("dispatchPharmacyAlert", "Failed to auto-process unmanaged pharmacy", err),
			"AlertID", alertID, "PharmacyID", pharmacyID)
		// If automatic processing fails, still give this pharmacy the normal
		// response window before advancing so the alert cannot get stuck.
		s.scheduleEscalation(alertID, pharmacyID)
	}
}

// processUnmanagedPharmacy automatically answers the current alert from stock.
// It uses the same response path as a real pharmacist: enough stock means
// available; otherwise unavailable immediately advances to the next pharmacy.
func (s *AlertService) processUnmanagedPharmacy(ctx context.Context, pharmacyID, alertID, medicationID int32) error {
	var requestedQuantity, availableQuantity int32
	if err := s.db.QueryRow(ctx, `
		SELECT COALESCE(ma.requested_quantity, 1),
			COALESCE((
				SELECT st.quantity
				FROM stock st
				WHERE st.pharmacy_id = $2 AND st.medication_id = $3
			), 0)
		FROM medication_alerts ma
		WHERE ma.id = $1`, alertID, pharmacyID, medicationID).Scan(&requestedQuantity, &availableQuantity); err != nil {
		return err
	}

	responseType := "unavailable"
	if availableQuantity >= requestedQuantity {
		responseType = "available"
	}

	return s.SubmitPharmacistResponse(ctx, PharmacistResponseRequest{
		AlertID:             alertID,
		PharmacyID:          pharmacyID,
		ResponseType:        responseType,
		ResponseTimeSeconds: 0,
	})
}

// scheduleEscalation moves on to the next pharmacy if this one stays silent.
func (s *AlertService) scheduleEscalation(alertID, pharmacyID int32) {
	time.AfterFunc(pharmacyResponseTimeout, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var responded bool
		if err := s.db.QueryRow(ctx, `
			SELECT EXISTS (SELECT 1 FROM pharmacist_responses WHERE alert_id = $1 AND pharmacy_id = $2)`,
			alertID, pharmacyID).Scan(&responded); err != nil || responded {
			return
		}
		s.advanceAlert(ctx, alertID, pharmacyID)
	})
}

// advanceAlert alerts the next nearest pharmacy after fromPharmacyID failed to
// help (unavailable or no answer). When nobody is left it tells the patient.
func (s *AlertService) advanceAlert(ctx context.Context, alertID, fromPharmacyID int32) {
	const operation = "advanceAlert"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		s.logError(ctx, NewDatabaseError(operation, "Failed to begin transaction", err), "AlertID", alertID)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lock the alert so a timeout and an "unavailable" reply can't both advance it.
	var status string
	var customerID int32
	var medicationID int32
	if err = tx.QueryRow(ctx, `SELECT status, customer_id, medication_id FROM medication_alerts WHERE id = $1 FOR UPDATE`, alertID).Scan(&status, &customerID, &medicationID); err != nil {
		return
	}
	if status != "pending" {
		return
	}

	// Only advance from the most recently alerted pharmacy.
	var latest int32
	if err = tx.QueryRow(ctx, `SELECT pharmacy_id FROM pharmacy_alert_notifications WHERE alert_id = $1 ORDER BY id DESC LIMIT 1`, alertID).Scan(&latest); err != nil || latest != fromPharmacyID {
		return
	}

	// If someone already said available/substitute the patient is choosing.
	var hasOffer bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pharmacist_responses WHERE alert_id = $1 AND response_type IN ('available','substitute'))`, alertID).Scan(&hasOffer); err != nil || hasOffer {
		return
	}

	nextID, found, err := s.notifyNextPharmacy(ctx, tx, alertID)
	if err != nil {
		s.logError(ctx, NewDatabaseError(operation, "Failed to alert next pharmacy", err), "AlertID", alertID)
		return
	}
	if !found {
		if _, err = tx.Exec(ctx, `UPDATE medication_alerts SET status = 'expired' WHERE id = $1 AND status = 'pending'`, alertID); err != nil {
			return
		}
		if err = tx.Commit(ctx); err != nil {
			return
		}
		log.Printf("Alert %d: no more pharmacies to try", alertID)
		if s.sseHub != nil {
			s.sseHub.BroadcastToAlert(alertID, CustomerAlertEvent{Type: "exhausted", AlertID: alertID})
		}
		return
	}

	if err = tx.Commit(ctx); err != nil {
		return
	}

	log.Printf("Alert %d: moved on to pharmacy %d", alertID, nextID)
	// Notify the patient before dispatching: an unmanaged pharmacy may answer
	// immediately and recursively advance the alert to another pharmacy.
	if s.sseHub != nil {
		s.sseHub.BroadcastToAlert(alertID, CustomerAlertEvent{
			Type:    "trying_next",
			AlertID: alertID,
			Data:    map[string]int{"seconds": int(pharmacyResponseTimeout / time.Second)},
		})
	}
	s.dispatchPharmacyAlert(ctx, nextID, alertID, medicationID)
}
