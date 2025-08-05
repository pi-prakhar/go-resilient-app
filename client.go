package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/slok/goresilience"
	bulkheadmw "github.com/slok/goresilience/bulkhead"
	circuitbreakermw "github.com/slok/goresilience/circuitbreaker"
	prometheusmw "github.com/slok/goresilience/metrics/prometheus"
	parallelmw "github.com/slok/goresilience/parallel"
	ratelimitmw "github.com/slok/goresilience/ratelimit"
	retrymw "github.com/slok/goresilience/retry"
)

func DoResilientCall(promReg *prometheus.Registry) {
	// Metrics middleware
	metrics := prometheusmw.NewMiddleware(prometheusmw.Config{
		Recorder: prometheusmw.NewRecorder(prometheusmw.RecorderConfig{
			Registry: promReg,
		}),
	})

	executor := goresilience.NewExecutor(
		metrics,
		retrymw.Middleware(retrymw.Config{
			Times:    3,
			WaitBase: 200 * time.Millisecond,
			WaitMax:  1 * time.Second,
			Jitter:   true,
		}),
		circuitbreakermw.Middleware(circuitbreakermw.Config{
			FailureThreshold:        0.5,
			MinimumRequestToTrip:    5,
			WaitDurationInOpenState: 10 * time.Second,
		}),
		ratelimitmw.Middleware(ratelimitmw.Config{
			Limit: 10,
			Every: time.Second,
		}),
		bulkheadmw.Middleware(bulkheadmw.Config{
			MaxConcurrentCalls: 5,
		}),
		parallelmw.Middleware(parallelmw.Config{
			MaxParallel:   2,
			StopOnSuccess: true,
		}),
	)

	err := executor.Run(context.Background(), func(ctx context.Context) error {
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://localhost:8081/flaky")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			return errors.New("received 5xx from flaky service")
		}

		fmt.Println("🎉 Resilient call succeeded")
		return nil
	})

	if err != nil {
		fmt.Printf("❌ Resilient call failed: %v\n", err)
	}
}
