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
	_ = v.ReadInConfig()
	out.ConfigFile = v.GetString("config.file")
	out.Server.BaseURL = v.GetString("server.base_url")
	out.TLS.CACertPath = v.GetString("tls.ca_cert_path")
	out.Log.Level = v.GetString("log.level")
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
	return nil
}
