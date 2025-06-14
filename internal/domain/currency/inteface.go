package domain

import "context"

type ICurrencyUsecase interface {
	GetConvertedCurrency(ctx context.Context, from, to, date string, amount float64) (float64, error)
	GetExchangeRate(ctx context.Context, from, to, date string) (float64, error)
}

type ICurrencyRepository interface {
	GetRate(ctx context.Context, from, to, date string) (float64, error)
	SaveRate(ctx context.Context, from, to, date string, rate float64) error
}

type IRateFetcher interface {
	FetchRate(ctx context.Context, from, to, date string) (float64, error)
}
