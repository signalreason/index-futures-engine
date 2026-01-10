package ingestion

import (
	"context"
	"fmt"

	"trading-algo-generator/internal/core"
	"trading-algo-generator/internal/storage"
)

// Pipeline connects a data source to a tick store.
type Pipeline struct {
	Store *storage.TickStore
}

func (p Pipeline) Run(ctx context.Context, ticks <-chan core.Tick, errs <-chan error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err, ok := <-errs:
			if ok && err != nil {
				return err
			}
			if !ok {
				errs = nil
			}
		case tick, ok := <-ticks:
			if !ok {
				return nil
			}
			if err := p.Store.Append(tick); err != nil {
				return fmt.Errorf("store append: %w", err)
			}
		}
	}
}
