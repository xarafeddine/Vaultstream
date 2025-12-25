package main

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/storage"
)

// handlerServeAssets serves files from the assets directory with support for
// HTTP Range requests, enabling video streaming (like S3 does).
func (cfg *apiConfig) handlerServeAssets(w http.ResponseWriter, r *http.Request) {
	// Get the file path from URL, strip the /assets/ prefix
	urlPath := r.URL.Path
	filePath := strings.TrimPrefix(urlPath, "/assets/")

	// Security: prevent directory traversal attacks
	if strings.Contains(filePath, "..") {
		respondWithError(w, http.StatusBadRequest, "Invalid path", nil)
		return
	}

	// Verify local presigned URL signature
	if localStorage, ok := cfg.storage.(*storage.LocalStorage); ok {
		expires := r.URL.Query().Get("expires")
		signature := r.URL.Query().Get("signature")
		if !localStorage.VerifyPresignedURL(filePath, expires, signature) {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired signature", nil)
			return
		}
	}

	// Build full file path
	fullPath := filepath.Join(cfg.assetsRoot, filePath)

	// Check if file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			respondWithError(w, http.StatusNotFound, "File not found", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error accessing file", err)
		return
	}

	// Don't serve directories
	if fileInfo.IsDir() {
		respondWithError(w, http.StatusBadRequest, "Not a file", nil)
		return
	}

	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error opening file", err)
		return
	}
	defer file.Close()

	// Detect content type from file extension
	ext := filepath.Ext(fullPath)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set content type header
	w.Header().Set("Content-Type", contentType)

	// Disable caching for development
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	// http.ServeContent handles:
	// - Range requests (for video streaming/seeking)
	// - If-Modified-Since headers
	// - Content-Length
	// - 206 Partial Content responses
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}
