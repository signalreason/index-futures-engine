import argparse
import json
from pathlib import Path

import joblib
import numpy as np
import pandas as pd
from sklearn.linear_model import LogisticRegression, RidgeClassifier
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score, f1_score
from sklearn.model_selection import train_test_split


def load_data(path: Path):
    df = pd.read_csv(path)
    if "label" not in df.columns:
        raise ValueError("label column missing")
    features = [c for c in df.columns if c not in ("timestamp", "label")]
    return df, features


def train_one(feature_name, x, y):
    x = x.reshape(-1, 1)
    x_train, x_test, y_train, y_test = train_test_split(x, y, test_size=0.2, shuffle=False)

    models = {
        "ridge": RidgeClassifier(),
        "logit": LogisticRegression(max_iter=200),
        "forest": RandomForestClassifier(n_estimators=200, max_depth=6, random_state=7),
    }

    results = {}
    for name, model in models.items():
        model.fit(x_train, y_train)
        preds = model.predict(x_test)
        results[name] = {
            "accuracy": float(accuracy_score(y_test, preds)),
            "f1": float(f1_score(y_test, preds, average="macro")),
            "model": model,
        }
    return results


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--features", required=True, help="feature CSV from Go exporter")
    parser.add_argument("--out", required=True, help="output directory")
    args = parser.parse_args()

    df, feature_cols = load_data(Path(args.features))
    out_dir = Path(args.out)
    out_dir.mkdir(parents=True, exist_ok=True)

    y = df["label"].values

    summary = {}
    for feature in feature_cols:
        x = df[feature].values.astype(np.float64)
        results = train_one(feature, x, y)
        summary[feature] = {
            name: {"accuracy": metrics["accuracy"], "f1": metrics["f1"]}
            for name, metrics in results.items()
        }
        for name, metrics in results.items():
            model_path = out_dir / f"{feature}_{name}.joblib"
            joblib.dump(metrics["model"], model_path)

    with (out_dir / "summary.json").open("w", encoding="utf-8") as f:
        json.dump(summary, f, indent=2)


if __name__ == "__main__":
    main()
