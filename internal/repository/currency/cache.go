package repository

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ItsDee25/exchange-rate-service/pkg/constants"
)

type RateCache struct {
	cache sync.Map
}

func NewRateCache() *RateCache {
	return &RateCache{}
}

func (c *RateCache) Get(ctx context.Context, key string) (float64, bool) {
	val, ok := c.cache.Load(key)
	if !ok {
		return 0, false
	}
	rate, ok := val.(float64)
	return rate, ok
}

func (c *RateCache) Set(ctx context.Context, key string, rate float64) {
	c.cache.Store(key, rate)
}

func (c *RateCache) Delete(ctx context.Context, key string) {
	c.cache.Delete(key)
}

func (c *RateCache) ScanAndDeleteExipred(ctx context.Context) {
	now := time.Now()
	deleted := 0
	c.cache.Range(func(key, value any) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}
		params := strings.Split(keyStr, "#")
		if len(params) < 3 {
			return true
		}
		rateDate, err := time.Parse(constants.DateLayout, params[2])
		if err != nil {
			return true
		}
		if now.Sub(rateDate) > ttlDuration {
			c.cache.Delete(key)
			deleted++
		}
		return true
	})
}
