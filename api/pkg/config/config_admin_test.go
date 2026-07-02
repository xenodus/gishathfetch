package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdminEnabled_RequiresAllSettings(t *testing.T) {
	keys := []string{
		AdminUsernameEnv,
		AdminPasswordEnv,
		AdminSessionSecretEnv,
		AdminLoginDynamoDBTableEnv,
	}
	original := make(map[string]string, len(keys))
	for _, key := range keys {
		original[key] = os.Getenv(key)
		_ = os.Unsetenv(key)
	}
	t.Cleanup(func() {
		for _, key := range keys {
			if value, ok := original[key]; ok && value != "" {
				_ = os.Setenv(key, value)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	})

	require.False(t, AdminEnabled())

	_ = os.Setenv(AdminUsernameEnv, "admin")
	require.False(t, AdminEnabled())

	for _, key := range keys {
		_ = os.Setenv(key, "value")
	}
	require.True(t, AdminEnabled())
}

func TestAdminLoginRateLimitsFromEnv_UsesDefaults(t *testing.T) {
	limits := AdminLoginRateLimitsFromEnv()
	require.Equal(t, DefaultAdminLoginMaxFailuresPerIP, limits.MaxFailuresPerIP)
	require.Equal(t, DefaultAdminLoginIPWindow, limits.IPWindow)
	require.Equal(t, DefaultAdminLoginMaxFailuresPerUser, limits.MaxFailuresPerUser)
}

func TestDurationFromEnv_InvalidFallsBack(t *testing.T) {
	t.Setenv(AdminSessionTTLEnv, "not-a-number")
	require.Equal(t, DefaultAdminSessionTTL, AdminSessionTTL())
}

func TestAdminAttemptLogRetention_UsesConfiguredDays(t *testing.T) {
	t.Setenv(AdminAttemptLogRetentionDaysEnv, "30")
	require.Equal(t, 30*24*time.Hour, AdminAttemptLogRetention())
}
