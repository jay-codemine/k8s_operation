$out = "D:\k8s-go\k8s_operation\storage\logs\backend_restart.out.log"
$err = "D:\k8s-go\k8s_operation\storage\logs\backend_restart.err.log"
Set-Location "D:\k8s-go\k8s_operation"
$p = Start-Process -FilePath "go.exe" -ArgumentList "run", "./cmd/k8soperation" -WorkingDirectory "D:\k8s-go\k8s_operation" -RedirectStandardOutput $out -RedirectStandardError $err -WindowStyle Hidden -PassThru
"PID=$($p.Id)"
Start-Sleep -Seconds 35
netstat -ano | findstr "LISTENING" | findstr ":8080"
