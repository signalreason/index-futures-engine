package config

import (
	"encoding/json"
	"os"

	"trading-algo-generator/internal/risk"
	"trading-algo-generator/internal/strategy"
)

// StrategyConfig is a top-level strategy configuration.
type StrategyConfig struct {
	Name    string          `json:"name"`
	Params  json.RawMessage `json:"params"`
	Risk    risk.Settings   `json:"risk"`
	Size    int64           `json:"size"`
	Symbol  string          `json:"symbol"`
	TickSize float64        `json:"tick_size"`
}

func LoadStrategyConfig(path string) (StrategyConfig, error) {
	var cfg StrategyConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func BuildStrategy(cfg StrategyConfig) (strategy.Strategy, error) {
	switch cfg.Name {
	case "breakout":
		var params strategy.BreakoutConfig
		if err := json.Unmarshal(cfg.Params, &params); err != nil {
			return nil, err
		}
		return &strategy.BreakoutStrategy{Config: params}, nil
	case "mean_reversion":
		var params strategy.MeanReversionConfig
		if err := json.Unmarshal(cfg.Params, &params); err != nil {
			return nil, err
		}
		return &strategy.MeanReversionStrategy{Config: params}, nil
	case "delta_trend":
		var params strategy.DeltaTrendConfig
		if err := json.Unmarshal(cfg.Params, &params); err != nil {
			return nil, err
		}
		return &strategy.DeltaTrendStrategy{Config: params}, nil
	default:
		return nil, os.ErrNotExist
	}
}
