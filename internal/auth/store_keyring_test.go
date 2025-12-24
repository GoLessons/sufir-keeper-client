package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeyringStoreSaveLoadClear(t *testing.T) {
	dir := t.TempDir()
	opts := KeyringOptions{
		ServiceName: "sufir-keeper-client",
		Backend:     "file",
		FileDir:     filepath.Join(dir, "keyring"),
		AccessKey:   "access_token",
		RefreshKey:  "refresh_token",
	}
	store, err := NewKeyringStore(opts)
	require.NoError(t, err)
	tokens := AuthTokens{
		AccessToken:  "a1",
		RefreshToken: "r1",
		TokenType:    "bearer",
		ExpiresIn:    3600,
	}
	err = store.SaveTokens(tokens)
	require.NoError(t, err)
	loaded, err := store.LoadTokens()
	require.NoError(t, err)
	require.Equal(t, tokens.AccessToken, loaded.AccessToken)
	require.Equal(t, tokens.RefreshToken, loaded.RefreshToken)
	at, ok := store.CurrentAccessToken()
	require.True(t, ok)
	require.Equal(t, tokens.AccessToken, at)
	require.True(t, store.HasRefreshToken())
	err = store.Clear()
	require.NoError(t, err)
	_, err = store.LoadTokens()
	require.Error(t, err)
	_ = os.RemoveAll(opts.FileDir)
}

