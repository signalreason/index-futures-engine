# Trading Algo Generator

Fully automated index-futures research, signal-generation, and execution engine.

## Build
```
go build ./cmd/tagen
```

## Core Commands
- Ingest CSV:
  - `./tagen ingest --input data.csv --output ticks.jsonl`
- Replay:
  - `./tagen replay --input ticks.jsonl --speed 50`
- Simulated live feed:
  - `./tagen live --input ticks.jsonl --config configs/strategies/breakout.json --speed 1`
- Generate features:
  - `./tagen features --input ticks.jsonl --output features.csv`
- Run strategy:
  - `./tagen run --input ticks.jsonl --config configs/strategies/breakout.json`
- Dashboard:
  - `./tagen dashboard --input ticks.jsonl --config configs/strategies/breakout.json`

## ML Workflow
```
python ml/train_per_feature.py --features features.csv --out ml/models
python ml/score_per_feature.py --features features.csv --models ml/models --out ml/scores.csv
```

## Documentation
See `docs/ENGINEERING.md` for full architecture and data details.
