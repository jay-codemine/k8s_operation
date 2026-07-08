# Jenkins 部署与插件安装指南

## 一、部署概览

| 项目 | 配置 |
|------|------|
| 部署方式 | K8s StatefulSet（单实例） |
| 命名空间 | `devops` |
| 镜像 | `jenkins/jenkins:lts` |
| 持久化 | PVC `jenkins-data`（挂载 `/var/jenkins_home`） |
| Web UI 访问 | NodePort `30080` → http://\<NodeIP\>:30080 |
| K8s 内部地址 | http://jenkins.devops.svc.cluster.local:8080 |
| Agent 通信端口 | 50000（ClusterIP `jenkins-agent`） |
| 管理员账号 | `ops-dev` / `admin123` |
| Setup Wizard | 已禁用（`-Djenkins.install.runSetupWizard=false`） |
| 配置方式 | JCasC（Configuration as Code） |
| 资源配额 | requests: 1C/2Gi, limits: 4C/8Gi |

---

## 二、插件安装

### 2.1 需要安装的 13 个插件

| # | 插件 ID | 插件名称 | 用途 |
|---|---------|---------|------|
| 1 | `kubernetes` | Kubernetes | K8s 动态 Pod Agent，每次构建创建独立 Pod |
| 2 | `workflow-aggregator` | Pipeline | Jenkins Pipeline 核心引擎 |
| 3 | `git` | Git | Git SCM 代码检出 |
| 4 | `configuration-as-code` | Configuration as Code (JCasC) | YAML 声明式配置 Jenkins |
| 5 | `credentials-binding` | Credentials Binding | Pipeline 中绑定和使用凭证 |
| 6 | `http_request` | HTTP Request | Pipeline 中发送 HTTP 请求（回调平台） |
| 7 | `pipeline-utility-steps` | Pipeline Utility Steps | readJSON、writeFile 等实用步骤 |
| 8 | `junit` | JUnit | 测试报告收集与展示 |
| 9 | `sonar` | SonarQube Scanner | SonarQube 代码质量扫描集成 |
| 10 | `pipeline-stage-view` | Pipeline: Stage View | 流水线阶段可视化 |
| 11 | `timestamper` | Timestamper | 构建日志添加时间戳 |
| 12 | `ws-cleanup` | Workspace Cleanup | 构建前/后清理工作空间 |
| 13 | `ansicolor` | AnsiColor | 控制台彩色输出支持 |

### 2.2 安装方式

#### 方式一：Init Container 自动安装（推荐）

部署时 init container 会自动安装上述 13 个插件：

```yaml
initContainers:
  - name: install-plugins
    image: jenkins/jenkins:lts
    command:
      - sh
      - -c
      - |
        echo "[init] 开始自动安装插件（超时5分钟）..."
        mkdir -p /var/jenkins_home/plugins
        timeout 300 jenkins-plugin-cli \
          --plugin-download-directory /var/jenkins_home/plugins \
          --plugins kubernetes workflow-aggregator git \
            configuration-as-code credentials-binding http_request \
            pipeline-utility-steps junit sonar \
            pipeline-stage-view timestamper ws-cleanup ansicolor \
          && echo "[init] 插件安装完成" \
          || echo "[init] 插件安装失败或超时，Jenkins 启动后可从 UI 手动安装"
```

**关键配置说明：**
- `--plugin-download-directory /var/jenkins_home/plugins`：必须指定此参数，否则插件安装到容器临时目录，Jenkins 主进程无法加载
- `timeout 300`：5 分钟超时，网络不佳时自动跳过，不阻塞 Jenkins 启动
- `|| true` 语义：安装失败不影响 Pod 启动，可事后通过 UI 补装

#### 方式二：Jenkins UI 手动安装

如果 init container 安装失败（网络问题等），可通过 UI 补装：

1. 访问 **http://\<NodeIP\>:30080/pluginManager/available**
2. 搜索插件名称，勾选后点击 **Install without restart**
3. 安装完成后建议重启 Jenkins（**Manage Jenkins → Restart**）

> **说明**：两种方式安装的插件存储在同一个 PVC 目录（`/var/jenkins_home/plugins/`），互不冲突、互相补充。

---

## 三、安装后验证

### 3.1 通过 API 验证

```bash
# 使用密码认证
curl -s -u "ops-dev:admin123" "http://127.0.0.1:30080/pluginManager/api/json?depth=1" \
  | python3 -c "
import json, sys
data = json.load(sys.stdin)
plugins = {p['shortName'] for p in data.get('plugins', [])}
needed = ['kubernetes','workflow-aggregator','git','configuration-as-code',
          'credentials-binding','http_request','pipeline-utility-steps',
          'junit','sonar','pipeline-stage-view','timestamper','ws-cleanup','ansicolor']
for p in needed:
    status = '✅' if p in plugins else '❌'
    print(f'{status} {p}')
"
```

### 3.2 通过 Jenkins UI 验证

访问 **Manage Jenkins → Plugins → Installed plugins**，确认 13 个插件均显示为已安装且启用。

---

## 四、生成 API Token

平台与 Jenkins 通信需要 API Token（非密码），需在部署后手动生成。

### 4.1 方式一：通过 curl 命令（推荐）

```bash
# 1. 获取 Crumb + Cookie（Jenkins CSRF 保护需要三要素：Token + Crumb + Cookie）
CRUMB=$(curl -s -c /tmp/jc -u "ops-dev:admin123" \
  'http://127.0.0.1:30080/crumbIssuer/api/json' \
  | python3 -c "import json,sys;print(json.load(sys.stdin)['crumb'])")

# 2. 生成 API Token
curl -s -b /tmp/jc -u "ops-dev:admin123" \
  -H "Jenkins-Crumb: $CRUMB" \
  'http://127.0.0.1:30080/user/ops-dev/descriptorByName/jenkins.security.ApiTokenProperty/generateNewToken' \
  --data 'newTokenName=platform-token'
```

### 4.2 方式二：通过 Script Console

1. 访问 **http://\<NodeIP\>:30080/script**
2. 执行以下 Groovy 脚本：

```groovy
import hudson.model.*
import jenkins.security.*

def user = User.getById("ops-dev", true)
def prop = user.getProperty(jenkins.security.ApiTokenProperty)
def token = prop.tokenStore.generateNewToken("platform-api-token")
println "API Token: ${token.plainValue}"
```

### 4.3 更新 Token 到配置

将生成的 Token 更新到：
- `configs/config.yaml` → `Jenkins.APIToken`
- `deploy/backend/secret.yaml` → `JENKINS_API_TOKEN`（base64 编码）

> **重要**：每次 Jenkins 彻底重装（删除 PVC）后，旧 Token 失效，必须重新生成。

---

## 五、凭证配置

Jenkins 构建流水线需要以下凭证（通过 JCasC 或 UI 配置）：

| 凭证 ID | 类型 | 用途 | 配置方式 |
|---------|------|------|---------|
| `hmac-secret` | Secret text | 平台回调 HMAC 签名 | JCasC 自动（Secret 注入） |
| `sonarqube-token` | Secret text | SonarQube 认证 | JCasC 自动（Secret 注入） |
| `gitee-id` | Username/Password | Git 仓库拉取代码 | UI 手动配置 |
| `harbor-registry` | Username/Password | 镜像仓库推送 | UI 手动配置 |

---

## 六、部署命令

```bash
# 首次部署
kubectl apply -k deploy/jenkins/

# 彻底重装（删除数据重建）
kubectl delete statefulset jenkins -n devops
kubectl delete pvc jenkins-data -n devops
kubectl apply -k deploy/jenkins/

# 查看状态
kubectl get pods -n devops -l app.kubernetes.io/name=jenkins
kubectl logs jenkins-0 -n devops -c install-plugins  # 查看插件安装日志
```

---

## 七、已知问题与故障排查

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| init container 卡住很久 | 网络下载插件慢 | 5 分钟超时后自动跳过，通过 UI 补装 |
| 插件安装成功但 Jenkins 没加载 | 缺少 `--plugin-download-directory` | 已修复，确保写入 `/var/jenkins_home/plugins` |
| API 返回 401 | Token 过期或重装后失效 | 重新生成 API Token 并更新 `config.yaml` |
| Pod CrashLoopBackOff | 内存不足或配置错误 | 检查 `kubectl describe pod` 和 `kubectl logs` |
| 登录页面空白（无用户名密码输入框） | JCasC 加载失败（`configuration-as-code` 在 `kubernetes` 插件之前加载） | 等所有插件安装完后 `kubectl delete pod jenkins-0 -n devops` 重启 |
| 流水线报 "Job不存在" | Job 名带 `.groovy` 后缀（Seed Job 自动创建导致） | 通过 Jenkins UI 或 API 重命名 Job，去掉 `.groovy` 后缀 |
| CSRF 403 "No valid crumb" | Jenkins 写操作需 Crumb+Cookie | 先 `curl -c cookie` 获取 crumb，再 `-b cookie -H Crumb` 发请求 |

### 7.1 JCasC 启动失败详解

**现象**：Jenkins 启动后页面空白，没有登录表单。

**根因**：init container 超时后，Jenkins 通过 UpdateCenter 在后台逐个安装插件。`configuration-as-code` 插件装好后立即尝试应用 JCasC 配置，但此时 `kubernetes` 插件尚未安装，导致：

```
io.jenkins.plugins.casc.ConfigurationAsCodeBootFailure
No hudson.slaves.Cloud implementation found for kubernetes
```

安全策略（`securityRealm`）配置未生效 → 无登录页面。

**修复**：等全部插件安装完后，删除 Pod 触发重建即可：

```bash
kubectl delete pod jenkins-0 -n devops
# StatefulSet 会自动重建，此次 JCasC 可正确加载
```

### 7.2 Job 名称 `.groovy` 后缀问题

**现象**：平台触发构建报错 `Job不存在: java-spring-pipeline`。

**根因**：Seed Job 从 Groovy 文件创建 Pipeline Job 时，默认使用文件名作为 Job 名（含 `.groovy` 后缀）。

**修复**：通过 API 批量重命名：

```bash
# 获取 Crumb + Cookie
CRUMB=$(curl -s -c /tmp/jc -u "ops-dev:<API_TOKEN>" \
  'http://127.0.0.1:30080/crumbIssuer/api/json' \
  | python3 -c "import json,sys;print(json.load(sys.stdin)['crumb'])")

# 重命名（每个 Job 需分别执行）
curl -s -b /tmp/jc -u "ops-dev:<API_TOKEN>" \
  -H "Jenkins-Crumb: $CRUMB" -X POST \
  'http://127.0.0.1:30080/job/java-spring-pipeline.groovy/doRename?newName=java-spring-pipeline'
```

需重命名的 Job 列表：
- `go-pipeline.groovy` → `go-pipeline`
- `java-spring-pipeline.groovy` → `java-spring-pipeline`
- `frontend-pipeline.groovy` → `frontend-pipeline`
- `python-pipeline.groovy` → `python-pipeline`
