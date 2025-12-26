package apiutil

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

type errRoundTripper struct{}

func (e errRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("transport error")
}

func TestDoHandlesUnauthorizedForbiddenAndSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/u":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
		case "/f":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"Forbidden"}`))
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	t.Cleanup(srv.Close)
	rc := retryablehttp.NewClient()
	rc.RetryMax = 0
	rc.HTTPClient = srv.Client()

	req1, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/u", nil)
	require.NoError(t, err)
	resp1, err := Do(rc, req1)
	require.Error(t, err)
	var apiErr Error
	require.True(t, errors.As(err, &apiErr))
	require.Equal(t, http.StatusUnauthorized, apiErr.Status)
	require.NotNil(t, resp1)

	req2, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/f", nil)
	require.NoError(t, err)
	resp2, err := Do(rc, req2)
	require.Error(t, err)
	require.True(t, errors.As(err, &apiErr))
	require.Equal(t, http.StatusForbidden, apiErr.Status)
	require.NotNil(t, resp2)

	req3, err := retryablehttp.NewRequest(http.MethodGet, srv.URL+"/ok", nil)
	require.NoError(t, err)
	resp3, err := Do(rc, req3)
	require.NoError(t, err)
	require.NotNil(t, resp3)
	require.Equal(t, http.StatusOK, resp3.StatusCode)
}

func TestDoPropagatesTransportErrors(t *testing.T) {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 0
	rc.HTTPClient = &http.Client{Transport: errRoundTripper{}}
	req, err := retryablehttp.NewRequest(http.MethodGet, "http://example.invalid/", nil)
	require.NoError(t, err)
	resp, err := Do(rc, req)
	require.Error(t, err)
	require.Nil(t, resp)
}
