package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/sync/singleflight"
)

type Authenticator interface {
	Register(ctx context.Context, baseURL, login, password string) error
	Login(ctx context.Context, baseURL, login, password string) (AuthTokens, error)
	Refresh(ctx context.Context, baseURL string) (AuthTokens, error)
	Logout(ctx context.Context, baseURL string) error
	Verify(ctx context.Context, baseURL string) (UserInfo, error)
}

type Manager struct {
	client *retryablehttp.Client
	store  TokenStore
	group  singleflight.Group
	mu     sync.Mutex
}

func NewManager(client *retryablehttp.Client, store TokenStore) *Manager {
	return &Manager{client: client, store: store}
}

func (m *Manager) Register(ctx context.Context, baseURL, login, password string) error {
	body := map[string]string{"login": login, "password": password}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/register", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return errors.New("register failed")
	}
	return nil
}

func (m *Manager) Login(ctx context.Context, baseURL, login, password string) (AuthTokens, error) {
	body := map[string]string{"login": login, "password": password}
	data, err := json.Marshal(body)
	if err != nil {
		return AuthTokens{}, err
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/auth", bytes.NewReader(data))
	if err != nil {
		return AuthTokens{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		return AuthTokens{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return AuthTokens{}, errors.New("login failed")
	}
	defer resp.Body.Close()
	var tokens AuthTokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return AuthTokens{}, err
	}
	if err := m.store.SaveTokens(tokens); err != nil {
		return AuthTokens{}, err
	}
	return tokens, nil
}

func (m *Manager) Refresh(ctx context.Context, baseURL string) (AuthTokens, error) {
	res, err, _ := m.group.Do("refresh", func() (any, error) {
		tokens, err := m.store.LoadTokens()
		if err != nil {
			return AuthTokens{}, err
		}
		if tokens.RefreshToken == "" {
			return AuthTokens{}, errors.New("no refresh token")
		}
		body := map[string]string{"refresh_token": tokens.RefreshToken}
		data, err := json.Marshal(body)
		if err != nil {
			return AuthTokens{}, err
		}
		req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPatch, baseURL+"/auth", bytes.NewReader(data))
		if err != nil {
			return AuthTokens{}, err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := m.client.Do(req)
		if err != nil {
			return AuthTokens{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return AuthTokens{}, errors.New("refresh failed")
		}
		defer resp.Body.Close()
		var newTokens AuthTokens
		if err := json.NewDecoder(resp.Body).Decode(&newTokens); err != nil {
			return AuthTokens{}, err
		}
		if err := m.store.SaveTokens(newTokens); err != nil {
			return AuthTokens{}, err
		}
		return newTokens, nil
	})
	if err != nil {
		return AuthTokens{}, err
	}
	return res.(AuthTokens), nil
}

func (m *Manager) Logout(ctx context.Context, baseURL string) error {
	access, ok := m.store.CurrentAccessToken()
	if ok && access != "" {
		req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodDelete, baseURL+"/auth", nil)
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+access)
			_, _ = m.client.Do(req)
		}
	}
	return m.store.Clear()
}

func (m *Manager) Verify(ctx context.Context, baseURL string) (UserInfo, error) {
	access, ok := m.store.CurrentAccessToken()
	if !ok || access == "" {
		return UserInfo{}, errors.New("no access token")
	}
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/auth-verify", nil)
	if err != nil {
		return UserInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+access)
	resp, err := m.client.Do(req)
	if err != nil {
		return UserInfo{}, err
	}
	if resp.StatusCode != http.StatusNoContent {
		return UserInfo{}, errors.New("verify failed")
	}
	userID := resp.Header.Get("X-User-Id")
	if userID == "" {
		return UserInfo{}, errors.New("no user id")
	}
	return UserInfo{UserID: userID}, nil
}

