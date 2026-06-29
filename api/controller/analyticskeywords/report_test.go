package analyticskeywords

import (
	"context"
	"strings"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/ga4"
)

type mockReporter struct {
	start6Months string
	start1Year   string
	last24Hours  []ga4.SearchTermCount
	last7Days    []ga4.SearchTermCount
	last30Days   []ga4.SearchTermCount
	last6Months  []ga4.SearchTermCount
	last1Year    []ga4.SearchTermCount
}

func (m *mockReporter) TopSearchTerms(_ context.Context, startDate, endDate string, limit int) ([]ga4.SearchTermCount, error) {
	switch {
	case startDate == "7daysAgo" && endDate == "today":
		return trimTerms(m.last7Days, limit), nil
	case startDate == "30daysAgo" && endDate == "today":
		return trimTerms(m.last30Days, limit), nil
	case startDate == m.start6Months && endDate == "today":
		return trimTerms(m.last6Months, limit), nil
	case startDate == m.start1Year && endDate == "today":
		return trimTerms(m.last1Year, limit), nil
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

func mockVerifyCardName(_ context.Context, query string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(query)) {
	case "opt", "lightning bolt", "sol ring":
		return strings.TrimSpace(query), nil
	default:
		return "", nil
	}
}

func TestBuildReport(t *testing.T) {
	fixedNow := time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	originalNowFunc := nowFunc
	originalVerifyFunc := verifyCardNameFunc
	nowFunc = func() time.Time { return fixedNow }
	verifyCardNameFunc = mockVerifyCardName
	defer func() {
		nowFunc = originalNowFunc
		verifyCardNameFunc = originalVerifyFunc
	}()

	reporter := &mockReporter{
		start6Months: ga4CalendarDate(fixedNow.AddDate(0, -6, 0)),
		start1Year:   ga4CalendarDate(fixedNow.AddDate(-1, 0, 0)),
		last24Hours:  []ga4.SearchTermCount{{Term: "Opt", Count: 4}},
		last7Days:   []ga4.SearchTermCount{{Term: "Lightning Bolt", Count: 10}},
		last30Days:  []ga4.SearchTermCount{{Term: "Sol Ring", Count: 20}},
		last6Months: []ga4.SearchTermCount{{Term: "Opt", Count: 40}},
		last1Year:   []ga4.SearchTermCount{{Term: "Lightning Bolt", Count: 100}},
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

	last6Months := report.Periods[periodLast6Months]
	if last6Months.StartDate != "2025-12-28" || last6Months.EndDate != "today" {
		t.Fatalf("unexpected 6mo window: %+v", last6Months)
	}
	if len(last6Months.Keywords) != 1 || last6Months.Keywords[0].Term != "Opt" {
		t.Fatalf("unexpected 6mo keywords: %+v", last6Months.Keywords)
	}

	last1Year := report.Periods[periodLast1Year]
	if last1Year.StartDate != "2025-06-28" || last1Year.EndDate != "today" {
		t.Fatalf("unexpected 1yr window: %+v", last1Year)
	}
	if len(last1Year.Keywords) != 1 || last1Year.Keywords[0].Term != "Lightning Bolt" {
		t.Fatalf("unexpected 1yr keywords: %+v", last1Year.Keywords)
	}
}

func TestValidateKeywords_FiltersInvalidAndMergesCanonicalNames(t *testing.T) {
	originalVerifyFunc := verifyCardNameFunc
	verifyCardNameFunc = func(_ context.Context, query string) (string, error) {
		switch strings.ToLower(strings.TrimSpace(query)) {
		case "opt":
			return "Opt", nil
		case "lightning bolt":
			return "Lightning Bolt", nil
		default:
			return "", nil
		}
	}
	defer func() { verifyCardNameFunc = originalVerifyFunc }()

	keywords, err := validateKeywords(context.Background(), []ga4.SearchTermCount{
		{Term: "asdfasdf", Count: 99},
		{Term: "opt", Count: 4},
		{Term: "Opt", Count: 2},
		{Term: "Lightning Bolt", Count: 10},
	}, 20)
	if err != nil {
		t.Fatalf("validate keywords: %v", err)
	}

	if len(keywords) != 2 {
		t.Fatalf("expected 2 keywords, got %d", len(keywords))
	}
	if keywords[0].Term != "Lightning Bolt" || keywords[0].Count != 10 {
		t.Fatalf("unexpected first keyword: %+v", keywords[0])
	}
	if keywords[1].Term != "Opt" || keywords[1].Count != 6 {
		t.Fatalf("unexpected second keyword: %+v", keywords[1])
	}
}

func TestValidateKeywords_RespectsLimit(t *testing.T) {
	originalVerifyFunc := verifyCardNameFunc
	verifyCardNameFunc = func(_ context.Context, query string) (string, error) {
		return query, nil
	}
	defer func() { verifyCardNameFunc = originalVerifyFunc }()

	keywords, err := validateKeywords(context.Background(), []ga4.SearchTermCount{
		{Term: "A", Count: 3},
		{Term: "B", Count: 2},
		{Term: "C", Count: 1},
	}, 2)
	if err != nil {
		t.Fatalf("validate keywords: %v", err)
	}
	if len(keywords) != 2 {
		t.Fatalf("expected 2 keywords, got %d", len(keywords))
	}
}
