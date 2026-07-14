# K8sOperation 平台完整部署指南

## 架构

```
用户 → 前端 (Nginx :80) ──/api/*──→ 后端 (8080)
                         ──/     ──→ 静态文件

前端 Service (ClusterIP) ─→ NodePort 30007（对外入口）

后端 Service (ClusterIP) ← 前端 Nginx 反向代理
后端 ServiceAccount + ClusterRole → 管理 K8s 多集群资源

Jenkins (独立 namespace: devops)
  ├── ClusterIP :8080（Web UI）
  ├── NodePort 30080（对外入口）
  ├── Agent :50000（JNLP）
  └── 动态 Pod Agent → 每次构建创建独立 Pod
```

---

## 一、环境准备

### 1.1 必要条件

| 组件 | 说明 |
|------|------|
| Kubernetes | ≥ 1.25，k3s 亦可 |
| MySQL | 已在集群内（`mysql.default.svc:3306`） |
| Redis | 已在 1.117.227.207:6379 |
| Jenkins | 部署在 devops namespace |
| NFS Server | 集群内节点（10.0.16.2） |
| 镜像仓库 | 阿里云 ACR `registry.cn-hangzhou.aliyuncs.com/k8s-gos/` |

### 1.2 NFS 配置

```bash
# /etc/exports
/data/nfs            10.0.16.0/24(rw,sync,no_subtree_check,no_root_squash)
/data/k8soperation   10.0.16.0/24(rw,sync,no_subtree_check,no_root_squash)
/data/jenkins        10.0.16.0/24(rw,sync,no_subtree_check,no_root_squash)

# 重载
exportfs -ra
```

### 1.3 创建存储目录

```bash
# 后端（全新创建）
mkdir -p /data/k8soperation/{artifacts,logs,agents}

# Jenkins（已有数据，确认存在即可）
ls /data/jenkins/{home,go-cache,maven-cache,npm-cache,pip-cache}
```

### 1.4 数据库初始化

```bash
mysql -u root -p < docs/sql/k8s_platform_full_init.sql
```

---

## 二、端口规划

| 端口 | 服务 | 类型 | 说明 |
|------|------|------|------|
| 80 | 前端 | ClusterIP | Nginx 静态 + 反向代理 |
| 8080 | 后端 | ClusterIP | 仅内部，不对外 |
| 50000 | Jenkins | ClusterIP | JNLP Agent 通信 |
| 30007 | 前端 | NodePort | 对外访问入口（默认注释） |
| 30080 | Jenkins | NodePort | Jenkins Web UI |
| 30081 | 后端 | NodePort | 调试用（默认注释） |

---

## 三、部署步骤

### 3.1 后端

```bash
# 1. 配置 Secret（敏感信息）
cp deploy/backend/secret.yaml.example deploy/backend/secret.yaml
# 编辑 secret.yaml，DB/Redis/Jenkins 连接信息已预设

# 2. 修改 PV NFS Server IP（如果不是 10.0.16.2）
vi deploy/backend/pv.yaml

# 3. 一键部署
kubectl apply -k deploy/backend/

# 4. 验证
kubectl -n k8soperation get pods
kubectl -n k8soperation get svc
kubectl -n k8soperation get pvc
```

### 3.2 前端

```bash
# 1. 修改 ConfigMap Nginx upstream（如需）
vi deploy/frontend/configmap.yaml

# 2. 部署
kubectl apply -k deploy/frontend/

# 3. 对外暴露（选一种）
# 方式A：NodePort
#   编辑 deploy/frontend/kustomization.yaml，取消 service-nodeport.yaml 注释
#   重新 kubectl apply -k deploy/frontend/

# 方式B：Ingress
#   编辑 deploy/frontend/kustomization.yaml，取消 ingress.yaml 注释
#   修改 ingress.yaml 中的域名
```

### 3.3 Jenkins

```bash
# 1. 配置 Secret
vi deploy/jenkins/secret.yaml
# 修改项：
#   - admin-password（Jenkins 管理员密码）
#   - registry-username/password（阿里云 ACR）
#   - gitee-username/password（Gitee Token）
#   - maven-repo-password（Java 项目必填）

# 2. 部署
kubectl apply -k deploy/jenkins/

# 3. 等待就绪
kubectl -n devops wait --for=condition=Ready pod/jenkins-0 --timeout=300s
```

### 3.4 安装 Jenkins 插件

Jenkins → Manage Jenkins → Plugins → Available plugins，搜索安装以下 11 个：

```
workflow-aggregator          Pipeline 引擎
kubernetes                   K8s 动态 Pod Agent
git                          代码拉取
http_request                 平台回调
credentials-binding          凭证注入
pipeline-utility-steps       JSON/文件操作
configuration-as-code        JCasC 自动配置
sonar                        SonarQube 扫描（可选）
throttle-concurrents         并发构建限制（可选）
junit                        测试报告（可选）
blueocean                    可视化 UI（可选）
```

完整清单见：`deploy/jenkins/plugins.txt`

### 3.5 创建 4 个 Pipeline Job

Jenkins → New Item → Pipeline，创建以下 Job，每个选择 "Pipeline script from SCM"：

| Job 名称 | Script Path |
|----------|-------------|
| `go-pipeline` | `configs/jenkins-templates/go-pipeline.groovy` |
| `java-spring-pipeline` | `configs/jenkins-templates/java-spring-pipeline.groovy` |
| `frontend-pipeline` | `configs/jenkins-templates/frontend-pipeline.groovy` |
| `python-pipeline` | `configs/jenkins-templates/python-pipeline.groovy` |

---

## 四、平台初始配置

1. **添加 K8s 集群**：前端 → 集群管理 → 添加集群，上传 kubeconfig
2. **创建流水线**：选择语言类型 → 填 Git 仓库 → 配置目标部署
3. **运行流水线**：点击"运行"，Jenkins 构建 → 回调平台 → 自动部署到 K8s

---

## 五、文件清单

### 后端 `deploy/backend/`

| 文件 | 说明 |
|------|------|
| `namespace.yaml` | 创建 `k8soperation` namespace |
| `secret.yaml` | DB/Redis/Jenkins/JWT/加密等敏感配置 |
| `configmap.yaml` | 非敏感配置（app.yaml 含占位符 `${VAR}`） |
| `pv.yaml` | NFS PersistentVolume × 3（制品/日志/探针） |
| `pvc.yaml` | PersistentVolumeClaim × 3 |
| `service.yaml` | ClusterIP Service + ServiceAccount + ClusterRole + ClusterRoleBinding |
| `service-nodeport.yaml` | NodePort 30081（调试用，默认禁用） |
| `deployment.yaml` | Deployment + 探针 + 资源限制 |
| `ingress.yaml` | Ingress（可选） |
| `kustomization.yaml` | Kustomize 编排文件 |

### 前端 `deploy/frontend/`

| 文件 | 说明 |
|------|------|
| `configmap.yaml` | Nginx 配置（upstream → 后端 Service） |
| `service.yaml` | ClusterIP Service :80 |
| `service-nodeport.yaml` | NodePort 30007（对外，默认禁用） |
| `deployment.yaml` | Deployment + 资源限制 |
| `ingress.yaml` | Ingress（可选） |
| `kustomization.yaml` | Kustomize 编排文件 |

### Jenkins `deploy/jenkins/`

| 文件 | 说明 |
|------|------|
| `namespace.yaml` | 创建 `devops` namespace |
| `secret.yaml` | Jenkins 凭证（admin/registry/gitee/maven/sonar） |
| `configmap.yaml` | JCasC 自动配置（K8s Cloud + 凭证 + SonarQube） |
| `pv.yaml` | NFS PersistentVolume × 5（主数据+4缓存） |
| `pvc.yaml` | PersistentVolumeClaim × 5 |
| `rbac.yaml` | ServiceAccount + ClusterRole（动态 Pod Agent 权限） |
| `service.yaml` | ClusterIP :8080 + :50000 + NodePort 30080 |
| `statefulset.yaml` | StatefulSet + 探针 + init container |
| `plugins.txt` | 必装插件清单 |
| `kustomization.yaml` | Kustomize 编排文件 |

---

## 六、一键部署命令

```bash
# 全部部署
kubectl apply -k deploy/backend/
kubectl apply -k deploy/frontend/
kubectl apply -k deploy/jenkins/

# 查看状态
kubectl -n k8soperation get pods,pvc,svc
kubectl -n devops get pods,pvc,svc
```

---

## 七、常见问题

### PVC Pending

```bash
# 检查 PV 是否存在
kubectl get pv | grep k8soperation
kubectl get pv | grep jenkins

# 确保 NFS 目录存在
ls /data/k8soperation/
ls /data/jenkins/

# 确保 NFS export 已生效
showmount -e 10.0.16.2
```

### 后端启动报 K8s 集群初始化失败

正常现象 — 后端先空启动，然后通过前端 UI 添加集群即可。

### Jenkins Pod 调度失败

```bash
# 检查节点资源
kubectl top nodes

# 降低资源请求（编辑 statefulset.yaml）
# requests: 1C/2Gi → 500m/1Gi
```

### 镜像拉取失败

```bash
# 创建 imagePullSecret
kubectl -n k8soperation create secret docker-registry aliyun-registry \
  --docker-server=registry.cn-hangzhou.aliyuncs.com \
  --docker-username=15862326490 \
  --docker-password=dc521521..0
```

### Redis 连接失败

```bash
# 检查 Redis 地址
kubectl -n k8soperation get secret k8soperation-secret -o jsonpath='{.data.REDIS_ADDRESS}' | base64 -d

# secret.yaml 中配置:
#   REDIS_ADDRESS: "1.117.227.207:6379"
```
