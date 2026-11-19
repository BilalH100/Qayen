package utils

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GoogleClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Nonce         string `json:"nonce"`
}

type GoogleVerifier struct {
	Client  *http.Client
	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
	certURL string
}

func NewGoogleVerifier() *GoogleVerifier {
	return &GoogleVerifier{Client: &http.Client{Timeout: 8 * time.Second}}
}

func (v *GoogleVerifier) key(ctx context.Context, token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok || kid == "" {
		return nil, errors.New("missing key id")
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if time.Now().Before(v.expires) {
		if key := v.keys[kid]; key != nil {
			return key, nil
		}
		return nil, errors.New("unknown Google signing key")
	}
	certURL := v.certURL
	if certURL == "" {
		certURL = "https://www.googleapis.com/oauth2/v1/certs"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, certURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google certificates unavailable")
	}
	var certs map[string]string
	if err = json.NewDecoder(io.LimitReader(resp.Body, 65536)).Decode(&certs); err != nil {
		return nil, err
	}
	keys := make(map[string]*rsa.PublicKey)
	for id, value := range certs {
		block, _ := pem.Decode([]byte(value))
		if block == nil {
			continue
		}
		cert, parseErr := x509.ParseCertificate(block.Bytes)
		if parseErr != nil {
			continue
		}
		if key, valid := cert.PublicKey.(*rsa.PublicKey); valid {
			keys[id] = key
		}
	}
	if len(keys) == 0 {
		return nil, errors.New("invalid Google certificates")
	}
	cacheTime := 5 * time.Minute
	for _, directive := range strings.Split(resp.Header.Get("Cache-Control"), ",") {
		if seconds, parseErr := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(directive), "max-age=")); strings.HasPrefix(strings.TrimSpace(directive), "max-age=") && parseErr == nil && seconds > 0 && seconds <= 86400 {
			cacheTime = time.Duration(seconds) * time.Second
		}
	}
	v.keys, v.expires = keys, time.Now().Add(cacheTime)
	if key := keys[kid]; key != nil {
		return key, nil
	}
	return nil, errors.New("unknown Google signing key")
}

func (v *GoogleVerifier) Verify(ctx context.Context, credential, audience, nonce string) (*GoogleClaims, error) {
	if audience == "" || nonce == "" {
		return nil, errors.New("missing Google sign-in configuration")
	}
	claims := &GoogleClaims{}
	token, err := jwt.ParseWithClaims(credential, claims, func(token *jwt.Token) (interface{}, error) { return v.key(ctx, token) },
		jwt.WithValidMethods([]string{"RS256"}), jwt.WithAudience(audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid {
		return nil, errors.New("invalid Google credential")
	}
	if claims.Issuer != "https://accounts.google.com" && claims.Issuer != "accounts.google.com" {
		return nil, errors.New("invalid Google issuer")
	}
	if claims.Subject == "" || claims.Email == "" || !claims.EmailVerified || claims.Nonce != nonce {
		return nil, errors.New("invalid Google identity")
	}
	return claims, nil
}
