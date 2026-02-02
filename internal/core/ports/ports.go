package ports

import (
	"context"
	"portfolio/internal/core/domain"
)

type Repo interface {
	SaveMessage(ctx context.Context, msg *domain.Message) error
	SaveMeta(ctx context.Context, metaReq *domain.MetaReq, metaIP *domain.MetaIP) error
	QueryLog(ctx context.Context) ([]*domain.Log, error)
}

type Service interface {
	PostMessage(ctx context.Context, msg *domain.Message) error
	VisitLog(metaReq *domain.MetaReq)
	IPFetcher(ctx context.Context, ip string) (*domain.MetaIP, error)
	PostQuery(ctx context.Context, query *domain.Query) ([]*domain.Log, error)
}

type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
}

type IPFetcher interface {
	FetchIPInfo(ctx context.Context, ip string) (*domain.MetaIP, error)
}

type RateLimiter interface {
	Allow(key string) bool
}

type MetaIPCache interface {
	Get(key string) (*domain.MetaIP, bool)
	Set(key string, val *domain.MetaIP)
	Del(key string)
	Query() int
	Clear() int
}
