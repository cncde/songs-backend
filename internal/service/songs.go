package service

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/cncde/songs-backend/internal/models"
	"github.com/cncde/songs-backend/internal/repository"
)

// SongService handles song business logic
type SongService struct {
	songRepo    *repository.SongRepository
	commentRepo *repository.CommentRepository
	validator   *validator.Validate
}

// NewSongService creates a new song service
func NewSongService(songRepo *repository.SongRepository, commentRepo *repository.CommentRepository) *SongService {
	return &SongService{
		songRepo:    songRepo,
		commentRepo: commentRepo,
		validator:   validator.New(),
	}
}

// CreateSong creates a new song with business logic validation
func (s *SongService) CreateSong(ctx context.Context, userID string, req models.CreateSongRequest) (*models.Song, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create song model
	song := models.NewSong(req, userID)


	// Save to repository
	if err := s.songRepo.Create(ctx, song); err != nil {
		if err == repository.ErrItemAlreadyExists {
			return nil, fmt.Errorf("song with this ID already exists")
		}
		return nil, fmt.Errorf("failed to create song: %w", err)
	}

	return song, nil
}

// GetSong retrieves a song by ID with authorization check
func (s *SongService) GetSong(ctx context.Context, songID, userID string) (*models.Song, error) {
	if songID == "" {
		return nil, fmt.Errorf("song ID is required")
	}

	song, err := s.songRepo.GetByID(ctx, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	// Check authorization - users can only access their own songs unless published
	if song.UserID != userID && !song.IsPublished() {
		return nil, repository.ErrUnauthorized
	}

	return song, nil
}

// GetSongForOwner retrieves a song by ID only if user owns it
func (s *SongService) GetSongForOwner(ctx context.Context, songID, userID string) (*models.Song, error) {
	if songID == "" {
		return nil, fmt.Errorf("song ID is required")
	}

	song, err := s.songRepo.GetByUserIDAndID(ctx, userID, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	return song, nil
}

// UpdateSong updates an existing song
func (s *SongService) UpdateSong(ctx context.Context, songID, userID string, req models.UpdateSongRequest) (*models.Song, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing song
	song, err := s.songRepo.GetByUserIDAndID(ctx, userID, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	// Check if song can be edited
	if !song.CanBeEdited() {
		return nil, fmt.Errorf("song cannot be edited in %s stage", song.Stage)
	}


	// Apply updates
	song.Update(req)

	// Save to repository
	if err := s.songRepo.Update(ctx, song); err != nil {
		return nil, fmt.Errorf("failed to update song: %w", err)
	}

	return song, nil
}

// DeleteSong deletes a song and all its comments
func (s *SongService) DeleteSong(ctx context.Context, songID, userID string) error {
	// Verify ownership
	song, err := s.songRepo.GetByUserIDAndID(ctx, userID, songID)
	if err != nil {
		return fmt.Errorf("failed to get song: %w", err)
	}

	// Check if song can be deleted (business rule: can't delete published songs)
	if song.IsPublished() {
		return fmt.Errorf("cannot delete published songs")
	}

	// Delete all comments for this song first
	if err := s.commentRepo.DeleteBySongID(ctx, songID); err != nil {
		return fmt.Errorf("failed to delete song comments: %w", err)
	}

	// Delete the song
	if err := s.songRepo.Delete(ctx, songID, userID); err != nil {
		return fmt.Errorf("failed to delete song: %w", err)
	}

	return nil
}

// ListUserSongs lists songs for a specific user
func (s *SongService) ListUserSongs(ctx context.Context, userID string, limit int32, cursor string) (*models.SongsListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}

	songs, nextCursor, err := s.songRepo.ListByUserID(ctx, userID, limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list user songs: %w", err)
	}

	return &models.SongsListResponse{
		Songs:      songs,
		NextCursor: nextCursor,
		Count:      len(songs),
	}, nil
}

// ListPublishedSongs lists all published songs (public endpoint)
func (s *SongService) ListPublishedSongs(ctx context.Context, limit int32, cursor string) (*models.SongsListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	songs, nextCursor, err := s.songRepo.ListByStage(ctx, "published", limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to list published songs: %w", err)
	}

	return &models.SongsListResponse{
		Songs:      songs,
		NextCursor: nextCursor,
		Count:      len(songs),
	}, nil
}

// SearchSongs searches songs by title
func (s *SongService) SearchSongs(ctx context.Context, query string, limit int32, cursor string) (*models.SongsListResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	songs, nextCursor, err := s.songRepo.SearchByTitle(ctx, query, limit, cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to search songs: %w", err)
	}

	// Filter out non-published songs for search results
	publishedSongs := make([]models.Song, 0)
	for _, song := range songs {
		if song.IsPublished() {
			publishedSongs = append(publishedSongs, song)
		}
	}

	return &models.SongsListResponse{
		Songs:      publishedSongs,
		NextCursor: nextCursor,
		Count:      len(publishedSongs),
	}, nil
}

// UpdateSongStage updates only the stage of a song
func (s *SongService) UpdateSongStage(ctx context.Context, songID, userID, newStage string) (*models.Song, error) {
	// Get current song to verify ownership
	_, err := s.songRepo.GetByUserIDAndID(ctx, userID, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	// Update stage in repository
	if err := s.songRepo.UpdateStage(ctx, songID, userID, newStage); err != nil {
		return nil, fmt.Errorf("failed to update song stage: %w", err)
	}

	// Return updated song
	return s.songRepo.GetByID(ctx, songID)
}

// AddImageURL adds an image URL to a song
func (s *SongService) AddImageURL(ctx context.Context, songID, userID, imageURL string) (*models.Song, error) {
	song, err := s.songRepo.GetByUserIDAndID(ctx, userID, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	if !song.CanBeEdited() {
		return nil, fmt.Errorf("song cannot be edited in %s stage", song.Stage)
	}

	// Add image URL to song
	if err := song.AddImageURL(imageURL); err != nil {
		return nil, fmt.Errorf("failed to add image URL: %w", err)
	}

	// Update in repository
	if err := s.songRepo.Update(ctx, song); err != nil {
		return nil, fmt.Errorf("failed to update song: %w", err)
	}

	return song, nil
}

// AddAudioURL adds an audio URL to a song
func (s *SongService) AddAudioURL(ctx context.Context, songID, userID, audioURL string) (*models.Song, error) {
	song, err := s.songRepo.GetByUserIDAndID(ctx, userID, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	if !song.CanBeEdited() {
		return nil, fmt.Errorf("song cannot be edited in %s stage", song.Stage)
	}

	// Add audio URL to song
	if err := song.AddAudioURL(audioURL); err != nil {
		return nil, fmt.Errorf("failed to add audio URL: %w", err)
	}

	// Update in repository
	if err := s.songRepo.Update(ctx, song); err != nil {
		return nil, fmt.Errorf("failed to update song: %w", err)
	}

	return song, nil
}

// RemoveImageURL removes an image URL from a song
func (s *SongService) RemoveImageURL(ctx context.Context, songID, userID, imageURL string) (*models.Song, error) {
	song, err := s.songRepo.GetByUserIDAndID(ctx, userID, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	if !song.CanBeEdited() {
		return nil, fmt.Errorf("song cannot be edited in %s stage", song.Stage)
	}

	// Remove image URL from song
	song.RemoveImageURL(imageURL)

	// Update in repository
	if err := s.songRepo.Update(ctx, song); err != nil {
		return nil, fmt.Errorf("failed to update song: %w", err)
	}

	return song, nil
}

// RemoveAudioURL removes an audio URL from a song
func (s *SongService) RemoveAudioURL(ctx context.Context, songID, userID, audioURL string) (*models.Song, error) {
	song, err := s.songRepo.GetByUserIDAndID(ctx, userID, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	if !song.CanBeEdited() {
		return nil, fmt.Errorf("song cannot be edited in %s stage", song.Stage)
	}

	// Remove audio URL from song
	song.RemoveAudioURL(audioURL)

	// Update in repository
	if err := s.songRepo.Update(ctx, song); err != nil {
		return nil, fmt.Errorf("failed to update song: %w", err)
	}

	return song, nil
}

// GetSongStats returns statistics about a user's songs
func (s *SongService) GetSongStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	totalCount, err := s.songRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song count: %w", err)
	}

	// Get counts by stage (this would require additional repository methods)
	stats := map[string]interface{}{
		"total_songs": totalCount,
		// Could add more stats here like:
		// "draft_count": draftCount,
		// "published_count": publishedCount,
		// "total_comments": commentCount,
	}

	return stats, nil
}

// ValidateSongOwnership checks if user owns the song (helper method)
func (s *SongService) ValidateSongOwnership(ctx context.Context, songID, userID string) error {
	_, err := s.songRepo.GetByUserIDAndID(ctx, userID, songID)
	if err != nil {
		if err == repository.ErrUnauthorized {
			return fmt.Errorf("user does not own this song")
		}
		return fmt.Errorf("failed to validate ownership: %w", err)
	}
	return nil
}