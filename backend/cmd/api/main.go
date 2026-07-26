package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

func main() {
	if isLambda() {
		startLambda()
	} else {
		startHTTPServer()
	}
}

func isLambda() bool {
	_, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME")
	return ok
}

// TODO: aws-lambda-go を組み込み、Lambda イベントを http.Request に変換して mux に流す。
func startLambda() {
	log.Println("Starting in Lambda mode (not yet implemented)")
}

func startHTTPServer() {
	addr := ":8080"
	srv := &http.Server{
		Addr:              addr,
		Handler:           newMux(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Printf("Starting HTTP server on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
