#!/bin/bash

# Build all

path_scripts=./scripts
deploy_path=$GOPATH/swpc/release

$path_scripts/test-go.sh
$path_scripts/lint-go.sh

rm -r "$deploy_path/public"
find "$deploy_path" -maxdepth 1 -type f -delete

mkdir -p "$deploy_path/data"

$path_scripts/build-ui.sh "$deploy_path"

GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o "$deploy_path/swpc-server" cmd/swpc-server/main.go

echo "Release deployment: '$deploy_path'"
