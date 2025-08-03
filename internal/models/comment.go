package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Comment represents a comment on a song's audio file
type Comment struct {
	ID        string `json:"id" dynamodbav:"id" validate:"required"`
	SongID    string `json:"song_id" dynamodbav:"song_id" validate:"required"`
	AudioURL  string `json:"audio_url" dynamodbav:"audio_url" validate:"required,url"`
	Content   string `json:"content" dynamodbav:"content" validate:"required,min=1,max=2000"`
	UserID    string `json:"user_id" dynamodbav:"user_id" validate:"required"`
	CreatedAt string `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt string `json:"updated_at" dynamodbav:"updated_at"`
}

// CreateCommentRequest represents the request payload for creating a comment
type CreateCommentRequest struct {
	AudioURL string `json:"audio_url" validate:"required,url"`
	Content  string `json:"content" validate:"required,min=1,max=2000"`
}

// UpdateCommentRequest represents the request payload for updating a comment
type UpdateCommentRequest struct {
	Content *string `json:"content,omitempty" validate:"omitempty,min=1,max=2000"`
}

// CommentResponse represents the response format for comment operations
type CommentResponse struct {
	Comment *Comment `json:"comment,omitempty"`
	Message string   `json:"message,omitempty"`
}

// CommentsListResponse represents the response format for listing comments
type CommentsListResponse struct {
	Comments   []Comment `json:"comments"`
	NextCursor string    `json:"next_cursor,omitempty"`
	Count      int       `json:"count"`
}

// NewComment creates a new Comment instance with generated ID and timestamps
func NewComment(req CreateCommentRequest, songID, userID string) *Comment {
	now := time.Now().UTC().Format(time.RFC3339)

	comment := &Comment{
		ID:        uuid.New().String(),
		SongID:    songID,
		AudioURL:  strings.TrimSpace(req.AudioURL),
		Content:   sanitizeContent(req.Content),
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return comment
}

// Update applies changes from UpdateCommentRequest to the comment
func (c *Comment) Update(req UpdateCommentRequest) {
	now := time.Now().UTC().Format(time.RFC3339)

	if req.Content != nil {
		c.Content = sanitizeContent(*req.Content)
		c.UpdatedAt = now
	}
}

// CanBeEditedBy checks if the comment can be edited by the given user
func (c *Comment) CanBeEditedBy(userID string) bool {
	return c.UserID == userID
}

// IsEmpty returns true if the comment content is empty after sanitization
func (c *Comment) IsEmpty() bool {
	return strings.TrimSpace(c.Content) == ""
}

// GetWordCount returns the word count of the comment content
func (c *Comment) GetWordCount() int {
	content := strings.TrimSpace(c.Content)
	if content == "" {
		return 0
	}
	
	words := strings.Fields(content)
	return len(words)
}

// IsTooLong returns true if the comment exceeds maximum length
func (c *Comment) IsTooLong() bool {
	return len(c.Content) > 2000
}

// GetCreatedTime returns the created time as time.Time
func (c *Comment) GetCreatedTime() (time.Time, error) {
	return time.Parse(time.RFC3339, c.CreatedAt)
}

// GetUpdatedTime returns the updated time as time.Time
func (c *Comment) GetUpdatedTime() (time.Time, error) {
	return time.Parse(time.RFC3339, c.UpdatedAt)
}

// IsRecent returns true if the comment was created within the specified duration
func (c *Comment) IsRecent(duration time.Duration) bool {
	created, err := c.GetCreatedTime()
	if err != nil {
		return false
	}
	
	return time.Since(created) <= duration
}

// WasModified returns true if the comment was modified after creation
func (c *Comment) WasModified() bool {
	return c.CreatedAt != c.UpdatedAt
}

// sanitizeContent cleans the comment content
func sanitizeContent(content string) string {
	// Trim whitespace
	content = strings.TrimSpace(content)
	
	// Replace multiple consecutive whitespace characters with single space
	content = strings.Join(strings.Fields(content), " ")
	
	return content
}

// CommentsByAudioURL groups comments by their audio URL
type CommentsByAudioURL map[string][]Comment

// GroupCommentsByAudioURL groups a slice of comments by their audio URL
func GroupCommentsByAudioURL(comments []Comment) CommentsByAudioURL {
	grouped := make(CommentsByAudioURL)
	
	for _, comment := range comments {
		grouped[comment.AudioURL] = append(grouped[comment.AudioURL], comment)
	}
	
	return grouped
}

// GetAudioURLs returns a unique list of audio URLs from the grouped comments
func (g CommentsByAudioURL) GetAudioURLs() []string {
	urls := make([]string, 0, len(g))
	for url := range g {
		urls = append(urls, url)
	}
	return urls
}

// GetTotalCount returns the total number of comments across all audio URLs
func (g CommentsByAudioURL) GetTotalCount() int {
	total := 0
	for _, comments := range g {
		total += len(comments)
	}
	return total
}

// CommentSortBy represents different sorting options for comments
type CommentSortBy string

const (
	SortByCreatedAtAsc  CommentSortBy = "created_at_asc"
	SortByCreatedAtDesc CommentSortBy = "created_at_desc"
	SortByUpdatedAtAsc  CommentSortBy = "updated_at_asc"
	SortByUpdatedAtDesc CommentSortBy = "updated_at_desc"
)

// IsValidSortBy checks if the sort option is valid
func IsValidSortBy(sortBy string) bool {
	switch CommentSortBy(sortBy) {
	case SortByCreatedAtAsc, SortByCreatedAtDesc, SortByUpdatedAtAsc, SortByUpdatedAtDesc:
		return true
	default:
		return false
	}
}