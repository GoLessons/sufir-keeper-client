package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCommandOutput(t *testing.T) {
	cmd := NewRootCmd("1.2.3", "abc123", "2025-01-01T00:00:00Z")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	for _, c := range cmd.Commands() {
		if c.Name() == "version" {
			c.SetOut(&buf)
			c.SetErr(&buf)
		}
	}
	var vc *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Name() == "version" {
			vc = c
			break
		}
	}
	if vc == nil {
		t.Fatalf("version command not found")
	}
	_ = vc.RunE(vc, []string{})
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
