package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func backup(ctx context.Context, executable, dbURL, path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("backup path already exists or is inaccessible")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	config, err := pgx.ParseConfig(dbURL)
	if err != nil {
		return fmt.Errorf("invalid database configuration")
	}
	command := exec.CommandContext(ctx, executable, "--format=custom", "--no-owner", "--no-acl", "--no-password", "--file", path)
	ssl := "disable"
	if config.TLSConfig != nil {
		ssl = "require"
	}
	command.Env = append(os.Environ(), "PGHOST="+config.Host, "PGPORT="+strconv.Itoa(int(config.Port)), "PGDATABASE="+config.Database, "PGUSER="+config.User, "PGPASSWORD="+config.Password, "PGSSLMODE="+ssl, "PGCONNECT_TIMEOUT=10")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err = command.Run(); err != nil {
		return fmt.Errorf("backup failed before writes: %s", strings.TrimSpace(stderr.String()))
	}
	restore := "pg_restore" + filepath.Ext(executable)
	if executable != filepath.Base(executable) {
		restore = filepath.Join(filepath.Dir(executable), restore)
	}
	if err = exec.CommandContext(ctx, restore, "--list", path).Run(); err != nil {
		return fmt.Errorf("backup verification failed")
	}
	return nil
}
