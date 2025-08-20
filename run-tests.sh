#!/bin/bash

cd src
mkdir test-out

set -euo pipefail

echo "Running unit tests..."

go test -v -coverprofile=./test-out/coverage-unit.out -coverpkg=./... ./test/unit/...

echo "Building coverage report..."

go tool cover -func=./test-out/coverage-unit.out
go tool cover -html=./test-out/coverage-unit.out -o ./test-out/coverage-unit.html

echo "Linting..."

go vet ./...