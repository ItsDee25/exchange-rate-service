package jobs

import (
	"context"
	"log"
	"time"

	domain "github.com/ItsDee25/exchange-rate-service/internal/domain/currency"
)

const cleanerFrequency = 24 * time.Hour

type cacheCleaner struct {
	cache domain.IRateCache
}

func NewCacheCleaner(cache domain.IRateCache) *cacheCleaner {
	return &cacheCleaner{cache: cache}
}

func (c *cacheCleaner) Start() {
	log.Println("[CacheCleaner] Starting cache cleaner job")
	ticker := time.NewTicker(cleanerFrequency) // Run every 24 hours
	go func() {
		for {
			select {
			case <-ticker.C:
				c.Run()
			}
		}
	}()
}

func (c *cacheCleaner) Run() {
	log.Printf("Running cache cleaner at time %v", time.Now().Format("2006-01-02 15:04:05"))
	c.cache.ScanAndDeleteExipred(context.Background())
	log.Printf("[CacheCleaner] Cache cleaner completed successfully at time %v", time.Now().Format("2006-01-02 15:04:05"))
}
