import argparse
from pathlib import Path

import pandas as pd
from fit import create_dataframe, fit_water_quality, fit_chlorine
from predict import predict
from sample import random_sample

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description='Generates random data and generates a model '
        'to obtain water quality and chlorine')

    subparsers = parser.add_subparsers(dest="command")

    fit_parser = subparsers.add_parser('fit', help='Generate model')
    fit_parser.add_argument('-s', '-sample',
                            type=str,
                            dest='sample',
                            required=True,
                            help='swpc sample file')
    fit_parser.add_argument('-t', '-decision_tree',
                            type=str,
                            dest='decision_tree',
                            required=True,
                            help='Decision tree graph png file')
    fit_parser.add_argument('-m', '-model',
                            type=str,
                            dest='model',
                            required=True,
                            help='Model file')

    sample_parser = subparsers.add_parser(
        'sample',
        help='Generates data sample')
    sample_parser.add_argument(
        '-s', '-size_dataset',
        type=int,
        default=1000,
        dest='size_dataset',
        help='Data sample size')
    sample_parser.add_argument(
        '-r', '-result',
        type=str,
        dest='result',
        help='Data sample file')

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

    if args.command == "fit":
        model_file = Path(args.model)
        model_wq_file = model_file.with_name(model_file.name + "_wq")
        model_c_file = model_file.with_name(model_file.name + "_cl")

        sample = create_dataframe(Path(args.sample))

        fit_chlorine(sample, model_c_file)
        fit_water_quality(sample, Path(args.decision_tree), model_wq_file)
    elif args.command == "sample":
        random_sample(Path(args.result), args.size_dataset)
    elif args.command == "predict_wq":
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
