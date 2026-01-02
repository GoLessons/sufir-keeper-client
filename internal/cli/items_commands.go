package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/GoLessons/sufir-keeper-client/internal/api"
	"github.com/GoLessons/sufir-keeper-client/internal/api/apigen"
	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
	"github.com/GoLessons/sufir-keeper-client/internal/service"
)

func AttachItemsCommands(root *cobra.Command) {
	root.AddCommand(newItemsListCmd())
	root.AddCommand(newItemsGetCmd())
	root.AddCommand(newItemsCreateCmd())
	root.AddCommand(newItemsUpdateCmd())
	root.AddCommand(newItemsDeleteCmd())
}

func newItemsListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Список записей",
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
			cm, cerr := newCache(cfg)
			if cerr != nil {
				return cerr
			}
			defer func() { _ = cm.Close() }()
			w := api.NewWrapper(cl)
			svc := service.NewItemsService(w, cm, cfg)
			var params apigen.GetItemsParams
			if s := strings.TrimSpace(cmd.Flag("search").Value.String()); s != "" {
				params.S = &s
			}
			if tt := strings.TrimSpace(cmd.Flag("type").Value.String()); tt != "" {
				t := apigen.ItemType(tt)
				params.Type = &t
			}
			if v := cmd.Flag("limit").Value.String(); v != "" {
				if n, err := strconv.Atoi(v); err == nil && n > 0 {
					params.Limit = &n
				}
			}
			if v := cmd.Flag("offset").Value.String(); v != "" {
				if n, err := strconv.Atoi(v); err == nil && n >= 0 {
					params.Offset = &n
				}
			}
			resp, err := svc.List(ctx, &params)
			if err != nil {
				return err
			}
			if resp.JSON200 != nil && resp.JSON200.Items != nil {
				for _, it := range *resp.JSON200.Items {
					title := ""
					if it.Title != nil {
						title = *it.Title
					}
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", title)
				}
			}
			return nil
		},
	}
	cmd.Flags().String("search", "", "Поиск по наименованию")
	cmd.Flags().String("type", "", "Тип записи: TEXT|CREDENTIAL|CARD|BINARY")
	cmd.Flags().Int("limit", 0, "Лимит")
	cmd.Flags().Int("offset", 0, "Смещение")
	return cmd
}

func newItemsGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Получить запись",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idv, err := uuid.Parse(args[0])
			if err != nil {
				return errors.New("некорректный UUID")
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
			cm, cerr := newCache(cfg)
			if cerr != nil {
				return cerr
			}
			defer func() { _ = cm.Close() }()
			w := api.NewWrapper(cl)
			svc := service.NewItemsService(w, cm, cfg)
			var id openapi_types.UUID
			if err := id.UnmarshalText([]byte(idv.String())); err != nil {
				return err
			}
			resp, err := svc.Get(ctx, id)
			if err != nil {
				return err
			}
			if resp.JSON200 != nil && resp.JSON200.Title != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", *resp.JSON200.Title)
			}
			return nil
		},
	}
}

func newItemsCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Создать запись",
		RunE: func(cmd *cobra.Command, args []string) error {
			title, _ := cmd.Flags().GetString("title")
			if strings.TrimSpace(title) == "" {
				return errors.New("требуется title")
			}
			ttype, _ := cmd.Flags().GetString("type")
			ttype = strings.ToUpper(strings.TrimSpace(ttype))
			var data apigen.ItemCreate_Data
			switch ttype {
			case "TEXT", "":
				value, _ := cmd.Flags().GetString("value")
				if strings.TrimSpace(value) == "" {
					return errors.New("требуется value для TEXT")
				}
				if err := data.FromTextData(apigen.TextData{Type: "TEXT", Value: value}); err != nil {
					return err
				}
			case "CREDENTIAL":
				login, _ := cmd.Flags().GetString("login")
				password, _ := cmd.Flags().GetString("password")
				if strings.TrimSpace(login) == "" || strings.TrimSpace(password) == "" {
					return errors.New("требуются login и password для CREDENTIAL")
				}
				if err := data.FromCredentialData(apigen.CredentialData{Type: "CREDENTIAL", Login: login, Password: password}); err != nil {
					return err
				}
			case "CARD":
				cardNumber, _ := cmd.Flags().GetString("card-number")
				cardHolder, _ := cmd.Flags().GetString("card-holder")
				expiryDate, _ := cmd.Flags().GetString("expiry-date")
				cvv, _ := cmd.Flags().GetString("cvv")
				if strings.TrimSpace(cardNumber) == "" || strings.TrimSpace(cardHolder) == "" || strings.TrimSpace(expiryDate) == "" || strings.TrimSpace(cvv) == "" {
					return errors.New("требуются card-number, card-holder, expiry-date, cvv для CARD")
				}
				if err := data.FromCardData(apigen.CardData{Type: "CARD", CardNumber: cardNumber, CardHolder: cardHolder, ExpiryDate: expiryDate, Cvv: cvv}); err != nil {
					return err
				}
			case "BINARY":
				filename, _ := cmd.Flags().GetString("filename")
				bid, _ := cmd.Flags().GetString("binary-id")
				if strings.TrimSpace(filename) == "" || strings.TrimSpace(bid) == "" {
					return errors.New("требуются filename и binary-id для BINARY")
				}
				u, err := uuid.Parse(bid)
				if err != nil {
					return errors.New("некорректный UUID в binary-id")
				}
				if err := data.FromBinaryData(apigen.BinaryData{Type: "BINARY", Filename: filename, Id: u}); err != nil {
					return err
				}
			default:
				return errors.New("неподдерживаемый type, используйте TEXT|CREDENTIAL|CARD|BINARY")
			}
			body := apigen.ItemCreate{
				Title: title,
				Data:  data,
			}
			if m := strings.TrimSpace(cmd.Flag("meta").Value.String()); m != "" {
				meta := parseMeta(m)
				body.Meta = &meta
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
			cm, cerr := newCache(cfg)
			if cerr != nil {
				return cerr
			}
			defer func() { _ = cm.Close() }()
			w := api.NewWrapper(cl)
			svc := service.NewItemsService(w, cm, cfg)
			resp, err := svc.Create(ctx, body)
			if err != nil {
				return err
			}
			if resp.JSON201 != nil && resp.JSON201.Id != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", resp.JSON201.Id.String())
			} else {
				_, _ = cmd.OutOrStdout().Write([]byte("OK\n"))
			}
			return nil
		},
	}
	cmd.Flags().String("title", "", "Заголовок")
	cmd.Flags().String("type", "TEXT", "Тип: TEXT|CREDENTIAL|CARD|BINARY")
	cmd.Flags().String("value", "", "Значение для TEXT")
	cmd.Flags().String("login", "", "Логин для CREDENTIAL")
	cmd.Flags().String("password", "", "Пароль для CREDENTIAL")
	cmd.Flags().String("card-number", "", "Номер карты для CARD")
	cmd.Flags().String("card-holder", "", "Владелец карты для CARD")
	cmd.Flags().String("expiry-date", "", "Срок действия (MM/YY) для CARD")
	cmd.Flags().String("cvv", "", "CVV для CARD")
	cmd.Flags().String("filename", "", "Имя файла для BINARY")
	cmd.Flags().String("binary-id", "", "UUID файла для BINARY")
	cmd.Flags().String("meta", "", "Метаданные key=value через запятую")
	return cmd
}

func newItemsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [id]",
		Short: "Обновить запись",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idv, err := uuid.Parse(args[0])
			if err != nil {
				return errors.New("некорректный UUID")
			}
			title, _ := cmd.Flags().GetString("title")
			ttype, _ := cmd.Flags().GetString("type")
			ttype = strings.ToUpper(strings.TrimSpace(ttype))
			value, _ := cmd.Flags().GetString("value")
			var body apigen.UpdateItemJSONRequestBody
			if strings.TrimSpace(title) != "" || strings.TrimSpace(value) != "" || strings.TrimSpace(cmd.Flag("meta").Value.String()) != "" || ttype != "" {
				var u apigen.ItemUpdate
				if strings.TrimSpace(title) != "" {
					u.Title = &title
				}
				if ttype != "" {
					var d apigen.ItemUpdate_Data
					switch ttype {
					case "TEXT":
						if strings.TrimSpace(value) == "" {
							return errors.New("требуется value для TEXT")
						}
						if err := d.UnmarshalJSON([]byte(fmt.Sprintf(`{"type":"TEXT","value":%q}`, value))); err != nil {
							return err
						}
						u.Data = &d
					case "CREDENTIAL":
						login, _ := cmd.Flags().GetString("login")
						password, _ := cmd.Flags().GetString("password")
						if strings.TrimSpace(login) == "" && strings.TrimSpace(password) == "" {
							return errors.New("нужно указать хотя бы один из login|password для CREDENTIAL")
						}
						obj := map[string]string{"type": "CREDENTIAL"}
						if strings.TrimSpace(login) != "" {
							obj["login"] = login
						}
						if strings.TrimSpace(password) != "" {
							obj["password"] = password
						}
						b := []byte(fmt.Sprintf(`{"type":"CREDENTIAL"%s%s}`,
							func() string {
								if strings.TrimSpace(login) != "" {
									return fmt.Sprintf(`,"login":%q`, login)
								}
								return ""
							}(),
							func() string {
								if strings.TrimSpace(password) != "" {
									return fmt.Sprintf(`,"password":%q`, password)
								}
								return ""
							}(),
						))
						if err := d.UnmarshalJSON(b); err != nil {
							return err
						}
						u.Data = &d
					case "CARD":
						cardNumber, _ := cmd.Flags().GetString("card-number")
						cardHolder, _ := cmd.Flags().GetString("card-holder")
						expiryDate, _ := cmd.Flags().GetString("expiry-date")
						cvv, _ := cmd.Flags().GetString("cvv")
						if strings.TrimSpace(cardNumber) == "" && strings.TrimSpace(cardHolder) == "" && strings.TrimSpace(expiryDate) == "" && strings.TrimSpace(cvv) == "" {
							return errors.New("нужно указать хотя бы одно из card-number|card-holder|expiry-date|cvv для CARD")
						}
						payload := `{"type":"CARD"}`
						if strings.TrimSpace(cardNumber) != "" {
							payload = payload[:len(payload)-1] + fmt.Sprintf(`,"card_number":%q}`, cardNumber)
						}
						if strings.TrimSpace(cardHolder) != "" {
							if payload[len(payload)-1] == '}' {
								payload = payload[:len(payload)-1] + fmt.Sprintf(`,"card_holder":%q}`, cardHolder)
							} else {
								payload = fmt.Sprintf(`{"type":"CARD","card_holder":%q}`, cardHolder)
							}
						}
						if strings.TrimSpace(expiryDate) != "" {
							if payload[len(payload)-1] == '}' {
								payload = payload[:len(payload)-1] + fmt.Sprintf(`,"expiry_date":%q}`, expiryDate)
							} else {
								payload = fmt.Sprintf(`{"type":"CARD","expiry_date":%q}`, expiryDate)
							}
						}
						if strings.TrimSpace(cvv) != "" {
							if payload[len(payload)-1] == '}' {
								payload = payload[:len(payload)-1] + fmt.Sprintf(`,"cvv":%q}`, cvv)
							} else {
								payload = fmt.Sprintf(`{"type":"CARD","cvv":%q}`, cvv)
							}
						}
						if err := d.UnmarshalJSON([]byte(payload)); err != nil {
							return err
						}
						u.Data = &d
					case "BINARY":
						filename, _ := cmd.Flags().GetString("filename")
						bid, _ := cmd.Flags().GetString("binary-id")
						var parts []string
						if strings.TrimSpace(filename) != "" {
							parts = append(parts, fmt.Sprintf(`"filename":%q`, filename))
						}
						if strings.TrimSpace(bid) != "" {
							if _, e := uuid.Parse(bid); e != nil {
								return errors.New("некорректный UUID в binary-id")
							}
							parts = append(parts, fmt.Sprintf(`"id":%q`, bid))
						}
						if len(parts) == 0 {
							return errors.New("нужно указать filename или binary-id для BINARY")
						}
						payload := fmt.Sprintf(`{"type":"BINARY",%s}`, strings.Join(parts, ","))
						var d apigen.ItemUpdate_Data
						if err := d.UnmarshalJSON([]byte(payload)); err != nil {
							return err
						}
						u.Data = &d
					default:
						return errors.New("неподдерживаемый type, используйте TEXT|CREDENTIAL|CARD|BINARY")
					}
				}
				if m := strings.TrimSpace(cmd.Flag("meta").Value.String()); m != "" {
					meta := parseMeta(m)
					u.Meta = &meta
				}
				body = u
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
			cm, cerr := newCache(cfg)
			if cerr != nil {
				return cerr
			}
			defer func() { _ = cm.Close() }()
			w := api.NewWrapper(cl)
			svc := service.NewItemsService(w, cm, cfg)
			var id openapi_types.UUID
			if err := id.UnmarshalText([]byte(idv.String())); err != nil {
				return err
			}
			resp, err := svc.Update(ctx, id, body)
			if err != nil {
				return err
			}
			if resp.JSON200 != nil && resp.JSON200.Id != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", resp.JSON200.Id.String())
			} else {
				_, _ = cmd.OutOrStdout().Write([]byte("OK\n"))
			}
			return nil
		},
	}
	cmd.Flags().String("title", "", "Заголовок")
	cmd.Flags().String("type", "", "Тип: TEXT|CREDENTIAL|CARD|BINARY")
	cmd.Flags().String("value", "", "Значение для TEXT")
	cmd.Flags().String("login", "", "Логин для CREDENTIAL")
	cmd.Flags().String("password", "", "Пароль для CREDENTIAL")
	cmd.Flags().String("card-number", "", "Номер карты для CARD")
	cmd.Flags().String("card-holder", "", "Владелец карты для CARD")
	cmd.Flags().String("expiry-date", "", "Срок действия (MM/YY) для CARD")
	cmd.Flags().String("cvv", "", "CVV для CARD")
	cmd.Flags().String("filename", "", "Имя файла для BINARY")
	cmd.Flags().String("binary-id", "", "UUID файла для BINARY")
	cmd.Flags().String("meta", "", "Метаданные key=value через запятую")
	return cmd
}

func newItemsDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [id]",
		Short: "Удалить запись",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idv, err := uuid.Parse(args[0])
			if err != nil {
				return errors.New("некорректный UUID")
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
			cm, cerr := newCache(cfg)
			if cerr != nil {
				return cerr
			}
			defer func() { _ = cm.Close() }()
			w := api.NewWrapper(cl)
			svc := service.NewItemsService(w, cm, cfg)
			var id openapi_types.UUID
			if err := id.UnmarshalText([]byte(idv.String())); err != nil {
				return err
			}
			_, err = svc.Delete(ctx, id)
			if err != nil {
				return err
			}
			_, _ = cmd.OutOrStdout().Write([]byte("OK\n"))
			return nil
		},
	}
}

func parseMeta(s string) map[string]string {
	res := make(map[string]string)
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		if len(kv) == 2 {
			res[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return res
}
