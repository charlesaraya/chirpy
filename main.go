package main

import (
	"net/http"
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

	// Start server
	server.ListenAndServe()
}
