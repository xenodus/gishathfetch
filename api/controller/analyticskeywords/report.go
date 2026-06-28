package analyticskeywords

import (
	"context"
	"fmt"
	"time"

	"mtg-price-checker-sg/gateway/ga4"
)

const (
	TopKeywordLimit = 20
	periodLast24Hours = "last24Hours"
	periodLast7Days        = "last7Days"
	periodLast30Days       = "last30Days"
)

// KeywordCount is one ranked search keyword.
type KeywordCount struct {
	Term  string `json:"term"`
	Count int64  `json:"count"`
}

// PeriodReport holds top keywords for one reporting window.
type PeriodReport struct {
	Start       string         `json:"start,omitempty"`
	End         string         `json:"end,omitempty"`
	StartDate   string         `json:"startDate,omitempty"`
	EndDate     string         `json:"endDate,omitempty"`
	Keywords    []KeywordCount `json:"keywords"`
	KeywordLimit int           `json:"keywordLimit"`
}

// Report is the full analytics export written to S3.
type Report struct {
	GeneratedAt string                    `json:"generatedAt"`
	PropertyID  string                    `json:"propertyId"`
	EventName   string                    `json:"eventName"`
	Periods     map[string]PeriodReport   `json:"periods"`
}

var nowFunc = time.Now

// BuildReport fetches top search keywords for the last 24 hours, 7 days, and 30 days.
func BuildReport(ctx context.Context, reporter ga4.Reporter, propertyID string, limit int) (*Report, error) {
	if reporter == nil {
		return nil, fmt.Errorf("analyticskeywords: reporter is required")
	}
	if limit <= 0 {
		limit = TopKeywordLimit
	}

	now := nowFunc().UTC()
	report := &Report{
		GeneratedAt: now.Format(time.RFC3339),
		PropertyID:  propertyID,
		EventName:   ga4.SearchEventName,
		Periods:     make(map[string]PeriodReport, 3),
	}

	last24Hours, err := reporter.TopSearchTermsLast24Hours(ctx, now, limit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 24 hours: %w", err)
	}
	report.Periods[periodLast24Hours] = PeriodReport{
		Start:        now.Add(-24 * time.Hour).Format(time.RFC3339),
		End:          now.Format(time.RFC3339),
		Keywords:     toKeywordCounts(last24Hours),
		KeywordLimit: limit,
	}

	last7Days, err := reporter.TopSearchTerms(ctx, "7daysAgo", "today", limit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 7 days: %w", err)
	}
	report.Periods[periodLast7Days] = PeriodReport{
		StartDate:    "7daysAgo",
		EndDate:      "today",
		Keywords:     toKeywordCounts(last7Days),
		KeywordLimit: limit,
	}

	last30Days, err := reporter.TopSearchTerms(ctx, "30daysAgo", "today", limit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 30 days: %w", err)
	}
	report.Periods[periodLast30Days] = PeriodReport{
		StartDate:    "30daysAgo",
		EndDate:      "today",
		Keywords:     toKeywordCounts(last30Days),
		KeywordLimit: limit,
	}

	return report, nil
}

func toKeywordCounts(terms []ga4.SearchTermCount) []KeywordCount {
	keywords := make([]KeywordCount, 0, len(terms))
	for _, term := range terms {
		keywords = append(keywords, KeywordCount{
			Term:  term.Term,
			Count: term.Count,
		})
	}
	return keywords
}
