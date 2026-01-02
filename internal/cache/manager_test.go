package cache

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/stretchr/testify/require"
)

func TestCachePutGetFresh(t *testing.T) {
	dir := t.TempDir()
	opts := Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 5,
		KeyringConfig: keyring.Config{
			AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
			FileDir:          dir,
			FilePasswordFunc: func(string) (string, error) { return "pw", nil },
			ServiceName:      "sufir-keeper-client",
		},
	}
	m, err := New(opts)
	require.NoError(t, err)
	defer func() { _ = m.Close() }()

	payloadJSON := []byte(`{"ok":true}`)
	payload := []byte("secret")
	require.NoError(t, m.Put("k1", payloadJSON, payload, "m"))

	pj, pl, ts, meta, err := m.Get("k1")
	require.NoError(t, err)
	require.Equal(t, payloadJSON, pj)
	require.Equal(t, payload, pl)
	require.Equal(t, "m", meta)
	require.True(t, m.IsFresh(ts))
}

func TestCacheExpired(t *testing.T) {
	dir := t.TempDir()
	opts := Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 1,
		KeyringConfig: keyring.Config{
			AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
			FileDir:          dir,
			FilePasswordFunc: func(string) (string, error) { return "pw", nil },
			ServiceName:      "sufir-keeper-client",
		},
	}
	m, err := New(opts)
	require.NoError(t, err)
	defer func() { _ = m.Close() }()

	require.NoError(t, m.PutWithTimestamp("k2", []byte(`{}`), []byte("x"), "", time.Now().Add(-2*time.Minute)))
	_, _, ts, _, err := m.Get("k2")
	require.NoError(t, err)
	require.False(t, m.IsFresh(ts))
}

func TestCacheDelete(t *testing.T) {
	dir := t.TempDir()
	opts := Options{
		Path:       filepath.Join(dir, "cache.db"),
		TTLMinutes: 10,
		KeyringConfig: keyring.Config{
			AllowedBackends:  []keyring.BackendType{keyring.FileBackend},
			FileDir:          dir,
			FilePasswordFunc: func(string) (string, error) { return "pw", nil },
			ServiceName:      "sufir-keeper-client",
		},
	}
	m, err := New(opts)
	require.NoError(t, err)
	defer func() { _ = m.Close() }()

	require.NoError(t, m.Put("k3", []byte(`{}`), []byte("v"), ""))
	require.NoError(t, m.Delete("k3"))
	_, _, _, _, err = m.Get("k3")
	require.Error(t, err)
}
