package strategy

import (
	"trading-algo-generator/internal/core"
)

// DeltaTrendConfig aligns order flow with profile skew.
type DeltaTrendConfig struct {
	DeltaThreshold float64
	ProfileSkew    float64
	MinVolume      float64
}

// DeltaTrendStrategy takes trend trades when order flow aligns with profile.
type DeltaTrendStrategy struct {
	Config DeltaTrendConfig
}

func (s *DeltaTrendStrategy) Name() string { return "delta_trend" }

func (s *DeltaTrendStrategy) OnTick(tick core.Tick, features core.FeatureSet, position core.Position) *core.Signal {
	if position.Open {
		return nil
	}
	delta := features.Values["delta_norm"]
	skew := features.Values["vp_skew"]
	vol := features.Values["ohlcv_vol_sma"]
	if vol < s.Config.MinVolume {
		return nil
	}
	if delta >= s.Config.DeltaThreshold && skew >= s.Config.ProfileSkew {
		return &core.Signal{Timestamp: tick.Timestamp, Direction: core.Long, Confidence: 0.65, Reason: "delta_profile_long"}
	}
	if delta <= -s.Config.DeltaThreshold && skew <= -s.Config.ProfileSkew {
		return &core.Signal{Timestamp: tick.Timestamp, Direction: core.Short, Confidence: 0.65, Reason: "delta_profile_short"}
	}
	return nil
}
