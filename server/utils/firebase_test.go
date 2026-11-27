package utils

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestFirebaseGoogleVerifier(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	verifier := &GoogleVerifier{keys: map[string]*rsa.PublicKey{"test": &key.PublicKey}, expires: time.Now().Add(time.Hour)}
	now := time.Now()
	base := FirebaseClaims{RegisteredClaims: jwt.RegisteredClaims{Issuer: "https://securetoken.google.com/qayen-2c858", Subject: "firebase-uid", Audience: jwt.ClaimStrings{"qayen-2c858"}, IssuedAt: jwt.NewNumericDate(now.Add(-time.Minute)), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour))}, Email: "user@gmail.com", EmailVerified: true, AuthTime: now.Add(-time.Minute).Unix()}
	base.Firebase.SignInProvider = "google.com"
	base.Firebase.Identities = map[string][]string{"google.com": {"original-google-subject"}}
	sign := func(claims FirebaseClaims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "test"
		value, e := token.SignedString(key)
		if e != nil {
			t.Fatal(e)
		}
		return value
	}
	verified, err := verifier.VerifyFirebase(context.Background(), sign(base), "qayen-2c858")
	if err != nil || verified.Firebase.Identities["google.com"][0] != "original-google-subject" {
		t.Fatalf("valid Google identity rejected: %v", err)
	}
	for _, tc := range []struct {
		name   string
		change func(*FirebaseClaims)
	}{
		{"other project audience", func(c *FirebaseClaims) { c.Audience = jwt.ClaimStrings{"attacker"} }},
		{"other project issuer", func(c *FirebaseClaims) { c.Issuer = "https://securetoken.google.com/attacker" }},
		{"Google OAuth token issuer", func(c *FirebaseClaims) { c.Issuer = "https://accounts.google.com" }},
		{"expired", func(c *FirebaseClaims) { c.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Hour)) }},
		{"missing expiry", func(c *FirebaseClaims) { c.ExpiresAt = nil }},
		{"missing issued at", func(c *FirebaseClaims) { c.IssuedAt = nil }},
		{"future issued at", func(c *FirebaseClaims) { c.IssuedAt = jwt.NewNumericDate(now.Add(time.Hour)) }},
		{"future authentication", func(c *FirebaseClaims) { c.AuthTime = now.Add(time.Hour).Unix() }},
		{"missing authentication", func(c *FirebaseClaims) { c.AuthTime = 0 }},
		{"unverified email", func(c *FirebaseClaims) { c.EmailVerified = false }},
		{"empty uid", func(c *FirebaseClaims) { c.Subject = "" }},
		{"password provider", func(c *FirebaseClaims) { c.Firebase.SignInProvider = "password" }},
		{"missing google identity", func(c *FirebaseClaims) { c.Firebase.Identities = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims := base
			tc.change(&claims)
			if _, e := verifier.VerifyFirebase(context.Background(), sign(claims), "qayen-2c858"); e == nil {
				t.Fatal("invalid identity accepted")
			}
		})
	}
	// Small real-world clock differences should not reject a fresh Google login.
	for _, seconds := range []int{1, 5, 20} {
		claims := base
		claims.IssuedAt = jwt.NewNumericDate(time.Now().Add(time.Duration(seconds) * time.Second))
		claims.AuthTime = claims.IssuedAt.Unix()
		if _, e := verifier.VerifyFirebase(context.Background(), sign(claims), "qayen-2c858"); e != nil {
			t.Fatalf("small clock difference rejected: %v", e)
		}
	}
	claimsBeyondTolerance := base
	claimsBeyondTolerance.IssuedAt = jwt.NewNumericDate(time.Now().Add(60 * time.Second))
	claimsBeyondTolerance.AuthTime = claimsBeyondTolerance.IssuedAt.Unix()
	if _, e := verifier.VerifyFirebase(context.Background(), sign(claimsBeyondTolerance), "qayen-2c858"); e == nil {
		t.Fatal("large future clock difference accepted")
	}
	claimsExpired := base
	claimsExpired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
	if _, e := verifier.VerifyFirebase(context.Background(), sign(claimsExpired), "qayen-2c858"); e == nil {
		t.Fatal("expired token accepted")
	}
	bad := jwt.NewWithClaims(jwt.SigningMethodHS256, base)
	bad.Header["kid"] = "test"
	raw, _ := bad.SignedString([]byte("forged"))
	if _, err = verifier.VerifyFirebase(context.Background(), raw, "qayen-2c858"); err == nil {
		t.Fatal("wrong algorithm accepted")
	}
	otherKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	bad = jwt.NewWithClaims(jwt.SigningMethodRS256, base)
	bad.Header["kid"] = "test"
	raw, _ = bad.SignedString(otherKey)
	if _, err = verifier.VerifyFirebase(context.Background(), raw, "qayen-2c858"); err == nil {
		t.Fatal("forged signature accepted")
	}
}
