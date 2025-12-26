package httpclient

import (
	"context"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

type tempNetErr struct{}

func (tempNetErr) Error() string   { return "temp" }
func (tempNetErr) Timeout() bool   { return true }
func (tempNetErr) Temporary() bool { return true }

func TestCheckRetryBranches(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	cert := srv.Certificate()
	f, err := os.CreateTemp("", "ca-*.crt")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	// write PEM
	pemContent := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	_, err = f.Write(pemContent)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{TLS: config.TLSConfig{CACertPath: f.Name()}}
	client, err := New(cfg, log)
	require.NoError(t, err)
	ok, err := client.CheckRetry(context.Background(), nil, nil)
	require.NoError(t, err)
	require.True(t, ok)
	resp := &http.Response{StatusCode: http.StatusOK}
	ok, err = client.CheckRetry(context.Background(), resp, nil)
	require.NoError(t, err)
	require.False(t, ok)
	ok, err = client.CheckRetry(context.Background(), nil, io.ErrUnexpectedEOF)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = client.CheckRetry(context.Background(), nil, tempNetErr{})
	require.NoError(t, err)
	require.True(t, ok)
}
