package apiutil

import "net/http"

type Error struct {
	Message string
	Status  int
}

func (e Error) Error() string { return e.Message }

func IsUnauthorized(resp *http.Response) bool {
	return resp != nil && resp.StatusCode == http.StatusUnauthorized
}

func IsForbidden(resp *http.Response) bool {
	return resp != nil && resp.StatusCode == http.StatusForbidden
}
