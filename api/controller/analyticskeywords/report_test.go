package analyticskeywords

import (
	"context"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/ga4"
)

type mockReporter struct {
	last24Hours []ga4.SearchTermCount
	last7Days   []ga4.SearchTermCount
	last30Days  []ga4.SearchTermCount
}

func (m *mockReporter) TopSearchTerms(_ context.Context, startDate, endDate string, limit int) ([]ga4.SearchTermCount, error) {
	switch {
	case startDate == "7daysAgo" && endDate == "today":
		return trimTerms(m.last7Days, limit), nil
	case startDate == "30daysAgo" && endDate == "today":
		return trimTerms(m.last30Days, limit), nil
	default:
		return nil, nil
	}
}

func (m *mockReporter) TopSearchTermsLast24Hours(_ context.Context, _ time.Time, limit int) ([]ga4.SearchTermCount, error) {
	return trimTerms(m.last24Hours, limit), nil
}

func trimTerms(terms []ga4.SearchTermCount, limit int) []ga4.SearchTermCount {
	if len(terms) <= limit {
		return terms
	}
	return terms[:limit]
}

func TestBuildReport(t *testing.T) {
	fixedNow := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	originalNowFunc := nowFunc
	nowFunc = func() time.Time { return fixedNow }
	defer func() { nowFunc = originalNowFunc }()

	reporter := &mockReporter{
		last24Hours: []ga4.SearchTermCount{{Term: "Opt", Count: 4}},
		last7Days:   []ga4.SearchTermCount{{Term: "Lightning Bolt", Count: 10}},
		last30Days:  []ga4.SearchTermCount{{Term: "Sol Ring", Count: 20}},
	}

	report, err := BuildReport(context.Background(), reporter, "123456789", 20)
	if err != nil {
		t.Fatalf("build report: %v", err)
	}

	if report.PropertyID != "123456789" || report.EventName != ga4.SearchEventName {
		t.Fatalf("unexpected report metadata: %+v", report)
	}

	last24Hours := report.Periods[periodLast24Hours]
	if last24Hours.Start != "2026-06-27T12:00:00Z" || last24Hours.End != "2026-06-28T12:00:00Z" {
		t.Fatalf("unexpected 24h window: %+v", last24Hours)
	}
	if len(last24Hours.Keywords) != 1 || last24Hours.Keywords[0].Term != "Opt" {
		t.Fatalf("unexpected 24h keywords: %+v", last24Hours.Keywords)
	}

	last7Days := report.Periods[periodLast7Days]
	if last7Days.StartDate != "7daysAgo" || last7Days.EndDate != "today" {
		t.Fatalf("unexpected 7d window: %+v", last7Days)
	}
	if len(last7Days.Keywords) != 1 || last7Days.Keywords[0].Term != "Lightning Bolt" {
		t.Fatalf("unexpected 7d keywords: %+v", last7Days.Keywords)
	}

	last30Days := report.Periods[periodLast30Days]
	if last30Days.StartDate != "30daysAgo" || last30Days.EndDate != "today" {
		t.Fatalf("unexpected 30d window: %+v", last30Days)
	}
	if len(last30Days.Keywords) != 1 || last30Days.Keywords[0].Term != "Sol Ring" {
		t.Fatalf("unexpected 30d keywords: %+v", last30Days.Keywords)
	}
}
