package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func pickupUser(w http.ResponseWriter, r *http.Request) (int32, bool) {
	parts := strings.Fields(r.Header.Get("Authorization"))
	secret := os.Getenv("JWT_SECRECT")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || secret == "" {
		http.Error(w, "Please sign in", 401)
		return 0, false
	}
	token, err := jwt.Parse(parts[1], func(_ *jwt.Token) (interface{}, error) { return []byte(secret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithJSONNumber())
	if err != nil || !token.Valid {
		http.Error(w, "Please sign in", 401)
		return 0, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Please sign in", 401)
		return 0, false
	}
	sub, ok := claims["sub"].(json.Number)
	id, err := strconv.ParseInt(string(sub), 10, 32)
	if !ok || err != nil || id <= 0 {
		http.Error(w, "Please sign in", 401)
		return 0, false
	}
	return int32(id), true
}
func (h *AlertHandler) ListPickups(w http.ResponseWriter, r *http.Request) {
	user, ok := pickupUser(w, r)
	if !ok {
		return
	}
	var pharmacy int64
	var err error
	if raw := chi.URLParam(r, "pharmacyID"); raw != "" {
		pharmacy, err = strconv.ParseInt(raw, 10, 32)
		if err != nil || pharmacy <= 0 {
			http.Error(w, "Invalid pharmacy", 400)
			return
		}
	}
	items, err := h.alertService.ListPickups(r.Context(), user, int32(pharmacy))
	if err != nil {
		h.handleAlertError(w, err, "Cannot load reservations")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": items})
}
func (h *AlertHandler) CollectPickup(w http.ResponseWriter, r *http.Request) {
	user, ok := pickupUser(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "pickupID"), 10, 32)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid reservation", 400)
		return
	}
	if err = h.alertService.CollectPickup(r.Context(), user, int32(id)); err != nil {
		h.handleAlertError(w, err, "Cannot confirm collection")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
