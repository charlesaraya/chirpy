package handlers

import (
	"fmt"
	"net/http"
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
		res.Header().Add("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(200)
		msg := fmt.Sprintf("Hits: %v", apiCfg.getHits())
		res.Write([]byte(msg))
	}
}
