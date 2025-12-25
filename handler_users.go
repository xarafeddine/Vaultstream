package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	// Validate required fields
	params.Email = strings.TrimSpace(params.Email)
	params.FullName = strings.TrimSpace(params.FullName)

	if params.Password == "" || params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required", nil)
		return
	}

	if params.FullName == "" {
		respondWithError(w, http.StatusBadRequest, "Full name is required", nil)
		return
	}

	if len(params.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 6 characters", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.CreateUser(database.CreateUserParams{
		Email:    params.Email,
		Password: hashedPassword,
		FullName: params.FullName,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, user)
}
