#!/bin/bash

# Build all

option=$1

path_scripts=./scripts
deploy_path=$GOPATH/swpc/release

mkdir -p "$deploy_path/data"

if [[ $option == '-test' || $option == '' ]]; then
  $path_scripts/test-go.sh
  $path_scripts/lint-go.sh
fi

if [[ $option == '-ui' || $option == '' ]]; then
  rm -r "$deploy_path/public"
  find "$deploy_path" -maxdepth 1 -type f -delete
  $path_scripts/build-ui.sh "$deploy_path" -no-map
fi

if [[ $option == '-ai' || $option == '' ]]; then
  $path_scripts/build-ai.sh "$deploy_path"
fi

GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o "$deploy_path/swpc-server" cmd/swpc-server/main.go

echo "Release deployment: '$deploy_path'"
