package ingester

import (
	"context"
	"log/slog"
	"time"
)

type Component struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Component {
	return &Component{
		logger: logger,
	}
}

func (c *Component) Name() string {
	return "ingester"
}

func (c *Component) Start(ctx context.Context) error {
	c.logger.Info("starting component", "name", c.Name())

	return c.ping(ctx)
}

func (c *Component) ping(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)

	for tc := ticker.C; ; {
		c.logger.Info("ping", "componentName", c.Name())
		select {
		case <-tc:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Component) Stop() error {
	c.logger.Info("stopping component", "name", c.Name())
	return nil
}