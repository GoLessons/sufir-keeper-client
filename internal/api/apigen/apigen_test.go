package apigen

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientConstructionAndOptions(t *testing.T) {
	base := "http://localhost:8080/api/v1"
	c, err := NewClientWithResponses(base)
	require.NoError(t, err)
	require.NotNil(t, c)
	c2, err := NewClientWithResponses(base, WithHTTPClient(&http.Client{}))
	require.NoError(t, err)
	require.NotNil(t, c2)
}

func TestParseResponsesJSONBranches(t *testing.T) {
	jsonHdr := http.Header{}
	jsonHdr.Set("Content-Type", "application/json")
	makeResp := func(code int, body string) *http.Response {
		return &http.Response{
			StatusCode: code,
			Header:     jsonHdr.Clone(),
			Body:       ioNopCloser(bytes.NewBufferString(body)),
		}
	}
	r1, err := ParseLoginUserResponse(makeResp(200, `{"access_token":"a","refresh_token":"r","token_type":"bearer","expires_in":1}`))
	require.NoError(t, err)
	require.Equal(t, 200, r1.StatusCode())
	r2, err := ParseRefreshTokenResponse(makeResp(200, `{"access_token":"a2","refresh_token":"r2","token_type":"bearer","expires_in":2}`))
	require.NoError(t, err)
	require.Equal(t, 200, r2.StatusCode())
	r3, err := ParseAuthVerifyGetResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r3.StatusCode())
	r4, err := ParseAuthVerifyPostResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r4.StatusCode())
	r5, err := ParsePresignFileResponse(makeResp(200, `{"upload_url":"u","key":"k","form_fields":{"a":"b"}}`))
	require.NoError(t, err)
	require.Equal(t, 200, r5.StatusCode())
	r6, err := ParsePresignFileResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r6.StatusCode())
	r7, err := ParseDownloadFileResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r7.StatusCode())
	r8, err := ParseGetItemsResponse(makeResp(200, `{"items":[{"id":"123e4567-e89b-12d3-a456-426614174000","title":"t"}],"total":1,"limit":1,"offset":0}`))
	require.NoError(t, err)
	require.Equal(t, 200, r8.StatusCode())
	r9, err := ParseGetItemsResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r9.StatusCode())
	r10, err := ParseCreateItemResponse(makeResp(201, `{"id":"123e4567-e89b-12d3-a456-426614174000","title":"t"}`))
	require.NoError(t, err)
	require.Equal(t, 201, r10.StatusCode())
	r11, err := ParseCreateItemResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r11.StatusCode())
	r12, err := ParseDeleteItemResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r12.StatusCode())
	r13, err := ParseGetItemResponse(makeResp(200, `{"id":"123e4567-e89b-12d3-a456-426614174000","title":"t"}`))
	require.NoError(t, err)
	require.Equal(t, 200, r13.StatusCode())
	r14, err := ParseGetItemResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r14.StatusCode())
	r15, err := ParseUpdateItemResponse(makeResp(200, `{"id":"123e4567-e89b-12d3-a456-426614174000","title":"t"}`))
	require.NoError(t, err)
	require.Equal(t, 200, r15.StatusCode())
	r16, err := ParseUpdateItemResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r16.StatusCode())
	r17, err := ParseLogoutUserResponse(makeResp(401, `{"error":"Unauthorized"}`))
	require.NoError(t, err)
	require.Equal(t, 401, r17.StatusCode())
}

type nopCloser struct{ *bytes.Buffer }

func (n nopCloser) Close() error { return nil }

func ioNopCloser(b *bytes.Buffer) nopCloser { return nopCloser{b} }
