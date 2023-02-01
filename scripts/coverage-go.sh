#!/bin/bash

# Coverage in GO

go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg ./... ./...