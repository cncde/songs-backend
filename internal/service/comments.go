package service

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/cncde/songs-backend/internal/models"
	"github.com/cncde/songs-backend/internal/repository"
)

// CommentService handles comment business logic
type CommentService struct {
	commentRepo *repository.CommentRepository
	songRepo    *repository.SongRepository
	validator   *validator.Validate
}

// NewCommentService creates a new comment service
func NewCommentService(commentRepo *repository.CommentRepository, songRepo *repository.SongRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		songRepo:    songRepo,
		validator:   validator.New(),
	}
}

// CreateComment creates a new comment with business logic validation
func (s *CommentService) CreateComment(ctx context.Context, songID, userID string, req models.CreateCommentRequest) (*models.Comment, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify the song exists and is accessible
	song, err := s.songRepo.GetByID(ctx, songID)
	if err != nil {
		return nil, fmt.Errorf("song not found: %w", err)
	}

	// Check if song accepts comments (business rule: only published songs accept comments)
	if !song.IsPublished() {
		return nil, fmt.Errorf("comments are only allowed on published songs")
	}

	// Validate that the audio URL exists in the song's audio URLs
	if !s.isValidAudioURL(req.AudioURL, song.AudioURLs) {
		return nil, fmt.Errorf("audio URL not found in song")
	}

	// Create comment
	comment := models.NewComment(req, songID, userID)

	// Additional business validation
	if comment.IsEmpty() {
		return nil, fmt.Errorf("comment content cannot be empty")
	}

	if comment.IsTooLong() {
		return nil, fmt.Errorf("comment content exceeds maximum length")
	}

	// Save to repository
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		if err == repository.ErrItemAlreadyExists {
			return nil, fmt.Errorf("comment with this ID already exists")
		}
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

// GetComment retrieves a comment by ID
func (s *CommentService) GetComment(ctx context.Context, commentID string) (*models.Comment, error) {
	if commentID == "" {
		return nil, fmt.Errorf("comment ID is required")
	}

	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return comment, nil
}

// UpdateComment updates an existing comment (only by owner)
func (s *CommentService) UpdateComment(ctx context.Context, commentID, userID string, req models.UpdateCommentRequest) (*models.Comment, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing comment and verify ownership
	comment, err := s.commentRepo.GetByUserIDAndID(ctx, userID, commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	// Business rule: comments can only be edited within a certain time frame
	if !s.canEditComment(comment) {
		return nil, fmt.Errorf("comment can no longer be edited")
	}

	// Apply updates
	comment.Update(req)

	// Additional validation after update
	if comment.IsEmpty() {
		return nil, fmt.Errorf("comment content cannot be empty")
	}

	if comment.IsTooLong() {
		return nil, fmt.Errorf("comment content exceeds maximum length")
	}

	// Save to repository
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	return comment, nil
}

// DeleteComment deletes a comment (only by owner or song owner)
func (s *CommentService) DeleteComment(ctx context.Context, commentID, userID string) error {
	// Get the comment to check ownership
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// Check if user can delete the comment (either they own it or they own the song)
	canDelete := false
	
	// User owns the comment
	if comment.UserID == userID {
		canDelete = true
	} else {
		// Check if user owns the song
		song, err := s.songRepo.GetByID(ctx, comment.SongID)
		if err != nil {
			return fmt.Errorf("failed to get song: %w", err)
		}
		if song.UserID == userID {
			canDelete = true
		}
	}

	if !canDelete {
		return fmt.Errorf("unauthorized to delete this comment")
	}

	// Delete the comment
	if err := s.commentRepo.Delete(ctx, commentID, comment.UserID); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

// ListSongComments lists all comments for a song
func (s *CommentService) ListSongComments(ctx context.Context, songID string, limit int32, cursor string) (*models.CommentsListResponse, error) {
	// Verify song exists
	_, err := s.songRepo.GetByID(ctx, songID)
	if err != nil {
		return nil, fmt.Errorf("song not found: %w", err)
	}

	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}

	comments, nextCursor, err := s.commentRepo.ListBySongID(ctx, songID, limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}

	return &models.CommentsListResponse{
		Comments:   comments,
		NextCursor: nextCursor,
		Count:      len(comments),
	}, nil
}

// ListCommentsByAudioURL lists comments for a specific audio file
func (s *CommentService) ListCommentsByAudioURL(ctx context.Context, songID, audioURL string, limit int32, cursor string) (*models.CommentsListResponse, error) {
	// Verify song exists and audio URL is valid
	song, err := s.songRepo.GetByID(ctx, songID)
	if err != nil {
		return nil, fmt.Errorf("song not found: %w", err)
	}

	if !s.isValidAudioURL(audioURL, song.AudioURLs) {
		return nil, fmt.Errorf("audio URL not found in song")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	comments, nextCursor, err := s.commentRepo.ListBySongIDAndAudioURL(ctx, songID, audioURL, limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments by audio URL: %w", err)
	}

	return &models.CommentsListResponse{
		Comments:   comments,
		NextCursor: nextCursor,
		Count:      len(comments),
	}, nil
}

// ListUserComments lists all comments by a user
func (s *CommentService) ListUserComments(ctx context.Context, userID string, limit int32, cursor string) (*models.CommentsListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	comments, nextCursor, err := s.commentRepo.ListByUserID(ctx, userID, limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list user comments: %w", err)
	}

	return &models.CommentsListResponse{
		Comments:   comments,
		NextCursor: nextCursor,
		Count:      len(comments),
	}, nil
}

// SearchComments searches comments by content
func (s *CommentService) SearchComments(ctx context.Context, query string, limit int32, cursor string) (*models.CommentsListResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	comments, nextCursor, err := s.commentRepo.SearchComments(ctx, query, limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to search comments: %w", err)
	}

	// Filter comments to only include those on published songs
	filteredComments := make([]models.Comment, 0)
	for _, comment := range comments {
		song, err := s.songRepo.GetByID(ctx, comment.SongID)
		if err != nil {
			continue // Skip if can't get song
		}
		if song.IsPublished() {
			filteredComments = append(filteredComments, comment)
		}
	}

	return &models.CommentsListResponse{
		Comments:   filteredComments,
		NextCursor: nextCursor,
		Count:      len(filteredComments),
	}, nil
}

// GetCommentStats returns statistics about comments
func (s *CommentService) GetCommentStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	userCommentCount, err := s.commentRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user comment count: %w", err)
	}

	stats := map[string]interface{}{
		"total_comments": userCommentCount,
	}

	return stats, nil
}

// GetSongCommentStats returns comment statistics for a song
func (s *CommentService) GetSongCommentStats(ctx context.Context, songID string) (map[string]interface{}, error) {
	// Verify song exists
	_, err := s.songRepo.GetByID(ctx, songID)
	if err != nil {
		return nil, fmt.Errorf("song not found: %w", err)
	}

	commentCount, err := s.commentRepo.CountBySongID(ctx, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song comment count: %w", err)
	}

	stats := map[string]interface{}{
		"comment_count": commentCount,
		"song_id":       songID,
	}

	return stats, nil
}

// GetCommentsGroupedByAudio groups comments by audio URL for a song
func (s *CommentService) GetCommentsGroupedByAudio(ctx context.Context, songID string) (models.CommentsByAudioURL, error) {
	// Verify song exists
	_, err := s.songRepo.GetByID(ctx, songID)
	if err != nil {
		return nil, fmt.Errorf("song not found: %w", err)
	}

	// Get all comments for the song
	comments, _, err := s.commentRepo.ListBySongID(ctx, songID, 1000, "") // Large limit to get all
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	// Group by audio URL
	grouped := models.GroupCommentsByAudioURL(comments)

	return grouped, nil
}

// isValidAudioURL checks if an audio URL exists in the song's audio URLs
func (s *CommentService) isValidAudioURL(audioURL string, songAudioURLs []string) bool {
	for _, url := range songAudioURLs {
		if url == audioURL {
			return true
		}
	}
	return false
}

// canEditComment determines if a comment can still be edited
func (s *CommentService) canEditComment(comment *models.Comment) bool {
	// Business rule: comments can be edited within 24 hours of creation
	// You could also implement this based on whether the song is still published
	// or other business rules
	return !comment.IsRecent(-24 * 60 * 60 * 1000000000) // 24 hours in nanoseconds, but negative so it returns true if within 24 hours
}

// ValidateCommentOwnership checks if user owns the comment (helper method)
func (s *CommentService) ValidateCommentOwnership(ctx context.Context, commentID, userID string) error {
	_, err := s.commentRepo.GetByUserIDAndID(ctx, userID, commentID)
	if err != nil {
		if err == repository.ErrUnauthorized {
			return fmt.Errorf("user does not own this comment")
		}
		return fmt.Errorf("failed to validate ownership: %w", err)
	}
	return nil
}

// DeleteCommentsByUser deletes all comments by a specific user (admin function)
func (s *CommentService) DeleteCommentsByUser(ctx context.Context, targetUserID, adminUserID string) error {
	// This would be an admin-only function in a real system
	// For now, just return an error as it's not implemented
	return fmt.Errorf("admin functions not implemented")
}

// ModerationAction represents moderation actions that can be taken on comments
type ModerationAction string

const (
	ModerationActionHide   ModerationAction = "hide"
	ModerationActionDelete ModerationAction = "delete"
	ModerationActionFlag   ModerationAction = "flag"
)

// ModerateComment applies moderation actions to comments (placeholder for future implementation)
func (s *CommentService) ModerateComment(ctx context.Context, commentID, moderatorUserID string, action ModerationAction, reason string) error {
	// This would be implemented in a full system with moderation features
	return fmt.Errorf("moderation features not implemented")
}