package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	go StartFlakyServer()

	// Setup Prometheus
	reg := prometheus.NewRegistry()
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":9090", nil)
	fmt.Println("📊 Prometheus metrics on :9090/metrics")

	// Run resilient calls every 3s
	ticker := time.NewTicker(3 * time.Second)
	for range ticker.C {
		DoResilientCall(reg)
	}
}
