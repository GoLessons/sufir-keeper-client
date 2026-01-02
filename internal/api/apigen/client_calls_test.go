package apigen

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestClientCalls_RequestBuilding(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"items":[],"total":0,"limit":0,"offset":0}`))
			return
		}
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"title":"t"}`))
			return
		}
	})
	mux.HandleFunc("/items/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"title":"t"}`))
			return
		}
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"title":"t2"}`))
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	})
	mux.HandleFunc("/files/presign", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"k","upload_url":"http://` + r.Host + `/upload","form_fields":{}}`))
	})
	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	})
	mux.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte("bin"))
			return
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	c, err := NewClientWithResponses(srv.URL, WithHTTPClient(srv.Client()))
	if err != nil {
		t.Fatal(err)
	}
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"a","refresh_token":"r","token_type":"bearer","expires_in":3600}`))
			return
		}
		if r.Method == http.MethodPatch {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"a2","refresh_token":"r2","token_type":"bearer","expires_in":3600}`))
			return
		}
	})
	mux.HandleFunc("/auth-verify", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"message":"ok"}`))
			return
		}
	})
	pt := ItemType("TEXT")
	s := "x"
	limit := 10
	offset := 1
	_, err = c.GetItemsWithResponse(t.Context(), &GetItemsParams{Type: &pt, S: &s, Limit: &limit, Offset: &offset})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetItemsWithResponse(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var id openapi_types.UUID
	if err := id.UnmarshalText([]byte("00000000-0000-0000-0000-000000000001")); err != nil {
		t.Fatal(err)
	}
	_, err = c.GetItemWithResponse(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	var d ItemCreate_Data
	if err := d.UnmarshalJSON([]byte(`{"type":"TEXT","value":"v"}`)); err != nil {
		t.Fatal(err)
	}
	_, err = c.CreateItemWithResponse(t.Context(), ItemCreate{Title: "t", Data: d})
	if err != nil {
		t.Fatal(err)
	}
	var ud ItemUpdate_Data
	if err := ud.UnmarshalJSON([]byte(`{"type":"TEXT","value":"v2"}`)); err != nil {
		t.Fatal(err)
	}
	_, err = c.UpdateItemWithResponse(t.Context(), id, ItemUpdate{Data: &ud})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.DeleteItemWithResponse(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.PresignFileWithResponse(t.Context(), PresignFileJSONRequestBody{FileId: id})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.UploadFileWithBodyWithResponse(t.Context(), nil, "application/octet-stream", bytes.NewReader([]byte("x")))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.DownloadFileWithResponse(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.LoginUserWithResponse(t.Context(), LoginUserJSONRequestBody{Login: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.RefreshTokenWithResponse(t.Context(), RefreshTokenJSONRequestBody{RefreshToken: "r"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.RegisterUserWithResponse(t.Context(), RegisterUserJSONRequestBody{Login: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.AuthVerifyGetWithResponse(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.AuthVerifyPostWithResponse(t.Context())
	if err != nil {
		t.Fatal(err)
	}
}
