# CICD 发布平台架构设计文档

> 本文档详细介绍平台的 CI/CD 全链路架构设计，包含 14 阶段闭环流水线、SonarQube 代码质量门禁、制品库管理、人工审批机制。

---

## 一、架构概览

### 1.1 设计理念

本平台参考 **Rancher、KubeSphere、Jenkins** 等企业级平台设计，实现了 **CI 与 CD 解耦**：

| 阶段 | 执行方 | 说明 |
|------|--------|------|
| CI（持续集成） | Jenkins | 代码检出、编译、测试、代码扫描、质量门禁、构建制品、上传制品、打包镜像、推送镜像 |
| CD（持续部署） | 平台 | 人工审批、K8s 部署、回滚、通知闭环 |

**核心优势**：
- 构建与部署解耦，职责清晰
- **14 阶段全链路可观测**，每阶段独立状态与日志
- **SonarQube 代码质量门禁**，质量不达标自动阻断发布
- 生产环境强制审批，防止误发布
- **制品库统一管理**，构建产物可追溯、可下载
- 部署支持多集群灵活选择
- 全流程可审计、可追溯

### 1.2 完整 CICD 架构图

```
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                              CI/CD 完整 14 阶段闭环流程                                    │
├──────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                          │
│   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│   │ 1.清理  │─▶│ 2.检出  │─▶│ 3.依赖  │─▶│ 4.编译  │─▶│ 5.测试  │─▶│ 6.检查  │        │
│   │  clean  │  │checkout │  │  deps   │  │compile  │  │  test   │  │  lint   │        │
│   └─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘  └────┬────┘        │
│                                                                          │              │
│   ┌─────────────────────────────────────────────────────────────────────┘              │
│   │                                                                                    │
│   ▼                                                                                    │
│   ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐                              │
│   │ 7.代码   │─▶│ 8.质量   │─▶│ 9.构建   │─▶│10.上传   │                              │
│   │   扫描   │  │   门禁   │  │   制品   │  │  制品库  │                              │
│   │  sonar   │  │  q_gate  │  │build_bin │  │upload_a  │                              │
│   └──────────┘  └──────────┘  └──────────┘  └────┬─────┘                              │
│                                                   │                                    │
│                     Jenkins CI 执行阶段           │                                    │
│   ════════════════════════════════════════════════╪════════════════════                │
│                     Jenkins CI 执行阶段           │                                    │
│                                                   │                                    │
│   ┌──────────┐  ┌──────────┐                      │                                    │
│   │11.打包   │─▶│12.推送   │ ◀────────────────────┘                                    │
│   │  镜像    │  │  镜像    │                                                            │
│   │  build   │  │  push    │                                                            │
│   └──────────┘  └────┬─────┘                                                            │
│                      │                                                                  │
│   ═══════════════════╪══════════════════════════════════════════════════                │
│                      │              平台 CD 执行阶段                                    │
│                      ▼                                                                  │
│                 ┌──────────┐         ┌──────────┐         ┌──────────┐                  │
│                 │13.人工   │────────▶│14.部署   │────────▶│ 钉钉通知 │                  │
│                 │  审批    │         │  deploy  │         │          │                  │
│                 │ approval │         └──────────┘         └──────────┘                  │
│                 └──────────┘                                                            │
│                                                                                          │
└──────────────────────────────────────────────────────────────────────────────────────────┘
```

### 1.3 阶段状态展示

```
清理 [✅ 3s] → 检出 [✅ 5s] → 依赖 [✅ 30s] → 编译 [✅ 45s] → 测试 [✅ 12s] → 检查 [✅ 8s]
                                                                                    ↓
代码扫描 [✅ 60s] → 质量门禁 [✅ OK] → 构建制品 [✅ 15s] → 上传制品库 [✅ 5s]
                                                                          ↓
打包镜像 [✅ 30s] → 推送镜像 [✅ 10s] → 审批 [⏳ 待通过] → 部署 [⏳ 等待]
```

### 1.4 14 阶段定义

| 序号 | 阶段类型 | 阶段名称 | 默认启用 | 说明 |
|:----:|----------|----------|:--------:|------|
| 1 | `clean` | 清理工作空间 | ✅ | 清理 Jenkins 工作目录 |
| 2 | `checkout` | 代码检出 | ✅ | Git 代码拉取 |
| 3 | `dependencies` | 依赖下载 | ✅ | go mod / mvn / npm install / pip install |
| 4 | `compile` | 编译检查 | ✅ | 编译源代码 |
| 5 | `test` | 单元测试 | ✅ | 运行测试用例 |
| 6 | `lint` | 代码检查 | ✅ | 静态代码分析（golangci-lint/eslint/pylint） |
| 7 | `sonar` | SonarQube 代码扫描 | ❌ | SonarQube 代码质量扫描（可选开启） |
| 8 | `quality_gate` | 质量门禁检查 | ❌ | 等待 SonarQube 质量门禁结果，不通过则中止 |
| 9 | `build_binary` | 构建制品 | ✅ | 编译产出二进制/JAR/dist 等制品 |
| 10 | `upload_artifact` | 上传制品库 | ✅ | 将制品上传到平台制品库 |
| 11 | `build` | 打包镜像 | ✅ | Docker build（纯运行时 Dockerfile，不含编译） |
| 12 | `push` | 推送镜像 | ✅ | 推送到 Harbor 镜像仓库 |
| 13 | `approval` | 人工审批 | ❌ | 生产环境或 require_approval=true 时自动激活 |
| 14 | `deploy` | 部署 | ❌ | 平台侧执行 K8s 部署（Deployment/StatefulSet/DaemonSet） |

---

## 二、SonarQube 代码质量门禁

### 2.1 集成架构

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Jenkins   │────▶│  SonarQube   │────▶│    平台     │
│             │     │              │     │             │
│ sonar-scan  │     │ 质量分析引擎 │     │ 指标存储    │
│ + 门禁等待  │     │ + 质量门禁   │     │ + 报告展示  │
└─────────────┘     └──────────────┘     └─────────────┘
      │                    │                    ▲
      │  1. 推送代码扫描   │                    │
      ├────────────────────┘                    │
      │  2. 等待质量门禁结果                     │
      ├────────────────────                     │
      │  3. 获取指标数据                         │
      ├──────────────────────────────────────────┘
      │  4. 回调平台保存 SonarQube 报告
      └──────────────────────────────────────────
```

### 2.2 质量门禁决策

| 门禁状态 | 值 | 行为 |
|----------|------|------|
| `OK` | 通过 | 继续后续阶段 |
| `WARN` | 警告 | 继续但记录警告 |
| `ERROR` | 未通过 | **中止流水线，error() 阻断** |
| `NONE` | 未扫描 | 跳过（EnableSonar=false） |

### 2.3 SonarQube 报告指标

```go
type StageSonarInfo struct {
    ProjectKey       string  // SonarQube 项目 Key
    QualityGate      string  // 质量门禁状态: OK/WARN/ERROR/NONE
    DashboardURL     string  // SonarQube Dashboard 链接
    Bugs             int     // Bug 数量
    Vulnerabilities  int     // 漏洞数量
    CodeSmells       int     // 代码异味数量
    Coverage         float64 // 代码覆盖率 (%)
    Duplications     float64 // 重复代码率 (%)
    SecurityHotspots int     // 安全热点数量
    ReliabilityRating string // 可靠性评级: A/B/C/D/E
    SecurityRating    string // 安全性评级
    Maintainability   string // 可维护性评级
}
```

### 2.4 四种语言的扫描方式

| 语言 | 扫描工具 | 覆盖率报告 |
|------|----------|-----------|
| Go | sonar-scanner CLI | `-Dsonar.go.coverage.reportPaths=coverage.out` |
| Java | `mvn sonar:sonar` (Maven 插件) | JaCoCo 自动集成 |
| Frontend | sonar-scanner CLI | lcov / clover |
| Python | sonar-scanner CLI | `-Dsonar.python.coverage.reportPaths=coverage.xml` |

---

## 三、制品库管理

### 3.1 制品类型

| 类型 | 常量值 | 说明 | 来源语言 |
|------|--------|------|----------|
| JAR | `jar` | Java JAR 包 | Java |
| WAR | `war` | Java WAR 包 | Java |
| Binary | `binary` | Go 二进制 | Go |
| Dist | `dist` | 前端构建产物 (dist.tar.gz) | Frontend |
| Wheel | `wheel` | Python wheel 包 | Python |
| Image | `image` | Docker 镜像引用（不存储文件） | 所有 |
| Archive | `archive` | 通用压缩包 | 通用 |

### 3.2 制品库 API 接口

| 接口 | 方法 | 功能 | 说明 |
|------|------|------|------|
| `/api/v1/k8s/cicd/artifact/list` | GET | 制品列表 | 分页 + 按类型/语言/状态筛选 |
| `/api/v1/k8s/cicd/artifact/detail` | GET | 制品详情 | 获取单个制品完整信息 |
| `/api/v1/k8s/cicd/artifact/by-run` | GET | 运行制品 | 获取某次流水线运行产出的所有制品 |
| `/api/v1/k8s/cicd/artifact/create` | POST | 创建记录 | 创建制品记录（镜像类型，无需上传文件） |
| `/api/v1/k8s/cicd/artifact/upload` | POST | 上传制品 | Jenkins 回调上传 / 手动上传文件 |
| `/api/v1/k8s/cicd/artifact/update` | POST | 更新制品 | 修改制品名称/版本/类型/镜像信息 |
| `/api/v1/k8s/cicd/artifact/download` | GET | 下载制品 | 下载制品文件（自动增加下载计数） |
| `/api/v1/k8s/cicd/artifact/delete` | POST | 删除制品 | 软删除 + 清理文件 |
| `/api/v1/k8s/cicd/artifact/batch-delete` | POST | 批量删除 | 批量软删除（最多100条） |
| `/api/v1/k8s/cicd/artifact/stats` | GET | 制品统计 | 按类型分组统计数量和总大小 |

### 3.3 制品存储架构

```
storage/artifacts/
├── {pipeline_id}/           # 按流水线分目录
│   └── {YYYYMMDD}/         # 按日期分目录
│       ├── app-v1.0.0.jar
│       ├── app-v1.0.1.jar
│       └── dist.tar.gz
├── manual/                  # 手动上传
│   └── {YYYYMMDD}/
│       └── custom-tool.tar.gz
```

### 3.4 制品生命周期

```
上传中 (uploading) ──▶ 就绪 (ready) ──▶ 已过期 (expired)
                                  │
                                  └──▶ 已删除 (deleted)  [软删除 + 文件清理]
```

---

## 四、人工审批机制（核心亮点）

### 4.1 审批触发条件

| 配置字段 | 类型 | 作用 |
|----------|------|------|
| `require_approval` | bool | 是否需要人工审批 |
| `deploy_env` | string | 部署环境（prod 生产环境建议强制开启） |
| `enable_sonar` | bool | 是否启用 SonarQube 代码扫描 |

**触发逻辑**：
```go
// 生产环境或显式配置需要审批
if pipeline.DeployEnv == "prod" || pipeline.RequireApproval {
    // 创建审批记录，状态设为 waiting
}

// SonarQube 开关控制
if pipeline.EnableSonar {
    // 动态插入 sonar + quality_gate 阶段
}
```

### 4.2 审批状态机

```
               构建成功 + 镜像推送完成
                        │
                        ▼
           ┌─────────────────────────┐
           │  审批阶段 (waiting)     │ ← 等待人工操作
           └───────────┬─────────────┘
                       │
              ┌───────┴───────┐
              │               │
           approve         reject
              │               │
              ▼               ▼
         ┌─────────┐    ┌─────────┐
         │ success │    │ failed  │
         └────┬────┘    └─────────┘
              │
              ▼
         自动触发部署阶段
```

### 4.3 审批 API 接口

| 接口 | 方法 | 功能 |
|------|------|------|
| `/api/v1/k8s/cicd/approval/list` | GET | 获取审批列表 |
| `/api/v1/k8s/cicd/approval/pending` | GET | 获取待审批列表 |
| `/api/v1/k8s/cicd/approval/action` | POST | 执行审批操作（approve/reject） |
| `/api/v1/k8s/cicd/stage/approve` | POST | 阶段级审批 |

---

## 五、Jenkins 与平台协作

### 5.1 回调机制

```
┌─────────────┐                      ┌─────────────┐
│   Jenkins   │                      │    平台     │
└──────┬──────┘                      └──────┬──────┘
       │                                    │
       │  1. 阶段回调 (每阶段实时)          │
       │  POST /stage/callback             │
       │  {stage_type, status}             │
       │ ──────────────────────────────────▶│
       │                                    │ 更新阶段状态
       │                                    │
       │  2. SonarQube 报告回调             │
       │  POST /pipeline/sonar-callback    │
       │  {quality_gate, bugs, coverage..} │
       │ ──────────────────────────────────▶│
       │                                    │ 保存代码质量数据
       │                                    │
       │  3. 制品上传回调                   │
       │  POST /artifact/upload            │
       │  {file, pipeline_id, run_id}      │
       │ ──────────────────────────────────▶│
       │                                    │ 保存制品到制品库
       │                                    │
       │  4. 最终回调 (构建完成)            │
       │  POST /pipeline/callback          │
       │  {status, image_url, digest}      │
       │ ──────────────────────────────────▶│
       │                                    │ 创建审批/部署阶段
       │                                    │ 发送钉钉通知
       │                                    │
```

### 5.2 回调安全：HMAC 签名验证

```go
// Jenkins 发送请求时计算签名
signature = HMAC-SHA256(secret, job_name + build_number + status)

// 请求头
X-Signature: <signature>

// 平台验证
func VerifyHMACSignature(signature, jobName string, buildNumber int, status string) bool {
    data := fmt.Sprintf("%s%d%s", jobName, buildNumber, status)
    expected := computeHMAC(hmacSecret, data)
    return hmac.Equal([]byte(signature), []byte(expected))
}
```

### 5.3 Jenkinsfile 完整阶段回调示例

```groovy
// 14 阶段中每个阶段的回调
stage('编译检查') {
    steps {
        script { stageCallback('compile', 'running') }
        sh 'go build ./...'
    }
    post {
        success { script { stageCallback('compile', 'success') } }
        failure { script { stageCallback('compile', 'failed') } }
    }
}

stage('SonarQube Analysis') {
    when { expression { return params.ENABLE_SONAR == 'true' } }
    steps {
        script { stageCallback('sonar', 'running') }
        withSonarQubeEnv('SonarQube') {
            sh "sonar-scanner -Dsonar.projectKey=${params.SONAR_PROJECT_KEY} ..."
        }
    }
    post {
        success { script { stageCallback('sonar', 'success'); sonarReportCallback() } }
        failure { script { stageCallback('sonar', 'failed') } }
    }
}

stage('Quality Gate') {
    when { expression { return params.ENABLE_SONAR == 'true' } }
    steps {
        script { stageCallback('quality_gate', 'running') }
        timeout(time: 5, unit: 'MINUTES') {
            waitForQualityGate abortPipeline: false
        }
    }
}
```

---

## 六、通知闭环设计

### 6.1 通知时机

```
┌──────────────────────────────────────────────────────────────────────┐
│                        钉钉群通知                                     │
├──────────────────────────────────────────────────────────────────────┤
│  ✅ 构建成功 → ⏳ 待审批 → ✅ 部署成功                                │
│  ❌ 构建失败              ❌ 部署失败                                 │
│  🚫 质量门禁未通过       ❌ 审批拒绝                                 │
│  📊 SonarQube 扫描报告                                               │
└──────────────────────────────────────────────────────────────────────┘
```

### 6.2 待审批通知示例

```markdown
### ⏳ 待审批

**流水线**: my-frontend-app
**环境**: 🚀 生产环境
**分支**: main
**构建号**: #42
**镜像**: harbor.example.com/proj/app:main-abc123
**SonarQube**: ✅ 通过 (覆盖率: 78.5%)
**时间**: 2024-01-01 12:00:00

---
✅ [点击进行审批](http://platform.example.com/#/cicd/pipeline/15)
🛠 [查看 Jenkins 构建日志](http://jenkins.example.com/job/app/42/console)
📊 [查看 SonarQube 报告](http://sonar.example.com/dashboard?id=app)
```

---

## 七、双路发布模型

### 7.1 Pipeline 模式（单项目 CI/CD）

```
用户点击"运行" → 触发 Jenkins 构建 → 14 阶段逐步执行 → 审批 → 部署
```

### 7.2 Release 模式（多集群批量部署）

```
创建发布单 → 选择多个集群/多个流水线 → 统一触发 → CAS 状态机管理 → 批量部署
```

| 特性 | Pipeline 模式 | Release 模式 |
|------|---------------|--------------|
| 适用场景 | 单项目迭代 | 版本发布/多环境同步 |
| 集群选择 | 固定一个 | 动态多集群 |
| 审批粒度 | 流水线级 | 发布单级 |
| 回滚能力 | 阶段级回滚 | 批量回滚 |

---

## 八、流水线配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `auto_deploy` | bool | 是否自动部署（构建成功后） |
| `require_approval` | bool | 是否需要人工审批 |
| `enable_sonar` | bool | 是否启用 SonarQube 代码扫描 |
| `target_cluster_id` | int64 | 目标集群ID |
| `target_namespace` | string | 目标命名空间 |
| `target_workload_kind` | string | 工作负载类型 (Deployment/StatefulSet/DaemonSet) |
| `target_workload_name` | string | 工作负载名称 |
| `target_container` | string | 容器名称 |
| `deploy_env` | string | 部署环境 (dev/test/staging/prod) |

---

## 九、技术亮点总结

| 特性 | 实现方式 | 价值 |
|------|----------|------|
| **14阶段全链路** | 14种阶段类型独立记录状态和日志 | 精细化追踪，问题定位快 |
| **SonarQube 门禁** | 质量不达标自动阻断发布 | 代码质量可量化、可管控 |
| **制品库管理** | 完整 CRUD + 文件存储 + SHA256 校验 | 构建产物可追溯、可审计 |
| **审批可配置** | 流水线级别可开关 | 灵活适配不同环境需求 |
| **审批记录** | 记录审批人、时间、意见 | 满足审计合规要求 |
| **审批后自动部署** | 审批通过自动激活 deploy 阶段 | 减少人工操作，提升效率 |
| **HMAC 签名** | Jenkins 回调签名验证 | 防止伪造回调，安全可靠 |
| **钉钉即时通知** | 构建/扫描/审批/部署实时推送 | 团队实时感知发布状态 |
| **平台编译+Docker纯打包** | Jenkins编译，Dockerfile仅打包运行时 | 镜像精简、安全、速度快 |
| **双路发布** | Pipeline 模式 + Release 模式 | 覆盖单项目和多集群批量场景 |
| **幂等回调** | job_name + build_number 唯一键 | 防止重复处理 |
| **CAS 状态机** | 发布单状态 Compare-And-Swap | 防并发状态竞争 |

---

## 十、相关代码文件

| 文件路径 | 说明 |
|----------|------|
| **模型层** | |
| `internal/app/models/cicd_pipeline.go` | 流水线/阶段/SonarQube 数据模型 |
| `internal/app/models/cicd_artifact.go` | 制品库数据模型 |
| **DAO层** | |
| `internal/app/dao/cicd_pipeline.go` | 流水线数据访问 |
| `internal/app/dao/cicd_artifact.go` | 制品库数据访问（CRUD + 统计） |
| **服务层** | |
| `internal/app/services/cicd_stage.go` | 阶段业务逻辑（14阶段定义 + 审批） |
| `internal/app/services/cicd_pipeline.go` | 流水线业务逻辑 |
| `internal/app/services/cicd_artifact.go` | 制品库业务逻辑（上传/下载/更新/批量删除） |
| `internal/app/services/cicd_release.go` | 发布单业务逻辑 |
| `internal/app/services/cicd_executor.go` | K8s 部署执行器 |
| `internal/app/services/cicd_finalize.go` | 发布单完结判定 |
| `internal/app/services/cicd_notify.go` | 钉钉通知服务 |
| **控制器层** | |
| `internal/app/controllers/api/v1/cicd/artifact_controller.go` | 制品库控制器（10个接口） |
| `internal/app/controllers/api/v1/cicd/pipeline_controller.go` | 流水线控制器 |
| `internal/app/controllers/api/v1/cicd/stage_controller.go` | 阶段控制器 |
| **路由层** | |
| `internal/app/routers/kube_cicd/cicd_router.go` | 需认证路由（含制品库路由） |
| `internal/app/routers/kube_cicd/cicd_callback_router.go` | 公开回调路由（HMAC 验证） |
| **Jenkins 模板** | |
| `configs/jenkins-templates/go-pipeline.groovy` | Go 流水线模板 |
| `configs/jenkins-templates/java-spring-pipeline.groovy` | Java 流水线模板 |
| `configs/jenkins-templates/frontend-pipeline.groovy` | 前端流水线模板 |
| `configs/jenkins-templates/python-pipeline.groovy` | Python 流水线模板 |
| **前端** | |
| `k8s-web/src/views/cicd/PipelineDetail.vue` | 流水线详情/阶段展示/审批交互 |
| `k8s-web/src/views/cicd/Artifacts.vue` | 制品库管理页面 |
| **SQL** | |
| `docs/sql/k8s_platform_full_init.sql` | 全量数据库初始化（含 cicd_artifact 表） |
| `docs/sql/cicd_artifact.sql` | 制品库独立建表 SQL |

---

## 十一、面试回答要点

1. **为什么要 CI/CD 解耦？**
   - Jenkins 专注构建，平台专注部署和审批
   - 生产环境部署需要人工卡点，防止误发布
   - 部署目标可灵活选择（多集群支持）

2. **14 阶段设计的价值？**
   - 每个阶段独立状态、日志、耗时记录
   - SonarQube 和质量门禁可按需开关
   - 制品上传与镜像构建分离，支持独立追溯

3. **SonarQube 质量门禁如何工作？**
   - Jenkins 推送扫描 → SonarQube 分析 → waitForQualityGate 等结果
   - 质量不达标（ERROR）自动 error() 中止流水线
   - 扫描指标通过回调保存到平台，前端可查看报告

4. **制品库的设计考虑？**
   - 完整 CRUD + 批量删除 + 按 RunID 查询
   - 文件本地存储 + SHA256 校验 + 下载计数
   - 自动推导制品类型（文件扩展名 + 语言类型）
   - 软删除保留记录，物理文件同步清理

5. **审批机制如何保证安全？**
   - 生产环境强制审批
   - 审批记录完整（who/when/why），满足审计
   - HMAC 签名防止回调伪造

6. **与 Rancher/KubeSphere 的区别？**
   - 更轻量，专注于多集群权限管控场景
   - 审批流程可配置，适配不同团队需求
   - 与 Jenkins 深度集成，复用现有 CI 能力
   - 自带制品库和 SonarQube 集成，一站式闭环
