package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockStorage struct {
	gauges   map[string]gauge
	counters map[string]counter
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		gauges:   make(map[string]gauge),
		counters: make(map[string]counter),
	}
}

func (m *mockStorage) Update(metricType, name, value string) error {
	switch metricType {
	case "gauge":
		m.gauges[name] = gauge(42.5) // фиксированное значение для теста
	case "counter":
		m.counters[name]++
	default:
		return nil
	}
	return nil
}

func (m *mockStorage) GetMetrics() []string {
	return []string{"test_metric: 42.5"}
}

func (m *mockStorage) GetCounter(name string) (counter, error) {
	return m.counters[name], nil
}

func (m *mockStorage) GetGauge(name string) (gauge, error) {
	return m.gauges[name], nil
}

func TestMemStorageUpdate(t *testing.T) {
	storage := &MemStorage{
		gauges:   make(map[string]gauge),
		counters: make(map[string]counter),
	}

	tests := []struct {
		metricType string
		name       string
		value      string
		expects    bool
	}{
		{"gauge", "TestGauge", "42.5", false},
		{"counter", "TestCounter", "10", false},
		{"gauge", "InvalidGauge", "abc", true},
		{"counter", "InvalidCounter", "xyz", true},
		{"unknown", "Metric", "10", true},
	}

	for _, tt := range tests {
		err := storage.Update(tt.metricType, tt.name, tt.value)
		if (err != nil) != tt.expects {
			t.Errorf("Update(%s, %s, %s) error = %v, expects %v", tt.metricType, tt.name, tt.value, err, tt.expects)
		}
	}
}

func TestUpdateHandler(t *testing.T) {
	storage := newMockStorage()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_metric/42.5", nil)
	res := httptest.NewRecorder()

	handler := updateHandler(storage)
	handler(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.Code)
	}
}

func TestGetMetricHandler(t *testing.T) {
	storage := newMockStorage()
	storage.Update("gauge", "test_metric", "42.5")

	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", getMetric(storage))

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/test_metric", nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req) // Используем маршрутизатор для обработки запроса

	if res.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.Code)
	}

	if strings.TrimSpace(res.Body.String()) != "42.5" {
		t.Errorf("Expected response '42.5', got %s", res.Body.String())
	}
}

func TestAllMetricsHandler(t *testing.T) {
	storage := newMockStorage()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	handler := allMetrics(storage)
	handler(res, req)

	if res.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.Code)
	}

	if !strings.Contains(res.Body.String(), "test_metric: 42.5") {
		t.Errorf("Expected metrics in response, got %s", res.Body.String())
	}
}

func TestCheckAddr(t *testing.T) {
	tests := []struct {
		addr    string
		expects bool
	}{
		{"localhost:8080", false},
		{":8080", true},
		{"localhost:", true},
		{"localhost", true},
	}

	for _, tt := range tests {
		err := CheckAddr(tt.addr)
		if (err != nil) != tt.expects {
			t.Errorf("CheckAddr(%s) error = %v, expects %v", tt.addr, err, tt.expects)
		}
	}
}
