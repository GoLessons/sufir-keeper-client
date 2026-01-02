package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestCLI_CreateTypes_Credential_Card_Binary(t *testing.T) {
	dir := t.TempDir()
	srv := newMockServer(t)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")

	// login first
	{
		var buf bytes.Buffer
		cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--config", filepath.Join(dir, "cfg.json"), "--server", srv.URL, "--ca-cert-path=", "login", "--login", "u", "--password", "p"})
		if err := cmd.ExecuteContext(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	// CREDENTIAL
	{
		var buf bytes.Buffer
		cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--config", filepath.Join(dir, "cfg.json"), "--server", srv.URL, "--ca-cert-path=", "create", "--title", "cred", "--type", "CREDENTIAL", "--login", "l", "--password", "p"})
		if err := cmd.ExecuteContext(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	// CARD
	{
		var buf bytes.Buffer
		cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{
			"--config", filepath.Join(dir, "cfg.json"),
			"--server", srv.URL, "--ca-cert-path=",
			"create", "--title", "card", "--type", "CARD",
			"--card-number", "4111111111111111", "--card-holder", "IVAN IVANOV", "--expiry-date", "12/25", "--cvv", "123",
		})
		if err := cmd.ExecuteContext(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	// BINARY
	{
		var buf bytes.Buffer
		cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{
			"--config", filepath.Join(dir, "cfg.json"),
			"--server", srv.URL, "--ca-cert-path=",
			"create", "--title", "bin", "--type", "BINARY",
			"--filename", "a.bin", "--binary-id", "00000000-0000-0000-0000-000000000001",
		})
		if err := cmd.ExecuteContext(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCLI_UpdateTypes_Credential_Card_Binary(t *testing.T) {
	dir := t.TempDir()
	srv := newMockServer(t)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")

	id := "00000000-0000-0000-0000-000000000001"

	// CREDENTIAL update password
	{
		var buf bytes.Buffer
		cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--config", filepath.Join(dir, "cfg.json"), "--server", srv.URL, "--ca-cert-path=", "update", id, "--type", "CREDENTIAL", "--password", "np"})
		if err := cmd.ExecuteContext(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	// CARD update holder
	{
		var buf bytes.Buffer
		cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--config", filepath.Join(dir, "cfg.json"), "--server", srv.URL, "--ca-cert-path=", "update", id, "--type", "CARD", "--card-holder", "IVAN PETROV"})
		if err := cmd.ExecuteContext(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	// BINARY update filename
	{
		var buf bytes.Buffer
		cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"--config", filepath.Join(dir, "cfg.json"), "--server", srv.URL, "--ca-cert-path=", "update", id, "--type", "BINARY", "--filename", "b.bin"})
		if err := cmd.ExecuteContext(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}
