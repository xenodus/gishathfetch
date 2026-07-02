package adminlogin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type memoryStore struct {
	records map[string]rateRecord
}

func newMemoryStore() *memoryStore {
	return &memoryStore{records: map[string]rateRecord{}}
}

func (m *memoryStore) CheckLockout(_ context.Context, ip, username string, now time.Time) (Lockout, error) {
	for _, pk := range []string{rateKey("ip", ip), rateKey("user", username)} {
		record, ok := m.records[pk]
		if !ok || record.LockedUntil == "" {
			continue
		}
		lockedUntil, err := time.Parse(time.RFC3339, record.LockedUntil)
		if err == nil && lockedUntil.After(now) {
			return Lockout{Locked: true, RetryAfter: lockedUntil.Sub(now)}, nil
		}
	}
	return Lockout{}, nil
}

func (m *memoryStore) RecordAttempt(_ context.Context, _ Attempt, _ time.Duration) error {
	return nil
}

func (m *memoryStore) RecordFailure(
	_ context.Context,
	ip, username string,
	now time.Time,
	limits RateLimits,
) error {
	if err := incrementFailureRecord(m.records, rateKey("ip", ip), now, limits.MaxFailuresPerIP, limits.IPWindow, limits.IPLockout); err != nil {
		return err
	}
	return incrementFailureRecord(
		m.records,
		rateKey("user", username),
		now,
		limits.MaxFailuresPerUser,
		limits.UserWindow,
		limits.UserLockout,
	)
}

func (m *memoryStore) ClearFailures(_ context.Context, ip, username string) error {
	delete(m.records, rateKey("ip", ip))
	delete(m.records, rateKey("user", username))
	return nil
}

func incrementFailureRecord(
	records map[string]rateRecord,
	pk string,
	now time.Time,
	maxFailures int,
	window time.Duration,
	lockout time.Duration,
) error {
	current, ok := records[pk]
	if !ok {
		current = rateRecord{
			PK:          pk,
			WindowStart: now.UTC().Format(time.RFC3339),
		}
	}

	windowStart, err := time.Parse(time.RFC3339, current.WindowStart)
	if err != nil || now.Sub(windowStart) > window {
		current.FailCount = 0
		current.WindowStart = now.UTC().Format(time.RFC3339)
		current.LockedUntil = ""
	}

	current.FailCount++
	if current.FailCount >= maxFailures {
		lockedUntil := now.Add(lockout)
		current.LockedUntil = lockedUntil.UTC().Format(time.RFC3339)
		current.TTL = lockedUntil.Add(time.Hour).Unix()
	} else {
		current.TTL = now.Add(window).Add(time.Hour).Unix()
	}

	records[pk] = current
	return nil
}

func TestService_RecordFailureLocksAfterThreshold(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store, RateLimits{
		MaxFailuresPerIP: 2,
		IPWindow:         time.Minute,
		IPLockout:        10 * time.Minute,
	})
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	require.NoError(t, service.RecordFailure(context.Background(), "203.0.113.1", "admin", now))
	lockout, err := service.CheckLockout(context.Background(), "203.0.113.1", "admin", now)
	require.NoError(t, err)
	require.False(t, lockout.Locked)

	require.NoError(t, service.RecordFailure(context.Background(), "203.0.113.1", "admin", now))
	lockout, err = service.CheckLockout(context.Background(), "203.0.113.1", "admin", now)
	require.NoError(t, err)
	require.True(t, lockout.Locked)
}

func TestRateKey_NormalizesValues(t *testing.T) {
	require.Equal(t, "rate#ip#203.0.113.1", rateKey("ip", " 203.0.113.1 "))
	require.Equal(t, "rate#user#admin", rateKey("user", "Admin"))
}
