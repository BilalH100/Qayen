package utils

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type FirebaseClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	AuthTime      int64  `json:"auth_time"`
	Firebase      struct {
		SignInProvider string              `json:"sign_in_provider"`
		Identities     map[string][]string `json:"identities"`
	} `json:"firebase"`
}

func NewFirebaseVerifier() *GoogleVerifier {
	verifier := NewGoogleVerifier()
	verifier.certURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
	return verifier
}

func (v *GoogleVerifier) VerifyFirebase(ctx context.Context, credential, projectID string) (*FirebaseClaims, error) {
	if projectID == "" {
		return nil, errors.New("missing Firebase project")
	}
	claims := &FirebaseClaims{}
	token, err := jwt.ParseWithClaims(credential, claims, func(token *jwt.Token) (interface{}, error) { return v.key(ctx, token) },
		jwt.WithValidMethods([]string{"RS256"}), jwt.WithAudience(projectID), jwt.WithIssuer("https://securetoken.google.com/"+projectID), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid {
		return nil, errors.New("invalid Firebase credential")
	}
	if claims.Subject == "" || len(claims.Subject) > 128 || claims.IssuedAt == nil || claims.AuthTime <= 0 || claims.AuthTime > time.Now().Unix() || claims.AuthTime > claims.IssuedAt.Unix() || claims.Email == "" || !claims.EmailVerified {
		return nil, errors.New("invalid Firebase identity")
	}
	// Use Google's stable provider ID so existing Qayen Google accounts remain the same.
	googleIDs := claims.Firebase.Identities["google.com"]
	if claims.Firebase.SignInProvider != "google.com" || len(googleIDs) != 1 || googleIDs[0] == "" {
		return nil, errors.New("Google authentication required")
	}
	return claims, nil
}
