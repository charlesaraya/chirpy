package main

import (
	"log"
	"net/http"

	"github.com/charlesaraya/chirpy/internal/api"
)

func main() {
	apiCfg, err := api.Load()
	if err != nil {
		log.Fatal("error loading api config")
	}
	// 1. Create Server
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	// 2. Set up handlers
	mux.Handle("/app/", api.GetHomeHandler(apiCfg, "./app", "/app"))

	mux.HandleFunc("POST /api/users", api.CreateUserHandler(apiCfg))

	mux.HandleFunc("PUT /api/users", api.UpdateUserHandler(apiCfg))

	mux.HandleFunc("POST /api/login", api.LoginUserHandler(apiCfg))

	mux.HandleFunc("POST /api/chirps", api.CreateChirpHandler(apiCfg))

	mux.HandleFunc("GET /api/chirps", api.GetChirpsHandler(apiCfg))

	mux.HandleFunc("GET /api/chirps/{chirpID}", api.GetSingleChirpHandler(apiCfg))

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", api.DeleteChirpHandler(apiCfg))

	mux.HandleFunc("POST /api/refresh", api.RefreshTokenHandler(apiCfg))

	mux.HandleFunc("POST /api/revoke", api.RevokeTokenHandler(apiCfg))

	mux.HandleFunc("POST /api/validate_chirp", api.ValidateChirpHandler)

	mux.HandleFunc("GET /api/healthz", api.GetHealthHandler)

	mux.HandleFunc("GET /admin/metrics", api.GetMetricsHandler(apiCfg, api.MetricsTemplatePath))

	mux.HandleFunc("POST /admin/reset", api.ResetMetricsHandler(apiCfg))

	// Webhooks
	mux.HandleFunc("POST /api/polka/webhooks", api.PolkaWebhookHandler(apiCfg))

	// 3. Start server
	server.ListenAndServe()
}
