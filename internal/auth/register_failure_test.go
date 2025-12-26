package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

func TestRegisterFailureOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/register" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	client := retryablehttp.NewClient()
	client.RetryMax = 0
	store, err := NewKeyringStore(KeyringOptions{ServiceName: "sufir-keeper-client", Backend: "file", FileDir: t.TempDir(), FilePassword: "test"})
	require.NoError(t, err)
	mgr := &Manager{client: client, store: store}
	err = mgr.Register(context.Background(), srv.URL, "u", "p")
	require.Error(t, err)
}
