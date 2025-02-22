package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// виды метрик - измеренное значение (gauge) и счетчик (counter)
type gauge float64
type counter int64

// интерфейс хранилища метрик
type Storage interface {
	Update(metricType, name, value string) error
}

// хранилище метрик
type MemStorage struct {
	gauges   map[string]gauge
	counters map[string]counter
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
		//разрешаем только POST
		if req.Method != http.MethodPost {
			http.Error(res, "405 only post requests are allowed", http.StatusMethodNotAllowed)
			return
		}

		pathURL := strings.TrimSuffix(req.URL.Path, "/")
		slicePath := strings.Split(pathURL, "/")

		if len(slicePath) < 5 {
			http.Error(res, "404 not Found", http.StatusNotFound)
			return
		}

		metricType := slicePath[2]
		metricName := slicePath[3]
		metricValue := slicePath[4]

		err := storage.Update(metricType, metricName, metricValue)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

func main() {
	var storage Storage = &MemStorage{
		gauges:   make(map[string]gauge),
		counters: make(map[string]counter),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateHandler(storage))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
