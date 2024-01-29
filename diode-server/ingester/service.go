package ingester

import (
	"context"
	"log/slog"
	"time"
)

type Service struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) Name() string {
	return "ingester"
}

func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("starting service", "name", s.Name())

	return s.ping(ctx)
}

func (s *Service) ping(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)

	for c := ticker.C; ; {
		s.logger.Info("ping", "serviceName", s.Name())
		select {
		case <-c:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Service) Stop() error {
	s.logger.Info("stopping service", "name", s.Name())
	return nil
}
