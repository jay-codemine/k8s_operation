# K8sOperation 平台部署手册

## 一、环境准备

### 1.1 必要条件

| 组件 | 版本要求 | 说明 |
|------|----------|------|
| Kubernetes | ≥ 1.25 | 集群已就绪 |
| MySQL | ≥ 5.7 | 数据库 |
| Redis | ≥ 6.0 | 单节点或 Cluster |
| Jenkins | 2.555.x LTS | CI/CD 引擎 |
| kubectl | 最新版 | K8s 命令行工具 |
| 镜像仓库 | Harbor / ACR | 存储构建产物 |

### 1.2 创建节点目录

```bash
# Jenkins hostPath PV 目录（在 K8s 节点上）
mkdir -p /data/jenkins/{home,go-cache,maven-cache,npm-cache,pip-cache}
```

### 1.3 准备数据库

```sql
-- 全新安装
mysql -u root -p < docs/sql/k8s_platform_full_init.sql

-- 已有库升级（金丝雀部署功能）
mysql -u root -p < scripts/migrations/002_add_canary_deploy.sql
```

---

## 二、部署后端

### 2.1 配置 Secret

```bash
cp deploy/backend/secret.yaml.example deploy/backend/secret.yaml
# 编辑 secret.yaml，填写真实数据库/Redis/Jenkins 连接信息
```

### 2.2 部署

```bash
# 默认部署（namespace + configmap + service + deployment）
kubectl apply -k deploy/backend/

# 如需持久化存储 + NodePort 暴露
# 编辑 deploy/backend/kustomization.yaml，取消 pv.yaml、pvc.yaml、service-nodeport.yaml 注释
```

### 2.3 验证

```bash
kubectl -n k8soperation get pods
kubectl -n k8soperation logs deployment/k8soperation
```

---

## 三、部署前端

```bash
kubectl apply -k deploy/frontend/

# 如需 NodePort 暴露（无需 Ingress Controller）
# 编辑 deploy/frontend/kustomization.yaml，取消 service-nodeport.yaml 注释
```

---

## 四、部署 Jenkins

### 4.1 配置 Secret

```bash
# 编辑 deploy/jenkins/secret.yaml，修改以下占位符：

# SonarQube Token
sonarqube-token: <base64编码的真实Token>

# Maven 私有仓库密码
maven-repo-password: <base64编码的真实密码>

# 其他凭证（Git/Registry/Admin）按需修改
```

### 4.2 部署

```bash
kubectl apply -k deploy/jenkins/

# 等待 Jenkins Pod 就绪
kubectl -n devops wait --for=condition=Ready pod/jenkins-0 --timeout=300s

# 获取初始密码（如果 Secret 中的 admin-password 不生效）
kubectl -n devops exec jenkins-0 -- cat /var/jenkins_home/secrets/initialAdminPassword
```

### 4.3 安装插件（详细步骤）

Jenkins 启动后，第一次登录需要安装插件，有两种方式：

**方式一：Web UI 安装（推荐）**

1. 浏览器打开 `http://<节点IP>:<NodePort>` 或通过 Ingress 访问
2. 使用 `admin / admin123`（Secret 中配置的密码）登录
3. 进入 **Manage Jenkins** → **Plugins** → **Available plugins**
4. 搜索并勾选以下 10 个插件，点击 **Install without restart**

**方式二：批量安装（jenkins-plugin-cli）**

```bash
kubectl -n devops exec -it jenkins-0 -- bash
cd /tmp
wget https://github.com/jenkinsci/plugin-installation-manager-tool/releases/download/2.13.0/jenkins-plugin-manager-2.13.0.jar

java -jar jenkins-plugin-manager-2.13.0.jar \
  --war /usr/share/jenkins/jenkins.war \
  --plugin-file /tmp/plugins.txt \
  --plugin-download-directory /var/jenkins_home/plugins \
  --jenkins-version 2.555.3
```

#### 插件清单说明

| 分类 | 插件 | 用途 |
|------|------|------|
| **必须** | workflow-aggregator | Pipeline 流水线引擎 |
| **必须** | kubernetes | K8s 动态 Pod Agent |
| **必须** | git | Git 代码拉取 |
| **必须** | http_request | HTTP 请求（平台回调 `httpRequest`） |
| **必须** | credentials-binding | 凭证注入（`withCredentials`） |
| **必须** | pipeline-utility-steps | JSON 文件操作（`readJSON`/`writeFile`） |
| 可选 | sonar | SonarQube 代码扫描 |
| 可选 | throttle-concurrents | 并发构建数量限制 |
| 可选 | junit | 单元测试报告（Java 模板使用） |
| 可选 | blueocean | 可视化 Pipeline UI |

> **注意**：`throttle-concurrents` 用于控制同一流水线最大并发构建数。
> 平台 `config.yaml` 的 `Jenkins.MaxConcurrentBuilds` 参数（默认 10）会自动注入到流水线模板的 `MAX_CONCURRENT_BUILDS` 参数，
> 该插件根据此参数限制同时运行的构建数量，防止并发过多导致集群资源耗尽。

所有插件安装完成后，**重启 Jenkins**：

```
Manage Jenkins → 页面上方黄色提醒 → 点击重启
或：http://<Jenkins地址>/safeRestart
```

### 4.4 验证插件安装

Manage Jenkins → Plugins → Installed plugins，确认 10 个插件状态为 `Enabled`。

### 4.5 创建 4 个 Pipeline Job

Jenkins → New Item → Pipeline，创建以下 Job：

| Job 名称 | Script Path（Pipeline script from SCM） |
|----------|----------------------------------------|
| `go-pipeline` | `configs/jenkins-templates/go-pipeline.groovy` |
| `java-spring-pipeline` | `configs/jenkins-templates/java-spring-pipeline.groovy` |
| `frontend-pipeline` | `configs/jenkins-templates/frontend-pipeline.groovy` |
| `python-pipeline` | `configs/jenkins-templates/python-pipeline.groovy` |

每个 Job 配置：
- Definition: `Pipeline script from SCM`
- SCM: Git
- Repository URL: 平台代码仓库地址
- Credentials: 对应 Git 凭证

---

## 五、平台初始配置

### 5.1 添加 K8s 集群

打开前端 → 集群管理 → 添加集群，上传目标集群的 kubeconfig。

### 5.2 创建流水线

打开前端 → 流水线管理 → 创建流水线：
1. 填写应用名称、Git 仓库
2. 选择语言类型（Go / Java / Node.js / Python）
3. 配置目标部署（集群、Namespace、工作负载名称）
4. 可选：启用金丝雀部署（设置流量比例和观察时长）

### 5.3 运行流水线

流水线详情页 → 点击"运行"，Jenkins 自动触发构建 → 回调平台 → 自动部署。

---

## 六、Jenkins 种子 Job 自动化（可选）

批量创建 4 个 Pipeline Job，避免手动操作：

```groovy
// Jenkins → New Item → Pipeline → 命名为 "seed-job"
// 将下方脚本粘贴到 Pipeline Script

def jobs = [
    [name: 'go-pipeline',       script: 'configs/jenkins-templates/go-pipeline.groovy'],
    [name: 'java-spring-pipeline', script: 'configs/jenkins-templates/java-spring-pipeline.groovy'],
    [name: 'frontend-pipeline', script: 'configs/jenkins-templates/frontend-pipeline.groovy'],
    [name: 'python-pipeline',   script: 'configs/jenkins-templates/python-pipeline.groovy'],
]

jobs.each { job ->
    pipelineJob(job.name) {
        definition {
            cpsScm {
                scm {
                    git {
                        remote { url('你的平台代码仓库地址') }
                        branch('*/main')
                    }
                }
                scriptPath(job.script)
            }
        }
    }
}
```

---

## 七、常见问题

### 后端启动报 K8s 集群初始化失败

```yaml
# configs/config.yaml
AutoInitK8s: false    # 改为 false，通过前端 UI 添加集群
AllowEmptyStart: true  # 允许空集群启动，不影响登录/CICD 功能
```

### Redis 集群模式

```yaml
# config.yaml
Cache:
  Address: ""                    # 单节点留空
  Addresses:                     # 集群模式
    - "redis-node1:6379"
    - "redis-node2:6379"
    - "redis-node3:6379"
  Username: ""                   # 可选
  Password: "your-password"      # 可选
```

### Jenkins Pod Agent 创建失败

检查：
1. `kubectl -n devops get sa jenkins` — ServiceAccount 是否存在
2. `kubectl auth can-i create pods --as=system:serviceaccount:devops:jenkins` — RBAC 权限
3. Jenkins Cloud 配置中 K8s URL 是否正确

### 构建缓存 PV 未绑定

```bash
# 确认 PV 和 PVC 匹配
kubectl get pv | grep jenkins
kubectl -n devops get pvc

# 如果 PVC Pending，检查 StorageClass
kubectl get pvc -n devops -o yaml | grep storageClassName
```
