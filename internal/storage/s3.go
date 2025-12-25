package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3Storage(client *s3.Client, bucket, region string) *S3Storage {
	return &S3Storage{
		client: client,
		bucket: bucket,
		region: region,
	}
}

// Save uploads the file to S3 and returns a storage reference in "bucket,key" format
func (s *S3Storage) Save(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if seeker, ok := data.(io.ReadSeeker); ok {
			_, err := seeker.Seek(0, io.SeekStart)
			if err != nil {
				return "", fmt.Errorf("couldn't seek to beginning: %w", err)
			}
		}

		_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      &s.bucket,
			Key:         &key,
			Body:        data,
			ContentType: &contentType,
		})

		if err == nil {
			// Return storage reference as "bucket,key"
			storageRef := fmt.Sprintf("%s,%s", s.bucket, key)
			return storageRef, nil
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(time.Second * time.Duration(attempt))
		}
	}

	return "", fmt.Errorf("S3 upload failed after %d attempts: %w", maxRetries, lastErr)
}

// GeneratePresignedURL creates a presigned URL for the given storage reference
func (s *S3Storage) GeneratePresignedURL(storageRef string, expireTime time.Duration) (string, error) {
	// Parse storage reference "bucket,key"
	parts := strings.SplitN(storageRef, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid storage reference: %s", storageRef)
	}
	bucket := parts[0]
	key := parts[1]

	// Create presign client
	presignClient := s3.NewPresignClient(s.client)

	// Generate presigned URL
	presignedReq, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expireTime))

	if err != nil {
		return "", fmt.Errorf("couldn't generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
}
