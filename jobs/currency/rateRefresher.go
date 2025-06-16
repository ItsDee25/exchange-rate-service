package jobs

import (
	"context"
	"log"
	"sync"
	"time"

	domain "github.com/ItsDee25/exchange-rate-service/internal/domain/currency"
)

const (
	lockId        = "rate_refresher_lock"
	lockTTL       = 25 * time.Minute
	refresherFreq = 30 * time.Minute
)

type RateRefresher struct {
	repo          domain.IRefresherRepository
	fetcher       domain.IRateFetcher
	locker        domain.ILocker
	currencyPairs [][2]string
}

func NewRateRefresher(repo domain.IRefresherRepository, fetcher domain.IRateFetcher, locker domain.ILocker, pairs [][2]string) *RateRefresher {
	return &RateRefresher{
		repo:          repo,
		fetcher:       fetcher,
		locker:        locker,
		currencyPairs: pairs,
	}
}

func (r *RateRefresher) Start() {
	log.Println("[RateRefresher] Starting rate refresher job")
	ticker := time.NewTicker(refresherFreq)
	go func() {
		for {
			select {
			case <-ticker.C:
				r.Run()
			}
		}
	}()
}

func (r *RateRefresher) Run() {
	log.Printf("Running rate refresher for pairs: %v\n at time %v", r.currencyPairs, time.Now().Format("2006-01-02 15:04:05"))
	ctx := context.Background()
	today := time.Now().Format("2006-01-02")
	locked, err := r.locker.AcquireLock(ctx, lockId, lockTTL)
	if err != nil {
		log.Printf("Failed to acquire lock: %v", err)
		return
	}
	if locked {
		mu := sync.Mutex{}
		rateKeys := make([]domain.RateKey, 0, len(r.currencyPairs))
		wg := sync.WaitGroup{}
		for _, pair := range r.currencyPairs {
			from, to := pair[0], pair[1]
			wg.Add(1)
			go func(from, to string) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Recovered from panic while refreshing rate for %s to %s: %v", from, to, r)
					}
				}()
				rate, err := r.fetcher.FetchRate(ctx, from, to, today)
				if err != nil {
					log.Printf("Failed to fetch rate for %s to %s: %v", from, to, err)
					return
				}
				mu.Lock()
				defer mu.Unlock()
				rateKeys = append(rateKeys, domain.RateKey{
					RateKeyRequest: domain.RateKeyRequest{
						From: from,
						To:   to,
						Date: today,
					},
					Rate: rate,
				})
			}(from, to)
		}
		wg.Wait()
		if len(rateKeys) > 0 {
			err := r.repo.BatchUpdateDB(ctx, rateKeys)
			if err != nil {
				log.Printf("Failed to update rates in batch: %v for keys %v", err, rateKeys)
			}
		}
		return
	}

	req := make([]domain.RateKeyRequest, len(r.currencyPairs))
	for i, pair := range r.currencyPairs {
		req[i] = domain.RateKeyRequest{
			From: pair[0],
			To:   pair[1],
			Date: today,
		}
	}
	rateKeys, err := r.repo.BatchGetFromDB(ctx, req)
	if err != nil {
		log.Printf("Failed to get rates from DB: %v for keys %v", err, req)
		return
	}
	if len(rateKeys) == 0 {
		log.Printf("No rates found in DB for pairs: %v", r.currencyPairs)
		return
	}
	err = r.repo.BatchUpdateCache(ctx, rateKeys)
	if err != nil {
		log.Printf("Failed to update cache in batch: %v for keys %v", err, rateKeys)
	}
	log.Printf("Rate refresher completed successfully for pairs: %v at time %v", r.currencyPairs, time.Now().Format("2006-01-02 15:04:05"))
}
