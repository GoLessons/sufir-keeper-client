package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Auth       AuthConfig
	Server     ServerConfig
	TLS        TLSConfig
	Log        LogConfig
	ConfigFile string
	Cache      CacheConfig
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

type CacheConfig struct {
	Path       string
	TTLMinutes int
	Enabled    bool
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
	_ = v.BindEnv("cache.path", "SUFIR_KEEPER_CACHE_PATH")
	_ = v.BindEnv("cache.ttl_minutes", "SUFIR_KEEPER_CACHE_TTL")
	_ = v.BindEnv("cache.enabled", "SUFIR_KEEPER_CACHE_ENABLED")
	out.ConfigFile = v.GetString("config.file")
	out.Server.BaseURL = v.GetString("server.base_url")
	out.TLS.CACertPath = v.GetString("tls.ca_cert_path")
	out.Log.Level = v.GetString("log.level")
	out.Auth.TokenStoreService = v.GetString("auth.token_store_service")
	out.Auth.Backend = v.GetString("auth.backend")
	out.Auth.FileDir = v.GetString("auth.file_dir")
	out.Cache.Path = v.GetString("cache.path")
	out.Cache.TTLMinutes = atoiSafe(v.GetString("cache.ttl_minutes"))
	out.Cache.Enabled = v.GetString("cache.enabled") == "true"
	if out.ConfigFile == "" {
		out.ConfigFile = os.Getenv("SUFIR_KEEPER_CONFIG")
	}
	if out.ConfigFile != "" {
		v.SetConfigFile(out.ConfigFile)
		if _, err := os.Stat(out.ConfigFile); err == nil {
			if err := v.ReadInConfig(); err != nil {
				return err
			}
		}
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
	if out.Cache.Path == "" {
		out.Cache.Path = os.Getenv("SUFIR_KEEPER_CACHE_PATH")
	}
	if out.Cache.TTLMinutes == 0 {
		if v := os.Getenv("SUFIR_KEEPER_CACHE_TTL"); v != "" {
			out.Cache.TTLMinutes = atoiSafe(v)
		}
	}
	if !out.Cache.Enabled {
		out.Cache.Enabled = os.Getenv("SUFIR_KEEPER_CACHE_ENABLED") == "true"
	}
	return nil
}

func atoiSafe(s string) int {
	if s == "" {
		return 0
	}
	var n int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0
		}
	}
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
