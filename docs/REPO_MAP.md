# Repo Map: index-futures-engine (trading-algo-generator)

## Purpose and scope
Automated index-futures research, signal generation, and execution engine combining a Go CLI with Python ML scripts.

## Quickstart commands
- Build: `go build ./cmd/tagen`
- Ingest: `./tagen ingest --input data.csv --output ticks.jsonl`
- Features: `./tagen features --input ticks.jsonl --output features.csv`
- Run strategy: `./tagen run --input ticks.jsonl --config configs/strategies/breakout.json`
- ML train: `python ml/train_per_feature.py --features features.csv --out ml/models`

## Top-level map
- `cmd/` - Go CLI entry points.
  - `cmd/tagen/` - main CLI for ingest/replay/run/dashboard.
- `internal/` - ingestion, replay, features, strategy, and broker logic.
- `configs/` - strategy and risk configuration templates.
  - `configs/strategies/` - JSON strategy configs.
- `ml/` - Python ML training and scoring scripts.
- `docs/` - architecture notes and data model.
- `go.mod`, `go.sum` - Go module definition.
- `README.md` - command overview.

## Key entry points
- `cmd/tagen/` - main Go CLI.
- `ml/train_per_feature.py` - per-feature model training.
- `ml/score_per_feature.py` - per-feature scoring.
- `docs/ENGINEERING.md` - architecture and data details.

## Core flows and data movement
- CSV input -> `tagen ingest` -> JSONL tick store.
- JSONL ticks -> `tagen features` -> CSV features.
- Features CSV -> ML train/score -> model outputs.
- JSONL ticks + strategy config -> `tagen run` -> simulated trades.

## External integrations
- Input data from CSV; outputs to JSONL/CSV.
- Prepared for future live broker integration (not included).

## Configuration and deployment
- Strategy configs in `configs/strategies/*.json`.
- No deployment manifests included.

## Common workflows (build/test/release)
- `go build ./cmd/tagen`
- `./tagen ingest|replay|live|features|run|dashboard ...`
- `python ml/train_per_feature.py ...`

## Read-next list
- `README.md`
- `docs/ENGINEERING.md`
- `cmd/tagen/`
- `internal/`
- `configs/strategies/`
- `ml/`

## Unknowns and follow-ups
- No test commands are documented in the repo.
