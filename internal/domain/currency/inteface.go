package domain

import (
	"context"
	"time"
)

type ICurrencyUsecase interface {
	GetConvertedCurrency(ctx context.Context, from, to, date string, amount float64) (float64, error)
	GetExchangeRate(ctx context.Context, from, to, date string) (float64, error)
}

type ICurrencyRepository interface {
	GetRate(ctx context.Context, from, to, date string) (float64, error)
}

type IRefresherRepository interface {
	ICurrencyRepository
	BatchGetFromDB(ctx context.Context, req []RateKeyRequest) ([]RateKey, error)
	BatchUpdateDB(ctx context.Context, req []RateKey) error
	BatchUpdateCache(ctx context.Context, req []RateKey) error
}

type IRateCache interface {
	Get(ctx context.Context, key string) (float64, bool)
	Set(ctx context.Context, key string, value float64) error
	Delete(ctx context.Context, key string) error
	ScanAndDeleteExipred(ctx context.Context)
}

type ILocker interface {
	AcquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

type IRateFetcher interface {
	FetchRate(ctx context.Context, from, to, date string) (float64, error)
}
