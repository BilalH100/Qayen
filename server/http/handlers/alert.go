package handlers

import (
	"encoding/json"
	"kayena/server/services"
	"kayena/server/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type AlertHandler struct {
	alertService *services.AlertService
}

func NewAlertHandler(alertService *services.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// CreateMedicationAlert handles POST /api/alerts
func (h *AlertHandler) CreateMedicationAlert(w http.ResponseWriter, r *http.Request) {
	var req services.AlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.CustomerID == 0 || req.MedicationID == 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Customer ID and Medication ID are required")
		return
	}

	if req.Latitude == 0 || req.Longitude == 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Location coordinates are required")
		return
	}

	// Set defaults
	if req.SearchRadius == 0 {
		req.SearchRadius = 10.0 // 10km default
	}
	if req.MaxResponseTime == 0 {
		req.MaxResponseTime = 2 // 2 minutes default
	}

	result, err := h.alertService.CreateMedicationAlert(r.Context(), req)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to create alert: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// GetAlertResults handles GET /api/alerts/{id}/results
func (h *AlertHandler) GetAlertResults(w http.ResponseWriter, r *http.Request) {
	alertIDStr := chi.URLParam(r, "id")
	alertID, err := strconv.ParseInt(alertIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid alert ID")
		return
	}

	// Get customer location from query params for distance calculation
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid latitude")
		return
	}
	
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid longitude")
		return
	}

	result, err := h.alertService.GetAlertResults(r.Context(), int32(alertID), lat, lng)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to get alert results: "+err.Error())
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
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid alert ID")
		return
	}

	var req services.PharmacistResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set alert ID from URL
	req.AlertID = int32(alertID)

	// Validate request
	if req.PharmacyID == 0 {
		utils.ErrorResponse(w, http.StatusBadRequest, "Pharmacy ID is required")
		return
	}

	if req.ResponseType != "available" && req.ResponseType != "unavailable" && req.ResponseType != "substitute" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid response type")
		return
	}

	if req.ResponseType == "substitute" && req.SubstituteMedicationID == nil && req.SubstituteBrand == "" {
		utils.ErrorResponse(w, http.StatusBadRequest, "Substitute details are required for substitute responses")
		return
	}

	err = h.alertService.SubmitPharmacistResponse(r.Context(), req)
	if err != nil {
		utils.ErrorResponse(w, http.StatusInternalServerError, "Failed to submit response: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Response submitted successfully",
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

	// For now, we'll implement a simplified dashboard
	// In a real implementation, this would call a specific service method
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"pharmacy_id": pharmacyID,
			"pending_alerts": []interface{}{}, // TODO: Implement GetPharmacyDashboardAlerts
			"today_stats": map[string]interface{}{
				"alerts_received": 0,
				"responses_sent":  0,
				"avg_response_time": 0,
				"availability_rate": 0,
			},
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
	_, err := strconv.ParseInt(pharmacyIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid pharmacy ID")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this connection
	clientChan := make(chan string, 10)

	// TODO: Register this client for real-time notifications
	// This would integrate with a WebSocket manager or pub/sub system
	
	// Keep connection alive and send heartbeat
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-clientChan:
			w.Write([]byte("data: " + msg + "\n\n"))
			w.(http.Flusher).Flush()
		case <-ticker.C:
			w.Write([]byte("data: {\"type\":\"heartbeat\"}\n\n"))
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}

// SSE endpoint for real-time updates to customers
func (h *AlertHandler) CustomerAlertsSSE(w http.ResponseWriter, r *http.Request) {
	alertIDStr := chi.URLParam(r, "id")
	_, err := strconv.ParseInt(alertIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(w, http.StatusBadRequest, "Invalid alert ID")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this connection
	clientChan := make(chan string, 10)

	// TODO: Register this client for real-time notifications about this alert
	
	// Keep connection alive and send heartbeat
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-clientChan:
			w.Write([]byte("data: " + msg + "\n\n"))
			w.(http.Flusher).Flush()
		case <-ticker.C:
			w.Write([]byte("data: {\"type\":\"heartbeat\"}\n\n"))
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}
