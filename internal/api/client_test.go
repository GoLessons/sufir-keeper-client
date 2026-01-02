package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/GoLessons/sufir-keeper-client/internal/auth"
	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

func TestClientWiresHTTPAuthAndAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{
		Server: config.ServerConfig{BaseURL: srv.URL},
	}
	store, err := auth.NewKeyringStore(auth.KeyringOptions{
		ServiceName:  "sufir-keeper-client",
		Backend:      "file",
		FileDir:      filepath.Join(t.TempDir(), "keyring"),
		FilePassword: "test",
	})
	require.NoError(t, err)
	cl, err := New(cfg, log, store)
	require.NoError(t, err)
	require.NotNil(t, cl.HTTP)
	require.NotNil(t, cl.API)
	require.NotNil(t, cl.Auth)
	_, ok := cl.HTTP.HTTPClient.Transport.(*auth.AuthRoundTripper)
	require.True(t, ok)
}
