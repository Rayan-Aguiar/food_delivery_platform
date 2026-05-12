package logger

import "testing"

func TestNewLoggerLevels(t *testing.T) {
	levels := []string{"debug", "warn", "error", "info", "unknown"}
	for _, lvl := range levels {
		l := New(Config{ServiceName: "svc", Level: lvl})
		if l == nil {
			t.Fatalf("expected logger for level %s", lvl)
		}
		l.Info("test log")
	}
}
