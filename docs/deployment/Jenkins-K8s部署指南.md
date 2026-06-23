# Jenkins K8s 集群部署指南（NodePort 模式）

> 将 Jenkins 部署到 K8s 集群内部，通过 NodePort 30080 对外暴露 Web UI。
> Jenkins 使用 Kubernetes Plugin 动态创建构建 Pod Agent。

---

## 前置条件

- K8s 集群已就绪（`kubectl cluster-info` 正常）
- 集群有默认 StorageClass（用于 jenkins-data PVC 动态供给）
- 节点端口 30080 未被占用

---

## 部署文件结构

```
deploy/jenkins/
├── kustomization.yaml    # Kustomize 入口
├── namespace.yaml        # devops 命名空间
├── rbac.yaml             # ServiceAccount + ClusterRole（动态 Pod Agent 权限）
├── secret.yaml           # 敏感配置（密码、凭证，base64 编码）
├── pv.yaml               # 构建缓存 hostPath PV（Go/Maven/NPM/Pip）
├── pvc.yaml              # PVC（jenkins-data + 4 个缓存）
├── configmap.yaml        # JCasC 自动化配置
├── statefulset.yaml      # Jenkins Master Pod
└── service.yaml          # ClusterIP + NodePort Service
```

---

## 第一步：修改 Secret（凭证配置）

编辑 `deploy/jenkins/secret.yaml`，替换为你的实际凭证：

```bash
# 生成 base64 编码
echo -n "你的密码" | base64
```

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `admin-password` | Jenkins 管理员密码 | `admin123` |
| `hmac-secret` | HMAC 签名密钥（需与后端一致） | `a9XkP7mQ2vNc8Lr4TzY6HwFd3BjUs5Ge` |
| `sonarqube-token` | SonarQube Token（不用可保持默认） | `your-sonarqube-token` |
| `registry-username` | 镜像仓库用户名（阿里云 ACR） | 按实际填写 |
| `registry-password` | 镜像仓库密码 | 按实际填写 |
| `gitee-username` | Git 用户名 | 按实际填写 |
| `gitee-password` | Git 密码或 Personal Access Token | 按实际填写 |

---

## 第二步：修改 ConfigMap（回调地址）

编辑 `deploy/jenkins/configmap.yaml`，确认回调地址指向后端 Service：

```yaml
globalNodeProperties:
  - envVars:
      env:
        - key: "PLATFORM_CALLBACK_URL"
          value: "http://k8soperation.k8soperation.svc:8080/api/v1/k8s/cicd/pipeline/callback"
        - key: "ARTIFACT_UPLOAD_URL"
          value: "http://k8soperation.k8soperation.svc:8080/api/v1/k8s/cicd/artifact/upload"
```

> **注意**：如果后端还没部署，这个地址暂时不通，但不影响 Jenkins 启动。后端部署后自动可达。

---

## 第三步：一键部署

```bash
# 部署 Jenkins 所有资源
kubectl apply -k deploy/jenkins/
```

该命令会创建以下资源：

| 资源 | 名称 | 说明 |
|------|------|------|
| Namespace | `devops` | Jenkins 专用命名空间 |
| ServiceAccount | `jenkins` | Pod Agent 使用的服务账号 |
| ClusterRole | `jenkins-agent-manager` | 动态创建/删除构建 Pod 的权限 |
| Secret | `jenkins-secret` | 密码、凭证 |
| PV × 4 | `jenkins-*-cache-pv` | Go/Maven/NPM/Pip 构建缓存 |
| PVC × 5 | `jenkins-data` + 4 缓存 | 持久化存储 |
| ConfigMap | `jenkins-casc-config` | JCasC 自动化配置 |
| StatefulSet | `jenkins` | Jenkins Master Pod |
| Service | `jenkins` (ClusterIP) | 集群内部访问 |
| Service | `jenkins-agent` (ClusterIP) | Agent 通信 |
| Service | `jenkins-nodeport` (NodePort) | 外部访问 Web UI |

---

## 第四步：等待 Jenkins 就绪

```bash
# 查看 Pod 状态
kubectl get pods -n devops -w

# 等待就绪（首次启动约 2-5 分钟，需下载插件）
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=jenkins -n devops --timeout=600s
```

**首次启动较慢的原因：**
1. initContainer 会自动安装插件（kubernetes、git、pipeline 等）
2. 如果网络不好，插件下载可能超时，Jenkins 启动后可在 UI 手动安装

查看启动日志：
```bash
# 查看 init 容器日志（插件安装）
kubectl logs -n devops jenkins-0 -c install-plugins

# 查看 Jenkins 主容器日志
kubectl logs -n devops jenkins-0 -c jenkins -f
```

---

## 第五步：访问 Jenkins

```bash
# 查看 NodePort Service
kubectl get svc jenkins-nodeport -n devops
```

浏览器访问：
```
http://<节点IP>:30080
```

登录凭证：
- 用户名：`ops-dev`
- 密码：`admin123`（或你在 secret.yaml 中设置的密码）

---

## 第六步：验证配置

### 6.1 验证 Kubernetes Cloud 配置

进入 Jenkins → 系统管理 → 节点管理 → Configure Clouds：
- Kubernetes URL: `https://kubernetes.default.svc.cluster.local`
- Jenkins URL: `http://jenkins.devops.svc.cluster.local:8080`
- Jenkins tunnel: `jenkins-agent.devops.svc.cluster.local:50000`

### 6.2 验证全局凭证

进入 Jenkins → 系统管理 → Credentials：
- `hmac-secret` — HMAC 签名密钥
- `harbor-registry` — 镜像仓库用户名/密码
- `gitee-id` — Git 用户名/密码
- `sonarqube-token` — SonarQube Token

### 6.3 验证全局环境变量

进入 Jenkins → 系统管理 → 系统配置 → 全局属性 → 环境变量：
- `PLATFORM_CALLBACK_URL` = `http://k8soperation.k8soperation.svc:8080/api/v1/k8s/cicd/pipeline/callback`
- `ARTIFACT_UPLOAD_URL` = `http://k8soperation.k8soperation.svc:8080/api/v1/k8s/cicd/artifact/upload`

---

## 第七步：获取 API Token（后端需要）

后端平台需要 Jenkins API Token 来触发构建：

1. 登录 Jenkins → 点击右上角用户名 → Configure
2. 找到 **API Token** 区域 → 点击 **Add new Token**
3. 输入名称（如 `k8soperation`） → Generate
4. **复制生成的 Token**（只显示一次！）
5. 将 Token 填入后端 `deploy/backend/secret.yaml` 的 `JENKINS_API_TOKEN` 字段

---

## 存储说明

| PVC 名称 | 容量 | 存储方式 | 用途 |
|----------|------|---------|------|
| `jenkins-data` | 20Gi | 默认 StorageClass 动态供给 | Jenkins 主目录 |
| `jenkins-go-cache` | 10Gi | hostPath `/data/jenkins/go-cache` | Go 模块缓存 |
| `jenkins-maven-cache` | 20Gi | hostPath `/data/jenkins/maven-cache` | Maven 本地仓库 |
| `jenkins-npm-cache` | 10Gi | hostPath `/data/jenkins/npm-cache` | npm 缓存 |
| `jenkins-pip-cache` | 5Gi | hostPath `/data/jenkins/pip-cache` | pip 缓存 |

> 单节点集群使用 hostPath，目录会自动创建（`DirectoryOrCreate`）。

---

## Service 端口说明

| Service | 类型 | 端口 | 用途 |
|---------|------|------|------|
| `jenkins` | ClusterIP | 8080 | 集群内部 API 调用（后端 → Jenkins） |
| `jenkins-agent` | ClusterIP | 50000 | 构建 Agent 连接 Master |
| `jenkins-nodeport` | NodePort | 30080 | 浏览器访问 Jenkins Web UI |

---

## 常见问题

### Q1: Pod 一直 Pending
```bash
kubectl describe pod jenkins-0 -n devops
```
常见原因：
- PVC 未绑定 → 检查 StorageClass 是否存在：`kubectl get sc`
- 资源不足 → Jenkins 需要至少 2Gi 内存

### Q2: 插件安装失败
```bash
kubectl logs -n devops jenkins-0 -c install-plugins
```
解决方案：
- 网络问题 → Jenkins 启动后在 UI 手动安装
- 或进入 Pod 手动安装：`kubectl exec -it jenkins-0 -n devops -- jenkins-plugin-cli --plugins kubernetes git`

### Q3: 无法访问 30080 端口
```bash
# 确认 Service 存在
kubectl get svc jenkins-nodeport -n devops

# 确认 Pod 就绪
kubectl get pods -n devops

# 如果是云服务器，检查安全组是否放开 30080 端口
```

### Q4: JCasC 配置未生效
```bash
# 重启 Jenkins 重新加载配置
kubectl rollout restart statefulset jenkins -n devops
```

---

## 下一步

Jenkins 部署完成后，继续部署：
1. **后端服务** → `kubectl apply -k deploy/backend/`
2. **前端服务** → `kubectl apply -k deploy/frontend/`

详见：[Deploy部署配置修改指南.md](./Deploy部署配置修改指南.md)
