package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/sync/singleflight"

	"github.com/GoLessons/sufir-keeper-client/internal/api/apigen"
	"github.com/GoLessons/sufir-keeper-client/internal/api/apiutil"
)

func DerefStringPointer(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func DerefIntPointer(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

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
}

func NewManager(client *retryablehttp.Client, store TokenStore) *Manager {
	return &Manager{client: client, store: store}
}

func (m *Manager) Register(ctx context.Context, baseURL, login, password string) error {
	body := apigen.UserRegister{Login: login, Password: password}
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return normalizeAPIError(resp)
	}
	return nil
}

func (m *Manager) Login(ctx context.Context, baseURL, login, password string) (AuthTokens, error) {
	body := apigen.UserLogin{Login: login, Password: password}
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
		defer resp.Body.Close()
		return AuthTokens{}, normalizeAPIError(resp)
	}
	defer resp.Body.Close()
	var ar apigen.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return AuthTokens{}, err
	}
	tokens := AuthTokens{
		AccessToken:  DerefStringPointer(ar.AccessToken),
		RefreshToken: DerefStringPointer(ar.RefreshToken),
		TokenType:    DerefStringPointer(ar.TokenType),
		ExpiresIn:    DerefIntPointer(ar.ExpiresIn),
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
		type rb struct {
			RefreshToken string `json:"refresh_token"`
		}
		body := rb{RefreshToken: tokens.RefreshToken}
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
			defer resp.Body.Close()
			return AuthTokens{}, normalizeAPIError(resp)
		}
		defer resp.Body.Close()
		var ar apigen.AuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
			return AuthTokens{}, err
		}
		newTokens := AuthTokens{
			AccessToken:  DerefStringPointer(ar.AccessToken),
			RefreshToken: DerefStringPointer(ar.RefreshToken),
			TokenType:    DerefStringPointer(ar.TokenType),
			ExpiresIn:    DerefIntPointer(ar.ExpiresIn),
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
		defer resp.Body.Close()
		return UserInfo{}, normalizeAPIError(resp)
	}
	userID := resp.Header.Get("X-User-Id")
	if userID == "" {
		return UserInfo{}, errors.New("no user id")
	}
	return UserInfo{UserID: userID}, nil
}

func normalizeAPIError(resp *http.Response) error {
	status := resp.StatusCode
	var serverErr apigen.Error
	var body []byte
	if resp.Body != nil {
		b, _ := io.ReadAll(resp.Body)
		body = b
	}
	var msg string
	if len(body) > 0 && strings.Contains(resp.Header.Get("Content-Type"), "json") {
		if json.Unmarshal(body, &serverErr) == nil {
			if serverErr.Message != nil && *serverErr.Message != "" {
				msg = *serverErr.Message
			}
		}
	}
	if msg == "" {
		if status > 0 {
			msg = http.StatusText(status)
		} else {
			msg = "request failed"
		}
	}
	return apiutil.Error{Status: status, Message: msg}
}
