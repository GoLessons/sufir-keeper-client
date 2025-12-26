package httpclient

import (
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

func writeServerCertToFile(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	cert := srv.Certificate()
	der := cert.Raw
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	f, err := os.CreateTemp("", "ca-*.crt")
	require.NoError(t, err)
	_, err = f.Write(pemBytes)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}

func TestTLSAndRequestSuccess(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	caPath := writeServerCertToFile(t, srv)
	t.Cleanup(func() { _ = os.Remove(caPath) })
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{
		TLS: config.TLSConfig{
			CACertPath: caPath,
		},
	}
	client, err := New(cfg, log)
	require.NoError(t, err)
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRetryLogic(t *testing.T) {
	var counter int64
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt64(&counter, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	caPath := writeServerCertToFile(t, srv)
	t.Cleanup(func() { _ = os.Remove(caPath) })
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{
		TLS: config.TLSConfig{
			CACertPath: caPath,
		},
	}
	client, err := New(cfg, log, WithRetryMax(3), WithRetryWait(10*time.Millisecond, 50*time.Millisecond))
	require.NoError(t, err)
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.GreaterOrEqual(t, atomic.LoadInt64(&counter), int64(3))
}

func TestDoNotRetryOn404(t *testing.T) {
	var counter int64
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&counter, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	caPath := writeServerCertToFile(t, srv)
	t.Cleanup(func() { _ = os.Remove(caPath) })
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{
		TLS: config.TLSConfig{
			CACertPath: caPath,
		},
	}
	client, err := New(cfg, log, WithRetryMax(3), WithRetryWait(10*time.Millisecond, 50*time.Millisecond))
	require.NoError(t, err)
	_, err = client.Get(srv.URL)
	require.NoError(t, err)
	require.Equal(t, int64(1), atomic.LoadInt64(&counter))
}

func TestOptionsApplied(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	caPath := writeServerCertToFile(t, srv)
	t.Cleanup(func() { _ = os.Remove(caPath) })
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{
		TLS: config.TLSConfig{
			CACertPath: caPath,
		},
	}
	client, err := New(
		cfg,
		log,
		WithTimeout(1*time.Second),
		WithRetryMax(5),
		WithRetryWait(10*time.Millisecond, 20*time.Millisecond),
		WithTransportMaxIdleConns(200),
		WithTransportMaxIdleConnsPerHost(50),
		WithTransportIdleConnTimeout(120*time.Second),
		WithTransportTLSHandshakeTimeout(5*time.Second),
		WithTransportExpectContinueTimeout(2*time.Second),
		WithTransportMaxResponseHeaderBytes(1<<20),
		WithTransportReadBufferSize(1<<16),
		WithTransportWriteBufferSize(1<<16),
	)
	require.NoError(t, err)
	require.Equal(t, 5, client.RetryMax)
	require.Equal(t, 10*time.Millisecond, client.RetryWaitMin)
	require.Equal(t, 20*time.Millisecond, client.RetryWaitMax)
	require.Equal(t, 1*time.Second, client.HTTPClient.Timeout)
	tr, ok := client.HTTPClient.Transport.(*http.Transport)
	require.True(t, ok)
	require.Equal(t, 200, tr.MaxIdleConns)
	require.Equal(t, 50, tr.MaxIdleConnsPerHost)
	require.Equal(t, 120*time.Second, tr.IdleConnTimeout)
	require.Equal(t, 5*time.Second, tr.TLSHandshakeTimeout)
	require.Equal(t, 2*time.Second, tr.ExpectContinueTimeout)
	require.Equal(t, int64(1<<20), tr.MaxResponseHeaderBytes)
	require.Equal(t, 1<<16, tr.ReadBufferSize)
	require.Equal(t, 1<<16, tr.WriteBufferSize)
}

func TestMissingCACertPath(t *testing.T) {
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{}
	_, err = New(cfg, log)
	require.NoError(t, err)
}

func TestInvalidCACert(t *testing.T) {
	f, err := os.CreateTemp("", "bad-*.crt")
	require.NoError(t, err)
	_, err = f.Write([]byte("not a pem"))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	t.Cleanup(func() { _ = os.Remove(f.Name()) })
	log, err := logging.NewLogger("error")
	require.NoError(t, err)
	cfg := config.Config{
		TLS: config.TLSConfig{
			CACertPath: f.Name(),
		},
	}
	_, err = New(cfg, log)
	require.Error(t, err)
}
