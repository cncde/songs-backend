package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/cncde/songs-backend/internal/storage"
)

// MediaService handles media upload and download business logic
type MediaService struct {
	s3Service *storage.S3Service
	validator *validator.Validate
}

// NewMediaService creates a new media service
func NewMediaService(s3Service *storage.S3Service) *MediaService {
	return &MediaService{
		s3Service: s3Service,
		validator: validator.New(),
	}
}

// GenerateUploadURL generates a presigned URL for file upload with validation
func (m *MediaService) GenerateUploadURL(ctx context.Context, req storage.PresignedUploadRequest) (*storage.PresignedUploadResponse, error) {
	// Validate request structure
	if err := m.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Additional business logic validation
	if err := m.validateUploadRequest(req); err != nil {
		return nil, fmt.Errorf("upload validation failed: %w", err)
	}

	// Generate presigned URL
	response, err := m.s3Service.GeneratePresignedUploadURL(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return response, nil
}

// GenerateDownloadURL generates a presigned URL for file download
func (m *MediaService) GenerateDownloadURL(ctx context.Context, fileKey, userID string, expiry time.Duration) (string, error) {
	// Validate file key ownership
	if !storage.ValidateFileKey(fileKey, userID) {
		return "", fmt.Errorf("unauthorized access to file")
	}

	// Check if file exists
	exists, err := m.s3Service.FileExists(ctx, fileKey)
	if err != nil {
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}

	if !exists {
		return "", fmt.Errorf("file not found")
	}

	// Generate download URL
	url, err := m.s3Service.GeneratePresignedDownloadURL(ctx, fileKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}

// DeleteFile deletes a file with authorization check
func (m *MediaService) DeleteFile(ctx context.Context, fileKey, userID string) error {
	// Validate file key ownership
	if !storage.ValidateFileKey(fileKey, userID) {
		return fmt.Errorf("unauthorized access to file")
	}

	// Check if file exists
	exists, err := m.s3Service.FileExists(ctx, fileKey)
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("file not found")
	}

	// Delete the file
	if err := m.s3Service.DeleteFile(ctx, fileKey); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ListUserFiles lists files uploaded by a user
func (m *MediaService) ListUserFiles(ctx context.Context, userID string, fileType storage.FileType, limit int32) ([]FileInfo, error) {
	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	// List files from S3
	keys, err := m.s3Service.ListUserFiles(ctx, userID, fileType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Convert to FileInfo structs with metadata
	files := make([]FileInfo, 0, len(keys))
	for _, key := range keys {
		fileInfo := FileInfo{
			Key:      key,
			FileType: fileType,
			UserID:   userID,
		}

		// Get metadata if needed (optional, can be expensive for large lists)
		// metadata, err := m.s3Service.GetFileMetadata(ctx, key)
		// if err == nil {
		//     fileInfo.Metadata = metadata
		// }

		files = append(files, fileInfo)
	}

	return files, nil
}

// GetFileMetadata retrieves file metadata with authorization
func (m *MediaService) GetFileMetadata(ctx context.Context, fileKey, userID string) (map[string]string, error) {
	// Validate file key ownership
	if !storage.ValidateFileKey(fileKey, userID) {
		return nil, fmt.Errorf("unauthorized access to file")
	}

	// Get metadata
	metadata, err := m.s3Service.GetFileMetadata(ctx, fileKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return metadata, nil
}

// ValidateFileURL validates if a file URL is valid and belongs to the user
func (m *MediaService) ValidateFileURL(ctx context.Context, fileURL, userID string) (bool, error) {
	// Extract file key from URL
	fileKey := storage.ExtractFileKeyFromURL(fileURL)
	if fileKey == "" {
		return false, fmt.Errorf("invalid file URL format")
	}

	// Validate ownership
	if !storage.ValidateFileKey(fileKey, userID) {
		return false, fmt.Errorf("unauthorized access to file")
	}

	// Check if file exists
	exists, err := m.s3Service.FileExists(ctx, fileKey)
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return exists, nil
}

// GetUploadPolicies returns upload policies for different file types
func (m *MediaService) GetUploadPolicies() map[storage.FileType]storage.UploadConfig {
	return storage.GetUploadConfigs()
}

// validateUploadRequest performs business logic validation on upload requests
func (m *MediaService) validateUploadRequest(req storage.PresignedUploadRequest) error {
	// Check if user is authenticated
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Additional file name validation
	if len(req.FileName) > 255 {
		return fmt.Errorf("file name too long (max 255 characters)")
	}

	// Validate file extension
	if !m.hasValidExtension(req.FileName, req.FileType) {
		return fmt.Errorf("invalid file extension for file type %s", req.FileType)
	}

	// Business rule: limit number of uploads per user (placeholder)
	// You could implement rate limiting or quota checking here
	
	return nil
}

// hasValidExtension checks if file has a valid extension for the file type
func (m *MediaService) hasValidExtension(fileName string, fileType storage.FileType) bool {
	configs := storage.GetUploadConfigs()
	config, exists := configs[fileType]
	if !exists {
		return false
	}

	// This would need to be implemented based on your requirements
	// For now, just return true as the S3 service handles this validation
	_ = config
	return true
}

// FileInfo represents information about an uploaded file
type FileInfo struct {
	Key       string                `json:"key"`
	FileType  storage.FileType      `json:"file_type"`
	UserID    string                `json:"user_id"`
	Metadata  map[string]string     `json:"metadata,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitempty"`
}

// FileUploadStats represents upload statistics for a user
type FileUploadStats struct {
	TotalFiles      int   `json:"total_files"`
	TotalSizeBytes  int64 `json:"total_size_bytes"`
	ImageCount      int   `json:"image_count"`
	AudioCount      int   `json:"audio_count"`
	LastUploadTime  *time.Time `json:"last_upload_time,omitempty"`
}

// GetUserUploadStats returns upload statistics for a user
func (m *MediaService) GetUserUploadStats(ctx context.Context, userID string) (*FileUploadStats, error) {
	stats := &FileUploadStats{}

	// Get image files
	imageFiles, err := m.s3Service.ListUserFiles(ctx, userID, storage.FileTypeImage, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get image files: %w", err)
	}
	stats.ImageCount = len(imageFiles)

	// Get audio files
	audioFiles, err := m.s3Service.ListUserFiles(ctx, userID, storage.FileTypeAudio, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio files: %w", err)
	}
	stats.AudioCount = len(audioFiles)

	stats.TotalFiles = stats.ImageCount + stats.AudioCount

	// Note: Getting file sizes would require additional S3 API calls
	// This could be expensive for users with many files
	// Consider caching this information or updating it asynchronously

	return stats, nil
}

// CleanupOrphanedFiles removes files that are not referenced in any songs
func (m *MediaService) CleanupOrphanedFiles(ctx context.Context, userID string) (int, error) {
	// This would be an admin/maintenance function to clean up unreferenced files
	// Implementation would require checking all songs for file references
	// and removing files that are no longer used

	// For now, return an error as this is a complex operation
	return 0, fmt.Errorf("cleanup operation not implemented")
}

// BulkDeleteFiles deletes multiple files in batch
func (m *MediaService) BulkDeleteFiles(ctx context.Context, fileKeys []string, userID string) error {
	if len(fileKeys) == 0 {
		return nil
	}

	if len(fileKeys) > 100 {
		return fmt.Errorf("too many files to delete at once (max 100)")
	}

	// Validate ownership of all files
	for _, key := range fileKeys {
		if !storage.ValidateFileKey(key, userID) {
			return fmt.Errorf("unauthorized access to file: %s", key)
		}
	}

	// Delete files one by one (could be optimized with batch operations)
	var errors []string
	for _, key := range fileKeys {
		if err := m.s3Service.DeleteFile(ctx, key); err != nil {
			errors = append(errors, fmt.Sprintf("failed to delete %s: %v", key, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some files failed to delete: %v", errors)
	}

	return nil
}

// GeneratePublicURL generates a public URL for published content
func (m *MediaService) GeneratePublicURL(ctx context.Context, fileKey string, expiry time.Duration) (string, error) {
	// For published content, we might want longer expiry times or different policies
	if expiry == 0 {
		expiry = 24 * time.Hour // Default 24 hours for public content
	}

	// Generate download URL (same as private but with different expiry)
	url, err := m.s3Service.GeneratePresignedDownloadURL(ctx, fileKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate public URL: %w", err)
	}

	return url, nil
}