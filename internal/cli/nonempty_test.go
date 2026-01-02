package cli

import "testing"

func TestNonEmpty(t *testing.T) {
	if nonEmpty("", "fb") != "fb" {
		t.Fatalf("fallback not applied")
	}
	if nonEmpty("v", "fb") != "v" {
		t.Fatalf("value not applied")
	}
}
