package cli

import (
	"bytes"
	"context"
	"testing"
)

func TestVersionCommandOutput(t *testing.T) {
	cmd := NewRootCmd("1.2.3", "abc123", "2025-01-01T00:00:00Z")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"version"})
	_ = cmd.ExecuteContext(context.Background())
	out := buf.String()
	if !containsAll(out, []string{"version: 1.2.3", "commit: abc123", "date: 2025-01-01T00:00:00Z"}) {
		t.Fatalf("unexpected version output: %s", out)
	}
}

func containsAll(s string, parts []string) bool {
	for _, p := range parts {
		if !bytes.Contains([]byte(s), []byte(p)) {
			return false
		}
	}
	return true
}
