package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	domain "github.com/ItsDee25/exchange-rate-service/internal/domain/currency"
	"github.com/ItsDee25/exchange-rate-service/pkg/constants"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	ttlDuration     = 90 * 24 * time.Hour
)

type currencyDynamoRepository struct {
	client      *dynamodb.Client
	cache       domain.IRateCache
	tableName   string
	rateFetcher domain.IRateFetcher
}

func NewDynamoRepository(client *dynamodb.Client, rateFetcher domain.IRateFetcher, cache domain.IRateCache) (*currencyDynamoRepository, error) {
	return &currencyDynamoRepository{
		client:      client,
		tableName:   constants.TableName,
		cache:       cache,
		rateFetcher: rateFetcher,
	}, nil
}

func getPartitionKey(from, to string) string {
	return fmt.Sprintf("%s#%s", from, to)
}

func getFromAndToFromPartitionKey(pk string) (string, string, error) {
	parts := strings.Split(pk, "#")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid partition key format: %s", pk)
	}
	return parts[0], parts[1], nil
}

func getCacheKey(from, to, date string) string {
	return fmt.Sprintf("%s#%s#%s", from, to, date)
}

func getDynamoItemTTL(date string) (int64, error) {
	rateDate, err := time.Parse(constants.DateLayout, date)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %w", err)
	}
	return rateDate.Add(ttlDuration).Unix(), nil
}

func (r *currencyDynamoRepository) GetDataFromDB(ctx context.Context, from, to, date string) (float64, error) {
	pk := getPartitionKey(from, to)
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			constants.PartitionKey: &types.AttributeValueMemberS{Value: pk},
			constants.SortKey:      &types.AttributeValueMemberS{Value: date},
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

	return item.Rate, nil
}

func (r *currencyDynamoRepository) SaveRateInDB(ctx context.Context, from, to, date string, rate float64) error {
	pk := getPartitionKey(from, to)
	ttl, err := getDynamoItemTTL(date)

	if err != nil {
		return err
	}

	item := map[string]interface{}{
		constants.PartitionKey: pk,
		constants.SortKey:      date,
		constants.Rate:         rate,
		constants.UpdatedAt:    time.Now().Unix(),
		constants.TTL:          ttl,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	})
	return err
}

func (r *currencyDynamoRepository) GetRate(ctx context.Context, from, to, date string) (float64, error) {
	cacheKey := getCacheKey(from, to, date)
	// Check local cache first
	if val, ok := r.cache.Get(ctx, cacheKey); ok {
		return val, nil
	}

	// Fetch from DynamoDB
	rate, err := r.GetDataFromDB(ctx, from, to, date)

	if err != nil {
		rate, err = r.rateFetcher.FetchRate(ctx, from, to, date)
		if err != nil {
			return 0, err
		}
		r.cache.Set(ctx, cacheKey, rate)

		go func() {
			err := r.SaveRateInDB(ctx, from, to, date, rate)
			if err != nil {
				// TODO: log the error
			}
		}()
		return rate, nil
	}
	// Update local cache
	r.cache.Set(ctx, cacheKey, rate)

	return rate, nil
}

func (r *currencyDynamoRepository) SaveRate(ctx context.Context, from, to, date string, rate float64) error {
	cacheKey := getCacheKey(from, to, date)
	err := r.SaveRateInDB(ctx, from, to, date, rate)
	if err != nil {
		return fmt.Errorf("error saving rate: %w", err)
	}

	r.cache.Set(ctx, cacheKey, rate)

	return nil
}

func (r *currencyDynamoRepository) BatchGetFromDB(ctx context.Context, req []domain.RateKeyRequest) ([]domain.RateKey, error) {
	if len(req) == 0 {
		return nil, nil
	}

	keys := make([]map[string]types.AttributeValue, 0, len(req))
	for _, k := range req {
		pk := getPartitionKey(k.From, k.To)
		keys = append(keys, map[string]types.AttributeValue{
			constants.PartitionKey: &types.AttributeValueMemberS{Value: pk},
			constants.SortKey:      &types.AttributeValueMemberS{Value: k.Date},
		})
	}

	out, err := r.client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			r.tableName: {
				Keys: keys,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("batch get failed: %w", err)
	}

	result := make([]domain.RateKey, 0, len(out.Responses[r.tableName]))
	for _, item := range out.Responses[r.tableName] {
		var decoded struct {
			PK   string  `dynamodbav:"pk"`
			SK   string  `dynamodbav:"sk"`
			Rate float64 `dynamodbav:"rate"`
		}
		if err := attributevalue.UnmarshalMap(item, &decoded); err != nil {
			log.Printf("unmarshal failed: %v", err)
			continue
		}
		from, to, err := getFromAndToFromPartitionKey(decoded.PK)
		if err != nil {
			log.Printf("get from/to failed: %v", err)
			continue
		}
		result = append(result, domain.RateKey{
			RateKeyRequest: domain.RateKeyRequest{
				From: from,
				To:   to,
				Date: decoded.SK,
			},
			Rate: decoded.Rate,
		})
	}

	return result, nil
}

func (r *currencyDynamoRepository) BatchUpdateDB(ctx context.Context, rates []domain.RateKey) error {
	if len(rates) == 0 {
		return nil
	}

	writeRequests := make([]types.WriteRequest, 0, len(rates))
	for _, rate := range rates {
		pk := getPartitionKey(rate.From, rate.To)
		item := map[string]interface{}{
			constants.PartitionKey: pk,
			constants.SortKey:      rate.Date,
			constants.Rate:         rate.Rate,
			constants.UpdatedAt:    time.Now().Unix(),
		}
		ttl, err := getDynamoItemTTL(rate.Date)
		if err == nil {
			item[constants.TTL] = ttl
		}

		av, err := attributevalue.MarshalMap(item)
		if err != nil {
			log.Printf("marshal error for %v: %v", rate, err)
			continue
		}
		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: av},
		})
	}

	batch := writeRequests
	for len(batch) > 0 {
		size := 25
		if len(batch) < size {
			size = len(batch)
		}
		chunk := batch[:size]
		batch = batch[size:]

		_, err := r.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				r.tableName: chunk,
			},
		})
		if err != nil {
			return fmt.Errorf("batch write failed: %w", err)
		}
	}

	return nil
}

func (r *currencyDynamoRepository) BatchUpdateCache(ctx context.Context, rates []domain.RateKey) error {
	for _, rate := range rates {
		cacheKey := getCacheKey(rate.From, rate.To, rate.Date)
		r.cache.Set(ctx, cacheKey, rate.Rate)
	}
	return nil
}

func (r *currencyDynamoRepository) BatchDeleteFromCache(ctx context.Context, req []domain.RateKeyRequest) error {
	for _, k := range req {
		cacheKey := getCacheKey(k.From, k.To, k.Date)
		r.cache.Delete(ctx, cacheKey)
	}
	return nil
}
