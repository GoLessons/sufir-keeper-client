package logging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnknownLogLevel(t *testing.T) {
	_, err := NewLogger("trace")
	require.Error(t, err)
}

