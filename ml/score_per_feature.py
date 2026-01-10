import argparse
import json
from pathlib import Path

import joblib
import numpy as np
import pandas as pd


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--features", required=True, help="feature CSV from Go exporter")
    parser.add_argument("--models", required=True, help="directory with per-feature models")
    parser.add_argument("--out", required=True, help="output scored CSV")
    args = parser.parse_args()

    df = pd.read_csv(args.features)
    feature_cols = [c for c in df.columns if c not in ("timestamp", "label")]

    model_dir = Path(args.models)
    scores = []
    for feature in feature_cols:
        feature_scores = []
        for model_name in ("ridge", "logit", "forest"):
            model_path = model_dir / f"{feature}_{model_name}.joblib"
            if not model_path.exists():
                continue
            model = joblib.load(model_path)
            x = df[feature].values.reshape(-1, 1)
            preds = model.predict(x)
            feature_scores.append(preds)
        if feature_scores:
            stacked = np.vstack(feature_scores)
            scores.append(np.sign(np.mean(stacked, axis=0)))

    if scores:
        ensemble = np.sign(np.mean(np.vstack(scores), axis=0))
    else:
        ensemble = np.zeros(len(df))

    out = pd.DataFrame({
        "timestamp": df["timestamp"],
        "signal": ensemble.astype(int),
    })
    out.to_csv(args.out, index=False)

    meta = {
        "features_used": feature_cols,
        "models_dir": str(model_dir),
    }
    with open(Path(args.out).with_suffix(".json"), "w", encoding="utf-8") as f:
        json.dump(meta, f, indent=2)


if __name__ == "__main__":
    main()
