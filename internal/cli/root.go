package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

type contextKey string

const (
	cfgContextKey contextKey = "config"
	logContextKey contextKey = "logger"
)

func NewRootCmd(version, commit, date string) *cobra.Command {
	v := viper.New()
	v.SetEnvPrefix("SUFIR_KEEPER")
	v.SetEnvKeyReplacer(config.EnvKeyReplacer())
	v.AutomaticEnv()

	cmd := &cobra.Command{
		Use:           "keepcli",
		Short:         "Sufir Keeper CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			executablePath, err := os.Executable()
			if err != nil {
				return err
			}
			execDir := filepath.Dir(executablePath)
			if !v.IsSet("config.file") || v.GetString("config.file") == "" {
				v.Set("config.file", filepath.Join(execDir, "config.json"))
			}
			v.SetConfigFile(v.GetString("config.file"))
			v.SetDefault("server.base_url", "https://localhost:8443/api/v1")
			v.SetDefault("log.level", "info")
			v.SetDefault("tls.ca_cert_path", "./var/ca.crt")
			v.SetDefault("auth.token_store_service", "sufir-keeper-client")
			v.SetDefault("auth.backend", "")
			v.SetDefault("auth.file_dir", "")
			v.SetDefault("cache.path", "~/.local/share/sufir-keeper-client/cache.db")
			v.SetDefault("cache.ttl_minutes", 180)
			v.SetDefault("cache.enabled", true)
			var cfg config.Config
			if err := config.Load(v, &cfg); err != nil {
				return err
			}
			l, err := logging.NewLogger(cfg.Log.Level)
			if err != nil {
				return err
			}
			ctx := context.WithValue(cmd.Context(), cfgContextKey, cfg)
			ctx = context.WithValue(ctx, logContextKey, l)
			cmd.SetContext(ctx)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	if version != "" {
		cmd.Version = version
	}

	cmd.PersistentFlags().String("config", "", "Путь к файлу конфигурации")
	cmd.PersistentFlags().String("server", "", "Адрес сервера API")
	cmd.PersistentFlags().String("log-level", "", "Уровень логирования")
	cmd.PersistentFlags().String("ca-cert-path", "", "Путь к dev CA сертификату")

	_ = v.BindPFlag("config.file", cmd.PersistentFlags().Lookup("config"))
	_ = v.BindPFlag("server.base_url", cmd.PersistentFlags().Lookup("server"))
	_ = v.BindPFlag("log.level", cmd.PersistentFlags().Lookup("log-level"))
	_ = v.BindPFlag("tls.ca_cert_path", cmd.PersistentFlags().Lookup("ca-cert-path"))
	_ = v.BindEnv("auth.token_store_service", "SUFIR_KEEPER_AUTH_TOKEN_STORE_SERVICE")
	_ = v.BindEnv("auth.backend", "SUFIR_KEEPER_AUTH_BACKEND")
	_ = v.BindEnv("auth.file_dir", "SUFIR_KEEPER_AUTH_FILE_DIR")

	if version != "" || commit != "" || date != "" {
		versionCmd := &cobra.Command{
			Use:   "version",
			Short: "Показать версию приложения",
			RunE: func(c *cobra.Command, args []string) error {
				out := fmt.Sprintf("version: %s\ncommit: %s\ndate: %s\n", nonEmpty(version, "dev"), nonEmpty(commit, "none"), nonEmpty(date, "unknown"))
				_, err := c.OutOrStdout().Write([]byte(out))
				return err
			},
		}
		cmd.AddCommand(versionCmd)
	}

	AttachAuthCommands(cmd)
	AttachItemsCommands(cmd)
	AttachFilesCommands(cmd)
	AttachCompletion(cmd)

	return cmd
}

func nonEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
