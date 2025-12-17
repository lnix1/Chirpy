package main

import (
	"net/http"
	"time"

	"github.com/lnix1/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type returnVals struct {
		Token 	string `json:"token"`
	}

	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No valid refresh token", err)
		return
	}

	dbBearer, err := cfg.db.GetRefreshToken(r.Context(), bearer)
	if err != nil || dbBearer.ExpiredBool == false || dbBearer.RevokeCheck == false {
		respondWithError(w, http.StatusUnauthorized, "No valid refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(dbBearer.UserID, cfg.secret, time.Duration(3600) * time.Second)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error generating bearer token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Token: accessToken,
	})
	return
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No valid refresh token", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), bearer)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No refresh token with provided value in database", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
	return
}
