package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttachFilesCommands(t *testing.T) {
	root := NewRootCmd("", "", "")
	names := make([]string, 0, len(root.Commands()))
	for _, c := range root.Commands() {
		names = append(names, c.Name())
	}
	require.Contains(t, names, "upload")
	require.Contains(t, names, "download")
}

func TestUploadRequiresPath(t *testing.T) {
	cmd := NewRootCmd("dev", "none", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"upload"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
}
