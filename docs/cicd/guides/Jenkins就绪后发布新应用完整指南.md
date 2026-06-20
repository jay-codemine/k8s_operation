# Jenkins 就绪后发布新应用完整指南

> 本文档回答：Jenkins 部署、凭证、插件都配置完成后，下一步该做什么？
> 核心结论：**接下来就是在平台注册一个新应用（流水线），触发第一次构建，然后发布到 K8s。**
> 整个过程约 10~30 分钟，Jenkins 端后续无需再改动。

---

## 一、整体流程概览

```
Jenkins 就绪
    │
    ▼
[1] 确认环境 checklist（平台 / Jenkins / 业务仓库 / K8s 集群）
    │
    ▼
[2] 首次创建 K8s 工作负载（Deployment + Service）  ← 新应用必做
    │
    ▼
[3] 平台创建应用/流水线（CI/CD → 流水线 → 新建应用）
    │
    ▼
[4] 触发构建（git push 或手动点击「立即构建」）
    │
    ▼
[5] 创建发布单 或 等待自动部署（CI/CD 只负责更新镜像）
    │
    ▼
[6] 验证 K8s 部署结果
```

---

## 二、环境确认清单（发布前必做）

### 2.1 Jenkins 端（一次性，已配置完成需核对）

| 检查项 | 要求 | 排查位置 |
|--------|------|----------|
| 4 个通用 Builder Job 已创建 | `k8s-builder-go`、`k8s-builder-java`、`k8s-builder-frontend`、`k8s-builder-python` | Jenkins 首页 |
| 3 个全局凭证已配置 | `gitee-id`、`harbor-registry`、`hmac-secret` | Manage Jenkins → Credentials |
| 插件已安装 | Pipeline、Git、HTTP Request，Java 项目还需 SonarQube Scanner | Manage Jenkins → Plugins |
| 网络连通 | Jenkins 能访问 Git 仓库、镜像仓库、平台回调地址 | Jenkins 容器/节点内 `curl` 测试 |
| 构建工具（按需） | Java 需 JDK + Maven；前端需 Node.js；Go/Python 需对应运行时 | Jenkins 服务器 |

> 详细 Jenkins 初始化步骤见 [CICD发布操作手册.md](CICD发布操作手册.md) 第 12 节。

### 2.2 平台端（config.yaml）

确保 `configs/config.yaml` 中 Jenkins 配置块正确：

```yaml
Jenkins:
  URL: "http://jenkins.devops.svc.cluster.local:8080/"  # Jenkins 服务地址
  Username: "admin"
  APIToken: "xxxxxxxx"                                    # Jenkins API Token，不是登录密码
  TriggerTimeout: 60
  CallbackURL: "http://k8soperation-backend.k8soperation.svc.cluster.local:38180"
  PlatformURL: "http://k8soperation-frontend.k8soperation.svc.cluster.local"
  HMACSecret: "xxxxx3VbN6dHe1Zx4JuWq"                     # 必须与 Jenkins 中 hmac-secret 凭证一致
  GitCredentialID: "gitee-id"
  RegistryCredentialID: "harbor-registry"
  HMACCredentialID: "hmac-secret"
  PollInterval: 15
  MaxBuildTime: 30
```

> 完整示例参考 [configs/config.yaml.example](../../../../configs/config.yaml.example)。

### 2.3 业务项目端

每个需要发布的新项目，只需要有**构建配置文件**即可：

```
my-app/
├── pom.xml / go.mod / package.json / requirements.txt  # 对应语言的构建配置
├── Dockerfile（可选，不放也行，模板会自动生成）
└── src/...
```

> - 业务项目**不需要** Jenkinsfile，模板由平台统一维护，参见 [Jenkinsfile](../../../../Jenkinsfile)。
> - 业务项目**不需要** Dockerfile，Pipeline 模板在构建时会自动生成生产级 Dockerfile（含时区、健康检查、非 root 用户等最佳实践）。
> - 如果项目根目录**有** Dockerfile，模板会优先使用项目自带的（自定义需求场景）。

### 2.4 K8s 集群端

- 目标集群已在平台「集群管理」中接入。
- 平台后端服务账号有权限操作目标命名空间的工作负载。

> **重要：CI/CD 模块只负责「更新镜像」，不负责「创建工作负载」。**
> 新应用首次上线前，必须先创建 Deployment + Service（详见下方「新应用首次创建工作负载」章节）。

---

## 三、新应用首次创建工作负载（Deployment + Service）

> **核心设计原则**：CI/CD 流水线 = 代码构建 + 镜像更新；基础设施创建 = 独立操作。
> 这是业界主流做法（阿里 EDAS、腾讯 TKE、华为 CCE 均如此），原因：
> - 创建 Deployment 涉及资源限制、探针、Volume、亲和性等复杂配置，不适合在发布时填
> - 职责分离：创建基础设施（低频、高权限） vs 发布代码（高频、低权限）
> - 安全性：运维管基础设施，开发管发布

### 3.0 新应用首次上线的完整路径

```
运维/管理员（一次性）：
  ① 在平台「资源管理 → Deployment → 创建」创建工作负载
     - 填写名称、初始镜像（可用占位符如 nginx:latest）、端口、副本数、资源限制
     - 可选同时创建 Service
  ② 如需对外暴露，创建 Ingress

开发者（日常）：
  ③ 在平台「CI/CD → 创建流水线」→ 关联已有的 Deployment
  ④ git push → 自动构建 → CI/CD 自动 Patch 镜像 → 滚动更新完成
```

### 3.0.1 方式一：通过平台 UI 创建（推荐）

**资源管理 → Deployment → 创建**

平台已提供完整的 Deployment 创建 API（`POST /api/v1/k8s/deployment/create`），支持：
- 设置容器名称、镜像、端口
- 设置资源 Requests / Limits
- 设置副本数
- 可选同时创建 Service（`is_create_service: true`）

首次镜像可使用占位符（如 `nginx:latest` 或 `busybox:latest`），构建成功后 CI/CD 会自动替换为真实镜像。

### 3.0.2 方式二：通过 YAML 创建

**资源管理 → Deployment → 从 YAML 创建**（`POST /api/v1/k8s/deployment/create-from-yaml`）

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  namespace: production
spec:
  replicas: 2
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service          # 容器名称（CI/CD 会用这个名称定位更新镜像）
        image: nginx:latest          # 占位镜像，CI/CD 首次构建后自动替换
        ports:
        - containerPort: 8080
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
---
apiVersion: v1
kind: Service
metadata:
  name: user-service
  namespace: production
spec:
  selector:
    app: user-service
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

### 3.0.3 方式三：kubectl / Helm（运维偏好）

```bash
# 简单场景
kubectl create deployment user-service --image=nginx:latest -n production
kubectl expose deployment user-service --port=8080 -n production

# 或用 Helm Chart
helm install user-service ./charts/web-app -n production \
  --set image.repository=nginx --set image.tag=latest
```

### 3.0.4 CI/CD 流水线创建时如何关联？

创建流水线时，在「自动部署配置」中填写：

| 字段 | 填什么 | 说明 |
|------|--------|------|
| 工作负载名称 | `user-service` | 必须与上面创建的 Deployment 名称**完全一致** |
| 容器名称 | `user-service` | 必须与 Deployment 中的 container name 一致 |
| 命名空间 | `production` | 必须与 Deployment 所在 namespace 一致 |

> CI/CD 构建成功后会自动执行 `StrategicMergePatch`，将镜像从占位符替换为刚构建的真实镜像。

---

## 四、Step 1：业务项目准备（零侵入）

### 3.1 核心结论：无需 Dockerfile

平台的 Pipeline 模板已内置 **Dockerfile 自动生成逻辑**，构建时按以下优先级处理：

| 优先级 | 条件 | 行为 |
|--------|------|------|
| 1 | 平台传入 `DOCKERFILE_PATH` 参数 | 使用指定路径 |
| 2 | 项目根目录存在 `Dockerfile` | 使用项目自带的 |
| 3 | 以上都没有 | ✅ **模板自动生成**（推荐，零侵入） |

### 3.2 自动生成的 Dockerfile 内容

各语言模板自动生成的 Dockerfile 已是生产级水准：

| 语言 | 基础镜像 | 内置能力 |
|------|----------|----------|
| **Go** | `alpine:3.20` | 时区、CA证书、非root用户、健康检查 |
| **Java** | `java:{version}-jre-alpine` | 时区、JVM 调优参数、G1GC、OOM Dump、非root用户 |
| **Frontend** | `nginx:1.25-alpine` | SPA 路由、时区、非root用户、健康检查 |
| **Python** | `python:{version}-slim` | 国内 pip 源、时区、非root用户、健康检查 |

> 自动生成的 Dockerfile 代码详见各语言模板的 `Build & Push Image` 阶段：
> - [go-pipeline.groovy](../../../../configs/jenkins-templates/go-pipeline.groovy)
> - [java-spring-pipeline.groovy](../../../../configs/jenkins-templates/java-spring-pipeline.groovy)
> - [frontend-pipeline.groovy](../../../../configs/jenkins-templates/frontend-pipeline.groovy)
> - [python-pipeline.groovy](../../../../configs/jenkins-templates/python-pipeline.groovy)

### 3.3 业务项目只需要什么？

```
my-app/
├── pom.xml          # Java 项目
├── go.mod           # Go 项目
├── package.json     # 前端项目
├── requirements.txt # Python 项目
└── src/...          # 源码
```

**不需要 Jenkinsfile，不需要 Dockerfile，不需要任何 CI/CD 相关文件。**

### 3.4 什么时候需要自定义 Dockerfile？

以下场景建议在项目根目录放自己的 Dockerfile：

- 需要安装额外系统依赖（如 ffmpeg、ImageMagick）
- 需要自定义 Nginx 配置（前端项目）
- 需要多阶段构建优化（自动生成版已够用，但想进一步定制）
- 需要特殊端口或 ENTRYPOINT 参数

> 更多自定义 Dockerfile 模板参考 [configs/dockerfile-templates](../../../../configs/dockerfile-templates) 与 [docs/dockerfile](../../../../docs/dockerfile)。

---

## 五、Step 2：在平台创建应用/流水线

### 4.1 入口

**导航栏 → CI/CD → 流水线 → 「新建应用」**

或在「发布管理」空状态页点击「**快速新建应用**」。

### 4.2 必填字段（极简模式）

| 字段 | 说明 | 示例 |
|------|------|------|
| **应用名称 / 流水线名称** | 全局唯一，建议 `{项目名}-{环境}` | `user-service-prod` |
| **Git 仓库地址** | 业务项目仓库 HTTPS 或 SSH 地址 | `https://gitee.com/org/user-service.git` |
| **语言 / 框架类型** | 决定使用哪个 Jenkins 模板 | Java / Go / Node.js / Python |

### 4.3 推荐填写字段

| 字段 | 说明 | 示例 |
|------|------|------|
| **镜像仓库地址** | 构建后推送的镜像 Repo | `harbor.example.com/project/user-service` |
| **构建成功后自动部署** | 勾选后构建成功自动发布 | ✅ 开启 |
| **目标集群** | 已接入的 K8s 集群 | cluster-prod-1 |
| **目标命名空间** | 默认 `default`，生产建议 `production` | `production` |
| **工作负载类型** | Deployment / StatefulSet / DaemonSet | `Deployment` |
| **工作负载名称** | 默认等于应用名 | `user-service` |
| **容器名称** | 默认更新第一个容器 | `app` |

### 4.4 平台自动完成的工作

点击「立即创建」后，平台会自动完成：

- ✅ 根据 `language_type` 自动映射 Jenkins Job（如 Java → `k8s-builder-java`）
- ✅ 绑定对应语言模板（如 `configs/jenkins-templates/java-spring-pipeline.groovy`）
- ✅ 注入 `GIT_REPO`、`IMAGE_REPO`、`LANGUAGE_TYPE` 等参数
- ✅ 写入数据库，生成 `pipeline_id`

> 自动映射规则定义在 [internal/app/models/cicd_pipeline.go](../../../../internal/app/models/cicd_pipeline.go) 的 `DefaultJenkinsJobMap`。

---

## 六、Step 3：触发构建

### 5.1 方式一：自动触发（推荐日常）

```bash
git push origin main
```

Jenkins 收到 Webhook 后自动开始：

```
清理工作空间 → 检出代码 → 依赖下载 → 编译 → 测试 → 镜像构建 → 推送镜像 → 回调平台
```

### 5.2 方式二：手动触发

**CI/CD → 流水线 → 找到对应应用 → 点击「立即构建」**

可临时覆盖：
- 分支（如 hotfix 分支）
- 环境变量（如 `SKIP_TESTS=true`、`ENABLE_SONAR=false`）

### 5.3 构建阶段说明

| 阶段 | 说明 | 默认启用 |
|------|------|----------|
| 清理工作空间 | 清理 Jenkins workspace | ✅ |
| 代码检出 | 拉取 Git 指定分支 | ✅ |
| 依赖下载 | 下载语言依赖 | ✅ |
| 编译检查 | 编译验证 | ✅ |
| 单元测试 | 执行测试 | ✅ |
| 代码检查 | lint / format | ✅ |
| SonarQube 扫描 | 代码质量扫描 | ❌ 按需 |
| 构建镜像 | `nerdctl build` | ✅ |
| 推送镜像 | `nerdctl push` 到 Harbor | ✅ |
| 回调平台 | 携带镜像地址与状态 | ✅ |

> 各语言具体模板见 [configs/jenkins-templates](../../../../configs/jenkins-templates)。

---

## 七、Step 4：发布到 K8s

### 6.1 如果你勾选了「自动部署」

构建成功后，平台会自动：

1. 接收 Jenkins 回调（HMAC 验签）
2. 使用 `client-go StrategicMergePatch` 更新 Deployment 镜像
3. 轮询等待 Rollout 完成
4. 写入发布记录并发送通知

**你只需在「发布管理」页查看结果。**

### 6.2 如果你未勾选「自动部署」

需要手动创建发布单：

**导航栏 → CI/CD → 发布管理 → 「创建发布」**

| 字段 | 说明 | 示例 |
|------|------|------|
| 选择应用 | 下拉选择刚才创建的应用 | `user-service` |
| 版本号 / 镜像标签 | 要发布的镜像 Tag | `v1.0.0` |
| 备注 | 发布说明 | 首次上线 |

> 选择应用后，命名空间、工作负载、镜像仓库自动继承，只需填版本号。

### 6.3 极速发布模式（Quick Deploy）

如果镜像已经存在于 Harbor，无需重新构建：

**CI/CD → 发布管理 → 创建发布单 → 填入镜像地址/Tag → 提交**

- 不需要 Jenkins
- 几秒即可完成 K8s 部署
- 适合紧急热修、回滚、多集群同步

---

## 八、Step 5：验证部署结果

### 7.1 查看发布单状态

| 状态 | 含义 | 可操作 |
|------|------|--------|
| 🟡 部署中 | 正在执行 | 可取消 |
| 🟢 成功 | 部署完成 | 可回滚 / 重新部署 |
| 🔴 失败 | 部署失败 | 可重试 / 查看日志 |
| 🟣 已回滚 | 已回滚到上一版本 | 可重新部署 |

### 7.2 查看 K8s 资源

进入平台「资源管理 → Deployment」页面，确认：

- Pod 全部 Running
- 镜像已更新为目标 Tag
- ReplicaSet 滚动更新完成

也可在服务器执行：

```bash
kubectl get deployment -n <namespace>
kubectl rollout status deployment/<workload-name> -n <namespace>
kubectl get pods -n <namespace> -l app=<workload-name>
```

### 7.3 验证业务接口

```bash
# 通过 Service/Port-forward 验证
curl http://<service-name>.<namespace>.svc.cluster.local:8080/health
```

---

## 九、完整发布 vs 极速发布如何选择？

| 场景 | 推荐方式 | 是否需要 Jenkins | 耗时 |
|------|----------|------------------|------|
| 代码变更后日常发布 | 完整发布（运行流水线） | ✅ | 3~10 分钟 |
| 首次上线 | 完整发布 | ✅ | 3~10 分钟 |
| 镜像已构建好，只需部署 | 极速发布 | ❌ | 几秒 |
| 紧急热修 | 极速发布 | ❌ | 几秒 |
| 回滚到历史版本 | 极速发布 / 重新部署 | ❌ | 几秒 |
| 多集群同步部署 | 极速发布 | ❌ | 按并发数 |

---

## 十、首次发布常见问题排查

### 9.1 构建失败

| 现象 | 可能原因 | 解决方案 |
|------|----------|----------|
| 代码检出失败 | Git 凭证无效 | 检查 Jenkins `gitee-id` 凭证 |
| 依赖下载超时 | 网络/代理问题 | 检查 Jenkins 网络、Maven/Go 镜像源 |
| 编译失败 | 代码错误 | 查看阶段日志 |
| 镜像推送失败 | Harbor 凭证错误 | 检查 Jenkins `harbor-registry` 凭证 |
| 回调失败 | 平台不可达或 HMAC 不一致 | 检查 `CallbackURL` 和 `HMACSecret` |

### 9.2 部署失败

| 现象 | 可能原因 | 解决方案 |
|------|----------|----------|
| Pod 镜像拉取失败 | 镜像地址错误或 Harbor 认证失败 | 检查镜像 Repo/Tag、ImagePullSecret |
| Pod 启动失败 | 应用启动错误 | 查看 Pod 日志 |
| Rollout 超时 | 健康检查未通过 | 检查 Dockerfile EXPOSE、应用健康接口 |
| 无权限 Patch | 平台服务账号 RBAC 不足 | 检查集群权限 |

### 9.3 流水线卡住「运行中」

1. 检查 Jenkins 是否正常
2. 检查平台后端是否收到回调
3. 使用「强制运行」重置状态
4. 或手动停止后重新触发

---

## 十一、日常发布（第二次及以后）

```
1. git push 代码变更
2. 等待 Jenkins 自动构建成功
3. 如未开启自动部署，进入「发布管理 → 创建发布单 → 选应用 + 填版本号 → 确认」
```

全程不超过 30 秒。

---

## 十二、相关文件索引

| 文件 | 说明 |
|------|------|
| [configs/config.yaml.example](../../../../configs/config.yaml.example) | 平台 Jenkins 配置示例 |
| [Jenkinsfile](../../../../Jenkinsfile) | Pipeline 模板分发器 |
| [configs/jenkins-templates/java-spring-pipeline.groovy](../../../../configs/jenkins-templates/java-spring-pipeline.groovy) | Java 构建模板 |
| [configs/jenkins-templates/go-pipeline.groovy](../../../../configs/jenkins-templates/go-pipeline.groovy) | Go 构建模板 |
| [configs/jenkins-templates/frontend-pipeline.groovy](../../../../configs/jenkins-templates/frontend-pipeline.groovy) | 前端构建模板 |
| [configs/jenkins-templates/python-pipeline.groovy](../../../../configs/jenkins-templates/python-pipeline.groovy) | Python 构建模板 |
| [configs/dockerfile-templates](../../../../configs/dockerfile-templates) | Dockerfile 模板 |
| [scripts/batch-import-pipelines.json](../../../../scripts/batch-import-pipelines.json) | 批量导入流水线模板 |
| [internal/app/models/cicd_pipeline.go](../../../../internal/app/models/cicd_pipeline.go) | 语言类型与 Jenkins Job 映射 |

---

## 十三、Checklist：从 Jenkins 就绪到首次发布成功

- [ ] Jenkins 中 4 个通用 Builder Job 已创建并保存
- [ ] Jenkins 中 3 个凭证（gitee-id / harbor-registry / hmac-secret）已配置
- [ ] 平台 `config.yaml` 中 Jenkins 配置正确且已重启生效
- [ ] 目标 K8s 集群已在平台接入
- [ ] **新应用已创建 Deployment + Service**（平台 UI / YAML / kubectl）
- [ ] 业务项目有构建配置文件（pom.xml/go.mod/package.json/requirements.txt），Dockerfile 可选
- [ ] 已在平台创建应用/流水线，语言类型选择正确，工作负载名称匹配
- [ ] 已触发首次构建并构建成功
- [ ] 已自动部署或手动创建发布单并部署成功
- [ ] 已验证 K8s Pod 镜像更新、业务接口正常

---

> 完成以上步骤后，你的第一个应用就已经成功发布到 K8s。
> 后续新增应用只需重复：**创建 Deployment → 平台创建流水线 → 触发构建** 即可，Jenkins 端完全零改动。
