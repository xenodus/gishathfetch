package ga4

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/config"

	"golang.org/x/oauth2/google"
	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/option"
)

const (
	// SearchEventName matches the GA4 event sent by the frontend on search start.
	SearchEventName = "search"
	// SearchTermDimension is populated from the frontend search_term parameter.
	SearchTermDimension = "searchTerm"
	notSetSearchTerm    = "(not set)"
)

// SearchTermCount is a keyword and how many search events it received.
type SearchTermCount struct {
	Term  string `json:"term"`
	Count int64  `json:"count"`
}

// Reporter fetches top search keywords from GA4.
type Reporter interface {
	TopSearchTerms(ctx context.Context, startDate, endDate string, limit int) ([]SearchTermCount, error)
	TopSearchTermsLast24Hours(ctx context.Context, now time.Time, limit int) ([]SearchTermCount, error)
}

// Client queries the GA4 Data API for search event metrics.
type Client struct {
	service    *analyticsdata.Service
	propertyID string
}

var newAnalyticsDataService = func(ctx context.Context, credentialsJSON []byte) (*analyticsdata.Service, error) {
	credentials, err := google.CredentialsFromJSON(ctx, credentialsJSON, analyticsdata.AnalyticsReadonlyScope)
	if err != nil {
		return nil, err
	}
	return analyticsdata.NewService(ctx, option.WithCredentials(credentials))
}

// NewClient builds a GA4 Data API client from environment configuration.
func NewClient(ctx context.Context) (*Client, error) {
	propertyID := strings.TrimSpace(os.Getenv(config.GA4PropertyIDEnv))
	if propertyID == "" {
		return nil, fmt.Errorf("ga4: %s is not set", config.GA4PropertyIDEnv)
	}

	credentialsJSON := strings.TrimSpace(os.Getenv(config.GA4CredentialsJSONEnv))
	if credentialsJSON == "" {
		return nil, fmt.Errorf("ga4: %s is not set", config.GA4CredentialsJSONEnv)
	}

	service, err := newAnalyticsDataService(ctx, []byte(credentialsJSON))
	if err != nil {
		return nil, err
	}

	return &Client{
		service:    service,
		propertyID: propertyID,
	}, nil
}

func (c *Client) TopSearchTerms(ctx context.Context, startDate, endDate string, limit int) ([]SearchTermCount, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("ga4: limit must be positive")
	}

	request := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{
			{StartDate: startDate, EndDate: endDate},
		},
		Dimensions: []*analyticsdata.Dimension{{Name: SearchTermDimension}},
		Metrics:    []*analyticsdata.Metric{{Name: "eventCount"}},
		DimensionFilter: &analyticsdata.FilterExpression{
			Filter: &analyticsdata.Filter{
				FieldName: "eventName",
				StringFilter: &analyticsdata.StringFilter{
					MatchType: "EXACT",
					Value:     SearchEventName,
				},
			},
		},
		OrderBys: []*analyticsdata.OrderBy{{
			Desc: true,
			Metric: &analyticsdata.MetricOrderBy{
				MetricName: "eventCount",
			},
		}},
		Limit: int64(limit),
	}

	response, err := c.service.Properties.RunReport(c.propertyResource(), request).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return parseSearchTermRows(response.Rows), nil
}

func (c *Client) TopSearchTermsLast24Hours(ctx context.Context, now time.Time, limit int) ([]SearchTermCount, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("ga4: limit must be positive")
	}

	since := now.UTC().Add(-24 * time.Hour)
	sinceDateHour, err := strconv.ParseInt(since.Format("2006010215"), 10, 64)
	if err != nil {
		return nil, err
	}

	request := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{
			{StartDate: "2daysAgo", EndDate: "today"},
		},
		Dimensions: []*analyticsdata.Dimension{
			{Name: SearchTermDimension},
			{Name: "dateHour"},
		},
		Metrics: []*analyticsdata.Metric{{Name: "eventCount"}},
		DimensionFilter: &analyticsdata.FilterExpression{
			AndGroup: &analyticsdata.FilterExpressionList{
				Expressions: []*analyticsdata.FilterExpression{
					{
						Filter: &analyticsdata.Filter{
							FieldName: "eventName",
							StringFilter: &analyticsdata.StringFilter{
								MatchType: "EXACT",
								Value:     SearchEventName,
							},
						},
					},
					{
						Filter: &analyticsdata.Filter{
							FieldName: "dateHour",
							NumericFilter: &analyticsdata.NumericFilter{
								Operation: "GREATER_THAN_OR_EQUAL",
								Value: &analyticsdata.NumericValue{
									Int64Value: sinceDateHour,
								},
							},
						},
					},
				},
			},
		},
		Limit: 10000,
	}

	response, err := c.service.Properties.RunReport(c.propertyResource(), request).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	aggregated := aggregateSearchTermRows(response.Rows)
	return topSearchTerms(aggregated, limit), nil
}

func (c *Client) propertyResource() string {
	return "properties/" + c.propertyID
}

func parseSearchTermRows(rows []*analyticsdata.Row) []SearchTermCount {
	results := make([]SearchTermCount, 0, len(rows))
	for _, row := range rows {
		if row == nil || len(row.DimensionValues) == 0 || len(row.MetricValues) == 0 {
			continue
		}

		term := strings.TrimSpace(row.DimensionValues[0].Value)
		if !isValidSearchTerm(term) {
			continue
		}

		count, err := strconv.ParseInt(row.MetricValues[0].Value, 10, 64)
		if err != nil || count <= 0 {
			continue
		}

		results = append(results, SearchTermCount{
			Term:  term,
			Count: count,
		})
	}
	return results
}

func aggregateSearchTermRows(rows []*analyticsdata.Row) map[string]int64 {
	aggregated := make(map[string]int64)
	for _, row := range rows {
		if row == nil || len(row.DimensionValues) < 2 || len(row.MetricValues) == 0 {
			continue
		}

		term := strings.TrimSpace(row.DimensionValues[0].Value)
		if !isValidSearchTerm(term) {
			continue
		}

		count, err := strconv.ParseInt(row.MetricValues[0].Value, 10, 64)
		if err != nil || count <= 0 {
			continue
		}

		aggregated[term] += count
	}
	return aggregated
}

func topSearchTerms(aggregated map[string]int64, limit int) []SearchTermCount {
	results := make([]SearchTermCount, 0, len(aggregated))
	for term, count := range aggregated {
		results = append(results, SearchTermCount{
			Term:  term,
			Count: count,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Count == results[j].Count {
			return results[i].Term < results[j].Term
		}
		return results[i].Count > results[j].Count
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func isValidSearchTerm(term string) bool {
	return term != "" && term != notSetSearchTerm
}
