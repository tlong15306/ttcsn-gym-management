[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$docsRoot = $PSScriptRoot
$required = @(
    'README.md',
    '00_quyet_dinh_va_quy_tac_nghiep_vu.md',
    '03_tac_nhan.md',
    '04_use_case_tong_quat.md',
    '06_ma_tran_phan_quyen.md',
    '07_SRS_v0.3.md',
    '08_use_case_chi_tiet.md',
    '09_ERD_va_CSDL.md',
    '09_schema.sql',
    '11_kien_truc_he_thong.md',
    '12_dac_ta_API.md',
    '13_backlog_v0.3.md',
    '15_ke_hoach_kiem_thu.md',
    '16_runbook_van_hanh_nang_cap.md',
    'CHANGELOG.md'
)

$errors = [System.Collections.Generic.List[string]]::new()

foreach ($name in $required) {
    $path = Join-Path $docsRoot $name
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
        $errors.Add("Missing required file: $name")
        continue
    }

    $text = Get-Content -LiteralPath $path -Raw -Encoding UTF8
    if ($text.Contains([char]0xFFFD)) {
        $errors.Add("Invalid replacement character in: $name")
    }

    if ($name.EndsWith('.md')) {
        $fenceCount = @((Get-Content -LiteralPath $path -Encoding UTF8) | Where-Object { $_ -match '^```' }).Count
        if ($fenceCount % 2 -ne 0) {
            $errors.Add("Unbalanced Markdown fences in: $name")
        }
    }
}

$srs = Get-Content -LiteralPath (Join-Path $docsRoot '07_SRS_v0.3.md') -Raw -Encoding UTF8
$catalog = Get-Content -LiteralPath (Join-Path $docsRoot '04_use_case_tong_quat.md') -Raw -Encoding UTF8
$detail = Get-Content -LiteralPath (Join-Path $docsRoot '08_use_case_chi_tiet.md') -Raw -Encoding UTF8
$backlog = Get-Content -LiteralPath (Join-Path $docsRoot '13_backlog_v0.3.md') -Raw -Encoding UTF8
$rules = Get-Content -LiteralPath (Join-Path $docsRoot '00_quyet_dinh_va_quy_tac_nghiep_vu.md') -Raw -Encoding UTF8
$schema = Get-Content -LiteralPath (Join-Path $docsRoot '09_schema.sql') -Raw -Encoding UTF8

$srsReq = @([regex]::Matches($srs, '(?m)^### REQ-(\d{2})') | ForEach-Object { $_.Groups[1].Value } | Sort-Object -Unique)
$backlogReq = @([regex]::Matches($backlog, '(?m)^\| REQ-(\d{2}) \|') | ForEach-Object { $_.Groups[1].Value } | Sort-Object -Unique)
$catalogUc = @([regex]::Matches($catalog, '(?m)^\| UC-(\d{2}) \|') | ForEach-Object { $_.Groups[1].Value } | Sort-Object -Unique)
$detailUc = @([regex]::Matches($detail, '(?m)^## UC-(\d{2})') | ForEach-Object { $_.Groups[1].Value } | Sort-Object -Unique)
$businessRules = @([regex]::Matches($rules, '(?m)^- \*\*BR-(\d{2})') | ForEach-Object { $_.Groups[1].Value } | Sort-Object -Unique)

if ($srsReq.Count -ne 20) { $errors.Add("Expected 20 SRS requirements; found $($srsReq.Count)") }
if ($backlogReq.Count -ne 20) { $errors.Add("Expected 20 backlog requirements; found $($backlogReq.Count)") }
if (Compare-Object $srsReq $backlogReq) { $errors.Add('SRS and backlog REQ IDs differ') }
if ($catalogUc.Count -ne 16) { $errors.Add("Expected 16 catalog use cases; found $($catalogUc.Count)") }
if ($detailUc.Count -ne 16) { $errors.Add("Expected 16 detailed use cases; found $($detailUc.Count)") }
if (Compare-Object $catalogUc $detailUc) { $errors.Add('Catalog and detailed UC IDs differ') }
if ($businessRules.Count -ne 26) { $errors.Add("Expected 26 business rules; found $($businessRules.Count)") }
if ($schema -match '(?mi)^\s*DROP\s+TABLE') { $errors.Add('Bootstrap schema contains DROP TABLE') }

if ($errors.Count -gt 0) {
    $errors | ForEach-Object { Write-Error $_ }
    exit 1
}

[pscustomobject]@{
    Result = 'PASS'
    RequiredFiles = $required.Count
    Requirements = $srsReq.Count
    UseCases = $catalogUc.Count
    BusinessRules = $businessRules.Count
    SchemaTables = @([regex]::Matches($schema, '(?mi)^CREATE TABLE ')).Count
    DropTableStatements = @([regex]::Matches($schema, '(?mi)^\s*DROP\s+TABLE')).Count
} | Format-List
