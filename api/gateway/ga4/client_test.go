package ga4

import (
	"testing"

	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
)

func TestParseSearchTermRows_FiltersInvalidTerms(t *testing.T) {
	rows := []*analyticsdata.Row{
		{
			DimensionValues: []*analyticsdata.DimensionValue{{Value: "Lightning Bolt"}},
			MetricValues:    []*analyticsdata.MetricValue{{Value: "12"}},
		},
		{
			DimensionValues: []*analyticsdata.DimensionValue{{Value: notSetSearchTerm}},
			MetricValues:    []*analyticsdata.MetricValue{{Value: "5"}},
		},
		{
			DimensionValues: []*analyticsdata.DimensionValue{{Value: "  "}},
			MetricValues:    []*analyticsdata.MetricValue{{Value: "3"}},
		},
	}

	got := parseSearchTermRows(rows)
	if len(got) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got))
	}
	if got[0].Term != "Lightning Bolt" || got[0].Count != 12 {
		t.Fatalf("unexpected result: %+v", got[0])
	}
}

func TestAggregateSearchTermRows_SumsByTerm(t *testing.T) {
	rows := []*analyticsdata.Row{
		{
			DimensionValues: []*analyticsdata.DimensionValue{{Value: "Opt"}, {Value: "2026062712"}},
			MetricValues:    []*analyticsdata.MetricValue{{Value: "2"}},
		},
		{
			DimensionValues: []*analyticsdata.DimensionValue{{Value: "Opt"}, {Value: "2026062713"}},
			MetricValues:    []*analyticsdata.MetricValue{{Value: "3"}},
		},
		{
			DimensionValues: []*analyticsdata.DimensionValue{{Value: "Sol Ring"}, {Value: "2026062713"}},
			MetricValues:    []*analyticsdata.MetricValue{{Value: "4"}},
		},
	}

	got := topSearchTerms(aggregateSearchTermRows(rows), 20)
	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}
	if got[0].Term != "Opt" || got[0].Count != 5 {
		t.Fatalf("expected Opt=5, got %+v", got[0])
	}
	if got[1].Term != "Sol Ring" || got[1].Count != 4 {
		t.Fatalf("expected Sol Ring=4, got %+v", got[1])
	}
}

func TestTopSearchTerms_RespectsLimit(t *testing.T) {
	aggregated := map[string]int64{
		"a": 1,
		"b": 3,
		"c": 2,
	}

	got := topSearchTerms(aggregated, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(got))
	}
	if got[0].Term != "b" || got[1].Term != "c" {
		t.Fatalf("unexpected order: %+v", got)
	}
}
