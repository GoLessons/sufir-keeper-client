package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/99designs/keyring"
	"github.com/stretchr/testify/require"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/GoLessons/sufir-keeper-client/internal/api"
	"github.com/GoLessons/sufir-keeper-client/internal/api/apigen"
	"github.com/GoLessons/sufir-keeper-client/internal/cache"
	"github.com/GoLessons/sufir-keeper-client/internal/config"
)

func TestInvalidation_Create_Update_Delete(t *testing.T) {
	dir := t.TempDir()
	opts := cache.Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 10,
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
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"title":"created"}`))
			return
		}
	})
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			_, _ = w.Write([]byte(`{"title":"updated"}`))
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	apiClient, err := apigen.NewClientWithResponses(srv.URL, apigen.WithHTTPClient(srv.Client()))
	require.NoError(t, err)
	w := api.NewWrapperFromAPI(apiClient)
	cfg := config.Config{}
	cfg.Cache.Enabled = true
	svc := NewItemsService(w, cm, cfg)

	// Seed cache entries
	require.NoError(t, cm.Put("items:list:type=;s=;limit=0;offset=0", []byte(`{"items":[]}`), nil, ""))
	require.NoError(t, cm.Put("items:get:00000000-0000-0000-0000-000000000001", []byte(`{"title":"old"}`), nil, ""))

	// Create invalidates list
	{
		var d apigen.ItemCreate_Data
		require.NoError(t, d.FromTextData(apigen.TextData{Type: "TEXT", Value: "v"}))
		_, err := svc.Create(context.Background(), apigen.ItemCreate{Title: "t", Data: d})
		require.NoError(t, err)
		_, _, _, _, gerr := cm.Get("items:list:type=;s=;limit=0;offset=0")
		require.Error(t, gerr)
	}

	// Update invalidates list and get
	{
		var id openapi_types.UUID
		require.NoError(t, id.UnmarshalText([]byte("00000000-0000-0000-0000-000000000001")))
		_, err := svc.Update(context.Background(), id, apigen.ItemUpdate{})
		require.NoError(t, err)
		_, _, _, _, gerr := cm.Get("items:get:00000000-0000-0000-0000-000000000001")
		require.Error(t, gerr)
	}

	// Put back entries and test Delete invalidation
	require.NoError(t, cm.Put("items:list:type=;s=;limit=0;offset=0", []byte(`{"items":[]}`), nil, ""))
	require.NoError(t, cm.Put("items:get:00000000-0000-0000-0000-000000000001", []byte(`{"title":"old"}`), nil, ""))
	{
		var id openapi_types.UUID
		require.NoError(t, id.UnmarshalText([]byte("00000000-0000-0000-0000-000000000001")))
		_, err := svc.Delete(context.Background(), id)
		require.NoError(t, err)
		_, _, _, _, gerr := cm.Get("items:list:type=;s=;limit=0;offset=0")
		require.Error(t, gerr)
		_, _, _, _, gerr2 := cm.Get("items:get:00000000-0000-0000-0000-000000000001")
		require.Error(t, gerr2)
	}
}
