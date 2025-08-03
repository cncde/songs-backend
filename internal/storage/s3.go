package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// S3Service handles file operations with S3
type S3Service struct {
	client    *s3.Client
	bucket    string
	urlExpiry time.Duration
}

// NewS3Service creates a new S3 service instance
func NewS3Service(client *s3.Client, bucket string) *S3Service {
	return &S3Service{
		client:    client,
		bucket:    bucket,
		urlExpiry: 15 * time.Minute, // Default URL expiry
	}
}

// FileType represents allowed file types
type FileType string

const (
	FileTypeImage FileType = "image"
	FileTypeAudio FileType = "audio"
)

// UploadConfig holds configuration for file uploads
type UploadConfig struct {
	MaxSize         int64    // Maximum file size in bytes
	AllowedMIMETypes []string // Allowed MIME types
	PathPrefix      string   // S3 key prefix
}

// GetUploadConfigs returns upload configurations for different file types
func GetUploadConfigs() map[FileType]UploadConfig {
	return map[FileType]UploadConfig{
		FileTypeImage: {
			MaxSize: 10 * 1024 * 1024, // 10MB
			AllowedMIMETypes: []string{
				"image/jpeg",
				"image/png",
				"image/webp",
				"image/gif",
			},
			PathPrefix: "images",
		},
		FileTypeAudio: {
			MaxSize: 100 * 1024 * 1024, // 100MB
			AllowedMIMETypes: []string{
				"audio/mpeg",
				"audio/wav",
				"audio/mp4",
				"audio/aac",
				"audio/ogg",
				"audio/flac",
			},
			PathPrefix: "audio",
		},
	}
}

// PresignedUploadRequest represents a request for presigned upload URL
type PresignedUploadRequest struct {
	FileName    string   `json:"file_name" validate:"required"`
	ContentType string   `json:"content_type" validate:"required"`
	FileSize    int64    `json:"file_size" validate:"required,min=1"`
	FileType    FileType `json:"file_type" validate:"required,oneof=image audio"`
	UserID      string   `json:"-"` // Set from authentication context
}

// PresignedUploadResponse represents the response with presigned URL
type PresignedUploadResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresAt string `json:"expires_at"`
}

// GeneratePresignedUploadURL generates a presigned URL for file upload
func (s *S3Service) GeneratePresignedUploadURL(ctx context.Context, req PresignedUploadRequest) (*PresignedUploadResponse, error) {
	// Validate file type and configuration
	configs := GetUploadConfigs()
	config, exists := configs[req.FileType]
	if !exists {
		return nil, fmt.Errorf("unsupported file type: %s", req.FileType)
	}

	// Validate file size
	if req.FileSize > config.MaxSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", req.FileSize, config.MaxSize)
	}

	// Validate content type
	if !isAllowedMIMEType(req.ContentType, config.AllowedMIMETypes) {
		return nil, fmt.Errorf("content type %s is not allowed for file type %s", req.ContentType, req.FileType)
	}

	// Validate file extension matches content type
	if !isValidExtensionForContentType(req.FileName, req.ContentType) {
		return nil, fmt.Errorf("file extension does not match content type")
	}

	// Generate unique file key
	fileKey := s.generateFileKey(req.UserID, string(req.FileType), req.FileName)

	// Create presigned URL
	presignClient := s3.NewPresignClient(s.client)
	presignReq, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(fileKey),
		ContentType: aws.String(req.ContentType),
		Metadata: map[string]string{
			"uploaded-by":   req.UserID,
			"upload-time":   time.Now().UTC().Format(time.RFC3339),
			"original-name": req.FileName,
			"file-type":     string(req.FileType),
		},
		ServerSideEncryption: "AES256",
	}, func(opts *s3.PresignOptions) {
		opts.Expires = s.urlExpiry
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	response := &PresignedUploadResponse{
		UploadURL: presignReq.URL,
		FileKey:   fileKey,
		ExpiresAt: time.Now().Add(s.urlExpiry).Format(time.RFC3339),
	}

	return response, nil
}

// GeneratePresignedDownloadURL generates a presigned URL for file download
func (s *S3Service) GeneratePresignedDownloadURL(ctx context.Context, fileKey string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = s.urlExpiry
	}

	presignClient := s3.NewPresignClient(s.client)
	presignReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return presignReq.URL, nil
}

// UploadFile uploads a file directly to S3 (alternative to presigned URLs)
func (s *S3Service) UploadFile(ctx context.Context, fileKey string, content io.Reader, contentType string, metadata map[string]string) error {
	if metadata == nil {
		metadata = make(map[string]string)
	}

	// Add default metadata
	metadata["upload-time"] = time.Now().UTC().Format(time.RFC3339)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(s.bucket),
		Key:                  aws.String(fileKey),
		Body:                 content,
		ContentType:          aws.String(contentType),
		Metadata:             metadata,
		ServerSideEncryption: "AES256",
	})

	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// DeleteFile deletes a file from S3
func (s *S3Service) DeleteFile(ctx context.Context, fileKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	})

	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileMetadata retrieves file metadata from S3
func (s *S3Service) GetFileMetadata(ctx context.Context, fileKey string) (map[string]string, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return result.Metadata, nil
}

// FileExists checks if a file exists in S3
func (s *S3Service) FileExists(ctx context.Context, fileKey string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fileKey),
	})

	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}

	return true, nil
}

// ListUserFiles lists files uploaded by a specific user
func (s *S3Service) ListUserFiles(ctx context.Context, userID string, fileType FileType, maxKeys int32) ([]string, error) {
	prefix := fmt.Sprintf("%s/%s/", string(fileType), userID)

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(maxKeys),
	}

	result, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list user files: %w", err)
	}

	keys := make([]string, 0, len(result.Contents))
	for _, obj := range result.Contents {
		if obj.Key != nil {
			keys = append(keys, *obj.Key)
		}
	}

	return keys, nil
}

// generateFileKey generates a unique S3 key for the file
func (s *S3Service) generateFileKey(userID, fileType, fileName string) string {
	// Clean the filename
	cleanName := filepath.Base(fileName)
	ext := filepath.Ext(cleanName)
	nameWithoutExt := strings.TrimSuffix(cleanName, ext)
	
	// Generate unique identifier
	uniqueID := uuid.New().String()
	
	// Create timestamp prefix
	timestamp := time.Now().UTC().Format("2006/01/02")
	
	// Build key: filetype/userid/timestamp/uniqueid_filename.ext
	return fmt.Sprintf("%s/%s/%s/%s_%s%s", fileType, userID, timestamp, uniqueID, nameWithoutExt, ext)
}

// isAllowedMIMEType checks if the content type is allowed
func isAllowedMIMEType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// isValidExtensionForContentType validates file extension matches content type
func isValidExtensionForContentType(fileName, contentType string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	expectedMIME := mime.TypeByExtension(ext)
	
	// Handle common cases where mime package might not match exactly
	mimeMap := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".m4a":  "audio/mp4",
		".aac":  "audio/aac",
		".ogg":  "audio/ogg",
		".flac": "audio/flac",
	}
	
	if mappedType, exists := mimeMap[ext]; exists {
		return mappedType == contentType
	}
	
	return expectedMIME == contentType
}

// ExtractFileKeyFromURL extracts the S3 file key from a full S3 URL
func ExtractFileKeyFromURL(url string) string {
	// Handle both path-style and virtual-hosted-style URLs
	if strings.Contains(url, ".s3.") || strings.Contains(url, ".amazonaws.com/") {
		parts := strings.Split(url, "/")
		if len(parts) >= 4 {
			// Join everything after the domain
			return strings.Join(parts[3:], "/")
		}
	}
	return ""
}

// ValidateFileKey validates that a file key follows expected patterns
func ValidateFileKey(fileKey, userID string) bool {
	// Expected format: filetype/userid/timestamp/uniqueid_filename.ext
	parts := strings.Split(fileKey, "/")
	if len(parts) < 4 {
		return false
	}
	
	// Check if the user ID in the key matches the requesting user
	if parts[1] != userID {
		return false
	}
	
	// Validate file type
	fileType := parts[0]
	validTypes := []string{"images", "audio"}
	isValidType := false
	for _, validType := range validTypes {
		if fileType == validType {
			isValidType = true
			break
		}
	}
	
	return isValidType
}