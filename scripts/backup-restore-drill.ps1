[CmdletBinding()]
param(
    [ValidateSet('ttcsn_gym_dev', 'ttcsn_gym_test')]
    [string]$SourceDatabase = 'ttcsn_gym_dev',
    [string]$RestoreDatabase,
    [string]$BackupRoot
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$container = 'ttcsn-gym-postgres'
$stamp = Get-Date -Format 'yyyyMMdd_HHmmss'

if (-not $RestoreDatabase) {
    $RestoreDatabase = "ttcsn_gym_restore_$stamp"
}
if ($RestoreDatabase -notmatch '^ttcsn_gym_restore_[a-z0-9_]+$') {
    throw 'RestoreDatabase must start with ttcsn_gym_restore_ and contain only lowercase letters, numbers, and underscores.'
}
if ($RestoreDatabase -eq $SourceDatabase) {
    throw 'RestoreDatabase must be different from SourceDatabase.'
}
if (-not $BackupRoot) {
    $BackupRoot = Join-Path $projectRoot '.artifact_work\backup-drills'
}
if (-not (Test-Path (Join-Path $projectRoot 'infra\.env.compose'))) {
    throw 'Missing infra/.env.compose. Create the local Compose configuration before running a backup drill.'
}

& docker version --format '{{.Server.Version}}' | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw 'Docker Desktop Linux engine is unavailable. Open Docker Desktop and wait until it is running.'
}
$running = & docker inspect --format '{{.State.Running}}' $container 2>$null
if ($LASTEXITCODE -ne 0 -or $running.Trim() -ne 'true') {
    throw "PostgreSQL container $container is not running. Start it with docker compose --env-file .\infra\.env.compose up -d postgres."
}

function Invoke-DatabaseScalar {
    param([string]$Database, [string]$Query)
    $value = & docker exec $container psql -U gym_app -d $Database -At -c $Query
    if ($LASTEXITCODE -ne 0) {
        throw "PostgreSQL query failed for $Database."
    }
    return ($value | Out-String).Trim()
}

function Get-DatabaseSummary {
    param([string]$Database)
    $query = @"
SELECT concat_ws('|',
  (SELECT max(version_id)::text FROM goose_db_version WHERE is_applied),
  (SELECT count(*)::text FROM users),
  (SELECT count(*)::text FROM rooms),
  (SELECT count(*)::text FROM bookings),
  (SELECT count(*)::text FROM audit_events),
  (SELECT count(*)::text FROM rooms WHERE cover_image IS NOT NULL)
);
"@
    $parts = (Invoke-DatabaseScalar -Database $Database -Query $query).Split('|')
    if ($parts.Count -ne 6) {
        throw "Unexpected summary from $Database."
    }
    return [ordered]@{
        schema_version = [int]$parts[0]
        users = [int]$parts[1]
        rooms = [int]$parts[2]
        bookings = [int]$parts[3]
        audit_events = [int]$parts[4]
        rooms_with_cover = [int]$parts[5]
    }
}

$existing = Invoke-DatabaseScalar -Database 'postgres' -Query "SELECT 1 FROM pg_database WHERE datname='$RestoreDatabase';"
if ($existing -eq '1') {
    throw "Refusing to overwrite existing restore database $RestoreDatabase. Choose a new name or remove it manually after verifying it is safe."
}

New-Item -ItemType Directory -Force -Path $BackupRoot | Out-Null
$archiveName = "${SourceDatabase}_$stamp.dump"
$containerArchive = "/tmp/$archiveName"
$dumpPath = Join-Path $BackupRoot $archiveName
$manifestPath = Join-Path $BackupRoot "${SourceDatabase}_$stamp.manifest.json"
$restoreUploadRoot = Join-Path $BackupRoot "${RestoreDatabase}_private_uploads"
$sourceSummary = Get-DatabaseSummary -Database $SourceDatabase

try {
    & docker exec $container pg_dump -U gym_app -d $SourceDatabase --format=custom --no-owner --no-privileges --file $containerArchive
    if ($LASTEXITCODE -ne 0) { throw "pg_dump failed for $SourceDatabase." }
    & docker cp "${container}:$containerArchive" $dumpPath
    if ($LASTEXITCODE -ne 0) { throw 'Could not copy PostgreSQL archive from the container.' }
} finally {
    & docker exec $container rm -f $containerArchive 2>$null | Out-Null
}

$uploadRoot = Join-Path $projectRoot 'uploads\private'
$uploadArchive = $null
if (Test-Path $uploadRoot) {
    $uploadFiles = @(Get-ChildItem -LiteralPath $uploadRoot -File -Recurse -Force)
    if ($uploadFiles.Count -gt 0) {
        $uploadArchive = Join-Path $BackupRoot "${SourceDatabase}_$stamp.private-uploads.zip"
        Compress-Archive -Path (Join-Path $uploadRoot '*') -DestinationPath $uploadArchive -CompressionLevel Optimal -Force
    }
}

try {
    & docker exec $container createdb -U gym_app $RestoreDatabase
    if ($LASTEXITCODE -ne 0) { throw "Could not create isolated restore database $RestoreDatabase." }
    & docker cp $dumpPath "${container}:$containerArchive"
    if ($LASTEXITCODE -ne 0) { throw 'Could not copy PostgreSQL archive into the container.' }
    & docker exec $container pg_restore -U gym_app -d $RestoreDatabase --exit-on-error --no-owner --no-privileges $containerArchive
    if ($LASTEXITCODE -ne 0) { throw "pg_restore failed for $RestoreDatabase." }
} finally {
    & docker exec $container rm -f $containerArchive 2>$null | Out-Null
}

if ($uploadArchive) {
    New-Item -ItemType Directory -Force -Path $restoreUploadRoot | Out-Null
    Expand-Archive -LiteralPath $uploadArchive -DestinationPath $restoreUploadRoot -Force
}

$restoreSummary = Get-DatabaseSummary -Database $RestoreDatabase
foreach ($name in $sourceSummary.Keys) {
    if ($sourceSummary[$name] -ne $restoreSummary[$name]) {
        throw "Restore verification mismatch for ${name}: source=$($sourceSummary[$name]) restore=$($restoreSummary[$name])."
    }
}

$manifest = [ordered]@{
    created_at_local = (Get-Date).ToString('o')
    source_database = $SourceDatabase
    restore_database = $RestoreDatabase
    archive_path = $dumpPath
    archive_sha256 = (Get-FileHash -LiteralPath $dumpPath -Algorithm SHA256).Hash
    archive_bytes = (Get-Item -LiteralPath $dumpPath).Length
    source_summary = $sourceSummary
    restore_summary = $restoreSummary
    private_uploads_archive = $uploadArchive
    private_uploads_restored_to = if ($uploadArchive) { $restoreUploadRoot } else { $null }
}
$manifest | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $manifestPath -Encoding UTF8

Write-Host "Backup archive: $dumpPath"
Write-Host "Manifest:       $manifestPath"
Write-Host "Restored to:    $RestoreDatabase"
Write-Host 'Restore target is retained for independent inspection. Do not use it as the development database.'
