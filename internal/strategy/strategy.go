package strategy

import "trading-algo-generator/internal/core"

// Strategy consumes ticks and features to emit signals.
type Strategy interface {
	Name() string
	OnTick(tick core.Tick, features core.FeatureSet, position core.Position) *core.Signal
}
