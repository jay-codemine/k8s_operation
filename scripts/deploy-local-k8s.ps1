# ============================================================
# K8sOperation - Local K8s One-Click Deploy (Windows PowerShell)
# ============================================================
# Usage: .\scripts\deploy-local-k8s.ps1
#
# Prerequisites:
#   - Docker Desktop (with Kubernetes enabled) or Rancher Desktop
#   - Local MySQL 8.x with docs/sql/k8s_platform_full_init.sql imported
#   - Local Redis running
#
# This script will:
#   1. Check environment (Docker / kubectl / Go)
#   2. Cross-compile Linux amd64 binary
#   3. Build Docker image (loaded locally into K8s)
#   4. Generate local K8s manifests and deploy
#   5. Wait for Pod Ready and print access URL
# ============================================================

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
if (-not $ProjectRoot) { $ProjectRoot = (Get-Location).Path }
Set-Location $ProjectRoot

$ImageName = "k8soperation"
$ImageTag = "local-" + (Get-Date -Format "yyyyMMddHHmm")
$Namespace = "k8soperation"
$NodePort = 30080

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "  K8sOperation Local K8s Deploy" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

# ============================================================
# Step 0: Environment Check
# ============================================================
Write-Host "[Step 0] Checking environment..." -ForegroundColor Yellow

$dockerPath = Get-Command docker -ErrorAction SilentlyContinue
if (-not $dockerPath) {
    Write-Host ""
    Write-Host "[FAIL] Docker not found. Please install Docker Desktop:" -ForegroundColor Red
    Write-Host ""
    Write-Host "  Option 1: Docker Desktop (recommended)" -ForegroundColor White
    Write-Host "    Download: https://www.docker.com/products/docker-desktop/" -ForegroundColor Gray
    Write-Host "    After install: Settings > Kubernetes > Enable Kubernetes > Apply" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Option 2: Rancher Desktop (lightweight alternative)" -ForegroundColor White
    Write-Host "    Download: https://rancherdesktop.io/" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Option 3: winget install Docker.DockerDesktop" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Re-open PowerShell after installation and run this script again." -ForegroundColor Yellow
    exit 1
}
$dockerVer = docker version --format "{{.Client.Version}}" 2>$null
Write-Host "  [OK] Docker: $dockerVer" -ForegroundColor Green

$dockerInfo = docker info 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}
Write-Host "  [OK] Docker Engine: running" -ForegroundColor Green

$kubectlPath = Get-Command kubectl -ErrorAction SilentlyContinue
if (-not $kubectlPath) {
    Write-Host ""
    Write-Host "[FAIL] kubectl not found." -ForegroundColor Red
    Write-Host ""
    Write-Host "  If using Docker Desktop:" -ForegroundColor White
    Write-Host "    Settings > Kubernetes > Enable Kubernetes > Apply and Restart" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Manual install: winget install Kubernetes.kubectl" -ForegroundColor Gray
    Write-Host ""
    exit 1
}

$clusterInfo = kubectl cluster-info 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Cannot connect to K8s cluster." -ForegroundColor Red
    Write-Host "  Ensure Docker Desktop Kubernetes is enabled and ready." -ForegroundColor Yellow
    exit 1
}
$kubectlVer = kubectl version --client --short 2>$null
Write-Host "  [OK] kubectl: $kubectlVer" -ForegroundColor Green
Write-Host "  [OK] K8s cluster: connected" -ForegroundColor Green

$goPath = Get-Command go -ErrorAction SilentlyContinue
if (-not $goPath) {
    Write-Host "[FAIL] Go compiler not found." -ForegroundColor Red
    Write-Host "  Install: winget install GoLang.Go" -ForegroundColor Gray
    exit 1
}
$goVer = go version 2>$null
Write-Host "  [OK] Go: $goVer" -ForegroundColor Green
Write-Host ""

# ============================================================
# Step 1: Cross-compile Linux binary
# ============================================================
Write-Host "[Step 1] Compiling Linux amd64 binary..." -ForegroundColor Yellow

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

$binOutput = "bin/k8s_operation"
go build -trimpath -ldflags="-s -w" -o $binOutput ./cmd/k8soperation/
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Build failed" -ForegroundColor Red
    exit 1
}

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

$fileSize = [math]::Round((Get-Item $binOutput).Length / 1MB, 1)
Write-Host "  [OK] Build success: $binOutput (${fileSize}MB)" -ForegroundColor Green
Write-Host ""

# ============================================================
# Step 2: Build Docker image
# ============================================================
Write-Host "[Step 2] Building Docker image: ${ImageName}:${ImageTag}..." -ForegroundColor Yellow

docker build -t "${ImageName}:${ImageTag}" -t "${ImageName}:latest" -f Dockerfile .
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Docker image build failed" -ForegroundColor Red
    exit 1
}
Write-Host "  [OK] Image built" -ForegroundColor Green
Write-Host ""

# ============================================================
# Step 3: Generate local K8s manifests
# ============================================================
Write-Host "[Step 3] Generating local K8s manifests..." -ForegroundColor Yellow

$deployDir = "deploy/local"
if (-not (Test-Path $deployDir)) {
    New-Item -ItemType Directory -Path $deployDir -Force | Out-Null
}

$hostAddr = "host.docker.internal"

# Generate namespace.yaml
$nsYaml = @"
apiVersion: v1
kind: Namespace
metadata:
  name: $Namespace
"@
$nsYaml | Set-Content "$deployDir/namespace.yaml" -Encoding UTF8

# Generate configmap.yaml
$cmYaml = @"
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8soperation-config
  namespace: $Namespace
data:
  config.yaml: |
    Server:
      RunMode: release
      Port: 8080
      ReadTimeout: 3600
      WriteTimeout: 3600
      IdleTimeout: 300
      ShutdownTimeout: 300
    Database:
      DBType: mysql
      Username: root
      Password: "123456"
      Host: $hostAddr
      Port: "3306"
      DBName: k8s-platform
      Charset: utf8mb4
      ParseTime: true
      MaxIdleConns: 10
      MaxOpenConns: 100
      MaxLifeSeconds: 300
    Cache:
      Type: redis
      Name: sk_sid
      Address: "${hostAddr}:6379"
      Username: ""
      Password: "123456"
      MaxConnect: 10
      Network: tcp
      Secret: "k8smana"
    App:
      LogLevel: info
      TIMEZONE: "Asia/Shanghai"
      LogType: single
      LogFileName: storage/logs/app.log
      BusinessLogFileName: storage/logs/biz.log
      LogMaxSize: 50
      LogMaxBackup: 5
      LogMaxAge: 30
      LogCompress: true
      MirrorBusinessToSystem: false
      JWTMaxRefreshTime: 86400
      JWTSigningKey: eoNB0%bv5M7995F1
      JWTExpireTime: 120000
      AppName: "k8soperation"
      GlobalKubeConfigPath: ""
      DefaultClusterID: 0
      AutoInitK8s: true
      AllowEmptyStart: true
    PodLog:
      EnableStreaming: false
      TailDefault: 500
      TailMax: 5000
      LimitBytes: 2097152
      Timestamps: false
      Previous: false
    ErrorCode:
      AllowOverride: true
    ClusterClient:
      TTL: 30m
      TTLJitter: 3m
    Pod:
      eviction:
        default_grace_seconds: 30
    Node:
      drain:
        max_grace_seconds: 300
        ignore_daemon_sets: true
        delete_empty_dir: false
    Jenkins:
      URL: ""
      Username: ""
      APIToken: ""
      TriggerTimeout: 60
      CallbackURL: "http://k8soperation.$Namespace.svc:8080"
      PlatformURL: "http://localhost:$NodePort"
      HMACSecret: "local-test-hmac-secret"
      PollInterval: 15
      MaxBuildTime: 30
      DingTalkWebhook: ""
      FeishuWebhook: ""
      FeishuSecret: ""
    Security:
      KubeConfigEncryptKey: "K8sOp@2024!SecureKey#AES256Encrypt"
      PasswordBcryptCost: 10
      AutoEncryptLegacyData: true
    AIAssistant:
      Enabled: false
      DefaultProvider: "qwen"
      SystemPrompt: "K8s AI Assistant"
      ApprovalExpire: 30
      MaxHistoryRound: 20
"@
$cmYaml | Set-Content "$deployDir/configmap.yaml" -Encoding UTF8

# Generate rbac.yaml
$rbacYaml = @"
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8soperation
  namespace: $Namespace
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8soperation
rules:
  - apiGroups: [""]
    resources: ["pods","pods/log","pods/exec","services","configmaps","secrets","persistentvolumes","persistentvolumeclaims","nodes","namespaces","events","serviceaccounts"]
    verbs: ["get","list","watch","create","update","patch","delete"]
  - apiGroups: ["apps"]
    resources: ["deployments","statefulsets","daemonsets","replicasets"]
    verbs: ["get","list","watch","create","update","patch","delete"]
  - apiGroups: ["batch"]
    resources: ["jobs","cronjobs"]
    verbs: ["get","list","watch","create","update","patch","delete"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses"]
    verbs: ["get","list","watch","create","update","patch","delete"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get","list","watch"]
  - apiGroups: ["metrics.k8s.io"]
    resources: ["nodes","pods"]
    verbs: ["get","list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8soperation
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8soperation
subjects:
  - kind: ServiceAccount
    name: k8soperation
    namespace: $Namespace
"@
$rbacYaml | Set-Content "$deployDir/rbac.yaml" -Encoding UTF8

# Generate deployment.yaml
$deployYaml = @"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8soperation
  namespace: $Namespace
spec:
  replicas: 1
  selector:
    matchLabels:
      app: k8soperation
  template:
    metadata:
      labels:
        app: k8soperation
    spec:
      serviceAccountName: k8soperation
      containers:
        - name: k8soperation
          image: ${ImageName}:${ImageTag}
          imagePullPolicy: Never
          ports:
            - name: http
              containerPort: 8080
          env:
            - name: GIN_MODE
              value: "release"
            - name: APP_CONFIG
              value: "/app/configs/config.yaml"
          livenessProbe:
            httpGet:
              path: /healthz/live
              port: http
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /healthz/ready
              port: http
            initialDelaySeconds: 15
            periodSeconds: 10
          startupProbe:
            httpGet:
              path: /healthz/live
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
            failureThreshold: 30
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: "1"
              memory: 512Mi
          volumeMounts:
            - name: config
              mountPath: /app/configs/config.yaml
              subPath: config.yaml
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: k8soperation-config
"@
$deployYaml | Set-Content "$deployDir/deployment.yaml" -Encoding UTF8

# Generate service.yaml (NodePort)
$svcYaml = @"
apiVersion: v1
kind: Service
metadata:
  name: k8soperation
  namespace: $Namespace
spec:
  type: NodePort
  selector:
    app: k8soperation
  ports:
    - name: http
      port: 8080
      targetPort: http
      nodePort: $NodePort
      protocol: TCP
"@
$svcYaml | Set-Content "$deployDir/service.yaml" -Encoding UTF8

# Generate kustomization.yaml
$kustomYaml = @"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: $Namespace
resources:
  - namespace.yaml
  - configmap.yaml
  - rbac.yaml
  - deployment.yaml
  - service.yaml
"@
$kustomYaml | Set-Content "$deployDir/kustomization.yaml" -Encoding UTF8

Write-Host "  [OK] Manifests generated at $deployDir/" -ForegroundColor Green
Write-Host ""

# ============================================================
# Step 4: Deploy to K8s
# ============================================================
Write-Host "[Step 4] Deploying to local K8s cluster..." -ForegroundColor Yellow

kubectl apply -k $deployDir
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Deployment failed" -ForegroundColor Red
    exit 1
}
Write-Host "  [OK] K8s resources created" -ForegroundColor Green
Write-Host ""

# ============================================================
# Step 5: Wait for Pod Ready
# ============================================================
Write-Host "[Step 5] Waiting for Pod ready (max 120s)..." -ForegroundColor Yellow

$ready = $false
$readyJsonpath = "jsonpath={.items[0].status.conditions[?(@.type==""Ready"")].status}"
$phaseJsonpath = "jsonpath={.items[0].status.phase}"

for ($i = 0; $i -lt 24; $i++) {
    Start-Sleep -Seconds 5
    $podStatus = kubectl get pods -n $Namespace -l app=k8soperation -o $readyJsonpath 2>$null
    if ($podStatus -eq "True") {
        $ready = $true
        break
    }
    $phase = kubectl get pods -n $Namespace -l app=k8soperation -o $phaseJsonpath 2>$null
    Write-Host "  ... Pod status: $phase ($i/24)" -ForegroundColor Gray
}

Write-Host ""
if ($ready) {
    Write-Host "============================================================" -ForegroundColor Green
    Write-Host "  [OK] Deploy SUCCESS! Platform is ready." -ForegroundColor Green
    Write-Host "============================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "  API URL: http://localhost:${NodePort}" -ForegroundColor White
    Write-Host "  Account: admin / admin123" -ForegroundColor White
    Write-Host ""
    Write-Host "  Commands:" -ForegroundColor Cyan
    Write-Host "    View pods:    kubectl get pods -n $Namespace" -ForegroundColor Gray
    Write-Host "    View logs:    kubectl logs -n $Namespace -l app=k8soperation -f" -ForegroundColor Gray
    Write-Host "    Shell:        kubectl exec -it -n $Namespace deploy/k8soperation -- sh" -ForegroundColor Gray
    Write-Host "    Teardown:     kubectl delete -k $deployDir" -ForegroundColor Gray
    Write-Host "    Redeploy:     kubectl rollout restart deploy/k8soperation -n $Namespace" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Health check:" -ForegroundColor Cyan
    Write-Host "    Invoke-RestMethod http://localhost:${NodePort}/healthz/ready" -ForegroundColor Gray
    Write-Host ""
} else {
    Write-Host "[WARN] Pod not ready within 120s. Check manually:" -ForegroundColor Yellow
    Write-Host ""
    kubectl get pods -n $Namespace
    Write-Host ""
    Write-Host "  kubectl describe pod -n $Namespace -l app=k8soperation" -ForegroundColor Gray
    Write-Host "  kubectl logs -n $Namespace -l app=k8soperation" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Common issues:" -ForegroundColor Cyan
    Write-Host "    1. Image pull fail: imagePullPolicy is Never (uses local image)" -ForegroundColor Gray
    Write-Host "    2. MySQL connect fail: ensure MySQL allows Docker network access" -ForegroundColor Gray
    Write-Host "    3. Redis connect fail: ensure Redis listens on 0.0.0.0" -ForegroundColor Gray
    exit 1
}
