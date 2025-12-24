package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/spf13/viper"
)

func TestLoadNilOut(t *testing.T) {
	v := viper.New()
	err := Load(v, nil)
	require.Error(t, err)
}

func TestAuthEnvFallback(t *testing.T) {
	t.Setenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE", "sufir-keeper-client")
	t.Setenv("SUFIR_KEEPER_AUTH_BACKEND", "file")
	t.Setenv("SUFIR_KEEPER_AUTH_FILE_DIR", "/tmp/keyring")
	t.Setenv("SUFIR_KEEPER_SERVER", "https://localhost:8443/api/v1")
	t.Setenv("SUFIR_KEEPER_CA_CERT", "./var/ca.crt")
	t.Setenv("SUFIR_KEEPER_LOG_LEVEL", "info")
	v := viper.New()
	var cfg Config
	err := Load(v, &cfg)
	require.NoError(t, err)
	require.Equal(t, "sufir-keeper-client", cfg.Auth.TokenStoreService)
	require.Equal(t, "file", cfg.Auth.Backend)
	require.Equal(t, "/tmp/keyring", cfg.Auth.FileDir)
	require.Equal(t, os.Getenv("SUFIR_KEEPER_SERVER"), cfg.Server.BaseURL)
	require.Equal(t, os.Getenv("SUFIR_KEEPER_CA_CERT"), cfg.TLS.CACertPath)
	require.Equal(t, os.Getenv("SUFIR_KEEPER_LOG_LEVEL"), cfg.Log.Level)
}

