package strategy

import (
	"math"

	"trading-algo-generator/internal/core"
)

// MeanReversionConfig controls mean reversion logic.
type MeanReversionConfig struct {
	ZThreshold float64
	Lookback   int
}

// MeanReversionStrategy fades extended moves.
type MeanReversionStrategy struct {
	Config MeanReversionConfig
	window []float64
}

func (s *MeanReversionStrategy) Name() string { return "mean_reversion" }

func (s *MeanReversionStrategy) OnTick(tick core.Tick, features core.FeatureSet, position core.Position) *core.Signal {
	s.window = append(s.window, tick.Close)
	if s.Config.Lookback <= 0 {
		s.Config.Lookback = 30
	}
	if len(s.window) > s.Config.Lookback {
		s.window = s.window[len(s.window)-s.Config.Lookback:]
	}
	if len(s.window) < s.Config.Lookback {
		return nil
	}

	mean, std := meanStd(s.window)
	if std == 0 {
		return nil
	}
	z := (tick.Close - mean) / std
	if position.Open {
		return nil
	}
	if z > s.Config.ZThreshold {
		return &core.Signal{Timestamp: tick.Timestamp, Direction: core.Short, Confidence: math.Min(0.9, 0.6+z*0.05), Reason: "zscore_high"}
	}
	if z < -s.Config.ZThreshold {
		return &core.Signal{Timestamp: tick.Timestamp, Direction: core.Long, Confidence: math.Min(0.9, 0.6+(-z)*0.05), Reason: "zscore_low"}
	}
	return nil
}

func meanStd(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	return mean, math.Sqrt(variance)
}
