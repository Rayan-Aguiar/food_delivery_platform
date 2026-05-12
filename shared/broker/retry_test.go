package broker

import (
	"testing"
	"time"
)

func TestRetryPolicyNextDelay(t *testing.T) {
	p := RetryPolicy{BaseDelay: time.Second, MaxDelay: 5 * time.Second}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 0, want: 1 * time.Second},
		{attempt: 1, want: 1 * time.Second},
		{attempt: 2, want: 2 * time.Second},
		{attempt: 3, want: 4 * time.Second},
		{attempt: 4, want: 5 * time.Second},
	}

	for _, tt := range tests {
		got := p.NextDelay(tt.attempt)
		if got != tt.want {
			t.Fatalf("attempt=%d got=%v want=%v", tt.attempt, got, tt.want)
		}
	}
}
