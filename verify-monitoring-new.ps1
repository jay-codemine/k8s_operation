$ErrorActionPreference = 'SilentlyContinue'

$loginRaw = curl.exe -s -X POST 'http://127.0.0.1:8080/api/v1/auth/login' -H 'Content-Type: application/json' -d '@tmp_login.json'
$login = $loginRaw | ConvertFrom-Json
$token = $login.data.token
if (-not $token) {
    Write-Host "[FAIL] login: $loginRaw" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] token: $($token.Substring(0,30))..." -ForegroundColor Green
$auth = "Authorization: Bearer $token"

function Test-Api($Name, $Url) {
    Write-Host ""
    Write-Host "=== $Name ===" -ForegroundColor Cyan
    Write-Host "GET $Url" -ForegroundColor DarkGray
    $body = curl.exe -s -H $auth $Url
    if (-not $body) {
        Write-Host "[FAIL] empty response" -ForegroundColor Red
        return
    }
    $obj = $body | ConvertFrom-Json -ErrorAction SilentlyContinue
    if ($obj -and $obj.code -eq 0) {
        $preview = if ($body.Length -gt 500) { $body.Substring(0, 500) + '...' } else { $body }
        Write-Host "[OK] code=0" -ForegroundColor Green
        Write-Host $preview -ForegroundColor DarkGray
    } else {
        Write-Host "[WARN] bad response" -ForegroundColor Yellow
        Write-Host $body -ForegroundColor Yellow
    }
}

Test-Api '1. health score'      'http://127.0.0.1:8080/api/v1/monitoring/score'
Test-Api '2. node heatmap CPU'  'http://127.0.0.1:8080/api/v1/monitoring/heatmap?metric=cpu&duration=1h'
Test-Api '3. pod status dist'   'http://127.0.0.1:8080/api/v1/monitoring/pod-status'
Test-Api '4. abnormal pods'     'http://127.0.0.1:8080/api/v1/monitoring/abnormal-pods'
Test-Api '5. namespace metrics' 'http://127.0.0.1:8080/api/v1/monitoring/namespaces'

Write-Host ""
Write-Host "=== DONE ===" -ForegroundColor Cyan
