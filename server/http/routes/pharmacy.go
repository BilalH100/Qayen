package routes

import (
	"kayena/server/http/handlers"
	"kayena/server/services"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func PharmacyRouter(s services.PharmacyService) http.Handler {
	r := chi.NewRouter()
	r.Get("/closest", handlers.GetClosestPharmacyHandler(s))
	return r
}
