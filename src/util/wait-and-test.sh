#!/bin/sh
set -e

echo "Clearing test cache..."
go clean -testcache

echo "Running tests..."
go test -v ./test/integration/...