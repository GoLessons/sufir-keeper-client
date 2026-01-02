package apigen

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func resp(status int, contentType string, body []byte) *http.Response {
	rr := httptest.NewRecorder()
	rr.Header().Set("Content-Type", contentType)
	rr.WriteHeader(status)
	_, _ = rr.Write(body)
	return rr.Result()
}

func TestParse_GetItem_404(t *testing.T) {
	r := resp(404, "text/plain", []byte(""))
	out, err := ParseGetItemResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_DownloadFile_403_404(t *testing.T) {
	r403 := resp(403, "text/plain", []byte(""))
	out403, err := ParseDownloadFileResponse(r403)
	if err != nil {
		t.Fatal(err)
	}
	if out403.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
	r404 := resp(404, "text/plain", []byte(""))
	out404, err := ParseDownloadFileResponse(r404)
	if err != nil {
		t.Fatal(err)
	}
	if out404.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_UploadFile_400(t *testing.T) {
	r := resp(400, "text/plain", []byte(""))
	out, err := ParseUploadFileResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_PresignFile_400(t *testing.T) {
	r := resp(400, "text/plain", []byte(""))
	out, err := ParsePresignFileResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_RefreshToken_401(t *testing.T) {
	r := resp(401, "application/json", []byte(`{"code":401}`))
	out, err := ParseRefreshTokenResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_AuthVerifyGet_401(t *testing.T) {
	r := resp(401, "application/json", []byte(`{"code":401,"error":"Unauthorized","message":"unauthorized"}`))
	out, err := ParseAuthVerifyGetResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}

func TestParse_AuthVerifyPost_401(t *testing.T) {
	r := resp(401, "application/json", []byte(`{"code":401}`))
	out, err := ParseAuthVerifyPostResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}

func TestParse_LogoutUser_401(t *testing.T) {
	r := resp(401, "application/json", []byte(`{"code":401}`))
	out, err := ParseLogoutUserResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}

func TestParse_DownloadFile_401(t *testing.T) {
	r := resp(401, "application/json", []byte(`{"code":401}`))
	out, err := ParseDownloadFileResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}

func TestParse_DownloadFile_200(t *testing.T) {
	r := resp(200, "application/octet-stream", []byte("x"))
	out, err := ParseDownloadFileResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_AuthVerifyGet_204(t *testing.T) {
	r := resp(204, "text/plain", []byte(""))
	out, err := ParseAuthVerifyGetResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_AuthVerifyPost_204(t *testing.T) {
	r := resp(204, "text/plain", []byte(""))
	out, err := ParseAuthVerifyPostResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_LogoutUser_204(t *testing.T) {
	r := resp(204, "text/plain", []byte(""))
	out, err := ParseLogoutUserResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_UploadFile_204(t *testing.T) {
	r := resp(204, "text/plain", []byte(""))
	out, err := ParseUploadFileResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_CreateItem_201(t *testing.T) {
	r := resp(201, "application/json", []byte(`{"title":"t"}`))
	out, err := ParseCreateItemResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON201 == nil {
		t.Fatal("expected JSON201")
	}
}

func TestParse_CreateItem_401(t *testing.T) {
	r := resp(401, "application/json", []byte(`{"code":401}`))
	out, err := ParseCreateItemResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}

func TestParse_UpdateItem_200_401(t *testing.T) {
	r200 := resp(200, "application/json", []byte(`{"title":"t"}`))
	out200, err := ParseUpdateItemResponse(r200)
	if err != nil {
		t.Fatal(err)
	}
	if out200.JSON200 == nil {
		t.Fatal("expected JSON200")
	}
	r401 := resp(401, "application/json", []byte(`{"code":401}`))
	out401, err := ParseUpdateItemResponse(r401)
	if err != nil {
		t.Fatal(err)
	}
	if out401.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}

func TestParse_PresignFile_200(t *testing.T) {
	r := resp(200, "application/json", []byte(`{"key":"k","upload_url":"u","form_fields":{}}`))
	out, err := ParsePresignFileResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON200 == nil {
		t.Fatal("expected JSON200")
	}
}

func TestParse_DeleteItem_401(t *testing.T) {
	r := resp(401, "application/json", []byte(`{"code":401,"error":"Unauthorized","message":"unauthorized"}`))
	out, err := ParseDeleteItemResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}

func TestParse_DeleteItem_204(t *testing.T) {
	r := resp(204, "text/plain", []byte(""))
	out, err := ParseDeleteItemResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_RefreshToken_200(t *testing.T) {
	r := resp(200, "application/json", []byte(`{"access_token":"a","refresh_token":"r","token_type":"bearer","expires_in":3600}`))
	out, err := ParseRefreshTokenResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON200 == nil {
		t.Fatal("expected JSON200")
	}
}

func TestParse_RegisterUser_201(t *testing.T) {
	r := resp(201, "application/json", []byte(`{"message":"ok"}`))
	out, err := ParseRegisterUserResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON201 == nil {
		t.Fatal("expected JSON201")
	}
}

func TestParse_RegisterUser_409(t *testing.T) {
	r := resp(409, "text/plain", []byte(""))
	out, err := ParseRegisterUserResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_LoginUser_200_401(t *testing.T) {
	r200 := resp(200, "application/json", []byte(`{"access_token":"a","refresh_token":"r","token_type":"bearer","expires_in":3600}`))
	out200, err := ParseLoginUserResponse(r200)
	if err != nil {
		t.Fatal(err)
	}
	if out200.JSON200 == nil {
		t.Fatal("expected JSON200")
	}
	r401 := resp(401, "application/json", []byte(`{"code":401}`))
	out401, err := ParseLoginUserResponse(r401)
	if err != nil {
		t.Fatal(err)
	}
	if out401.HTTPResponse == nil || out401.HTTPResponse.StatusCode != 401 {
		t.Fatal("expected 401")
	}
}

func TestParse_UploadFile_409(t *testing.T) {
	r := resp(409, "text/plain", []byte(""))
	out, err := ParseUploadFileResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_UploadFile_401(t *testing.T) {
	r := resp(401, "application/json", []byte(`{"code":401,"error":"Unauthorized","message":"unauthorized"}`))
	out, err := ParseUploadFileResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}

func TestParse_GetItems_200_401(t *testing.T) {
	r200 := resp(200, "application/json", []byte(`{"items":[],"total":0,"limit":0,"offset":0}`))
	out200, err := ParseGetItemsResponse(r200)
	if err != nil {
		t.Fatal(err)
	}
	if out200.JSON200 == nil {
		t.Fatal("expected JSON200")
	}
	r401 := resp(401, "application/json", []byte(`{"code":401}`))
	out401, err := ParseGetItemsResponse(r401)
	if err != nil {
		t.Fatal(err)
	}
	if out401.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}

func TestParse_GetItems_404(t *testing.T) {
	r := resp(404, "text/plain", []byte(""))
	out, err := ParseGetItemsResponse(r)
	if err != nil {
		t.Fatal(err)
	}
	if out.HTTPResponse == nil {
		t.Fatal("expected HTTPResponse")
	}
}

func TestParse_GetItem_200_401(t *testing.T) {
	r200 := resp(200, "application/json", []byte(`{"title":"t"}`))
	out200, err := ParseGetItemResponse(r200)
	if err != nil {
		t.Fatal(err)
	}
	if out200.JSON200 == nil {
		t.Fatal("expected JSON200")
	}
	r401 := resp(401, "application/json", []byte(`{"code":401}`))
	out401, err := ParseGetItemResponse(r401)
	if err != nil {
		t.Fatal(err)
	}
	if out401.JSON401 == nil {
		t.Fatal("expected JSON401")
	}
}
