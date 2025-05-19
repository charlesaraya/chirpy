package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/charlesaraya/chirpy/internal/auth"
	"github.com/charlesaraya/chirpy/internal/database"
	"github.com/google/uuid"
)

const (
	HealthOK                    string        = "OK"
	MaxChirpLen                 int           = 140
	ErrorChirpTooLong           string        = "Chirp is too long"
	ErrorSomethingWentWrong     string        = "Something went wrong"
	ErrorInternalServerError    string        = "Internal Server Error"
	ErrorResourceNotFound       string        = "Resource Not Found"
	MetricsTemplatePath         string        = "./templates/metrics.html"
	allowedPlatform             string        = "dev"
	TimeFormat                  string        = "2006-01-02 15:04:05.000000"
	MaxSessionDurationInSeconds time.Duration = time.Hour
)

var ProfaneWords = []string{"kerfuffle", "sharbert", "fornax"}

type ApiConfig struct {
	ServerHits  atomic.Int32
	DBQueries   *database.Queries
	Platform    string
	TokenSecret string
}

type chirpPayload struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	UserID    string `json:"user_id"`
	Body      string `json:"body"`
}

type loginPayload struct {
	Email                 string `json:"email"`
	Password              string `json:"password"`
	SessionExpirationTime int    `json:"expires_in_seconds"`
}

type userPayload struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
	Token     string `json:"token"`
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

func GetHomeHandler(apiCfg *ApiConfig, name string, prefix string) http.HandlerFunc {
	return apiCfg.incHits(http.StripPrefix(prefix, http.FileServer(http.Dir(name))))
}

func ValidateChirpHandler(res http.ResponseWriter, req *http.Request) {
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
		res.Header().Set("Content-Type", "application/json")
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
		res.Header().Set("Content-Type", "application/json")
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

func GetHealthHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain; charset=utf-8")
	res.Write([]byte(HealthOK))
}

func GetMetricsHandler(apiCfg *ApiConfig, tmplPath string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		rawTemplate, err := os.ReadFile(tmplPath)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		msg := fmt.Sprintf(string(rawTemplate), apiCfg.getHits())
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.Write([]byte(msg))
	}
}

func ResetMetricsHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if apiCfg.Platform == allowedPlatform {
			apiCfg.resetHits()
			if err := apiCfg.DBQueries.DeleteUsers(req.Context()); err != nil {
				http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
				return
			}
			res.WriteHeader(http.StatusOK)
			return
		}
		res.WriteHeader(http.StatusForbidden)
	}
}

func CreateUserHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		params := loginPayload{}
		if err := decoder.Decode(&params); err != nil {
			http.Error(res, ErrorSomethingWentWrong, http.StatusBadRequest)
			return
		}
		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		if err := auth.CheckPasswordHash(hashedPassword, params.Password); err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		userParams := database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		}
		user, err := apiCfg.DBQueries.CreateUser(req.Context(), userParams)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		resBody := userPayload{
			ID:        user.ID.String(),
			CreatedAt: user.CreatedAt.String(),
			UpdatedAt: user.UpdatedAt.String(),
			Email:     user.Email,
		}
		data, err := json.Marshal(resBody)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}

func LoginUserHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		params := loginPayload{}
		if err := decoder.Decode(&params); err != nil {
			http.Error(res, ErrorSomethingWentWrong, http.StatusBadRequest)
			return
		}
		user, err := apiCfg.DBQueries.GetUser(req.Context(), params.Email)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		if err := auth.CheckPasswordHash(user.HashedPassword, params.Password); err != nil {
			http.Error(res, ErrorSomethingWentWrong, http.StatusUnauthorized)
			return
		}
		sessionDuration := MaxSessionDurationInSeconds
		if time.Duration(params.SessionExpirationTime) <= MaxSessionDurationInSeconds {
			sessionDuration = time.Duration(params.SessionExpirationTime)
		}
		token, err := auth.MakeJWT(user.ID, apiCfg.TokenSecret, sessionDuration)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		payload := userPayload{
			ID:        user.ID.String(),
			CreatedAt: user.CreatedAt.Format(TimeFormat),
			UpdatedAt: user.UpdatedAt.Format(TimeFormat),
			Email:     user.Email,
			Token:     token,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}

func CreateChirpHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		type reqPayload struct {
			UserID uuid.UUID `json:"user_id"`
			Body   string    `json:"body"`
		}
		params := reqPayload{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&params); err != nil {
			http.Error(res, ErrorSomethingWentWrong, http.StatusBadRequest)
			return
		}
		chirpParams := database.CreateChirpParams{
			UserID: uuid.UUID(params.UserID),
			Body:   params.Body,
		}
		chirp, err := apiCfg.DBQueries.CreateChirp(req.Context(), chirpParams)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		payload := chirpPayload{
			ID:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.Format(TimeFormat),
			UpdatedAt: chirp.UpdatedAt.Format(TimeFormat),
			UserID:    chirp.UserID.String(),
			Body:      chirp.Body,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
		}
		res.WriteHeader(http.StatusCreated)
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}

func GetChirpsHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		chirps, err := apiCfg.DBQueries.GetChirps(req.Context())
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		payload := make([]chirpPayload, len(chirps))
		for i, chirp := range chirps {
			payload[i] = chirpPayload{
				ID:        chirp.ID.String(),
				CreatedAt: chirp.CreatedAt.Format(TimeFormat),
				UpdatedAt: chirp.UpdatedAt.Format(TimeFormat),
				UserID:    chirp.UserID.String(),
				Body:      chirp.Body,
			}
		}
		data, err := json.Marshal(payload)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}

func GetSingleChirpHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		id, err := uuid.Parse(req.PathValue("chirpID"))
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		chirp, err := apiCfg.DBQueries.GetSingleChirp(req.Context(), id)
		if err != nil {
			http.Error(res, ErrorResourceNotFound, http.StatusNotFound)
			return
		}
		payload := chirpPayload{
			ID:        chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.Format(TimeFormat),
			UpdatedAt: chirp.UpdatedAt.Format(TimeFormat),
			UserID:    chirp.ID.String(),
			Body:      chirp.Body,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}
