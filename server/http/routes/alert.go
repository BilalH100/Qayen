package routes

import (
	"kayena/server/http/handlers"

	"github.com/go-chi/chi/v5"
)

func AlertRoutes(alertHandler *handlers.AlertHandler) chi.Router {
	r := chi.NewRouter()

	// Customer
	r.Post("/", alertHandler.CreateMedicationAlert)
	r.Get("/{id}/results", alertHandler.GetAlertResults)
	r.Get("/{id}/stream", alertHandler.CustomerAlertsSSE) //  (SSE)

	// Pharmacist
	r.Post("/{id}/response", alertHandler.SubmitPharmacistResponse)

	// Customer confirms pickup of a given quantity from a responding pharmacy
	r.Post("/{id}/confirm", alertHandler.ConfirmPickup)

	// Admin
	r.Post("/cleanup", alertHandler.CleanupExpiredAlerts)

	return r
}

// Pharmacy dashboard and analytics
func PharmacyAlertRoutes(alertHandler *handlers.AlertHandler) chi.Router {
	r := chi.NewRouter()

	r.Get("/{id}/dashboard", alertHandler.GetPharmacyDashboard)
	r.Get("/{id}/analytics", alertHandler.GetPharmacyAnalytics)
	r.Get("/{id}/alerts/stream", alertHandler.PharmacyAlertsSSE) // (SSE)

	return r
}
