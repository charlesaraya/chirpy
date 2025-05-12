package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
)

const (
	HealthOK string = "OK"
)

type ApiConfig struct {
	ServerHits atomic.Int32
}

func (cfg *ApiConfig) getHits() int32 {
	hits := cfg.ServerHits.Load()
	return hits
}

func (cfg *ApiConfig) incHits(next http.Handler) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		cfg.ServerHits.Add(1)
		next.ServeHTTP(res, req)
	}
}

func (cfg *ApiConfig) resetHits() {
	cfg.ServerHits = atomic.Int32{}
}

func GetHome(apiCfg *ApiConfig, name string, prefix string) http.HandlerFunc {
	return apiCfg.incHits(http.StripPrefix(prefix, http.FileServer(http.Dir(name))))
}

func GetHealth(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte(HealthOK))
}

func GetMetrics(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		path := filepath.Join("templates", "metrics.html")
		rawTemplate, err := os.ReadFile(path)
		if err != nil {
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		msg := fmt.Sprintf(string(rawTemplate), apiCfg.getHits())
		res.Header().Add("Content-Type", "text/html; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(msg))
	}
}

func ResetMetrics(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		apiCfg.resetHits()
	}
}
