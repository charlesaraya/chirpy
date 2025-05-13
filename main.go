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
	apiCfg := handlers.ApiConfig{}

	// 2. Set up handlers
	mux.Handle("/app/", handlers.GetHome(&apiCfg, ".", "/app"))

	mux.HandleFunc("POST /api/validate_chirp", handlers.ValidateChirp)

	mux.HandleFunc("GET /api/healthz", handlers.GetHealth)

	mux.HandleFunc("GET /admin/metrics", handlers.GetMetrics(&apiCfg, handlers.MetricsTemplatePath))

	mux.HandleFunc("POST /admin/reset", handlers.ResetMetrics(&apiCfg))

	// 3. Start server
	server.ListenAndServe()
}
