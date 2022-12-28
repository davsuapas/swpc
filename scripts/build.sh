#!/bin/bash

path_target=~/go/bin/usr

mkdir -p "$path_target"

go build -o $path_target/swpc-encrypt cmd/swpc-encrypt/main.go
go build -o $path_target/swpc-emulator cmd/swpc-emulator/main.go
go build -o $path_target/swpc-server cmd/swpc-server/main.go
