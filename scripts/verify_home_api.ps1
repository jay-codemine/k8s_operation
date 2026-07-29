# Home page API verification script (ASCII only for PS 5.1 compatibility)
# Usage: .\verify_home_api.ps1 -Token "<JWT>"
param([Parameter(Mandatory = $true)][string]$Token)

$base = "http://127.0.0.1:8080"
$headers = @{ Authorization = "Bearer $Token" }

$apis = @(
    @{ Name = "monitoring-overview";  Url = "/api/v1/monitoring/overview" },
    @{ Name = "trend-cpu";            Url = "/api/v1/monitoring/trend/cpu?duration=1h" },
    @{ Name = "trend-memory";         Url = "/api/v1/monitoring/trend/memory?duration=1h" },
    @{ Name = "trend-network";        Url = "/api/v1/monitoring/trend/network?duration=1h" },
    @{ Name = "pod-status";           Url = "/api/v1/monitoring/pod-status" },
    @{ Name = "alert-events";         Url = "/api/v1/monitoring/alert-event?page=1&size=5" },
    @{ Name = "alert-stats";          Url = "/api/v1/monitoring/alert-event/stats" },
    @{ Name = "platform-health";      Url = "/api/v1/platform/health" },
    @{ Name = "cluster-list";         Url = "/api/v1/k8s/cluster/list?page=1&limit=100" },
    @{ Name = "pipeline-list";        Url = "/api/v1/k8s/cicd/pipeline/list?page=1&page_size=5" },
    @{ Name = "build-stats";          Url = "/api/v1/k8s/cicd/pipeline/build-stats?days=7" },
    @{ Name = "release-list";         Url = "/api/v1/k8s/cicd/release/list?page=1&page_size=5" }
)

foreach ($api in $apis) {
    try {
        $resp = Invoke-RestMethod -Uri "$base$($api.Url)" -Headers $headers -TimeoutSec 40
        $preview = ($resp.data | ConvertTo-Json -Depth 3 -Compress)
        if ($preview -and $preview.Length -gt 500) { $preview = $preview.Substring(0, 500) + "..." }
        Write-Host ("[OK  ] {0,-20} code={1} data={2}" -f $api.Name, $resp.code, $preview)
    }
    catch {
        $status = ""
        if ($_.Exception.Response) { $status = [int]$_.Exception.Response.StatusCode }
        Write-Host ("[FAIL] {0,-20} HTTP={1} {2}" -f $api.Name, $status, $_.Exception.Message)
    }
}
