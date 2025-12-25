package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	// 	Set an upload limit of 1 GB (1 << 30 bytes) using http.MaxBytesReader.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)
	// Extract the videoID from the URL path parameters and parse it as a UUID
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

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get video", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You are not the owner of this video", nil)
		return
	}

	fmt.Println("uploading video for video", videoID, "by user", userID)

	// "video" should match the HTML form input name
	file, header, err := r.FormFile("video")
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
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Unsupported media type", nil)
		return
	}

	//     Use os.CreateTemp to create a temporary file. I passed in an empty string for the directory
	// to use the system default, and the name "vaultstream-upload.mp4" (but you can use whatever you want)
	tempFile, err := os.CreateTemp("", "vaultstream-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create temp file", err)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	//     io.Copy the contents over from the wire to the temp file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't copy file", err)
		return
	}

	// 	Reset the tempFile's file pointer to the beginning
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't seek to beginning of file", err)
		return
	}

	// Get aspect ratio and determine S3 key prefix
	aspectRatio, err := getVideoAspectRatio(tempFile.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get video aspect ratio", err)
		return
	}

	var prefix string
	switch aspectRatio {
	case "16:9":
		prefix = "landscape/"
	case "9:16":
		prefix = "portrait/"
	default:
		prefix = "other/"
	}

	s3Key := fmt.Sprintf("%s%s.mp4", prefix, videoID.String())

	// Validate file size before processing
	fileInfo, err := tempFile.Stat()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get file info", err)
		return
	}
	if fileInfo.Size() == 0 {
		respondWithError(w, http.StatusBadRequest, "Uploaded file is empty", nil)
		return
	}

	// Close original temp file before processing
	tempFile.Close()

	// Process video for fast start - creates a new processed file
	processedFilePath, err := processVideoForFastStart(tempFile.Name())
	if err != nil {
		os.Remove(tempFile.Name()) // Clean up original
		respondWithError(w, http.StatusInternalServerError, "Couldn't process video for fast start", err)
		return
	}
	defer os.Remove(tempFile.Name())   // Clean up original temp file
	defer os.Remove(processedFilePath) // Clean up processed file after upload

	// Open processed file for upload
	processedFile, err := os.Open(processedFilePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't open processed video file", err)
		return
	}
	defer processedFile.Close()

	// Upload processed video using our abstract storage interface
	storageRef, err := cfg.storage.Save(r.Context(), s3Key, processedFile, mediaType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't upload video", err)
		return
	}

	video.UpdatedAt = time.Now()
	video.VideoURL = &storageRef
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

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL != nil {
		presignedURL, err := cfg.storage.GeneratePresignedURL(*video.VideoURL, 15*time.Minute)
		if err != nil {
			return database.Video{}, err
		}
		video.VideoURL = &presignedURL
	}

	if video.ThumbnailURL != nil {
		presignedURL, err := cfg.storage.GeneratePresignedURL(*video.ThumbnailURL, 15*time.Minute)
		if err != nil {
			return database.Video{}, err
		}
		video.ThumbnailURL = &presignedURL
	}

	return video, nil
}
