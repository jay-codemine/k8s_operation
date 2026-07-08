# macOS 本地构建镜像与 K8s 部署指南

> **场景**：在 macOS 开发机上构建前后端 Docker 镜像（含交叉编译），推送到镜像仓库，部署到 Linux K8s 集群。
> 
> **核心原则**：交叉编译全部在 Dockerfile 多阶段构建中完成，macOS 上只需要 Docker Desktop，无需安装 Go / Node 环境。

---

## 一、架构总览

```
┌──────────────────────────────────────────────────────────────┐
│  macOS 开发机                                                 │
│                                                                │
│  ┌──────────────────┐    ┌──────────────────┐                │
│  │ docker build     │    │ docker build     │                │
│  │ (后端多阶段)      │    │ (前端多阶段)      │                │
│  │                  │    │                  │                │
│  │ Stage1: Go 交叉  │    │ Stage1: Node 构建│                │
│  │   编译 → linux   │    │   npm build      │                │
│  │ Stage2: Alpine   │    │ Stage2: Nginx    │                │
│  │   运行时镜像     │    │   运行时镜像     │                │
│  └────────┬─────────┘    └────────┬─────────┘                │
│           │                        │                          │
│           ▼                        ▼                          │
│      docker push              docker push                     │
└───────────┼────────────────────────┼──────────────────────────┘
            │                        │
            ▼                        ▼
┌──────────────────────────────────────────────────────────────┐
│  镜像仓库 (Harbor / DockerHub / 阿里云ACR)                    │
│  registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest              │
│  registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest              │
└───────────┬────────────────────────┬──────────────────────────┘
            │                        │
            ▼                        ▼
┌──────────────────────────────────────────────────────────────┐
│  Linux K8s 集群                                               │
│  kubectl apply -k deploy/  → 拉取镜像并部署                   │
└──────────────────────────────────────────────────────────────┘
```

---

## 二、前置条件

| 组件 | 要求 | 验证命令 |
|------|------|----------|
| Docker Desktop | 4.x+（含 buildx） | `docker version` |
| Docker 已登录镜像仓库 | 能 push 到目标 registry | `docker login registry.cn-hangzhou.aliyuncs.com` |
| kubectl | 已配置目标集群 kubeconfig | `kubectl cluster-info` |
| 网络 | macOS 可访问镜像仓库 + K8s 集群 | — |

> **重要**：macOS 上的 Docker Desktop 构建镜像时，默认目标平台为 `linux/amd64`（或 `linux/arm64` on Apple Silicon），
> 而我们的 Dockerfile 内已显式设置 `GOOS=linux`，所以**无需本地安装 Go 或 Node**。

---

## 三、文件说明

| 文件 | 用途 | 交叉编译方式 |
|------|------|-------------|
| `docker/backend/Dockerfile` | 后端镜像（推荐） | Docker 内 Go 编译，`GOOS=linux CGO_ENABLED=0` |
| `docker/backend/Dockerfile.runtime` | 后端镜像（需先本地编译） | 需手动 `go build`，Dockerfile 仅打包 |
| `docker/frontend/Dockerfile` | 前端镜像 | Docker 内 Node 编译 + Nginx 打包 |

---

## 四、完整操作步骤

### Step 0：确认 Docker Desktop 运行中

```bash
# 确认 Docker 正在运行
docker info

# 登录阿里云镜像仓库（首次执行，后续不需要）
docker login registry.cn-hangzhou.aliyuncs.com
# 输入用户名和密码
```

---

### Step 1：构建后端镜像（多阶段，含交叉编译）

```bash
# 在项目根目录执行
cd /Users/dai/k8s_operation-main

# 使用多阶段 Dockerfile（Docker 内自动完成 Go 交叉编译）
docker build -f docker/backend/Dockerfile \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest \
  .
```

**构建过程说明：**

```
Dockerfile 内部流程：
┌─────────────────────────────────────────────────────────┐
│ Stage 1 (golang:1.22-alpine)                            │
│                                                          │
│  ENV GOOS=linux        ← 关键：强制编译 Linux 二进制     │
│  ENV CGO_ENABLED=0     ← 静态链接，无 C 依赖            │
│  ENV GOPROXY=https://goproxy.cn,direct                  │
│                                                          │
│  go mod download       ← 下载依赖（Docker 缓存加速）    │
│  go build ... -o /app/devops-be ./cmd/k8soperation      │
│                                                          │
│  产出：/app/devops-be (linux/amd64 ELF 二进制)          │
└────────────────────────────┬────────────────────────────┘
                             │ COPY --from=builder
                             ▼
┌─────────────────────────────────────────────────────────┐
│ Stage 2 (alpine:3.20)                                    │
│                                                          │
│  最终镜像 ≈ 25MB                                         │
│  包含：二进制 + CA证书 + 时区 + Jenkins模板              │
│  用户：app (UID 1000)                                    │
│  端口：8080                                              │
└─────────────────────────────────────────────────────────┘
```

---

### Step 2：构建前端镜像（多阶段）

```bash
# 在项目根目录执行
docker build -f docker/frontend/Dockerfile \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest \
  ./k8s-web
```

**构建过程说明：**

```
docker/frontend/Dockerfile 内部流程：
┌─────────────────────────────────────────────────────────┐
│ Stage 1 (node:22-alpine)                                 │
│                                                          │
│  npm config set registry https://registry.npmmirror.com  │
│  npm ci                ← 安装依赖                        │
│  npm run build         ← Vite 构建产物到 dist/           │
│                                                          │
│  产出：/app/dist/ (静态 HTML/JS/CSS)                     │
└────────────────────────────┬────────────────────────────┘
                             │ COPY --from=builder
                             ▼
┌─────────────────────────────────────────────────────────┐
│ Stage 2 (nginx:1.27-alpine)                              │
│                                                          │
│  最终镜像 ≈ 40MB                                         │
│  包含：Nginx + 静态文件 + API 反向代理配置               │
│  支持：API_BACKEND_URL 环境变量动态注入后端地址          │
│  端口：80                                                │
└─────────────────────────────────────────────────────────┘
```

---

### Step 3：推送镜像到仓库

```bash
# 推送后端镜像
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest

# 推送前端镜像
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest
```

---

### Step 4：部署到 K8s

```bash
# 确认 kubectl 指向目标集群
kubectl cluster-info

# 一键部署前后端
kubectl apply -k deploy/

# 查看部署状态
kubectl get pods -n k8soperation -w
```

---

## 五、Apple Silicon (M1/M2/M3) 注意事项

Apple Silicon Mac 默认构建 `linux/arm64` 镜像。如果 K8s 集群是 **x86 (amd64)** 服务器，需要指定目标平台：

### 方式 A：指定 --platform（推荐）

```bash
# 后端 - 强制构建 amd64
docker build -f docker/backend/Dockerfile \
  --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest \
  .

# 前端 - 强制构建 amd64
docker build -f docker/frontend/Dockerfile \
  --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest \
  ./k8s-web
```

### 方式 B：多架构构建（同时支持 amd64 + arm64）

```bash
# 创建 buildx builder（首次执行）
docker buildx create --name multiarch --use
docker buildx inspect --bootstrap

# 后端 - 多架构构建并直接推送
docker buildx build -f docker/backend/Dockerfile \
  --platform linux/amd64,linux/arm64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest \
  --push \
  .

# 前端 - 多架构构建并直接推送
docker buildx build -f docker/frontend/Dockerfile \
  --platform linux/amd64,linux/arm64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest \
  --push \
  ./k8s-web
```

> **提示**：`--push` 会直接推送到仓库，无需单独 `docker push`。

---

## 六、一键构建脚本

如果你经常需要执行构建，可以用以下命令一次完成：

```bash
# ============================================================
# 一键构建 + 推送（在项目根目录执行）
# ============================================================

# 变量定义
REGISTRY="registry.cn-hangzhou.aliyuncs.com/k8s-gos"
BE_IMAGE="${REGISTRY}/devops-be:latest"
FE_IMAGE="${REGISTRY}/devops-fe:latest"
PLATFORM="linux/amd64"  # Apple Silicon 用户必须指定

# 1. 构建后端（Docker 内交叉编译）
echo "🔨 构建后端镜像..."
docker build -f docker/backend/Dockerfile --platform ${PLATFORM} -t ${BE_IMAGE} .

# 2. 构建前端（Docker 内 npm build）
echo "🔨 构建前端镜像..."
docker build -f docker/frontend/Dockerfile --platform ${PLATFORM} -t ${FE_IMAGE} ./k8s-web

# 3. 推送
echo "📦 推送镜像..."
docker push ${BE_IMAGE}
docker push ${FE_IMAGE}

# 4. 部署（可选）
echo "🚀 部署到 K8s..."
kubectl apply -k deploy/
kubectl rollout restart deployment/k8soperation -n k8soperation
kubectl rollout restart deployment/k8soperation-web -n k8soperation

echo "✅ 完成！"
```

---

## 七、版本标签管理

生产环境不建议只用 `latest`，推荐加上版本号：

```bash
# 以 git commit hash 作为版本号
VERSION=$(git rev-parse --short HEAD)

# 后端
docker build -f docker/backend/Dockerfile --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:${VERSION} \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest \
  .

# 前端
docker build -f docker/frontend/Dockerfile --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:${VERSION} \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest \
  ./k8s-web

# 推送所有标签
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:${VERSION}
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:${VERSION}
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest
```

K8s 部署时指定具体版本：
```bash
kubectl set image deployment/k8soperation \
  k8soperation=registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:abc1234 \
  -n k8soperation
```

---

## 八、常见问题

### Q1: 构建很慢？

**原因**：首次构建需要下载 Go 依赖和 npm 依赖。

**优化**：Docker 的层缓存机制会缓存 `go mod download` 和 `npm ci` 层，第二次构建会很快（通常 < 30 秒）。

```bash
# 查看缓存使用情况
docker system df

# 不要随意 prune，会清掉缓存
# docker builder prune  # 谨慎执行
```

### Q2: Apple Silicon 构建 amd64 性能差？

这是 QEMU 模拟的固有问题。优化方案：

1. **后端**：Go 交叉编译本身很快（纯编译，不依赖 QEMU）
2. **前端**：npm 阶段在 QEMU 下较慢，可以考虑先本地 `npm run build`，然后用纯打包 Dockerfile

### Q3: 网络超时（Go mod download / npm install）？

Dockerfile 内已配置国内镜像：
- Go：`GOPROXY=https://goproxy.cn,direct`
- npm：`registry https://registry.npmmirror.com`
- Alpine：`mirrors.aliyun.com`

如果仍超时，检查 Docker Desktop 网络设置（VPN/代理可能影响容器网络）。

### Q4: 镜像推送认证失败？

```bash
# 重新登录
docker login registry.cn-hangzhou.aliyuncs.com

# 检查凭证
cat ~/.docker/config.json | grep harbor
```

### Q5: K8s 拉取镜像失败 (ImagePullBackOff)？

```bash
# 确认集群已创建 harbor-secret
kubectl get secret harbor-secret -n k8soperation

# 如果没有，创建：
kubectl create secret docker-registry harbor-secret \
  --namespace=k8soperation \
  --docker-server=registry.cn-hangzhou.aliyuncs.com \
  --docker-username=YOUR_USER \
  --docker-password=YOUR_PASS
```

---

## 九、流程总结

```
macOS 开发机                          Linux K8s 集群
─────────────                         ──────────────
     │
     │ 1. docker build (多阶段，含交叉编译)
     │    ├── 后端: Dockerfile.multistage
     │    └── 前端: k8s-web/Dockerfile
     │
     │ 2. docker push (推到镜像仓库)
     │
     ├─────────── 镜像仓库 ──────────────┤
     │                                    │
     │                                    │ 3. kubectl apply -k deploy/
     │                                    │    (集群从仓库拉取镜像)
     │                                    │
     │                                    │ 4. Pod 启动运行
     │                                    │    ├── 后端: 8080
     │                                    │    └── 前端: 80 (Nginx)
     │                                    │
     │ 5. 验证                            │
     │    kubectl get pods -n k8soperation│
     │    kubectl logs ...                │
```

**核心回答：是的，交叉编译全部放在 Dockerfile 多阶段构建里面完成，macOS 上只需要执行 `docker build` 即可生成 Linux 镜像，不需要额外操作。**
