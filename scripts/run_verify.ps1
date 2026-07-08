$outFile = "D:\k8s-go\k8s_operation\verify_result.txt"
"=== Verification Start: $(Get-Date) ===" | Out-File $outFile

# Check if backend is listening on 8080
$listen8080 = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue
if ($listen8080) {
    "BACKEND: Already listening on port 8080 (PID: $($listen8080[0].OwningProcess))" | Out-File $outFile -Append
} else {
    "BACKEND: Not running, starting..." | Out-File $outFile -Append
    Start-Process powershell -ArgumentList "-Command","cd D:\k8s-go\k8s_operation; go run ./cmd/k8soperation" -WindowStyle Hidden
    Start-Sleep -Seconds 10
    $listen8080 = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue
    if ($listen8080) {
        "BACKEND: Started successfully (PID: $($listen8080[0].OwningProcess))" | Out-File $outFile -Append
    } else {
        "BACKEND: FAILED to start!" | Out-File $outFile -Append
    }
}

# Check frontend on 5173
$listen5173 = Get-NetTCPConnection -LocalPort 5173 -State Listen -ErrorAction SilentlyContinue
if ($listen5173) {
    "FRONTEND: Running on port 5173 (PID: $($listen5173[0].OwningProcess))" | Out-File $outFile -Append
} else {
    "FRONTEND: Not running!" | Out-File $outFile -Append
}

# Test login endpoint
"" | Out-File $outFile -Append
"=== API Tests ===" | Out-File $outFile -Append
try {
    $body = '{"username":"admin","password":"admin123"}'
    $loginResp = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/auth/login" -Method POST -Body $body -ContentType "application/json" -ErrorAction Stop
    $token = $loginResp.data.token
    "LOGIN: SUCCESS (token length: $($token.Length))" | Out-File $outFile -Append
} catch {
    "LOGIN: FAILED - $($_.Exception.Message)" | Out-File $outFile -Append
    $token = $null
}

if ($token) {
    $headers = @{ "Authorization" = "Bearer $token" }
    
    # Test AIOps Dashboard
    try {
        $dash = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/ai/aiops/dashboard" -Headers $headers -ErrorAction Stop
        "DASHBOARD: SUCCESS - $($dash | ConvertTo-Json -Compress -Depth 3)" | Out-File $outFile -Append
    } catch {
        "DASHBOARD: FAILED - $($_.Exception.Message)" | Out-File $outFile -Append
    }

    # Test Inspection List
    try {
        $list = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/ai/aiops/inspection/list?page=1&page_size=5" -Headers $headers -ErrorAction Stop
        "INSPECTION_LIST: SUCCESS - total=$($list.data.total), items=$($list.data.items.Count)" | Out-File $outFile -Append
    } catch {
        "INSPECTION_LIST: FAILED - $($_.Exception.Message)" | Out-File $outFile -Append
    }

    # Test Channels
    try {
        $ch = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/ai/aiops/channels" -Headers $headers -ErrorAction Stop
        "CHANNELS: SUCCESS - count=$($ch.data.Count)" | Out-File $outFile -Append
    } catch {
        "CHANNELS: FAILED - $($_.Exception.Message)" | Out-File $outFile -Append
    }

    # Test Export (using first report id if exists)
    try {
        $list2 = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/ai/aiops/inspection/list?page=1&page_size=1" -Headers $headers -ErrorAction Stop
        if ($list2.data.items -and $list2.data.items.Count -gt 0) {
            $rid = $list2.data.items[0].id
            $export = Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/ai/aiops/inspection/$rid/export" -Headers $headers -ErrorAction Stop
            "EXPORT: SUCCESS - report_id=$rid, content_length=$($export.data.Length)" | Out-File $outFile -Append
        } else {
            "EXPORT: SKIPPED - no reports" | Out-File $outFile -Append
        }
    } catch {
        "EXPORT: FAILED - $($_.Exception.Message)" | Out-File $outFile -Append
    }
}

# Frontend page test
try {
    $fr = Invoke-WebRequest -Uri "http://127.0.0.1:5173/platform/aiops" -UseBasicParsing -ErrorAction Stop
    "FRONTEND_PAGE: SUCCESS - Status=$($fr.StatusCode), Length=$($fr.Content.Length)" | Out-File $outFile -Append
} catch {
    "FRONTEND_PAGE: FAILED - $($_.Exception.Message)" | Out-File $outFile -Append
}

"" | Out-File $outFile -Append
"=== Verification Complete: $(Get-Date) ===" | Out-File $outFile -Append
