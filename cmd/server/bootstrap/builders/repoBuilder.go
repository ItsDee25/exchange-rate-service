package builders

import (
	infra "github.com/ItsDee25/exchange-rate-service/infra/ratefetcher"
	repository "github.com/ItsDee25/exchange-rate-service/internal/repository/currency"
)

type repositories struct {
	CurrencyDynamoRepository *repository.CurrencyDynamoRepository
	CurrecyCache *repository.RateCache
	DynamoLocker *infra.DynamoLocker
}

func NewRepositories() *repositories {
	return &repositories{}
}

func (r *repositories) WithCurrencyRepository(repo *repository.CurrencyDynamoRepository) *repositories {
	r.CurrencyDynamoRepository = repo
	return r
}

func (r *repositories) WithCurrencyCache(c *repository.RateCache) *repositories {
	r.CurrecyCache = c
	return r
}

func (r *repositories) WithDynamoLocker(l *infra.DynamoLocker) *repositories {
	r.DynamoLocker = l
	return r
}