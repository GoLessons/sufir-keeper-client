package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/stretchr/testify/require"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/GoLessons/sufir-keeper-client/internal/api"
	"github.com/GoLessons/sufir-keeper-client/internal/api/apigen"
	"github.com/GoLessons/sufir-keeper-client/internal/cache"
	"github.com/GoLessons/sufir-keeper-client/internal/config"
)

type failingDoer struct{}

func (f failingDoer) Do(*http.Request) (*http.Response, error) {
	return nil, &timeoutError{}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestItemsService_List_FallbackToCache(t *testing.T) {
	dir := t.TempDir()
	opts := cache.Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 5,
		KeyringConfig: keyring.Config{
			AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
			FileDir:          dir,
			FilePasswordFunc: func(string) (string, error) { return "pw", nil },
			ServiceName:      "sufir-keeper-client",
		},
	}
	cm, err := cache.New(opts)
	require.NoError(t, err)
	defer func() { _ = cm.Close() }()

	apiClient, err := apigen.NewClientWithResponses("http://127.0.0.1:1", apigen.WithHTTPClient(failingDoer{}))
	require.NoError(t, err)
	w := api.NewWrapperFromAPI(apiClient)
	cfg := config.Config{}
	cfg.Cache.Enabled = true
	svc := NewItemsService(w, cm, cfg)

	body := []byte(`{"items":[{"title":"t"}],"limit":1,"offset":0,"total":1}`)
	require.NoError(t, cm.Put("items:list:type=;s=;limit=0;offset=0", body, nil, ""))

	ctx := context.Background()
	resp, err := svc.List(ctx, &apigen.GetItemsParams{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, body, resp.Body)
}

func TestItemsService_Get_FallbackToCache(t *testing.T) {
	dir := t.TempDir()
	opts := cache.Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 5,
		KeyringConfig: keyring.Config{
			AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
			FileDir:          dir,
			FilePasswordFunc: func(string) (string, error) { return "pw", nil },
			ServiceName:      "sufir-keeper-client",
		},
	}
	cm, err := cache.New(opts)
	require.NoError(t, err)
	defer func() { _ = cm.Close() }()

	apiClient, err := apigen.NewClientWithResponses("http://127.0.0.1:1", apigen.WithHTTPClient(failingDoer{}))
	require.NoError(t, err)
	w := api.NewWrapperFromAPI(apiClient)
	cfg := config.Config{}
	cfg.Cache.Enabled = true
	svc := NewItemsService(w, cm, cfg)

	id := openapiUUIDFromString(t, "00000000-0000-0000-0000-000000000001")
	body := []byte(`{"title":"x","id":"00000000-0000-0000-0000-000000000001"}`)
	require.NoError(t, cm.Put("items:get:00000000-0000-0000-0000-000000000001", body, nil, ""))

	ctx := context.Background()
	resp, err := svc.Get(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, body, resp.Body)
}

func TestItemsService_List_SuccessUpdatesCache(t *testing.T) {
	dir := t.TempDir()
	opts := cache.Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 5,
		KeyringConfig: keyring.Config{
			AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
			FileDir:          dir,
			FilePasswordFunc: func(string) (string, error) { return "pw", nil },
			ServiceName:      "sufir-keeper-client",
		},
	}
	cm, err := cache.New(opts)
	require.NoError(t, err)
	defer func() { _ = cm.Close() }()
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"title":"t"}],"limit":1,"offset":0,"total":1}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	apiClient, err := apigen.NewClientWithResponses(srv.URL, apigen.WithHTTPClient(srv.Client()))
	require.NoError(t, err)
	w := api.NewWrapperFromAPI(apiClient)
	cfg := config.Config{}
	cfg.Cache.Enabled = true
	svc := NewItemsService(w, cm, cfg)
	_, err = svc.List(context.Background(), &apigen.GetItemsParams{})
	require.NoError(t, err)
	pj, _, _, _, gerr := cm.Get("items:list:type=;s=;limit=0;offset=0")
	require.NoError(t, gerr)
	require.Equal(t, []byte(`{"items":[{"title":"t"}],"limit":1,"offset":0,"total":1}`), pj)
}

func TestItemsService_List_NoFallbackOnStatusError(t *testing.T) {
	dir := t.TempDir()
	opts := cache.Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 5,
		KeyringConfig: keyring.Config{
			AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
			FileDir:          dir,
			FilePasswordFunc: func(string) (string, error) { return "pw", nil },
			ServiceName:      "sufir-keeper-client",
		},
	}
	cm, err := cache.New(opts)
	require.NoError(t, err)
	defer func() { _ = cm.Close() }()
	require.NoError(t, cm.Put("items:list:type=;s=;limit=0;offset=0", []byte(`{"items":[],"limit":0,"offset":0,"total":0}`), nil, ""))
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401,"message":"unauthorized"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	apiClient, err := apigen.NewClientWithResponses(srv.URL, apigen.WithHTTPClient(srv.Client()))
	require.NoError(t, err)
	w := api.NewWrapperFromAPI(apiClient)
	cfg := config.Config{}
	cfg.Cache.Enabled = true
	svc := NewItemsService(w, cm, cfg)
	_, err = svc.List(context.Background(), &apigen.GetItemsParams{})
	require.Error(t, err)
}

func TestItemsService_List_NoFallbackOn500(t *testing.T) {
	dir := t.TempDir()
	opts := cache.Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 5,
		KeyringConfig: keyring.Config{
			AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
			FileDir:          dir,
			FilePasswordFunc: func(string) (string, error) { return "pw", nil },
			ServiceName:      "sufir-keeper-client",
		},
	}
	cm, err := cache.New(opts)
	require.NoError(t, err)
	defer func() { _ = cm.Close() }()
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	apiClient, err := apigen.NewClientWithResponses(srv.URL, apigen.WithHTTPClient(srv.Client()))
	require.NoError(t, err)
	w := api.NewWrapperFromAPI(apiClient)
	cfg := config.Config{}
	cfg.Cache.Enabled = true
	svc := NewItemsService(w, cm, cfg)
	_, err = svc.List(context.Background(), &apigen.GetItemsParams{})
	require.Error(t, err)
}

func TestItemsService_List_TTLExpired_NoFallback(t *testing.T) {
	dir := t.TempDir()
	opts := cache.Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 1,
		KeyringConfig: keyring.Config{
			AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
			FileDir:          dir,
			FilePasswordFunc: func(string) (string, error) { return "pw", nil },
			ServiceName:      "sufir-keeper-client",
		},
	}
	cm, err := cache.New(opts)
	require.NoError(t, err)
	defer func() { _ = cm.Close() }()
	require.NoError(t, cm.PutWithTimestamp("items:list:type=;s=;limit=0;offset=0", []byte(`{"items":[]}`), nil, "", time.Now().Add(-3*time.Minute)))
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	apiClient, err := apigen.NewClientWithResponses(srv.URL, apigen.WithHTTPClient(srv.Client()))
	require.NoError(t, err)
	w := api.NewWrapperFromAPI(apiClient)
	cfg := config.Config{}
	cfg.Cache.Enabled = true
	svc := NewItemsService(w, cm, cfg)
	_, err = svc.List(context.Background(), &apigen.GetItemsParams{})
	require.Error(t, err)
}

func openapiUUIDFromString(t *testing.T, s string) openapi_types.UUID {
	t.Helper()
	var rr httptest.ResponseRecorder
	_ = rr
	var u openapi_types.UUID
	if err := u.UnmarshalText([]byte(s)); err != nil {
		t.Fatal(err)
	}
	return u
}
