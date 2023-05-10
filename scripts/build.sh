#!/bin/bash

# Build all

path_scripts=./scripts
path_target=~/go/bin/usr/swpc

$path_scripts/test-go.sh
$path_scripts/lint-go.sh

mkdir -p "$path_target/data"

$path_scripts/build-ui.sh "$path_target"

go build -o $path_target/swpc-emulator cmd/swpc-emulator/main.go
go build -o $path_target/swpc-server cmd/swpc-server/main.go
