package repository

import (
	"strings"
	"sync"
	"time"

	"github.com/ItsDee25/exchange-rate-service/pkg/constants"
)

type rateCache struct {
	cache sync.Map
}

func NewRateCache() *rateCache {
	return &rateCache{}
}

func (c *rateCache) Get(key string) (float64, bool) {
	val, ok := c.cache.Load(key)
	if !ok {
		return 0, false
	}
	rate, ok := val.(float64)
	return rate, ok
}

func (c *rateCache) Set(key string, rate float64) {
	c.cache.Store(key, rate)
}

func (c *rateCache) Delete(key string) {
	c.cache.Delete(key)
}

func (c *rateCache) ScanAndDeleteExipred() {
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
