package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	counter := 0
	http.HandleFunc("/flaky", func(w http.ResponseWriter, r *http.Request) {
		counter++
		// chance := rand.Intn(100)
		chance := 61
		fmt.Printf("🌀 Incoming request, chance=%d %d\n", chance, counter)

		switch {
		case chance < 30:
			http.Error(w, "400 bad request", http.StatusBadRequest)
		case chance < 60:
			time.Sleep(2 * time.Second)
			fmt.Println("😴 slow response")
		default:
			fmt.Println("✅ success")
		}
	})

	fmt.Println("🚀 Flaky server running on :8081")
	http.ListenAndServe(":8081", nil)
}
