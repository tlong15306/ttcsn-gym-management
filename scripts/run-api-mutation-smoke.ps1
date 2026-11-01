[CmdletBinding()]
param(
    [ValidateRange(1025, 65535)]
    [int]$Port = 18080,
    [string]$OutputDirectory = '.artifact_work\api-mutation-smoke'
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$serverRoot = Join-Path $projectRoot 'server'
$collection = Join-Path $projectRoot 'docs\GYM_Mutation_Smoke.postman_collection.json'
$composeEnv = Join-Path $projectRoot 'infra\.env.compose'
$outputRoot = Join-Path $projectRoot $OutputDirectory
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$apiProcess = $null

function Test-Listener {
    param([int]$LocalPort)
    return $null -ne (Get-NetTCPConnection -State Listen -LocalPort $LocalPort -ErrorAction SilentlyContinue | Select-Object -First 1)
}

function Wait-HttpOk {
    param([string]$Uri, [int]$Attempts = 30)
    for ($i = 1; $i -le $Attempts; $i++) {
        try {
            $response = Invoke-WebRequest -UseBasicParsing -Uri $Uri -TimeoutSec 2
            if ($response.StatusCode -eq 200) { return }
        } catch {
            if ($i -eq $Attempts) { break }
            Start-Sleep -Seconds 1
        }
    }
    throw "Test API did not become ready at $Uri."
}

if (-not (Test-Path $collection) -or -not (Test-Path $composeEnv)) {
    throw 'Mutation collection or local Compose configuration is missing.'
}
if (Test-Listener $Port) {
    throw "Port $Port is already listening. Choose a dedicated port for the isolated smoke API."
}

& docker version --format '{{.Server.Version}}' | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'Docker Desktop Linux engine is unavailable.' }

$composeText = Get-Content $composeEnv -Raw
$composePassword = ($composeText -split '=', 2)[1].Trim()
$encodedPassword = [Uri]::EscapeDataString($composePassword)
$env:DATABASE_URL = "postgres://gym_app:$encodedPassword@127.0.0.1:5433/ttcsn_gym_test?sslmode=disable"
$bytes = [byte[]]::new(32)
$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
try { $rng.GetBytes($bytes) } finally { $rng.Dispose() }
$env:JWT_SECRET = -join ($bytes | ForEach-Object { $_.ToString('x2') })
$env:HTTP_ADDR = "127.0.0.1:$Port"
$env:APP_ENV = 'test'
$env:CORS_ORIGINS = 'http://127.0.0.1:5173'
$env:UPLOAD_DIR = Join-Path $projectRoot '.artifact_work\api-mutation-smoke-private'
$env:GOCACHE = Join-Path $projectRoot '.artifact_work\go-cache'

New-Item -ItemType Directory -Force -Path $outputRoot | Out-Null
$apiPath = Join-Path $outputRoot "ttcsn-gym-mutation-api-$stamp.exe"
$apiLog = Join-Path $outputRoot "api-mutation-$stamp.stdout.log"
$apiErrorLog = Join-Path $outputRoot "api-mutation-$stamp.stderr.log"
$newmanName = "newman-mutation-smoke-$stamp.json"
$newmanPath = Join-Path $outputRoot $newmanName
$summaryPath = Join-Path $outputRoot "mutation-smoke-$stamp.summary.json"

Push-Location $serverRoot
try {
    & go run .\cmd\reset-test
    if ($LASTEXITCODE -ne 0) { throw 'Could not reset ttcsn_gym_test before mutation smoke.' }
    & go run .\cmd\seed
    if ($LASTEXITCODE -ne 0) { throw 'Could not seed isolated mutation smoke database.' }
    & go build -o $apiPath .\cmd\api
    if ($LASTEXITCODE -ne 0) { throw 'Could not build isolated mutation smoke API.' }
} finally {
    Pop-Location
}

try {
    $apiProcess = Start-Process -FilePath $apiPath -WorkingDirectory $serverRoot -RedirectStandardOutput $apiLog -RedirectStandardError $apiErrorLog -WindowStyle Hidden -PassThru
    Wait-HttpOk -Uri "http://127.0.0.1:$Port/health/ready"

    $containerBaseUrl = "http://host.docker.internal:$Port/api/v1"
    & docker run --rm `
        -v "$projectRoot\docs:/etc/newman:ro" `
        -v "${outputRoot}:/output" `
        postman/newman:alpine run /etc/newman/GYM_Mutation_Smoke.postman_collection.json `
        --env-var "base_url=$containerBaseUrl" `
        --reporters cli,json `
        --reporter-json-export "/output/$newmanName"
    if ($LASTEXITCODE -ne 0) { throw "Newman mutation assertions failed. Report: $newmanPath" }

    $auditCount = (& docker exec ttcsn-gym-postgres psql -U gym_app -d ttcsn_gym_test -At -c "SELECT count(*) FROM audit_events WHERE action IN ('account.created','account.status_changed','account.password_reset');").Trim()
    if ($LASTEXITCODE -ne 0 -or [int]$auditCount -lt 4) {
        throw "Expected account mutation audit events, found $auditCount."
    }
    [ordered]@{
        created_at_local = (Get-Date).ToString('o')
        database = 'ttcsn_gym_test'
        api_base = "http://127.0.0.1:$Port/api/v1"
        newman_report = $newmanPath
        account_mutation_audits = [int]$auditCount
        cleanup = 'ttcsn_gym_test reset after run'
    } | ConvertTo-Json | Set-Content -LiteralPath $summaryPath -Encoding UTF8
    Write-Host "PASS mutation smoke. Summary: $summaryPath"
} finally {
    if ($apiProcess -and -not $apiProcess.HasExited) {
        Stop-Process -Id $apiProcess.Id -Force
        $apiProcess.WaitForExit()
    }
    Push-Location $serverRoot
    try {
        & go run .\cmd\reset-test
        if ($LASTEXITCODE -ne 0) { throw 'Could not reset ttcsn_gym_test after mutation smoke.' }
    } finally {
        Pop-Location
    }
}
