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

// CommentRepository handles comment data operations in DynamoDB
type CommentRepository struct {
	client    *DynamoDBClient
	tableName string
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(client *DynamoDBClient, tableName string) *CommentRepository {
	return &CommentRepository{
		client:    client,
		tableName: tableName,
	}
}

// Create creates a new comment
func (r *CommentRepository) Create(ctx context.Context, comment *models.Comment) error {
	conditionExpression := aws.String("attribute_not_exists(id)")
	
	if err := r.client.PutItem(ctx, r.tableName, comment, conditionExpression); err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// GetByID retrieves a comment by its ID
func (r *CommentRepository) GetByID(ctx context.Context, id string) (*models.Comment, error) {
	key := BuildKey("id", id)
	
	var comment models.Comment
	if err := r.client.GetItem(ctx, r.tableName, key, &comment); err != nil {
		if err == ErrItemNotFound {
			return nil, fmt.Errorf("comment with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &comment, nil
}

// Update updates an existing comment
func (r *CommentRepository) Update(ctx context.Context, comment *models.Comment) error {
	key := BuildKey("id", comment.ID)
	
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
		UpdateExpression: aws.String("SET content = :content, updated_at = :updated_at"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":content":    &types.AttributeValueMemberS{Value: comment.Content},
			":updated_at": &types.AttributeValueMemberS{Value: comment.UpdatedAt},
			":user_id":    &types.AttributeValueMemberS{Value: comment.UserID},
		},
		ConditionExpression: aws.String("attribute_exists(id) AND user_id = :user_id"),
	}

	_, err := r.client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	return nil
}

// Delete deletes a comment by ID (only if user owns it)
func (r *CommentRepository) Delete(ctx context.Context, id, userID string) error {
	key := BuildKey("id", id)
	conditionExpression := aws.String("attribute_exists(id) AND user_id = :user_id")
	
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
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

// ListBySongID lists all comments for a specific song with pagination
func (r *CommentRepository) ListBySongID(ctx context.Context, songID string, limit int32, cursor string) ([]models.Comment, string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("SongIDIndex"), // Assume this GSI exists
		KeyConditionExpression: aws.String("song_id = :song_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":song_id": &types.AttributeValueMemberS{Value: songID},
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
		return nil, "", fmt.Errorf("failed to list comments: %w", err)
	}

	comments := make([]models.Comment, 0, len(result.Items))
	for _, item := range result.Items {
		var comment models.Comment
		if err := attributevalue.UnmarshalMap(item, &comment); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal comment: %w", err)
		}
		comments = append(comments, comment)
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

	return comments, nextCursor, nil
}

// ListBySongIDAndAudioURL lists comments for a specific song and audio file
func (r *CommentRepository) ListBySongIDAndAudioURL(ctx context.Context, songID, audioURL string, limit int32, cursor string) ([]models.Comment, string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("SongID-AudioURL-Index"), // Composite GSI
		KeyConditionExpression: aws.String("song_id = :song_id AND audio_url = :audio_url"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":song_id":   &types.AttributeValueMemberS{Value: songID},
			":audio_url": &types.AttributeValueMemberS{Value: audioURL},
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
		return nil, "", fmt.Errorf("failed to list comments by audio URL: %w", err)
	}

	comments := make([]models.Comment, 0, len(result.Items))
	for _, item := range result.Items {
		var comment models.Comment
		if err := attributevalue.UnmarshalMap(item, &comment); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal comment: %w", err)
		}
		comments = append(comments, comment)
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

	return comments, nextCursor, nil
}

// ListByUserID lists all comments by a specific user with pagination
func (r *CommentRepository) ListByUserID(ctx context.Context, userID string, limit int32, cursor string) ([]models.Comment, string, error) {
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
		return nil, "", fmt.Errorf("failed to list user comments: %w", err)
	}

	comments := make([]models.Comment, 0, len(result.Items))
	for _, item := range result.Items {
		var comment models.Comment
		if err := attributevalue.UnmarshalMap(item, &comment); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal comment: %w", err)
		}
		comments = append(comments, comment)
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

	return comments, nextCursor, nil
}

// GetByUserIDAndID retrieves a comment by user ID and comment ID for authorization
func (r *CommentRepository) GetByUserIDAndID(ctx context.Context, userID, commentID string) (*models.Comment, error) {
	comment, err := r.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	if comment.UserID != userID {
		return nil, ErrUnauthorized
	}

	return comment, nil
}

// CountBySongID returns the total number of comments for a song
func (r *CommentRepository) CountBySongID(ctx context.Context, songID string) (int, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("SongIDIndex"),
		KeyConditionExpression: aws.String("song_id = :song_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":song_id": &types.AttributeValueMemberS{Value: songID},
		},
		Select: types.SelectCount,
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}

	return int(result.Count), nil
}

// CountByUserID returns the total number of comments by a user
func (r *CommentRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
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
		return 0, fmt.Errorf("failed to count user comments: %w", err)
	}

	return int(result.Count), nil
}

// DeleteBySongID deletes all comments for a song (used when song is deleted)
func (r *CommentRepository) DeleteBySongID(ctx context.Context, songID string) error {
	// First, get all comments for the song
	comments, _, err := r.ListBySongID(ctx, songID, 100, "")
	if err != nil {
		return fmt.Errorf("failed to list comments for deletion: %w", err)
	}

	// Delete comments in batches
	for i := 0; i < len(comments); i += 25 { // DynamoDB batch write limit is 25
		end := i + 25
		if end > len(comments) {
			end = len(comments)
		}

		batch := comments[i:end]
		if err := r.batchDeleteComments(ctx, batch); err != nil {
			return fmt.Errorf("failed to batch delete comments: %w", err)
		}
	}

	return nil
}

// batchDeleteComments deletes a batch of comments
func (r *CommentRepository) batchDeleteComments(ctx context.Context, comments []models.Comment) error {
	if len(comments) == 0 {
		return nil
	}

	writeRequests := make([]types.WriteRequest, 0, len(comments))
	
	for _, comment := range comments {
		key := BuildKey("id", comment.ID)
		writeRequests = append(writeRequests, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: key,
			},
		})
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			r.tableName: writeRequests,
		},
	}

	_, err := r.client.BatchWriteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to batch delete comments: %w", err)
	}

	return nil
}

// SearchComments searches comments by content (basic text search using scan)
func (r *CommentRepository) SearchComments(ctx context.Context, query string, limit int32, cursor string) ([]models.Comment, string, error) {
	input := &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("contains(content, :query)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":query": &types.AttributeValueMemberS{Value: query},
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
		return nil, "", fmt.Errorf("failed to search comments: %w", err)
	}

	comments := make([]models.Comment, 0, len(result.Items))
	for _, item := range result.Items {
		var comment models.Comment
		if err := attributevalue.UnmarshalMap(item, &comment); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal comment: %w", err)
		}
		comments = append(comments, comment)
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

	return comments, nextCursor, nil
}