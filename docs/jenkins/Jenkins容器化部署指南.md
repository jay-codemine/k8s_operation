# Jenkins 全容器化部署指南

## 架构概述

```
┌─────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                     │
│                                                          │
│  ┌─────────────────┐    ┌──────────────────────────┐    │
│  │  devops namespace│    │ k8soperation namespace   │    │
│  │                  │    │                          │    │
│  │  ┌────────────┐ │    │  ┌─────────┐  ┌──────┐  │    │
│  │  │  Jenkins   │ │◄───┤  │ Backend │  │Front │  │    │
│  │  │  Master    │ │    │  │  API    │  │ end  │  │    │
│  │  │(StatefulSet)│ │───►│  └─────────┘  └──────┘  │    │
│  │  └────────────┘ │    │                          │    │
│  │        │        │    └──────────────────────────┘    │
│  │        ▼        │                                     │
│  │  ┌────────────┐ │                                     │
│  │  │ Dynamic Pod│ │  ┌─────────────────┐               │
│  │  │  Agents    │ │  │  PVC (缓存)     │               │
│  │  │ ┌────┐┌──┐│ │  │ - maven-cache   │               │
│  │  │ │build││ka││ │  │ - go-cache      │               │
│  │  │ │tool ││ni││◄┤  │ - npm-cache     │               │
│  │  │ └────┘│ko││ │  │ - pip-cache     │               │
│  │  │       └──┘│ │  └─────────────────┘               │
│  │  └────────────┘ │                                     │
│  └─────────────────┘                                     │
└─────────────────────────────────────────────────────────┘
```

**核心变更**：
- Jenkins Master：VM → K8s StatefulSet
- 构建 Agent：固定节点 → 动态 Pod Agent（按需创建、用完销毁）
- 镜像构建：nerdctl/docker → Kaniko（无 daemon、无特权）
- 缓存：本地磁盘 → PVC 持久化卷

---

## 前置条件

| 组件 | 要求 |
|------|------|
| Kubernetes | >= 1.24 |
| StorageClass | 支持 ReadWriteOnce + ReadWriteMany |
| 网络 | Pod 间可互相通信 |
| 镜像仓库 | Harbor/DockerHub/阿里云 ACR 等 |

---

## 一、部署 Jenkins

### 1.1 修改 Secret（必须）

编辑 `deploy/jenkins/secret.yaml`，替换为你的实际密码：

```bash
# 生成 base64 编码
echo -n "your-strong-password" | base64
# 替换 secret.yaml 中的 admin-password 值
```

### 1.2 修改 StorageClass（按需）

如果集群使用非默认 StorageClass，编辑 `deploy/jenkins/pvc.yaml`：

```yaml
spec:
  storageClassName: "your-storage-class"  # 取消注释并填写
```

### 1.3 一键部署

```bash
# 部署 Jenkins
kubectl apply -k deploy/jenkins/

# 验证部署状态
kubectl -n devops get pods -w
kubectl -n devops get svc
```

### 1.4 等待就绪

```bash
# Jenkins 启动较慢（约 2-3 分钟），等待 Ready
kubectl -n devops rollout status statefulset/jenkins --timeout=300s
```

### 1.5 访问 Jenkins

| 方式 | 地址 |
|------|------|
| 集群内 | `http://jenkins.devops.svc.cluster.local:8080` |
| NodePort | `http://<任意节点IP>:30080` |

---

## 二、Jenkins 初始配置

Jenkins 使用 JCasC（Configuration as Code）自动完成以下配置：
- Kubernetes Cloud 连接
- 管理员用户
- 安全策略

### 2.1 手动配置镜像仓库凭证

登录 Jenkins → 「Manage Jenkins」→「Credentials」→ 添加：

| 字段 | 值 |
|------|------|
| Kind | Username with password |
| ID | `robot$test-k8soperation`（与 config.yaml 中 RegistryCredentialID 一致） |
| Username | 镜像仓库用户名 |
| Password | 镜像仓库密码/Token |

### 2.2 配置 Git 凭证

| 字段 | 值 |
|------|------|
| Kind | Username with password |
| ID | `k8soperation`（与 config.yaml 中 GitCredentialID 一致） |
| Username | Git 用户名 |
| Password | Git 密码/Token |

### 2.3 配置 HMAC 凭证

| 字段 | 值 |
|------|------|
| Kind | Secret text |
| ID | `hmac-secret` |
| Secret | 与 config.yaml 中 HMACSecret 值相同 |

---

## 三、Pipeline 模板说明

所有 Pipeline 模板已改造为 **K8s Pod Agent + Kaniko** 模式：

| 模板 | 构建容器 | 缓存 PVC |
|------|---------|----------|
| go-pipeline.groovy | `golang:1.24` | jenkins-go-cache |
| java-spring-pipeline.groovy | `maven:3.9-eclipse-temurin-17` | jenkins-maven-cache |
| frontend-pipeline.groovy | `node:18-alpine` | jenkins-npm-cache |
| python-pipeline.groovy | `python:3.11-slim` | jenkins-pip-cache |

### 工作原理

```
1. 平台触发构建 → Jenkins API 创建 Job
2. Jenkins Kubernetes Plugin 在 devops namespace 创建临时 Pod
3. Pod 包含两个容器：
   - 构建工具容器（golang/maven/node/python）→ 编译代码
   - Kaniko 容器 → 构建镜像 + 推送到仓库
4. 构建完成 → Pod 自动销毁
5. Jenkins 回调平台 → 更新构建状态
```

### Kaniko 说明

- **无需安装**：Kaniko 作为容器镜像运行，不需要在节点上安装任何东西
- **无特权**：不需要 Docker daemon，不需要特权模式
- **带缓存**：支持 `--cache=true` 加速重复构建

---

## 四、平台配置

`configs/config.yaml` 中 Jenkins 相关配置已更新为 K8s Service 地址：

```yaml
Jenkins:
  URL: "http://jenkins.devops.svc.cluster.local:8080/"
  CallbackURL: "http://k8soperation-backend.k8soperation.svc.cluster.local:38180"
  PlatformURL: "http://k8soperation-frontend.k8soperation.svc.cluster.local"
```

> **注意**：如果 Jenkins 和平台不在同一集群，需要使用实际可达的地址（NodePort/Ingress/LoadBalancer）

---

## 五、验证部署

### 5.1 验证 Jenkins 连接

```bash
# 在平台后端 Pod 中测试连通性
kubectl -n k8soperation exec -it deploy/k8soperation-backend -- \
  curl -s -o /dev/null -w "%{http_code}" \
  http://jenkins.devops.svc.cluster.local:8080/api/json
# 期望返回 200 或 403（表示网络通，需要认证）
```

### 5.2 验证动态 Agent

在 Jenkins 中创建一个测试 Pipeline：

```groovy
pipeline {
    agent {
        kubernetes {
            yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: test
    image: alpine:3.19
    command: ['sleep', '30']
"""
        }
    }
    stages {
        stage('Test') {
            steps {
                container('test') {
                    sh 'echo "K8s Pod Agent works!"'
                }
            }
        }
    }
}
```

### 5.3 通过平台触发构建

在平台 Web UI 中创建一个测试项目，触发构建，观察：
1. Jenkins 是否成功创建 Job
2. 构建 Pod 是否在 devops namespace 中创建
3. 镜像是否成功推送到仓库
4. 构建状态是否回调到平台

---

## 六、故障排查

| 问题 | 排查方式 |
|------|---------|
| Jenkins 启动失败 | `kubectl -n devops logs statefulset/jenkins` |
| Pod Agent 创建失败 | `kubectl -n devops get events --sort-by='.lastTimestamp'` |
| Kaniko 推送失败 | 检查 Jenkins Credentials 中镜像仓库凭证是否正确 |
| 回调失败 | 检查 devops → k8soperation namespace 网络策略 |
| 缓存不生效 | `kubectl -n devops get pvc` 确认 PVC 正常 Bound |

---

## 七、资源清理

```bash
# 删除 Jenkins 部署
kubectl delete -k deploy/jenkins/

# 注意：PVC 默认不会被删除（保护数据），手动清理：
kubectl -n devops delete pvc jenkins-data jenkins-go-cache jenkins-maven-cache jenkins-npm-cache jenkins-pip-cache
```

---

## 八、与 VM 模式的对比

| 对比项 | VM 模式（旧） | K8s 容器化（新） |
|--------|--------------|-----------------|
| Jenkins 部署 | 虚拟机手动安装 | StatefulSet 声明式部署 |
| 构建环境 | 固定在 Jenkins VM 上 | 动态 Pod，按需创建 |
| 镜像构建 | nerdctl + buildkitd | Kaniko（无 daemon） |
| JDK/Maven/Go | VM 上全局安装 | 容器镜像自带 |
| 扩容 | 增加 VM | 自动 Pod 调度 |
| 资源利用 | VM 常驻占用 | 构建完即释放 |
| 安全性 | 需要 root/docker 权限 | 无特权容器 |
