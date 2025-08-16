package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/slok/goresilience"
	"github.com/slok/goresilience/circuitbreaker"
	rerror "github.com/slok/goresilience/errors"
	"github.com/slok/goresilience/retry"
	"github.com/slok/goresilience/timeout"
)

// Custom error type for ignorable errors (like 400s).

var runner = goresilience.RunnerChain(
	retry.NewMiddleware(retry.Config{
		Times:          3,
		DisableBackoff: false,
		WaitBase:       100 * time.Millisecond,
	}),
	timeout.NewMiddleware(timeout.Config{
		Timeout: 1 * time.Second,
	}),
	circuitbreaker.NewMiddleware(circuitbreaker.Config{
		ErrorPercentThresholdToOpen: 50,
		MinimumRequestToOpen:        4,
	}),
)

func DoResilientCall() {
	err := runner.Run(context.Background(), func(ctx context.Context) error {
		client := http.Client{}
		resp, err := client.Get("http://flaky:8081/flaky")
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			fmt.Printf("🔄 Ignoring 4xx error: %d\n", resp.StatusCode)
			return nil
		}
		if resp.StatusCode >= 500 {
			return errors.New("5xx from flaky service")
		}
		return errors.New("unexpected status code: " + resp.Status)
	})

	// Handle the error at the caller side
	if err != nil {
		if errors.Is(err, rerror.ErrCircuitOpen) {
			fmt.Println("🔒 Circuit breaker is open")
		} else {
			fmt.Printf("❌ Resilient call failed: %v\n", err)
		}
	} else {
		fmt.Println("✅ Resilient call succeeded")
	}
}

func main() {
	go http.ListenAndServe(":9090", nil)

	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		DoResilientCall()
	}
}
