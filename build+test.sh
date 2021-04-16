#!/bin/bash -e

cd "$(dirname $0)"
PATH=$HOME/go/bin:$PATH

echo "go mod download"
go mod download

echo "go test ./..."
go test ./...

echo "gofmt -s -w -l ..."
gofmt -s -w -l $(find . -type f -name \*.go)

echo "go vet ./..."
go vet ./...

#echo "go install ./cmd/..."
#go install ./cmd/...
