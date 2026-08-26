package config

import (
	"testing"
)

func TestAgoraSearchEnabled(t *testing.T) {
	if !AgoraSearchEnabled {
		t.Fatalf("expected agora search to be enabled")
	}
}

func TestUseDedicatedProxy(t *testing.T) {
	t.Run("defaults to enabled when unset", func(t *testing.T) {
		t.Setenv(UseDedicatedProxyEnv, "")
		if !UseDedicatedProxy() {
			t.Fatalf("expected dedicated proxy to be enabled by default")
		}
	})

	t.Run("respects explicit true", func(t *testing.T) {
		t.Setenv(UseDedicatedProxyEnv, "true")
		if !UseDedicatedProxy() {
			t.Fatalf("expected dedicated proxy to be enabled")
		}
	})

	t.Run("respects explicit false", func(t *testing.T) {
		t.Setenv(UseDedicatedProxyEnv, "false")
		if UseDedicatedProxy() {
			t.Fatalf("expected dedicated proxy to be disabled")
		}
	})

	t.Run("defaults to enabled for invalid value", func(t *testing.T) {
		t.Setenv(UseDedicatedProxyEnv, "not-a-bool")
		if !UseDedicatedProxy() {
			t.Fatalf("expected invalid toggle to default to enabled")
		}
	})
}

func TestAPIMaintenanceMode(t *testing.T) {
	t.Run("defaults to disabled when unset", func(t *testing.T) {
		t.Setenv(APIMaintenanceModeEnv, "")
		if APIMaintenanceMode() {
			t.Fatalf("expected maintenance mode to be disabled by default")
		}
	})

	t.Run("respects explicit true", func(t *testing.T) {
		t.Setenv(APIMaintenanceModeEnv, "true")
		if !APIMaintenanceMode() {
			t.Fatalf("expected maintenance mode to be enabled")
		}
	})

	t.Run("respects explicit false", func(t *testing.T) {
		t.Setenv(APIMaintenanceModeEnv, "false")
		if APIMaintenanceMode() {
			t.Fatalf("expected maintenance mode to be disabled")
		}
	})
}

func TestAPIMaintenanceMessage(t *testing.T) {
	t.Run("uses custom message when set", func(t *testing.T) {
		t.Setenv(APIMaintenanceMessageEnv, "Custom maintenance message.")
		if got := APIMaintenanceMessage(); got != "Custom maintenance message." {
			t.Fatalf("unexpected message: %q", got)
		}
	})

	t.Run("uses default when unset", func(t *testing.T) {
		t.Setenv(APIMaintenanceMessageEnv, "")
		if got := APIMaintenanceMessage(); got != DefaultAPIMaintenanceMessage {
			t.Fatalf("unexpected default message: %q", got)
		}
	})
}

func TestAPINoticeMessage(t *testing.T) {
	t.Run("returns empty when unset", func(t *testing.T) {
		t.Setenv(APINoticeMessageEnv, "")
		if got := APINoticeMessage(); got != "" {
			t.Fatalf("unexpected notice message: %q", got)
		}
	})

	t.Run("returns trimmed message when set", func(t *testing.T) {
		t.Setenv(APINoticeMessageEnv, "  Prices may be delayed.  ")
		if got := APINoticeMessage(); got != "Prices may be delayed." {
			t.Fatalf("unexpected notice message: %q", got)
		}
	})
}

func TestCKPriceLookupEnabled(t *testing.T) {
	t.Run("defaults to enabled when dynamodb table is configured", func(t *testing.T) {
		t.Setenv(CKPriceLookupEnabledEnv, "")
		t.Setenv(CKDynamoDBTableEnv, "mtg-ck-prices")
		if !CKPriceLookupEnabled() {
			t.Fatalf("expected ck price lookup to be enabled when table is configured")
		}
	})

	t.Run("defaults to disabled when unset", func(t *testing.T) {
		t.Setenv(CKPriceLookupEnabledEnv, "")
		t.Setenv(CKDynamoDBTableEnv, "")
		if CKPriceLookupEnabled() {
			t.Fatalf("expected ck price lookup to be disabled by default")
		}
	})

	t.Run("respects explicit true", func(t *testing.T) {
		t.Setenv(CKPriceLookupEnabledEnv, "true")
		if !CKPriceLookupEnabled() {
			t.Fatalf("expected ck price lookup to be enabled")
		}
	})

	t.Run("respects explicit false", func(t *testing.T) {
		t.Setenv(CKPriceLookupEnabledEnv, "false")
		if CKPriceLookupEnabled() {
			t.Fatalf("expected ck price lookup to be disabled")
		}
	})
}
