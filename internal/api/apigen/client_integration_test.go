package apigen

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestClientWithResponsesCoversEndpoints(t *testing.T) {
	const (
		pathAuth         = "/auth"
		pathAuthVerify   = "/auth-verify"
		pathFiles        = "/files"
		pathFilesPresign = "/files/presign"
		pathFileByID     = "/files/00000000-0000-0000-0000-000000000000"
		pathItems        = "/items"
		pathItemGet      = "/items/00000000-0000-0000-0000-000000000001"
		pathItemUpdate   = "/items/00000000-0000-0000-0000-000000000002"
		pathItemDelete   = "/items/00000000-0000-0000-0000-000000000003"
		pathRegister     = "/register"
	)
	type jsonResp map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete && r.URL.Path == pathAuth:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPatch && r.URL.Path == pathAuth:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(jsonResp{
				"access_token":  "a",
				"refresh_token": "r",
				"token_type":    "bearer",
				"expires_in":    1,
			})
		case r.Method == http.MethodPost && r.URL.Path == pathAuth:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(jsonResp{
				"access_token":  "a2",
				"refresh_token": "r2",
				"token_type":    "bearer",
				"expires_in":    2,
			})
		case r.Method == http.MethodGet && r.URL.Path == pathAuthVerify:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(jsonResp{"error": "Unauthorized"})
		case r.Method == http.MethodPost && r.URL.Path == pathAuthVerify:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(jsonResp{"error": "Unauthorized"})
		case r.Method == http.MethodPost && r.URL.Path == pathFiles:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == pathFilesPresign:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(jsonResp{
				"upload_url": "https://upload",
				"key":        "k",
				"form_fields": map[string]string{
					"a": "b",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == pathFileByID:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(jsonResp{"error": "Unauthorized"})
		case r.Method == http.MethodGet && r.URL.Path == pathItems:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(jsonResp{
				"items":  []jsonResp{{"id": uuid.New().String(), "title": "t"}},
				"total":  1,
				"limit":  1,
				"offset": 0,
			})
		case r.Method == http.MethodPost && r.URL.Path == pathItems:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(jsonResp{
				"id":         uuid.New().String(),
				"title":      "t",
				"created_at": time.Now().UTC(),
			})
		case r.Method == http.MethodGet && r.URL.Path == pathItemGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(jsonResp{
				"id":    uuid.New().String(),
				"title": "t",
			})
		case r.Method == http.MethodPut && r.URL.Path == pathItemUpdate:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(jsonResp{
				"id":    uuid.New().String(),
				"title": "t2",
			})
		case r.Method == http.MethodDelete && r.URL.Path == pathItemDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == pathRegister:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(jsonResp{"message": "ok"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	c, err := NewClientWithResponses(srv.URL, WithHTTPClient(srv.Client()))
	require.NoError(t, err)
	_, err = c.LoginUserWithResponse(context.Background(), LoginUserJSONRequestBody{Login: "u", Password: "p"})
	require.NoError(t, err)
	_, err = c.RefreshTokenWithResponse(context.Background(), RefreshTokenJSONRequestBody{RefreshToken: "r"})
	require.NoError(t, err)
	_, err = c.AuthVerifyGetWithResponse(context.Background())
	require.NoError(t, err)
	_, err = c.AuthVerifyPostWithResponse(context.Background())
	require.NoError(t, err)
	_, err = c.PresignFileWithResponse(context.Background(), PresignFileJSONRequestBody{FileId: uuid.Nil})
	require.NoError(t, err)
	_, err = c.DownloadFileWithResponse(context.Background(), uuid.Nil)
	require.NoError(t, err)
	_, err = c.GetItemsWithResponse(context.Background(), &GetItemsParams{Limit: intPtr(1), Offset: intPtr(0)})
	require.NoError(t, err)
	body := CreateItemJSONRequestBody{Title: "t", Data: ItemCreate_Data{}}
	_ = body.Data.FromTextData(TextData{Type: "TEXT", Value: "v"})
	_, err = c.CreateItemWithResponse(context.Background(), body)
	require.NoError(t, err)
	_, err = c.GetItemWithResponse(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	require.NoError(t, err)
	upd := UpdateItemJSONRequestBody{Data: &ItemUpdate_Data{}}
	_ = upd.Data.FromTextData(TextData{Type: "TEXT", Value: "v2"})
	_, err = c.UpdateItemWithResponse(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000002"), upd)
	require.NoError(t, err)
	_, err = c.DeleteItemWithResponse(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000003"))
	require.NoError(t, err)
	_, err = c.RegisterUserWithResponse(context.Background(), RegisterUserJSONRequestBody{Login: "u2", Password: "p2"})
	require.NoError(t, err)

	req, err := NewUploadFileRequestWithBody(srv.URL, &UploadFileParams{XFileID: uuid.Nil}, "multipart/form-data", bytes.NewBufferString("x"))
	require.NoError(t, err)
	require.Equal(t, "POST", req.Method)
	require.Equal(t, "multipart/form-data", req.Header.Get("Content-Type"))
	require.NotEmpty(t, req.Header.Get("X-File-ID"))
}

func intPtr(v int) *int { return &v }
