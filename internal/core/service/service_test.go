package service

import (
	"context"
	"errors"
	"portfolio/internal/core/domain"
	"sync"
	"testing"
	"time"
)

type repoStub struct {
	mu sync.Mutex

	saveMessageCalls int
	savedMessage     *domain.Message
	saveMessageErr   error

	saveMetaCalls int
	lastMetaReq   *domain.MetaReq
	lastMetaIP    *domain.MetaIP
	saveMetaErr   error
	saveMetaCh    chan struct{}

	queryCalls int
	queryLogs  []*domain.Log
	queryErr   error
}

func (r *repoStub) SaveMessage(_ context.Context, msg *domain.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saveMessageCalls++
	r.savedMessage = msg
	return r.saveMessageErr
}

func (r *repoStub) SaveMeta(_ context.Context, metaReq *domain.MetaReq, metaIP *domain.MetaIP) error {
	r.mu.Lock()
	r.saveMetaCalls++
	r.lastMetaReq = metaReq
	r.lastMetaIP = metaIP
	ch := r.saveMetaCh
	err := r.saveMetaErr
	r.mu.Unlock()

	if ch != nil {
		select {
		case ch <- struct{}{}:
		default:
		}
	}

	return err
}

func (r *repoStub) QueryLog(_ context.Context) ([]*domain.Log, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.queryCalls++
	return r.queryLogs, r.queryErr
}

type loggerStub struct{}

func (loggerStub) Info(string, ...any)  {}
func (loggerStub) Error(string, ...any) {}
func (loggerStub) Debug(string, ...any) {}

type ipFetcherStub struct {
	mu   sync.Mutex
	meta *domain.MetaIP
	err  error

	calls int
}

func (f *ipFetcherStub) FetchIPInfo(_ context.Context, _ string) (*domain.MetaIP, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.meta, f.err
}

type cacheStub struct {
	mu   sync.Mutex
	data map[string]*domain.MetaIP
}

func newCacheStub() *cacheStub {
	return &cacheStub{data: make(map[string]*domain.MetaIP)}
}

func (c *cacheStub) Get(key string) (*domain.MetaIP, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *cacheStub) Set(key string, val *domain.MetaIP) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = val
}

func (c *cacheStub) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *cacheStub) Query() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.data)
}

func (c *cacheStub) Clear() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := len(c.data)
	c.data = make(map[string]*domain.MetaIP)
	return n
}

func newTestService(repo *repoStub, fetcher *ipFetcherStub, cache *cacheStub, queryPass string) *Service {
	return &Service{
		repo:        repo,
		logger:      loggerStub{},
		ipfetcher:   fetcher,
		metaIpCache: cache,
		queryPass:   queryPass,
	}
}

func TestPostMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &repoStub{}
		svc := newTestService(repo, &ipFetcherStub{}, newCacheStub(), "secret")

		msg := &domain.Message{Name: "n", Email: "e", Text: "t"}
		if err := svc.PostMessage(context.Background(), msg); err != nil {
			t.Fatalf("PostMessage() error = %v, want nil", err)
		}

		repo.mu.Lock()
		defer repo.mu.Unlock()
		if repo.saveMessageCalls != 1 {
			t.Fatalf("SaveMessage calls = %d, want 1", repo.saveMessageCalls)
		}
		if repo.savedMessage != msg {
			t.Fatal("SaveMessage received unexpected message pointer")
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &repoStub{saveMessageErr: errors.New("db down")}
		svc := newTestService(repo, &ipFetcherStub{}, newCacheStub(), "secret")

		err := svc.PostMessage(context.Background(), &domain.Message{})
		if err == nil {
			t.Fatal("PostMessage() error = nil, want non-nil")
		}
	})
}

func TestPostQuery(t *testing.T) {
	t.Run("incorrect password", func(t *testing.T) {
		repo := &repoStub{}
		svc := newTestService(repo, &ipFetcherStub{}, newCacheStub(), "secret")

		_, err := svc.PostQuery(context.Background(), &domain.Query{Pass: "wrong"})
		if !errors.Is(err, domain.ErrIncorrectPassword) {
			t.Fatalf("PostQuery() error = %v, want ErrIncorrectPassword", err)
		}

		repo.mu.Lock()
		defer repo.mu.Unlock()
		if repo.queryCalls != 0 {
			t.Fatalf("QueryLog calls = %d, want 0", repo.queryCalls)
		}
	})

	t.Run("success", func(t *testing.T) {
		logs := []*domain.Log{{ID: "1"}, {ID: "2"}}
		repo := &repoStub{queryLogs: logs}
		svc := newTestService(repo, &ipFetcherStub{}, newCacheStub(), "secret")

		got, err := svc.PostQuery(context.Background(), &domain.Query{Pass: "secret"})
		if err != nil {
			t.Fatalf("PostQuery() error = %v, want nil", err)
		}
		if len(got) != len(logs) {
			t.Fatalf("PostQuery() logs length = %d, want %d", len(got), len(logs))
		}
	})
}

func TestIPFetcher(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		cache := newCacheStub()
		cached := &domain.MetaIP{IP: "1.1.1.1"}
		cache.Set("1.1.1.1", cached)
		fetcher := &ipFetcherStub{}
		svc := newTestService(&repoStub{}, fetcher, cache, "secret")

		got, err := svc.IPFetcher(context.Background(), "1.1.1.1")
		if err != nil {
			t.Fatalf("IPFetcher() error = %v, want nil", err)
		}
		if got != cached {
			t.Fatal("IPFetcher() did not return cached value")
		}

		fetcher.mu.Lock()
		defer fetcher.mu.Unlock()
		if fetcher.calls != 0 {
			t.Fatalf("FetchIPInfo calls = %d, want 0", fetcher.calls)
		}
	})

	t.Run("cache miss stores result", func(t *testing.T) {
		cache := newCacheStub()
		fetcher := &ipFetcherStub{meta: &domain.MetaIP{IP: "8.8.8.8", ISP: "Google"}}
		svc := newTestService(&repoStub{}, fetcher, cache, "secret")

		got, err := svc.IPFetcher(context.Background(), "8.8.8.8")
		if err != nil {
			t.Fatalf("IPFetcher() error = %v, want nil", err)
		}
		if got == nil || got.IP != "8.8.8.8" {
			t.Fatalf("IPFetcher() unexpected value: %+v", got)
		}

		if _, ok := cache.Get("8.8.8.8"); !ok {
			t.Fatal("cache does not contain fetched IP")
		}
	})

	t.Run("fetch error", func(t *testing.T) {
		fetcher := &ipFetcherStub{err: errors.New("upstream error")}
		svc := newTestService(&repoStub{}, fetcher, newCacheStub(), "secret")

		_, err := svc.IPFetcher(context.Background(), "9.9.9.9")
		if err == nil {
			t.Fatal("IPFetcher() error = nil, want non-nil")
		}
	})
}

func TestVisitLog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &repoStub{saveMetaCh: make(chan struct{}, 1)}
		fetcher := &ipFetcherStub{meta: &domain.MetaIP{IP: "1.2.3.4", ISP: "X"}}
		svc := newTestService(repo, fetcher, newCacheStub(), "secret")

		req := &domain.MetaReq{
			RequestID: "rid-1",
			IP:        "1.2.3.4",
			Path:      "/",
			Useragent: "ua",
		}
		svc.VisitLog(req)

		select {
		case <-repo.saveMetaCh:
		case <-time.After(2 * time.Second):
			t.Fatal("SaveMeta was not called")
		}

		repo.mu.Lock()
		defer repo.mu.Unlock()
		if repo.saveMetaCalls != 1 {
			t.Fatalf("SaveMeta calls = %d, want 1", repo.saveMetaCalls)
		}
		if repo.lastMetaReq == nil || repo.lastMetaReq.RequestID != "rid-1" {
			t.Fatalf("SaveMeta metaReq = %+v", repo.lastMetaReq)
		}
	})

	t.Run("fetch error does not save", func(t *testing.T) {
		repo := &repoStub{}
		fetcher := &ipFetcherStub{err: errors.New("fail")}
		svc := newTestService(repo, fetcher, newCacheStub(), "secret")

		svc.VisitLog(&domain.MetaReq{IP: "1.1.1.1"})
		time.Sleep(100 * time.Millisecond)

		repo.mu.Lock()
		defer repo.mu.Unlock()
		if repo.saveMetaCalls != 0 {
			t.Fatalf("SaveMeta calls = %d, want 0", repo.saveMetaCalls)
		}
	})
}
