$ErrorActionPreference = 'Continue'

Write-Host "===== 1. 端口监听检查 =====" -ForegroundColor Cyan
$p8080 = Test-NetConnection -ComputerName 127.0.0.1 -Port 8080 -WarningAction SilentlyContinue
$p5173 = Test-NetConnection -ComputerName 127.0.0.1 -Port 5173 -WarningAction SilentlyContinue
Write-Host ("backend  :8080 -> " + $p8080.TcpTestSucceeded)
Write-Host ("frontend :5173 -> " + $p5173.TcpTestSucceeded)

Write-Host ""
Write-Host "===== 2. login 接口 =====" -ForegroundColor Cyan
$loginRaw = curl.exe -s -X POST http://127.0.0.1:8080/api/v1/auth/login `
    -H "Content-Type: application/json" `
    --data-binary "@D:\k8s-go\k8s_operation\tmp_login.json"
Write-Host $loginRaw
$login = $loginRaw | ConvertFrom-Json
$token = $login.data.token
if (-not $token) { Write-Host "[FAIL] no token" -ForegroundColor Red; exit 1 }
Write-Host ("[OK] token len = " + $token.Length)

Write-Host ""
Write-Host "===== 3. monitoring/health =====" -ForegroundColor Cyan
curl.exe -s "http://127.0.0.1:8080/api/v1/monitoring/health" -H "Authorization: Bearer $token"

Write-Host ""
Write-Host ""
Write-Host "===== 4. monitoring/overview =====" -ForegroundColor Cyan
$ov = curl.exe -s "http://127.0.0.1:8080/api/v1/monitoring/overview" -H "Authorization: Bearer $token"
if ($ov.Length -gt 400) { Write-Host ($ov.Substring(0,400) + " ...(truncated)") } else { Write-Host $ov }

Write-Host ""
Write-Host "===== 5. monitoring/datasource (验证 DB 默认源) =====" -ForegroundColor Cyan
curl.exe -s "http://127.0.0.1:8080/api/v1/monitoring/datasource?page=1&page_size=10" -H "Authorization: Bearer $token"
Write-Host ""

Write-Host ""
Write-Host "===== 6. 前端首页 =====" -ForegroundColor Cyan
$fe = curl.exe -s -o NUL -w "%{http_code}" http://127.0.0.1:5173/
Write-Host ("frontend GET / -> HTTP " + $fe)
