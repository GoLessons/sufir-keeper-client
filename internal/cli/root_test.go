package cli

import (
	"context"
	"testing"

	"github.com/GoLessons/sufir-keeper-server/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmdConfigFlow(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--config", "/workspace/var/config.yaml",
		"--server", "https://localhost:8443/api/v1",
		"--log-level", "info",
		"--ca-cert-path", "/workspace/var/ca.crt",
	})
	err := cmd.ExecuteContext(context.Background())
	require.NoError(t, err)
	val := cmd.Context().Value(cfgContextKey)
	require.NotNil(t, val)
	cfg, ok := val.(config.Config)
	require.True(t, ok)
	require.Equal(t, "https://localhost:8443/api/v1", cfg.Server.BaseURL)
	require.Equal(t, "/workspace/var/ca.crt", cfg.TLS.CACertPath)
	require.Equal(t, "info", cfg.Log.Level)
	logVal := cmd.Context().Value(logContextKey)
	require.NotNil(t, logVal)
}
