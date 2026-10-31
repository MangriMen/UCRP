#!/usr/bin/env bash
set -euo pipefail

golangci_lint_version=2.15.0
api_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../apps/api" && pwd)"
cd "$api_dir"

find_golangci_lint() {
  if command -v golangci-lint >/dev/null 2>&1; then
    command -v golangci-lint
    return
  fi
  local candidate
  for candidate in "$(go env GOPATH)/bin/golangci-lint" "$(go env GOPATH)/bin/golangci-lint.exe"; do
    if [ -x "$candidate" ]; then
      echo "$candidate"
      return
    fi
  done
  echo "golangci-lint v$golangci_lint_version not found. Install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v$golangci_lint_version" >&2
  return 1
}

check_fmt() {
  local unformatted
  unformatted="$(gofmt -l .)"
  if [ -n "$unformatted" ]; then
    echo "$unformatted" >&2
    echo "gofmt: run 'gofmt -w' on the files above" >&2
    return 1
  fi
}

check_lint() {
  local binary
  binary="$(find_golangci_lint)"
  if ! "$binary" version 2>&1 | grep -q "version $golangci_lint_version"; then
    echo "golangci-lint v$golangci_lint_version required, found: $("$binary" version 2>&1 | head -1)" >&2
    return 1
  fi
  "$binary" run ./...
}

run_step() {
  case "$1" in
    fmt) check_fmt ;;
    vet) go vet ./... ;;
    lint) check_lint ;;
    test) go test -race ./... ;;
    build) go build ./... ;;
    static) check_fmt && go vet ./... && check_lint ;;
    all) check_fmt && go vet ./... && check_lint && go test -race ./... && go build ./... ;;
    *)
      echo "usage: $0 {fmt|vet|lint|test|build|static|all}..." >&2
      return 2
      ;;
  esac
}

if [ "$#" -eq 0 ]; then
  run_step all
  exit
fi

for step in "$@"; do
  run_step "$step"
done
