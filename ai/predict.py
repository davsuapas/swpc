from pathlib import Path

import joblib
import pandas as pd


def predict(
        x: pd.DataFrame,
        model_file: Path,
        result_file: Path):
    """Predicts based on model

    Args:
        x: Dataframe
        model_file (Path): Model file
        result_file (Path): Prediction result
    """
    model = joblib.load(model_file)
    predictions = model.predict(x)

    result_file.write_text(str(predictions[0]))
