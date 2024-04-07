#!/bin/bash

# Wrapper to predict the water quality and chlorine
# This wrapper will be called from go

source ai/.venv/swpc_predict/bin/activate

echo $(python ai/main_predict.py $@)

if [ $? -ne 0 ]; then
    exit 1
fi

