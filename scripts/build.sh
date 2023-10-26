#!/bin/bash

# Build all

path_scripts=./scripts
deploy_path=$GOPATH/swpc/release

$path_scripts/test-go.sh
$path_scripts/lint-go.sh

rm -r "$deploy_path"

mkdir -p "$deploy_path/data"

$path_scripts/build-ui.sh "$deploy_path"

GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o "$deploy_path/swpc-emulator" cmd/swpc-emulator/main.go
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o "$deploy_path/swpc-server" cmd/swpc-server/main.go

echo "Release deployment: '$deploy_path'"
