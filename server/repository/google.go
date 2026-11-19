package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"kayena/server/models"
	"strings"
)

var ErrGoogleEmailExists = errors.New("This email already has a Qayen account. Please sign in with your password.")

func (r *sqlcUserRepo) GoogleLogin(ctx context.Context, subject, email, name string) (*models.User, error) {
	var storedEmail string
	err := r.db.QueryRow(ctx, `SELECT email FROM users WHERE google_subject=$1`, subject).Scan(&storedEmail)
	if err == nil {
		return r.GetByEmail(ctx, storedEmail)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	// Never link an existing password account merely because its email matches.
	var exists bool
	if err = r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email)=lower($1))`, email).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrGoogleEmailExists
	}
	secret := make([]byte, 32)
	if _, err = rand.Read(secret); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(base64.RawURLEncoding.EncodeToString(secret)), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRow(ctx, `INSERT INTO users(email,name,phone,password,user_role,google_subject)
 VALUES($1,$2,'',$3,'regular',$4)
 ON CONFLICT(google_subject) DO UPDATE SET google_subject=EXCLUDED.google_subject
 RETURNING email`, strings.ToLower(email), name, string(hash), subject).Scan(&storedEmail)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return nil, ErrGoogleEmailExists
	}
	if err != nil {
		return nil, err
	}
	return r.GetByEmail(ctx, storedEmail)
}
