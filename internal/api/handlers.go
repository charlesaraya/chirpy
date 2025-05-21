package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charlesaraya/chirpy/internal/auth"
	"github.com/charlesaraya/chirpy/internal/database"
	"github.com/google/uuid"
)

const (
	HealthOK                 string        = "OK"
	MaxChirpLen              int           = 140
	ErrorChirpTooLong        string        = "Chirp is too long"
	ErrorSomethingWentWrong  string        = "Something went wrong"
	ErrorInternalServerError string        = "Internal Server Error"
	ErrorUnauthorized        string        = "Unauthorized"
	ErrorForbidden           string        = "Forbidden"
	ErrorNotFound            string        = "NotFound"
	MetricsTemplatePath      string        = "./templates/metrics.html"
	allowedPlatform          string        = "dev"
	TimeFormat               string        = "2006-01-02 15:04:05.000000"
	MaxSessionDuration       time.Duration = time.Hour
)

var ProfaneWords = []string{"kerfuffle", "sharbert", "fornax"}

type chirpPayload struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	UserID    string `json:"user_id"`
	Body      string `json:"body"`
}

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenPayload struct {
	AccessToken string `json:"token"`
}

func GetHomeHandler(apiCfg *ApiConfig, name string, prefix string) http.HandlerFunc {
	return apiCfg.IncHits(http.StripPrefix(prefix, http.FileServer(http.Dir(name))))
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
		msg := fmt.Sprintf(string(rawTemplate), apiCfg.GetHits())
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.Write([]byte(msg))
	}
}

func ResetMetricsHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if apiCfg.Platform == allowedPlatform {
			apiCfg.ResetHits()
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
		resBody := UserPayload{
			ID:          user.ID.String(),
			CreatedAt:   user.CreatedAt.String(),
			UpdatedAt:   user.UpdatedAt.String(),
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
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

func UpdateUserHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		params := loginPayload{}
		if err := decoder.Decode(&params); err != nil {
			http.Error(res, ErrorSomethingWentWrong, http.StatusBadRequest)
			return
		}
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(res, ErrorUnauthorized, http.StatusUnauthorized)
			return
		}
		userUUID, err := auth.ValidateJWT(token, apiCfg.TokenSecret)
		if err != nil {
			http.Error(res, ErrorUnauthorized, http.StatusUnauthorized)
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
		userParams := database.UpdateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
			ID:             userUUID,
		}
		user, err := apiCfg.DBQueries.UpdateUser(req.Context(), userParams)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
		}
		payload := UserPayload{
			ID:          user.ID.String(),
			CreatedAt:   user.CreatedAt.Format(TimeFormat),
			UpdatedAt:   user.UpdatedAt.Format(TimeFormat),
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
			Token:       token,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		res.Header().Add("Content-Type", "application/json")
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
			http.Error(res, ErrorUnauthorized, http.StatusUnauthorized)
			return
		}
		token, err := auth.MakeJWT(user.ID, apiCfg.TokenSecret, MaxSessionDuration)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		refreshToken, _ := auth.MakeRefreshToken()
		refereshTokensParams := database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
		}
		refreshTokenDB, err := apiCfg.DBQueries.CreateRefreshToken(req.Context(), refereshTokensParams)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		payload := UserPayload{
			ID:           user.ID.String(),
			CreatedAt:    user.CreatedAt.Format(TimeFormat),
			UpdatedAt:    user.UpdatedAt.Format(TimeFormat),
			Email:        user.Email,
			IsChirpyRed:  user.IsChirpyRed,
			Token:        token,
			RefreshToken: refreshTokenDB.Token,
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

func RefreshTokenHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		refreshToken, err := apiCfg.DBQueries.GetRefreshToken(req.Context(), token)
		if err != nil || refreshToken.ExpiresAt.Before(time.Now()) || refreshToken.RevokedAt.Valid {
			http.Error(res, ErrorUnauthorized, http.StatusUnauthorized)
			return
		}
		accessToken, err := auth.MakeJWT(refreshToken.UserID, apiCfg.TokenSecret, MaxSessionDuration)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		payload := tokenPayload{
			AccessToken: accessToken,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
		}
		res.Header().Set("Content-Type", "application/json")
		res.Write(data)
	}
}

func RevokeTokenHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		err = apiCfg.DBQueries.RevokeRefreshToken(req.Context(), token)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}

func CreateChirpHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		type reqPayload struct {
			Body string `json:"body"`
		}
		params := reqPayload{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&params); err != nil {
			http.Error(res, ErrorSomethingWentWrong, http.StatusBadRequest)
			return
		}
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		userUUID, err := auth.ValidateJWT(token, apiCfg.TokenSecret)
		if err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}
		chirpParams := database.CreateChirpParams{
			UserID: userUUID,
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
		var chirps []database.Chirp
		var err error
		authorID := req.URL.Query().Get("author_id")
		if authorID != "" {
			userUUID, err := uuid.Parse(authorID)
			if err != nil {
				http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
				return
			}
			chirps, err = apiCfg.DBQueries.GetChirpsFromUser(req.Context(), userUUID)
			if err != nil {
				http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
				return
			}
		} else {
			chirps, err = apiCfg.DBQueries.GetChirps(req.Context())
			if err != nil {
				http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
				return
			}
		}
		sortType := req.URL.Query().Get("sort")
		if sortType == "desc" {
			sort.Slice(chirps, func(i, j int) bool {
				return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
			})
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
			http.Error(res, ErrorNotFound, http.StatusNotFound)
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

func DeleteChirpHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		token, err := auth.GetBearerToken(req.Header)
		if err != nil {
			http.Error(res, ErrorUnauthorized, http.StatusUnauthorized)
			return
		}
		userUUID, err := auth.ValidateJWT(token, apiCfg.TokenSecret)
		if err != nil {
			http.Error(res, ErrorUnauthorized, http.StatusUnauthorized)
			return
		}
		chirpID, err := uuid.Parse(req.PathValue("chirpID"))
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		params := database.DeleteChirpParams{
			ID:     chirpID,
			UserID: userUUID,
		}
		result, err := apiCfg.DBQueries.DeleteChirp(req.Context(), params)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(res, ErrorForbidden, http.StatusForbidden)
			return
		}
		if rowsAffected == 0 {
			_, err := apiCfg.DBQueries.GetSingleChirp(req.Context(), chirpID)
			if err != nil {
				http.Error(res, ErrorNotFound, http.StatusNotFound)
				return
			}
			http.Error(res, ErrorForbidden, http.StatusForbidden)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}
