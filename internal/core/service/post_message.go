package service

import (
	"context"
	"portfolio/internal/core/domain"
)

func (svc *Service) PostMessage(ctx context.Context, msg *domain.Message) error {
	if err := svc.repo.SaveMessage(ctx, msg); err != nil {
		svc.logger.Error("failed to save message", "error", err)
		return err
	}

	return nil
}
