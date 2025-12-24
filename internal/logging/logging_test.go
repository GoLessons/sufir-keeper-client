package logging

import "testing"

func TestNewLoggerLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	for _, lv := range levels {
		l, err := NewLogger(lv)
		if err != nil {
			t.Fatalf("unexpected error for level %s: %v", lv, err)
		}
		if l == nil {
			t.Fatalf("nil logger for level %s", lv)
		}
		_ = l.Sync()
	}
	if _, err := NewLogger("unknown"); err == nil {
		t.Fatalf("expected error for unknown level")
	}
}
