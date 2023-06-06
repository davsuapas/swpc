#!/bin/bash

# Build all

path_scripts=./scripts
deploy_path=$GOPATH/swpc/release

$path_scripts/test-go.sh
$path_scripts/lint-go.sh

mkdir -p "$deploy_path/data"

chmod -R 755 $deploy_path/

$path_scripts/build-ui.sh "$deploy_path"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $deploy_path/swpc-emulator cmd/swpc-emulator/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $deploy_path/swpc-server cmd/swpc-server/main.go

chmod -R 555 $deploy_path
chmod 755 $deploy_path/swpc-*

echo "Target Deployment: '$deploy_path'"
