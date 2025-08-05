package main

import (
    "fmt"
    "math/rand"
    "net/http"
    "time"
)

func StartFlakyServer() {
    rand.Seed(time.Now().UnixNano())
    http.HandleFunc("/flaky", func(w http.ResponseWriter, r *http.Request) {
        chance := rand.Intn(100)
        fmt.Printf("🌀 Incoming request, chance=%d\n", chance)

        switch {
        case chance < 30:
            http.Error(w, "500 internal error", http.StatusInternalServerError)
        case chance < 60:
            time.Sleep(2 * time.Second)
            fmt.Fprintln(w, "😴 slow response")
        default:
            fmt.Fprintln(w, "✅ success")
        }
    })

    fmt.Println("🚀 Flaky server running on :8081")
    http.ListenAndServe(":8081", nil)
}
