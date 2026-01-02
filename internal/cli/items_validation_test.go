package cli

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestCLI_CreateCredentialMissingPassword(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--config", filepath.Join(dir, "cfg.json"),
		"--server", "https://localhost",
		"--ca-cert-path=",
		"create", "--title", "cred", "--type", "CREDENTIAL", "--login", "l",
	})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestCLI_CreateBinaryInvalidUUID(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--config", filepath.Join(dir, "cfg.json"),
		"--server", "https://localhost",
		"--ca-cert-path=",
		"create", "--title", "bin", "--type", "BINARY", "--filename", "a.bin", "--binary-id", "bad-uuid",
	})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestCLI_CreateCardMissingFields(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--config", filepath.Join(dir, "cfg.json"),
		"--server", "https://localhost",
		"--ca-cert-path=",
		"create", "--title", "card", "--type", "CARD", "--card-number", "4111111111111111",
	})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected error for missing card fields")
	}
}

func TestCLI_ListInvalidParamsHandled(t *testing.T) {
	dir := t.TempDir()
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"total":0,"limit":0,"offset":0}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--config", filepath.Join(dir, "cfg.json"),
		"--server", srv.URL,
		"--ca-cert-path=",
		"list", "--type", "BAD", "--limit", "-1", "--offset", "-1",
	})
	err := cmd.ExecuteContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestCLI_CreateMissingTitle(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--config", filepath.Join(dir, "cfg.json"),
		"--server", "http://localhost",
		"--ca-cert-path=",
		"create", "--type", "TEXT", "--value", "v",
	})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestCLI_UpdateUnsupportedType(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--config", filepath.Join(dir, "cfg.json"),
		"--server", "https://localhost",
		"--ca-cert-path=",
		"update", "00000000-0000-0000-0000-000000000001", "--type", "UNKNOWN",
	})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
