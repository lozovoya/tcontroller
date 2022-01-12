package timer

import (
	"Smoozzy/TicketSystemController/internal/cache"
	"Smoozzy/TicketSystemController/internal/model"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type timer struct {
	ctx   context.Context
	TTL   time.Duration
	cache cache.Cache
	lg    *zap.Logger
}

func NewTimer(ctx context.Context, TTL time.Duration, cache cache.Cache) *timer {
	return &timer{ctx: ctx, TTL: TTL, cache: cache}
}

func (t *timer) FindExpired() (*[]cache.CacheRecord, error) {
	expiredRecords := make([]cache.CacheRecord, 0)
	keys, err := t.cache.GetAllKeysFromCache(t.ctx)

	if err != nil {
		return &expiredRecords, fmt.Errorf("timer.findExpired: %w", err)
	}
	for _, k := range keys {
		record, err := t.cache.GetFromCacheByKey(t.ctx, k)
		log.Printf("checking key %v", record)
		if err != nil {
			return &expiredRecords, fmt.Errorf("timer.findExpired: %w", err)
		}
		if (record.Status == model.Creating) || (record.Status == model.Error) {
			log.Printf("found right ticket %v", keys)
			timeNow := time.Now()
			i, err := strconv.ParseInt(record.Modified, 10, 64)
			if err != nil {
				return &expiredRecords, fmt.Errorf("timer.findExpired: %w", err)
			}
			modifiedTime := time.Unix(i, 0)
			log.Printf("time: %v - %v", timeNow, modifiedTime.Add(time.Minute*30))
			if timeNow.After(modifiedTime.Add(time.Minute * 30)) {
				log.Printf("succesful time comparison %s", record.Modified)
			}
		}
	}
	return &expiredRecords, nil
}
