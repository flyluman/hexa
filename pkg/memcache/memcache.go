package memcache

import (
	"context"
	"sync"
	"time"
)

type item[T any] struct {
	v   T
	exp int64
}

type shard[T any] struct {
	mu sync.RWMutex
	m  map[string]*item[T]
}

type Cache[T any] struct {
	shards []shard[T]
	mask   uint32
}

func NewCache[T any](ctx context.Context, shards int, cleanEnable bool, cleanInterval time.Duration) *Cache[T] {
	if shards <= 0 {
		shards = 128
	}

	n := 1
	for n < shards {
		n <<= 1
	}

	c := &Cache[T]{
		shards: make([]shard[T], n),
		mask:   uint32(n - 1),
	}

	for i := range c.shards {
		c.shards[i].m = make(map[string]*item[T])
	}

	if cleanEnable {
		go c.cleanup(ctx, cleanInterval)
	}

	return c
}

func (c *Cache[T]) getShard(k string) *shard[T] {
	var h uint32 = 2166136261
	for i := 0; i < len(k); i++ {
		h ^= uint32(k[i])
		h *= 16777619
	}
	return &c.shards[h&c.mask]
}

func (c *Cache[T]) Set(key string, val T, ttl time.Duration) {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	newIt := &item[T]{v: val, exp: exp}
	s := c.getShard(key)

	s.mu.Lock()
	s.m[key] = newIt
	s.mu.Unlock()
}

func (c *Cache[T]) Get(key string) (T, bool) {
	s := c.getShard(key)
	s.mu.RLock()
	it, ok := s.m[key]
	s.mu.RUnlock()

	if !ok {
		var zero T
		return zero, false
	}

	if it.exp > 0 && time.Now().UnixNano() > it.exp {
		var zero T
		return zero, false
	}

	return it.v, true
}

func (c *Cache[T]) Del(key string) {
	s := c.getShard(key)
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()
}

func (c *Cache[T]) Clear() int {
	var count int
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.Lock()
		count += len(s.m)
		s.m = make(map[string]*item[T])
		s.mu.Unlock()
	}
	return count
}

func (c *Cache[T]) Query() int {
	var count int
	for i := range c.shards {
		s := &c.shards[i]
		s.mu.RLock()
		count += len(s.m)
		s.mu.RUnlock()
	}
	return count
}

func (c *Cache[T]) cleanup(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			now := time.Now().UnixNano()
			for i := range c.shards {
				s := &c.shards[i]
				s.mu.Lock()
				for k, v := range s.m {
					if v.exp > 0 && now > v.exp {
						delete(s.m, k)
					}
				}
				s.mu.Unlock()
			}
		}
	}
}
