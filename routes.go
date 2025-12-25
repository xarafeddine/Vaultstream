package main

import (
	"net/http"
)

// RegisterRoutes sets up all API routes with appropriate middleware
func (cfg *apiConfig) RegisterRoutes(mux *http.ServeMux) {
	// ============================================
	// Static Files
	// ============================================
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(cfg.filepathRoot)))
	mux.Handle("/app/", appHandler)

	// Assets with presigned URL verification (supports Range requests for video streaming)
	mux.HandleFunc("/assets/", cfg.handlerServeAssets)

	// ============================================
	// Public Routes (No Auth Required)
	// ============================================

	// Authentication
	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)

	// User Registration
	mux.HandleFunc("POST /api/users", cfg.handlerUsersCreate)

	// Password Reset
	mux.HandleFunc("POST /api/forgot-password", cfg.handlerForgotPassword)
	mux.HandleFunc("POST /api/reset-password", cfg.handlerResetPassword)

	// ============================================
	// Protected Routes (Auth Required)
	// ============================================

	// Videos - CRUD
	mux.Handle("POST /api/videos", cfg.AuthHandler(cfg.handlerVideoMetaCreate))
	mux.Handle("GET /api/videos", cfg.AuthHandler(cfg.handlerVideosRetrieve))
	mux.Handle("GET /api/videos/{videoID}", cfg.AuthHandler(cfg.handlerVideoGet))
	mux.Handle("PUT /api/videos/{videoID}", cfg.AuthHandler(cfg.handlerVideoMetaUpdate))
	mux.Handle("DELETE /api/videos/{videoID}", cfg.AuthHandler(cfg.handlerVideoMetaDelete))

	// Video File Uploads
	mux.Handle("POST /api/thumbnail_upload/{videoID}", cfg.AuthHandler(cfg.handlerUploadThumbnail))
	mux.Handle("POST /api/video_upload/{videoID}", cfg.AuthHandler(cfg.handlerUploadVideo))

	// Selective Deletion
	mux.Handle("DELETE /api/videos/{videoID}/thumbnail", cfg.AuthHandler(cfg.handlerDeleteThumbnail))
	mux.Handle("DELETE /api/videos/{videoID}/video-file", cfg.AuthHandler(cfg.handlerDeleteVideoFile))

	// ============================================
	// Admin Routes
	// ============================================
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)
}
