$ErrorActionPreference = 'SilentlyContinue'

# 清理可能已存在的同名转发进程
Get-Process kubectl -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -like '*port-forward*' } | Stop-Process -Force -ErrorAction SilentlyContinue

$log = 'D:\2026-7-11-k8sopearionn\k8s_operation-v16.7\output'

# 前端: localhost:8082 -> svc/k8soperation-web:80
Start-Process -FilePath 'kubectl' -ArgumentList 'port-forward','-n','k8soperation','svc/k8soperation-web','8082:80','--address','127.0.0.1' -WindowStyle Hidden -RedirectStandardOutput "$log\pf-web.log" -RedirectStandardError "$log\pf-web.err"

# 后端: localhost:8088 -> svc/k8soperation:8080
Start-Process -FilePath 'kubectl' -ArgumentList 'port-forward','-n','k8soperation','svc/k8soperation','8088:8080','--address','127.0.0.1' -WindowStyle Hidden -RedirectStandardOutput "$log\pf-be.log" -RedirectStandardError "$log\pf-be.err"

# Jenkins: localhost:8081 -> svc/jenkins:8080
Start-Process -FilePath 'kubectl' -ArgumentList 'port-forward','-n','devops','svc/jenkins','8081:8080','--address','127.0.0.1' -WindowStyle Hidden -RedirectStandardOutput "$log\pf-jenkins.log" -RedirectStandardError "$log\pf-jenkins.err"

Start-Sleep -Seconds 6

$out = "$log\diag.txt"
"=== port-forward access test ===" | Out-File $out
$targets = @(
    @{ Name = 'frontend'; Url = 'http://localhost:8082' },
    @{ Name = 'backend-ready'; Url = 'http://localhost:8088/healthz/ready' },
    @{ Name = 'jenkins'; Url = 'http://localhost:8081/login' }
)
foreach ($t in $targets) {
    try {
        $r = Invoke-WebRequest -Uri $t.Url -UseBasicParsing -TimeoutSec 10
        ("{0,-14} {1} => HTTP {2}" -f $t.Name, $t.Url, $r.StatusCode) | Out-File $out -Append
    } catch {
        $code = $_.Exception.Response.StatusCode.value__
        if ($code) {
            ("{0,-14} {1} => HTTP {2}" -f $t.Name, $t.Url, $code) | Out-File $out -Append
        } else {
            ("{0,-14} {1} => ERR {2}" -f $t.Name, $t.Url, $_.Exception.Message) | Out-File $out -Append
        }
    }
}
"done" | Out-File $out -Append
