param(
    [int]$ConcurrentUsers = 10,
    [int]$RequestsPerUser = 50,
    [string]$BaseURL = "http://localhost:8080"
)

$ErrorActionPreference = "Continue"

Write-Host '============================================' -ForegroundColor Cyan
Write-Host '  CI/CD API Benchmark v1.0' -ForegroundColor Cyan
Write-Host "  Concurrent: $ConcurrentUsers  ReqPerUser: $RequestsPerUser  Total: $($ConcurrentUsers * $RequestsPerUser)" -ForegroundColor Cyan
Write-Host '============================================' -ForegroundColor Cyan

# Step 1: Login
Write-Host ''
Write-Host '[1/4] Login...' -ForegroundColor Yellow
$loginBody = @{username='admin'; password='123456'} | ConvertTo-Json
try {
    $loginResp = Invoke-RestMethod -Uri "$BaseURL/api/v1/auth/login" -Method POST -Body $loginBody -ContentType 'application/json; charset=utf-8'
    if ($loginResp.code -ne 0) {
        Write-Host "  Login failed: $($loginResp.msg)" -ForegroundColor Red
        exit 1
    }
    $token = $loginResp.data.token
    Write-Host "  Login OK, Token: $($token.Substring(0,20))..." -ForegroundColor Green
} catch {
    Write-Host "  Login request failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{ 'Authorization' = "Bearer $token" }

# Step 2: Define test cases
$testCases = @(
    @{
        Name = 'PipelineList'
        URL  = "$BaseURL/api/v1/k8s/cicd/pipeline/list?page=1&page_size=10"
        Method = 'GET'
    },
    @{
        Name = 'PipelineHistory'
        URL  = "$BaseURL/api/v1/k8s/cicd/pipeline/history?id=4&page=1&page_size=20"
        Method = 'GET'
    },
    @{
        Name = 'PipelineStatus'
        URL  = "$BaseURL/api/v1/k8s/cicd/pipeline/status?id=4"
        Method = 'GET'
    },
    @{
        Name = 'PipelineStages'
        URL  = "$BaseURL/api/v1/k8s/cicd/pipeline/stages?id=4"
        Method = 'GET'
    }
)

# Step 3: Run benchmark
$allResults = @()

foreach ($tc in $testCases) {
    Write-Host ''
    Write-Host "[Bench] $($tc.Name)" -ForegroundColor Yellow
    Write-Host "  URL: $($tc.URL)" -ForegroundColor DarkGray

    # Warmup
    try {
        $null = Invoke-RestMethod -Uri $tc.URL -Method $tc.Method -Headers $headers -TimeoutSec 30
    } catch {}

    $jobs = @()
    $scriptBlock = {
        param($url, $method, $hdrs, $count)
        $results = @()
        for ($i = 0; $i -lt $count; $i++) {
            $sw = [System.Diagnostics.Stopwatch]::StartNew()
            try {
                $resp = Invoke-RestMethod -Uri $url -Method $method -Headers $hdrs -TimeoutSec 30
                $sw.Stop()
                $results += @{
                    Status   = 'OK'
                    Code     = $resp.code
                    Duration = $sw.ElapsedMilliseconds
                }
            } catch {
                $sw.Stop()
                $results += @{
                    Status   = 'FAIL'
                    Code     = -1
                    Duration = $sw.ElapsedMilliseconds
                }
            }
        }
        return $results
    }

    $totalSW = [System.Diagnostics.Stopwatch]::StartNew()

    for ($u = 0; $u -lt $ConcurrentUsers; $u++) {
        $jobs += Start-Job -ScriptBlock $scriptBlock -ArgumentList $tc.URL, $tc.Method, $headers, $RequestsPerUser
    }

    $jobs | Wait-Job | Out-Null
    $totalSW.Stop()

    $durations = @()
    $successCount = 0
    $failCount = 0

    foreach ($job in $jobs) {
        $jobResults = Receive-Job -Job $job
        foreach ($r in $jobResults) {
            if ($r.Status -eq 'OK') {
                $successCount++
                $durations += $r.Duration
            } else {
                $failCount++
            }
        }
        Remove-Job -Job $job
    }

    $totalRequests = $ConcurrentUsers * $RequestsPerUser
    $totalTimeSec  = $totalSW.ElapsedMilliseconds / 1000.0

    if ($durations.Count -gt 0) {
        $sorted = $durations | Sort-Object
        $avg = ($durations | Measure-Object -Average).Average
        $min = $sorted[0]
        $max = $sorted[-1]
        $p50 = $sorted[[math]::Floor($sorted.Count * 0.5)]
        $p95 = $sorted[[math]::Floor($sorted.Count * 0.95)]
        $p99 = $sorted[[math]::Min($sorted.Count - 1, [math]::Floor($sorted.Count * 0.99))]
        $qps = [math]::Round($totalRequests / $totalTimeSec, 2)
    } else {
        $avg = 0; $min = 0; $max = 0; $p50 = 0; $p95 = 0; $p99 = 0; $qps = 0
    }

    $result = @{
        Name         = $tc.Name
        Total        = $totalRequests
        Success      = $successCount
        Fail         = $failCount
        TotalTimeSec = [math]::Round($totalTimeSec, 2)
        QPS          = $qps
        AvgMs        = [math]::Round($avg, 1)
        MinMs        = $min
        MaxMs        = $max
        P50Ms        = $p50
        P95Ms        = $p95
        P99Ms        = $p99
    }

    $allResults += $result

    $color = if ($failCount -eq 0) { 'Green' } else { 'Red' }
    Write-Host "  Total: $totalRequests  Success: $successCount  Fail: $failCount" -ForegroundColor $color
    Write-Host "  Time: $($result.TotalTimeSec)s  QPS: $qps" -ForegroundColor Green
    Write-Host "  Latency(ms): avg=$([math]::Round($avg,1)) min=$min max=$max p50=$p50 p95=$p95 p99=$p99" -ForegroundColor Green
}

# Step 4: Summary
Write-Host ''
Write-Host '============================================' -ForegroundColor Cyan
Write-Host '  Benchmark Summary' -ForegroundColor Cyan
Write-Host '============================================' -ForegroundColor Cyan

foreach ($r in $allResults) {
    $line = '{0,-20} Total={1,5} OK={2,5} Fail={3,3} QPS={4,8} Avg={5,6}ms P50={6,5}ms P95={7,5}ms P99={8,5}ms Max={9,5}ms' -f $r.Name, $r.Total, $r.Success, $r.Fail, $r.QPS, $r.AvgMs, $r.P50Ms, $r.P95Ms, $r.P99Ms, $r.MaxMs
    Write-Host $line
}

# Save Markdown results
$sb = [System.Text.StringBuilder]::new()
[void]$sb.AppendLine('| API | Concurrency | Total | Success | Fail | QPS | Avg(ms) | P50(ms) | P95(ms) | P99(ms) | Max(ms) |')
[void]$sb.AppendLine('|-----|-------------|-------|---------|------|-----|---------|---------|---------|---------|---------|')
foreach ($r in $allResults) {
    $row = '| {0} | {1} | {2} | {3} | {4} | {5} | {6} | {7} | {8} | {9} | {10} |' -f $r.Name, $ConcurrentUsers, $r.Total, $r.Success, $r.Fail, $r.QPS, $r.AvgMs, $r.P50Ms, $r.P95Ms, $r.P99Ms, $r.MaxMs
    [void]$sb.AppendLine($row)
}

$mdContent = $sb.ToString()
Write-Host ''
Write-Host '[Markdown Table]' -ForegroundColor Yellow
Write-Host $mdContent

$outputFile = Join-Path $PSScriptRoot 'benchmark_results.txt'
$mdContent | Out-File -FilePath $outputFile -Encoding utf8
Write-Host "Results saved to: $outputFile" -ForegroundColor Green
