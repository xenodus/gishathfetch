package analyticskeywords

import (
	"strings"
	"testing"
)

func TestCollectUniqueSearchTerms_DeduplicatesAcrossPeriods(t *testing.T) {
	report := &Report{
		Periods: map[string]PeriodReport{
			periodLast24Hours: {
				Keywords: []KeywordCount{{Term: "Opt", Count: 4}},
			},
			periodLast7Days: {
				Keywords: []KeywordCount{
					{Term: "Opt", Count: 10},
					{Term: "Lightning Bolt", Count: 8},
				},
			},
			periodLast30Days: {
				Keywords: []KeywordCount{{Term: "Sol Ring", Count: 20}},
			},
		},
	}

	terms := CollectUniqueSearchTerms(report)
	if len(terms) != 3 {
		t.Fatalf("expected 3 unique terms, got %d: %v", len(terms), terms)
	}
	if terms[0] != "Lightning Bolt" || terms[1] != "Opt" || terms[2] != "Sol Ring" {
		t.Fatalf("unexpected sorted terms: %v", terms)
	}
}

func TestBuildSearchPageURL_EncodesQuery(t *testing.T) {
	got := BuildSearchPageURL("https://gishathfetch.com/", "Lightning Bolt")
	want := "https://gishathfetch.com/?s=Lightning+Bolt"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBuildRobotsTxt_IncludesUniqueKeywordURLs(t *testing.T) {
	report := &Report{
		Periods: map[string]PeriodReport{
			periodLast24Hours: {
				Keywords: []KeywordCount{{Term: "Opt", Count: 4}},
			},
			periodLast7Days: {
				Keywords: []KeywordCount{
					{Term: "Opt", Count: 10},
					{Term: "Lightning Bolt", Count: 8},
				},
			},
		},
	}

	robotsTxt := BuildRobotsTxt(report, "https://gishathfetch.com/")
	if !strings.HasPrefix(robotsTxt, RobotsTxtBase) {
		t.Fatalf("expected robots.txt to start with base policy")
	}
	if strings.Count(robotsTxt, "Allow: https://gishathfetch.com/?s=") != 2 {
		t.Fatalf("expected 2 keyword allow URLs, got:\n%s", robotsTxt)
	}
	if !strings.Contains(robotsTxt, "Allow: https://gishathfetch.com/?s=Lightning+Bolt\n") {
		t.Fatalf("missing Lightning Bolt URL in:\n%s", robotsTxt)
	}
	if !strings.Contains(robotsTxt, "Allow: https://gishathfetch.com/?s=Opt\n") {
		t.Fatalf("missing Opt URL in:\n%s", robotsTxt)
	}
}

func TestBuildRobotsTxt_ReturnsBaseWhenNoKeywords(t *testing.T) {
	robotsTxt := BuildRobotsTxt(&Report{}, "https://gishathfetch.com/")
	if robotsTxt != RobotsTxtBase {
		t.Fatalf("expected base robots.txt only, got:\n%s", robotsTxt)
	}
}
