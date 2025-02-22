package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Мок хранилища для тестирования
type MockStorage struct {
	gauges   map[string]gauge
	counters map[string]counter
}

func (ms *MockStorage) Update(metricType, name, value string) error {
	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid gauge value")
		}
		ms.gauges[name] = gauge(v)
	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter value")
		}
		ms.counters[name] += counter(v)
	default:
		return fmt.Errorf("invalid metric type")
	}
	return nil
}

func TestUpdateHandler(t *testing.T) {
	mockStorage := &MockStorage{
		gauges:   make(map[string]gauge),
		counters: make(map[string]counter),
	}

	handler := updateHandler(mockStorage)
	tests := []struct {
		name       string
		method     string
		url        string
		wantStatus int
	}{
		{
			name:       "Valid gauge metric",
			method:     http.MethodPost,
			url:        "/update/gauge/cpu/4.5",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Valid counter metric",
			method:     http.MethodPost,
			url:        "/update/counter/requests/10",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid metric type",
			method:     http.MethodPost,
			url:        "/update/invalid/metric/123",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid value for gauge",
			method:     http.MethodPost,
			url:        "/update/gauge/memory/abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid value for counter",
			method:     http.MethodPost,
			url:        "/update/counter/events/xyz",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid method",
			method:     http.MethodGet,
			url:        "/update/gauge/memory/3.3",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()
			handler(rr, req)

			res := rr.Result()
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)

			assert.Equal(t, tt.wantStatus, res.StatusCode, "Response: %s", string(body))
		})
	}
}
