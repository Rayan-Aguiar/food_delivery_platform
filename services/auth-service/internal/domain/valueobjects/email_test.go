package valueobjects

import "testing"

func TestNewEmail(t *testing.T) {
	e, err := NewEmail("  USER@Example.COM ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.String() != "user@example.com" {
		t.Fatalf("unexpected normalized email: %s", e.String())
	}

	invalid := []string{"", "  ", "abc", "foo@", "@bar.com"}
	for _, raw := range invalid {
		if _, err := NewEmail(raw); err == nil {
			t.Fatalf("expected invalid email for %q", raw)
		}
	}
}
