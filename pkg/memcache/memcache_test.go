package memcache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 8, true, time.Minute)

	c.Set("a", 123, time.Second)

	v, ok := c.Get("a")
	if !ok {
		t.Fatal("expected value, got miss")
	}
	if v != 123 {
		t.Fatalf("got %d, want 123", v)
	}
}

func TestGet_Miss(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 8, true, time.Minute)

	if _, ok := c.Get("missing"); ok {
		t.Fatal("expected miss")
	}
}

func TestDelete(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 8, true, time.Minute)

	c.Set("a", 1, time.Minute)
	c.Del("a")

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected miss after delete")
	}
}

func TestTTL_Expires(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 8, true, time.Minute)

	c.Set("a", 1, 50*time.Millisecond)
	time.Sleep(80 * time.Millisecond)

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected expired item")
	}
}

func TestTTL_NoExpiration(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 8, true, time.Minute)

	c.Set("a", 1, 0)

	time.Sleep(50 * time.Millisecond)

	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected item to persist")
	}
}

func TestConcurrentSetGet(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 16, true, time.Minute)

	const goroutines = 32
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := "k"
				c.Set(key, id, time.Minute)
				_, _ = c.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentReaders(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 16, true, time.Minute)

	c.Set("a", 42, time.Minute)

	const readers = 64
	var wg sync.WaitGroup
	wg.Add(readers)

	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				v, ok := c.Get("a")
				if !ok || v != 42 {
					t.Errorf("unexpected value %v", v)
				}
			}
		}()
	}

	wg.Wait()
}

func TestCleanupRemovesExpired(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 4, true, 20*time.Millisecond)

	c.Set("a", 1, 10*time.Millisecond)
	time.Sleep(60 * time.Millisecond)

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected expired item to be cleaned up")
	}
}

func TestClear(t *testing.T) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 4, true, time.Minute)

	for i := 0; i < 10; i++ {
		c.Set(string(rune('a'+i)), i, time.Minute)
	}

	n := c.Clear()
	if n != 10 {
		t.Fatalf("cleared %d items, want 10", n)
	}

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected empty cache after clear")
	}
}

func BenchmarkGet_SingleKey(b *testing.B) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 32, true, time.Minute)

	c.Set("a", 1, time.Minute)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.Get("a")
		}
	})
}

func BenchmarkSet_SingleKey(b *testing.B) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 32, true, time.Minute)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set("a", 1, time.Minute)
		}
	})
}

func BenchmarkGet_MultiKey(b *testing.B) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 64, true, time.Minute)

	for i := 0; i < 1000; i++ {
		c.Set("k"+string(rune(i)), i, time.Minute)
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "k" + string(rune(i%1000))
			_, _ = c.Get(key)
			i++
		}
	})
}

func BenchmarkSet_MultiKey(b *testing.B) {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	c := NewCache[int](ctx, 64, true, time.Minute)

	var counter int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := atomic.AddInt64(&counter, 1)
			key := "k" + string(rune(n%1000))
			c.Set(key, int(n), time.Minute)
		}
	})
}
