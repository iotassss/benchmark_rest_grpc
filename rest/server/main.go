// server.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type EchoRequest struct {
	Message string `json:"message"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	// JSON decode もオーバーヘッドに含める
	var req EchoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	_ = req // シンプルなビジネスロジックの代わり

	resp := EchoResponse{Message: "ok"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		log.Println("encode error:", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", echoHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("REST server listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
