package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUpdateMetrics(t *testing.T) {
	metrics := NewMetrics()
	metrics.updateMetrics()

	if len(metrics.gauges) == 0 {
		t.Errorf("Expected gauges to be populated, but got empty map")
	}

	if metrics.counters["PollCount"] == 0 {
		t.Errorf("Expected PollCount to be incremented, but got 0")
	}
}

func TestSendMetrics(t *testing.T) {
	metrics := NewMetrics()
	metrics.updateMetrics()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/update/") {
			t.Errorf("Unexpected request path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	metrics.sendMetrics()
}

func BenchmarkUpdateMetrics(b *testing.B) {
	metrics := NewMetrics()
	for i := 0; i < b.N; i++ {
		metrics.updateMetrics()
	}
}

func BenchmarkSendMetrics(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	metrics := NewMetrics()
	metrics.updateMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.sendMetrics()
	}
}
