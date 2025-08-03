package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Song represents a song with all its metadata
type Song struct {
	ID          string   `json:"id" dynamodbav:"id" validate:"required"`
	Title       string   `json:"title" dynamodbav:"title" validate:"required,min=1,max=255"`
	Subtitle    string   `json:"subtitle" dynamodbav:"subtitle" validate:"max=255"`
	Description string   `json:"description" dynamodbav:"description" validate:"max=2000"`
	ImageURLs   []string `json:"image_urls" dynamodbav:"image_urls"`
	AudioURLs   []string `json:"audio_urls" dynamodbav:"audio_urls"`
	Lyrics      string   `json:"lyrics" dynamodbav:"lyrics" validate:"max=10000"`
	Stage       string   `json:"stage" dynamodbav:"stage" validate:"oneof=draft review published archived"`
	Tags        []string `json:"tags" dynamodbav:"tags"`
	UserID      string   `json:"user_id" dynamodbav:"user_id" validate:"required"`
	CreatedAt   string   `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt   string   `json:"updated_at" dynamodbav:"updated_at"`
}

// CreateSongRequest represents the request payload for creating a song
type CreateSongRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=255"`
	Subtitle    string   `json:"subtitle" validate:"max=255"`
	Description string   `json:"description" validate:"max=2000"`
	Lyrics      string   `json:"lyrics" validate:"max=10000"`
	Stage       string   `json:"stage" validate:"oneof=draft review published"`
	Tags        []string `json:"tags" validate:"max=20,dive,min=1,max=50"`
}

// UpdateSongRequest represents the request payload for updating a song
type UpdateSongRequest struct {
	Title       *string  `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Subtitle    *string  `json:"subtitle,omitempty" validate:"omitempty,max=255"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=2000"`
	Lyrics      *string  `json:"lyrics,omitempty" validate:"omitempty,max=10000"`
	Stage       *string  `json:"stage,omitempty" validate:"omitempty,oneof=draft review published archived"`
	Tags        []string `json:"tags,omitempty" validate:"omitempty,max=20,dive,min=1,max=50"`
}

// SongResponse represents the response format for song operations
type SongResponse struct {
	Song    *Song  `json:"song,omitempty"`
	Message string `json:"message,omitempty"`
}

// SongsListResponse represents the response format for listing songs
type SongsListResponse struct {
	Songs      []Song `json:"songs"`
	NextCursor string `json:"next_cursor,omitempty"`
	Count      int    `json:"count"`
}

// NewSong creates a new Song instance with generated ID and timestamps
func NewSong(req CreateSongRequest, userID string) *Song {
	now := time.Now().UTC().Format(time.RFC3339)
	
	song := &Song{
		ID:          uuid.New().String(),
		Title:       strings.TrimSpace(req.Title),
		Subtitle:    strings.TrimSpace(req.Subtitle),
		Description: strings.TrimSpace(req.Description),
		ImageURLs:   []string{}, // Initialize as empty slice
		AudioURLs:   []string{}, // Initialize as empty slice
		Lyrics:      strings.TrimSpace(req.Lyrics),
		Stage:       getStageOrDefault(req.Stage),
		Tags:        sanitizeTags(req.Tags),
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return song
}

// Update applies changes from UpdateSongRequest to the song
func (s *Song) Update(req UpdateSongRequest) {
	now := time.Now().UTC().Format(time.RFC3339)

	if req.Title != nil {
		s.Title = strings.TrimSpace(*req.Title)
	}
	if req.Subtitle != nil {
		s.Subtitle = strings.TrimSpace(*req.Subtitle)
	}
	if req.Description != nil {
		s.Description = strings.TrimSpace(*req.Description)
	}
	if req.Lyrics != nil {
		s.Lyrics = strings.TrimSpace(*req.Lyrics)
	}
	if req.Stage != nil {
		s.Stage = *req.Stage
	}
	if req.Tags != nil {
		s.Tags = sanitizeTags(req.Tags)
	}

	s.UpdatedAt = now
}

// AddImageURL adds an image URL to the song
func (s *Song) AddImageURL(url string) error {
	if url == "" {
		return errors.New("image URL cannot be empty")
	}

	// Check if URL already exists
	for _, existingURL := range s.ImageURLs {
		if existingURL == url {
			return errors.New("image URL already exists")
		}
	}

	s.ImageURLs = append(s.ImageURLs, url)
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return nil
}

// AddAudioURL adds an audio URL to the song
func (s *Song) AddAudioURL(url string) error {
	if url == "" {
		return errors.New("audio URL cannot be empty")
	}

	// Check if URL already exists
	for _, existingURL := range s.AudioURLs {
		if existingURL == url {
			return errors.New("audio URL already exists")
		}
	}

	s.AudioURLs = append(s.AudioURLs, url)
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return nil
}

// RemoveImageURL removes an image URL from the song
func (s *Song) RemoveImageURL(url string) {
	for i, existingURL := range s.ImageURLs {
		if existingURL == url {
			s.ImageURLs = append(s.ImageURLs[:i], s.ImageURLs[i+1:]...)
			s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			return
		}
	}
}

// RemoveAudioURL removes an audio URL from the song
func (s *Song) RemoveAudioURL(url string) {
	for i, existingURL := range s.AudioURLs {
		if existingURL == url {
			s.AudioURLs = append(s.AudioURLs[:i], s.AudioURLs[i+1:]...)
			s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			return
		}
	}
}

// IsPublished returns true if the song is in published stage
func (s *Song) IsPublished() bool {
	return s.Stage == "published"
}

// IsDraft returns true if the song is in draft stage
func (s *Song) IsDraft() bool {
	return s.Stage == "draft"
}

// CanBeEdited returns true if the song can be edited (not archived)
func (s *Song) CanBeEdited() bool {
	return s.Stage != "archived"
}

// getStageOrDefault returns the stage or default to "draft" if empty
func getStageOrDefault(stage string) string {
	if stage == "" {
		return "draft"
	}
	return stage
}

// sanitizeTags cleans and deduplicates tags
func sanitizeTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}

	seen := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		trimmed := strings.TrimSpace(strings.ToLower(tag))
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			result = append(result, trimmed)
		}
	}

	return result
}

