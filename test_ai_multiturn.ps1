# =====================================================================
# AI Multi-Turn Conversation Verification Script (ASCII-safe)
# Usage: powershell -ExecutionPolicy Bypass -File .\test_ai_multiturn.ps1
# =====================================================================

$ErrorActionPreference = "Stop"
$BASE = "http://localhost:8080"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host " AI Multi-Turn Dialogue Verification" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

# 1. Login
Write-Host ""
Write-Host "[1/4] Login..." -ForegroundColor Yellow
$loginBody = @{username="admin";password="123456"} | ConvertTo-Json
try {
    $login = Invoke-RestMethod -Uri "$BASE/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json; charset=utf-8"
} catch {
    Write-Host "  [X] Backend not reachable: $_" -ForegroundColor Red
    Write-Host "  Please start: go run cmd/k8soperation/main.go" -ForegroundColor Yellow
    exit 1
}
if ($login.code -ne 0) {
    Write-Host "  [X] Login failed: $($login.msg)" -ForegroundColor Red
    exit 1
}
$token = $login.data.token
Write-Host "  [OK] Login success" -ForegroundColor Green

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type"  = "application/json; charset=utf-8"
}

# Helper: send a chat message and return response
function Send-AIChat {
    param([string]$Text, $ConvId)
    $payload = @{ message = $Text }
    if ($ConvId) { $payload.conversation_id = [int]$ConvId }
    $json = $payload | ConvertTo-Json -Compress
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($json)
    return Invoke-RestMethod -Uri "$BASE/api/v1/ai/chat" -Method POST -Headers $headers -Body $bytes
}

# 2. Round 1: tell AI a fact
Write-Host ""
Write-Host "[2/4] Round 1: Please remember my name is Alice and favorite color is purple" -ForegroundColor Yellow
$r1 = Send-AIChat -Text "Please remember: my name is Alice and my favorite color is purple. Just reply OK."
if ($r1.code -ne 0) {
    Write-Host "  [X] Chat failed: $($r1.msg)" -ForegroundColor Red
    exit 1
}
$convId = $r1.data.conversation_id
Write-Host "  conversation_id : $convId" -ForegroundColor Gray
Write-Host "  context_round   : $($r1.data.context_round)" -ForegroundColor Gray
Write-Host "  history_count   : $($r1.data.history_count)" -ForegroundColor Gray
Write-Host "  AI reply        : $($r1.data.reply)" -ForegroundColor White

# 3. Round 2: ask AI to recall (same conversation_id)
Write-Host ""
Write-Host "[3/4] Round 2 (same conv): What is my name and favorite color?" -ForegroundColor Yellow
$r2 = Send-AIChat -Text "Based on what I just told you, what is my name and what is my favorite color? Reply directly." -ConvId $convId
Write-Host "  conversation_id : $($r2.data.conversation_id)" -ForegroundColor Gray
Write-Host "  context_round   : $($r2.data.context_round)" -ForegroundColor Gray
Write-Host "  history_count   : $($r2.data.history_count)" -ForegroundColor Gray
Write-Host "  AI reply        : $($r2.data.reply)" -ForegroundColor White

# 4. Verify
Write-Host ""
Write-Host "[4/4] Verifying..." -ForegroundColor Yellow
$reply = $r2.data.reply
$pass1 = $reply -match "Alice"
$pass2 = $reply -match "purple"
$ctxOk = ($r2.data.context_round -ge 2) -and ($r2.data.history_count -ge 2)

Write-Host "  [+] Reply contains 'Alice'         : " -NoNewline
if ($pass1) { Write-Host "PASS" -ForegroundColor Green } else { Write-Host "FAIL" -ForegroundColor Red }

Write-Host "  [+] Reply contains 'purple'        : " -NoNewline
if ($pass2) { Write-Host "PASS" -ForegroundColor Green } else { Write-Host "FAIL" -ForegroundColor Red }

Write-Host "  [+] Context counter correct        : " -NoNewline
if ($ctxOk) {
    Write-Host "PASS (round=$($r2.data.context_round), history=$($r2.data.history_count))" -ForegroundColor Green
} else {
    Write-Host "FAIL (round=$($r2.data.context_round), history=$($r2.data.history_count))" -ForegroundColor Red
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
if ($pass1 -and $pass2 -and $ctxOk) {
    Write-Host " [SUCCESS] Multi-turn dialogue verified!" -ForegroundColor Green
    Write-Host " AI correctly remembered the context." -ForegroundColor Green
} else {
    Write-Host " [WARN] Verification incomplete" -ForegroundColor Yellow
    Write-Host " Check backend log: storage/logs/ai.log" -ForegroundColor Yellow
}
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Full conversation: GET $BASE/api/v1/ai/conversations/$convId/messages" -ForegroundColor DarkGray
