package utils

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"
)

type googleKeyRefreshTransport func(*http.Request) (*http.Response, error)

func (f googleKeyRefreshTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestFirebaseSigningKeyRotation(t *testing.T) {
	old, _ := rsa.GenerateKey(rand.Reader, 2048)
	fresh, _ := rsa.GenerateKey(rand.Reader, 2048)
	now := time.Now()
	template := &x509.Certificate{SerialNumber: big.NewInt(42), NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &fresh.PublicKey, fresh)
	if err != nil {
		t.Fatal(err)
	}
	certs, _ := json.Marshal(map[string]string{"rotated": string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))})
	calls := 0
	v := NewFirebaseVerifier()
	v.keys = map[string]*rsa.PublicKey{"old": &old.PublicKey}
	v.expires = now.Add(time.Hour)
	v.Client = &http.Client{Transport: googleKeyRefreshTransport(func(*http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Header: http.Header{"Cache-Control": []string{"max-age=3600"}}, Body: io.NopCloser(strings.NewReader(string(certs)))}, nil
	})}
	claims := FirebaseClaims{RegisteredClaims: jwt.RegisteredClaims{Issuer: "https://securetoken.google.com/project", Subject: "user-id", Audience: jwt.ClaimStrings{"project"}, IssuedAt: jwt.NewNumericDate(now.Add(-time.Minute)), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}, Email: "user@example.com", EmailVerified: true, AuthTime: now.Add(-time.Minute).Unix()}
	claims.Firebase.SignInProvider = "google.com"
	claims.Firebase.Identities = map[string][]string{"google.com": {"google-id"}}
	sign := func(kid string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = kid
		raw, e := token.SignedString(fresh)
		if e != nil {
			t.Fatal(e)
		}
		return raw
	}
	for i := 0; i < 2; i++ {
		if _, err = v.VerifyFirebase(context.Background(), sign("rotated"), "project"); err != nil {
			t.Fatalf("rotated Google key rejected: %v", err)
		}
	}
	if calls != 1 {
		t.Fatalf("certificate refreshes=%d, expected 1", calls)
	}
	for _, kid := range []string{"unknown-1", "unknown-2"} {
		if _, err = v.VerifyFirebase(context.Background(), sign(kid), "project"); err == nil {
			t.Fatal("unknown signing key accepted")
		}
	}
	if calls != 1 {
		t.Fatal("untrusted key IDs bypassed refresh cooldown")
	}
}
func TestFirebaseCertificateOutageIsClassified(t *testing.T) {
	v := NewFirebaseVerifier()
	v.Client = &http.Client{Transport: googleKeyRefreshTransport(func(*http.Request) (*http.Response, error) { return nil, io.ErrUnexpectedEOF })}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{})
	token.Header["kid"] = "key"
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	raw, _ := token.SignedString(key)
	_, err := v.VerifyFirebase(context.Background(), raw, "project")
	if !errors.Is(err, ErrGoogleCertificatesUnavailable) || FirebaseVerificationReason(err) != "certificates_unavailable" {
		t.Fatalf("outage misclassified: %v", err)
	}
}
