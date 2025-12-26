package auth

import (
	"context"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func TestVerifyFailsWhenNoAccessToken(t *testing.T) {
	client := retryablehttp.NewClient()
	client.RetryMax = 0
	store, err := NewKeyringStore(KeyringOptions{ServiceName: "sufir-keeper-client", Backend: "file", FileDir: t.TempDir(), FilePassword: "test"})
	require.NoError(t, err)
	mgr := &Manager{client: client, store: store}
	_, err = mgr.Verify(context.Background(), "http://example.invalid")
	require.Error(t, err)
}
