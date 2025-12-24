#!/usr/bin/env bash
set -euo pipefail

export GOFLAGS="-buildvcs=false"

mkdir -p var

if [ "${DEV_UPDATE_DEPS:-0}" = "1" ]; then
  go get -u
fi
go mod download
go mod tidy

gofmt -s -w .
fmt_out=$(gofmt -s -l . || true)
if [ -n "$fmt_out" ]; then
  echo "$fmt_out" && exit 1
fi

MODULE=$(go list -m)
goimports -local "$MODULE" -w .
imports_out=$(goimports -l -local "$MODULE" ./ || true)
if [ -n "$imports_out" ]; then
  echo "$imports_out" && exit 1
fi

go vet ./...

golangci-lint run --config tools/.golangci-lint.yml --timeout 3m --fix ./...
golangci-lint run --config tools/.golangci-lint.yml --timeout 3m ./...

go test ./... -covermode=atomic -coverprofile=var/coverage.out
go tool cover -func=var/coverage.out | tail -n 1
go tool cover -html=var/coverage.out -o var/coverage.html

git config --global --add safe.directory /workspace || true
LDV_VERSION="$(git describe --tags --always 2>/dev/null || echo dev)"
LDV_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo none)"
LDV_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
go build -v -ldflags "-X github.com/GoLessons/sufir-keeper-client/internal/buildinfo.version=${LDV_VERSION} -X github.com/GoLessons/sufir-keeper-client/internal/buildinfo.commit=${LDV_COMMIT} -X github.com/GoLessons/sufir-keeper-client/internal/buildinfo.date=${LDV_DATE}" ./cmd/keepcli
./keepcli --version >/dev/null
