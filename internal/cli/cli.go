package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"trading-algo-generator/internal/config"
	"trading-algo-generator/internal/core"
	"trading-algo-generator/internal/eval"
	"trading-algo-generator/internal/execution"
	"trading-algo-generator/internal/features"
	"trading-algo-generator/internal/ingestion"
	"trading-algo-generator/internal/replay"
	"trading-algo-generator/internal/risk"
	"trading-algo-generator/internal/storage"
)

func Run() error {
	if len(os.Args) < 2 {
		return usage()
	}
	switch os.Args[1] {
	case "ingest":
		return ingestCmd(os.Args[2:])
	case "features":
		return featuresCmd(os.Args[2:])
	case "replay":
		return replayCmd(os.Args[2:])
	case "live":
		return liveCmd(os.Args[2:])
	case "run":
		return runCmd(os.Args[2:])
	case "dashboard":
		return dashboardCmd(os.Args[2:])
	default:
		return usage()
	}
}

func usage() error {
	fmt.Fprintln(os.Stderr, "Usage: tagen <ingest|features|replay|live|run|dashboard> [args]")
	return fmt.Errorf("invalid command")
}

func ingestCmd(args []string) error {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	input := fs.String("input", "", "path to CSV input")
	output := fs.String("output", "", "path to tick store")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" || *output == "" {
		return fmt.Errorf("input and output required")
	}

	reader := ingestion.CSVReader{Path: *input}
	ticks, errs := reader.Stream()
	store := storage.TickStore{Path: *output}
	pipeline := ingestion.Pipeline{Store: &store}
	ctx := context.Background()
	return pipeline.Run(ctx, ticks, errs)
}

func featuresCmd(args []string) error {
	fs := flag.NewFlagSet("features", flag.ExitOnError)
	input := fs.String("input", "", "path to tick store")
	output := fs.String("output", "", "path to feature CSV")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" || *output == "" {
		return fmt.Errorf("input and output required")
	}

	store := storage.TickStore{Path: *input}
	ticks, err := store.LoadAll()
	if err != nil {
		return err
	}
	engine := features.Engine{
		Generators: []features.Generator{
			&features.OHLCVGenerator{Window: 20},
			features.DeltaGenerator{},
			features.VolumeProfileGenerator{},
			features.SessionGenerator{},
			features.TimeGenerator{},
		},
	}
	featureSets := make([]core.FeatureSet, 0, len(ticks))
	for _, tick := range ticks {
		featureSets = append(featureSets, engine.Build(tick))
	}
	labelFunc := func(idx int, fs core.FeatureSet) int {
		// Simple directional label: next close higher or lower.
		if idx+1 >= len(ticks) {
			return 0
		}
		if ticks[idx+1].Close > ticks[idx].Close {
			return 1
		}
		if ticks[idx+1].Close < ticks[idx].Close {
			return -1
		}
		return 0
	}
	return features.Export(*output, featureSets, labelFunc)
}

func replayCmd(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	input := fs.String("input", "", "path to tick store")
	speed := fs.Float64("speed", 50, "replay speed factor")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		return fmt.Errorf("input required")
	}

	store := storage.TickStore{Path: *input}
	ticks, err := store.LoadAll()
	if err != nil {
		return err
	}
	engine := replay.Engine{Speed: *speed}
	engine.Run(ticks, func(tick core.Tick) {
		fmt.Printf("%s %.2f %.0f %s\n", tick.Timestamp.Format(time.RFC3339Nano), tick.Close, float64(tick.Volume), tick.Session)
	})
	return nil
}

func liveCmd(args []string) error {
	fs := flag.NewFlagSet("live", flag.ExitOnError)
	input := fs.String("input", "", "path to tick store")
	configPath := fs.String("config", "", "path to strategy config")
	speed := fs.Float64("speed", 1, "replay speed factor")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" || *configPath == "" {
		return fmt.Errorf("input and config required")
	}

	cfg, err := config.LoadStrategyConfig(*configPath)
	if err != nil {
		return err
	}
	applyRiskTickSize(&cfg)
	strat, err := config.BuildStrategy(cfg)
	if err != nil {
		return err
	}

	store := storage.TickStore{Path: *input}
	tickStream, errStream := store.Stream()
	sim := ingestion.LiveSimulator{Speed: *speed}
	liveTicks := sim.Stream(tickStream)

	engine := core.Engine{
		Strategy: strat,
		Features: features.Engine{Generators: []features.Generator{
			&features.OHLCVGenerator{Window: 20},
			features.DeltaGenerator{},
			features.VolumeProfileGenerator{},
			features.SessionGenerator{},
			features.TimeGenerator{},
		}},
		Risk: &risk.Manager{Settings: cfg.Risk},
		Broker: &execution.MockBroker{},
		Evaluator: &eval.Evaluator{},
		TickSize: cfg.TickSize,
		TradeSize: cfg.Size,
		Symbol: cfg.Symbol,
	}
	if err := engine.Validate(); err != nil {
		return err
	}

	for tick := range liveTicks {
		if err := engine.OnTick(tick); err != nil {
			return err
		}
	}
	if err := <-errStream; err != nil {
		return err
	}
	printDashboard(&engine)
	return nil
}

func runCmd(args []string) error {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	input := fs.String("input", "", "path to tick store")
	configPath := fs.String("config", "", "path to strategy config")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" || *configPath == "" {
		return fmt.Errorf("input and config required")
	}

	cfg, err := config.LoadStrategyConfig(*configPath)
	if err != nil {
		return err
	}
	applyRiskTickSize(&cfg)
	strat, err := config.BuildStrategy(cfg)
	if err != nil {
		return err
	}
	store := storage.TickStore{Path: *input}
	ticks, err := store.LoadAll()
	if err != nil {
		return err
	}
	engine := core.Engine{
		Strategy: strat,
		Features: features.Engine{Generators: []features.Generator{
			&features.OHLCVGenerator{Window: 20},
			features.DeltaGenerator{},
			features.VolumeProfileGenerator{},
			features.SessionGenerator{},
			features.TimeGenerator{},
		}},
		Risk: &risk.Manager{Settings: cfg.Risk},
		Broker: &execution.MockBroker{},
		Evaluator: &eval.Evaluator{},
		TickSize: cfg.TickSize,
		TradeSize: cfg.Size,
		Symbol: cfg.Symbol,
	}
	if err := engine.Validate(); err != nil {
		return err
	}
	for _, tick := range ticks {
		if err := engine.OnTick(tick); err != nil {
			return err
		}
	}
	summary := engine.Evaluator.Summary()
	fmt.Printf("Trades: %d Wins: %d Losses: %d WinRate: %.2f Expectancy: %.2f MaxDD: %.2f\n",
		summary.TotalTrades, summary.Wins, summary.Losses, summary.WinRate, summary.Expectancy, summary.MaxDrawdown)
	fmt.Printf("Win/Loss Distribution: %+v\n", summary.WinLossDist)
	return nil
}

func dashboardCmd(args []string) error {
	fs := flag.NewFlagSet("dashboard", flag.ExitOnError)
	input := fs.String("input", "", "path to tick store")
	configPath := fs.String("config", "", "path to strategy config")
	refresh := fs.Duration("refresh", 2*time.Second, "dashboard refresh interval")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" || *configPath == "" {
		return fmt.Errorf("input and config required")
	}

	cfg, err := config.LoadStrategyConfig(*configPath)
	if err != nil {
		return err
	}
	applyRiskTickSize(&cfg)
	strat, err := config.BuildStrategy(cfg)
	if err != nil {
		return err
	}

	store := storage.TickStore{Path: *input}
	ticks, err := store.LoadAll()
	if err != nil {
		return err
	}

	engine := core.Engine{
		Strategy: strat,
		Features: features.Engine{Generators: []features.Generator{
			&features.OHLCVGenerator{Window: 20},
			features.DeltaGenerator{},
			features.VolumeProfileGenerator{},
			features.SessionGenerator{},
			features.TimeGenerator{},
		}},
		Risk: &risk.Manager{Settings: cfg.Risk},
		Broker: &execution.MockBroker{},
		Evaluator: &eval.Evaluator{},
		TickSize: cfg.TickSize,
		TradeSize: cfg.Size,
		Symbol: cfg.Symbol,
	}
	if err := engine.Validate(); err != nil {
		return err
	}

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(*refresh)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				printDashboard(&engine)
			case <-stop:
				return
			}
		}
	}()

	for _, tick := range ticks {
		if err := engine.OnTick(tick); err != nil {
			return err
		}
	}
	close(stop)
	printDashboard(&engine)
	return nil
}

func printDashboard(engine *core.Engine) {
	summary := engine.Evaluator.Summary()
	pos := "FLAT"
	if engine.Position.Open {
		if engine.Position.Direction == core.Long {
			pos = "LONG"
		} else {
			pos = "SHORT"
		}
	}
	fmt.Printf("POS: %s Trades: %d WinRate: %.2f Expectancy: %.2f DailyPnL: %.2f\n",
		pos, summary.TotalTrades, summary.WinRate, summary.Expectancy, engine.Risk.DailyPnL)
}

func applyRiskTickSize(cfg *config.StrategyConfig) {
	if cfg.Risk.TickSize == 0 && cfg.TickSize > 0 {
		cfg.Risk.TickSize = cfg.TickSize
	}
}
