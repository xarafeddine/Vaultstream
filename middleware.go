package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

// ============================================
// Types
// ============================================

// responseWrapper captures the status code for logging
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// ContextKey for storing values in request context
type ContextKey string

const UserIDKey ContextKey = "userID"

// ============================================
// Middleware Stack
// ============================================

type Middleware func(http.Handler) http.Handler

// MiddlewareStack chains multiple middleware together
func MiddlewareStack(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// ============================================
// Logger Middleware
// ============================================

// Logger logs all requests with method, path, status code, and duration
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Color-coded status for terminal
		statusColor := getStatusColor(rw.statusCode)
		methodColor := getMethodColor(r.Method)

		log.Printf("%s%-7s%s %s%d%s %s %v",
			methodColor, r.Method, colorReset,
			statusColor, rw.statusCode, colorReset,
			r.URL.Path,
			duration,
		)
	})
}

// Terminal colors for better readability
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func getStatusColor(status int) string {
	switch {
	case status >= 500:
		return colorRed
	case status >= 400:
		return colorYellow
	case status >= 300:
		return colorCyan
	case status >= 200:
		return colorGreen
	default:
		return colorWhite
	}
}

func getMethodColor(method string) string {
	switch method {
	case "GET":
		return colorBlue
	case "POST":
		return colorGreen
	case "PUT", "PATCH":
		return colorYellow
	case "DELETE":
		return colorRed
	default:
		return colorWhite
	}
}

// ============================================
// Auth Middleware
// ============================================

// AuthMiddleware validates JWT and adds user ID to context
func (cfg *apiConfig) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Missing or invalid authorization header", err)
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token", err)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// AuthHandler wraps a handler function that requires authentication
func (cfg *apiConfig) AuthHandler(handler http.HandlerFunc) http.Handler {
	return cfg.AuthMiddleware(handler)
}

// ============================================
// CORS Middleware (optional)
// ============================================

// CORS adds Cross-Origin Resource Sharing headers
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================
// Recovery Middleware
// ============================================

// Recovery catches panics and returns a 500 error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("%s[PANIC]%s %v", colorRed, colorReset, err)
				respondWithError(w, http.StatusInternalServerError, "Internal server error", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
