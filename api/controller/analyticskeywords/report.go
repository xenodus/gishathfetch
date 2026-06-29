package analyticskeywords

import (
	"context"
	"fmt"
	"sort"
	"time"

	"mtg-price-checker-sg/gateway/ga4"
	"mtg-price-checker-sg/gateway/scryfall"
)

const (
	TopKeywordLimit   = 20
	ga4CandidateLimit = 20
	periodLast24Hours = "last24Hours"
	periodLast7Days   = "last7Days"
	periodLast30Days  = "last30Days"
	periodLast6Months = "last6Months"
	periodLast1Year   = "last1Year"
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

var (
	nowFunc            = time.Now
	verifyCardNameFunc = scryfall.VerifyCardName
)

// BuildReport fetches top search keywords for the last 24 hours, 7 days, 30 days, 6 months, and 1 year.
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
		Periods:     make(map[string]PeriodReport, 5),
	}

	last24Hours, err := reporter.TopSearchTermsLast24Hours(ctx, now, ga4CandidateLimit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 24 hours: %w", err)
	}
	validated24Hours, err := validateKeywords(ctx, last24Hours, limit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 24 hours: %w", err)
	}
	report.Periods[periodLast24Hours] = PeriodReport{
		Start:        now.Add(-24 * time.Hour).Format(time.RFC3339),
		End:          now.Format(time.RFC3339),
		Keywords:     validated24Hours,
		KeywordLimit: limit,
	}

	last7Days, err := reporter.TopSearchTerms(ctx, "7daysAgo", "today", ga4CandidateLimit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 7 days: %w", err)
	}
	validated7Days, err := validateKeywords(ctx, last7Days, limit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 7 days: %w", err)
	}
	report.Periods[periodLast7Days] = PeriodReport{
		StartDate:    "7daysAgo",
		EndDate:      "today",
		Keywords:     validated7Days,
		KeywordLimit: limit,
	}

	last30Days, err := reporter.TopSearchTerms(ctx, "30daysAgo", "today", ga4CandidateLimit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 30 days: %w", err)
	}
	validated30Days, err := validateKeywords(ctx, last30Days, limit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 30 days: %w", err)
	}
	report.Periods[periodLast30Days] = PeriodReport{
		StartDate:    "30daysAgo",
		EndDate:      "today",
		Keywords:     validated30Days,
		KeywordLimit: limit,
	}

	last6Months, err := reporter.TopSearchTerms(ctx, "6monthsAgo", "today", ga4CandidateLimit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 6 months: %w", err)
	}
	validated6Months, err := validateKeywords(ctx, last6Months, limit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 6 months: %w", err)
	}
	report.Periods[periodLast6Months] = PeriodReport{
		StartDate:    "6monthsAgo",
		EndDate:      "today",
		Keywords:     validated6Months,
		KeywordLimit: limit,
	}

	last1Year, err := reporter.TopSearchTerms(ctx, "365daysAgo", "today", ga4CandidateLimit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 1 year: %w", err)
	}
	validated1Year, err := validateKeywords(ctx, last1Year, limit)
	if err != nil {
		return nil, fmt.Errorf("analyticskeywords: last 1 year: %w", err)
	}
	report.Periods[periodLast1Year] = PeriodReport{
		StartDate:    "365daysAgo",
		EndDate:      "today",
		Keywords:     validated1Year,
		KeywordLimit: limit,
	}

	return report, nil
}

func validateKeywords(ctx context.Context, terms []ga4.SearchTermCount, limit int) ([]KeywordCount, error) {
	merged := make(map[string]int64)
	for _, term := range terms {
		verifiedName, err := verifyCardNameFunc(ctx, term.Term)
		if err != nil {
			return nil, fmt.Errorf("verify %q: %w", term.Term, err)
		}
		if verifiedName == "" {
			continue
		}
		merged[verifiedName] += term.Count
	}

	keywords := make([]KeywordCount, 0, len(merged))
	for term, count := range merged {
		keywords = append(keywords, KeywordCount{
			Term:  term,
			Count: count,
		})
	}

	sort.Slice(keywords, func(i, j int) bool {
		if keywords[i].Count == keywords[j].Count {
			return keywords[i].Term < keywords[j].Term
		}
		return keywords[i].Count > keywords[j].Count
	})

	if len(keywords) > limit {
		keywords = keywords[:limit]
	}
	return keywords, nil
}
