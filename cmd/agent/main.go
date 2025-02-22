package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

const (
	serverURL      = "http://localhost:8080/update/"
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

// виды метрик - измеренное значение (gauge) и счетчик (counter)
type gauge float64
type counter int64

// структура хранения метрик
type Metrics struct {
	gauges   map[string]gauge
	counters map[string]counter
}

// создает новый объект с метриками
func NewMetrics() *Metrics {
	return &Metrics{
		gauges:   make(map[string]gauge),
		counters: make(map[string]counter),
	}
}

func (m *Metrics) updateMetrics() {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	m.gauges = map[string]gauge{
		"Alloc":         gauge(memStats.Alloc),
		"BuckHashSys":   gauge(memStats.BuckHashSys),
		"Frees":         gauge(memStats.Frees),
		"GCCPUFraction": gauge(memStats.GCCPUFraction),
		"GCSys":         gauge(memStats.GCSys),
		"HeapAlloc":     gauge(memStats.HeapAlloc),
		"HeapIdle":      gauge(memStats.HeapIdle),
		"HeapInuse":     gauge(memStats.HeapInuse),
		"HeapObjects":   gauge(memStats.HeapObjects),
		"HeapReleased":  gauge(memStats.HeapReleased),
		"HeapSys":       gauge(memStats.HeapSys),
		"LastGC":        gauge(memStats.LastGC),
		"Lookups":       gauge(memStats.Lookups),
		"MCacheInuse":   gauge(memStats.MCacheInuse),
		"MCacheSys":     gauge(memStats.MCacheSys),
		"MSpanInuse":    gauge(memStats.MSpanInuse),
		"MSpanSys":      gauge(memStats.MSpanSys),
		"Mallocs":       gauge(memStats.Mallocs),
		"NextGC":        gauge(memStats.NextGC),
		"NumForcedGC":   gauge(memStats.NumForcedGC),
		"NumGC":         gauge(memStats.NumGC),
		"OtherSys":      gauge(memStats.OtherSys),
		"PauseTotalNs":  gauge(memStats.PauseTotalNs),
		"StackInuse":    gauge(memStats.StackInuse),
		"StackSys":      gauge(memStats.StackSys),
		"Sys":           gauge(memStats.Sys),
		"TotalAlloc":    gauge(memStats.TotalAlloc),
		"RandomValue":   gauge(rand.Float64() * 100),
	}

	m.counters["PollCount"]++
}

func (m *Metrics) sendMetrics() {
	client := &http.Client{}
	for name, value := range m.gauges {
		url := serverURL + "gauge/" + name + "/" + strconv.FormatFloat(float64(value), 'f', -1, 64)
		resp, err := client.Post(url, "text/plain", nil)
		if err != nil {
			fmt.Println("Failed to send metric:", err)
			continue
		}
		resp.Body.Close()
	}

	for name, value := range m.counters {
		url := serverURL + "counter/" + name + "/" + strconv.FormatInt(int64(value), 10)
		resp, err := client.Post(url, "text/plain", nil)
		if err != nil {
			fmt.Println("Failed to send metric:", err)
			continue
		}
		resp.Body.Close()
	}
}

func (m *Metrics) Run() {
	metrics := NewMetrics()

	tickerPoll := time.NewTicker(pollInterval)
	tickerReport := time.NewTicker(reportInterval)

	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	for {
		select {
		case <-tickerPoll.C:
			metrics.updateMetrics()
		case <-tickerReport.C:
			metrics.sendMetrics()
		}
	}
}

func main() {
	agent := Metrics{}
	agent.Run()
}
