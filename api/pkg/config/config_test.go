package config

import (
	"testing"
)

func TestUseDynamicProxy(t *testing.T) {
	t.Run("defaults to enabled when unset", func(t *testing.T) {
		t.Setenv(UseDynamicProxyEnv, "")
		if !UseDynamicProxy() {
			t.Fatalf("expected dynamic proxy to be enabled by default")
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

	t.Run("defaults to enabled for invalid value", func(t *testing.T) {
		t.Setenv(UseDynamicProxyEnv, "not-a-bool")
		if !UseDynamicProxy() {
			t.Fatalf("expected invalid toggle to default to enabled")
		}
	})
}

func TestCKPriceLookupEnabled(t *testing.T) {
	t.Run("defaults to disabled when unset", func(t *testing.T) {
		t.Setenv(CKPriceLookupEnabledEnv, "")
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
