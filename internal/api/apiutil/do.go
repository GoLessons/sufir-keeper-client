package apiutil

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

func Do(client *retryablehttp.Client, req *retryablehttp.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	if IsUnauthorized(resp) {
		return resp, Error{Status: http.StatusUnauthorized, Message: "unauthorized"}
	}
	if IsForbidden(resp) {
		return resp, Error{Status: http.StatusForbidden, Message: "forbidden"}
	}
	return resp, nil
}
