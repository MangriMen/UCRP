#!/bin/sh
# Runs the UCRP E2E suite against a live API instance.
# Usage: ./apps/api/e2e/run.sh [--report-only]
set -eu

report_only=0
for arg in "$@"; do
    case "$arg" in
        --report-only) report_only=1 ;;
    esac
done

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../../.." && pwd)
api_root="$repo_root/apps/api"

export UCRP_E2E_BASE_URL="${UCRP_E2E_BASE_URL:-http://127.0.0.1:8080}"
export UCRP_E2E_DATABASE_URL="${UCRP_E2E_DATABASE_URL:-postgres://ucrp:ucrp@127.0.0.1:5543/ucrp?sslmode=disable}"
export UCRP_E2E_ADMIN_EMAIL="${UCRP_E2E_ADMIN_EMAIL:-admin@ucrp.local}"
export UCRP_E2E_ADMIN_PASSWORD="${UCRP_E2E_ADMIN_PASSWORD:-admin12345}"
export UCRP_E2E_OPENAPI="${UCRP_E2E_OPENAPI:-$repo_root/docs/swagger/openapi.json}"
export UCRP_E2E_JUNIT="${UCRP_E2E_JUNIT:-$api_root/e2e-report.xml}"
export UCRP_E2E_DRIFT_REPORT="${UCRP_E2E_DRIFT_REPORT:-$api_root/e2e-contract-drift.md}"

if [ "$report_only" -eq 1 ]; then
    export UCRP_E2E_STRICT_CONTRACT=0
else
    export UCRP_E2E_STRICT_CONTRACT=1
fi

cd "$api_root"
go test ./e2e/... -count=1 -v
