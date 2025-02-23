package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/caarlos0/env/v9"
)

func TestCheckAddr(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		expects bool
	}{
		{"valid address", "localhost:8080", false},
		{"missing port", "localhost:", true},
		{"missing host", ":8080", true},
		{"invalid port", "localhost:abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAddr(tt.input)
			if (err != nil) != tt.expects {
				t.Errorf("CheckAddr(%s) error = %v, wantErr %v", tt.input, err, tt.expects)
			}
		})
	}
}

func TestUpdateMetrics(t *testing.T) {
	m := NewMetrics()
	m.updateMetrics()

	if len(m.gauges) == 0 {
		t.Errorf("Expected gauges to be populated, got empty")
	}

	if m.counters["PollCount"] != 1 {
		t.Errorf("Expected PollCount to be 1, got %d", m.counters["PollCount"])
	}
}

func TestSendMetrics(t *testing.T) {
	m := NewMetrics()
	m.updateMetrics()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/update/") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	m.sendMetrics(strings.TrimPrefix(ts.URL, "http://"))
}

func TestRun(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	agent := Metrics{}
	go agent.Run(strings.TrimPrefix(ts.URL, "http://"), 1*time.Second, 2*time.Second)
	time.Sleep(time.Millisecond * 500)
}
func TestConfigParsing(t *testing.T) {
	os.Setenv("ADDRESS", "127.0.0.1:9090")
	os.Setenv("POLL_INTERVAL", "5")
	os.Setenv("REPORT_INTERVAL", "15")

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		t.Fatalf("Failed to parse env vars: %v", err)
	}

	if *cfg.Address != "127.0.0.1:9090" {
		t.Errorf("Expected ADDRESS to be 127.0.0.1:9090, got %s", *cfg.Address)
	}
	if *cfg.PollInterval != 5 {
		t.Errorf("Expected POLL_INTERVAL to be 5, got %d", *cfg.PollInterval)
	}
	if *cfg.ReportInterval != 15 {
		t.Errorf("Expected REPORT_INTERVAL to be 15, got %d", *cfg.ReportInterval)
	}
}
