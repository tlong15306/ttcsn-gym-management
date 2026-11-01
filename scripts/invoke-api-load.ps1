[CmdletBinding()]
param(
    [ValidateRange(1, 200)]
    [int]$Concurrency = 100,
    [ValidateRange(1, 10)]
    [int]$RequestsPerUser = 3,
    [ValidateRange(0, 300)]
    [int]$RampSeconds = 0,
    [switch]$WarmSessions,
    [string]$ApiBase = 'http://127.0.0.1:8080/api/v1',
    [string]$OutputPath
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
if (-not $OutputPath) {
    $outputRoot = Join-Path $projectRoot '.artifact_work\load'
    New-Item -ItemType Directory -Force -Path $outputRoot | Out-Null
    $OutputPath = Join-Path $outputRoot ('api-load-' + (Get-Date -Format 'yyyyMMdd_HHmmss') + '.json')
}

try {
    $ready = Invoke-WebRequest -UseBasicParsing -Uri ($ApiBase -replace '/api/v1$', '/health/ready') -TimeoutSec 3
    if ($ready.StatusCode -ne 200) { throw 'unexpected readiness status' }
} catch {
    throw "API is not ready at $ApiBase. Start the project first."
}

$worker = {
    param($base, $username, $password, $role, $requestCount, $preissuedToken)
    function Invoke-LoadRequest {
        param($phase, $method, $uri, $headers, $body)
        $watch = [System.Diagnostics.Stopwatch]::StartNew()
        try {
            $arguments = @{ Uri = $uri; Method = $method; UseBasicParsing = $true; TimeoutSec = 20 }
            if ($headers) { $arguments.Headers = $headers }
            if ($body) { $arguments.ContentType = 'application/json'; $arguments.Body = $body }
            $response = Invoke-WebRequest @arguments
            [pscustomobject]@{ phase = $phase; role = $role; status = [int]$response.StatusCode; milliseconds = [math]::Round($watch.Elapsed.TotalMilliseconds, 2) }
        } catch {
            $status = 0
            if ($_.Exception.Response) {
                try { $status = [int]$_.Exception.Response.StatusCode } catch { $status = 0 }
            }
            [pscustomobject]@{ phase = $phase; role = $role; status = $status; milliseconds = [math]::Round($watch.Elapsed.TotalMilliseconds, 2) }
        }
    }

    $result = @()
    $token = $preissuedToken
    if (-not $token) {
        $loginBody = @{ username = $username; password = $password } | ConvertTo-Json -Compress
        $watch = [System.Diagnostics.Stopwatch]::StartNew()
        try {
            $loginResponse = Invoke-WebRequest -UseBasicParsing -Uri "$base/auth/login" -Method POST -ContentType 'application/json' -Body $loginBody -TimeoutSec 20
            $token = (($loginResponse.Content | ConvertFrom-Json).data.access_token)
            $result += [pscustomobject]@{ phase = 'login'; role = $role; status = [int]$loginResponse.StatusCode; milliseconds = [math]::Round($watch.Elapsed.TotalMilliseconds, 2) }
        } catch {
            $status = 0
            if ($_.Exception.Response) {
                try { $status = [int]$_.Exception.Response.StatusCode } catch { $status = 0 }
            }
            return @([pscustomobject]@{ phase = 'login'; role = $role; status = $status; milliseconds = [math]::Round($watch.Elapsed.TotalMilliseconds, 2) })
        }
    }
    $headers = @{ Authorization = "Bearer $token" }
    $paths = @('me', 'rooms?page_size=20', 'bookings?page_size=20')
    if ($role -eq 'ADMIN' -or $role -eq 'MANAGER') { $paths[2] = 'reports/operations' }
    for ($index = 0; $index -lt $requestCount; $index++) {
        $result += Invoke-LoadRequest -phase 'read' -method 'GET' -uri "$base/$($paths[$index % $paths.Count])" -headers $headers -body $null
    }
    return $result
}

$accounts = @(
    @{ username = 'admin.demo'; password = 'DemoPass123!'; role = 'ADMIN' },
    @{ username = 'manager.demo'; password = 'DemoPass123!'; role = 'MANAGER' },
    @{ username = 'reception.demo'; password = 'DemoPass123!'; role = 'RECEPTIONIST' },
    @{ username = 'pt.demo'; password = 'DemoPass123!'; role = 'PT' },
    @{ username = 'member.demo'; password = 'DemoPass123!'; role = 'MEMBER' }
)

# Warmed sessions isolate the authenticated-read phase from the CPU cost of a
# simultaneous bcrypt burst. Tokens remain only in-memory and are omitted from
# the JSON report. The default remains the all-at-once login + read workload.
$warmupElapsedSeconds = 0.0
$issuedSessions = @()
if ($WarmSessions) {
    $warmupStarted = Get-Date
    for ($index = 0; $index -lt $Concurrency; $index++) {
        $account = $accounts[$index % $accounts.Count]
        $loginBody = @{ username = $account.username; password = $account.password } | ConvertTo-Json -Compress
        try {
            $response = Invoke-WebRequest -UseBasicParsing -Uri "$ApiBase/auth/login" -Method POST -ContentType 'application/json' -Body $loginBody -TimeoutSec 20
            $token = (($response.Content | ConvertFrom-Json).data.access_token)
        } catch {
            throw "Warm-session login failed for role $($account.role)."
        }
        if ([string]::IsNullOrWhiteSpace($token)) {
            throw "Warm-session login returned no access token for role $($account.role)."
        }
        $issuedSessions += [pscustomobject]@{ username = $account.username; role = $account.role; token = $token }
    }
    $warmupElapsedSeconds = [math]::Round(((Get-Date) - $warmupStarted).TotalSeconds, 2)
}

$pool = [RunspaceFactory]::CreateRunspacePool(1, $Concurrency)
$pool.Open()
$workers = @()
$started = Get-Date
try {
    for ($index = 0; $index -lt $Concurrency; $index++) {
        $account = $accounts[$index % $accounts.Count]
        $preissuedToken = $null
        if ($WarmSessions) { $preissuedToken = $issuedSessions[$index].token }
        $shell = [PowerShell]::Create()
        $shell.RunspacePool = $pool
        [void]$shell.AddScript($worker).AddArgument($ApiBase).AddArgument($account.username).AddArgument($account.password).AddArgument($account.role).AddArgument($RequestsPerUser).AddArgument($preissuedToken)
        $workers += [pscustomobject]@{ shell = $shell; handle = $shell.BeginInvoke() }
        if ($RampSeconds -gt 0 -and $index -lt ($Concurrency - 1)) {
            $milliseconds = [int][math]::Round(($RampSeconds * 1000.0) / [math]::Max($Concurrency - 1, 1))
            if ($milliseconds -gt 0) { Start-Sleep -Milliseconds $milliseconds }
        }
    }
    $results = @()
    foreach ($entry in $workers) {
        $results += @($entry.shell.EndInvoke($entry.handle))
        $entry.shell.Dispose()
        $entry.shell = $null
    }
} finally {
    foreach ($entry in $workers) {
        if ($entry.shell) { $entry.shell.Dispose() }
    }
    $pool.Close()
    $pool.Dispose()
}
$elapsedSeconds = [math]::Max(((Get-Date) - $started).TotalSeconds, 0.001)
$latencies = @($results | Where-Object { $_.milliseconds -gt 0 } | ForEach-Object { [double]$_.milliseconds } | Sort-Object)
function Get-Percentile {
    param([double[]]$Values, [double]$Percentile)
    if ($Values.Count -eq 0) { return $null }
    $position = [math]::Max(0, [math]::Ceiling($Values.Count * $Percentile) - 1)
    return [math]::Round($Values[$position], 2)
}
$phasePercentiles = @($results | Group-Object phase | ForEach-Object {
    $phaseLatencies = @($_.Group | Where-Object { $_.milliseconds -gt 0 } | ForEach-Object { [double]$_.milliseconds } | Sort-Object)
    [ordered]@{
        name = $_.Name
        requests = $_.Count
        p50_milliseconds = Get-Percentile -Values $phaseLatencies -Percentile 0.50
        p95_milliseconds = Get-Percentile -Values $phaseLatencies -Percentile 0.95
        p99_milliseconds = Get-Percentile -Values $phaseLatencies -Percentile 0.99
    }
})

$summary = [ordered]@{
    started_at_local = $started.ToString('o')
    api_base = $ApiBase
    concurrency = $Concurrency
    requests_per_user = $RequestsPerUser
    ramp_seconds = $RampSeconds
    warm_sessions = [bool]$WarmSessions
    warmup_logins = if ($WarmSessions) { $Concurrency } else { 0 }
    warmup_elapsed_seconds = $warmupElapsedSeconds
    total_requests = @($results).Count
    successful_2xx = @($results | Where-Object { $_.status -ge 200 -and $_.status -lt 300 }).Count
    client_4xx = @($results | Where-Object { $_.status -ge 400 -and $_.status -lt 500 }).Count
    server_5xx_or_network = @($results | Where-Object { $_.status -eq 0 -or $_.status -ge 500 }).Count
    elapsed_seconds = [math]::Round($elapsedSeconds, 2)
    throughput_requests_per_second = [math]::Round((@($results).Count / $elapsedSeconds), 2)
    p50_milliseconds = Get-Percentile -Values $latencies -Percentile 0.50
    p95_milliseconds = Get-Percentile -Values $latencies -Percentile 0.95
    p99_milliseconds = Get-Percentile -Values $latencies -Percentile 0.99
    phases = @($results | Group-Object phase | ForEach-Object {
        [ordered]@{ name = $_.Name; requests = $_.Count; successful_2xx = @($_.Group | Where-Object { $_.status -ge 200 -and $_.status -lt 300 }).Count }
    })
    phase_percentiles = $phasePercentiles
}

[ordered]@{ summary = $summary; results = $results } | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $OutputPath -Encoding UTF8
$summary | Format-List
Write-Host "Load report: $OutputPath"
if ($summary.server_5xx_or_network -gt 0 -or $summary.client_4xx -gt 0) { exit 1 }
