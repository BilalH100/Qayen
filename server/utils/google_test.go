package utils

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestGoogleVerifier(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	verifier := &GoogleVerifier{keys: map[string]*rsa.PublicKey{"test": &key.PublicKey}, expires: time.Now().Add(time.Hour)}
	base := GoogleClaims{RegisteredClaims: jwt.RegisteredClaims{Issuer: "https://accounts.google.com", Subject: "google-user", Audience: jwt.ClaimStrings{"client-id"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}, Email: "user@gmail.com", EmailVerified: true, Nonce: "nonce"}
	sign := func(claims GoogleClaims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "test"
		value, e := token.SignedString(key)
		if e != nil {
			t.Fatal(e)
		}
		return value
	}
	if _, err = verifier.Verify(context.Background(), sign(base), "client-id", "nonce"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*GoogleClaims)
	}{
		{"wrong audience", func(c *GoogleClaims) { c.Audience = jwt.ClaimStrings{"other-client"} }},
		{"expired", func(c *GoogleClaims) { c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour)) }},
		{"missing expiry", func(c *GoogleClaims) { c.ExpiresAt = nil }},
		{"wrong issuer", func(c *GoogleClaims) { c.Issuer = "https://attacker.invalid" }},
		{"unverified email", func(c *GoogleClaims) { c.EmailVerified = false }},
		{"wrong nonce", func(c *GoogleClaims) { c.Nonce = "attacker" }},
		{"missing subject", func(c *GoogleClaims) { c.Subject = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := base
			tc.change(&c)
			if _, e := verifier.Verify(context.Background(), sign(c), "client-id", "nonce"); e == nil {
				t.Fatal("accepted invalid identity")
			}
		})
	}
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, base)
	forged.Header["kid"] = "test"
	forgedToken, _ := forged.SignedString([]byte("untrusted"))
	if _, err = verifier.Verify(context.Background(), forgedToken, "client-id", "nonce"); err == nil {
		t.Fatal("accepted wrong algorithm")
	}
	otherKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	forged = jwt.NewWithClaims(jwt.SigningMethodRS256, base)
	forged.Header["kid"] = "test"
	forgedToken, _ = forged.SignedString(otherKey)
	if _, err = verifier.Verify(context.Background(), forgedToken, "client-id", "nonce"); err == nil {
		t.Fatal("accepted forged signature")
	}
}
