package services

import (
	"context"
	"fmt"
	sqlc "kayena/server/database/generated"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
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
}

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
	AlertID    int32                  `json:"alert_id"`
	Status     string                 `json:"status"`
	Pharmacies []PharmacyAvailability `json:"pharmacies"`
	CreatedAt  time.Time              `json:"created_at"`
	ExpiresAt  time.Time              `json:"expires_at"`
}

type PharmacyAvailability struct {
	PharmacyID      int32   `json:"pharmacy_id"`
	PharmacyName    string  `json:"pharmacy_name"`
	PharmacyAddress string  `json:"pharmacy_address"`
	PharmacyPhone   string  `json:"pharmacy_phone"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	DistanceKM      float64 `json:"distance_km"`
	ResponseType    string  `json:"response_type"`
	SubstituteBrand string  `json:"substitute_brand,omitempty"`
	SubstituteNotes string  `json:"substitute_notes,omitempty"`
	ResponseTime    int32   `json:"response_time_seconds"`
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

	// Get nearby pharmacies that actually have this medication in stock
	nearbyPharmacies, err := qtx.GetNearbyPharmaciesWithStockForAlert(ctx, sqlc.GetNearbyPharmaciesWithStockForAlertParams{
		Radians:      req.Latitude,                                          // $1 - customer latitude
		Radians_2:    req.Longitude,                                         // $2 - customer longitude
		Latitude:     pgtype.Float8{Float64: req.SearchRadius, Valid: true}, // $3 - search radius
		MedicationID: pgtype.Int4{Int32: req.MedicationID, Valid: true},     // $4 - requested medication
	})
	if err != nil {
		dbErr := NewDatabaseError(operation, "Failed to fetch nearby pharmacies with stock", err)
		s.logError(ctx, dbErr, "AlertID", alert.ID, "SearchRadius", req.SearchRadius)
		return nil, dbErr
	}

	log.Printf("Found %d nearby pharmacies for alert %d", len(nearbyPharmacies), alert.ID)

	notifiedCount := 0
	failedNotifications := 0

	for i, pharmacy := range nearbyPharmacies {
		pharmacyContext := fmt.Sprintf("Pharmacy[%d] ID:%d Name:%s", i, pharmacy.ID, pharmacy.Name)

		// Check if pharmacy is open
		isOpen, err := qtx.CheckPharmacyIsOpen(ctx, pharmacy.ID)
		if err != nil {
			dbErr := NewDatabaseError(operation, "Error checking pharmacy open status", err)
			s.logError(ctx, dbErr, "Context", pharmacyContext)
			failedNotifications++
			continue
		}
		if !isOpen {
			log.Printf("Pharmacy %d (%s) is closed, skipping notification", pharmacy.ID, pharmacy.Name)
			continue
		}

		// Check rate limiting with enhanced error handling
		rateLimit, err := qtx.CheckAlertRateLimit(ctx, sqlc.CheckAlertRateLimitParams{
			PharmacyID:      pgtype.Int4{Int32: pharmacy.ID, Valid: true},
			MedicationID:    pgtype.Int4{Int32: req.MedicationID, Valid: true},
			AlertCountToday: pgtype.Int4{Int32: 10, Valid: true}, // Max 10 alerts per day
		})
		if err != nil {
			// If no rate limit record exists, we can send the alert
			if err == pgx.ErrNoRows {
				log.Printf("No rate limit record exists for %s, creating new one", pharmacyContext)
				// Create new rate limit record
				_, createErr := qtx.UpdateAlertRateLimit(ctx, sqlc.UpdateAlertRateLimitParams{
					PharmacyID:   pgtype.Int4{Int32: pharmacy.ID, Valid: true},
					MedicationID: pgtype.Int4{Int32: req.MedicationID, Valid: true},
				})
				if createErr != nil {
					dbErr := NewDatabaseError(operation, "Failed to create rate limit record", createErr)
					s.logError(ctx, dbErr, "Context", pharmacyContext)
					failedNotifications++
					continue
				}
			} else {
				dbErr := NewDatabaseError(operation, "Error checking rate limit", err)
				s.logError(ctx, dbErr, "Context", pharmacyContext)
				failedNotifications++
				continue
			}
		} else if !rateLimit.CanSendAlert {
			log.Printf("Rate limit exceeded for %s, skipping notification", pharmacyContext)
			continue
		} else {
			// Update rate limit
			_, updateErr := qtx.UpdateAlertRateLimit(ctx, sqlc.UpdateAlertRateLimitParams{
				PharmacyID:   pgtype.Int4{Int32: pharmacy.ID, Valid: true},
				MedicationID: pgtype.Int4{Int32: req.MedicationID, Valid: true},
			})
			if updateErr != nil {
				dbErr := NewDatabaseError(operation, "Failed to update rate limit", updateErr)
				s.logError(ctx, dbErr, "Context", pharmacyContext)
				failedNotifications++
				continue
			}
		}

		// Create notification record
		_, err = qtx.CreatePharmacyAlertNotification(ctx, sqlc.CreatePharmacyAlertNotificationParams{
			AlertID:    pgtype.Int4{Int32: alert.ID, Valid: true},
			PharmacyID: pgtype.Int4{Int32: pharmacy.ID, Valid: true},
		})
		if err != nil {
			dbErr := NewDatabaseError(operation, "Failed to create notification record", err)
			s.logError(ctx, dbErr, "Context", pharmacyContext, "AlertID", alert.ID)
			failedNotifications++
			continue
		}

		// Send real-time notification to pharmacy
		notifyErr := s.sendNotificationToPharmacy(ctx, pharmacy.ID, alert.ID, req.MedicationID)
		if notifyErr != nil {
			internalErr := NewInternalError(operation, "Failed to send notification to pharmacy", notifyErr)
			s.logError(ctx, internalErr, "Context", pharmacyContext, "AlertID", alert.ID)
			failedNotifications++
			continue
		}

		log.Printf("Successfully notified %s for alert %d", pharmacyContext, alert.ID)
		notifiedCount++
	}

	log.Printf("Alert %d notification summary - Total pharmacies: %d, Notified: %d, Failed: %d",
		alert.ID, len(nearbyPharmacies), notifiedCount, failedNotifications)

	if notifiedCount == 0 {
		businessErr := NewBusinessLogicError(operation, "No pharmacies could be notified",
			fmt.Sprintf("Total pharmacies found: %d, Failed notifications: %d", len(nearbyPharmacies), failedNotifications))
		s.logError(ctx, businessErr, "AlertID", alert.ID)
		return nil, businessErr
	}

	// Update alert with notification count
	updatedAlert, err := qtx.UpdateMedicationAlertStatus(ctx, sqlc.UpdateMedicationAlertStatusParams{
		ID:                     alert.ID,
		Status:                 "pending",
		TotalResponsesReceived: pgtype.Int4{Int32: int32(notifiedCount), Valid: true},
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

	result := &AlertResult{
		AlertID:    updatedAlert.ID,
		Status:     updatedAlert.Status,
		CreatedAt:  updatedAlert.CreatedAt.Time,
		ExpiresAt:  updatedAlert.ExpiresAt.Time,
		Pharmacies: []PharmacyAvailability{}, // Will be populated as responses come in
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

	return nil
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
			PharmacyID:      resp.PharmacyID.Int32,
			PharmacyName:    resp.PharmacyName,
			PharmacyAddress: resp.PharmacyAddress,
			PharmacyPhone:   resp.PharmacyPhone,
			Latitude:        resp.PharmacyLatitude.Float64,
			Longitude:       resp.PharmacyLongitude.Float64,
			DistanceKM:      float64(resp.DistanceKm),
			ResponseType:    resp.ResponseType,
			SubstituteBrand: resp.SubstituteBrand.String,
			SubstituteNotes: resp.SubstituteNotes.String,
			ResponseTime:    resp.ResponseTimeSeconds.Int32,
		}
	}

	result := &AlertResult{
		AlertID:    alert.ID,
		Status:     alert.Status,
		Pharmacies: pharmacies,
		CreatedAt:  alert.CreatedAt.Time,
		ExpiresAt:  alert.ExpiresAt.Time,
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
