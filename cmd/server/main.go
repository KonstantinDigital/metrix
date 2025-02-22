package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// виды метрик - измеренное значение (gauge) и счетчик (counter)
type gauge float64
type counter int64

// интерфейс хранилища метрик
type Storage interface {
	Update(metricType, name, value string) error
	GetMetrics() []string
	GetCounter(name string) (counter, error)
	GetGauge(name string) (gauge, error)
}

// хранилище метрик
type MemStorage struct {
	gauges   map[string]gauge
	counters map[string]counter
}

// возвращаем значение по типу и имени метрики
func (ms *MemStorage) GetCounter(n string) (counter, error) {
	v, ok := ms.counters[n]
	if ok {
		return v, nil
	}
	return v, fmt.Errorf("invalid counter name")
}

// возвращаем значение по типу и имени метрики
func (ms *MemStorage) GetGauge(n string) (gauge, error) {
	v, ok := ms.gauges[n]
	if ok {
		return v, nil
	}
	return v, fmt.Errorf("invalid gauge name")
}

// метод возвращает список метрик и их значений
func (ms *MemStorage) GetMetrics() []string {
	var list []string
	for name, value := range ms.gauges {
		raw := fmt.Sprintf("%s: %s", name, strconv.FormatFloat(float64(value), 'f', -1, 64))
		list = append(list, raw)
	}

	return list
}

// реализация метода Update интерфейса Storage
func (ms *MemStorage) Update(metricType, name, value string) error {
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

// HTTP обработчик эндпоинта /update/<metric_type>/<metric_name>/<metric_value>
func updateHandler(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		metricValue := chi.URLParam(req, "value")

		err := storage.Update(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

func getMetric(s Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		t := chi.URLParam(req, "type")
		n := chi.URLParam(req, "name")

		switch t {
		case "gauge":
			val, err := s.GetGauge(n)
			if err != nil {
				http.Error(res, err.Error(), http.StatusNotFound)
				return
			}
			res.Write([]byte(strconv.FormatFloat(float64(val), 'f', -1, 64)))
		case "counter":
			val, err := s.GetCounter(n)
			if err != nil {
				http.Error(res, err.Error(), http.StatusNotFound)
				return
			}
			res.Write([]byte(strconv.FormatInt(int64(val), 10)))
		default:
			http.Error(res, "invalid metric type", http.StatusNotFound)
		}
	}
}

func allMetrics(s Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		var sb strings.Builder
		sb.WriteString("<html><head><title>Metrics</title></head><body>")
		sb.WriteString("<h1>Current Metrics</h1><ul>")
		for _, metric := range s.GetMetrics() {
			sb.WriteString("<li>" + metric + "</li>")
		}
		sb.WriteString("</ul></body></html>")
		res.Write([]byte(sb.String()))
		// io.WriteString(res, strings.Join(s.GetMetrics(), ", "))
	}
}

// обработчик эндпоинтов
func HandleRouter(s Storage) chi.Router {
	r := chi.NewRouter()

	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", updateHandler(s))
	})

	r.Route("/", func(r chi.Router) {
		r.Get("/", allMetrics(s))
		r.Get("/value/{type}/{name}", getMetric(s))
	})

	return r
}

func main() {
	var storage Storage = &MemStorage{
		gauges:   make(map[string]gauge),
		counters: make(map[string]counter),
	}

	log.Fatal(http.ListenAndServe(":8080", HandleRouter(storage)))
}
