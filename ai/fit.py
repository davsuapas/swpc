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
import pandas as pd
from sklearn.linear_model import LinearRegression
from sklearn.discriminant_analysis import StandardScaler
from sklearn.model_selection import train_test_split
from sklearn.tree import DecisionTreeClassifier
from sklearn.metrics import classification_report, accuracy_score, r2_score
from sklearn.metrics import mean_absolute_error, mean_squared_error
from sklearn import tree
import graphviz
import joblib

from sample import field_cl, field_quality, field_temp


def create_dataframe(swpc_sample: Path) -> pd.DataFrame:
    """Create dataframe from path

    Args:
        swpc_sample (Path): Sample path

    Returns:
        pd.DataFrame: Sample dataframe
    """
    df = pd.read_csv(swpc_sample)

    quality_map = {
        0: 'bad',
        1: 'regular',
        2: "good",
    }

    df[field_quality] = df[field_quality].map(quality_map)

    return df


def fit_water_quality(
        swpc_sample: pd.DataFrame,
        decision_tree: Path,
        model_file: Path):
    """Create water quality model

    Args:
        swpc_sample (pd.DataFrame): swpc sample 
        decision_tree (Path): Decision tree graph png file path
        model_file (Path): Model file path
    """
    data = swpc_sample.drop(field_cl, axis=1, inplace=False)

    # The target is separated from the rest of the variables
    y = data[field_quality]
    x = data.drop(field_quality, axis=1, inplace=False)
    feature_names = x.columns
    class_names = y.unique()

    # Splitting data into training and test sets
    x_train, x_test, y_train, y_test = \
        train_test_split(x, y, test_size=0.2, random_state=42)

    model = DecisionTreeClassifier(max_depth=3)
    model.fit(x_train, y_train)

    y_pred = model.predict(x_test)

    print('\nWATER QUALITY MODEL')
    print('-------------------\n')

    print(f'Accuracy Score: {accuracy_score(y_test, y_pred)}\n')
    print(f'Classification report:\n{classification_report(y_test, y_pred)}')

    dot_data = tree.export_graphviz(
        model, out_file=None,
        feature_names=feature_names,
        class_names=class_names,
        filled=True, rounded=True,
        special_characters=True
    )

    graph = graphviz.Source(dot_data)
    graph.render(decision_tree, format='png', cleanup=True)

    print(f'Decision tree graph file: {str(decision_tree)}\n')

    # Save the trained model to a file
    joblib.dump(model, model_file)
    print(f'Saved model in the file: {str(model_file)}\n')


def fit_chlorine(swpc_sample: pd.DataFrame, model_file: Path):
    """Create chlorine model

    Args:
        swpc_sample (pd.DataFrame): Sample dataframe
        model_file (Path): Model file path
    """
    data = swpc_sample

    # The target is separated from the rest of the variables
    y = data[field_cl]
    x = data.drop([field_quality, field_temp, field_cl], axis=1, inplace=False)
    feature_names = x.columns

    scaler = StandardScaler()

    # Scaling the data
    x_scaled = x.copy()
    x_scaled = scaler.fit_transform(x_scaled)

    # Splitting data into training and test sets
    x_train, x_test, y_train, y_test = \
        train_test_split(x_scaled, y, test_size=0.2, random_state=42)

    model = LinearRegression()
    model.fit(x_train, y_train)

    coefficients = model.coef_
    feature_names = x.columns

    print('\nCHLORINE MODEL')
    print('--------------\n')

    print("Coefficients: ")
    for name, coefficient in zip(feature_names, coefficients):
        print(f"{name}: {coefficient}")
    print("\n")

    y_pred = model.predict(x_test)

    # Calculate metrics
    r2 = r2_score(y_test, y_pred)
    mse = mean_squared_error(y_test, y_pred)
    mae = mean_absolute_error(y_test, y_pred)

    print('Metrics:')
    print(f'R2: {r2}')
    print(f'Mean Squared Error (MSE): {mse}')
    print(f'Mean Absolute Error (MAE): {mae}\n')

    # Save the trained model to a file
    joblib.dump(model, model_file)
    print(f'Saved model in the file: {str(model_file)}\n')
