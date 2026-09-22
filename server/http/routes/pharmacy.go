package routes

import (
	"kayena/server/http/handlers"
	"kayena/server/services"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func PharmacyRouter(s services.PharmacyService) http.Handler {
	r := chi.NewRouter()
	r.Get("/", handlers.GetAllPharmaciesHandler(s))
	r.Get("/closest", handlers.GetClosestPharmacyHandler(s))
	r.Get("/id/{id}", handlers.GetPharmacyDetailsHandler(s))
	r.Put("/id/{id}", handlers.UpdatePharmacyHandler(s))
	r.Get("/id/{id}/stock", handlers.ListPharmacyStockHandler(s))
	r.Post("/id/{id}/stock", handlers.UpsertPharmacyStockHandler(s))
	r.Delete("/id/{id}/stock/{medId}", handlers.DeletePharmacyStockHandler(s))
	return r
}
