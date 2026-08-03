# Verify that docs/sql/k8s_platform_full_init.sql imports cleanly into an empty DB.
#
# Steps:
#   1) Static check: no MariaDB-only "ADD COLUMN IF NOT EXISTS" left in executable DDL
#   2) Create a throwaway database and import the whole init script into it
#   3) Check multi-tenant preconditions (tenant_id on every table, per-tenant role unique key)
#   4) Re-run the tenant section to prove idempotency, then drop the throwaway database
#
# The init script pins its own target schema (CREATE DATABASE / USE `k8s-platform`), which is
# what makes it one-click for users but would hijack this test into the live database. Those
# two statements are therefore stripped from the copy fed to the throwaway schema.
#
# ASCII only on purpose: PowerShell 5.1 reads BOM-less UTF-8 files as ANSI/GBK, which corrupts
# non-ASCII literals and breaks parsing.
#
# Usage: powershell -ExecutionPolicy Bypass -File scripts\verify-init-sql.ps1

$ErrorActionPreference = 'Stop'

$repo   = Split-Path -Parent $PSScriptRoot
$sql    = Join-Path $repo 'docs\sql\k8s_platform_full_init.sql'
$mysql  = 'D:\develop\mysql-8.0.31-winx64\bin\mysql.exe'
$testDb = 'k8s_platform_import_test'

$bytes = [System.IO.File]::ReadAllBytes($sql)
$text  = [System.IO.File]::ReadAllText($sql)
# Must not use Get-Content here: PowerShell 5.1 decodes BOM-less UTF-8 as ANSI/GBK, which turns
# the Chinese COMMENT literals into mojibake and unbalances their quotes.
$lines = $text -split "`n"

# ---------- 1) static file checks ----------
# Only non-comment lines matter: the rationale comments legitimately quote the MariaDB syntax.
$badDdl = @($lines | Where-Object { $_ -match 'ADD COLUMN IF NOT EXISTS' -and $_ -notmatch '^\s*--' })

Write-Output '=== FILE ==='
Write-Output ("BYTES          = {0}" -f $bytes.Length)
Write-Output ("HAS_BOM        = {0}   expect False" -f (($bytes[0] -eq 0xEF) -and ($bytes[1] -eq 0xBB) -and ($bytes[2] -eq 0xBF)))
Write-Output ("CR_COUNT       = {0}   expect 0 (LF only)" -f ($text.ToCharArray() | Where-Object { $_ -eq [char]13 }).Count)
Write-Output ("LINES          = {0}" -f $lines.Count)
Write-Output ("MARIADB_DDL    = {0}   expect 0" -f $badDdl.Count)
Write-Output ("CALL_COUNT     = {0}   expect 56" -f @($lines | Where-Object { $_ -match '^CALL ' }).Count)
Write-Output ("PROC_COUNT     = {0}   expect 1" -f @($lines | Where-Object { $_ -match '^CREATE PROCEDURE' }).Count)
Write-Output ("PROC_DROPPED   = {0}   expect 2" -f @($lines | Where-Object { $_ -match '^DROP PROCEDURE IF EXISTS' }).Count)
Write-Output ("UK_REFS        = {0}   expect 2" -f @($lines | Where-Object { $_ -match 'uk_sys_role_tenant_name' }).Count)

# ---------- 2) read local DB password (never echoed) ----------
$cfg = Get-Content -LiteralPath (Join-Path $repo 'configs\config.yaml') -Raw
$pw  = ([regex]::Match($cfg, '(?m)^\s*Password:\s*(.+?)\s*$')).Groups[1].Value.Trim('"').Trim("'")
if (-not $pw) { throw 'cannot read DB password from configs/config.yaml' }
$env:MYSQL_PWD = $pw

$common = @('--host=127.0.0.1', '--port=3306', '--user=root',
            '--default-character-set=utf8mb4', '--batch', '--skip-column-names')

# ---------- 3) full import into an empty database ----------
Write-Output ''
Write-Output "=== IMPORT INTO $testDb ==="
& $mysql @common --execute="DROP DATABASE IF EXISTS ``$testDb``; CREATE DATABASE ``$testDb`` DEFAULT CHARACTER SET utf8mb4;"
Write-Output ("PREPARE_EXIT   = {0}" -f $LASTEXITCODE)

$portable = Join-Path $env:TEMP 'init_no_use.sql'
$stripped = @($lines | Where-Object { $_ -notmatch '^\s*(CREATE DATABASE|USE )' })
[System.IO.File]::WriteAllText($portable, ($stripped -join "`n") + "`n", (New-Object System.Text.UTF8Encoding($false)))
Write-Output ("STRIPPED_LINES = {0}   expect 2" -f ($lines.Count - $stripped.Count))

& $mysql @common "--database=$testDb" --execute="source $($portable.Replace('\','/'))"
Write-Output ("IMPORT_EXIT    = {0}   expect 0" -f $LASTEXITCODE)

# ---------- 4) post-import self check ----------
Write-Output ''
Write-Output '=== POST-IMPORT CHECK ==='
$q = @'
SELECT CONCAT('TABLES=', COUNT(*)) FROM information_schema.TABLES
 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_TYPE = 'BASE TABLE';
SELECT CONCAT('VIEWS=', COUNT(*)) FROM information_schema.VIEWS
 WHERE TABLE_SCHEMA = DATABASE();
SELECT CONCAT('MISSING_TENANT_ID=', COUNT(*)) FROM information_schema.TABLES t
 WHERE t.TABLE_SCHEMA = DATABASE() AND t.TABLE_TYPE = 'BASE TABLE'
   AND NOT EXISTS (SELECT 1 FROM information_schema.COLUMNS c
     WHERE c.TABLE_SCHEMA = t.TABLE_SCHEMA AND c.TABLE_NAME = t.TABLE_NAME
       AND c.COLUMN_NAME = 'tenant_id');
SELECT CONCAT('TENANT_ID_INDEXES=', COUNT(DISTINCT TABLE_NAME)) FROM information_schema.STATISTICS
 WHERE TABLE_SCHEMA = DATABASE() AND INDEX_NAME = 'idx_tenant_id';
SELECT CONCAT('LEGACY_ROLE_INDEX=', COUNT(*)) FROM information_schema.STATISTICS
 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_role'
   AND INDEX_NAME = 'idx_sys_role_name';
SELECT CONCAT('UK_TENANT_NAME=', COUNT(*)) FROM information_schema.STATISTICS
 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_role'
   AND INDEX_NAME = 'uk_sys_role_tenant_name';
SELECT CONCAT('LEFTOVER_PROC=', COUNT(*)) FROM information_schema.ROUTINES
 WHERE ROUTINE_SCHEMA = DATABASE() AND ROUTINE_NAME = '__add_tenant_id';
SELECT CONCAT('SEED_ROLES=', COUNT(*)) FROM sys_role;
SELECT CONCAT('SEED_USERS=', COUNT(*)) FROM `user`;
SELECT CONCAT('TENANTS=', COUNT(*)) FROM tenant;
SELECT CONCAT('ADMIN_BOUND=', COUNT(*)) FROM sys_user_role;
'@
& $mysql @common "--database=$testDb" --execute=$q
Write-Output ("CHECK_EXIT     = {0}   expect 0" -f $LASTEXITCODE)

# ---------- 5) re-run tenant section to prove idempotency ----------
Write-Output ''
Write-Output '=== IDEMPOTENT RE-RUN (tenant section only) ==='
$marker = 'DROP PROCEDURE IF EXISTS `__add_tenant_id`;'
$idx = $text.IndexOf($marker)
if ($idx -lt 0) {
  Write-Output 'SECTION_NOT_FOUND'
} else {
  $tmp = Join-Path $env:TEMP 'tenant_section_rerun.sql'
  [System.IO.File]::WriteAllText($tmp, $text.Substring($idx), (New-Object System.Text.UTF8Encoding($false)))
  & $mysql @common "--database=$testDb" --execute="source $($tmp.Replace('\','/'))"
  Write-Output ("RERUN_EXIT     = {0}   expect 0" -f $LASTEXITCODE)
  Remove-Item -LiteralPath $tmp -Force
}

# ---------- 6) cleanup ----------
Write-Output ''
& $mysql @common --execute="DROP DATABASE IF EXISTS ``$testDb``;"
Write-Output ("DROP_TESTDB_EXIT = {0}" -f $LASTEXITCODE)
Remove-Item -LiteralPath $portable -Force

$env:MYSQL_PWD = $null
Write-Output 'DONE'
