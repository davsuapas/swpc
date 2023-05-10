#!/bin/bash

# Coverage in GO

param1=$1

go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg ./... ./...

# The emulator main is excluded because the emulator is a tool for testing the hub,
# and it is not necessary to test on the emulator.
grep -v "/cmd/swpc-emulator/main.go" ./coverage.out > ./coverage-final.out
mv ./coverage-final.out ./coverage.out

if [[ $param1 == "-report" ]]; then
    go tool cover -func=coverage.out
fi
