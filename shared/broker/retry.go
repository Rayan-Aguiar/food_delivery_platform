package broker

import "time"

type RetryPolicy struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
}

func (p RetryPolicy) NextDelay(attempt int) time.Duration {
    if attempt <= 0 {
        attempt = 1
    }
    d := p.BaseDelay * time.Duration(1<<(attempt-1))
    if d > p.MaxDelay {
        return p.MaxDelay
    }
    return d
}