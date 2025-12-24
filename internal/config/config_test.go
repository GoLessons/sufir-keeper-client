package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestEnvKeyReplacer(t *testing.T) {
	r := EnvKeyReplacer()
	out := r.Replace("tls.ca_cert_path")
	if out != "tls_ca_cert_path" {
		t.Fatalf("expected tls_ca_cert_path, got %s", out)
	}
}

func TestLoadWithExplicitValues(t *testing.T) {
	v := viper.New()
	v.Set("config.file", "/workspace/var/config.yaml")
	v.Set("server.base_url", "https://localhost:8443/api/v1")
	v.Set("tls.ca_cert_path", "/workspace/var/ca.crt")
	v.Set("log.level", "debug")
	var cfg Config
	if err := Load(v, &cfg); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if cfg.ConfigFile != "/workspace/var/config.yaml" {
		t.Fatalf("unexpected config file: %s", cfg.ConfigFile)
	}
	if cfg.Server.BaseURL != "https://localhost:8443/api/v1" {
		t.Fatalf("unexpected base url: %s", cfg.Server.BaseURL)
	}
	if cfg.TLS.CACertPath != "/workspace/var/ca.crt" {
		t.Fatalf("unexpected ca cert path: %s", cfg.TLS.CACertPath)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("unexpected log level: %s", cfg.Log.Level)
	}
}

func TestLoadFromEnv(t *testing.T) {
	_ = os.Setenv("SUFIR_KEEPER_CONFIG", "/workspace/var/envconfig.yaml")
	_ = os.Setenv("SUFIR_KEEPER_SERVER", "https://s.example/api/v1")
	_ = os.Setenv("SUFIR_KEEPER_CA_CERT", "/workspace/var/envca.crt")
	_ = os.Setenv("SUFIR_KEEPER_LOG_LEVEL", "info")
	t.Cleanup(func() {
		_ = os.Unsetenv("SUFIR_KEEPER_CONFIG")
		_ = os.Unsetenv("SUFIR_KEEPER_SERVER")
		_ = os.Unsetenv("SUFIR_KEEPER_CA_CERT")
		_ = os.Unsetenv("SUFIR_KEEPER_LOG_LEVEL")
	})
	v := viper.New()
	v.SetEnvPrefix("SUFIR_KEEPER")
	v.SetEnvKeyReplacer(EnvKeyReplacer())
	v.AutomaticEnv()
	var cfg Config
	if err := Load(v, &cfg); err != nil {
		t.Fatalf("load error: %v", err)
	}
	if cfg.ConfigFile != "/workspace/var/envconfig.yaml" {
		t.Fatalf("unexpected config file: %s", cfg.ConfigFile)
	}
	if cfg.Server.BaseURL != "https://s.example/api/v1" {
		t.Fatalf("unexpected base url: %s", cfg.Server.BaseURL)
	}
	if cfg.TLS.CACertPath != "/workspace/var/envca.crt" {
		t.Fatalf("unexpected ca cert path: %s", cfg.TLS.CACertPath)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("unexpected log level: %s", cfg.Log.Level)
	}
}
