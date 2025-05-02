#!/bin/bash

# Deploy AI module

path_target=$1

if [[ -z "$path_target" ]]; then
    path_target="./cmd/swpc-server"
fi

path_target="$path_target/ai"

rm -r "$path_target"
mkdir -p "$path_target"

cp ./ai/main_predict.py "$path_target/"
cp ./ai/predict.* "$path_target/"
chmod +x "$path_target/predict.sh"

cp ./ai/model/*_wq "$path_target/"
cp ./ai/model/*_cl "$path_target/"

python3 -m venv "$path_target/.venv/swpc_predict"
source "$path_target/.venv/swpc_predict/bin/activate"

pip install -r ./ai/requirement_predict.yml
