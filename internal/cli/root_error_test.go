package cli

import (
	"os"
	"testing"
)

func TestRootPreRunInvalidLogLevel(t *testing.T) {
	os.Setenv("SUFIR_KEEPER_LOG_LEVEL", "trace")
	cmd := NewRootCmd("v", "c", "d")
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error due to invalid log level")
	}
}
