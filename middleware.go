package main

import (
	"log"
	"net/http"
	"time"
)

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
}

type Middleware func(http.Handler) http.Handler

func MiddlewareStack(arr ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := 0; i < len(arr); i++ {
			middleware := arr[len(arr)-1-i]
			next = middleware(next)
		}
		return next
	}
}

var ApplyMiddleware = MiddlewareStack(Logger)

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Println("Implenet auth")
		next(w, r)
	}
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(rw, r)
			log.Printf("| %d | %s | %s | %v", rw.statusCode, r.Method, r.URL.Path, time.Since(start))
		},
	)
}
