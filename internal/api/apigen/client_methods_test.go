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

func TestClientMethodsInvokeEndpoints(t *testing.T) {
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
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == pathAuthVerify:
			w.WriteHeader(http.StatusNoContent)
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
			w.WriteHeader(http.StatusOK)
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
	cl, err := NewClient(srv.URL, WithHTTPClient(srv.Client()), WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("X-Edited", "1")
		return nil
	}))
	require.NoError(t, err)
	resp, err := cl.LogoutUser(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp, err = cl.RefreshToken(context.Background(), RefreshTokenJSONRequestBody{RefreshToken: "r"})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp, err = cl.LoginUser(context.Background(), LoginUserJSONRequestBody{Login: "u", Password: "p"})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp, err = cl.AuthVerifyGet(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp, err = cl.AuthVerifyPost(context.Background())
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp, err = cl.UploadFileWithBody(context.Background(), &UploadFileParams{XFileID: uuid.Nil}, "multipart/form-data", bytes.NewBufferString("x"))
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp, err = cl.PresignFileWithBody(context.Background(), "application/json", bytes.NewBufferString(`{"fileId":"00000000-0000-0000-0000-000000000000"}`))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp, err = cl.PresignFile(context.Background(), PresignFileJSONRequestBody{FileId: uuid.Nil})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp, err = cl.DownloadFile(context.Background(), uuid.Nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp, err = cl.GetItems(context.Background(), &GetItemsParams{})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	b := CreateItemJSONRequestBody{Title: "t", Data: ItemCreate_Data{}}
	_ = b.Data.FromTextData(TextData{Type: "TEXT", Value: "v"})
	resp, err = cl.CreateItem(context.Background(), b)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp, err = cl.CreateItemWithBody(context.Background(), "application/json", bytes.NewBufferString(`{"title":"t","data":{"type":"TEXT","value":"v"}}`))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp, err = cl.GetItem(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	upd := UpdateItemJSONRequestBody{Data: &ItemUpdate_Data{}}
	_ = upd.Data.FromTextData(TextData{Type: "TEXT", Value: "v2"})
	resp, err = cl.UpdateItem(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000002"), upd)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp, err = cl.UpdateItemWithBody(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000002"), "application/json", bytes.NewBufferString(`{"data":{"type":"TEXT","value":"v2"}}`))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp, err = cl.DeleteItem(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000003"))
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp, err = cl.RegisterUser(context.Background(), RegisterUserJSONRequestBody{Login: "u2", Password: "p2"})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}
