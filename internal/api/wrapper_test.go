package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/GoLessons/sufir-keeper-client/internal/api/apigen"
	"github.com/GoLessons/sufir-keeper-client/internal/api/apiutil"
)

func newApigen(t *testing.T, srv *httptest.Server) *apigen.ClientWithResponses {
	t.Helper()
	api, err := apigen.NewClientWithResponses(srv.URL, apigen.WithHTTPClient(srv.Client()))
	require.NoError(t, err)
	return api
}

func TestWrapper_GetItems_RetryOn500(t *testing.T) {
	mux := http.NewServeMux()
	var calls int
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			Items  []apigen.ItemListResponse `json:"items"`
			Limit  int                       `json:"limit"`
			Offset int                       `json:"offset"`
			Total  int                       `json:"total"`
		}{Items: []apigen.ItemListResponse{}, Limit: 0, Offset: 0, Total: 0}
		_ = json.NewEncoder(w).Encode(resp)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	w := NewWrapperFromAPI(newApigen(t, srv))
	w.retryMax = 3
	w.retryWaitMin = 10 * time.Millisecond
	w.retryWaitMax = 20 * time.Millisecond

	ctx := context.Background()
	resp, err := w.GetItems(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 3, calls)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestWrapper_GetItems_RetryAfter429(t *testing.T) {
	mux := http.NewServeMux()
	var calls int
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			Items  []apigen.ItemListResponse `json:"items"`
			Limit  int                       `json:"limit"`
			Offset int                       `json:"offset"`
			Total  int                       `json:"total"`
		}{Items: []apigen.ItemListResponse{}, Limit: 0, Offset: 0, Total: 0}
		_ = json.NewEncoder(w).Encode(resp)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	w := NewWrapperFromAPI(newApigen(t, srv))
	w.retryMax = 2
	w.retryWaitMin = 10 * time.Millisecond
	w.retryWaitMax = 20 * time.Millisecond

	start := time.Now()
	ctx := context.Background()
	resp, err := w.GetItems(ctx, nil)
	elapsed := time.Since(start)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, calls)
	require.GreaterOrEqual(t, int64(elapsed/time.Millisecond), int64(900))
}

func TestWrapper_Normalize401ToError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(apigen.Error{
			Code:    ptrInt(http.StatusUnauthorized),
			Error:   ptrStr("Unauthorized"),
			Message: ptrStr("unauthorized"),
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	w := NewWrapperFromAPI(newApigen(t, srv))
	ctx := context.Background()
	resp, err := w.GetItems(ctx, nil)
	require.Nil(t, resp)
	var aerr apiutil.Error
	require.ErrorAs(t, err, &aerr)
	require.Equal(t, http.StatusUnauthorized, aerr.Status)
	require.Equal(t, "unauthorized", aerr.Message)
}

func ptrInt(v int) *int       { return &v }
func ptrStr(v string) *string { return &v }

func TestWrapper_GetItem_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"title":"ok"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.GetItem(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestWrapper_CreateItem_201(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"title":"created"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	var d apigen.ItemCreate_Data
	_ = d.FromTextData(apigen.TextData{Type: "TEXT", Value: "v"})
	resp, err := w.CreateItem(context.Background(), apigen.ItemCreate{Title: "t", Data: d})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestWrapper_UpdateItem_200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"title":"updated"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.UpdateItem(context.Background(), id, apigen.ItemUpdate{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestWrapper_DeleteItem_204(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.DeleteItem(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestWrapper_DeleteItem_401(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401,"error":"Unauthorized","message":"unauthorized"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.DeleteItem(context.Background(), id)
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestWrapper_UpdateItem_401(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401,"error":"Unauthorized","message":"unauthorized"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.UpdateItem(context.Background(), id, apigen.ItemUpdate{})
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestWrapper_PresignFile_401(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/files/presign", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.PresignFile(context.Background(), apigen.PresignFileJSONRequestBody{FileId: id})
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestWrapper_AuthVerifyPost_204(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth-verify", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	resp, err := w.AuthVerifyPost(context.Background())
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestWrapper_LogoutUser_204(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	resp, err := w.LogoutUser(context.Background())
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestWrapper_DownloadFile_401(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.DownloadFile(context.Background(), id)
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestWrapper_DownloadFile_403(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":403}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.DownloadFile(context.Background(), id)
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestWrapper_UploadFileWithBody_401(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	resp, err := w.UploadFileWithBody(context.Background(), nil, "application/octet-stream", bytes.NewReader([]byte("x")))
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestWrapper_CreateItem_401(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	var d apigen.ItemCreate_Data
	_ = d.FromTextData(apigen.TextData{Type: "TEXT", Value: "v"})
	resp, err := w.CreateItem(context.Background(), apigen.ItemCreate{Title: "t", Data: d})
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestWrapper_DownloadFile_200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("x"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.DownloadFile(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestWrapper_GetItem_RetryOn500(t *testing.T) {
	mux := http.NewServeMux()
	var calls int
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"title":"ok"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.GetItem(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, calls)
}

func TestWrapper_GetItem_RetryAfter429(t *testing.T) {
	mux := http.NewServeMux()
	var calls int
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"title":"ok"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	start := time.Now()
	resp, err := w.GetItem(context.Background(), id)
	elapsed := time.Since(start)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, calls)
	require.GreaterOrEqual(t, int64(elapsed/time.Millisecond), int64(900))
}

func TestWrapper_AuthVerifyGet_401Error(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth-verify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":401,"error":"Unauthorized","message":"unauthorized"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	resp, err := w.AuthVerifyGet(context.Background())
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestWrapper_GetItems_Forbidden(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":403,"error":"Forbidden","message":"forbidden"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	resp, err := w.GetItems(context.Background(), nil)
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestWrapper_GetItems_200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[],"total":0,"limit":0,"offset":0}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	resp, err := w.GetItems(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestWrapper_PresignFile_200(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/files/presign", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"k","upload_url":"u","form_fields":{}}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	id := openapiUUID(t, "00000000-0000-0000-0000-000000000001")
	resp, err := w.PresignFile(context.Background(), apigen.PresignFileJSONRequestBody{FileId: id})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func openapiUUID(t *testing.T, s string) openapi_types.UUID {
	t.Helper()
	var u openapi_types.UUID
	_ = u.UnmarshalText([]byte(s))
	return u
}

func TestWrapper_UploadFileWithBody_204(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	api := newApigen(t, srv)
	w := NewWrapperFromAPI(api)
	resp, err := w.UploadFileWithBody(context.Background(), nil, "application/octet-stream", bytes.NewReader([]byte("x")))
	require.NoError(t, err)
	require.NotNil(t, resp)
}
