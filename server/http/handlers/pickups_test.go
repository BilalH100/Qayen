package handlers

import (
	"github.com/golang-jwt/jwt/v5"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPickupAuthentication(t *testing.T) {
	t.Setenv("JWT_SECRECT", "pickup-test-secret")
	for _, tc := range []struct {
		name           string
		method         jwt.SigningMethod
		sub            interface{}
		expired, valid bool
	}{
		{"valid patient", jwt.SigningMethodHS256, 42, false, true},
		{"invalid subject", jwt.SigningMethodHS256, "other", false, false},
		{"expired", jwt.SigningMethodHS256, 42, true, false},
		{"wrong signing method", jwt.SigningMethodHS384, 42, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			expiry := time.Now().Add(time.Hour)
			if tc.expired {
				expiry = time.Now().Add(-time.Hour)
			}
			token, err := jwt.NewWithClaims(tc.method, jwt.MapClaims{"sub": tc.sub, "exp": expiry.Unix()}).SignedString([]byte("pickup-test-secret"))
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			id, ok := pickupUser(w, req)
			if ok != tc.valid || (ok && id != 42) || (!ok && w.Code != 401) {
				t.Fatalf("unexpected authentication: %d %v", id, ok)
			}
		})
	}
	w := httptest.NewRecorder()
	if _, ok := pickupUser(w, httptest.NewRequest("GET", "/", nil)); ok || w.Code != 401 {
		t.Fatal("missing token accepted")
	}
}
