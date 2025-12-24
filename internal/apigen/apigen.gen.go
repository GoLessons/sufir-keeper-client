package apigen

import (
	"net/http"
)

type UserRegister struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserLogin struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type ClientWithResponses struct {
	BaseURL    string
	HTTPClient *http.Client
}

type ClientOption func(*ClientWithResponses)

func WithHTTPClient(c *http.Client) ClientOption {
	return func(cl *ClientWithResponses) {
		cl.HTTPClient = c
	}
}

func NewClientWithResponses(baseURL string, opts ...ClientOption) (*ClientWithResponses, error) {
	c := &ClientWithResponses{BaseURL: baseURL, HTTPClient: http.DefaultClient}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

