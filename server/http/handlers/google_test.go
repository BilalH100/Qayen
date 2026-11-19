package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGoogleSignInGuards(t *testing.T) {
	t.Setenv("GOOGLE_ALLOWED_ORIGINS", "http://localhost:3000")
	t.Setenv("GOOGLE_CLIENT_ID", "")
	request := func(method, path, body string) *http.Request {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Origin", "http://localhost:3000")
		return r
	}
	w := httptest.NewRecorder()
	GoogleNonceHandler(w, request("GET", "/api/v1/users/google/nonce", ""))
	if w.Code != 503 {
		t.Fatalf("unconfigured status %d", w.Code)
	}
	t.Setenv("GOOGLE_CLIENT_ID", "client-id")
	w = httptest.NewRecorder()
	GoogleNonceHandler(w, request("GET", "/api/v1/users/google/nonce", ""))
	cookies := w.Result().Cookies()
	if w.Code != 200 || len(cookies) != 1 || !cookies[0].HttpOnly || len(cookies[0].Value) != 43 {
		t.Fatal("missing nonce cookie")
	}
	r := request("GET", "/api/v1/users/google/nonce", "")
	r.Header.Set("Origin", "https://attacker.invalid")
	w = httptest.NewRecorder()
	GoogleNonceHandler(w, r)
	if w.Code != 403 {
		t.Fatal("accepted untrusted origin")
	}
	handler := GoogleLoginHandler(nil)
	w = httptest.NewRecorder()
	handler(w, request("POST", "/api/v1/users/google", `{"credential":"forged"}`))
	if w.Code != 401 {
		t.Fatal("accepted missing nonce")
	}
	r = request("POST", "/api/v1/users/google", `{"credential":"forged"}`)
	r.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	handler(w, r)
	if w.Code != 401 {
		t.Fatal("accepted forged token")
	}
}
