package cache

import (
	"context"
	"portfolio/internal/core/domain"
	"portfolio/pkg/memcache"
)

type MetaIPCache struct {
	bucket *memcache.Cache[*domain.MetaIP]
}

func NewMetaIPCache(ctx context.Context, shards int) *MetaIPCache {
	return &MetaIPCache{
		bucket: memcache.NewCache[*domain.MetaIP](ctx, shards, false, -1),
	}
}

func (c *MetaIPCache) Get(key string) (*domain.MetaIP, bool) {
	return c.bucket.Get(key)
}

func (c *MetaIPCache) Set(key string, val *domain.MetaIP) {
	c.bucket.Set(key, val, -1)
}

func (c *MetaIPCache) Del(key string) {
	c.bucket.Del(key)
}

func (c *MetaIPCache) Query() int {
	return c.bucket.Query()
}

func (c *MetaIPCache) Clear() int {
	return c.bucket.Clear()
}
