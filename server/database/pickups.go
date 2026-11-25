package database

import (
	"context"
	_ "embed"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

//go:embed migrations/20261109000000_add_pickup_reservations.sql
var pickupMigration string

//go:embed migrations/20261111000000_add_pickup_email_reminders.sql
var pickupEmailMigration string

//go:embed migrations/20261125000000_add_pharmacist_alert_emails.sql
var pharmacistEmailMigration string

// EnsurePickupSchema also supports local installations without a migration runner.
func EnsurePickupSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, strings.Split(pickupMigration, "-- +goose Down")[0]+"\n"+strings.Split(pickupEmailMigration, "-- +goose Down")[0]+"\n"+strings.Split(pharmacistEmailMigration, "-- +goose Down")[0])
	return err
}
