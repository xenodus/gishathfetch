package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mtg-price-checker-sg/controller/analyticskeywords"
	"mtg-price-checker-sg/gateway/ga4"
	"mtg-price-checker-sg/store/analyticsreport"
)

type mockAnalyticsReporter struct{}

func (m *mockAnalyticsReporter) TopSearchTerms(_ context.Context, _, _ string, _ int) ([]ga4.SearchTermCount, error) {
	return []ga4.SearchTermCount{{Term: "Opt", Count: 1}}, nil
}

func (m *mockAnalyticsReporter) TopSearchTermsLast24Hours(_ context.Context, _ time.Time, _ int) ([]ga4.SearchTermCount, error) {
	return []ga4.SearchTermCount{{Term: "Opt", Count: 1}}, nil
}

type mockAnalyticsWriter struct {
	written bool
}

func (m *mockAnalyticsWriter) Write(_ context.Context, _ *analyticskeywords.Report) error {
	m.written = true
	return nil
}

func TestHandle_RoutesAnalyticsKeywordsExportRun(t *testing.T) {
	originalReporterFunc := newGA4ReporterFunc
	originalBuildFunc := buildAnalyticsKeywordsReportFunc
	originalWriterFunc := newAnalyticsReportWriterFunc
	defer func() {
		newGA4ReporterFunc = originalReporterFunc
		buildAnalyticsKeywordsReportFunc = originalBuildFunc
		newAnalyticsReportWriterFunc = originalWriterFunc
	}()

	writer := &mockAnalyticsWriter{}
	newGA4ReporterFunc = func(_ context.Context) (ga4.Reporter, error) {
		return &mockAnalyticsReporter{}, nil
	}
	buildAnalyticsKeywordsReportFunc = analyticskeywords.BuildReport
	newAnalyticsReportWriterFunc = func(_ context.Context) (analyticsreport.Writer, error) {
		return writer, nil
	}

	event, err := json.Marshal(map[string]string{"action": analyticsKeywordsExportRunAction})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	if _, err = Handle(context.Background(), event); err != nil {
		t.Fatalf("handle event: %v", err)
	}
	if !writer.written {
		t.Fatalf("expected analytics report to be written")
	}
}
