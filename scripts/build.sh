#!/bin/bash

path_target=~/go/bin/usr

go build -o $path_target/swpc-encrypt cmd/swpc-encrypt/main.go
go build -o $path_target/swpc-emulator cmd/swpc-emulator/main.go
