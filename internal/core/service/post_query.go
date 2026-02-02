package service

import (
	"context"
	"portfolio/internal/core/domain"
)

func (svc *Service) PostQuery(ctx context.Context, query *domain.Query) ([]*domain.Log, error) {
	if query.Pass != svc.queryPass {
		svc.logger.Error("password mismatch", "error", domain.ErrIncorrectPassword)
		return nil, domain.ErrIncorrectPassword
	}

	logs, err := svc.repo.QueryLog(ctx)
	if err != nil {
		svc.logger.Error("failed to query log", "error", err)
		return nil, err
	}

	return logs, nil
}
