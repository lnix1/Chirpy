package main

import (
	"time"
	"net/http"
	"encoding/json"

	"github.com/lnix1/Chirpy/internal/auth"
	"github.com/lnix1/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password 	string	`json:"password"`
		Email		string	`json:"email"`
	}
	type returnVals struct {
		Id		uuid.UUID 	`json:"id"`
		Created_at 	time.Time	`json:"created_at"`
		Updated_at 	time.Time	`json:"updated_at"`
		Email 		string		`json:"email"`
		ChirpyRed	bool		`json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	if params.Email == "" || params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password or email missing", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating password hash", err)
		return
	}
	
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{params.Email, hashedPassword})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't create new user", err)
	}
	
	respondWithJSON(w, http.StatusCreated, returnVals{
		Id: user.ID,
		Created_at: user.CreatedAt,
		Updated_at: user.UpdatedAt,
		Email: user.Email,
		ChirpyRed: user.IsChirpyRed,
	})
	return
}

func (cfg *apiConfig) handlerUpdateEmailPassword(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication token missing", err)
	}
	
	bearerUserID, err := auth.ValidateJWT(bearer, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid authentication token", err)
	}

	type parameters struct {
		Password 	string	`json:"password"`
		Email		string	`json:"email"`
	}
	type returnVals struct {
		Id		uuid.UUID 	`json:"id"`
		Created_at 	time.Time	`json:"created_at"`
		Updated_at 	time.Time	`json:"updated_at"`
		Email 		string		`json:"email"`
		ChirpyRed	bool		`json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	if params.Email == "" || params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "New password or email missing", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error with password creation", err)
		return
	}
	updateArgs := database.UpdateEmailUserParams{Email: params.Email, HashedPassword: hashedPassword, ID: bearerUserID}
	user, err := cfg.db.UpdateEmailUser(r.Context(), updateArgs)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to udpate email and password", err)
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Id: user.ID,
		Created_at: user.CreatedAt,
		Updated_at: user.UpdatedAt,
		Email: user.Email,
		ChirpyRed: user.IsChirpyRed,
	})
	return
}
