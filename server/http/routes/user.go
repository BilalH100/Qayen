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
	r.Get("/list", handlers.ListAllUsersHandler(s))
	r.Post("/register", handlers.CreateUserHandler(s))
	r.Post("/login", handlers.LoginHandler(s))
	r.Patch("/me", handlers.UpdateMyProfileHandler(s))
	r.Get("/google/nonce", handlers.GoogleNonceHandler)
	r.Post("/google", handlers.GoogleLoginHandler(s))
	r.Post("/google/firebase", handlers.FirebaseGoogleLoginHandler(s))
	r.Delete("/id/{id}", handlers.DeleteUserHandler(s))
	r.Put("/id/{id}/role", handlers.UpdateUserRole(s))
	r.Put("/id/{id}/assign-pharmacy", handlers.UpdateUserManagedPharmacyHandler(s))
	return r
}
