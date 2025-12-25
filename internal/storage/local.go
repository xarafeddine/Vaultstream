package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type LocalStorage struct {
	baseDir   string
	baseURL   string // e.g. "http://localhost:8080/assets"
	secretKey string // used for signing presigned URLs
}

func NewLocalStorage(baseDir, baseURL, secretKey string) *LocalStorage {
	return &LocalStorage{
		baseDir:   baseDir,
		baseURL:   baseURL,
		secretKey: secretKey,
	}
}

// Save stores the file locally and returns a storage reference in "local,key" format
func (s *LocalStorage) Save(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	// key might contain subdirectories (e.g. "landscape/uuid.mp4")
	filePath := filepath.Join(s.baseDir, key)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", fmt.Errorf("couldn't create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("couldn't create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, data); err != nil {
		return "", fmt.Errorf("couldn't save file: %w", err)
	}

	// Return storage reference as "local,key"
	storageRef := fmt.Sprintf("local,%s", key)
	return storageRef, nil
}

// GeneratePresignedURL creates a signed URL that expires after the given duration
// This mimics S3's presigned URL behavior for local storage
func (s *LocalStorage) GeneratePresignedURL(storageRef string, expireTime time.Duration) (string, error) {
	// Parse storage reference "local,key"
	parts := strings.SplitN(storageRef, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid storage reference: %s", storageRef)
	}
	key := parts[1]

	// Calculate expiration timestamp
	expires := time.Now().Add(expireTime).Unix()

	// Create signature: HMAC-SHA256(key + expires, secretKey)
	message := fmt.Sprintf("%s:%d", key, expires)
	signature := s.sign(message)

	// Build presigned URL with query parameters
	presignedURL := fmt.Sprintf("%s/%s?expires=%d&signature=%s", s.baseURL, key, expires, signature)

	return presignedURL, nil
}

// sign creates an HMAC-SHA256 signature
func (s *LocalStorage) sign(message string) string {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyPresignedURL validates a presigned URL's signature and expiration
// This should be called by the assets handler to verify the request
func (s *LocalStorage) VerifyPresignedURL(key string, expiresStr, signature string) bool {
	// Parse expiration timestamp
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return false
	}

	// Check if URL has expired
	if time.Now().Unix() > expires {
		return false
	}

	// Verify signature
	message := fmt.Sprintf("%s:%d", key, expires)
	expectedSignature := s.sign(message)

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// DeleteFile removes a file from local storage
func (s *LocalStorage) DeleteFile(storageRef string) error {
	// Parse storage reference "local,key"
	parts := strings.SplitN(storageRef, ",", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid storage reference: %s", storageRef)
	}
	key := parts[1]

	filePath := filepath.Join(s.baseDir, key)
	return os.Remove(filePath)
}
