package main

import (
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse form", err)
		return
	}

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse media type", err)
		return
	}
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Unsupported media type", nil)
		return
	}

	// imageData, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "Couldn't read the file", err)
	// 	return
	// }

	video, err := cfg.db.GetVideo(videoID)
	if err != nil || video.ID == uuid.Nil {
		respondWithError(w, http.StatusNotFound, "Video not found", nil)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized", nil)
		return
	}

	fileExtension := mediaType[len("image/"):]
	thumbnailKey := fmt.Sprintf("thumbnails/%s.%s", videoID.String(), fileExtension)

	// Save thumbnail using abstract storage
	thumbnailURL, err := cfg.storage.Save(r.Context(), thumbnailKey, file, mediaType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save thumbnail", err)
		return
	}

	video.ThumbnailURL = &thumbnailURL
	video.UpdatedAt = time.Now()

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video metadata in database", err)
		return
	}
	signedVideo, err := cfg.dbVideoToSignedVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate presigned URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, signedVideo)
}
