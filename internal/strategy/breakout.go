package strategy

import (
	"math"

	"trading-algo-generator/internal/core"
)

// BreakoutConfig controls the breakout template.
type BreakoutConfig struct {
	Lookback int
	MinRange float64
	Confidence float64
}

// BreakoutStrategy trades when price breaks out of recent range.
type BreakoutStrategy struct {
	Config BreakoutConfig
	window []core.Tick
}

func (s *BreakoutStrategy) Name() string { return "breakout" }

func (s *BreakoutStrategy) OnTick(tick core.Tick, features core.FeatureSet, position core.Position) *core.Signal {
	s.window = append(s.window, tick)
	if s.Config.Lookback <= 0 {
		s.Config.Lookback = 20
	}
	if len(s.window) > s.Config.Lookback {
		s.window = s.window[len(s.window)-s.Config.Lookback:]
	}
	if len(s.window) < s.Config.Lookback {
		return nil
	}

	high := s.window[0].High
	low := s.window[0].Low
	for _, t := range s.window {
		if t.High > high {
			high = t.High
		}
		if t.Low < low {
			low = t.Low
		}
	}
	rangeSize := high - low
	if rangeSize < s.Config.MinRange {
		return nil
	}

	if position.Open {
		return nil
	}
	if tick.Close > high {
		return &core.Signal{Timestamp: tick.Timestamp, Direction: core.Long, Confidence: confidence(s.Config.Confidence, rangeSize), Reason: "range_break_high"}
	}
	if tick.Close < low {
		return &core.Signal{Timestamp: tick.Timestamp, Direction: core.Short, Confidence: confidence(s.Config.Confidence, rangeSize), Reason: "range_break_low"}
	}
	return nil
}

func confidence(base, rangeSize float64) float64 {
	if base <= 0 {
		base = 0.55
	}
	return math.Min(0.99, base+rangeSize*0.01)
}
