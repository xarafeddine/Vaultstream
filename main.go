package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/storage"

	"github.com/joho/godotenv"
)

type apiConfig struct {
	db           database.Client
	jwtSecret    string
	platform     string
	filepathRoot string
	assetsRoot   string
	port         string
	storage      storage.FileStorage
}

func main() {
	godotenv.Load(".env")

	pathToDB := os.Getenv("DB_PATH")
	if pathToDB == "" {
		log.Fatal("DB_URL must be set")
	}

	db, err := database.NewClient(pathToDB)
	if err != nil {
		log.Fatalf("Couldn't connect to database: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM environment variable is not set")
	}

	filepathRoot := os.Getenv("FILEPATH_ROOT")
	if filepathRoot == "" {
		log.Fatal("FILEPATH_ROOT environment variable is not set")
	}

	assetsRoot := os.Getenv("ASSETS_ROOT")
	if assetsRoot == "" {
		log.Fatal("ASSETS_ROOT environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	var storageBackend storage.FileStorage

	// Choose storage backend based on PLATFORM or a dedicated STORAGE_TYPE env var
	// For this exercise, let's use S3 by default if bucket is set, otherwise local.
	s3Bucket := os.Getenv("S3_BUCKET")
	s3Region := os.Getenv("S3_REGION")

	if s3Bucket != "" && s3Region != "" {
		aws_cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(s3Region))
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}
		s3Client := s3.NewFromConfig(aws_cfg)
		storageBackend = storage.NewS3Storage(s3Client, s3Bucket, s3Region)
		log.Println("Using S3 storage")
	} else {
		// Use jwtSecret for signing local presigned URLs (mimics S3 behavior)
		storageBackend = storage.NewLocalStorage(assetsRoot, "http://localhost:"+port+"/assets", jwtSecret)
		log.Println("Using local storage")
	}

	cfg := apiConfig{
		db:           db,
		jwtSecret:    jwtSecret,
		platform:     platform,
		filepathRoot: filepathRoot,
		assetsRoot:   assetsRoot,
		port:         port,
		storage:      storageBackend,
	}

	err = cfg.ensureAssetsDir()
	if err != nil {
		log.Fatalf("Couldn't create assets directory: %v", err)
	}

	mux := http.NewServeMux()

	// Register all routes
	cfg.RegisterRoutes(mux)

	// Apply global middleware stack
	handler := MiddlewareStack(
		Recovery,
		Logger,
		CORS,
	)(mux)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	log.Printf("🚀 Vaultstream server running on http://localhost:%s/app/\n", port)
	log.Fatal(srv.ListenAndServe())
}
