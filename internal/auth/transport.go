package auth

import (
	"bytes"
	"io"
	"net/http"
)

type AuthRoundTripper struct {
	base    http.RoundTripper
	store   TokenStore
	manager *Manager
	baseURL string
}

func NewAuthRoundTripper(base http.RoundTripper, manager *Manager, baseURL string, store TokenStore) *AuthRoundTripper {
	return &AuthRoundTripper{
		base:    base,
		manager: manager,
		baseURL: baseURL,
		store:   store,
	}
}

func (t *AuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	access, has := t.store.CurrentAccessToken()
	if has && access != "" {
		req.Header.Set("Authorization", "Bearer "+access)
	}
	var bodyBytes []byte
	var canReplay bool
	if req.Body != nil && req.GetBody != nil {
		rc, err := req.GetBody()
		if err == nil {
			defer rc.Close()
			bodyBytes, _ = io.ReadAll(rc)
			canReplay = true
		}
	}
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		_, rerr := t.manager.Refresh(req.Context(), t.baseURL)
		if rerr != nil {
			return resp, nil
		}
		nreq := req.Clone(req.Context())
		if canReplay {
			nreq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		} else {
			if req.Body != nil {
				return resp, nil
			}
		}
		access2, ok := t.store.CurrentAccessToken()
		if ok && access2 != "" {
			nreq.Header.Set("Authorization", "Bearer "+access2)
		}
		return t.base.RoundTrip(nreq)
	}
	return resp, nil
}
