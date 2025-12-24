package cli

import (
	"context"
	"testing"

	"github.com/GoLessons/sufir-keeper-server/internal/config"
)

func TestNewRootCmdConfigFlow(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--config", "/workspace/var/config.yaml",
		"--server", "https://localhost:8443/api/v1",
		"--log-level", "info",
		"--ca-cert-path", "/workspace/var/ca.crt",
	})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("execute error: %v", err)
	}
	val := cmd.Context().Value(cfgContextKey)
	if val == nil {
		t.Fatalf("missing config in context")
	}
	cfg, ok := val.(config.Config)
	if !ok {
		t.Fatalf("unexpected config type")
	}
	if cfg.Server.BaseURL != "https://localhost:8443/api/v1" {
		t.Fatalf("unexpected base url: %s", cfg.Server.BaseURL)
	}
	if cfg.TLS.CACertPath != "/workspace/var/ca.crt" {
		t.Fatalf("unexpected ca cert path: %s", cfg.TLS.CACertPath)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("unexpected log level: %s", cfg.Log.Level)
	}
	logVal := cmd.Context().Value(logContextKey)
	if logVal == nil {
		t.Fatalf("missing logger in context")
	}
}
