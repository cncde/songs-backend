package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient wraps the AWS DynamoDB client with common operations
type DynamoDBClient struct {
	client *dynamodb.Client
}

// NewDynamoDBClient creates a new DynamoDB client wrapper
func NewDynamoDBClient(client *dynamodb.Client) *DynamoDBClient {
	return &DynamoDBClient{
		client: client,
	}
}

// PutItem puts an item into a DynamoDB table
func (d *DynamoDBClient) PutItem(ctx context.Context, tableName string, item interface{}, conditionExpression *string) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	if conditionExpression != nil {
		input.ConditionExpression = conditionExpression
	}

	_, err = d.client.PutItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return ErrItemAlreadyExists
		}
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

// GetItem gets an item from a DynamoDB table by key
func (d *DynamoDBClient) GetItem(ctx context.Context, tableName string, key map[string]types.AttributeValue, result interface{}) error {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	}

	output, err := d.client.GetItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	if output.Item == nil {
		return ErrItemNotFound
	}

	if err := attributevalue.UnmarshalMap(output.Item, result); err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return nil
}

// UpdateItem updates an item in a DynamoDB table
func (d *DynamoDBClient) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	output, err := d.client.UpdateItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return nil, ErrItemNotFound
		}
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return output, nil
}

// DeleteItem deletes an item from a DynamoDB table
func (d *DynamoDBClient) DeleteItem(ctx context.Context, tableName string, key map[string]types.AttributeValue, conditionExpression *string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	}

	if conditionExpression != nil {
		input.ConditionExpression = conditionExpression
	}

	_, err := d.client.DeleteItem(ctx, input)
	if err != nil {
		var conditionalCheckFailedException *types.ConditionalCheckFailedException
		if errors.As(err, &conditionalCheckFailedException) {
			return ErrItemNotFound
		}
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// Query performs a query operation on a DynamoDB table
func (d *DynamoDBClient) Query(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	output, err := d.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	return output, nil
}

// Scan performs a scan operation on a DynamoDB table
func (d *DynamoDBClient) Scan(ctx context.Context, input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	output, err := d.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan items: %w", err)
	}

	return output, nil
}

// BatchGetItem gets multiple items from DynamoDB tables
func (d *DynamoDBClient) BatchGetItem(ctx context.Context, input *dynamodb.BatchGetItemInput) (*dynamodb.BatchGetItemOutput, error) {
	output, err := d.client.BatchGetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get items: %w", err)
	}

	return output, nil
}

// BatchWriteItem writes multiple items to DynamoDB tables
func (d *DynamoDBClient) BatchWriteItem(ctx context.Context, input *dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	output, err := d.client.BatchWriteItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to batch write items: %w", err)
	}

	return output, nil
}

// Common repository errors
var (
	ErrItemNotFound      = errors.New("item not found")
	ErrItemAlreadyExists = errors.New("item already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
)

// BuildKey is a helper function to build DynamoDB attribute value keys
func BuildKey(keyName, keyValue string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		keyName: &types.AttributeValueMemberS{Value: keyValue},
	}
}

// BuildCompositeKey is a helper function to build DynamoDB composite keys
func BuildCompositeKey(partitionKey, partitionValue, sortKey, sortValue string) map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		partitionKey: &types.AttributeValueMemberS{Value: partitionValue},
		sortKey:      &types.AttributeValueMemberS{Value: sortValue},
	}
}

// PaginationToken represents cursor-based pagination
type PaginationToken struct {
	LastEvaluatedKey map[string]types.AttributeValue
	Limit            int32
}

// EncodePaginationToken encodes pagination token to string (implement based on your needs)
func EncodePaginationToken(token *PaginationToken) (string, error) {
	if token == nil || len(token.LastEvaluatedKey) == 0 {
		return "", nil
	}
	
	// For now, return empty string - implement base64 encoding if needed
	// You can implement proper encoding/decoding based on your security requirements
	return "", nil
}

// DecodePaginationToken decodes pagination token from string (implement based on your needs)
func DecodePaginationToken(tokenStr string) (*PaginationToken, error) {
	if tokenStr == "" {
		return nil, nil
	}
	
	// For now, return nil - implement base64 decoding if needed
	// You can implement proper encoding/decoding based on your security requirements
	return nil, nil
}

// IsRetryableError checks if a DynamoDB error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var (
		provisionedThroughputExceededException *types.ProvisionedThroughputExceededException
		requestLimitExceeded                   *types.RequestLimitExceeded
		internalServerError                    *types.InternalServerError
	)

	return errors.As(err, &provisionedThroughputExceededException) ||
		errors.As(err, &requestLimitExceeded) ||
		errors.As(err, &internalServerError)
}