package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvKeyReplacer(t *testing.T) {
	r := EnvKeyReplacer()
	out := r.Replace("tls.ca_cert_path")
	assert.Equal(t, "tls_ca_cert_path", out)
}

func TestLoadWithExplicitValues(t *testing.T) {
	v := viper.New()
	v.Set("config.file", "/workspace/var/config.yaml")
	v.Set("server.base_url", "https://localhost:8443/api/v1")
	v.Set("tls.ca_cert_path", "/workspace/var/ca.crt")
	v.Set("log.level", "debug")
	var cfg Config
	err := Load(v, &cfg)
	require.NoError(t, err)
	assert.Equal(t, "/workspace/var/config.yaml", cfg.ConfigFile)
	assert.Equal(t, "https://localhost:8443/api/v1", cfg.Server.BaseURL)
	assert.Equal(t, "/workspace/var/ca.crt", cfg.TLS.CACertPath)
	assert.Equal(t, "debug", cfg.Log.Level)
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("SUFIR_KEEPER_CONFIG", "/workspace/var/envconfig.yaml")
	t.Setenv("SUFIR_KEEPER_SERVER", "https://s.example/api/v1")
	t.Setenv("SUFIR_KEEPER_CA_CERT", "/workspace/var/envca.crt")
	t.Setenv("SUFIR_KEEPER_LOG_LEVEL", "info")
	v := viper.New()
	v.SetEnvPrefix("SUFIR_KEEPER")
	v.SetEnvKeyReplacer(EnvKeyReplacer())
	v.AutomaticEnv()
	var cfg Config
	err := Load(v, &cfg)
	require.NoError(t, err)
	assert.Equal(t, "/workspace/var/envconfig.yaml", cfg.ConfigFile)
	assert.Equal(t, "https://s.example/api/v1", cfg.Server.BaseURL)
	assert.Equal(t, "/workspace/var/envca.crt", cfg.TLS.CACertPath)
	assert.Equal(t, "info", cfg.Log.Level)
}
