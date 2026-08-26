package handler

import "testing"

func TestFormatCKPriceRefreshSuccess_IncludesTransportOrder(t *testing.T) {
	got := formatCKPriceRefreshSuccess(42, 1, 1, "2026-07-11T12:00:00Z", "direct → dedicated")
	want := "CK price refresh finished: refreshed=42, top=1, bottom=1, generatedAt=2026-07-11T12:00:00Z, transportOrder=direct → dedicated"
	if got != want {
		t.Fatalf("alert = %q, want %q", got, want)
	}
}

func TestFormatCKPriceRefreshSuccess_UsesUnknownWhenTransportMissing(t *testing.T) {
	got := formatCKPriceRefreshSuccess(1, 0, 0, "2026-07-11T12:00:00Z", "")
	if got != "CK price refresh finished: refreshed=1, top=0, bottom=0, generatedAt=2026-07-11T12:00:00Z, transportOrder=unknown" {
		t.Fatalf("unexpected alert: %q", got)
	}
}
