package cli

import (
	"context"
	"os"
	"path/filepath"

	"github.com/GoLessons/sufir-keeper-server/internal/config"
	"github.com/GoLessons/sufir-keeper-server/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type contextKey string

const cfgContextKey contextKey = "config"
const logContextKey contextKey = "logger"

func NewRootCmd() *cobra.Command {
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
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			if !v.IsSet("config.file") || v.GetString("config.file") == "" {
				v.Set("config.file", filepath.Join(homeDir, ".config", "keepcli", "config.yaml"))
			}
			v.SetConfigFile(v.GetString("config.file"))
			v.SetDefault("server.base_url", "https://localhost:8443/api/v1")
			v.SetDefault("log.level", "info")
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

	cmd.PersistentFlags().String("config", "", "Путь к файлу конфигурации")
	cmd.PersistentFlags().String("server", "", "Адрес сервера API")
	cmd.PersistentFlags().String("log-level", "", "Уровень логирования")
	cmd.PersistentFlags().String("ca-cert-path", "", "Путь к dev CA сертификату")

	_ = v.BindPFlag("config.file", cmd.PersistentFlags().Lookup("config"))
	_ = v.BindPFlag("server.base_url", cmd.PersistentFlags().Lookup("server"))
	_ = v.BindPFlag("log.level", cmd.PersistentFlags().Lookup("log-level"))
	_ = v.BindPFlag("tls.ca_cert_path", cmd.PersistentFlags().Lookup("ca-cert-path"))
	
	return cmd
}
