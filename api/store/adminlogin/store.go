package adminlogin

import (
	"context"
	"time"
)

// Attempt is an append-only audit record for a login try.
type Attempt struct {
	ID        string
	IP        string
	Username  string
	Success   bool
	Blocked   bool
	UserAgent string
	CreatedAt time.Time
}

// Lockout reports whether a client is temporarily blocked from logging in.
type Lockout struct {
	Locked     bool
	RetryAfter time.Duration
}

// Store persists login attempt logs and rate-limit counters.
type Store interface {
	CheckLockout(ctx context.Context, ip, username string, now time.Time) (Lockout, error)
	RecordAttempt(ctx context.Context, attempt Attempt, retention time.Duration) error
	RecordFailure(ctx context.Context, ip, username string, now time.Time, limits RateLimits) error
	ClearFailures(ctx context.Context, ip, username string) error
}

// RateLimits configures brute-force protection thresholds.
type RateLimits struct {
	MaxFailuresPerIP   int
	IPWindow           time.Duration
	IPLockout          time.Duration
	MaxFailuresPerUser int
	UserWindow         time.Duration
	UserLockout        time.Duration
}
