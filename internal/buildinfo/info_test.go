package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInfoReturnsCurrentValues(t *testing.T) {
	version = "vX"
	commit = "abc123"
	date = "2025-01-01T00:00:00Z"
	v, c, d := Info()
	require.Equal(t, "vX", v)
	require.Equal(t, "abc123", c)
	require.Equal(t, "2025-01-01T00:00:00Z", d)
}
