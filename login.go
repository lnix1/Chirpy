package main

import (
	"net/http"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lnix1/Chirpy/internal/auth"
	"github.com/lnix1/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password 	string	`json:"password"`
		Email		string	`json:"email"`
	}
	type returnVals struct {
		Id		uuid.UUID 	`json:"id"`
		Created_at 	time.Time	`json:"created_at"`
		Updated_at 	time.Time	`json:"updated_at"`
		Email 		string		`json:"email"`
		Token		string		`json:"token"`
		RefreshToken	string		`json:"refresh_token"`
		ChirpyRed	bool		`json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	
	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	passCheck, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if passCheck == false || err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	
	accessToken, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(3600) * time.Second)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error generating bearer token", err)
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: refreshToken, UserID: user.ID})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Id: user.ID,
		Created_at: user.CreatedAt,
		Updated_at: user.UpdatedAt,
		Email: user.Email,
		Token: accessToken,
		RefreshToken: refreshToken,
		ChirpyRed: user.IsChirpyRed,
	})
	return
}
