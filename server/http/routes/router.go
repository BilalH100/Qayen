package routes

import (
	"kayena/server/services"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRouter(s *services.Services) http.Handler {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			return true
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Route("/api/v1", func(r chi.Router) {
		r.Mount("/users", UserRouter(s.UserService))
		r.Mount("/meds", MedRouter(s.MedService))
		r.Mount("/pharmacies", PharmacyRouter(s.PharmacyService))
	})

	return router
}
