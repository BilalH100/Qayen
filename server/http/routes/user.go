package routes

import (
	"kayena/server/http/handlers"
	"kayena/server/services"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func UserRouter(s services.UserService) http.Handler {
	r := chi.NewRouter()
	r.Get("/id/{id}", handlers.GetUserProfileHandler(s))
	r.Post("/register", handlers.CreateUserHandler(s))
	r.Post("/login", handlers.LoginHandler(s))
	return r
}
