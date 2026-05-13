package valueobjects

import "testing"

func TestNewDefaultPasswordPolicy(t *testing.T) {
	p := NewDefaultPasswordPolicy()
	if p.MinLength != 8 || !p.RequireUpper || !p.RequireLower || !p.RequireNumber || !p.RequireSpecial {
		t.Fatalf("unexpected default policy: %+v", p)
	}
}

func TestPasswordPolicyValidate(t *testing.T) {
	p := NewDefaultPasswordPolicy()

	if err := p.Validate("Abcdef1!"); err != nil {
		t.Fatalf("expected strong password, got: %v", err)
	}

	weak := []string{"short1!", "abcdef1!", "ABCDEF1!", "Abcdefg!", "Abcdefg1"}
	for _, w := range weak {
		if err := p.Validate(w); err == nil {
			t.Fatalf("expected weak password error for %q", w)
		}
	}

	p.MinLength = 0
	if err := p.Validate("Abcdef1!"); err == nil {
		t.Fatal("expected invalid policy error")
	}
}
