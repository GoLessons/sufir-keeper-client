package auth

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/httpclient"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

func TestRefreshSingleflightEnsuresSingleCall(t *testing.T) {
	const pathAuth = "/auth"
	var refreshCalls int64
	var currentAccess string
	var currentRefresh string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPatch && r.URL.Path == pathAuth:
			atomic.AddInt64(&refreshCalls, 1)
			type rb struct {
				RefreshToken string `json:"refresh_token"`
			}
			var body rb
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = r.Body.Close()
			if body.RefreshToken != currentRefresh {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			currentAccess = "access-refreshed"
			currentRefresh = "refresh-refreshed"
			resp := map[string]any{
				"access_token":  currentAccess,
				"refresh_token": currentRefresh,
				"token_type":    "bearer",
				"expires_in":    3600,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodPost && r.URL.Path == pathAuth:
			resp := map[string]any{
				"access_token":  "access-init",
				"refresh_token": "refresh-init",
				"token_type":    "bearer",
				"expires_in":    3600,
			}
			currentAccess = resp["access_token"].(string)
			currentRefresh = resp["refresh_token"].(string)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
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
	store, err := NewKeyringStore(KeyringOptions{ServiceName: "sufir-keeper-client", Backend: "file", FileDir: t.TempDir(), FilePassword: "test"})
	require.NoError(t, err)
	mgr := NewManager(rc, store)
	_, err = mgr.Login(context.Background(), cfg.Server.BaseURL, "u", "p")
	require.NoError(t, err)
	var wg sync.WaitGroup
	const n = 10
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, _ = mgr.Refresh(context.Background(), cfg.Server.BaseURL)
		}()
	}
	wg.Wait()
	require.Equal(t, int64(1), atomic.LoadInt64(&refreshCalls))
	at, ok := store.CurrentAccessToken()
	require.True(t, ok)
	require.Equal(t, currentAccess, at)
}
