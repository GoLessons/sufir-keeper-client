package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAttachAuthCommands(t *testing.T) {
	root := NewRootCmd("", "", "")
	names := make([]string, 0, len(root.Commands()))
	for _, c := range root.Commands() {
		names = append(names, c.Name())
	}
	require.Contains(t, names, "login")
	require.Contains(t, names, "status")
	require.Contains(t, names, "logout")
	require.Contains(t, names, "register")
	// auth-verify убран: фоновые верификация/refresh происходят прозрачно
}

func TestStatusNotAuthorized(t *testing.T) {
	root := NewRootCmd("dev", "none", "2025-01-01")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--ca-cert-path=", "status"})
	err := root.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() == "" {
		t.Fatal("empty output")
	}
}
