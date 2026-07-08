# Jenkins 完整配置操作手册

> 适用版本：k8s_operation v13.3+  
> Jenkins 版本：LTS（运行于 K8s devops 命名空间）  
> 访问地址：`http://<节点IP>:30080`  
> 管理员账号：`ops-dev` / `admin123`

---

## 目录

1. [架构说明](#1-架构说明)
2. [前置准备：外部平台 Token 生成](#2-前置准备外部平台-token-生成)
3. [插件清单（已自动安装）](#3-插件清单已自动安装)
4. [Jenkins UI 操作步骤](#4-jenkins-ui-操作步骤)
   - 4.1 [创建 Git 凭证](#41-创建-git-凭证-gitee-id)
   - 4.2 [创建镜像仓库凭证](#42-创建镜像仓库凭证-harbor-registry)
   - 4.3 [创建 4 个 Pipeline Job](#43-创建-4-个-pipeline-job)
   - 4.4 [生成 Jenkins API Token](#44-生成-jenkins-api-token)
5. [后端配置更新](#5-后端配置更新)
6. [凭证 ID 自定义说明](#6-凭证-id-自定义说明)
7. [完成验证清单](#7-完成验证清单)
8. [常见问题](#8-常见问题)

---

## 1. 架构说明

```
平台后端 ──触发构建──▶ Jenkins Master (K8s Pod)
                              │
                    动态创建构建 Pod
                              │
                    ┌─────────┼──────────┐
                    ▼         ▼          ▼
                Git Clone  Build Image  Push Image
               (gitee-id) (Kaniko)   (harbor-registry)
                              │
                    构建完成回调平台接口
                  (hmac-secret 签名验证)
```

**关键说明：**

- Jenkins Master 以 **StatefulSet** 运行在 `devops` 命名空间
- 每次构建动态创建一个 **Pod Agent**，构建完自动销毁（无常驻 Agent）
- 插件存储在 PVC `jenkins-data`，重建 Pod 不会丢失
- 通过 JCasC（Configuration as Code）自动初始化 K8s Cloud 配置和 `hmac-secret` 凭证

---

## 2. 前置准备：外部平台 Token 生成

在操作 Jenkins 之前，需要先在外部平台生成认证 Token。

---

### 2.1 Gitee 生成 Access Token

**用途**：Jenkins 用此 Token 克隆任意私有/公开代码仓库。

**操作步骤：**

```
1. 登录 Gitee（https://gitee.com）
2. 右上角头像 → 设置
3. 左侧菜单 → 安全设置 → 私人令牌
4. 点击「生成新令牌」
5. 填写令牌描述，例如：jenkins-cicd
6. 权限勾选：✅ projects（仓库读取权限）
7. 点击「提交」
8. ⚠️ 复制并保存 Token（页面关闭后无法再次查看！）
```

> **注意**：使用 Token 时，用户名填你的 Gitee 账号，密码填此 Token。

---

### 2.2 阿里云容器镜像服务 ACR 生成凭证

**用途**：Jenkins 用此凭证将构建好的 Docker 镜像推送到阿里云镜像仓库。

**操作步骤：**

```
1. 登录阿里云控制台（https://cr.console.aliyun.com）
2. 选择「个人版」或「企业版」实例
3. 左侧菜单 → 访问凭证
4. 设置固定密码（首次需要设置 Registry 登录密码）
5. 记录以下信息：
   - 仓库地址（Registry 地址）：如 registry.cn-hangzhou.aliyuncs.com
   - 用户名：阿里云账号全名（如 user@example.com）或子账号
   - 密码：刚才设置的固定密码
```

> **镜像仓库地址**根据地区不同而不同，常见地区：
> - 华东（杭州）：`registry.cn-hangzhou.aliyuncs.com`
> - 华北（北京）：`registry.cn-beijing.aliyuncs.com`
> - 华南（深圳）：`registry.cn-shenzhen.aliyuncs.com`

---

## 3. 插件清单（已自动安装）

以下 13 个插件已由 initContainer 在 Jenkins 启动时自动安装到 PVC，无需手动操作。

| # | 插件 ID | 版本 | 用途 |
|---|---------|------|------|
| 1 | `kubernetes` | v4423 | K8s 动态 Pod Agent，每次构建独立 Pod |
| 2 | `workflow-aggregator` | v608 | Pipeline（流水线）核心语法支持 |
| 3 | `git` | v5.10.1 | 克隆 Git 代码仓库 |
| 4 | `configuration-as-code` | v2088 | JCasC 自动初始化 Jenkins 配置 |
| 5 | `credentials-binding` | v725 | 将凭证安全注入 Pipeline 环境变量 |
| 6 | `http_request` | v1.25 | 构建完成后回调平台 REST 接口 |
| 7 | `pipeline-utility-steps` | v3.810 | `readJSON` 等工具函数 |
| 8 | `junit` | v1413 | Java 单元测试报告展示 |
| 9 | `sonar` | v2.18.3 | SonarQube 代码质量扫描（可选功能） |
| 10 | `pipeline-stage-view` | v2.41 | 流水线阶段可视化视图 |
| 11 | `timestamper` | v1.30 | 控制台日志添加时间戳 |
| 12 | `ws-cleanup` | v0.49 | 构建完成自动清理工作空间 |
| 13 | `ansicolor` | v536 | 支持彩色 ANSI 控制台日志输出 |

**验证命令（可选）：**

```bash
kubectl -n devops exec jenkins-0 -- ls /var/jenkins_home/plugins/ | grep -v '\.jpi$' | sort
```

---

## 4. Jenkins UI 操作步骤

打开浏览器，访问：`http://<节点IP>:30080`

登录信息：
- 用户名：`ops-dev`
- 密码：`admin123`

---

### 4.1 创建 Git 凭证（`gitee-id`）

```
导航路径：
Manage Jenkins → Credentials → System → Global credentials (unrestricted)
→ 右侧「Add Credentials」按钮

填写内容：
  Kind（类型）：    Username with password
  Scope（范围）：   Global
  Username：       你的 Gitee 用户名（如：zhangsan）
  Password：       第 2.1 步生成的 Gitee Access Token
  ID：             gitee-id               ← 必须与此完全一致
  Description：    Gitee repository credentials

→ 点击「Create」保存
```

---

### 4.2 创建镜像仓库凭证（`harbor-registry`）

```
同一页面继续 → 「Add Credentials」

填写内容：
  Kind（类型）：    Username with password
  Scope（范围）：   Global
  Username：       阿里云 ACR 用户名
  Password：       阿里云 ACR 密码（第 2.2 步设置的固定密码）
  ID：             harbor-registry        ← 必须与此完全一致
  Description：    Aliyun ACR registry credentials

→ 点击「Create」保存
```

> `hmac-secret` 和 `sonarqube-token` 凭证已由 JCasC 自动创建，**无需手动添加**。

---

### 4.3 创建 4 个 Pipeline Job

平台支持 4 种语言的 CI/CD，每种语言对应一个 Jenkins Job。

**创建步骤（重复 4 次）：**

```
Dashboard → New Item（新建任务）
→ 输入 Job 名称（见下表）
→ 选择类型：Pipeline
→ 点击「OK」

进入配置页：
→ 找到「Pipeline」区域
→ Definition：Pipeline script
→ Script：粘贴对应模板文件的完整内容

→ 点击「Save」保存
```

**4 个 Job 对应关系：**

| Job 名称 | 模板文件路径 | 适用语言 | 说明 |
|---------|------------|---------|------|
| `go-pipeline` | `configs/jenkins-templates/go-pipeline.groovy` | Go | 支持自动生成 Dockerfile |
| `java-spring-pipeline` | `configs/jenkins-templates/java-spring-pipeline.groovy` | Java/Spring Boot | 支持多版本 JDK（8/11/17）|
| `frontend-pipeline` | `configs/jenkins-templates/frontend-pipeline.groovy` | Vue/React | 支持 Node.js 构建 |
| `python-pipeline` | `configs/jenkins-templates/python-pipeline.groovy` | Python | 支持 Flask/FastAPI |

**如何查看模板文件内容：**

```bash
# 查看 Go 模板
cat configs/jenkins-templates/go-pipeline.groovy

# 查看 Java 模板
cat configs/jenkins-templates/java-spring-pipeline.groovy
```

---

### 4.4 生成 Jenkins API Token

**用途**：后端平台调用 Jenkins REST API 触发构建时使用。

```
操作路径：
右上角点击用户名「ops-dev」→ Configure（配置）
→ API Token 区域
→ 点击「Add new Token」
→ Token Name 填：platform-token
→ 点击「Generate」
→ ⚠️ 立即复制 Token（页面刷新后不再显示！）
   示例格式：11abc123def456ghi789jkl000mno111
```

---

## 5. 后端配置更新

将第 4.4 步生成的 Jenkins API Token 填入后端配置文件。

**编辑文件：** `deploy/backend/secret.yaml`

找到以下配置项，替换为实际值：

```yaml
# Jenkins 配置
JENKINS_URL: "http://jenkins.devops.svc.cluster.local:8080/"   # 不需要修改（K8s 内部地址）
JENKINS_USERNAME: "ops-dev"                                      # 不需要修改
JENKINS_API_TOKEN: "← 填入第4.4步生成的 Token"                  # ← 必须替换！

# Jenkins 凭证 ID（与 Jenkins UI 中创建的凭证 ID 保持一致）
GIT_CREDENTIAL_ID: "gitee-id"          # 与 4.1 步的 ID 一致
REGISTRY_CREDENTIAL_ID: "harbor-registry"  # 与 4.2 步的 ID 一致
HMAC_CREDENTIAL_ID: "hmac-secret"      # 自动配置，无需修改
```

**应用配置并重启后端：**

```bash
# 应用 Secret
kubectl apply -f deploy/backend/secret.yaml

# 重启后端 Pod 使配置生效
kubectl -n k8soperation rollout restart deployment/k8soperation

# 等待重启完成
kubectl -n k8soperation rollout status deployment/k8soperation
```

---

## 6. 凭证 ID 自定义说明

如果你在 Jenkins UI 中使用了不同的凭证 ID（不是 `gitee-id` / `harbor-registry`），需要同步修改 `deploy/backend/secret.yaml`：

```yaml
GIT_CREDENTIAL_ID: "你在Jenkins中实际创建的Git凭证ID"
REGISTRY_CREDENTIAL_ID: "你在Jenkins中实际创建的镜像仓库凭证ID"
```

**三处必须保持一致：**

```
Jenkins UI 凭证 ID
      ↕ 必须相同
secret.yaml 中的值
      ↕ 后端启动时读取
Pipeline 参数 GIT_CREDENTIAL_ID 默认值
```

---

## 7. 完成验证清单

完成所有步骤后，逐项检查：

| # | 检查项 | 验证方法 | 预期结果 |
|---|--------|---------|---------|
| 1 | Jenkins 正常运行 | 访问 `http://<节点IP>:30080` | 显示登录页，可正常登录 |
| 2 | 13 个插件已安装 | Manage Jenkins → Plugins → Installed | 可搜索到所有插件 |
| 3 | `gitee-id` 凭证存在 | Credentials → System → Global | 可见该凭证条目 |
| 4 | `harbor-registry` 凭证存在 | Credentials → System → Global | 可见该凭证条目 |
| 5 | `hmac-secret` 凭证存在 | Credentials → System → Global | 可见该凭证条目（自动创建）|
| 6 | 4 个 Pipeline Job 已创建 | Dashboard 首页 | 可见 4 个 Job |
| 7 | 后端能调用 Jenkins | 在平台发起一次构建 | Jenkins 收到构建请求 |

---

## 8. 常见问题

### Q1：Jenkins Pod 一直处于 `Init:0/1` 状态

**原因**：initContainer 尝试从 `updates.jenkins.io` 下载插件，国内网络无法访问。

**解决**：当前已将 initContainer 改为跳过模式，插件从 PVC 历史数据加载。  
若 PVC 是全新创建（无历史数据），需要手动在 Jenkins UI → Manage Jenkins → Plugins 安装插件。

---

### Q2：Pod 报 `CreateContainerConfigError`

**原因**：K8s Secret `jenkins-secret` 缺少某个被引用的 key。

**排查命令：**
```bash
kubectl -n devops describe pod jenkins-0 | grep "couldn't find key"
```

**解决：**
```bash
kubectl apply -f deploy/jenkins/secret.yaml
kubectl -n devops delete pod jenkins-0
```

---

### Q3：Pipeline 构建报 `CredentialsNotFoundException`

**原因**：Pipeline 使用的凭证 ID 在 Jenkins 中不存在。

**解决**：确认 Jenkins UI → Credentials 中存在对应 ID 的凭证，ID 必须完全匹配（区分大小写）。

---

### Q4：构建时 `git clone` 失败，提示认证错误

**原因**：`gitee-id` 凭证中的用户名或 Token 填写错误。

**解决**：  
1. Jenkins UI → Credentials → `gitee-id` → Update  
2. 确认用户名为 Gitee 账号，密码为 Access Token（不是登录密码）

---

### Q5：镜像推送失败，提示 `unauthorized`

**原因**：`harbor-registry` 凭证错误，或镜像仓库地址格式不对。

**解决**：  
1. 确认凭证用户名/密码正确  
2. 构建参数 `IMAGE_REPO` 格式应为：`registry.cn-hangzhou.aliyuncs.com/命名空间/镜像名`

---

### Q6：Jenkins API Token 忘记复制

**解决**：重新生成一个 Token（旧 Token 会失效），然后重新更新 `deploy/backend/secret.yaml`。

---

*最后更新：v13.3*
