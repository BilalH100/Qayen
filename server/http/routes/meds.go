package routes

import (
	"kayena/server/http/handlers"
	"kayena/server/services"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func MedRouter(s services.MedService) http.Handler {
	r := chi.NewRouter()
	r.Get("/id/{id}", handlers.GetMedicationByIdHandler(s))
	r.Get("/code/{code}", handlers.GetMedicationByCodeHandler(s))
	r.Get("/all", handlers.GetMedicationsHandler(s))
	r.Get("/categories/all", handlers.GetMedsCategories(s))
	r.Get("/search", handlers.SearchMedicationsHandler(s))
	r.Post("/create", handlers.CreateMedicationHandler(s))
	r.Delete("/id/{id}", handlers.DeleteMedicationHandler(s))
	r.Put("/id/{id}", handlers.UpdateMedicationHandler(s))
	return r
}
