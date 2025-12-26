package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/GoLessons/sufir-keeper-client/internal/api/apiutil"
	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/httpclient"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

func TestTransportNoReplayOnBodyWithoutGetBody(t *testing.T) {
	const pathAuth = "/auth"
	var access string
	var refresh string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == pathAuth:
			resp := map[string]any{
				"access_token":  "acc",
				"refresh_token": "ref",
				"token_type":    "bearer",
				"expires_in":    3600,
			}
			access = "acc"
			refresh = "ref"
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodPatch && r.URL.Path == pathAuth:
			type rb struct {
				RefreshToken string `json:"refresh_token"`
			}
			var body rb
			_ = json.NewDecoder(r.Body).Decode(&body)
			_ = r.Body.Close()
			if body.RefreshToken != refresh {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			access = "acc2"
			refresh = "ref2"
			resp := map[string]any{
				"access_token":  access,
				"refresh_token": refresh,
				"token_type":    "bearer",
				"expires_in":    3600,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodPost && r.URL.Path == "/protected":
			if r.Header.Get("Authorization") != "Bearer "+access {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
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
	_ = store.Clear()
	rt := NewAuthRoundTripper(rc.HTTPClient.Transport, mgr, cfg.Server.BaseURL, store)
	rc.HTTPClient.Transport = rt
	// Body без GetBody
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/protected", bytes.NewBufferString("x"))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid")
	resp, err := rc.HTTPClient.Do(req)
	require.Error(t, err)
	var apiErr apiutil.Error
	require.True(t, errors.As(err, &apiErr))
	require.Equal(t, http.StatusUnauthorized, apiErr.Status)
	require.Nil(t, resp)
}
