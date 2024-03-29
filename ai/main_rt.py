import argparse
from pathlib import Path

import pandas as pd
from predict import predict

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description='Predicts water quality and chlorine')

    subparsers = parser.add_subparsers(dest="command")

    pred_wq_parser = subparsers.add_parser(
        'predict_wq',
        help='Predict water quality')
    pred_wq_parser.add_argument('-temp',
                                type=float,
                                dest='temp',
                                required=True,
                                help='Temperature')
    pred_wq_parser.add_argument('-ph',
                                type=float,
                                dest='ph',
                                required=True,
                                help='PH')
    pred_wq_parser.add_argument('-orp',
                                type=float,
                                dest='orp',
                                required=True,
                                help='ORP')
    pred_wq_parser.add_argument('-m', '-model',
                                type=str,
                                dest='model',
                                required=True,
                                help='Model file')
    pred_wq_parser.add_argument('-r', '-result_file',
                                type=str,
                                dest='result_file',
                                required=True,
                                help='Result file')

    pred_cl_parser = subparsers.add_parser(
        'predict_cl',
        help='Predict chlourine')
    pred_cl_parser.add_argument('-ph',
                                type=float,
                                dest='ph',
                                required=True,
                                help='PH')
    pred_cl_parser.add_argument('-orp',
                                type=float,
                                dest='orp',
                                required=True,
                                help='ORP')
    pred_cl_parser.add_argument('-m', '-model',
                                type=str,
                                dest='model',
                                required=True,
                                help='Model file')
    pred_cl_parser.add_argument('-r', '-result_file',
                                type=str,
                                dest='result_file',
                                required=True,
                                help='Result file')

    args = parser.parse_args()

    if args.command == "predict_wq":
        x = pd.DataFrame(columns=['temp', 'ph', 'orp'])
        x.loc[1] = [args.temp, args.ph, args.orp]

        predict(
            x,
            Path(args.model),
            Path(args.result_file))
    elif args.command == "predict_cl":
        x = pd.DataFrame([[args.ph, args.orp]])

        predict(
            x,
            Path(args.model),
            Path(args.result_file))
    else:
        print(f"Unrecognised command: {args.command}")
