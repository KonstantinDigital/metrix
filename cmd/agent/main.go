package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
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

func (m *Metrics) sendMetrics(serverURL string) {
	client := &http.Client{}
	prefix := "http://"
	serverURL = prefix + serverURL
	serverURL += "/update/"
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

func (m *Metrics) Run(serverURL string, pi time.Duration, ri time.Duration) {
	metrics := NewMetrics()

	tickerPoll := time.NewTicker(pi)
	tickerReport := time.NewTicker(ri)

	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	for {
		select {
		case <-tickerPoll.C:
			metrics.updateMetrics()
		case <-tickerReport.C:
			metrics.sendMetrics(serverURL)
		}
	}
}

func CheckAddr(a string) error {
	hp := strings.Split(a, ":")
	if len(hp) != 2 {
		return fmt.Errorf("address must be <host>:<port>")
	}
	return nil
}

func main() {
	serverURL := flag.String("a", "localhost:8080", "Net address host:port")
	pollInterval := flag.Int("p", 2, "Poll interval in seconds")
	reportInterval := flag.Int("r", 10, "Report interval in seconds")
	flag.Parse()

	err := CheckAddr(*serverURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	pi := time.Duration(*pollInterval) * time.Second
	ri := time.Duration(*reportInterval) * time.Second

	agent := Metrics{}
	agent.Run(*serverURL, pi, ri)
}
