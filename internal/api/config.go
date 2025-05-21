package api

import (
	"net/http"
	"sync/atomic"

	"github.com/charlesaraya/chirpy/internal/database"
)

type ApiConfig struct {
	ServerHits  atomic.Int32
	DBQueries   *database.Queries
	Platform    string
	TokenSecret string
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
