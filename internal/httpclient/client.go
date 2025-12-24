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
	timeout      time.Duration
	retryMax     int
	retryWaitMin time.Duration
	retryWaitMax time.Duration
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

type loggerAdapter struct {
	log logging.Logger
}

func (l loggerAdapter) Printf(format string, v ...interface{}) {
	l.log.Debug(fmt.Sprintf(format, v...))
}

func New(cfg config.Config, log logging.Logger, opts ...Option) (*retryablehttp.Client, error) {
	if cfg.TLS.CACertPath == "" {
		return nil, errors.New("tls ca cert path required")
	}
	pem, err := os.ReadFile(cfg.TLS.CACertPath)

	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()

	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("failed to append ca cert")
	}
	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    pool,
	}

	transport := &http.Transport{
		TLSClientConfig:       tlsCfg,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	base := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
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
