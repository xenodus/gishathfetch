package config

import (
	"testing"
)

func TestAgoraSearchEnabled(t *testing.T) {
	if !AgoraSearchEnabled {
		t.Fatalf("expected agora search to be enabled")
	}
}

func TestUseDynamicProxy(t *testing.T) {
	t.Run("defaults to disabled when unset", func(t *testing.T) {
		t.Setenv(UseDynamicProxyEnv, "")
		if UseDynamicProxy() {
			t.Fatalf("expected dynamic proxy to be disabled by default")
		}
	})

	t.Run("respects explicit true", func(t *testing.T) {
		t.Setenv(UseDynamicProxyEnv, "true")
		if !UseDynamicProxy() {
			t.Fatalf("expected dynamic proxy to be enabled")
		}
	})

	t.Run("respects explicit false", func(t *testing.T) {
		t.Setenv(UseDynamicProxyEnv, "false")
		if UseDynamicProxy() {
			t.Fatalf("expected dynamic proxy to be disabled")
		}
	})

	t.Run("defaults to disabled for invalid value", func(t *testing.T) {
		t.Setenv(UseDynamicProxyEnv, "not-a-bool")
		if UseDynamicProxy() {
			t.Fatalf("expected invalid toggle to default to disabled")
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
