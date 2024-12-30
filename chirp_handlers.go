package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/pedroomedicina/chirpy/internal/auth"
	"github.com/pedroomedicina/chirpy/internal/database"
	"net/http"
	"strings"
)

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

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

	var chirp Chirp
	err = json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	err = cfg.validateChirp(chirp)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}

	cleanedChirpBody := cleanProfanity(chirp.Body)

	dbChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedChirpBody,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	apiChirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, apiChirp)
}

func validateSortDirection(sort string) (string, error) {
	sort = strings.ToUpper(sort)

	if sort == "" {
		return "ASC", nil
	}

	if sort == "ASC" || sort == "DESC" {
		return sort, nil
	}

	return "", fmt.Errorf("invalid sort direction: %s", sort)
}

func scanChirps(rows *sql.Rows) ([]Chirp, error) {
	var chirps []Chirp
	for rows.Next() {
		var chirp Chirp
		err := rows.Scan(&chirp.ID, &chirp.CreatedAt, &chirp.UpdatedAt, &chirp.Body, &chirp.UserID)
		if err != nil {
			return nil, err
		}
		chirps = append(chirps, chirp)
	}
	return chirps, rows.Err()
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	sortQueryParam := r.URL.Query().Get("sort")
	validatedSort, err := validateSortDirection(sortQueryParam)
	if err != nil {
		fmt.Printf("error when validating sort: %s, validated sort:%s\n", err.Error(), validatedSort)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	authorId := r.URL.Query().Get("author_id")
	if authorId != "" {
		authorUuid, err := uuid.Parse(authorId)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		query := fmt.Sprintf("SELECT * FROM chirps WHERE user_id = '%s' ORDER BY created_at %s", authorUuid, sortQueryParam)
		rows, err := cfg.db.QueryContext(r.Context(), query)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {

			}
		}(rows)

		chirps, err := scanChirps(rows)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, chirps)
		return
	}

	query := fmt.Sprintf("SELECT * FROM chirps ORDER BY created_at %s", sortQueryParam)
	rows, err := cfg.db.QueryContext(r.Context(), query)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	chirps, err := scanChirps(rows)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var apiChirps []Chirp
	for _, chirp := range chirps {
		apiChirps = append(apiChirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, apiChirps)
}

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("id")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID")
		return
	}

	dbChirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization token")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	chirpIDString := r.PathValue("id")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp ID")
		return
	}

	dbChirp, err := cfg.dbQueries.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
			return
		}

		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "you are not authorized to delete this chirp")
		return
	}

	err = cfg.dbQueries.DeleteChirp(r.Context(), dbChirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
