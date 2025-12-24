package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Server     ServerConfig
	TLS        TLSConfig
	Log        LogConfig
	Auth       AuthConfig
	ConfigFile string
}

type ServerConfig struct {
	BaseURL string
}

type TLSConfig struct {
	CACertPath string
}

type LogConfig struct {
	Level string
}

type AuthConfig struct {
	TokenStoreService string
	Backend           string
	FileDir           string
}

type Reader interface {
	Set(string, any)
	SetDefault(string, any)
	SetConfigFile(string)
	AutomaticEnv()
	SetEnvPrefix(string)
	SetEnvKeyReplacer(*strings.Replacer)
	ReadInConfig() error
	BindEnv(...string) error
	GetString(string) string
	IsSet(string) bool
}

func EnvKeyReplacer() *strings.Replacer {
	return strings.NewReplacer(".", "_")
}

func Load(v Reader, out *Config) error {
	if out == nil {
		return errors.New("nil config output")
	}
	_ = v.BindEnv("config.file", "SUFIR_KEEPER_CONFIG")
	_ = v.BindEnv("server.base_url", "SUFIR_KEEPER_SERVER")
	_ = v.BindEnv("tls.ca_cert_path", "SUFIR_KEEPER_CA_CERT")
	_ = v.BindEnv("log.level", "SUFIR_KEEPER_LOG_LEVEL")
	_ = v.BindEnv("auth.token_store_service", "SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE")
	_ = v.BindEnv("auth.backend", "SUFIR_KEEPER_AUTH_BACKEND")
	_ = v.BindEnv("auth.file_dir", "SUFIR_KEEPER_AUTH_FILE_DIR")
	_ = v.ReadInConfig()
	out.ConfigFile = v.GetString("config.file")
	out.Server.BaseURL = v.GetString("server.base_url")
	out.TLS.CACertPath = v.GetString("tls.ca_cert_path")
	out.Log.Level = v.GetString("log.level")
	out.Auth.TokenStoreService = v.GetString("auth.token_store_service")
	out.Auth.Backend = v.GetString("auth.backend")
	out.Auth.FileDir = v.GetString("auth.file_dir")
	if out.ConfigFile == "" {
		out.ConfigFile = os.Getenv("SUFIR_KEEPER_CONFIG")
	}
	if out.Server.BaseURL == "" {
		out.Server.BaseURL = os.Getenv("SUFIR_KEEPER_SERVER")
	}
	if out.TLS.CACertPath == "" {
		out.TLS.CACertPath = os.Getenv("SUFIR_KEEPER_CA_CERT")
	}
	if out.Log.Level == "" {
		out.Log.Level = os.Getenv("SUFIR_KEEPER_LOG_LEVEL")
	}
	if out.Auth.TokenStoreService == "" {
		out.Auth.TokenStoreService = os.Getenv("SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE")
	}
	if out.Auth.Backend == "" {
		out.Auth.Backend = os.Getenv("SUFIR_KEEPER_AUTH_BACKEND")
	}
	if out.Auth.FileDir == "" {
		out.Auth.FileDir = os.Getenv("SUFIR_KEEPER_AUTH_FILE_DIR")
	}
	return nil
}
