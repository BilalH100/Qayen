package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	httpx "kayena/server/http"
	"kayena/server/repository"
	"kayena/server/schemas"
	"kayena/server/services"
	"kayena/server/utils"
	"net/http"
	"os"
	"strings"
	"time"
)

const googleNonceCookie = "qayen_google_nonce"

func googleOriginAllowed(w http.ResponseWriter, r *http.Request) bool {
	allowed := os.Getenv("GOOGLE_ALLOWED_ORIGINS")
	if allowed == "" {
		allowed = "http://localhost:3000,http://127.0.0.1:3000"
	}
	for _, origin := range strings.Split(allowed, ",") {
		if r.Header.Get("Origin") == strings.TrimSpace(origin) && strings.TrimSpace(origin) != "" {
			return true
		}
	}
	httpx.RespondWithJSON(w, 403, schemas.Message{Message: "Google sign-in is not enabled for this website address."})
	return false
}

func GoogleNonceHandler(w http.ResponseWriter, r *http.Request) {
	if !googleOriginAllowed(w, r) {
		return
	}
	if os.Getenv("GOOGLE_CLIENT_ID") == "" && os.Getenv("FIREBASE_PROJECT_ID") == "" {
		httpx.RespondWithJSON(w, 503, schemas.Message{Message: "Google sign-in is not configured yet."})
		return
	}
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		httpx.RespondWithJSON(w, 500, schemas.Message{Message: "Unable to start Google sign-in."})
		return
	}
	value := base64.RawURLEncoding.EncodeToString(nonce)
	http.SetCookie(w, &http.Cookie{Name: googleNonceCookie, Value: value, Path: "/api/v1/users/google", HttpOnly: true, Secure: r.TLS != nil || os.Getenv("GOOGLE_COOKIE_SECURE") == "true", SameSite: http.SameSiteLaxMode, MaxAge: 300, Expires: time.Now().Add(5 * time.Minute)})
	w.Header().Set("Cache-Control", "no-store")
	httpx.RespondWithJSON(w, 200, map[string]string{"nonce": value})
}

func GoogleLoginHandler(s services.UserService) http.HandlerFunc {
	verifier := utils.NewGoogleVerifier()
	return func(w http.ResponseWriter, r *http.Request) {
		if !googleOriginAllowed(w, r) {
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		clientID := os.Getenv("GOOGLE_CLIENT_ID")
		if clientID == "" {
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
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16384)).Decode(&body) != nil || body.Credential == "" {
			httpx.RespondWithJSON(w, 400, schemas.Message{Message: "Invalid Google sign-in request."})
			return
		}
		claims, err := verifier.Verify(r.Context(), body.Credential, clientID, cookie.Value)
		if err != nil {
			httpx.RespondWithJSON(w, 401, schemas.Message{Message: "Google sign-in could not be verified. Please try again."})
			return
		}
		completeGoogleLogin(w, r, s, claims.Subject, claims.Email, claims.Name)
	}
}

func completeGoogleLogin(w http.ResponseWriter, r *http.Request, s services.UserService, subject, email, name string) {
	user, err := s.GoogleLogin(r.Context(), subject, email, name)
	if errors.Is(err, repository.ErrGoogleEmailExists) {
		httpx.RespondWithJSON(w, 409, schemas.Message{Message: err.Error()})
		return
	}
	if err != nil {
		httpx.RespondWithJSON(w, 500, schemas.Message{Message: "Google sign-in is temporarily unavailable."})
		return
	}
	token, err := utils.GenerateJWT(*user)
	if err != nil {
		httpx.RespondWithJSON(w, 500, schemas.Message{Message: "Unable to create a session."})
		return
	}
	httpx.RespondWithJSON(w, 200, schemas.LoginResponse{Token: token, User: schemas.GetUserResponse{Id: user.ID, Email: user.Email, Name: user.Name, Phone: user.Phone, Role: user.Role, ManagedPharmacyId: user.ManagedPharmacyID}})
}
