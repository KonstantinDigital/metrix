package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// вспомогательная функция для установки chi-контекста в запрос
func newChiRequest(method, url string, params map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, nil)
	chiCtx := chi.NewRouteContext()
	for key, value := range params {
		chiCtx.URLParams.Add(key, value)
	}
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx) // исправлено
	return req.WithContext(ctx)
}

// тестируем updateHandler
func TestUpdateHandler(t *testing.T) {
	storage := &MemStorage{
		gauges:   make(map[string]gauge),
		counters: make(map[string]counter),
	}

	handler := updateHandler(storage)

	req := newChiRequest(http.MethodPost, "/update/gauge/cpu/75.5", map[string]string{
		"type":  "gauge",
		"name":  "cpu",
		"value": "75.5",
	})

	rec := httptest.NewRecorder()
	handler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	val, err := storage.GetGauge("cpu")
	assert.NoError(t, err)
	assert.Equal(t, gauge(75.5), val)
}

// тестируем getMetric
func TestGetMetric(t *testing.T) {
	storage := &MemStorage{
		gauges: map[string]gauge{"cpu": 75.5},
		counters: map[string]counter{
			"requests": 10,
		},
	}

	handler := getMetric(storage)

	req := newChiRequest(http.MethodGet, "/value/gauge/cpu", map[string]string{
		"type": "gauge",
		"name": "cpu",
	})

	rec := httptest.NewRecorder()
	handler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "75.5", rec.Body.String())

	req = newChiRequest(http.MethodGet, "/value/counter/requests", map[string]string{
		"type": "counter",
		"name": "requests",
	})

	rec = httptest.NewRecorder()
	handler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "10", rec.Body.String())
}

// тестируем allMetrics
func TestAllMetrics(t *testing.T) {
	storage := &MemStorage{
		gauges: map[string]gauge{"cpu": 75.5},
	}

	handler := allMetrics(storage)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "cpu: 75.5")
}
