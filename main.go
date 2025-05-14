package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/charlesaraya/chirpy/internal/database"
	"github.com/charlesaraya/chirpy/internal/handlers"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Print("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")

	// 0. Open DB connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error opening the database")
	}
	dbQueries := database.New(db)

	apiCfg := handlers.ApiConfig{
		DBQueries: dbQueries,
		Platform:  platform,
	}
	// 1. Create Server
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	// 2. Set up handlers
	mux.Handle("/app/", handlers.GetHome(&apiCfg, ".", "/app"))

	mux.HandleFunc("POST /api/users", handlers.CreateUserHandler(&apiCfg))

	mux.HandleFunc("POST /api/chirps", handlers.CreateChirpHandler(&apiCfg))

	mux.HandleFunc("POST /api/validate_chirp", handlers.ValidateChirp)

	mux.HandleFunc("GET /api/healthz", handlers.GetHealth)

	mux.HandleFunc("GET /admin/metrics", handlers.GetMetrics(&apiCfg, handlers.MetricsTemplatePath))

	mux.HandleFunc("POST /admin/reset", handlers.ResetMetrics(&apiCfg))

	// 3. Start server
	server.ListenAndServe()
}
