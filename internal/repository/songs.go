package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/cncde/songs-backend/internal/models"
)

// SongRepository handles song data operations in DynamoDB
type SongRepository struct {
	client    *DynamoDBClient
	tableName string
}

// NewSongRepository creates a new song repository
func NewSongRepository(client *DynamoDBClient, tableName string) *SongRepository {
	return &SongRepository{
		client:    client,
		tableName: tableName,
	}
}

// Create creates a new song
func (r *SongRepository) Create(ctx context.Context, song *models.Song) error {
	conditionExpression := aws.String("attribute_not_exists(id)")
	
	if err := r.client.PutItem(ctx, r.tableName, song, conditionExpression); err != nil {
		return fmt.Errorf("failed to create song: %w", err)
	}

	return nil
}

// GetByID retrieves a song by its ID
func (r *SongRepository) GetByID(ctx context.Context, id string) (*models.Song, error) {
	key := BuildKey("id", id)
	
	var song models.Song
	if err := r.client.GetItem(ctx, r.tableName, key, &song); err != nil {
		if err == ErrItemNotFound {
			return nil, fmt.Errorf("song with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	return &song, nil
}

// Update updates an existing song
func (r *SongRepository) Update(ctx context.Context, song *models.Song) error {
	key := BuildKey("id", song.ID)
	
	// Build update expression
	updateExpr := "SET title = :title, subtitle = :subtitle, description = :description, " +
		"image_urls = :image_urls, audio_urls = :audio_urls, lyrics = :lyrics, " +
		"stage = :stage, tags = :tags, updated_at = :updated_at"
	
	expressionAttributeValues := map[string]types.AttributeValue{
		":title":       &types.AttributeValueMemberS{Value: song.Title},
		":subtitle":    &types.AttributeValueMemberS{Value: song.Subtitle},
		":description": &types.AttributeValueMemberS{Value: song.Description},
		":lyrics":      &types.AttributeValueMemberS{Value: song.Lyrics},
		":stage":       &types.AttributeValueMemberS{Value: song.Stage},
		":updated_at":  &types.AttributeValueMemberS{Value: song.UpdatedAt},
	}

	// Handle image URLs
	imageURLsAttr, err := attributevalue.Marshal(song.ImageURLs)
	if err != nil {
		return fmt.Errorf("failed to marshal image URLs: %w", err)
	}
	expressionAttributeValues[":image_urls"] = imageURLsAttr

	// Handle audio URLs
	audioURLsAttr, err := attributevalue.Marshal(song.AudioURLs)
	if err != nil {
		return fmt.Errorf("failed to marshal audio URLs: %w", err)
	}
	expressionAttributeValues[":audio_urls"] = audioURLsAttr

	// Handle tags
	tagsAttr, err := attributevalue.Marshal(song.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	expressionAttributeValues[":tags"] = tagsAttr

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.tableName),
		Key:                       key,
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: expressionAttributeValues,
		ConditionExpression:       aws.String("attribute_exists(id)"),
	}

	_, err = r.client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update song: %w", err)
	}

	return nil
}

// Delete deletes a song by ID
func (r *SongRepository) Delete(ctx context.Context, id, userID string) error {
	key := BuildKey("id", id)
	conditionExpression := aws.String("attribute_exists(id) AND user_id = :user_id")
	
	// We need to pass expression attribute values for the condition
	input := &dynamodb.DeleteItemInput{
		TableName:           aws.String(r.tableName),
		Key:                 key,
		ConditionExpression: conditionExpression,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: userID},
		},
	}

	_, err := r.client.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete song: %w", err)
	}

	return nil
}

// ListByUserID lists songs for a specific user with pagination
func (r *SongRepository) ListByUserID(ctx context.Context, userID string, limit int32, cursor string) ([]models.Song, string, error) {
	// Build query input for GSI on user_id
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("UserIDIndex"), // Assume this GSI exists
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: userID},
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
		Limit:            aws.Int32(limit),
	}

	// Handle pagination cursor
	if cursor != "" {
		token, err := DecodePaginationToken(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		if token != nil {
			input.ExclusiveStartKey = token.LastEvaluatedKey
		}
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list songs: %w", err)
	}

	songs := make([]models.Song, 0, len(result.Items))
	for _, item := range result.Items {
		var song models.Song
		if err := attributevalue.UnmarshalMap(item, &song); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal song: %w", err)
		}
		songs = append(songs, song)
	}

	// Build next cursor
	var nextCursor string
	if result.LastEvaluatedKey != nil {
		token := &PaginationToken{
			LastEvaluatedKey: result.LastEvaluatedKey,
			Limit:            limit,
		}
		nextCursor, err = EncodePaginationToken(token)
		if err != nil {
			return nil, "", fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	return songs, nextCursor, nil
}

// ListByStage lists songs by stage with pagination
func (r *SongRepository) ListByStage(ctx context.Context, stage string, limit int32, cursor string) ([]models.Song, string, error) {
	// Build query input for GSI on stage
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("StageIndex"), // Assume this GSI exists
		KeyConditionExpression: aws.String("stage = :stage"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":stage": &types.AttributeValueMemberS{Value: stage},
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
		Limit:            aws.Int32(limit),
	}

	// Handle pagination cursor
	if cursor != "" {
		token, err := DecodePaginationToken(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		if token != nil {
			input.ExclusiveStartKey = token.LastEvaluatedKey
		}
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list songs by stage: %w", err)
	}

	songs := make([]models.Song, 0, len(result.Items))
	for _, item := range result.Items {
		var song models.Song
		if err := attributevalue.UnmarshalMap(item, &song); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal song: %w", err)
		}
		songs = append(songs, song)
	}

	// Build next cursor
	var nextCursor string
	if result.LastEvaluatedKey != nil {
		token := &PaginationToken{
			LastEvaluatedKey: result.LastEvaluatedKey,
			Limit:            limit,
		}
		nextCursor, err = EncodePaginationToken(token)
		if err != nil {
			return nil, "", fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	return songs, nextCursor, nil
}

// SearchByTitle searches songs by title using scan (could be optimized with search service)
func (r *SongRepository) SearchByTitle(ctx context.Context, titleQuery string, limit int32, cursor string) ([]models.Song, string, error) {
	input := &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("contains(title, :title)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":title": &types.AttributeValueMemberS{Value: titleQuery},
		},
		Limit: aws.Int32(limit),
	}

	// Handle pagination cursor
	if cursor != "" {
		token, err := DecodePaginationToken(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
		if token != nil {
			input.ExclusiveStartKey = token.LastEvaluatedKey
		}
	}

	result, err := r.client.Scan(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("failed to search songs: %w", err)
	}

	songs := make([]models.Song, 0, len(result.Items))
	for _, item := range result.Items {
		var song models.Song
		if err := attributevalue.UnmarshalMap(item, &song); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal song: %w", err)
		}
		songs = append(songs, song)
	}

	// Build next cursor
	var nextCursor string
	if result.LastEvaluatedKey != nil {
		token := &PaginationToken{
			LastEvaluatedKey: result.LastEvaluatedKey,
			Limit:            limit,
		}
		nextCursor, err = EncodePaginationToken(token)
		if err != nil {
			return nil, "", fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	return songs, nextCursor, nil
}

// GetByUserIDAndID retrieves a song by user ID and song ID for authorization
func (r *SongRepository) GetByUserIDAndID(ctx context.Context, userID, songID string) (*models.Song, error) {
	song, err := r.GetByID(ctx, songID)
	if err != nil {
		return nil, err
	}

	if song.UserID != userID {
		return nil, ErrUnauthorized
	}

	return song, nil
}

// UpdateStage updates only the stage of a song
func (r *SongRepository) UpdateStage(ctx context.Context, songID, userID, newStage string) error {
	key := BuildKey("id", songID)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
		UpdateExpression: aws.String("SET stage = :stage, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":stage":      &types.AttributeValueMemberS{Value: newStage},
			":updated_at": &types.AttributeValueMemberS{Value: models.Song{}.UpdatedAt}, // This should be current time
			":user_id":    &types.AttributeValueMemberS{Value: userID},
		},
		ConditionExpression: aws.String("attribute_exists(id) AND user_id = :user_id"),
	}

	_, err := r.client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update song stage: %w", err)
	}

	return nil
}

// CountByUserID returns the total number of songs for a user
func (r *SongRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("UserIDIndex"),
		KeyConditionExpression: aws.String("user_id = :user_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":user_id": &types.AttributeValueMemberS{Value: userID},
		},
		Select: types.SelectCount,
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to count songs: %w", err)
	}

	return int(result.Count), nil
}