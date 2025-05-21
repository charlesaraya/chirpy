package api

import (
	"encoding/json"
	"net/http"

	"github.com/charlesaraya/chirpy/internal/auth"
	"github.com/google/uuid"
)

const (
	validPolkaEvent string = "user.upgraded"
)

type polkaPayload struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func PolkaWebhookHandler(apiCfg *ApiConfig) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		apiKey, err := auth.GetApiKey(req.Header)
		if err != nil {
			http.Error(res, ErrorUnauthorized, http.StatusUnauthorized)
			return
		}
		if apiKey != apiCfg.PolkaApiKey {
			http.Error(res, ErrorUnauthorized, http.StatusUnauthorized)
			return
		}
		reqPayload := polkaPayload{}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&reqPayload); err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		if reqPayload.Event != validPolkaEvent {
			res.WriteHeader(http.StatusNoContent)
			return
		}
		userUUID, err := uuid.Parse(reqPayload.Data.UserID)
		if err != nil {
			http.Error(res, ErrorInternalServerError, http.StatusInternalServerError)
			return
		}
		_, err = apiCfg.DBQueries.UpgradeUser(req.Context(), userUUID)
		if err != nil {
			http.Error(res, ErrorNotFound, http.StatusNotFound)
			return
		}
		res.WriteHeader(http.StatusNoContent)
	}
}
