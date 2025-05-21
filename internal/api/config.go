package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/charlesaraya/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type ApiConfig struct {
	ServerHits  atomic.Int32
	DBQueries   *database.Queries
	Platform    string
	TokenSecret string
	PolkaApiKey string
}

func (cfg *ApiConfig) GetHits() int32 {
	hits := cfg.ServerHits.Load()
	return hits
}

func (cfg *ApiConfig) IncHits(next http.Handler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		cfg.ServerHits.Add(1)
		next.ServeHTTP(res, req)
	}
}

func (cfg *ApiConfig) ResetHits() {
	cfg.ServerHits = atomic.Int32{}
}

func Load() (*ApiConfig, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Print("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")

	// 0. Open DB connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, errors.New("error opening the database")
	}

	return &ApiConfig{
		DBQueries:   database.New(db),
		Platform:    os.Getenv("PLATFORM"),
		TokenSecret: os.Getenv("TOKEN_SECRET"),
		PolkaApiKey: os.Getenv("POLKA_API_KEY"),
	}, nil
}

type UserPayload struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Email        string `json:"email"`
	IsChirpyRed  bool   `json:"is_chirpy_red"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}
