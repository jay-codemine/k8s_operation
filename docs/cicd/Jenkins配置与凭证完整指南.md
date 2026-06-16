# Jenkins 配置与凭证完整指南

> 本文档详细说明 K8sOperation 平台与 Jenkins 对接所需的所有配置项、凭证创建、Job 配置、插件安装等。

---

## 目录

- [一、整体架构](#一整体架构)
- [二、Jenkins 必装插件](#二jenkins-必装插件)
- [三、Jenkins 凭证配置](#三jenkins-凭证配置)
- [四、Jenkins 全局工具配置](#四jenkins-全局工具配置)
- [五、Jenkins Job 创建](#五jenkins-job-创建)
- [六、平台后端配置（config.yaml）](#六平台后端配置configyaml)
- [七、SonarQube 集成配置（可选）](#七sonarqube-集成配置可选)
- [八、回调机制说明](#八回调机制说明)
- [九、验证清单](#九验证清单)
- [十、常见问题](#十常见问题)

---

## 一、整体架构

```
┌────────────────────────────────────────────────────────────────┐
│                      K8sOperation 平台                          │
│                                                                │
│  前端 UI ──→ 后端 API ───→ Jenkins API (触发构建)               │
│                  ↑                    │                         │
│                  │                    ▼                         │
│             回调接口 ←────── Jenkins Pipeline (构建完成回调)     │
│                                      │                         │
│                                      ▼                         │
│                              镜像仓库 (Harbor/ACR)              │
└────────────────────────────────────────────────────────────────┘
```

**通信关系：**
1. **平台 → Jenkins**：通过 Jenkins API + Basic Auth（用户名 + API Token）触发构建
2. **Jenkins → 平台**：构建完成后通过 HTTP 回调通知平台（HMAC 签名验证）
3. **Jenkins → Git**：拉取项目源代码（Git 凭证）
4. **Jenkins → 镜像仓库**：推送构建好的 Docker 镜像（Harbor 凭证）
5. **Jenkins → SonarQube**（可选）：代码质量扫描

---

## 二、Jenkins 必装插件

进入 **Manage Jenkins → Plugins → Available plugins**，搜索并安装：

### 必须安装

| 插件名称 | 用途 |
|---------|------|
| **Pipeline** | Pipeline 流水线支持（通常已预装） |
| **Pipeline: SCM Step** | 从 SCM 加载 Pipeline 脚本 |
| **Git** | Git 仓库拉取 |
| **Credentials Binding** | 凭证绑定到环境变量 |
| **HTTP Request** | Pipeline 中发送 HTTP 请求（回调平台用） |
| **Docker Pipeline** | Docker 镜像构建与推送 |
| **Pipeline Utility Steps** | `readJSON`、`writeFile` 等工具步骤 |

### 可选安装（按需）

| 插件名称 | 用途 |
|---------|------|
| **SonarQube Scanner** | 代码质量扫描集成 |
| **NodeJS** | 前端项目 Node.js 环境 |
| **Timestamper** | 构建日志加时间戳 |
| **Blue Ocean** | 美化 Pipeline 可视化 |
| **Build Timeout** | 构建超时控制 |

---

## 三、Jenkins 凭证配置

进入 **Manage Jenkins → Credentials → System → Global credentials → Add Credentials**

### 3.1 Git 仓库凭证（必须）

| 字段 | 值 |
|------|-----|
| **Kind** | Username with password |
| **Scope** | Global |
| **ID** | `gitee-id` |
| **Username** | Git 仓库用户名（如 Gitee 账号） |
| **Password** | Git 仓库密码 或 Personal Access Token |
| **Description** | Gitee/GitHub 代码拉取凭证 |

> **说明**：所有流水线模板默认使用 `gitee-id` 作为 Git 凭证 ID。如果你用其他 ID，需在平台创建流水线时指定 `GIT_CREDENTIAL_ID` 参数。

---

### 3.2 镜像仓库凭证（必须）

| 字段 | 值 |
|------|-----|
| **Kind** | Username with password |
| **Scope** | Global |
| **ID** | `harbor-registry` |
| **Username** | 镜像仓库用户名（如 Harbor/ACR 用户） |
| **Password** | 镜像仓库密码 |
| **Description** | Docker 镜像推送凭证（Harbor/ACR/DockerHub） |

> **说明**：流水线中通过 `credentials('harbor-registry')` 引用，绑定到 `REGISTRY_CREDS_USR` 和 `REGISTRY_CREDS_PSW` 环境变量。

**各镜像仓库示例：**

| 仓库类型 | Username | Password |
|---------|----------|----------|
| 阿里云 ACR | 阿里云账号（如 `your@email.com`） | 在 ACR 设置的固定密码 |
| Harbor | Harbor 用户名 | Harbor 密码 |
| DockerHub | Docker ID | Docker Access Token |

---

### 3.3 HMAC 签名密钥（必须）

| 字段 | 值 |
|------|-----|
| **Kind** | Secret text |
| **Scope** | Global |
| **ID** | `hmac-secret` |
| **Secret** | 与平台 `config.yaml` 中 `Jenkins.HMACSecret` **完全相同** 的字符串 |
| **Description** | 平台回调 HMAC 签名验证密钥 |

> **重要**：此密钥必须与后端配置文件中的 `HMACSecret` **完全一致**，否则回调签名验证会失败。

**生成随机密钥：**
```bash
# Linux/Mac
openssl rand -hex 32

# 输出示例：a7f9c2e1b4d5...（64 位十六进制字符串）
```

---

### 3.4 凭证一览表

| 凭证 ID | 类型 | 用途 | 必须 |
|---------|------|------|------|
| `gitee-id` | Username with password | Git 代码拉取 | ✅ |
| `harbor-registry` | Username with password | Docker 镜像推送 | ✅ |
| `hmac-secret` | Secret text | 回调签名验证 | ✅ |

---

## 四、Jenkins 全局工具配置

进入 **Manage Jenkins → Tools**

### 4.1 Go 环境（Go 项目必须）

Jenkins 服务器上需安装 Go：

```bash
# 安装 Go 1.24（示例）
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile
go version
```

流水线模板中的环境变量预设：
```groovy
GOROOT     = "/usr/local/go"
GOPATH     = "/var/lib/jenkins/go"
GOMODCACHE = "/var/lib/jenkins/go/pkg/mod"
GOCACHE    = "/var/lib/jenkins/.cache/go-build"
GOPROXY    = "https://goproxy.cn,direct"
```

### 4.2 Maven + JDK（Java 项目必须）

在 Tools 页面配置：

| 工具 | Name | 安装路径 |
|------|------|---------|
| **Maven** | `Maven-3.9` | `/opt/apache-maven-3.9.9` |
| **JDK** | `JDK-21` | `/usr/lib/jvm/java-21` |

**安装 Maven：**
```bash
cd /opt
wget https://mirrors.aliyun.com/apache/maven/maven-3/3.9.9/binaries/apache-maven-3.9.9-bin.tar.gz
tar xzf apache-maven-3.9.9-bin.tar.gz
echo 'export MAVEN_HOME=/opt/apache-maven-3.9.9' >> /etc/profile
echo 'export PATH=$PATH:$MAVEN_HOME/bin' >> /etc/profile
source /etc/profile
mvn --version
```

### 4.3 Node.js（前端项目必须）

```bash
# 安装 Node.js 18 LTS
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt-get install -y nodejs
node -v && npm -v

# 设置国内镜像
npm config set registry https://registry.npmmirror.com
```

### 4.4 Docker（所有项目必须）

```bash
# Jenkins 用户需要有 Docker 权限
usermod -aG docker jenkins
systemctl restart jenkins

# 验证
su - jenkins -c "docker info"
```

### 4.5 Python（Python 项目必须）

```bash
# 安装 Python 3.11
apt-get install -y python3.11 python3.11-venv python3-pip
update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.11 1

# 设置国内镜像
pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
```

---

## 五、Jenkins Job 创建

### 5.1 推荐方式：通用分发器（一个 Job 搞定所有）

创建一个 Pipeline Job，通过 `Jenkinsfile` 自动路由到对应模板：

1. **Jenkins → New Item → Pipeline** → 命名为 `k8s-platform-builder`
2. **Pipeline 区域：**
   - Definition: **Pipeline script from SCM**
   - SCM: **Git**
   - Repository URL: `https://gitee.com/jay-kim/k8s_operation.git`
   - Credentials: 选择 `gitee-id`
   - Branch Specifier: `*/main`
   - Script Path: `Jenkinsfile`
3. **勾选**：✅ This project is parameterized（平台会自动注入参数）
4. **保存**

### 5.2 按语言分离 Job（精细化管理）

| Job 名称 | Script Path | 适用项目 |
|----------|------------|---------|
| `k8s-builder-go` | `configs/jenkins-templates/go-pipeline.groovy` | Go |
| `k8s-builder-java` | `configs/jenkins-templates/java-spring-pipeline.groovy` | Java/Spring |
| `k8s-builder-frontend` | `configs/jenkins-templates/frontend-pipeline.groovy` | Vue/React |
| `k8s-builder-python` | `configs/jenkins-templates/python-pipeline.groovy` | Python |

**创建步骤（以 Go 为例）：**

1. **New Item → Pipeline** → 命名 `k8s-builder-go`
2. Pipeline → Definition: **Pipeline script from SCM**
3. SCM: Git → URL: `https://gitee.com/jay-kim/k8s_operation.git`
4. Credentials: `gitee-id`
5. Branch: `*/main`
6. Script Path: `configs/jenkins-templates/go-pipeline.groovy`
7. 保存

### 5.3 Job 参数说明

平台触发构建时会自动注入以下参数（无需手动填写）：

| 参数 | 说明 | 示例 |
|------|------|------|
| `GIT_REPO` | 项目 Git 仓库地址 | `https://gitee.com/org/myapp.git` |
| `GIT_BRANCH` | 构建分支 | `main` / `release/v1.0` |
| `IMAGE_REPO` | 镜像仓库地址 | `harbor.example.com/project/myapp` |
| `IMAGE_TAG` | 镜像标签（空则自动生成） | `v1.2.0` / `abc1234-20240101120000` |
| `LANGUAGE_TYPE` | 语言类型（分发器路由用） | `go` / `java` / `frontend` / `python` |
| `PIPELINE_ID` | 平台流水线 ID | `12` |
| `RUN_ID` | 平台运行记录 ID | `56` |
| `PLATFORM_CALLBACK_URL` | 平台回调地址 | `http://k8sop-api:8080` |
| `GIT_CREDENTIAL_ID` | Git 凭证 ID | `gitee-id` |
| `SKIP_TESTS` | 是否跳过测试 | `true` / `false` |
| `ENABLE_SONAR` | 是否启用代码扫描 | `true` / `false` |

---

## 六、平台后端配置（config.yaml）

```yaml
Jenkins:
  URL: "http://192.168.1.100:8080/"         # Jenkins 服务器地址（必须以 / 结尾）
  Username: "admin"                          # Jenkins 管理员用户名
  APIToken: "11a3f4b5c6d7e8f9..."           # Jenkins API Token（见下方生成方法）
  TriggerTimeout: 60                         # 触发构建超时时间（秒）
  CallbackURL: "http://192.168.1.200:8080"   # 平台后端 API 地址（Jenkins 能访问到）
  PlatformURL: "http://192.168.1.200:5173"   # 前端页面地址（用于钉钉通知链接）
  HMACSecret: "a7f9c2e1b4d5..."             # HMAC 签名密钥（与 Jenkins 凭证一致）
  PollInterval: 15                           # 轮询构建状态间隔（秒）
  MaxBuildTime: 30                           # 最大构建时间（分钟，超时判定失败）
  DingTalkWebhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"  # 钉钉通知（可选）
```

### 6.1 生成 Jenkins API Token

1. 登录 Jenkins → 右上角用户名 → **Configure**
2. **API Token** 区域 → **Add new Token**
3. 输入名称（如 `k8s-platform`）→ **Generate**
4. **复制生成的 Token**（只显示一次！）
5. 填入 `config.yaml` 的 `APIToken` 字段

### 6.2 配置字段详解

| 字段 | 必填 | 说明 |
|------|------|------|
| `URL` | ✅ | Jenkins 地址，确保平台后端能访问 |
| `Username` | ✅ | Jenkins 用户（需有 Job 的构建、配置权限） |
| `APIToken` | ✅ | Jenkins API Token（非登录密码） |
| `TriggerTimeout` | ❌ | 触发构建后等待进入队列的超时（默认 60s） |
| `CallbackURL` | ✅ | 平台后端地址，Jenkins 构建完成后回调此地址 |
| `PlatformURL` | ❌ | 前端地址，用于通知消息中的链接跳转 |
| `HMACSecret` | ✅ | 回调验签密钥，必须与 Jenkins 凭证 `hmac-secret` 一致 |
| `PollInterval` | ❌ | 轮询间隔（秒），默认 15 |
| `MaxBuildTime` | ❌ | 超时阈值（分钟），超过则判定失败，默认 30 |
| `DingTalkWebhook` | ❌ | 钉钉机器人 Webhook（构建通知用） |

---

## 七、SonarQube 集成配置（可选）

如果启用代码质量扫描，需额外配置：

### 7.1 Jenkins 端配置

1. **安装插件**：SonarQube Scanner for Jenkins
2. **Manage Jenkins → System → SonarQube servers：**
   - Name: `SonarQube`（必须叫这个名字，模板中 `withSonarQubeEnv('SonarQube')` 引用）
   - Server URL: `http://sonarqube.example.com:9000`
   - Server authentication token: 添加 Secret text 类型凭证（SonarQube Token）
3. **Manage Jenkins → Tools → SonarQube Scanner：**
   - Name: `SonarQube Scanner`
   - 勾选自动安装 或 指定安装路径

### 7.2 SonarQube 端生成 Token

1. 登录 SonarQube → **My Account → Security**
2. Generate Token → 类型选 **Global Analysis Token**
3. 复制 Token → 添加到 Jenkins 凭证

### 7.3 安装 sonar-scanner（Jenkins 服务器）

```bash
# 下载安装
cd /opt
wget https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-5.0.1.3006-linux.zip
unzip sonar-scanner-cli-5.0.1.3006-linux.zip
ln -s /opt/sonar-scanner-5.0.1.3006-linux/bin/sonar-scanner /usr/local/bin/sonar-scanner
sonar-scanner --version
```

---

## 八、回调机制说明

### 8.1 回调流程

```
Jenkins Pipeline 执行中
    │
    ├── 每个 Stage 完成 → 回调 /api/v1/cicd/pipelines/callback/stage
    │   Body: { pipeline_id, run_id, stage, status, message, ... }
    │
    └── 整体完成 → 回调 /api/v1/cicd/pipelines/callback/pipeline
        Body: { pipeline_id, run_id, status, image, duration, ... }
```

### 8.2 签名验证

回调请求携带 `X-Signature` Header：
```
X-Signature = HMAC-SHA256(hmac_secret, "{JOB_NAME}:{BUILD_NUMBER}:{stage_name}")
```

平台后端验证签名，防止伪造回调。

### 8.3 网络要求

| 方向 | 源 | 目标 | 端口 | 用途 |
|------|-----|------|------|------|
| 平台 → Jenkins | 平台后端 | Jenkins | 8080 | 触发构建/查询状态 |
| Jenkins → 平台 | Jenkins | 平台后端 | 8080 | 构建回调通知 |
| Jenkins → Git | Jenkins | Gitee/GitHub | 443 | 拉取代码 |
| Jenkins → 镜像仓库 | Jenkins | Harbor/ACR | 443/80 | 推送镜像 |
| Jenkins → SonarQube | Jenkins | SonarQube | 9000 | 代码扫描（可选） |

---

## 九、验证清单

### 9.1 一键验证脚本

```bash
# 设置环境变量
export JENKINS_URL="http://192.168.1.100:8080"
export JENKINS_USERNAME="admin"
export JENKINS_API_TOKEN="your-api-token"

# 验证 Jenkins 连接
curl -s -u "${JENKINS_USERNAME}:${JENKINS_API_TOKEN}" "${JENKINS_URL}/api/json" | head -c 200

# 验证 CSRF Crumb 获取
curl -s -u "${JENKINS_USERNAME}:${JENKINS_API_TOKEN}" "${JENKINS_URL}/crumbIssuer/api/json"
```

### 9.2 检查项清单

| # | 检查项 | 验证方法 |
|---|--------|---------|
| 1 | Jenkins 可访问 | 浏览器打开 Jenkins URL |
| 2 | API Token 有效 | `curl -u user:token /api/json` 返回 200 |
| 3 | Git 凭证可用 | 手动触发 Job，观察代码拉取是否成功 |
| 4 | Harbor 凭证可用 | 手动触发 Job，观察镜像推送是否成功 |
| 5 | HMAC 密钥一致 | 检查 Jenkins `hmac-secret` 凭证值 = config.yaml `HMACSecret` |
| 6 | 网络互通 | 平台能访问 Jenkins，Jenkins 能回调平台 |
| 7 | Docker 可用 | Jenkins 用户执行 `docker info` 成功 |
| 8 | Go/Maven/Node 已装 | 对应命令 `go version` / `mvn -v` / `node -v` 正常 |

---

## 十、常见问题

### Q1: 平台触发构建报 "认证失败"

**原因**：API Token 错误或已过期
**解决**：
1. 确认 `config.yaml` 中 `Jenkins.Username` 和 `APIToken` 正确
2. 重新生成 API Token（Jenkins 用户配置页面）
3. 确认用户有 Job 的 Build 权限

### Q2: 代码拉取失败 "Authentication failed"

**原因**：Git 凭证配置错误
**解决**：
1. 确认 Jenkins 凭证 ID 为 `gitee-id`（或流水线指定的 ID）
2. Gitee 用户使用 **Personal Access Token** 而非密码
3. Token 需有 `projects` 权限

### Q3: 镜像推送失败 "unauthorized"

**原因**：Harbor/ACR 凭证配置错误
**解决**：
1. 确认 Jenkins 凭证 ID 为 `harbor-registry`
2. 在 Jenkins 服务器手动测试：`docker login harbor.example.com`
3. ACR 需在「容器镜像服务 → 访问凭证」设置固定密码

### Q4: 回调不生效，平台状态一直"构建中"

**原因**：Jenkins 无法访问平台回调地址 或 HMAC 签名不一致
**解决**：
1. 从 Jenkins 服务器执行：`curl http://平台地址:8080/healthz/live`
2. 确认 `config.yaml` 中 `CallbackURL` 是 Jenkins 能访问的地址
3. 检查 HMAC 密钥是否两端一致
4. 如果回调不通，平台会通过轮询机制（`PollInterval`）兜底获取状态

### Q5: SonarQube 扫描失败

**原因**：SonarQube 服务不可达 或 Token 无效
**解决**：
1. 确认 Jenkins System 配置中 SonarQube 服务器名称为 `SonarQube`
2. 从 Jenkins 测试：`curl http://sonarqube:9000/api/system/status`
3. 重新生成 SonarQube Token

### Q6: Pipeline 报 "Scripts not permitted to use method"

**原因**：Groovy 沙箱安全限制
**解决**：
- **Manage Jenkins → In-process Script Approval** → 批准被阻断的方法签名
- 或在 Job 配置取消勾选 "Use Groovy Sandbox"（不推荐生产环境）

---

## 附录：配置速查表

### 需要在 Jenkins 上配置的内容

```
┌─────────────────────────────────────────────────────────┐
│                    Jenkins 配置总览                       │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  📦 插件 (Manage Jenkins → Plugins)                     │
│    ├── Pipeline                                         │
│    ├── Git                                              │
│    ├── Credentials Binding                              │
│    ├── HTTP Request                                     │
│    ├── Docker Pipeline                                  │
│    ├── Pipeline Utility Steps                           │
│    └── SonarQube Scanner (可选)                         │
│                                                         │
│  🔑 凭证 (Manage Jenkins → Credentials)                 │
│    ├── gitee-id          (Username/Password) → Git 拉取  │
│    ├── harbor-registry   (Username/Password) → 镜像推送  │
│    └── hmac-secret       (Secret Text)       → 回调签名  │
│                                                         │
│  🔧 工具 (Manage Jenkins → Tools)                       │
│    ├── Maven-3.9         → /opt/apache-maven-3.9.9      │
│    └── JDK-21            → /usr/lib/jvm/java-21         │
│                                                         │
│  🖥️ 服务器环境                                          │
│    ├── Go 1.24+          → /usr/local/go                │
│    ├── Node.js 18+       → /usr/bin/node                │
│    ├── Python 3.11+      → /usr/bin/python3             │
│    ├── Docker            → jenkins 用户在 docker 组      │
│    └── sonar-scanner     → /usr/local/bin (可选)        │
│                                                         │
│  📋 Job 配置                                            │
│    └── k8s-platform-builder (Pipeline from SCM)         │
│        ├── SCM: Git → 平台仓库地址                       │
│        ├── Credentials: gitee-id                        │
│        └── Script Path: Jenkinsfile                     │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### 需要在平台 config.yaml 配置的内容

```yaml
Jenkins:
  URL: "http://<Jenkins地址>:8080/"
  Username: "<Jenkins用户名>"
  APIToken: "<Jenkins API Token>"
  TriggerTimeout: 60
  CallbackURL: "http://<平台后端地址>:8080"
  PlatformURL: "http://<前端地址>:5173"
  HMACSecret: "<与Jenkins hmac-secret凭证相同的密钥>"
  PollInterval: 15
  MaxBuildTime: 30
  DingTalkWebhook: "<钉钉Webhook地址（可选）>"
```
