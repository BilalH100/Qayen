package handlers

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"kayena/server/models"
	"kayena/server/services"
	"kayena/server/utils"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type profileTestUsers struct {
	services.UserService
	id          int32
	name, phone string
	calls       int
}

func (s *profileTestUsers) UpdateProfile(_ context.Context, id int32, name, phone string) (*models.User, error) {
	s.id = id
	s.name = name
	s.phone = phone
	s.calls++
	return &models.User{ID: id, Name: name, Phone: phone, Email: "account@example.com", Role: models.Regular}, nil
}

func TestUpdateMyProfile(t *testing.T) {
	t.Setenv("JWT_SECRECT", "profile-test-secret")
	valid, err := utils.GenerateJWT(models.User{ID: 7, Name: "Original"})
	if err != nil {
		t.Fatal(err)
	}
	expired, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": 7, "exp": time.Now().Add(-time.Hour).Unix()}).SignedString([]byte("profile-test-secret"))
	forged, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": 7, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("attacker-secret"))
	for _, tc := range []struct {
		name, token, body string
		status            int
	}{
		{"own profile", valid, `{"name":" New Name ","phone":" +212 600-000-000 "}`, 200},
		{"optional phone", valid, `{"name":"اسم جديد","phone":""}`, 200},
		{"no session", "", `{"name":"Test","phone":""}`, 401},
		{"expired session", expired, `{"name":"Test","phone":""}`, 401},
		{"forged session", forged, `{"name":"Test","phone":""}`, 401},
		{"blank name", valid, `{"name":"   ","phone":""}`, 400},
		{"bad phone", valid, `{"name":"Test","phone":"abc"}`, 400},
		{"too short phone", valid, `{"name":"Test","phone":"123"}`, 400},
		{"cannot edit another user", valid, `{"id":99,"name":"Test","phone":""}`, 400},
		{"cannot change role", valid, `{"role":"admin","name":"Test","phone":""}`, 400},
		{"email stays linked", valid, `{"email":"other@example.com","name":"Test","phone":""}`, 400},
		{"malformed JSON", valid, `{"name":`, 400},
		{"trailing JSON", valid, `{"name":"Test","phone":""}{}`, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &profileTestUsers{}
			r := httptest.NewRequest("PATCH", "/api/v1/users/me", strings.NewReader(tc.body))
			if tc.token != "" {
				r.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			UpdateMyProfileHandler(s)(w, r)
			if w.Code != tc.status {
				t.Fatalf("got status %d, want %d: %s", w.Code, tc.status, w.Body.String())
			}
			if tc.status == 200 {
				if s.calls != 1 || s.id != 7 || s.name != strings.TrimSpace(s.name) || s.phone != strings.TrimSpace(s.phone) {
					t.Fatal("wrong user or untrimmed data")
				}
				if strings.Contains(w.Body.String(), "password") {
					t.Fatal("password exposed")
				}
			} else if s.calls != 0 {
				t.Fatal("invalid request reached database")
			}
		})
	}
}
