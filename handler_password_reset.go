package main

import (
	"encoding/json"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

// handlerForgotPassword initiates the password reset process
// In a real app, this would send an email with the reset link
func (cfg *apiConfig) handlerForgotPassword(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	type response struct {
		Message string `json:"message"`
		Token   string `json:"token,omitempty"` // Only for demo; in production, send via email
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if params.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Email is required", nil)
		return
	}

	// Find user by email
	user, err := cfg.db.GetUserByEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	// Always return success to prevent email enumeration attacks
	if user.ID.String() == "00000000-0000-0000-0000-000000000000" {
		respondWithJSON(w, http.StatusOK, response{
			Message: "If an account with that email exists, a password reset link has been sent.",
		})
		return
	}

	// Create password reset token
	resetToken, err := cfg.db.CreatePasswordResetToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create reset token", err)
		return
	}

	// In a real app, you would send this token via email
	// For demo purposes, we return it in the response
	respondWithJSON(w, http.StatusOK, response{
		Message: "If an account with that email exists, a password reset link has been sent.",
		Token:   resetToken.Token, // Remove this in production!
	})
}

// handlerResetPassword completes the password reset process
func (cfg *apiConfig) handlerResetPassword(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	type response struct {
		Message string `json:"message"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if params.Token == "" || params.NewPassword == "" {
		respondWithError(w, http.StatusBadRequest, "Token and new password are required", nil)
		return
	}

	if len(params.NewPassword) < 6 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 6 characters", nil)
		return
	}

	// Validate the reset token
	resetToken, err := cfg.db.GetPasswordResetToken(params.Token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	if resetToken == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid or expired reset token", nil)
		return
	}

	// Hash the new password
	hashedPassword, err := auth.HashPassword(params.NewPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// Update the user's password
	if err := cfg.db.UpdateUserPassword(resetToken.UserID, hashedPassword); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update password", err)
		return
	}

	// Mark the token as used
	if err := cfg.db.MarkPasswordResetTokenUsed(params.Token); err != nil {
		// Log but don't fail - password was already updated
	}

	respondWithJSON(w, http.StatusOK, response{
		Message: "Password has been reset successfully. You can now log in with your new password.",
	})
}
