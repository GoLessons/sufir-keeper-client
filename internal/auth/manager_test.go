package auth

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/httpclient"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

func writeServerCertToFile(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	cert := srv.Certificate()
	der := cert.Raw
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	f, err := os.CreateTemp("", "ca-*.crt")
	require.NoError(t, err)
	_, err = f.Write(pemBytes)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestAuthFlowLoginRefreshVerifyLogout(t *testing.T) {
	var refreshCalls int64
	var currentAccess string
	var currentRefresh string
	userID := "11111111-1111-1111-1111-111111111111"
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/register":
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPost && r.URL.Path == "/auth":
			type loginBody struct {
				Login    string `json:"login"`
				Password string `json:"password"`
			}
			var lb loginBody
			_ = json.NewDecoder(r.Body).Decode(&lb)
			_ = r.Body.Close()
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
		case r.Method == http.MethodPatch && r.URL.Path == "/auth":
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
		case r.Method == http.MethodDelete && r.URL.Path == "/auth":
			if r.Header.Get("Authorization") == "Bearer "+currentAccess {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		case r.Method == http.MethodGet && r.URL.Path == "/auth-verify":
			if r.Header.Get("Authorization") != "Bearer "+currentAccess {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("X-User-Id", userID)
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/protected":
			if r.Header.Get("Authorization") != "Bearer "+currentAccess {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	caPath := writeServerCertToFile(t, srv)
	t.Cleanup(func() { _ = os.Remove(caPath) })
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{
		Server: config.ServerConfig{BaseURL: srv.URL},
		TLS:    config.TLSConfig{CACertPath: caPath},
	}
	client, err := httpclient.New(cfg, log)
	require.NoError(t, err)
	dir := t.TempDir()
	store, err := NewKeyringStore(KeyringOptions{
		ServiceName: "sufir-keeper-client",
		Backend:     "file",
		FileDir:     filepath.Join(dir, "keyring"),
	})
	require.NoError(t, err)
	mgr := NewManager(client, store)
	err = mgr.Register(contextWithTimeout(), cfg.Server.BaseURL, "user", "password")
	require.NoError(t, err)
	tokens, err := mgr.Login(contextWithTimeout(), cfg.Server.BaseURL, "user", "password")
	require.NoError(t, err)
	require.Equal(t, "access-init", tokens.AccessToken)
	require.Equal(t, "refresh-init", tokens.RefreshToken)
	info, err := mgr.Verify(contextWithTimeout(), cfg.Server.BaseURL)
	require.NoError(t, err)
	require.Equal(t, userID, info.UserID)
	rt := NewAuthRoundTripper(client.HTTPClient.Transport, mgr, cfg.Server.BaseURL, store)
	client.HTTPClient.Transport = rt
	_, err = client.Get(srv.URL + "/protected")
	require.NoError(t, err)
	require.GreaterOrEqual(t, atomic.LoadInt64(&refreshCalls), int64(0))
	err = mgr.Logout(contextWithTimeout(), cfg.Server.BaseURL)
	require.NoError(t, err)
	_, ok := store.CurrentAccessToken()
	require.False(t, ok)
}

func contextWithTimeout() (ctx context.Context) {
	c, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return c
}
