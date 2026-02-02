package rest

import (
	"portfolio/internal/core/ports"
)

type RestHandler struct {
	svc    ports.Service
	logger ports.Logger
	rl     ports.RateLimiter
}

func NewRestHandler(svc ports.Service, logger ports.Logger, rl ports.RateLimiter) *RestHandler {
	return &RestHandler{
		svc:    svc,
		logger: logger,
		rl:     rl,
	}
}
