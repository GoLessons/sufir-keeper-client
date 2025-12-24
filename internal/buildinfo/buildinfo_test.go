package buildinfo

import (
	"os"
	"os/exec"
	"testing"
)

func TestValidateDefault(t *testing.T) {
	err := Validate()
	if err == nil {
		t.Fatal("expected error for default build metadata")
	}
}

func TestInfoReturnsValues(t *testing.T) {
	v, c, d := Info()
	if v == "" || c == "" || d == "" {
		t.Fatalf("unexpected empty info: %s %s %s", v, c, d)
	}
}

func TestValidateSuccess(t *testing.T) {
	version = "1.0.0"
	date = "2025-01-01T00:00:00Z"
	commit = "abc123"
	if err := Validate(); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
}

func TestEnsureSuccess(t *testing.T) {
	version = "1.0.0"
	date = "2025-01-01T00:00:00Z"
	commit = "abc123"
	Ensure()
}

func TestEnsureExit(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run", "TestEnsureExitHelper")
	cmd.Env = append(os.Environ(), "BUILDINFO_TEST_EXIT=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected process to exit with error")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 2 {
			t.Fatalf("unexpected exit code: %d", exitErr.ExitCode())
		}
		return
	}
	t.Fatalf("unexpected error type: %v", err)
}

func TestEnsureExitHelper(t *testing.T) {
	if os.Getenv("BUILDINFO_TEST_EXIT") != "1" {
		t.Skip("helper")
	}
	version = "dev"
	date = "unknown"
	commit = "none"
	Ensure()
}
