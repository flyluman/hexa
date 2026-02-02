package ratelimiter

import (
	"context"
	"portfolio/pkg/memcache"
	"sync/atomic"
	"time"
)

type counter struct {
	sec   atomic.Int64
	count atomic.Int64
}

type RateLimiter struct {
	limit int64
	cache *memcache.Cache[*counter]
}

func NewRateLimiter(ctx context.Context, maxPerSecond int64) *RateLimiter {
	return &RateLimiter{
		limit: maxPerSecond,
		cache: memcache.NewCache[*counter](ctx, 32, true, 10*time.Minute),
	}
}

func (r *RateLimiter) Allow(key string) bool {
	nowSec := time.Now().Unix()

	val, ok := r.cache.Get(key)
	if !ok {
		c := &counter{}
		c.sec.Store(nowSec)
		c.count.Store(1)

		r.cache.Set(key, c, 10*time.Minute)
		return true
	}

	return r.allowWithCounter(val, nowSec)
}

func (r *RateLimiter) allowWithCounter(c *counter, nowSec int64) bool {
	prevSec := c.sec.Load()

	if prevSec != nowSec {
		if c.sec.CompareAndSwap(prevSec, nowSec) {
			c.count.Store(1)
			return true
		}
		return r.allowWithCounter(c, nowSec)
	}

	n := c.count.Add(1)
	return n <= r.limit
}
