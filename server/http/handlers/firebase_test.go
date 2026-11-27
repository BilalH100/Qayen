package handlers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"kayena/server/models"
	"kayena/server/services"
	"kayena/server/utils"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type firebaseRoundTripper func(*http.Request) (*http.Response, error)

func (f firebaseRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type firebaseTestUsers struct {
	services.UserService
	subject string
}

func (s *firebaseTestUsers) GoogleLogin(_ context.Context, subject, email, name string) (*models.User, error) {
	s.subject = subject
	return &models.User{ID: 7, Email: email, Name: name, Role: models.Regular}, nil
}

func TestFirebaseNetworkLogin(t *testing.T) {
	t.Setenv("FIREBASE_PROJECT_ID", "qayen-2c858")
	t.Setenv("GOOGLE_ALLOWED_ORIGINS", "http://192.168.3.114:3000")
	t.Setenv("JWT_SECRECT", "isolated-test-signing-secret")
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certificates, _ := json.Marshal(map[string]string{"test": string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))})
	verifier := utils.NewFirebaseVerifier()
	verifier.Client = &http.Client{Transport: firebaseRoundTripper(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com" {
			t.Fatal("wrong certificate endpoint")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=3600"}}, Body: io.NopCloser(strings.NewReader(string(certificates)))}, nil
	})}
	claims := utils.FirebaseClaims{RegisteredClaims: jwt.RegisteredClaims{Issuer: "https://securetoken.google.com/qayen-2c858", Subject: "firebase-uid", Audience: jwt.ClaimStrings{"qayen-2c858"}, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}, Email: "test@gmail.com", Name: "Test", EmailVerified: true, AuthTime: now.Unix()}
	claims.Firebase.SignInProvider = "google.com"
	claims.Firebase.Identities = map[string][]string{"google.com": {"original-subject"}}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test"
	credential, _ := token.SignedString(key)
	nonce := strings.Repeat("x", 43)
	users := &firebaseTestUsers{}
	handler := firebaseGoogleLoginHandler(users, verifier)
	for _, tc := range []struct {
		name, origin, bodyNonce string
		cookie                  bool
		status                  int
	}{
		{"network success", "http://192.168.3.114:3000", nonce, true, 200},
		{"untrusted origin", "http://attacker.invalid", nonce, true, 403},
		{"missing cookie", "http://192.168.3.114:3000", nonce, false, 401},
		{"wrong csrf nonce", "http://192.168.3.114:3000", "attacker", true, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"credential": credential, "nonce": tc.bodyNonce})
			r := httptest.NewRequest("POST", "/api/v1/users/google/firebase", strings.NewReader(string(body)))
			r.Header.Set("Origin", tc.origin)
			if tc.cookie {
				r.AddCookie(&http.Cookie{Name: googleNonceCookie, Value: nonce})
			}
			w := httptest.NewRecorder()
			handler(w, r)
			if w.Code != tc.status {
				t.Fatalf("status %d, want %d", w.Code, tc.status)
			}
			if tc.status == 200 && (!strings.Contains(w.Body.String(), `"token"`) || users.subject != "original-subject") {
				t.Fatal("missing Qayen session or identity continuity")
			}
		})
	}

	t.Run("project mismatch gives actionable configuration error", func(t *testing.T) {
		t.Setenv("FIREBASE_PROJECT_ID", "other-project")
		body, _ := json.Marshal(map[string]string{"credential": credential, "nonce": nonce})
		req := httptest.NewRequest("POST", "/api/v1/users/google/firebase", strings.NewReader(string(body)))
		req.Header.Set("Origin", "http://192.168.3.114:3000")
		req.AddCookie(&http.Cookie{Name: googleNonceCookie, Value: nonce})
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != 503 || !strings.Contains(w.Body.String(), "configuration does not match") {
			t.Fatalf("unexpected failure: %d %s", w.Code, w.Body.String())
		}
	})
	t.Run("certificate outage is a temporary service error", func(t *testing.T) {
		offline := utils.NewFirebaseVerifier()
		offline.Client = &http.Client{Transport: firebaseRoundTripper(func(*http.Request) (*http.Response, error) { return nil, io.ErrUnexpectedEOF })}
		body, _ := json.Marshal(map[string]string{"credential": credential, "nonce": nonce})
		req := httptest.NewRequest("POST", "/api/v1/users/google/firebase", strings.NewReader(string(body)))
		req.Header.Set("Origin", "http://192.168.3.114:3000")
		req.AddCookie(&http.Cookie{Name: googleNonceCookie, Value: nonce})
		w := httptest.NewRecorder()
		firebaseGoogleLoginHandler(users, offline)(w, req)
		if w.Code != 503 || !strings.Contains(w.Body.String(), "temporarily unreachable") {
			t.Fatalf("unexpected outage response: %d %s", w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), credential) || strings.Contains(w.Body.String(), "test@gmail.com") {
			t.Fatal("credential or identity leaked")
		}
	})
}
