package auth

import (
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/httpclient"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

func TestLogoutWithoutAccessToken(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	cert := srv.Certificate()
	f, err := os.CreateTemp("", "ca-*.crt")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	_, err = f.Write(pemBytes)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{Server: config.ServerConfig{BaseURL: srv.URL}, TLS: config.TLSConfig{CACertPath: f.Name()}}
	rc, err := httpclient.New(cfg, log)
	require.NoError(t, err)
	store, err := NewKeyringStore(KeyringOptions{ServiceName: "sufir-keeper-client", Backend: "file", FileDir: t.TempDir()})
	require.NoError(t, err)
	mgr := NewManager(rc, store)
	err = mgr.Logout(rc.HTTPClient.Context(), cfg.Server.BaseURL)
	require.NoError(t, err)
	_, ok := store.CurrentAccessToken()
	require.False(t, ok)
}

