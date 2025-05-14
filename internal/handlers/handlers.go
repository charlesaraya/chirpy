package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

const (
	HealthOK                 string = "OK"
	MaxChirpLen              int    = 140
	ErrorChirpTooLong        string = "Chirp is too long"
	ErrorSomethingWentWrong  string = "Something went wrong"
	ErrorInternalServerError string = "Internal Server Error"
	MetricsTemplatePath      string = "./templates/metrics.html"
)

var ProfaneWords = []string{"kerfuffle", "sharbert", "fornax"}

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

func ValidateChirp(res http.ResponseWriter, req *http.Request) {
	type reqPayload struct {
		Body string `json:"body"`
	}
	type resErrorPayload struct {
		Error string `json:"error"`
	}
	type resPayload struct {
		CleanedBody string `json:"cleaned_body"`
	}
	decoder := json.NewDecoder(req.Body)
	pl := reqPayload{}
	if err := decoder.Decode(&pl); err != nil {
		http.Error(res, ErrorSomethingWentWrong, http.StatusBadRequest)
		return
	}
	if len(pl.Body) > 0 && len(pl.Body) <= MaxChirpLen {
		resBody := resPayload{
			CleanedBody: cleanProfanity(pl.Body),
		}
		data, err := json.Marshal(resBody)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write(data)
	} else {
		errorMsg := ""
		if len(pl.Body) == 0 {
			errorMsg = ErrorSomethingWentWrong
		} else {
			errorMsg = ErrorChirpTooLong
		}
		resBody := resErrorPayload{
			Error: errorMsg,
		}
		data, err := json.Marshal(resBody)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusBadRequest)
		res.Write(data)
	}
}

func cleanProfanity(chirp string) string {
	for _, word := range ProfaneWords {
		splitChirp := strings.Split(chirp, " ")
		chirp = ""
		for i, split := range splitChirp {
			if strings.ToLower(split) == word {
				split = "****"
			}
			chirp += split
			if i >= 0 && i < len(splitChirp)-1 {
				chirp += " "
			}
		}
	}
	return chirp
}

func GetHealth(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	res.Write([]byte(HealthOK))
}

func GetMetrics(apiCfg *ApiConfig, tmplPath string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		rawTemplate, err := os.ReadFile(tmplPath)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
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
