package buildinfo

import (
	"testing"
)

func TestValidateFailsOnDefaults(t *testing.T) {
	version = "dev"
	date = "unknown"
	if err := Validate(); err == nil {
		t.Fatalf("expected error on defaults")
	}
}

func TestValidateFailsOnUnknownDate(t *testing.T) {
	version = "vX"
	date = "unknown"
	if err := Validate(); err == nil {
		t.Fatalf("expected error on unknown date")
	}
}

func TestValidateSuccess(t *testing.T) {
	version = "v1.0.0"
	date = "2025-01-01T00:00:00Z"
	if err := Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
