package service

import (
	"context"
	"portfolio/internal/core/domain"
	"time"
)

func (svc *Service) VisitLog(metaReq *domain.MetaReq) {
	go func() {
		ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()

		metaIP, err := svc.IPFetcher(ctx, metaReq.IP)
		if err != nil {
			return
		}

		if err := svc.repo.SaveMeta(ctx, metaReq, metaIP); err != nil {
			svc.logger.Error("failed to save ipinfo", "error", err)
		}
	}()
}
