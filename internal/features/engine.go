package features

import (
	"math"
	"time"

	"trading-algo-generator/internal/core"
)

// Generator produces named features per tick.
type Generator interface {
	Name() string
	Generate(tick core.Tick) map[string]float64
}

// Engine maintains generators and merges their outputs.
type Engine struct {
	Generators []Generator
}

func (e Engine) Build(tick core.Tick) core.FeatureSet {
	values := make(map[string]float64)
	for _, gen := range e.Generators {
		for k, v := range gen.Generate(tick) {
			values[k] = v
		}
	}
	return core.FeatureSet{Timestamp: tick.Timestamp, Values: values}
}

// OHLCVGenerator creates price action features.
type OHLCVGenerator struct {
	Window int
	prices []float64
	vols   []int64
}

func (g *OHLCVGenerator) Name() string { return "ohlcv" }

func (g *OHLCVGenerator) Generate(tick core.Tick) map[string]float64 {
	g.prices = append(g.prices, tick.Close)
	g.vols = append(g.vols, tick.Volume)
	if g.Window > 0 && len(g.prices) > g.Window {
		g.prices = g.prices[len(g.prices)-g.Window:]
		g.vols = g.vols[len(g.vols)-g.Window:]
	}
	avg := averageFloat(g.prices)
	return map[string]float64{
		"ohlcv_close": tick.Close,
		"ohlcv_range": tick.High - tick.Low,
		"ohlcv_body":  tick.Close - tick.Open,
		"ohlcv_sma":   avg,
		"ohlcv_sma_dist": tick.Close - avg,
		"ohlcv_vol_sma":  float64(averageInt(g.vols)),
	}
}

// DeltaGenerator creates order flow features.
type DeltaGenerator struct{}

func (g DeltaGenerator) Name() string { return "delta" }

func (g DeltaGenerator) Generate(tick core.Tick) map[string]float64 {
	return map[string]float64{
		"delta_raw": float64(tick.BidAskDelta),
		"delta_norm": float64(tick.BidAskDelta) / math.Max(1, float64(tick.Volume)),
	}
}

// VolumeProfileGenerator extracts simple profile metrics.
type VolumeProfileGenerator struct{}

func (g VolumeProfileGenerator) Name() string { return "volume_profile" }

func (g VolumeProfileGenerator) Generate(tick core.Tick) map[string]float64 {
	if len(tick.VolumeProfile) == 0 {
		return map[string]float64{
			"vp_levels": 0,
			"vp_skew":   0,
		}
	}
	var totalVol int64
	var weighted float64
	min := tick.VolumeProfile[0].Price
	max := tick.VolumeProfile[0].Price
	for _, level := range tick.VolumeProfile {
		totalVol += level.Volume
		weighted += level.Price * float64(level.Volume)
		if level.Price < min {
			min = level.Price
		}
		if level.Price > max {
			max = level.Price
		}
	}
	mean := weighted / math.Max(1, float64(totalVol))
	mid := (min + max) / 2
	skew := (mean - mid) / math.Max(0.0001, max-min)
	return map[string]float64{
		"vp_levels": float64(len(tick.VolumeProfile)),
		"vp_skew":   skew,
	}
}

// SessionGenerator encodes session markers.
type SessionGenerator struct{}

func (g SessionGenerator) Name() string { return "session" }

func (g SessionGenerator) Generate(tick core.Tick) map[string]float64 {
	switch tick.Session {
	case "RTH":
		return map[string]float64{"session_rth": 1}
	case "ETH":
		return map[string]float64{"session_eth": 1}
	default:
		return map[string]float64{"session_other": 1}
	}
}

// TimeGenerator adds cyclical time-of-day features.
type TimeGenerator struct{}

func (g TimeGenerator) Name() string { return "time" }

func (g TimeGenerator) Generate(tick core.Tick) map[string]float64 {
	t := tick.Timestamp.In(time.UTC)
	seconds := float64(t.Hour()*3600 + t.Minute()*60 + t.Second())
	angle := 2 * math.Pi * (seconds / 86400.0)
	return map[string]float64{
		"tod_sin": math.Sin(angle),
		"tod_cos": math.Cos(angle),
	}
}

func averageFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func averageInt(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	var sum int64
	for _, v := range values {
		sum += v
	}
	return sum / int64(len(values))
}
