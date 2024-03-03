import argparse
from pathlib import Path
from fit import create_dataframe, fit_water_quality, fit_chlorine
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
                            help='swpc sample file')
    fit_parser.add_argument('-t', '-decision_tree',
                            type=str,
                            dest='decision_tree',
                            help='Decision tree graph png file')

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

    args = parser.parse_args()

    if args.command == "fit":
        sample = create_dataframe(Path(args.sample))
        # fit_chlorine(sample)
        fit_water_quality(sample, Path(args.decision_tree))
    elif args.command == "sample":
        random_sample(Path(args.result), args.size_dataset)
    else:
        print(f"Unrecognised command: {args.command}")
