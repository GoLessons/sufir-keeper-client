package cli

import (
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/GoLessons/sufir-keeper-client/internal/api"
	"github.com/GoLessons/sufir-keeper-client/internal/api/apigen"
	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

func AttachFilesCommands(root *cobra.Command) {
	root.AddCommand(newFilesUploadCmd())
	root.AddCommand(newFilesDownloadCmd())
}

func newFilesUploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Загрузить файл на сервер",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := cmd.Flags().GetString("path")
			if path == "" {
				return fmt.Errorf("не указан путь к файлу")
			}
			fi, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fi.Close()
			st, err := fi.Stat()
			if err != nil {
				return err
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
			w := api.NewWrapper(cl)
			presp, err := w.PresignFile(ctx, apigen.PresignFileJSONRequestBody{
				FileId:   uuid.New(),
				Filename: ptrString(filepath.Base(path)),
			})
			if err != nil {
				return err
			}
			if presp.JSON200 == nil || presp.JSON200.UploadUrl == nil {
				return fmt.Errorf("presign failed")
			}
			fields := map[string]string{}
			if presp.JSON200.FormFields != nil {
				for k, v := range *presp.JSON200.FormFields {
					fields[k] = v
				}
			}
			bodyReader, contentType, err := buildPresignedMultipartWithProgress(fi, filepath.Base(path), st.Size(), fields, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, *presp.JSON200.UploadUrl, bodyReader)
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", contentType)
			rresp, err := cl.HTTP.HTTPClient.Do(req)
			if err != nil {
				return err
			}
			if rresp.StatusCode == http.StatusNoContent || rresp.StatusCode == http.StatusOK {
				_, _ = cmd.OutOrStdout().Write([]byte("OK\n"))
				return nil
			}
			return fmt.Errorf("upload failed: %d", rresp.StatusCode)
		},
	}
	cmd.Flags().String("path", "", "Путь к файлу")
	return cmd
}

func newFilesDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download [id] [out]",
		Short: "Скачать файл",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			idv, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("некорректный UUID")
			}
			out := args[1]
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
			w := api.NewWrapper(cl)
			var id openapi_types.UUID
			if err := id.UnmarshalText([]byte(idv.String())); err != nil {
				return err
			}
			resp, err := w.DownloadFile(ctx, id)
			if err != nil {
				return err
			}
			total := len(resp.Body)
			chunk := 64 * 1024
			var written int
			f, err := os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
			if err != nil {
				return err
			}
			defer f.Close()
			last := time.Now()
			for i := 0; i < total; i += chunk {
				end := i + chunk
				if end > total {
					end = total
				}
				n, werr := f.Write(resp.Body[i:end])
				if werr != nil {
					return werr
				}
				written += n
				if time.Since(last) >= 200*time.Millisecond || written == total {
					printProgress(cmd.OutOrStdout(), int64(written), int64(total))
					last = time.Now()
				}
			}
			_, _ = cmd.OutOrStdout().Write([]byte("OK\n"))
			return nil
		},
	}
	return cmd
}

func buildPresignedMultipartWithProgress(file io.Reader, filename string, total int64, fields map[string]string, out io.Writer) (io.Reader, string, error) {
	pr, pw := io.Pipe()
	w := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		defer w.Close()
		for k, v := range fields {
			if err := w.WriteField(k, v); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
		fw, err := w.CreateFormFile("file", filename)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		var written int64
		last := time.Now()
		buf := make([]byte, 64*1024)
		for {
			n, rerr := file.Read(buf)
			if n > 0 {
				if _, werr := fw.Write(buf[:n]); werr != nil {
					pw.CloseWithError(werr)
					return
				}
				written += int64(n)
				if time.Since(last) >= 200*time.Millisecond || written == total {
					printProgress(out, written, total)
					last = time.Now()
				}
			}
			if rerr != nil {
				if rerr == io.EOF {
					break
				}
				pw.CloseWithError(rerr)
				return
			}
		}
	}()
	return pr, w.FormDataContentType(), nil
}

func printProgress(out io.Writer, done, total int64) {
	if total <= 0 {
		return
	}
	p := int(math.Round(float64(done) * 100.0 / float64(total)))
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	_, _ = fmt.Fprintf(out, "progress: %d%%\n", p)
}

func ptrString(s string) *string { return &s }
