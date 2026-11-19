package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// ReminderMailSender is injectable so tests never contact real inboxes.
type ReminderMailSender interface {
	Send(context.Context, string, string, string, string) error
}
type SMTPReminderSender struct{ host, port, user, password, from string }

func newSMTPReminderSender() (*SMTPReminderSender, error) {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	if host == "" {
		host = "smtp.gmail.com"
	}
	if port == "" {
		port = "587"
	}
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	password := os.Getenv("SMTP_PASSWORD")
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if from == "" {
		from = user
	}
	if user == "" || password == "" {
		return nil, errors.New("set SMTP_USER and SMTP_PASSWORD in server/.env")
	}
	if _, err := plainEmailAddress(from); err != nil {
		return nil, errors.New("SMTP_FROM must be an email address")
	}
	if port != "587" && port != "465" {
		return nil, errors.New("SMTP_PORT must be 587 (STARTTLS) or 465 (TLS)")
	}
	return &SMTPReminderSender{host: host, port: port, user: user, password: password, from: from}, nil
}
func plainEmailAddress(value string) (string, error) {
	if strings.ContainsAny(value, "\r\n") {
		return "", errors.New("invalid email address")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", errors.New("invalid email address")
	}
	return address.Address, nil
}
func reminderMessage(from, to, id, subject, body string) ([]byte, error) {
	if _, err := plainEmailAddress(from); err != nil {
		return nil, err
	}
	if _, err := plainEmailAddress(to); err != nil {
		return nil, err
	}
	if strings.ContainsAny(id+subject, "\r\n") {
		return nil, errors.New("invalid message headers")
	}
	var message bytes.Buffer
	fmt.Fprintf(&message, "From: Qayen <%s>\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: <%s>\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", from, to, subject, time.Now().Format(time.RFC1123Z), id)
	writer := quotedprintable.NewWriter(&message)
	if _, err := writer.Write([]byte(body)); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return message.Bytes(), nil
}
func (s *SMTPReminderSender) Send(ctx context.Context, to, id, subject, body string) error {
	message, err := reminderMessage(s.from, to, id, subject, body)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(10 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	dialer := &net.Dialer{Deadline: deadline}
	address := net.JoinHostPort(s.host, s.port)
	var conn net.Conn
	tlsConfig := &tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12}
	if s.port == "465" {
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(ctx, "tcp", address)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return errors.New("SMTP connection failed")
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	if err = conn.SetDeadline(deadline); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return errors.New("SMTP greeting failed")
	}
	defer client.Close()
	if s.port == "587" {
		if err = client.StartTLS(tlsConfig); err != nil {
			return errors.New("SMTP TLS failed")
		}
	}
	if err = client.Auth(smtp.PlainAuth("", s.user, s.password, s.host)); err != nil {
		return errors.New("SMTP authentication failed; check the sender App Password")
	}
	if err = client.Mail(s.from); err != nil {
		return errors.New("SMTP sender rejected")
	}
	if err = client.Rcpt(to); err != nil {
		return errors.New("SMTP recipient rejected")
	}
	data, err := client.Data()
	if err != nil {
		return errors.New("SMTP message rejected")
	}
	if _, err = data.Write(message); err != nil {
		data.Close()
		return errors.New("SMTP message transmission failed")
	}
	if err = data.Close(); err != nil {
		return errors.New("SMTP acceptance not confirmed")
	}
	// DATA acceptance is success. A failed QUIT must not trigger a duplicate email.
	_ = client.Quit()
	return nil
}
