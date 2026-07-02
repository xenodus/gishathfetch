package adminlogin

import (
	"context"
	"time"
)

// Service coordinates login attempt logging and rate limiting.
type Service struct {
	store  Store
	limits RateLimits
}

func NewService(store Store, limits RateLimits) *Service {
	return &Service{
		store:  store,
		limits: limits,
	}
}

func (s *Service) CheckLockout(ctx context.Context, ip, username string, now time.Time) (Lockout, error) {
	return s.store.CheckLockout(ctx, ip, username, now)
}

func (s *Service) RecordAttempt(ctx context.Context, attempt Attempt, retention time.Duration) error {
	return s.store.RecordAttempt(ctx, attempt, retention)
}

func (s *Service) RecordFailure(ctx context.Context, ip, username string, now time.Time) error {
	return s.store.RecordFailure(ctx, ip, username, now, s.limits)
}

func (s *Service) ClearFailures(ctx context.Context, ip, username string) error {
	return s.store.ClearFailures(ctx, ip, username)
}
