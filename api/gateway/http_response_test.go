package gateway

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"testing"
)

func TestReadResponseBody_plainJSON(t *testing.T) {
	body := []byte(`{"ok":true}`)
	resp := &http.Response{
		Body:          io.NopCloser(bytes.NewReader(body)),
		Uncompressed:  true,
		Header:        make(http.Header),
	}
	got, err := ReadResponseBody(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(body) {
		t.Fatalf("got %q, want %q", got, body)
	}
}

func TestReadResponseBody_gzipJSON(t *testing.T) {
	plain := []byte(`{"resources":{"results":{"products":[]}}}`)
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	if _, err := zw.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	resp := &http.Response{
		Body: io.NopCloser(bytes.NewReader(compressed.Bytes())),
		Header: http.Header{
			"Content-Encoding": []string{"gzip"},
		},
	}
	got, err := ReadResponseBody(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(plain) {
		t.Fatalf("got %q, want %q", got, plain)
	}
}

func TestReadResponseBody_gzipMagicWithoutHeader(t *testing.T) {
	plain := []byte(`{"ok":true}`)
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	if _, err := zw.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	resp := &http.Response{
		Body:   io.NopCloser(bytes.NewReader(compressed.Bytes())),
		Header: make(http.Header),
	}
	got, err := ReadResponseBody(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(plain) {
		t.Fatalf("got %q, want %q", got, plain)
	}
}
