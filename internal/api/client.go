package api

import (
	"github.com/hashicorp/go-retryablehttp"

	"github.com/GoLessons/sufir-keeper-client/internal/api/apigen"
	"github.com/GoLessons/sufir-keeper-client/internal/auth"
	"github.com/GoLessons/sufir-keeper-client/internal/config"
	"github.com/GoLessons/sufir-keeper-client/internal/httpclient"
	"github.com/GoLessons/sufir-keeper-client/internal/logging"
)

type Client struct {
	HTTP *retryablehttp.Client
	API  *apigen.ClientWithResponses
	Auth *auth.Manager
}

func New(cfg config.Config, log logging.Logger, store auth.TokenStore) (*Client, error) {
	rc, err := httpclient.New(cfg, log)
	if err != nil {
		return nil, err
	}
	mgr := auth.NewManager(rc, store)
	base := rc.HTTPClient.Transport
	rt := auth.NewAuthRoundTripper(base, mgr, cfg.Server.BaseURL, store)
	rc.HTTPClient.Transport = rt
	api, err := apigen.NewClientWithResponses(cfg.Server.BaseURL, apigen.WithHTTPClient(rc.HTTPClient))
	if err != nil {
		return nil, err
	}
	return &Client{
		HTTP: rc,
		API:  api,
		Auth: mgr,
	}, nil
}
