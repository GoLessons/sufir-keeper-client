package apiutil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsUnauthorizedAndForbidden(t *testing.T) {
	var nilResp *http.Response
	require.False(t, IsUnauthorized(nilResp))
	require.False(t, IsForbidden(nilResp))
	r1 := &http.Response{StatusCode: http.StatusUnauthorized}
	r2 := &http.Response{StatusCode: http.StatusForbidden}
	require.True(t, IsUnauthorized(r1))
	require.True(t, IsForbidden(r2))
}
