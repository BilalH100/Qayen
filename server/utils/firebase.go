package utils

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const firebaseClockTolerance = 30 * time.Second

var ErrFirebaseClockSkew = errors.New("Firebase token is newer than the server clock")
var ErrFirebaseProjectMismatch = errors.New("Firebase project configuration mismatch")
var ErrFirebaseTokenExpired = errors.New("Firebase session expired")

// Only fixed diagnostic codes are logged; credentials and identity claims stay private.
func FirebaseVerificationReason(err error) string {
	switch {
	case errors.Is(err, ErrGoogleCertificatesUnavailable):
		return "certificates_unavailable"
	case errors.Is(err, ErrGoogleSigningKeyUnknown):
		return "unknown_signing_key"
	case errors.Is(err, ErrFirebaseClockSkew):
		return "server_clock_behind"
	case errors.Is(err, ErrFirebaseProjectMismatch):
		return "project_mismatch"
	case errors.Is(err, ErrFirebaseTokenExpired):
		return "expired_session"
	default:
		return "invalid_identity_or_signature"
	}
}

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
		jwt.WithValidMethods([]string{"RS256"}), jwt.WithAudience(projectID), jwt.WithIssuer("https://securetoken.google.com/"+projectID), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(firebaseClockTolerance))
	if err != nil {
		switch {
		case errors.Is(err, ErrGoogleCertificatesUnavailable), errors.Is(err, ErrGoogleSigningKeyUnknown):
			return nil, err
		case errors.Is(err, jwt.ErrTokenInvalidAudience), errors.Is(err, jwt.ErrTokenInvalidIssuer):
			return nil, ErrFirebaseProjectMismatch
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrFirebaseTokenExpired
		case errors.Is(err, jwt.ErrTokenUsedBeforeIssued), errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, ErrFirebaseClockSkew
		default:
			return nil, errors.New("invalid Firebase credential")
		}
	}
	if token == nil || !token.Valid {
		return nil, errors.New("invalid Firebase credential")
	}
	if claims.AuthTime > time.Now().Add(firebaseClockTolerance).Unix() {
		return nil, ErrFirebaseClockSkew
	}
	if claims.Subject == "" || len(claims.Subject) > 128 || claims.IssuedAt == nil || claims.AuthTime <= 0 || claims.AuthTime > claims.IssuedAt.Unix() || claims.Email == "" || !claims.EmailVerified {
		return nil, errors.New("invalid Firebase identity")
	}
	// Use Google's stable provider ID so existing Qayen Google accounts remain the same.
	googleIDs := claims.Firebase.Identities["google.com"]
	if claims.Firebase.SignInProvider != "google.com" || len(googleIDs) != 1 || googleIDs[0] == "" {
		return nil, errors.New("Google authentication required")
	}
	return claims, nil
}
