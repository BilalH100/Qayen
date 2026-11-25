package services

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPharmacistAlertEmailBody(t *testing.T) {
	t.Setenv("QAYEN_PUBLIC_URL", "https://qayen.example/some-path?discard=true")
	body := pharmacistAlertEmailBody(7, 10, "Test pharmacy", "Doliprane", "500mg", "Tablet", "Paracetamol", 3, "Test patient", "+212600000000", time.Now(), time.Now().Add(30*time.Second))
	for _, text := range []string{"Medication: Doliprane", "Dosage: 500mg", "Active ingredient: Paracetamol", "Requested quantity: 3", "Patient: Test patient", "Patient phone: +212600000000", "https://qayen.example/pharmacy", "Response window ends (Morocco)"} {
		if !strings.Contains(body, text) {
			t.Errorf("missing %q", text)
		}
	}
	if strings.Contains(body, "discard=true") {
		t.Fatal("dashboard link retained arbitrary query")
	}
	t.Setenv("QAYEN_PUBLIC_URL", "javascript:alert(1)")
	body = pharmacistAlertEmailBody(7, 10, "Test pharmacy", "Medicine", "", "", "", 1, "Patient", "", time.Now(), time.Now())
	if !strings.Contains(body, "Patient phone: Not provided") || strings.Contains(body, "javascript:") {
		t.Fatal("invalid link or empty contact handling")
	}
}

func TestPharmacistAlertEmailDelivery(t *testing.T) {
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
	cfg, err := pgxpool.ParseConfig(env["DB_URL"])
	if err != nil {
		t.Fatal("configuration invalid")
	}
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal("database unavailable")
	}
	defer pool.Close()
	// Temporary objects and sequences shadow the live tables. No real emails.
	_, err = pool.Exec(ctx, `CREATE TEMP TABLE users(id integer PRIMARY KEY,email text,name text,phone text,user_role text,managed_pharmacy_id integer);
 CREATE TEMP TABLE pharmacies(id integer PRIMARY KEY,name text);
 CREATE TEMP TABLE medications(id integer PRIMARY KEY,speciality text,dosage text,form text,active_substance text);
 CREATE TEMP TABLE medication_alerts(id integer PRIMARY KEY,customer_id integer,medication_id integer,status text,created_at timestamptz,expires_at timestamptz,requested_quantity integer);
 CREATE TEMP TABLE pharmacy_alert_notifications(id integer PRIMARY KEY,alert_id integer,pharmacy_id integer,sent_at timestamptz);
 CREATE TEMP TABLE pharmacist_responses(alert_id integer,pharmacy_id integer);
 INSERT INTO pharmacies VALUES(10,'Target pharmacy'),(11,'Other pharmacy'),(12,'Unlinked pharmacy');
 INSERT INTO medications VALUES(100,'Doliprane','500mg','Tablet','Paracetamol');
 INSERT INTO users VALUES(1,'patient@example.com','Test patient','+212600000000','regular',NULL),
 (2,'pharmacist@example.com','Target pharmacist','','pharmacist',10),
 (3,'other@example.com','Other pharmacist','','pharmacist',11),
 (4,'admin@example.com','Admin','','admin',10),(5,'regular@example.com','Regular','','regular',10);`)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../database/migrations/20261125000000_add_pharmacist_alert_emails.sql")
	if err != nil {
		t.Fatal(err)
	}
	schema := strings.ReplaceAll(strings.Split(string(migration), "-- +goose Down")[0], "CREATE TABLE IF NOT EXISTS", "CREATE TEMP TABLE")
	if _, err = pool.Exec(ctx, schema); err != nil {
		t.Fatal(err)
	}
	service := NewAlertService(pool, nil)
	reset := func() {
		t.Helper()
		_, e := pool.Exec(ctx, `TRUNCATE pharmacist_alert_emails RESTART IDENTITY;DELETE FROM pharmacist_responses;DELETE FROM pharmacy_alert_notifications;DELETE FROM medication_alerts;
 UPDATE users SET email='pharmacist@example.com',managed_pharmacy_id=10 WHERE id=2;
 INSERT INTO medication_alerts VALUES(100,1,100,'pending',now(),now()+interval '2 minutes',3);
 INSERT INTO pharmacy_alert_notifications VALUES(100,100,10,now());`)
		if e != nil {
			t.Fatal(e)
		}
	}
	process := func(sender ReminderMailSender) {
		t.Helper()
		if e := service.ProcessPharmacistAlertEmails(ctx, sender); e != nil {
			t.Fatal(e)
		}
	}
	queue := func() {
		t.Helper()
		if _, e := pool.Exec(ctx, `INSERT INTO pharmacist_alert_emails(notification_id,user_id) VALUES(100,2)`); e != nil {
			t.Fatal(e)
		}
	}
	t.Run("target recipient and complete request without duplication", func(t *testing.T) {
		reset()
		sender := &fakeReminderSender{}
		process(sender)
		process(sender)
		if len(sender.messages) != 1 || sender.messages[0].to != "pharmacist@example.com" {
			t.Fatal("wrong pharmacy/role, missing email, or repeated delivery")
		}
		for _, text := range []string{"Medication: Doliprane", "Requested quantity: 3", "Patient: Test patient", "Patient phone: +212600000000"} {
			if !strings.Contains(sender.messages[0].body, text) {
				t.Fatal("request details missing")
			}
		}
		if !strings.Contains(sender.messages[0].id, "pharmacy-alert-100-user-2") {
			t.Fatal("message identity missing")
		}
	})
	t.Run("retry SMTP failure with backoff", func(t *testing.T) {
		reset()
		sender := &fakeReminderSender{fail: true}
		process(sender)
		process(sender)
		if sender.calls != 1 {
			t.Fatal("backoff ignored")
		}
		if _, err = pool.Exec(ctx, `UPDATE pharmacist_alert_emails SET next_attempt_at=now()-interval '1 second'`); err != nil {
			t.Fatal(err)
		}
		sender.fail = false
		process(sender)
		process(sender)
		if len(sender.messages) != 1 {
			t.Fatal("retry missing or duplicated")
		}
	})
	for _, condition := range []string{"completed", "cancelled", "expired", "answered", "timed_out", "unlinked", "invalid_recipient"} {
		t.Run("skip "+condition, func(t *testing.T) {
			reset()
			queue()
			switch condition {
			case "completed", "cancelled":
				_, err = pool.Exec(ctx, `UPDATE medication_alerts SET status=$1`, condition)
			case "expired":
				_, err = pool.Exec(ctx, `UPDATE medication_alerts SET expires_at=now()-interval '1 second'`)
			case "answered":
				_, err = pool.Exec(ctx, `INSERT INTO pharmacist_responses VALUES(100,10)`)
			case "timed_out":
				_, err = pool.Exec(ctx, `UPDATE pharmacy_alert_notifications SET sent_at=now()-interval '31 seconds'`)
			case "unlinked":
				_, err = pool.Exec(ctx, `UPDATE users SET managed_pharmacy_id=NULL WHERE id=2`)
			case "invalid_recipient":
				_, err = pool.Exec(ctx, `UPDATE users SET email='invalid' WHERE id=2`)
			}
			if err != nil {
				t.Fatal(err)
			}
			sender := &fakeReminderSender{}
			process(sender)
			if sender.calls != 0 {
				t.Fatal("ineligible pharmacy email sent")
			}
			var status string
			if err = pool.QueryRow(ctx, `SELECT status FROM pharmacist_alert_emails WHERE notification_id=100 AND user_id=2`).Scan(&status); err != nil || status != "skipped" {
				t.Fatal("ineligible job not skipped")
			}
		})
	}
	t.Run("escalation goes only to next selected pharmacy", func(t *testing.T) {
		reset()
		queue()
		if _, err = pool.Exec(ctx, `INSERT INTO pharmacy_alert_notifications VALUES(101,100,11,now())`); err != nil {
			t.Fatal(err)
		}
		sender := &fakeReminderSender{}
		process(sender)
		process(sender)
		if len(sender.messages) != 1 || sender.messages[0].to != "other@example.com" {
			t.Fatal("escalation emailed the wrong pharmacy")
		}
	})
	t.Run("no account means no recipient", func(t *testing.T) {
		reset()
		if _, err = pool.Exec(ctx, `UPDATE pharmacy_alert_notifications SET pharmacy_id=12`); err != nil {
			t.Fatal(err)
		}
		sender := &fakeReminderSender{}
		process(sender)
		if sender.calls != 0 {
			t.Fatal("unlinked catalog pharmacy emailed")
		}
	})
	t.Run("concurrent passes preserve one delivery", func(t *testing.T) {
		reset()
		sender := &fakeReminderSender{}
		var wg sync.WaitGroup
		errs := make(chan error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); errs <- service.ProcessPharmacistAlertEmails(ctx, sender) }()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(sender.messages) != 1 {
			t.Fatal("duplicate delivery")
		}
	})
}
