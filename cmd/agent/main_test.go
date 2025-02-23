package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheckAddr(t *testing.T) {
	tests := []struct {
		addr    string
		wantErr bool
	}{
		{"localhost:8080", false},
		{"127.0.0.1:9000", false},
		{"example.com:80", false},
		{"invalid_address", true},
	}

	for _, tt := range tests {
		err := CheckAddr(tt.addr)
		if (err != nil) != tt.wantErr {
			t.Errorf("CheckAddr(%q) error = %v, wantErr %v", tt.addr, err, tt.wantErr)
		}
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
	go agent.Run(strings.TrimPrefix(ts.URL, "http://"), time.Millisecond*100, time.Millisecond*200)
	time.Sleep(time.Millisecond * 500)
}
