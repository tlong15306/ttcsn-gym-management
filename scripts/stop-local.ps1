[CmdletBinding()]
param([switch]$KeepDatabase)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$runtimeRoot = Join-Path $projectRoot '.artifact_work\runtime'

function Stop-TrackedProcess {
    param(
        [string]$MetadataPath,
        [string]$ExpectedCommandFragment
    )

    if (-not (Test-Path $MetadataPath)) { return }
    $metadata = Get-Content $MetadataPath -Raw | ConvertFrom-Json
    $process = Get-CimInstance Win32_Process -Filter "ProcessId = $($metadata.process_id)" -ErrorAction SilentlyContinue
    if ($null -eq $process) {
        Remove-Item $MetadataPath -Force -ErrorAction SilentlyContinue
        return
    }
    if ([string]$process.CommandLine -notlike "*$ExpectedCommandFragment*") {
        Write-Warning "Refusing to stop PID $($metadata.process_id): it is not the tracked TTCSN process."
        return
    }
    Stop-Process -Id $metadata.process_id -Force
    Remove-Item $MetadataPath -Force -ErrorAction SilentlyContinue
}

Stop-TrackedProcess -MetadataPath (Join-Path $runtimeRoot 'client.process.json') -ExpectedCommandFragment 'node_modules\vite\bin\vite.js'
Stop-TrackedProcess -MetadataPath (Join-Path $runtimeRoot 'api.process.json') -ExpectedCommandFragment 'ttcsn-gym-api.exe'

if (-not $KeepDatabase) {
    $composeEnv = Join-Path $projectRoot 'infra\.env.compose'
    if (Test-Path $composeEnv) {
        Push-Location $projectRoot
        try {
            & docker compose --env-file $composeEnv stop postgres
        } finally {
            Pop-Location
        }
    }
}

Write-Host 'Tracked TTCSN API/client processes stopped. PostgreSQL data volume was preserved.'
