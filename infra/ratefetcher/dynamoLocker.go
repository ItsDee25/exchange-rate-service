package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ItsDee25/exchange-rate-service/pkg/constants"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoLocker struct {
	tableName string
	client    *dynamodb.Client
}

func NewDynamoLocker(client *dynamodb.Client) *DynamoLocker {
	return &DynamoLocker{
		tableName: constants.TableName,
		client:    client,
	}
}

func (l *DynamoLocker) AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	expiresAt := time.Now().Add(ttl).Unix()

	item := map[string]types.AttributeValue{
		constants.PartitionKey: &types.AttributeValueMemberS{Value: key},
		constants.ExpiresAt:    &types.AttributeValueMemberN{Value: string(expiresAt)},
	}

	condExpr := "attribute_not_exists(pk) OR expires_at < :now"

	_, err := l.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(l.tableName),
		Item:                item,
		ConditionExpression: aws.String(condExpr),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":now": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
		},
	})
	if err != nil {
		var cce *types.ConditionalCheckFailedException
		if errors.As(err, &cce) {
			return false, nil // Lock not acquired
		}
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return true, nil
}
