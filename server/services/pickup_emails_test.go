package services

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeReminderMail struct{ to, id, body string }
type fakeReminderSender struct {
	mu       sync.Mutex
	fail     bool
	messages []fakeReminderMail
	calls    int
}

func (f *fakeReminderSender) Send(_ context.Context, to, id, subject, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.fail {
		return errors.New("simulated SMTP failure")
	}
	f.messages = append(f.messages, fakeReminderMail{to, id, body})
	return nil
}
func TestReminderSMTPConfigAndMessage(t *testing.T) {
	for _, key := range []string{"SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_PASSWORD", "SMTP_FROM"} {
		t.Setenv(key, "")
	}
	if _, err := newSMTPReminderSender(); err == nil {
		t.Fatal("missing sender credentials accepted")
	}
	t.Setenv("SMTP_USER", "sender@example.com")
	t.Setenv("SMTP_PASSWORD", "test-app-password")
	sender, err := newSMTPReminderSender()
	if err != nil || sender.host != "smtp.gmail.com" || sender.port != "587" {
		t.Fatal("Gmail defaults unavailable")
	}
	t.Setenv("SMTP_PORT", "25")
	if _, err = newSMTPReminderSender(); err == nil {
		t.Fatal("plaintext SMTP accepted")
	}
	for _, address := range []string{"bad", "victim@example.com\r\nBcc: other@example.com"} {
		if _, err = reminderMessage("sender@example.com", address, "pickup-1@qayen", "Reminder", "body"); err == nil {
			t.Fatal("invalid recipient accepted")
		}
	}
	if _, err = reminderMessage("sender@example.com", "patient@example.com", "id@qayen", "Reminder\r\nBcc: victim@example.com", "body"); err == nil {
		t.Fatal("header injection accepted")
	}
	body := pickupEmailBody("Patient", "Pharmacy", "Address", time.Now().Add(time.Minute), 2)
	if !strings.Contains(body, "about to expire") || !strings.Contains(body, "Pickup deadline (Morocco)") {
		t.Fatal("final reminder content missing")
	}
	message, err := reminderMessage("sender@example.com", "patient@example.com", "pickup-1-reminder-2@qayen", "Reminder", body)
	if err != nil || !strings.Contains(string(message), "Message-ID: <pickup-1-reminder-2@qayen>") {
		t.Fatal("message identity missing")
	}
}
func TestPickupEmailDelivery(t *testing.T) {
	path := os.Getenv("EMAIL_TEST_DB_ENV")
	if path == "" {
		t.Skip("set EMAIL_TEST_DB_ENV for isolated PostgreSQL tests")
	}
	env, err := godotenv.Read(path)
	if err != nil {
		t.Fatal("configuration unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(env["DB_URL"])
	if err != nil {
		t.Fatal("configuration invalid")
	}
	config.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal("database unavailable")
	}
	defer pool.Close()
	// Every table and sequence is temporary: no patient/catalog data is touched.
	_, err = pool.Exec(ctx, `CREATE TEMP TABLE users(id integer PRIMARY KEY,email text,name text);
 CREATE TEMP TABLE pharmacies(id integer PRIMARY KEY,name text,address text);
 CREATE TEMP TABLE pickup_reservations(id integer PRIMARY KEY,customer_id integer,pharmacy_id integer,status text,created_at timestamptz,expires_at timestamptz);
 INSERT INTO users VALUES(1,'patient@example.com','Patient');INSERT INTO pharmacies VALUES(1,'Pharmacy','Address');`)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../database/migrations/20261111000000_add_pickup_email_reminders.sql")
	if err != nil {
		t.Fatal(err)
	}
	schema := strings.Split(string(migration), "-- +goose Down")[0]
	schema = strings.ReplaceAll(schema, "CREATE TABLE IF NOT EXISTS", "CREATE TEMP TABLE")
	if _, err = pool.Exec(ctx, schema); err != nil {
		t.Fatal(err)
	}
	service := NewAlertService(pool, nil)
	reset := func() {
		t.Helper()
		if _, e := pool.Exec(ctx, `TRUNCATE pickup_email_reminders RESTART IDENTITY;DELETE FROM pickup_reservations;INSERT INTO pickup_reservations VALUES(1,1,1,'awaiting',now()-interval '181 seconds',now()+interval '119 seconds')`); e != nil {
			t.Fatal(e)
		}
	}
	process := func(sender ReminderMailSender) {
		t.Helper()
		if e := service.ProcessPickupEmails(ctx, sender); e != nil {
			t.Fatal(e)
		}
	}
	t.Run("both scheduled reminders and no repeat", func(t *testing.T) {
		reset()
		sender := &fakeReminderSender{}
		process(sender)
		process(sender)
		if len(sender.messages) != 1 || sender.messages[0].to != "patient@example.com" || !strings.Contains(sender.messages[0].id, "reminder-1") {
			t.Fatal("first reminder missing, duplicated, or wrong recipient")
		}
		_, err = pool.Exec(ctx, `UPDATE pickup_reservations SET created_at=now()-interval '241 seconds',expires_at=now()+interval '59 seconds';UPDATE pickup_email_reminders SET scheduled_at=now()-interval '1 second' WHERE stage=2`)
		if err != nil {
			t.Fatal(err)
		}
		process(sender)
		process(sender)
		if len(sender.messages) != 2 || !strings.Contains(sender.messages[1].id, "reminder-2") {
			t.Fatal("second reminder missing or duplicated")
		}
	})
	t.Run("retry failed delivery", func(t *testing.T) {
		reset()
		sender := &fakeReminderSender{fail: true}
		process(sender)
		process(sender)
		if sender.calls != 1 || len(sender.messages) != 0 {
			t.Fatal("retry backoff not respected")
		}
		if _, err = pool.Exec(ctx, `UPDATE pickup_email_reminders SET next_attempt_at=now()-interval '1 second'`); err != nil {
			t.Fatal(err)
		}
		sender.fail = false
		process(sender)
		if len(sender.messages) != 1 {
			t.Fatal("failed email not retried")
		}
	})
	for _, status := range []string{"picked_up", "canceled", "expired"} {
		t.Run("skip "+status, func(t *testing.T) {
			reset()
			if _, err = pool.Exec(ctx, `INSERT INTO pickup_email_reminders(reservation_id,stage,scheduled_at) VALUES(1,1,now()-interval '1 second')`); err != nil {
				t.Fatal(err)
			}
			if status == "expired" {
				_, err = pool.Exec(ctx, `UPDATE pickup_reservations SET expires_at=now()-interval '1 second'`)
			} else {
				_, err = pool.Exec(ctx, `UPDATE pickup_reservations SET status=$1`, status)
			}
			if err != nil {
				t.Fatal(err)
			}
			sender := &fakeReminderSender{}
			process(sender)
			if sender.calls != 0 {
				t.Fatal("obsolete reminder sent")
			}
			var state string
			if err = pool.QueryRow(ctx, `SELECT status FROM pickup_email_reminders WHERE stage=1`).Scan(&state); err != nil || state != "skipped" {
				t.Fatal("obsolete email not skipped")
			}
		})
	}
	t.Run("restart catches up only relevant reminder", func(t *testing.T) {
		reset()
		if _, err = pool.Exec(ctx, `UPDATE pickup_reservations SET created_at=now()-interval '241 seconds',expires_at=now()+interval '59 seconds'`); err != nil {
			t.Fatal(err)
		}
		sender := &fakeReminderSender{}
		process(sender)
		process(sender)
		if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].id, "reminder-2") {
			t.Fatal("stale reminders sent after downtime")
		}
	})
	t.Run("concurrent workers do not repeat delivery", func(t *testing.T) {
		reset()
		sender := &fakeReminderSender{}
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); errs <- service.ProcessPickupEmails(ctx, sender) }()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(sender.messages) != 1 {
			t.Fatal("concurrent delivery duplicated")
		}
	})
}
