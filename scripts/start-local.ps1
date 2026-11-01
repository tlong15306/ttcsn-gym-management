[CmdletBinding()]
param(
    [switch]$SkipSeed,
    [switch]$WithBookingFixtures,
    [switch]$NoClient,
    [switch]$ClientOnly
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$serverRoot = Join-Path $projectRoot 'server'
$clientRoot = Join-Path $projectRoot 'client'
$runtimeRoot = Join-Path $projectRoot '.artifact_work\runtime'
$composeEnv = Join-Path $projectRoot 'infra\.env.compose'
$serverConfig = Join-Path $serverRoot '.env.local'

function Test-Listener {
    param([int]$Port)
    return $null -ne (Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue | Select-Object -First 1)
}

function Wait-HttpOk {
    param(
        [string]$Url,
        [string]$Label,
        [int]$Attempts = 30
    )

    for ($i = 1; $i -le $Attempts; $i++) {
        try {
            $response = Invoke-WebRequest -UseBasicParsing -Uri $Url -TimeoutSec 2
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 300) { return }
        } catch {
            if ($i -eq $Attempts) { break }
            Start-Sleep -Seconds 1
        }
    }
    throw "$Label did not become ready at $Url. See .artifact_work/runtime logs."
}

if ($ClientOnly) {
    if (-not (Test-Path (Join-Path $clientRoot 'node_modules\vite\bin\vite.js'))) {
        throw 'Missing client/node_modules. Run npm.cmd ci from client before using this script.'
    }
    if (Test-Listener 5173) { throw 'Port 5173 is already listening. Stop the existing project client before starting a new one.' }

    New-Item -ItemType Directory -Force -Path $runtimeRoot | Out-Null
    $nodeExecutable = (Get-Command node.exe -ErrorAction Stop).Source
    $viteScript = Join-Path $clientRoot 'node_modules\vite\bin\vite.js'
    $clientLog = Join-Path $runtimeRoot 'client.stdout.log'
    $clientErrorLog = Join-Path $runtimeRoot 'client.stderr.log'
    Remove-Item $clientLog,$clientErrorLog -Force -ErrorAction SilentlyContinue
    $clientProcess = Start-Process -FilePath $nodeExecutable -ArgumentList ('"' + $viteScript + '" --host 127.0.0.1 --port 5173 --strictPort') -WorkingDirectory $clientRoot -RedirectStandardOutput $clientLog -RedirectStandardError $clientErrorLog -WindowStyle Hidden -PassThru
    @{ process_id = $clientProcess.Id; executable = $nodeExecutable; script = $viteScript; started_utc = (Get-Date).ToUniversalTime().ToString('o') } |
        ConvertTo-Json | Set-Content (Join-Path $runtimeRoot 'client.process.json')
    Wait-HttpOk -Url 'http://127.0.0.1:5173/login' -Label 'Client'
    Write-Host 'Client: http://127.0.0.1:5173/login'
    return
}

if (-not (Test-Path $composeEnv)) {
    throw "Missing $composeEnv. Copy infra/.env.compose.example to infra/.env.compose and set POSTGRES_PASSWORD."
}
if (-not (Test-Path $serverConfig)) {
    throw "Missing $serverConfig. Follow docs/17_huong_dan_chay_local.md to create local API configuration."
}
if (-not (Test-Path (Join-Path $clientRoot 'node_modules\vite\bin\vite.js')) -and -not $NoClient) {
    throw 'Missing client/node_modules. Run npm.cmd ci from client before using this script.'
}

& docker version --format '{{.Server.Version}}' | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw 'Docker Desktop Linux engine is unavailable. Open Docker Desktop and wait until it is running.'
}

Push-Location $projectRoot
try {
    & docker compose --env-file $composeEnv up -d --wait postgres
    if ($LASTEXITCODE -ne 0) { throw 'PostgreSQL startup failed.' }
} finally {
    Pop-Location
}

if (Test-Listener 8080) { throw 'Port 8080 is already listening. Stop the existing project API before starting a new one.' }
if (-not $NoClient -and (Test-Listener 5173)) { throw 'Port 5173 is already listening. Stop the existing project client before starting a new one.' }

New-Item -ItemType Directory -Force -Path $runtimeRoot | Out-Null
$env:GOCACHE = Join-Path $projectRoot '.artifact_work\go-cache'

Push-Location $serverRoot
try {
    & go run .\cmd\migrate up
    if ($LASTEXITCODE -ne 0) { throw 'Migration failed.' }
    if (-not $SkipSeed) {
        & go run .\cmd\seed
        if ($LASTEXITCODE -ne 0) { throw 'Account seed failed.' }
    }
    if ($WithBookingFixtures) {
        & go run .\cmd\seed-booking-demo
        if ($LASTEXITCODE -ne 0) { throw 'Booking fixture seed failed.' }
    }
    $apiExecutable = Join-Path $runtimeRoot 'ttcsn-gym-api.exe'
    & go build -o $apiExecutable .\cmd\api
    if ($LASTEXITCODE -ne 0) { throw 'API build failed.' }
} finally {
    Pop-Location
}

$apiLog = Join-Path $runtimeRoot 'api.stdout.log'
$apiErrorLog = Join-Path $runtimeRoot 'api.stderr.log'
Remove-Item $apiLog,$apiErrorLog -Force -ErrorAction SilentlyContinue
$apiProcess = Start-Process -FilePath $apiExecutable -WorkingDirectory $serverRoot -RedirectStandardOutput $apiLog -RedirectStandardError $apiErrorLog -WindowStyle Hidden -PassThru
@{ process_id = $apiProcess.Id; executable = $apiExecutable; started_utc = (Get-Date).ToUniversalTime().ToString('o') } |
    ConvertTo-Json | Set-Content (Join-Path $runtimeRoot 'api.process.json')

try {
    Wait-HttpOk -Url 'http://127.0.0.1:8080/health/ready' -Label 'API'
    if (-not $NoClient) {
        $nodeExecutable = (Get-Command node.exe -ErrorAction Stop).Source
        $viteScript = Join-Path $clientRoot 'node_modules\vite\bin\vite.js'
        $clientLog = Join-Path $runtimeRoot 'client.stdout.log'
        $clientErrorLog = Join-Path $runtimeRoot 'client.stderr.log'
        Remove-Item $clientLog,$clientErrorLog -Force -ErrorAction SilentlyContinue
        $clientProcess = Start-Process -FilePath $nodeExecutable -ArgumentList ('"' + $viteScript + '" --host 127.0.0.1 --port 5173 --strictPort') -WorkingDirectory $clientRoot -RedirectStandardOutput $clientLog -RedirectStandardError $clientErrorLog -WindowStyle Hidden -PassThru
        @{ process_id = $clientProcess.Id; executable = $nodeExecutable; script = $viteScript; started_utc = (Get-Date).ToUniversalTime().ToString('o') } |
            ConvertTo-Json | Set-Content (Join-Path $runtimeRoot 'client.process.json')
        Wait-HttpOk -Url 'http://127.0.0.1:5173/login' -Label 'Client'
    }
} catch {
    & (Join-Path $PSScriptRoot 'stop-local.ps1') -KeepDatabase
    throw
}

Write-Host 'Local GYM is ready.'
Write-Host 'API:    http://127.0.0.1:8080/health/ready'
if (-not $NoClient) { Write-Host 'Client: http://127.0.0.1:5173/login' }
Write-Host 'Stop:   powershell -ExecutionPolicy Bypass -File .\scripts\stop-local.ps1'
