# Trading Algo Generator - Engineering Notes

## Architecture Overview
- Go backend handles ingestion, replay, features, strategy evaluation, risk, and mock execution.
- Python scripts train per-feature directional classifiers (ridge/logit/forest) and score signals.
- Storage uses JSON Lines for ticks; features are exported as CSV for ML.

## Data Model
### CSV Input (historical)
Required columns (header names are case-insensitive):
- timestamp (RFC3339Nano)
- open, high, low, close
- volume
- bid_ask_delta
- volume_profile (price:vol|price:vol...)
- session (RTH/ETH/other)
- symbol

### Tick Store
Stored as JSON Lines. Each line is a `core.Tick` serialized to JSON.

## Pipelines
### Ingestion
1. `tagen ingest --input data.csv --output ticks.jsonl`
2. CSV reader validates and parses each row.
3. Pipeline appends ticks into the JSONL tick store.

### Replay
- `tagen replay --input ticks.jsonl --speed 50`
- Replays ticks using timestamp deltas, scaled by `--speed`.

### Live Simulated Feed
- `tagen live --input ticks.jsonl --config configs/strategies/breakout.json --speed 1`
- Uses `LiveSimulator` to emit ticks with timestamp pacing.

### Features
- `tagen features --input ticks.jsonl --output features.csv`
- Features:
  - OHLCV: close, range, body, SMA, distance from SMA, volume SMA
  - Delta: raw delta and normalized delta
  - Volume profile: level count + skew
  - Session markers
  - Time-of-day sin/cos
- Labels: next-tick directional move (1 up, -1 down, 0 flat).

## Strategy Engine
- `tagen run --input ticks.jsonl --config configs/strategies/breakout.json`
- Uses configured strategy template + risk manager + mock broker.

### Risk Controls
- Hard daily stop loss (halt new trades once crossed)
- Per-trade stop in ticks
- Break-even +1 tick logic after a favorable move
- Trailing stop based on max favorable excursion

## ML Workflow
1. Export features from Go:
   - `tagen features --input ticks.jsonl --output features.csv`
2. Train per-feature models:
   - `python ml/train_per_feature.py --features features.csv --out ml/models`
3. Score and create per-tick signals:
   - `python ml/score_per_feature.py --features features.csv --models ml/models --out ml/scores.csv`

## Configurable Templates
Configs live in `configs/strategies/`.
- `breakout.json`
- `mean_reversion.json`
- `delta_trend.json`

## CLI Dashboards
- `tagen dashboard --input ticks.jsonl --config configs/strategies/breakout.json`
- Prints rolling stats (position, trades, win rate, expectancy, daily PnL).

## Notes
- All strategy and risk logic is deterministic and non-ML.
- ML models are per-feature only; no monolithic feature vectors are used.
- The system is prepared for replacing the mock broker with a live API.
