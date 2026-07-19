$ErrorActionPreference = 'Continue'
$out = 'D:\2026-7-11-k8sopearionn\k8s_operation-v16.7\output\diag.txt'
"=== current-context ===" | Out-File $out
kubectl config current-context 2>&1 | Out-File $out -Append
"" | Out-File $out -Append
"=== server URL & auth (minified, secrets redacted) ===" | Out-File $out -Append
kubectl config view --minify 2>&1 | Out-File $out -Append
"" | Out-File $out -Append
"=== k8soperation SA exists? ===" | Out-File $out -Append
kubectl get sa k8soperation -n k8soperation 2>&1 | Out-File $out -Append
"" | Out-File $out -Append
"=== ClusterRoleBinding for k8soperation ===" | Out-File $out -Append
kubectl get clusterrolebinding k8soperation -o wide 2>&1 | Out-File $out -Append
"done" | Out-File $out -Append
