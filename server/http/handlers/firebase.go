package handlers

import (
	"crypto/subtle"
	"encoding/json"
	httpx "kayena/server/http"
	"kayena/server/schemas"
	"kayena/server/services"
	"kayena/server/utils"
	"net/http"
	"os"
)

func FirebaseGoogleLoginHandler(s services.UserService) http.HandlerFunc {
	return firebaseGoogleLoginHandler(s, utils.NewFirebaseVerifier())
}

func firebaseGoogleLoginHandler(s services.UserService, verifier *utils.GoogleVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !googleOriginAllowed(w, r) {
			return
		}
		projectID := os.Getenv("FIREBASE_PROJECT_ID")
		if projectID == "" {
			httpx.RespondWithJSON(w, 503, schemas.Message{Message: "Google sign-in is not configured yet."})
			return
		}
		cookie, err := r.Cookie(googleNonceCookie)
		if err != nil || len(cookie.Value) != 43 {
			httpx.RespondWithJSON(w, 401, schemas.Message{Message: "Please reopen the sign-in window and try again."})
			return
		}
		http.SetCookie(w, &http.Cookie{Name: googleNonceCookie, Value: "", Path: "/api/v1/users/google", HttpOnly: true, Secure: r.TLS != nil || os.Getenv("GOOGLE_COOKIE_SECURE") == "true", SameSite: http.SameSiteLaxMode, MaxAge: -1})
		var body struct {
			Credential string `json:"credential"`
			Nonce      string `json:"nonce"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384)).Decode(&body) != nil || body.Credential == "" || subtle.ConstantTimeCompare([]byte(body.Nonce), []byte(cookie.Value)) != 1 {
			httpx.RespondWithJSON(w, 401, schemas.Message{Message: "Invalid Google sign-in request."})
			return
		}
		claims, err := verifier.VerifyFirebase(r.Context(), body.Credential, projectID)
		if err != nil {
			httpx.RespondWithJSON(w, 401, schemas.Message{Message: "Google sign-in could not be verified. Please try again."})
			return
		}
		completeGoogleLogin(w, r, s, claims.Firebase.Identities["google.com"][0], claims.Email, claims.Name)
	}
}
