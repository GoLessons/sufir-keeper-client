package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCLI_ItemsList_And_FilesDownload(t *testing.T) {
	dir := t.TempDir()
	srv := newMockServer(t)
	defer srv.Close()

	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "list"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if out == "" {
		t.Fatal("empty output")
	}

	buf.Reset()
	outPath := filepath.Join(dir, "out.bin")
	cmd = NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "download", "00000000-0000-0000-0000-000000000001", outPath})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatal(err)
	}
}

func TestCLI_ItemsListWithFlags(t *testing.T) {
	dir := t.TempDir()
	srv := newMockServer(t)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--config", cfgPath,
		"--server", srv.URL, "--ca-cert-path=",
		"list", "--type", "TEXT", "--search", "t", "--limit", "10", "--offset", "1",
	})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if out == "" {
		t.Fatal("empty output")
	}
}

func TestCLI_ItemsCreateUpdateDelete(t *testing.T) {
	dir := t.TempDir()
	srv := newMockServer(t)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "login", "--login", "u", "--password", "p"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	cmd = NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "create", "--title", "t", "--value", "v"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	cmd = NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "update", "00000000-0000-0000-0000-000000000001", "--title", "t2"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	cmd = NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "delete", "00000000-0000-0000-0000-000000000001"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestCLI_ItemsCreateUnauthorized(t *testing.T) {
	dir := t.TempDir()
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":401}`))
			return
		}
	})
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "create", "--title", "t", "--type", "TEXT", "--value", "v"})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestCLI_ItemsUpdateUnauthorized(t *testing.T) {
	dir := t.TempDir()
	mux := http.NewServeMux()
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":401}`))
			return
		}
	})
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "update", "00000000-0000-0000-0000-000000000001", "--type", "TEXT", "--value", "v"})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestCLI_ItemsDeleteUnauthorized(t *testing.T) {
	dir := t.TempDir()
	mux := http.NewServeMux()
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"code":401}`))
			return
		}
	})
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "delete", "00000000-0000-0000-0000-000000000001"})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestCLI_UpdateTextWithMeta(t *testing.T) {
	dir := t.TempDir()
	srv := newMockServer(t)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "update", "00000000-0000-0000-0000-000000000001", "--type", "TEXT", "--value", "v", "--meta", "k=v"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestCLI_ItemsGet(t *testing.T) {
	dir := t.TempDir()
	srv := newMockServer(t)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "get", "00000000-0000-0000-0000-000000000001"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if buf.String() == "" {
		t.Fatal("empty get output")
	}
}

func TestCLI_ItemsGetUnauthorized(t *testing.T) {
	dir := t.TempDir()
	mux := http.NewServeMux()
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401}`))
	})
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "get", "00000000-0000-0000-0000-000000000001"})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestCLI_FilesUpload(t *testing.T) {
	dir := t.TempDir()
	srv := newMockServer(t)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	fp := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(fp, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "upload", "--path", fp})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestCLI_DownloadUnauthorized(t *testing.T) {
	dir := t.TempDir()
	mux := http.NewServeMux()
	mux.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401}`))
	})
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "download", "00000000-0000-0000-0000-000000000001", filepath.Join(dir, "out.bin")})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestCLI_UploadError(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(fp, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/files/presign", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			FormFields map[string]string `json:"form_fields,omitempty"`
			Key        string            `json:"key,omitempty"`
			UploadURL  string            `json:"upload_url,omitempty"`
		}{
			FormFields: map[string]string{},
			Key:        "k",
			UploadURL:  "http://" + r.Host + "/upload",
		})
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "upload", "--path", fp})
	err := cmd.ExecuteContext(context.Background())
	if err == nil {
		t.Fatal("expected upload error")
	}
}

func TestCLI_FilesPresign(t *testing.T) {
	// presign выполняется прозрачно внутри upload; отдельная команда не нужна
}

func newMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		type item struct {
			Title string `json:"title"`
		}
		type list struct {
			Items  []item `json:"items"`
			Limit  int    `json:"limit"`
			Offset int    `json:"offset"`
			Total  int    `json:"total"`
		}
		resp := list{Items: []item{{Title: "one"}}, Limit: 1, Offset: 0, Total: 1}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"title":"x"}`))
			return
		}
		if r.Method == http.MethodPut {
			_, _ = w.Write([]byte(`{"title":"updated"}`))
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	})
	mux.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte("data"))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/files/presign", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			FormFields map[string]string `json:"form_fields,omitempty"`
			Key        string            `json:"key,omitempty"`
			UploadURL  string            `json:"upload_url,omitempty"`
		}{
			FormFields: map[string]string{"x-meta": "v"},
			Key:        "k",
			UploadURL:  "http://" + r.Host + "/upload",
		})
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"access_token":"a","refresh_token":"r","token_type":"bearer","expires_in":3600}`))
			return
		}
		if r.Method == http.MethodPatch {
			_, _ = w.Write([]byte(`{"access_token":"a2","refresh_token":"r2","token_type":"bearer","expires_in":3600}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/auth-verify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-User-Id", "user-1")
		w.WriteHeader(http.StatusNoContent)
	})
	return httptest.NewServer(mux)
}

func TestCLI_AuthLoginStatusLogout(t *testing.T) {
	dir := t.TempDir()
	srv := newMockServer(t)
	defer srv.Close()
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", dir)
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	cfgPath := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(cfgPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	cmd := NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "login", "--login", "u", "--password", "p"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	cmd = NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "status"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if buf.String() == "" {
		t.Fatal("empty status")
	}
	buf.Reset()
	cmd = NewRootCmd("dev", "none", time.Now().Format(time.RFC3339))
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", cfgPath, "--server", srv.URL, "--ca-cert-path=", "logout"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
}
