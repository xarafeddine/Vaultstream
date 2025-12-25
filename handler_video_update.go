package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// handlerVideoMetaUpdate updates video metadata (title, description)
func (cfg *apiConfig) handlerVideoMetaUpdate(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse video ID from URL
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID format", err)
		return
	}

	// Get existing video
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Video not found", err)
		return
	}

	// Check ownership
	if video.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You don't have permission to edit this video", nil)
		return
	}

	// Parse update parameters
	type parameters struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Update fields if provided
	if params.Title != nil {
		if *params.Title == "" {
			respondWithError(w, http.StatusBadRequest, "Title cannot be empty", nil)
			return
		}
		video.Title = *params.Title
	}
	if params.Description != nil {
		video.Description = *params.Description
	}

	// Save updates
	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update video", err)
		return
	}

	// Return updated video with presigned URLs
	signedVideo, err := cfg.dbVideoToSignedVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate presigned URLs", err)
		return
	}

	respondWithJSON(w, http.StatusOK, signedVideo)
}
