package handlers

import (
	"encoding/json"
	"errors"
	"kayena/server/services"
	"kayena/server/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type AlertHandler struct {
	alertService *services.AlertService
	sseHub       *services.SSEHub
}

func NewAlertHandler(alertService *services.AlertService, sseHub *services.SSEHub) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		sseHub:       sseHub,
	}
}

func (h *AlertHandler) CreateMedicationAlert(w http.ResponseWriter, r *http.Request) {
	var req services.AlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON in request body: "+err.Error())
		return
	}
	if req.CustomerID == 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Customer ID is required")
		return
	}

	if req.MedicationID == 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Medication ID is required")
		return
	}

	if req.Latitude == 0 && req.Longitude == 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Location coordinates are required")
		return
	}

	if req.SearchRadius <= 0 {
		req.SearchRadius = 10.0 // 10km default
	}
	if req.MaxResponseTime <= 0 {
		req.MaxResponseTime = 2 // 2 minutes default
	}

	// Limit search radius to reasonable bounds
	if req.SearchRadius > 100 {
		req.SearchRadius = 100
	}

	result, err := h.alertService.CreateMedicationAlert(r.Context(), req)
	if err != nil {
		h.handleAlertError(w, err, "Failed to create medication alert")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
		"message": "Medication alert created successfully",
	})
}

// GetAlertResults handles GET /api/alerts/{id}/results
func (h *AlertHandler) GetAlertResults(w http.ResponseWriter, r *http.Request) {
	alertIDStr := chi.URLParam(r, "id")
	alertID, err := strconv.ParseInt(alertIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid alert ID format: must be a positive integer")
		return
	}

	if alertID <= 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Alert ID must be positive")
		return
	}

	// Get customer location from query params for distance calculation
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")

	if latStr == "" || lngStr == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Customer location (lat, lng) query parameters are required")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid latitude format: "+err.Error())
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid longitude format: "+err.Error())
		return
	}

	result, err := h.alertService.GetAlertResults(r.Context(), int32(alertID), lat, lng)
	if err != nil {
		h.handleAlertError(w, err, "Failed to get alert results")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// SubmitPharmacistResponse handles POST /api/alerts/{id}/response
func (h *AlertHandler) SubmitPharmacistResponse(w http.ResponseWriter, r *http.Request) {
	alertIDStr := chi.URLParam(r, "id")
	alertID, err := strconv.ParseInt(alertIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid alert ID format: must be a positive integer")
		return
	}

	if alertID <= 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Alert ID must be positive")
		return
	}

	var req services.PharmacistResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON in request body: "+err.Error())
		return
	}

	// Set alert ID from URL parameter
	req.AlertID = int32(alertID)

	// Basic validation (detailed validation is done in service layer)
	if req.PharmacyID <= 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Valid pharmacy ID is required")
		return
	}

	validResponseTypes := []string{"available", "unavailable", "substitute"}
	validType := false
	for _, vt := range validResponseTypes {
		if req.ResponseType == vt {
			validType = true
			break
		}
	}
	if !validType {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid response type. Valid types: available, unavailable, substitute")
		return
	}

	// Set response time if not provided (time since alert was received)
	if req.ResponseTimeSeconds <= 0 {
		req.ResponseTimeSeconds = 30 // Default 30 seconds response time
	}

	err = h.alertService.SubmitPharmacistResponse(r.Context(), req)
	if err != nil {
		h.handleAlertError(w, err, "Failed to submit pharmacist response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Pharmacist response submitted successfully",
		"data": map[string]interface{}{
			"alert_id":      req.AlertID,
			"pharmacy_id":   req.PharmacyID,
			"response_type": req.ResponseType,
		},
	})
}

// GetPharmacyDashboard handles GET /api/pharmacy/{id}/dashboard
func (h *AlertHandler) GetPharmacyDashboard(w http.ResponseWriter, r *http.Request) {
	pharmacyIDStr := chi.URLParam(r, "id")
	pharmacyID, err := strconv.ParseInt(pharmacyIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid pharmacy ID")
		return
	}

	alerts, err := h.alertService.GetPharmacyDashboardAlerts(r.Context(), int32(pharmacyID))
	if err != nil {
		h.handleAlertError(w, err, "Failed to get pharmacy dashboard")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"pharmacy_id":    pharmacyID,
			"pending_alerts": alerts,
		},
	})
}

// CleanupExpiredAlerts handles POST /api/admin/alerts/cleanup (admin endpoint)
func (h *AlertHandler) CleanupExpiredAlerts(w http.ResponseWriter, r *http.Request) {
	err := h.alertService.CleanupExpiredAlerts(r.Context())
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to cleanup alerts: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Expired alerts cleaned up successfully",
	})
}

// GetPharmacyAnalytics handles GET /api/pharmacy/{id}/analytics
func (h *AlertHandler) GetPharmacyAnalytics(w http.ResponseWriter, r *http.Request) {
	pharmacyIDStr := chi.URLParam(r, "id")
	pharmacyID, err := strconv.ParseInt(pharmacyIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid pharmacy ID")
		return
	}

	// Get limit from query params (default 30 days)
	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// TODO: Implement GetPharmacyAnalytics in service
	// analytics, err := h.alertService.GetPharmacyAnalytics(r.Context(), int32(pharmacyID), int32(limit))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"pharmacy_id": pharmacyID,
			"period_days": limit,
			"analytics":   []interface{}{}, // TODO: Return actual analytics
		},
	})
}

// SSE endpoint for real-time alerts to pharmacies
func (h *AlertHandler) PharmacyAlertsSSE(w http.ResponseWriter, r *http.Request) {
	pharmacyIDStr := chi.URLParam(r, "id")
	pharmacyID64, err := strconv.ParseInt(pharmacyIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid pharmacy ID")
		return
	}
	pharmacyID := int32(pharmacyID64)

	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Streaming unsupported")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Create a channel for this connection and register it with the hub
	clientChan := make(chan string, 10)
	h.sseHub.RegisterPharmacy(pharmacyID, clientChan)
	defer h.sseHub.UnregisterPharmacy(pharmacyID, clientChan)

	// Keep connection alive and send heartbeat
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-clientChan:
			if !ok {
				// Hub closed the channel (e.g. during unregister elsewhere)
				return
			}
			w.Write([]byte("data: " + msg + "\n\n"))
			flusher.Flush()
		case <-ticker.C:
			w.Write([]byte("data: {\"type\":\"heartbeat\"}\n\n"))
			flusher.Flush()
		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// SSE endpoint for real-time updates to customers
func (h *AlertHandler) CustomerAlertsSSE(w http.ResponseWriter, r *http.Request) {
	alertIDStr := chi.URLParam(r, "id")
	alertID64, err := strconv.ParseInt(alertIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid alert ID")
		return
	}
	alertID := int32(alertID64)

	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Streaming unsupported")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Create a channel for this connection and register it with the hub
	clientChan := make(chan string, 10)
	h.sseHub.RegisterAlert(alertID, clientChan)
	defer h.sseHub.UnregisterAlert(alertID, clientChan)

	// Keep connection alive and send heartbeat
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			w.Write([]byte("data: " + msg + "\n\n"))
			flusher.Flush()
		case <-ticker.C:
			w.Write([]byte("data: {\"type\":\"heartbeat\"}\n\n"))
			flusher.Flush()
		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// Helper function to handle AlertError types and return appropriate HTTP responses
func (h *AlertHandler) handleAlertError(w http.ResponseWriter, err error, defaultMessage string) {
	var alertErr *services.AlertError
	if errors.As(err, &alertErr) {
		var statusCode int
		switch alertErr.Code {
		case "VALIDATION_ERROR":
			statusCode = http.StatusBadRequest
		case "NOT_FOUND":
			statusCode = http.StatusNotFound
		case "BUSINESS_LOGIC_ERROR":
			statusCode = http.StatusConflict
		case "DATABASE_ERROR":
			statusCode = http.StatusInternalServerError
		case "INTERNAL_ERROR":
			statusCode = http.StatusInternalServerError
		default:
			statusCode = http.StatusInternalServerError
		}

		// Create detailed error response
		errorResponse := map[string]interface{}{
			"success": false,
			"error": map[string]interface{}{
				"code":      alertErr.Code,
				"message":   alertErr.Message,
				"operation": alertErr.Operation,
			},
		}

		if alertErr.Details != "" {
			errorResponse["error"].(map[string]interface{})["details"] = alertErr.Details
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(errorResponse)
	} else {
		// Fallback for non-AlertError types
		utils.ErrorResponse(w, http.StatusInternalServerError, defaultMessage+": "+err.Error())
	}
}
