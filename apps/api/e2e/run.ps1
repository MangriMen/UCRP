#!/usr/bin/env pwsh
# Runs the UCRP E2E suite against a live API instance.
# Usage: ./apps/api/e2e/run.ps1 [-ReportOnly]

param(
    [switch]$ReportOnly
)

$ErrorActionPreference = 'Stop'

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$apiRoot = Join-Path $repoRoot 'apps/api'

function Set-DefaultEnv([string]$Name, [string]$Value) {
    if (-not [Environment]::GetEnvironmentVariable($Name)) {
        [Environment]::SetEnvironmentVariable($Name, $Value, 'Process')
    }
}

Set-DefaultEnv 'UCRP_E2E_BASE_URL' 'http://127.0.0.1:8080'
Set-DefaultEnv 'UCRP_E2E_DATABASE_URL' 'postgres://ucrp:ucrp@127.0.0.1:5543/ucrp?sslmode=disable'
Set-DefaultEnv 'UCRP_E2E_ADMIN_EMAIL' '[EMAIL]'
Set-DefaultEnv 'UCRP_E2E_ADMIN_PASSWORD' 'admin12345'
Set-DefaultEnv 'UCRP_E2E_OPENAPI' (Join-Path $repoRoot 'docs/swagger/openapi.json')
Set-DefaultEnv 'UCRP_E2E_JUNIT' (Join-Path $apiRoot 'e2e-report.xml')
Set-DefaultEnv 'UCRP_E2E_DRIFT_REPORT' (Join-Path $apiRoot 'e2e-contract-drift.md')

if ($ReportOnly) {
    $env:UCRP_E2E_STRICT_CONTRACT = '0'
} else {
    $env:UCRP_E2E_STRICT_CONTRACT = '1'
}

Push-Location $apiRoot
try {
    go test ./e2e/... -count=1 -v
    $exit = $LASTEXITCODE
} finally {
    Pop-Location
}

Write-Host ''
Write-Host "JUnit:  $env:UCRP_E2E_JUNIT"
Write-Host "Drift:  $env:UCRP_E2E_DRIFT_REPORT"
exit $exit
