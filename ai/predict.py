"""
  Copyright (c) 2022 ELIPCERO
  All rights reserved.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
"""

from pathlib import Path

import joblib
import pandas as pd


def predict(x: pd.DataFrame, model_file: Path):
    """Predicts based on model

    Args:
        x: Dataframe
        model_file (Path): Model file
    """
    model = joblib.load(model_file)
    predictions = model.predict(x)

    return str(predictions[0])
