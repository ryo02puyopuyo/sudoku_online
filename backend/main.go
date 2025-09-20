package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
        // CORS 許可
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        fmt.Fprintln(w, "pong")
    })

    fmt.Println("Go server running on http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}