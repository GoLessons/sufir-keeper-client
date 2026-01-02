package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/99designs/keyring"

	"github.com/GoLessons/sufir-keeper-client/internal/api"
	"github.com/GoLessons/sufir-keeper-client/internal/auth"
	"github.com/GoLessons/sufir-keeper-client/internal/cache"
	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

func AttachAuthCommands(root *cobra.Command) {
	root.AddCommand(newLoginCmd())
	root.AddCommand(newStatusCmd())
	root.AddCommand(newLogoutCmd())
	root.AddCommand(newRegisterCmd())
}

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Войти в систему",
		RunE: func(cmd *cobra.Command, args []string) error {
			login, _ := cmd.Flags().GetString("login")
			if strings.TrimSpace(login) == "" {
				return errors.New("не указан логин")
			}
			password, _ := cmd.Flags().GetString("password")
			if password == "" {
				pw, err := readPassword(cmd, "Введите пароль: ")
				if err != nil {
					return err
				}
				password = pw
			}
			ctx := cmd.Context()
			cfg := ctx.Value(cfgContextKey).(config.Config)
			log := ctx.Value(logContextKey).(logging.Logger)
			store, err := newStore(cfg)
			if err != nil {
				return err
			}
			cl, err := api.New(cfg, log, store)
			if err != nil {
				return err
			}
			_, err = cl.Auth.Login(ctx, cfg.Server.BaseURL, login, password)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write([]byte("OK\n"))
			return err
		},
	}
	cmd.Flags().String("login", "", "Логин")
	cmd.Flags().String("password", "", "Пароль")
	return cmd
}

func newRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Создать учетную запись",
		RunE: func(cmd *cobra.Command, args []string) error {
			login, _ := cmd.Flags().GetString("login")
			if strings.TrimSpace(login) == "" {
				return errors.New("не указан логин")
			}
			password, _ := cmd.Flags().GetString("password")
			if password == "" {
				pw, err := readPassword(cmd, "Введите пароль: ")
				if err != nil {
					return err
				}
				password = pw
			}
			ctx := cmd.Context()
			cfg := ctx.Value(cfgContextKey).(config.Config)
			log := ctx.Value(logContextKey).(logging.Logger)
			store, err := newStore(cfg)
			if err != nil {
				return err
			}
			cl, err := api.New(cfg, log, store)
			if err != nil {
				return err
			}
			err = cl.Auth.Register(ctx, cfg.Server.BaseURL, login, password)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write([]byte("OK\n"))
			return err
		},
	}
	cmd.Flags().String("login", "", "Логин")
	cmd.Flags().String("password", "", "Пароль")
	return cmd
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Выйти из системы",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ctx.Value(cfgContextKey).(config.Config)
			log := ctx.Value(logContextKey).(logging.Logger)
			store, err := newStore(cfg)
			if err != nil {
				return err
			}
			cl, err := api.New(cfg, log, store)
			if err != nil {
				return err
			}
			if err := cl.Auth.Logout(ctx, cfg.Server.BaseURL); err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write([]byte("OK\n"))
			return err
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Показать статус аутентификации",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := ctx.Value(cfgContextKey).(config.Config)
			log := ctx.Value(logContextKey).(logging.Logger)
			store, err := newStore(cfg)
			if err != nil {
				return err
			}
			cl, err := api.New(cfg, log, store)
			if err != nil {
				return err
			}
			access, ok := store.CurrentAccessToken()
			if !ok || access == "" {
				_, werr := cmd.OutOrStdout().Write([]byte("Не авторизован\n"))
				return werr
			}
			info, verr := cl.Auth.Verify(ctx, cfg.Server.BaseURL)
			if verr != nil {
				_, werr := cmd.OutOrStdout().Write([]byte("Токен недействителен\n"))
				return werr
			}
			out := fmt.Sprintf("Авторизован: %s\n", info.UserID)
			_, werr := cmd.OutOrStdout().Write([]byte(out))
			return werr
		},
	}
}

func readPassword(cmd *cobra.Command, prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		_, _ = cmd.OutOrStdout().Write([]byte(prompt))
		b, err := term.ReadPassword(fd)
		_, _ = cmd.OutOrStdout().Write([]byte("\n"))
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	_, _ = cmd.OutOrStdout().Write([]byte(prompt))
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func newStore(cfg config.Config) (auth.TokenStore, error) {
	opts := auth.KeyringOptions{
		ServiceName: cfg.Auth.TokenStoreService,
		Backend:     cfg.Auth.Backend,
		FileDir:     cfg.Auth.FileDir,
	}
	return auth.NewKeyringStore(opts)
}

func newCache(cfg config.Config) (*cache.Manager, error) {
	kr := keyringConfigFromAuth(cfg)
	return cache.New(cache.Options{
		Path:          cfg.Cache.Path,
		TTLMinutes:    cfg.Cache.TTLMinutes,
		KeyringConfig: kr,
		KeyName:       "cache_key",
	})
}

func keyringConfigFromAuth(cfg config.Config) keyring.Config {
	c := keyring.Config{
		ServiceName: cfg.Auth.TokenStoreService,
	}
	if cfg.Auth.Backend == "file" {
		c.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
		c.FileDir = cfg.Auth.FileDir
		c.FilePasswordFunc = func(string) (string, error) { return "sufir-keeper-dev", nil }
	}
	return c
}
