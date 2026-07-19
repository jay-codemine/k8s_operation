# CI/CD 发布完整配置清单

> 本文档提供从零创建一条流水线到成功发布到 K8s 的完整配置清单，适用于开发团队快速上手。

---

## 一、基础设施准备（一次性配置）

| # | 配置项 | 配置位置 | 说明 |
|---|--------|---------|------|
| 1 | MySQL 数据库 | `config.yaml` → Database | 需导入 `docs/sql/k8s_platform_full_init.sql` 初始化 |
| 2 | Redis | `config.yaml` → Cache | 支持单点/集群，密码可为空 |
| 3 | Jenkins 服务 | `config.yaml` → Jenkins | URL + Username + APIToken |
| 4 | 镜像仓库（Harbor） | Jenkins 凭证 | 凭证 ID: `harbor-registry` |
| 5 | K8s 集群接入 | 平台页面「集群管理」 | 上传 kubeconfig 文件 |
| 6 | Git 凭证 | Jenkins 凭证 | 凭证 ID: `gitee-id` |
| 7 | HMAC 签名密钥 | Jenkins 凭证 + config.yaml | 凭证 ID: `hmac-secret`（双方一致） |

---

## 二、Jenkins Job 创建（按语言类型，一次性）

在 Jenkins 中创建以下 Pipeline Job：

| 语言类型 | Job 名称 | Script Path |
|---------|---------|-------------|
| Go | `go-pipeline` | `configs/jenkins-templates/go-pipeline.groovy` |
| Java | `java-spring-pipeline` | `configs/jenkins-templates/java-spring-pipeline.groovy` |
| 前端 | `frontend-pipeline` | `configs/jenkins-templates/frontend-pipeline.groovy` |
| Python | `python-pipeline` | `configs/jenkins-templates/python-pipeline.groovy` |

**创建步骤：**
1. Jenkins → New Item → Pipeline
2. 命名为对应 Job 名称（如 `go-pipeline`）
3. Pipeline → Definition: **Pipeline script from SCM**
4. SCM: Git → Repository URL: 平台仓库地址
5. Script Path: 填入对应路径

---

## 三、创建流水线配置

### 3.1 必填字段

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `name` | string | 流水线名称（全局唯一） | `user-service-prod` |
| `git_repo` | string | Git 仓库 HTTPS/SSH 地址 | `https://gitee.com/org/user-service.git` |
| `git_branch` | string | 默认构建分支 | `main` |
| `language_type` | string | 语言类型，决定使用哪个构建模板 | `go` / `java` / `frontend` / `python` / `custom` |

### 3.2 自动部署配置（推荐填写）

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `auto_deploy` | bool | 构建成功后是否自动部署到 K8s | `true` |
| `target_cluster_id` | int64 | 目标集群 ID（平台中已接入的集群） | `1` |
| `target_namespace` | string | 目标命名空间 | `production` |
| `target_workload_kind` | string | 工作负载类型 | `Deployment` / `StatefulSet` / `DaemonSet` |
| `target_workload_name` | string | 工作负载名称 | `user-service` |
| `target_container` | string | 容器名称（留空则更新第一个容器） | `app` |
| `deploy_env` | string | 部署环境标识 | `dev` / `test` / `staging` / `prod` |

### 3.3 高级配置（可选）

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `require_approval` | bool | 部署前是否需要人工审批 | `false` |
| `enable_sonar` | bool | 启用 SonarQube 代码质量扫描 | `false` |
| `enable_artifact_upload` | bool | 构建完成后上传制品到制品库 | `false` |
| `enable_deploy_silence` | bool | 发布期间自动静默告警 | `false` |
| `silence_buffer_minutes` | int | 部署完成后延长静默时间（分钟） | `10` |
| `silence_severities` | string | 静默的告警级别 | `warning,info` |

### 3.4 环境变量配置

#### 平台自动注入的变量（无需手动配置）

| 语言类型 | 自动注入 |
|---------|---------|
| **通用** | `GIT_REPO`, `GIT_BRANCH`, `PIPELINE_ID`, `RUN_ID`, `PLATFORM_CALLBACK_URL`, `LANGUAGE_TYPE` |
| **go** | `GO_VERSION=1.24`, `SKIP_TESTS=false` |
| **java** | `JAVA_VERSION=17`, `MAVEN_GOALS=clean package -DskipTests -B` |
| **frontend** | `NODE_VERSION=18`, `BUILD_COMMAND=npm run build`, `BUILD_OUTPUT_DIR=dist` |
| **python** | `PYTHON_VERSION=3.11`, `SKIP_TESTS=false` |

#### 需要手动配置的变量

| 变量名 | 是否必须 | 何时需要 | 示例 |
|--------|---------|---------|------|
| `IMAGE_REPO` | ⚠️ **推荐** | 告诉 Jenkins 镜像推送到哪个仓库 | `harbor.example.com/proj/user-service` |
| `DOCKERFILE_PATH` | ❌ 可选 | 自定义 Dockerfile 路径（默认自动生成） | `./deploy/Dockerfile` |
| `GO_VERSION` | ❌ 可选 | 需要非默认 Go 版本 | `1.22` |
| `NODE_VERSION` | ❌ 可选 | 前端需要指定 Node 版本 | `20` |
| `JAVA_VERSION` | ❌ 可选 | 需要非默认 Java 版本 | `21` |
| `SKIP_TESTS` | ❌ 可选 | 跳过测试 | `true` |
| `GIT_CREDENTIAL_ID` | ❌ 可选 | 非默认 Git 凭证 | `my-git-key` |

---

## 四、运行流水线（触发构建）

| 参数 | 是否必填 | 说明 |
|------|---------|------|
| `id` | ✅ 必填 | 流水线 ID |
| `branch` | ❌ 可选 | 覆盖默认分支（如临时发布 hotfix） |
| `env_vars` | ❌ 可选 | 运行时覆盖环境变量（优先级最高） |
| `force` | ❌ 可选 | 强制运行：清理旧构建后重新触发 |

---

## 五、创建发布单（多集群部署场景）

> 适用于：已有构建好的镜像，需要发布到一个或多个集群

### 5.1 必填字段

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `app_name` | string | 应用名称 | `user-service` |
| `namespace` | string | 目标命名空间 | `production` |
| `workload_kind` | string | 工作负载类型 | `Deployment` |
| `workload_name` | string | 工作负载名称 | `user-service` |
| `container_name` | string | 目标容器名称 | `app` |
| `image_repo` | string | 镜像仓库地址 | `harbor.example.com/proj/user-service` |
| `image_tag` | string | 镜像版本标签 | `v1.2.0` |
| `cluster_ids` | []int64 | 目标集群 ID 列表（支持多集群） | `[1, 2, 3]` |

### 5.2 可选字段

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `strategy` | string | 发布策略 | `rolling` |
| `timeout_sec` | uint32 | 单个任务超时时间（秒） | `300` |
| `concurrency` | uint32 | 并发部署集群数 | `3` |
| `image_digest` | string | 镜像 SHA256 摘要（精确锁版本） | 空 |
| `request_id` | string | 幂等键（防重复提交） | 空 |

### 5.3 发布单状态流转

```
Pending → Queued → Running → Succeeded
                          ↘ Failed → Retry → Running
                                   ↘ Rollback
              ↘ Canceled
```

| 状态 | 说明 |
|------|------|
| `Pending` | 已创建，等待入队 |
| `Queued` | 已入队，等待 Worker 处理 |
| `Running` | 正在执行部署 |
| `Succeeded` | 所有集群部署成功 |
| `Failed` | 部分或全部集群部署失败 |
| `Canceled` | 已取消 |
| `Rollback` | 已回滚 |

---

## 六、完整示例

### 示例 1：Go 项目流水线（自动部署）

```json
{
  "name": "user-service-prod",
  "description": "用户服务生产环境流水线",
  "git_repo": "https://gitee.com/myorg/user-service.git",
  "git_branch": "main",
  "language_type": "go",
  "auto_deploy": true,
  "target_cluster_id": 1,
  "target_namespace": "production",
  "target_workload_kind": "Deployment",
  "target_workload_name": "user-service",
  "target_container": "app",
  "deploy_env": "prod",
  "require_approval": true,
  "enable_sonar": false,
  "env_vars": [
    {"name": "IMAGE_REPO", "value": "harbor.example.com/proj/user-service"}
  ]
}
```

### 示例 2：前端项目流水线

```json
{
  "name": "frontend-web-prod",
  "description": "前端管理系统生产发布",
  "git_repo": "https://gitee.com/myorg/admin-web.git",
  "git_branch": "main",
  "language_type": "frontend",
  "auto_deploy": true,
  "target_cluster_id": 1,
  "target_namespace": "production",
  "target_workload_kind": "Deployment",
  "target_workload_name": "admin-web",
  "target_container": "nginx",
  "deploy_env": "prod",
  "env_vars": [
    {"name": "IMAGE_REPO", "value": "harbor.example.com/proj/admin-web"},
    {"name": "NODE_VERSION", "value": "18"},
    {"name": "BUILD_COMMAND", "value": "npm run build:prod"}
  ]
}
```

### 示例 3：创建发布单（多集群）

```json
{
  "app_name": "user-service",
  "namespace": "production",
  "workload_kind": "Deployment",
  "workload_name": "user-service",
  "container_name": "app",
  "image_repo": "harbor.example.com/proj/user-service",
  "image_tag": "abc1234-20240601120000",
  "cluster_ids": [1, 2, 3],
  "strategy": "rolling",
  "timeout_sec": 300,
  "concurrency": 2
}
```

---

## 七、发布前检查清单（Checklist）

### 基础设施

- [ ] Jenkins 服务可访问，config.yaml 中 Jenkins URL / APIToken 正确
- [ ] Jenkins 中已创建对应语言的 Builder Job（如 `go-pipeline`）
- [ ] Jenkins 凭证已配置：`gitee-id`（Git）、`harbor-registry`（镜像仓库）、`hmac-secret`（签名）
- [ ] 平台已接入目标 K8s 集群（kubeconfig 有效）
- [ ] 回调地址 `CallbackURL` 正确（Jenkins 网络能访问到平台后端）

### 流水线配置

- [ ] 流水线名称唯一，无重复
- [ ] Git 仓库地址正确且凭证有拉取权限
- [ ] 语言类型选择正确（go/java/frontend/python/custom）
- [ ] `IMAGE_REPO` 环境变量已配置（镜像推送目标地址）
- [ ] 镜像仓库有推送权限（Harbor 凭证正确）

### 自动部署配置

- [ ] `auto_deploy` 已开启
- [ ] 目标集群 ID 正确（平台已接入）
- [ ] 目标命名空间存在
- [ ] 目标 Deployment/StatefulSet 已存在于 K8s 中（首次需手动创建）
- [ ] 容器名称匹配（或留空使用第一个容器）

### 发布单配置

- [ ] 镜像仓库地址和 Tag 正确
- [ ] 目标集群 ID 列表正确
- [ ] 工作负载名称在目标命名空间中存在
- [ ] 容器名称正确

---

## 八、常见问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| Jenkins 触发失败 | Job 不存在 | 检查 Jenkins 中是否创建了对应 Job |
| 代码拉取失败 | Git 凭证无效 | 检查 Jenkins 中 `gitee-id` 凭证 |
| 镜像推送失败 | Registry 认证失败 | 检查 `harbor-registry` 凭证 |
| 回调未收到 | 网络不通 | 确认 Jenkins 能访问平台 CallbackURL |
| 部署失败 | 集群连接异常 | 检查 kubeconfig 是否过期 |
| 部署失败 | Deployment 不存在 | 首次部署需先在 K8s 中创建工作负载 |
| 流水线卡在运行中 | Jenkins 已完成但回调丢失 | 使用「强制运行」重置状态 |
| IMAGE_REPO 为空 | 未配置环境变量 | 在流水线环境变量中添加 `IMAGE_REPO` |

---

## 九、一句话速记

> **最小配置 = 4 必填字段 + 1 环境变量 + 5 部署字段**
>
> - 必填：`name` + `git_repo` + `git_branch` + `language_type`
> - 环境变量：`IMAGE_REPO`（镜像推到哪）
> - 部署：`cluster_id` + `namespace` + `workload_kind` + `workload_name` + `container`
