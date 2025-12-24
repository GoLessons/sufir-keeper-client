package logging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLoggerLevels(t *testing.T) {
	cases := []struct {
		name      string
		level     string
		wantError bool
	}{
		{name: "debug", level: "debug"},
		{name: "info", level: "info"},
		{name: "warn", level: "warn"},
		{name: "error", level: "error"},
		{name: "unknown", level: "unknown", wantError: true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			l, err := NewLogger(tt.level)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, l)
			_ = l.Sync()
		})
	}
}

func TestLoggerWrappers(t *testing.T) {
	l, err := NewLogger("debug")
	require.NoError(t, err)
	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")
}
