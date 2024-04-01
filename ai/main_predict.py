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

import argparse
from pathlib import Path
import sys

import pandas as pd
from predict import predict

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description='Predicts water quality and chlorine')

    parser.add_argument('-mwq', '-model_wq',
                        type=str,
                        dest='model_wq',
                        required=True,
                        help='Water quality model file')
    parser.add_argument('-mcl', '-model_cl',
                        type=str,
                        dest='model_cl',
                        required=True,
                        help='Water chlorine model file')

    parser.add_argument('-e', '-error_file',
                        type=str,
                        dest='error_file',
                        required=True,
                        help='Error file')

    parser.add_argument('-temp',
                        type=float,
                        dest='temp',
                        required=True,
                        help='Temperature')
    parser.add_argument('-ph',
                        type=float,
                        dest='ph',
                        required=True,
                        help='PH')
    parser.add_argument('-orp',
                        type=float,
                        dest='orp',
                        required=True,
                        help='ORP')

    args = parser.parse_args()

    try:
        x = pd.DataFrame(columns=['temp', 'ph', 'orp'])
        x.loc[1] = [args.temp, args.ph, args.orp]
        wq = predict(x, Path(args.model_wq))

        x = pd.DataFrame([[args.ph, args.orp]])
        cl = predict(x, Path(args.model_cl))

        print(f'{wq};{cl}')
    except Exception as ex:
        Path(args.error_file).write_text(str(ex))
        sys.exit(1)
