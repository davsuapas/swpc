#!/bin/bash

# Coverage in GO

param1=$1

go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg ./... ./...

if [[ $param1 == "-report" ]]; then
    go tool cover -func=coverage.out
fi
