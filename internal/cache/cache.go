package cache

import (
	"context"
)

type Cache interface {
	WriteToCache(ctx context.Context, ticket *CacheRecord) error
	DeleteFromCache(ctx context.Context, ticket *CacheRecord) error
	UpdateCache(ctx context.Context, ticket *CacheRecord) error
	GetFromCacheByKey(ctx context.Context, key string) (*CacheRecord, error)
	GetFromCacheByCustomerID(ctx context.Context, customerInternalID string) (*CacheRecord, error)
	GetStatusFromCache(ctx context.Context, customerInternalID string) (string, error)
	GetSourceFromCache(ctx context.Context, customerInternalID string) (string, error)
	GetProcessingSystemFromCache(ctx context.Context, customerInternalID string) (string, error)
	GetAllKeysFromCache(ctx context.Context) ([]string, error)
}
