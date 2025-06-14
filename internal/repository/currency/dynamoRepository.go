package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ItsDee25/exchange-rate-service/pkg/constants"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	PartitionKey = "pk"
	SortKey      = "sk"
	TTL          = "ttl"
	Rate         = "rate"
	UpdatedAt    = "updated_at"
)

const TTLDuration = 90 * 24 * time.Hour // 90 days

type currencyDynamoRepository struct {
	client    *dynamodb.Client
	cache     *rateCache
	tableName string
}

type rateCache struct {
	cache sync.Map
}

func newRateCache() *rateCache {
	return &rateCache{}
}

func (c *rateCache) get(key string) (float64, bool) {
	val, ok := c.cache.Load(key)
	if !ok {
		return 0, false
	}
	rate, ok := val.(float64)
	return rate, ok
}

func (c *rateCache) set(key string, rate float64) {
	c.cache.Store(key, rate)
}

func NewDynamoRepository(ctx context.Context, client *dynamodb.Client) (*currencyDynamoRepository, error) {
	return &currencyDynamoRepository{
		client:    client,
		tableName: "currency_rates",
		cache:     newRateCache(),
	}, nil
}

func getPartitionKey(from, to string) string {
	return fmt.Sprintf("%s#%s", from, to)
}

func getCacheKey(from, to, date string) string {
	return fmt.Sprintf("%s#%s#%s", from, to, date)
}

func getDynamoItemTTL(date string) (int64, error) {
	rateDate, err := time.Parse(constants.DateLayout, date)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %w", err)
	}
	return rateDate.Add(TTLDuration).Unix(), nil
}

func (r *currencyDynamoRepository) GetRate(ctx context.Context, from, to, date string) (float64, error) {
	pk := getPartitionKey(from, to)
	cacheKey := getCacheKey(from, to, date)
	// Check local cache first
	if val, ok := r.cache.get(cacheKey); ok {
		return val, nil
	}

	// Fetch from DynamoDB
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			PartitionKey: &types.AttributeValueMemberS{Value: pk},
			SortKey:      &types.AttributeValueMemberS{Value: date},
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
	r.cache.set(cacheKey, item.Rate)

	return item.Rate, nil
}

func (r *currencyDynamoRepository) SaveRate(ctx context.Context, from, to, date string, rate float64) error {
	pk := getPartitionKey(from, to)
	cacheKey := getCacheKey(from, to, date)

	ttl, err := getDynamoItemTTL(date)

	if err != nil {
		return err
	}

	item := map[string]interface{}{
		PartitionKey: pk,
		SortKey:      date,
		Rate:         rate,
		UpdatedAt:    time.Now().Unix(),
		TTL:          ttl,
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
		return fmt.Errorf("error saving rate: %w", err)
	}

	// Update cache
	r.cache.set(cacheKey, rate)
	return nil
}
