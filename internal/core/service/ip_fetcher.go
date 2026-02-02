package service

import (
	"context"
	"portfolio/internal/core/domain"
)

func (svc *Service) IPFetcher(ctx context.Context, ip string) (*domain.MetaIP, error) {
	if metaIP, found := svc.metaIpCache.Get(ip); found {
		return metaIP, nil
	}

	metaIP, err := svc.ipfetcher.FetchIPInfo(ctx, ip)
	if err != nil {
		svc.logger.Error("failed to fetch ipinfo", "error", err)
		return nil, err
	}

	svc.metaIpCache.Set(ip, metaIP)
	return metaIP, nil
}
