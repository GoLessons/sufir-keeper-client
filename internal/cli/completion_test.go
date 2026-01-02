package cli

import (
	"bytes"
	"context"
	"testing"
)

func TestCompletionBash(t *testing.T) {
	cmd := NewRootCmd("dev", "none", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"completion", "bash"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty completion output")
	}
}

func TestCompletionZsh(t *testing.T) {
	cmd := NewRootCmd("dev", "none", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"completion", "zsh"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty completion output")
	}
}

func TestCompletionFish(t *testing.T) {
	cmd := NewRootCmd("dev", "none", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"completion", "fish"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty completion output")
	}
}

func TestCompletionPowershell(t *testing.T) {
	cmd := NewRootCmd("dev", "none", "2025-01-01")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"completion", "powershell"})
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty completion output")
	}
}
