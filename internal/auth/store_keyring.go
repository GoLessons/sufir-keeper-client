package auth

import (
	"errors"
	"sync"

	"github.com/99designs/keyring"
)

type TokenStore interface {
	SaveTokens(AuthTokens) error
	LoadTokens() (AuthTokens, error)
	Clear() error
	CurrentAccessToken() (string, bool)
	HasRefreshToken() bool
}

type KeyringStore struct {
	ring        keyring.Keyring
	serviceName string
	accessKey   string
	refreshKey  string
	cached      AuthTokens
	mu          sync.Mutex
}

type KeyringOptions struct {
	ServiceName  string
	Backend      string
	FileDir      string
	AccessKey    string
	RefreshKey   string
	FilePassword string
}

func NewKeyringStore(opts KeyringOptions) (*KeyringStore, error) {
	cfg := keyring.Config{
		ServiceName: opts.ServiceName,
	}
	if opts.Backend == "file" {
		cfg.AllowedBackends = []keyring.BackendType{keyring.FileBackend}
		cfg.FileDir = opts.FileDir
		pass := opts.FilePassword
		if pass == "" {
			pass = "sufir-keeper-dev"
		}
		cfg.FilePasswordFunc = func(prompt string) (string, error) { return pass, nil }
	}
	r, err := keyring.Open(cfg)
	if err != nil {
		return nil, err
	}
	accessKey := opts.AccessKey
	if accessKey == "" {
		accessKey = "access_token"
	}
	refreshKey := opts.RefreshKey
	if refreshKey == "" {
		refreshKey = "refresh_token"
	}
	return &KeyringStore{
		ring:        r,
		serviceName: opts.ServiceName,
		accessKey:   accessKey,
		refreshKey:  refreshKey,
	}, nil
}

func (s *KeyringStore) SaveTokens(t AuthTokens) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.AccessToken == "" || t.RefreshToken == "" {
		return errors.New("tokens required")
	}
	if err := s.ring.Set(keyring.Item{Key: s.accessKey, Data: []byte(t.AccessToken)}); err != nil {
		return err
	}
	if err := s.ring.Set(keyring.Item{Key: s.refreshKey, Data: []byte(t.RefreshToken)}); err != nil {
		return err
	}
	s.cached = t
	return nil
}

func (s *KeyringStore) LoadTokens() (AuthTokens, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached.AccessToken != "" && s.cached.RefreshToken != "" {
		return s.cached, nil
	}
	access, err := s.ring.Get(s.accessKey)
	if err != nil {
		return AuthTokens{}, err
	}
	refresh, err := s.ring.Get(s.refreshKey)
	if err != nil {
		return AuthTokens{}, err
	}
	s.cached = AuthTokens{
		AccessToken:  string(access.Data),
		RefreshToken: string(refresh.Data),
	}
	return s.cached, nil
}

func (s *KeyringStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.ring.Remove(s.accessKey)
	_ = s.ring.Remove(s.refreshKey)
	s.cached = AuthTokens{}
	return nil
}

func (s *KeyringStore) CurrentAccessToken() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached.AccessToken != "" {
		return s.cached.AccessToken, true
	}
	item, err := s.ring.Get(s.accessKey)
	if err != nil {
		return "", false
	}
	token := string(item.Data)
	if token == "" {
		return "", false
	}
	s.cached.AccessToken = token
	return token, true
}

func (s *KeyringStore) HasRefreshToken() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached.RefreshToken != "" {
		return true
	}
	_, err := s.ring.Get(s.refreshKey)
	return err == nil
}
