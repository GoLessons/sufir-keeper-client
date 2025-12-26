package httpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

type settings struct {
	timeout                         time.Duration
	retryMax                        int
	retryWaitMin                    time.Duration
	retryWaitMax                    time.Duration
	transportMaxIdleConns           int
	transportMaxIdleConnsPerHost    int
	transportIdleConnTimeout        time.Duration
	transportTLSHandshakeTimeout    time.Duration
	transportExpectContinueTimeout  time.Duration
	transportMaxResponseHeaderBytes int64
	transportReadBufferSize         int
	transportWriteBufferSize        int
}

type Option func(*settings)

func WithTimeout(d time.Duration) Option {
	return func(s *settings) {
		s.timeout = d
	}
}

func WithRetryMax(n int) Option {
	return func(s *settings) {
		s.retryMax = n
	}
}

func WithRetryWait(min, max time.Duration) Option {
	return func(s *settings) {
		s.retryWaitMin = min
		s.retryWaitMax = max
	}
}

func WithTransportMaxIdleConns(n int) Option {
	return func(s *settings) {
		s.transportMaxIdleConns = n
	}
}

func WithTransportMaxIdleConnsPerHost(n int) Option {
	return func(s *settings) {
		s.transportMaxIdleConnsPerHost = n
	}
}

func WithTransportIdleConnTimeout(d time.Duration) Option {
	return func(s *settings) {
		s.transportIdleConnTimeout = d
	}
}

func WithTransportTLSHandshakeTimeout(d time.Duration) Option {
	return func(s *settings) {
		s.transportTLSHandshakeTimeout = d
	}
}

func WithTransportExpectContinueTimeout(d time.Duration) Option {
	return func(s *settings) {
		s.transportExpectContinueTimeout = d
	}
}

func WithTransportMaxResponseHeaderBytes(n int64) Option {
	return func(s *settings) {
		s.transportMaxResponseHeaderBytes = n
	}
}

func WithTransportReadBufferSize(n int) Option {
	return func(s *settings) {
		s.transportReadBufferSize = n
	}
}

func WithTransportWriteBufferSize(n int) Option {
	return func(s *settings) {
		s.transportWriteBufferSize = n
	}
}

type loggerAdapter struct {
	log logging.Logger
}

func (l loggerAdapter) Printf(format string, v ...interface{}) {
	l.log.Debug(fmt.Sprintf(format, v...))
}

func New(cfg config.Config, log logging.Logger, opts ...Option) (*retryablehttp.Client, error) {
	var pool *x509.CertPool
	var err error
	pool, err = x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if cfg.TLS.CACertPath != "" {
		pem, rerr := os.ReadFile(cfg.TLS.CACertPath)
		if rerr != nil {
			return nil, fmt.Errorf("read ca cert %s: %w", cfg.TLS.CACertPath, rerr)
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("append ca cert failed for %s", cfg.TLS.CACertPath)
		}
	}
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    pool,
	}

	s := settings{
		timeout:      30 * time.Second,
		retryMax:     3,
		retryWaitMin: 200 * time.Millisecond,
		retryWaitMax: 2 * time.Second,
	}
	for _, o := range opts {
		o(&s)
	}

	transport := &http.Transport{
		TLSClientConfig:       tlsCfg,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if s.transportMaxIdleConns > 0 {
		transport.MaxIdleConns = s.transportMaxIdleConns
	}
	if s.transportMaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = s.transportMaxIdleConnsPerHost
	}
	if s.transportIdleConnTimeout > 0 {
		transport.IdleConnTimeout = s.transportIdleConnTimeout
	}
	if s.transportTLSHandshakeTimeout > 0 {
		transport.TLSHandshakeTimeout = s.transportTLSHandshakeTimeout
	}
	if s.transportExpectContinueTimeout > 0 {
		transport.ExpectContinueTimeout = s.transportExpectContinueTimeout
	}
	if s.transportMaxResponseHeaderBytes > 0 {
		transport.MaxResponseHeaderBytes = s.transportMaxResponseHeaderBytes
	}
	if s.transportReadBufferSize > 0 {
		transport.ReadBufferSize = s.transportReadBufferSize
	}
	if s.transportWriteBufferSize > 0 {
		transport.WriteBufferSize = s.transportWriteBufferSize
	}

	base := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	base.Timeout = s.timeout
	rc := retryablehttp.NewClient()
	rc.RetryMax = s.retryMax
	rc.RetryWaitMin = s.retryWaitMin
	rc.RetryWaitMax = s.retryWaitMax
	rc.Logger = loggerAdapter{log: log}
	rc.HTTPClient = base
	rc.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) {
				return true, nil
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return true, nil
			}
			return false, nil
		}
		if resp == nil {
			return true, nil
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			return true, nil
		}
		if resp.StatusCode >= 500 {
			return true, nil
		}
		return false, nil
	}

	return rc, nil
}
