package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/pedroomedicina/chirpy/internal/auth"
	"github.com/pedroomedicina/chirpy/internal/database"
	"log"
	"net/http"
	"time"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil || reqBody.Email == "" || reqBody.Password == "" {
		respondWithError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	hashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	dbUser, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          reqBody.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		log.Printf("Error creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	apiUser := User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusCreated, apiUser)
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil || reqBody.Email == "" || reqBody.Password == "" {
		respondWithError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		return
	}

	dbUser, err := cfg.dbQueries.GetUserByEmail(r.Context(), reqBody.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		} else {
			respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		return
	}

	err = auth.CheckPasswordHash(reqBody.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	accessTokenExpiresIn := time.Hour
	token, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, accessTokenExpiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	refreshTokenExpiresAt := time.Now().Add(60 * 24 * time.Hour)
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: refreshTokenExpiresAt,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save refresh token")
		return
	}

	apiUser := User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  dbUser.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, apiUser)
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing authorization token")
		return
	}

	dbUser, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to validate refresh token")
		}
		return
	}

	accessToken, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": accessToken,
	})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or missing authorization token")
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusUnauthorized, "Invalid or already revoked token")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to revoke token")
		}
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization token")
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil || reqBody.Email == "" || reqBody.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	hashedPassword, err := auth.HashPassword(reqBody.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid or expired password")
		return
	}

	dbUser, err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          reqBody.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user")
		return
	}

	apiUser := User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, apiUser)
}
