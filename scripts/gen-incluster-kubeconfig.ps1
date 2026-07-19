$ErrorActionPreference = 'Continue'
$log = 'D:\2026-7-11-k8sopearionn\k8s_operation-v16.7\output'
$out = "$log\diag.txt"

# 1) Create a long-lived ServiceAccount token Secret (K8s 1.24+ no longer auto-creates it)
$secretYaml = @"
apiVersion: v1
kind: Secret
metadata:
  name: k8soperation-sa-token
  namespace: k8soperation
  annotations:
    kubernetes.io/service-account.name: k8soperation
type: kubernetes.io/service-account-token
"@
$secretYaml | kubectl apply -f - 2>&1 | Out-File $out
Start-Sleep -Seconds 4

# 2) Read the token (the token controller fills .data.token automatically)
$jp = '{.data.token}'
$tokenB64 = kubectl get secret k8soperation-sa-token -n k8soperation -o "jsonpath=$jp" 2>$null
if ([string]::IsNullOrWhiteSpace($tokenB64)) {
    "ERROR: token not generated yet, retry later" | Out-File $out -Append
    return
}
$token = [System.Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($tokenB64))

# 3) Build a dedicated kubeconfig (in-cluster server address; insecure skips TLS, matches platform tuneRESTConfig)
$kubeconfig = @"
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://kubernetes.default.svc:443
    insecure-skip-tls-verify: true
  name: local-docker-desktop
contexts:
- context:
    cluster: local-docker-desktop
    user: k8soperation
  name: local-docker-desktop
current-context: local-docker-desktop
users:
- name: k8soperation
  user:
    token: $token
"@

$kcPath = "$log\local-incluster.kubeconfig"
$kubeconfig | Out-File -FilePath $kcPath -Encoding ascii

"=== kubeconfig generated: $kcPath ===" | Out-File $out -Append
"=== content (token partly redacted) ===" | Out-File $out -Append
($kubeconfig -replace [regex]::Escape($token), ($token.Substring(0,20) + '...<redacted>...')) | Out-File $out -Append
"done" | Out-File $out -Append
