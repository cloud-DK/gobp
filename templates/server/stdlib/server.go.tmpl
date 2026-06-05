package server

import (
	"encoding/json"
	"net/http"
	"runtime"
)

func StartServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{"status": "ok", "Active Goroutines": string(rune(runtime.NumGoroutine()))}
		jsonResponse, _ := json.Marshal(response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	})

	http.ListenAndServe(":8080", nil)
}
