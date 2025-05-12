package main

import (
	"net/http"

	"github.com/charlesaraya/chirpy/internal/handlers"
)

func main() {
	// 1. Create Server
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	// Set up handlers
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	mux.HandleFunc("GET /health", handlers.GetHealth)

	// Start server
	server.ListenAndServe()
}
