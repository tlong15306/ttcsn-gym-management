[CmdletBinding()]
param(
    [string]$BaseUrl = 'http://127.0.0.1:8080/api/v1',
    [string]$OutputDirectory = '.artifact_work\api-smoke'
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$collection = Join-Path $projectRoot 'docs\GYM_Role_Smoke.postman_collection.json'
$outputRoot = Join-Path $projectRoot $OutputDirectory
New-Item -ItemType Directory -Force -Path $outputRoot | Out-Null

try {
    Invoke-WebRequest -UseBasicParsing -Uri (($BaseUrl -replace '/api/v1$', '') + '/health/ready') -TimeoutSec 3 | Out-Null
} catch {
    throw "API is not ready at $BaseUrl. Run scripts/start-local.ps1 first."
}

$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$reportName = "newman-role-smoke-$timestamp.json"
$reportPath = Join-Path $outputRoot $reportName
$containerBaseUrl = $BaseUrl -replace '127\.0\.0\.1', 'host.docker.internal' -replace 'localhost', 'host.docker.internal'
Push-Location $projectRoot
try {
    & docker version --format '{{.Server.Version}}' | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Docker Desktop Linux engine is unavailable.' }
    & docker run --rm `
        -v "$projectRoot\docs:/etc/newman:ro" `
        -v "${outputRoot}:/output" `
        postman/newman:alpine run /etc/newman/GYM_Role_Smoke.postman_collection.json `
        --env-var "base_url=$containerBaseUrl" `
        --reporters cli,json `
        --reporter-json-export "/output/$reportName"
    if ($LASTEXITCODE -ne 0) { throw "Newman assertions failed. Report: $reportPath" }
} finally {
    Pop-Location
}

Write-Host "PASS role smoke. JSON report: $reportPath"
