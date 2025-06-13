package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoRepository struct {
	client    *dynamodb.Client
	tableName string
	cache     sync.Map // in-memory: map[string]float64
}

func NewDynamoRepository(ctx context.Context, tableName string) (*DynamoRepository, error) {
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if ep := os.Getenv("DYNAMO_ENDPOINT"); ep != "" {
			return aws.Endpoint{URL: ep}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("no endpoint set")
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-west-2"),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		return nil, err
	}

	return &DynamoRepository{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

func (r *DynamoRepository) GetRate(ctx context.Context, from, to, date string) (float64, error) {
	key := fmt.Sprintf("%s#%s#%s", from, to, date)

	// Check local cache first
	if val, ok := r.cache.Load(key); ok {
		if rate, ok := val.(float64); ok {
			return rate, nil
		}
	}

	// Fetch from DynamoDB
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"rate_key": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return 0, err
	}
	if result.Item == nil {
		return 0, errors.New("rate not found")
	}

	var item struct {
		Rate float64 `dynamodbav:"rate"`
	}

	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return 0, fmt.Errorf("unmarshal error: %w", err)
	}

	// Update local cache
	r.cache.Store(key, item.Rate)

	return item.Rate, nil
}

func (r *DynamoRepository) SaveRate(ctx context.Context, from, to, date string, rate float64) error {
	key := fmt.Sprintf("%s#%s#%s", from, to, date)

	rateDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}
	ttl := rateDate.Add(90 * 24 * time.Hour).Unix()

	item := map[string]interface{}{
		"rate_key":   key,
		"rate":       rate,
		"updated_at": time.Now().Unix(),
		"ttl":        ttl,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	if err != nil {
		return err
	}

	// Update cache
	r.cache.Store(key, rate)
	return nil
}
