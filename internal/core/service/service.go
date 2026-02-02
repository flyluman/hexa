package service

import (
	"portfolio/internal/core/ports"
)

type Service struct {
	repo        ports.Repo
	logger      ports.Logger
	ipfetcher   ports.IPFetcher
	metaIpCache ports.MetaIPCache
	queryPass   string
}

func NewService(repo ports.Repo, logger ports.Logger, ipfetcher ports.IPFetcher, metaIpCache ports.MetaIPCache, queryPass string) ports.Service {
	return &Service{
		repo:        repo,
		logger:      logger,
		ipfetcher:   ipfetcher,
		metaIpCache: metaIpCache,
		queryPass:   queryPass,
	}
}
