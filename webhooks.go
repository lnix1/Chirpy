package main

import (
	"net/http"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lnix1/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUpgradeUser(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Api key missing or incorrect", err)
	}

	type payload struct {
		UserId		string `json:"user_id"`
	}
	type paramaters struct {
		Event 		string `json:"event"`
		Data		payload	`json:"data"`
	}

	params := paramaters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}
	
	userID, err := uuid.Parse(params.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not a real user ID", err)
		return
	}

	_, err = cfg.db.UpgradeUserAccount(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to upgrade user account status", err)
		return
	}
	
	respondWithJSON(w, http.StatusNoContent, nil)
	return
}
