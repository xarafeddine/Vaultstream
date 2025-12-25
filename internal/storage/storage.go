package storage

import (
	"context"
	"io"
	"time"
)

// FileStorage defines an interface for saving files to a storage backend
type FileStorage interface {
	// Save stores the file and returns a storage reference (e.g., "bucket,key" for S3 or "local,key" for local)
	// This reference is stored in the database and used to generate presigned URLs later
	Save(ctx context.Context, key string, data io.Reader, contentType string) (storageRef string, err error)

	// GeneratePresignedURL creates a temporary URL for accessing the file
	// For S3: uses AWS SDK presigning
	// For Local: creates a signed token-based URL
	GeneratePresignedURL(storageRef string, expireTime time.Duration) (string, error)
}
