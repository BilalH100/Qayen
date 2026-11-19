package handlers

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"io"
	httpx "kayena/server/http"
	"kayena/server/schemas"
	"kayena/server/services"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var profilePhoneCharacters = regexp.MustCompile(`^\+?[0-9 ()-]+$`)

// Only the authenticated user can edit their own contact data.
func UpdateMyProfileHandler(s services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		parts := strings.Fields(r.Header.Get("Authorization"))
		secret := os.Getenv("JWT_SECRECT")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || secret == "" {
			httpx.RespondWithJSON(w, 401, schemas.Message{Message: "Please sign in again to edit your profile."})
			return
		}
		token, err := jwt.Parse(parts[1], func(_ *jwt.Token) (interface{}, error) { return []byte(secret), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithJSONNumber())
		if err != nil || !token.Valid {
			httpx.RespondWithJSON(w, 401, schemas.Message{Message: "Please sign in again to edit your profile."})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			httpx.RespondWithJSON(w, 401, schemas.Message{Message: "Please sign in again to edit your profile."})
			return
		}
		subject, ok := claims["sub"].(json.Number)
		id, err := strconv.ParseInt(string(subject), 10, 32)
		if !ok || err != nil || id <= 0 {
			httpx.RespondWithJSON(w, 401, schemas.Message{Message: "Please sign in again to edit your profile."})
			return
		}
		var req struct {
			Name  string `json:"name"`
			Phone string `json:"phone"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&req) != nil || decoder.Decode(&struct{}{}) != io.EOF {
			httpx.RespondWithJSON(w, 400, schemas.Message{Message: "Invalid profile data."})
			return
		}
		req.Name, req.Phone = strings.TrimSpace(req.Name), strings.TrimSpace(req.Phone)
		if req.Name == "" || utf8.RuneCountInString(req.Name) > 100 {
			httpx.RespondWithJSON(w, 400, schemas.Message{Message: "Enter a name between 1 and 100 characters."})
			return
		}
		digits := 0
		for _, c := range req.Phone {
			if c >= '0' && c <= '9' {
				digits++
			}
		}
		if req.Phone != "" && (len(req.Phone) > 40 || !profilePhoneCharacters.MatchString(req.Phone) || digits < 10 || digits > 15) {
			httpx.RespondWithJSON(w, 400, schemas.Message{Message: "Enter a valid phone number with 10 to 15 digits."})
			return
		}
		user, err := s.UpdateProfile(r.Context(), int32(id), req.Name, req.Phone)
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondWithJSON(w, 401, schemas.Message{Message: "Please sign in again to edit your profile."})
			return
		}
		if err != nil {
			httpx.RespondWithJSON(w, 500, schemas.Message{Message: "Could not save your profile. Please try again."})
			return
		}
		httpx.RespondWithJSON(w, 200, schemas.GetUserResponse{Id: user.ID, Name: user.Name, Email: user.Email, Phone: user.Phone, Role: user.Role, ManagedPharmacyId: user.ManagedPharmacyID})
	}
}
