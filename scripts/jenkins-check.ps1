$ErrorActionPreference = 'Continue'
$out = 'D:\2026-7-11-k8sopearionn\k8s_operation-v16.7\output\jenkins-check.txt'
$p = Get-Process kubectl -ErrorAction SilentlyContinue
("kubectl processes running: " + (@($p).Count)) | Out-File $out
try {
    $r = Invoke-WebRequest -Uri 'http://localhost:8081/login' -UseBasicParsing -TimeoutSec 8
    ("jenkins http://localhost:8081/login => HTTP " + $r.StatusCode) | Out-File $out -Append
} catch {
    ("jenkins => ERR " + $_.Exception.Message) | Out-File $out -Append
}
"done" | Out-File $out -Append
