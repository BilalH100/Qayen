package routes

import "net/http"

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	// mux.Handle("/med")
	// mux.Handle("/auth")
	// mux.Handle("/pharmacies")
	return mux
}
