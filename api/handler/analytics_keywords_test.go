package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"mtg-price-checker-sg/controller/analyticskeywords"
	"mtg-price-checker-sg/gateway/ga4"
	"mtg-price-checker-sg/store/analyticsreport"
)

func TestHandle_RoutesAnalyticsKeywordsExportRun(t *testing.T) {
	originalReporterFunc := newGA4ReporterFunc
	originalBuildFunc := buildAnalyticsKeywordsReportFunc
	originalWriterFunc := newAnalyticsReportWriterFunc
	originalAlertFunc := sendJobAlert
	defer func() {
		newGA4ReporterFunc = originalReporterFunc
		buildAnalyticsKeywordsReportFunc = originalBuildFunc
		newAnalyticsReportWriterFunc = originalWriterFunc
		sendJobAlert = originalAlertFunc
	}()

	sendJobAlert = func(string) {}

	writer := &mockAnalyticsWriter{}
	newGA4ReporterFunc = func(_ context.Context) (ga4.Reporter, error) {
		return &mockAnalyticsReporter{}, nil
	}
	buildAnalyticsKeywordsReportFunc = func(_ context.Context, _ ga4.Reporter, propertyID string, _ int) (*analyticskeywords.Report, error) {
		return &analyticskeywords.Report{
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			PropertyID:  propertyID,
			EventName:   ga4.SearchEventName,
		}, nil
	}
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

func TestRunAnalyticsKeywordsExport_SendsSuccessAlert(t *testing.T) {
	originalReporterFunc := newGA4ReporterFunc
	originalBuildFunc := buildAnalyticsKeywordsReportFunc
	originalWriterFunc := newAnalyticsReportWriterFunc
	originalAlertFunc := sendJobAlert
	defer func() {
		newGA4ReporterFunc = originalReporterFunc
		buildAnalyticsKeywordsReportFunc = originalBuildFunc
		newAnalyticsReportWriterFunc = originalWriterFunc
		sendJobAlert = originalAlertFunc
	}()

	var gotAlert string
	sendJobAlert = func(message string) {
		gotAlert = message
	}

	writer := &mockAnalyticsWriter{}
	newGA4ReporterFunc = func(_ context.Context) (ga4.Reporter, error) {
		return &mockAnalyticsReporter{}, nil
	}
	buildAnalyticsKeywordsReportFunc = func(_ context.Context, _ ga4.Reporter, propertyID string, _ int) (*analyticskeywords.Report, error) {
		return &analyticskeywords.Report{
			GeneratedAt: "2026-07-11T12:00:00Z",
			PropertyID:  propertyID,
			EventName:   ga4.SearchEventName,
		}, nil
	}
	newAnalyticsReportWriterFunc = func(_ context.Context) (analyticsreport.Writer, error) {
		return writer, nil
	}

	if err := runAnalyticsKeywordsExport(context.Background()); err != nil {
		t.Fatalf("runAnalyticsKeywordsExport: %v", err)
	}
	if !writer.written {
		t.Fatal("expected analytics report to be written")
	}

	want := "Analytics keywords export finished: generatedAt=2026-07-11T12:00:00Z"
	if gotAlert != want {
		t.Fatalf("alert = %q, want %q", gotAlert, want)
	}
}

func TestRunAnalyticsKeywordsExport_SendsFailureAlert(t *testing.T) {
	originalReporterFunc := newGA4ReporterFunc
	originalAlertFunc := sendJobAlert
	defer func() {
		newGA4ReporterFunc = originalReporterFunc
		sendJobAlert = originalAlertFunc
	}()

	reportErr := errors.New("ga4 credentials missing")
	var gotAlert string
	sendJobAlert = func(message string) {
		gotAlert = message
	}
	newGA4ReporterFunc = func(_ context.Context) (ga4.Reporter, error) {
		return nil, reportErr
	}

	if err := runAnalyticsKeywordsExport(context.Background()); err == nil {
		t.Fatal("expected error")
	}

	want := "Analytics keywords export failed: ga4 credentials missing"
	if gotAlert != want {
		t.Fatalf("alert = %q, want %q", gotAlert, want)
	}
}

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
