package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

const (
	TIMEOUT = time.Millisecond * 500
)

type apiCache struct {
	pool *redis.Pool
	ttl  int64
	lg   *zap.Logger
}

func NewRedisCache(pool *redis.Pool, ttl int64, lg *zap.Logger) Cache {
	return &apiCache{pool: pool, ttl: ttl, lg: lg}
}

func InitCache(url string) *redis.Pool {
	cache := &redis.Pool{
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return redis.DialURL(url)
		},
	}
	return cache
}

func (a *apiCache) WriteToCache(ctx context.Context, record *CacheRecord) error {
	conn, err := a.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("WriteToCache: %w", err)
	}
	defer conn.Close()
	//Ключ для Redis в формате CustomerInternalID:record.CustomerInternalID
	key := fmt.Sprintf("CustomerInternalID:%s", record.CustomerInternalID)
	_, err = redis.DoWithTimeout(conn, TIMEOUT, "HSET", redis.Args{}.Add(key).AddFlat(record)...)
	if err != nil {
		return fmt.Errorf("WriteToCache: %w", err)
	}
	//TTL записи в Redis
	_, err = redis.DoWithTimeout(conn, TIMEOUT, "EXPIRE", key, a.ttl)
	if err != nil {
		return fmt.Errorf("WriteToCache: %w", err)
	}
	return nil
}

func (a *apiCache) DeleteFromCache(ctx context.Context, record *CacheRecord) error {
	conn, err := a.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("DeleteFromCache: %w", err)
	}
	defer conn.Close()
	key := fmt.Sprintf("CustomerInternalID:%s", record.CustomerInternalID)
	_, err = redis.DoWithTimeout(conn, TIMEOUT, "HDEL", key,
		"status",
		"source",
		"id_channel_operator",
		"id_channel_operator_for_billing",
		"operator_tt_id",
		"created",
		"modified",
	)
	if err != nil {
		return fmt.Errorf("DeleteFromCache: %w", err)
	}
	return nil
}

func (a *apiCache) UpdateCache(ctx context.Context, record *CacheRecord) error {
	panic("implement me")
}

func (a *apiCache) GetFromCacheByCustomerID(ctx context.Context, customerInternalID string) (*CacheRecord, error) {
	var record = CacheRecord{}
	conn, err := a.pool.GetContext(ctx)
	if err != nil {
		return &record, fmt.Errorf("GetFromCacheByKey: %w", err)
	}
	defer conn.Close()
	key := fmt.Sprintf("CustomerInternalID:%s", customerInternalID)
	redisResponce, err := redis.DoWithTimeout(conn, TIMEOUT, "HGETALL", key)
	if err != nil {
		return &record, fmt.Errorf("GetFromCacheByKey: %w", err)
	}
	values, err := redis.Values(redisResponce, nil)
	if err != nil {
		return &record, fmt.Errorf("GetFromCacheByKey: %w", err)
	}
	err = redis.ScanStruct(values, &record)
	if err != nil {
		return &record, fmt.Errorf("GetFromCacheByKey: %w", err)
	}
	return &record, nil
}

func (a *apiCache) GetFromCacheByKey(ctx context.Context, key string) (*CacheRecord, error) {
	var record = CacheRecord{}
	conn, err := a.pool.GetContext(ctx)
	if err != nil {
		return &record, fmt.Errorf("GetFromCacheByKey: %w", err)
	}
	defer conn.Close()
	redisResponce, err := redis.DoWithTimeout(conn, TIMEOUT, "HGETALL", key)
	if err != nil {
		return &record, fmt.Errorf("GetFromCacheByKey: %w", err)
	}
	values, err := redis.Values(redisResponce, nil)
	if err != nil {
		return &record, fmt.Errorf("GetFromCacheByKey: %w", err)
	}
	err = redis.ScanStruct(values, &record)
	if err != nil {
		return &record, fmt.Errorf("GetFromCacheByKey: %w", err)
	}
	return &record, nil
}

func (a *apiCache) GetStatusFromCache(ctx context.Context, customerInternalID string) (string, error) {
	var status string
	conn, err := a.pool.GetContext(ctx)
	if err != nil {
		return status, fmt.Errorf("GetStatusFromCache: %w", err)
	}
	defer conn.Close()
	key := fmt.Sprintf("CustomerInternalID:%s", customerInternalID)
	redisResponce, err := redis.DoWithTimeout(conn, TIMEOUT, "HGET", key, "Status")
	if err != nil {
		return status, fmt.Errorf("GetStatusFromCache: %w", err)
	}
	status, err = redis.String(redisResponce, nil)
	if err != nil {
		return status, fmt.Errorf("GetStatusFromCache: %w", err)
	}
	return status, nil
}

func (a *apiCache) GetSourceFromCache(ctx context.Context, customerInternalID string) (string, error) {
	var source string
	conn, err := a.pool.GetContext(ctx)
	if err != nil {
		return source, fmt.Errorf("GetSourceFromCache: %w", err)
	}
	defer conn.Close()
	key := fmt.Sprintf("CustomerInternalID:%s", customerInternalID)
	redisResponce, err := redis.DoWithTimeout(conn, TIMEOUT, "HGET", key, "Source")
	if err != nil {
		return source, fmt.Errorf("GetSourceFromCache: %w", err)
	}
	source, err = redis.String(redisResponce, nil)
	if err != nil {
		return source, fmt.Errorf("GetSourceFromCache: %w", err)
	}
	return source, nil
}

func (a *apiCache) GetProcessingSystemFromCache(ctx context.Context, customerInternalID string) (string, error) {
	var idChannelOperatorForBilling string
	conn, err := a.pool.GetContext(ctx)
	if err != nil {
		return idChannelOperatorForBilling, fmt.Errorf("GetSourceFromCache: %w", err)
	}
	defer conn.Close()
	key := fmt.Sprintf("CustomerInternalID:%s", customerInternalID)
	redisResponce, err := redis.DoWithTimeout(conn, TIMEOUT, "HGET", key, "IDChannelOperatorForBilling")
	if err != nil {
		return idChannelOperatorForBilling, fmt.Errorf("GetSourceFromCache: %w", err)
	}
	idChannelOperatorForBilling, err = redis.String(redisResponce, nil)
	if err != nil {
		return idChannelOperatorForBilling, fmt.Errorf("GetSourceFromCache: %w", err)
	}
	return idChannelOperatorForBilling, nil
}

func (a *apiCache) GetAllKeysFromCache(ctx context.Context) ([]string, error) {
	var keys []string
	conn, err := a.pool.GetContext(ctx)
	if err != nil {
		return keys, fmt.Errorf("cache.GetKeysFromCache: %w", err)
	}
	defer conn.Close()
	var cursor = 0
	var counter = 10000
	data, err := redis.Values(redis.DoWithTimeout(conn, TIMEOUT, "SCAN", cursor, "COUNT", counter))
	if err != nil {
		return keys, fmt.Errorf("cache.GetKeysFromCache: %w", err)
	}
	cursor, _ = redis.Int(data[0], nil)
	if cursor != 0 {
		a.lg.Info("Over 10 000 records in cache")
	}
	keys, err = redis.Strings(data[1], nil)
	if err != nil {
		return keys, fmt.Errorf("cache.GetKeysFromCache: %w", err)
	}
	return keys, nil
}
