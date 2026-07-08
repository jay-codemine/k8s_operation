# CI/CD 流水线阶段化与通知闭环

## 一、功能概述

本功能参考 Rancher、KubeSphere、Jenkins 等平台设计，实现了：

1. **流水线阶段化执行**：每个阶段独立记录状态和日志
2. **人工审批节点**：生产环境部署前可配置审批流程
3. **平台内嵌部署**：部署操作在平台侧执行，非 Jenkins
4. **钉钉通知闭环**：构建/审批/部署结果实时推送

### 阶段流程图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CI/CD 完整流程                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   ┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐            │
│   │ 代码检出 │──▶│   构建   │──▶│   测试   │──▶│ 推送镜像 │            │
│   │ checkout │   │  build   │   │   test   │   │   push   │            │
│   └──────────┘   └──────────┘   └──────────┘   └────┬─────┘            │
│                                                      │                  │
│                        Jenkins 执行阶段              │                  │
│   ═══════════════════════════════════════════════════╪═════════════    │
│                        平台执行阶段                  │                  │
│                                                      ▼                  │
│                                               ┌──────────┐             │
│                                               │ 人工审批 │ (可选)      │
│                                               │ approval │             │
│                                               └────┬─────┘             │
│                                                    │                   │
│                                                    ▼                   │
│                                               ┌──────────┐             │
│                                               │   部署   │             │
│                                               │  deploy  │             │
│                                               └────┬─────┘             │
│                                                    │                   │
│                                                    ▼                   │
│                                               ┌──────────┐             │
│                                               │ 钉钉通知 │             │
│                                               └──────────┘             │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 阶段状态展示

```
代码检出  [✅ 完成 3s]  →  构建  [🔄 运行中]  →  测试  [⏳ 等待]  →  推送镜像  [⏳ 等待]  →  审批  [⏳ 等待]  →  部署  [⏳ 等待]
```

---

## 二、数据库迁移

### 迁移脚本

执行文件：`docs/sql/cicd_pipeline_stage_migration.sql`

```sql
-- 流水线阶段执行记录表
CREATE TABLE IF NOT EXISTS `cicd_pipeline_stage` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `run_id` bigint(20) NOT NULL COMMENT '运行记录ID',
    `pipeline_id` bigint(20) NOT NULL COMMENT '流水线ID',
    `stage_order` int(11) NOT NULL DEFAULT 0 COMMENT '阶段顺序',
    `stage_type` varchar(32) NOT NULL COMMENT '阶段类型: checkout/build/test/push/approval/deploy',
    `stage_name` varchar(100) NOT NULL COMMENT '阶段名称',
    `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '状态: pending/running/success/failed/skipped/waiting/aborted',
    
    -- 审批相关字段
    `approver_id` bigint(20) DEFAULT NULL COMMENT '审批人ID',
    `approve_time` bigint(20) DEFAULT NULL COMMENT '审批时间',
    `approve_comment` varchar(500) DEFAULT '' COMMENT '审批备注',
    
    -- 部署相关字段
    `deploy_cluster_id` bigint(20) DEFAULT NULL COMMENT '部署目标集群ID',
    `deploy_namespace` varchar(100) DEFAULT '' COMMENT '部署命名空间',
    `deploy_workload_kind` varchar(50) DEFAULT '' COMMENT '工作负载类型',
    `deploy_workload_name` varchar(200) DEFAULT '' COMMENT '工作负载名称',
    `deploy_container` varchar(100) DEFAULT '' COMMENT '容器名称',
    `deploy_image` varchar(500) DEFAULT '' COMMENT '部署镜像',
    
    -- 执行信息
    `started_at` bigint(20) DEFAULT NULL COMMENT '开始时间',
    `finished_at` bigint(20) DEFAULT NULL COMMENT '结束时间',
    `duration_sec` int(11) DEFAULT 0 COMMENT '执行耗时(秒)',
    `logs` text COMMENT '执行日志',
    `error_message` varchar(1000) DEFAULT '' COMMENT '错误信息',
    
    `created_at` bigint(20) NOT NULL COMMENT '创建时间',
    `modified_at` bigint(20) NOT NULL COMMENT '修改时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_run_id` (`run_id`),
    KEY `idx_pipeline_id` (`pipeline_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流水线阶段执行记录';
```

### 执行命令

```bash
mysql -u root -p k8s-platform < docs/sql/cicd_pipeline_stage_migration.sql
```

---

## 三、配置说明

### config.yaml 配置

```yaml
# Jenkins 配置（平台驱动 CI/CD）
Jenkins:
  URL: "http://jenkins.example.com:8080"           # Jenkins 服务器地址
  Username: "admin"                                 # Jenkins 用户名
  APIToken: "your-jenkins-api-token"                # Jenkins API Token
  TriggerTimeout: 60                                # 触发构建等待超时(秒)
  
  # 回调机制配置
  CallbackURL: "http://platform.example.com:8080"   # 平台基础地址（用于生成跳转链接）
  HMACSecret: "your-hmac-secret-key"                # HMAC 签名密钥
  PollInterval: 15                                  # 轮询间隔(秒)
  MaxBuildTime: 30                                  # 最大构建时间(分钟)
  
  # 通知配置
  DingTalkWebhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"  # 钉钉机器人 Webhook
```

### 流水线配置选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `auto_deploy` | bool | 是否自动部署（构建成功后） |
| `require_approval` | bool | 是否需要人工审批 |
| `target_cluster_id` | int64 | 目标集群ID |
| `target_namespace` | string | 目标命名空间 |
| `target_workload_kind` | string | 工作负载类型 (Deployment/StatefulSet/DaemonSet) |
| `target_workload_name` | string | 工作负载名称 |
| `target_container` | string | 容器名称 |
| `deploy_env` | string | 部署环境 (dev/staging/prod) |

---

## 四、钉钉通知效果

### 4.1 构建成功通知

```markdown
### ✅ 构建成功

**流水线**: my-frontend-app
**分支**: main
**构建号**: #42
**镜像**: harbor.example.com/proj/app:main-abc123
**耗时**: 120s
**时间**: 2024-01-01 12:00:00

⏳ **等待审批**: 请前往平台进行人工审批

---
🔗 [查看流水线详情](http://platform.example.com/#/cicd/pipeline/15)

🛠 [查看 Jenkins 构建](http://jenkins.example.com:8080/job/my-frontend-app/42/console)
```

### 4.2 待审批通知

```markdown
### ⏳ 待审批

**流水线**: my-frontend-app
**环境**: 🚀 生产环境
**分支**: main
**构建号**: #42
**镜像**: harbor.example.com/proj/app:main-abc123
**时间**: 2024-01-01 12:00:00

---
✅ [点击进行审批](http://platform.example.com/#/cicd/pipeline/15)

🛠 [查看 Jenkins 构建日志](http://jenkins.example.com:8080/job/my-frontend-app/42/console)
```

### 4.3 部署成功通知

```markdown
### ✅ 部署成功

**流水线**: my-frontend-app
**环境**: 🚀 生产环境
**命名空间**: production
**工作负载**: Deployment/my-app
**镜像**: harbor.example.com/proj/app:main-abc123
**时间**: 2024-01-01 12:05:00

---
🔗 [查看流水线详情](http://platform.example.com/#/cicd/pipeline/15)

🛠 [查看 Jenkins 构建](http://jenkins.example.com:8080/job/my-frontend-app/42/console)
```

### 4.4 部署失败通知

```markdown
### ❌ 部署失败

**流水线**: my-frontend-app
**环境**: 🚀 生产环境
**命名空间**: production
**工作负载**: Deployment/my-app
**镜像**: harbor.example.com/proj/app:main-abc123
**时间**: 2024-01-01 12:05:00

**错误**: deployment "my-app" not found

---
🔗 [查看流水线详情](http://platform.example.com/#/cicd/pipeline/15)

🛠 [查看 Jenkins 构建](http://jenkins.example.com:8080/job/my-frontend-app/42/console)
```

---

## 五、完美闭环设计

```
┌────────────────────────────────────────────────────────────────────────────┐
│                           完美 CI/CD 闭环                                  │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│   ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐               │
│   │ Jenkins │───▶│  平台   │───▶│  审批   │───▶│  部署   │               │
│   │  构建   │    │  回调   │    │ (可选)  │    │  K8s    │               │
│   └────┬────┘    └────┬────┘    └────┬────┘    └────┬────┘               │
│        │              │              │              │                     │
│        ▼              ▼              ▼              ▼                     │
│   ┌─────────────────────────────────────────────────────────┐            │
│   │                    钉钉群通知                            │            │
│   │  ✅ 构建成功 → ⏳ 待审批 → ✅ 部署成功                   │            │
│   │  ❌ 构建失败                ❌ 部署失败                  │            │
│   └─────────────────────────────────────────────────────────┘            │
│                                                                            │
│   ┌─────────────────────────────────────────────────────────┐            │
│   │                    平台 UI 实时展示                      │            │
│   │  代码检出[✅] → 构建[✅] → 测试[✅] → 推送[✅]          │            │
│   │                        ↓                                 │            │
│   │              审批[⏳等待] → 部署[✅完成]                 │            │
│   └─────────────────────────────────────────────────────────┘            │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 六、上线准备清单

### 6.1 必须完成

- [ ] **执行数据库迁移**
  ```bash
  mysql -u root -p k8s-platform < docs/sql/cicd_pipeline_stage_migration.sql
  ```

- [ ] **更新配置文件** (config.yaml)
  - 确认 Jenkins 配置完整
  - 添加 DingTalkWebhook（如需通知）

- [ ] **重启后端服务**

### 6.2 Jenkins 配置

- [ ] **添加凭据**（如使用 HMAC 签名验证）
  - Jenkins → 凭据管理 → 添加 `hmac-secret`

- [ ] **更新 Jenkinsfile**（如需阶段级回调）

- [ ] **确认插件**
  - HTTP Request Plugin（用于回调）

### 6.3 钉钉机器人配置（可选）

1. 钉钉群 → 群设置 → 智能群助手 → 添加机器人
2. 选择「自定义」机器人
3. 安全设置选择「自定义关键词」，添加：`构建`、`部署`、`审批`
4. 复制 Webhook URL 到 config.yaml

### 6.4 验证测试

```bash
# 1. 检查数据库表
mysql -u root -p -e "SHOW TABLES LIKE 'cicd_pipeline_stage';" k8s-platform

# 2. 编译后端
cd D:\k8s_re\k8s_operation
go build ./cmd/...

# 3. 启动服务
.\k8soperation.exe
```

---

## 七、API 接口

### 7.1 阶段列表

```
GET /api/v1/k8s/cicd/stage/list?run_id={run_id}
```

### 7.2 阶段日志

```
GET /api/v1/k8s/cicd/stage/logs?stage_id={stage_id}
```

### 7.3 审批操作

```
POST /api/v1/k8s/cicd/stage/approve

Body:
{
    "stage_id": 123,
    "approved": true,
    "comment": "审批通过"
}
```

### 7.4 执行部署

```
POST /api/v1/k8s/cicd/stage/deploy

Body:
{
    "stage_id": 123
}
```

### 7.5 阶段回调（Jenkins 调用）

```
POST /api/v1/k8s/cicd/stage/callback

Body:
{
    "job_name": "my-app",
    "build_number": 42,
    "pipeline_id": 15,
    "stage_type": "build",
    "status": "success"
}
```

---

## 八、相关文件

| 文件路径 | 说明 |
|----------|------|
| `internal/app/models/cicd_pipeline.go` | 阶段数据模型定义 |
| `internal/app/dao/cicd_stage.go` | 阶段 DAO 层 |
| `internal/app/services/cicd_stage.go` | 阶段业务逻辑 |
| `internal/app/services/cicd_notify.go` | 钉钉通知服务 |
| `internal/app/controllers/api/v1/cicd/stage_controller.go` | 阶段控制器 |
| `internal/app/routers/kube_cicd/cicd_router.go` | 路由注册 |
| `k8s-web/src/views/cicd/PipelineDetail.vue` | 前端阶段展示 |
| `docs/sql/cicd_pipeline_stage_migration.sql` | 数据库迁移脚本 |
| `pkg/setting/section.go` | 配置结构体 |
| `configs/config.yaml.example` | 配置示例 |

---

## 九、完整 Jenkinsfile

### 9.1 Jenkinsfile（带阶段级回调）

```groovy
pipeline {
    agent any

    environment {
        // ============ 必填配置 ============
        HARBOR_URL       = 'harbor.example.com'
        HARBOR_PROJECT   = 'myproject'
        IMAGE_NAME       = 'myapp'
        
        // 平台配置
        PLATFORM_URL     = 'http://platform.example.com:8080'
        PIPELINE_ID      = '15'  // 流水线ID（创建流水线后获取）
        
        // Git 配置
        GIT_CREDENTIAL   = 'gitee-id'
        
        // ============ 自动生成 ============
        TIMESTAMP        = sh(script: 'date +%Y%m%d%H%M%S', returnStdout: true).trim()
        SHORT_COMMIT     = ''
        FULL_IMAGE       = ''
        IMAGE_DIGEST     = ''
    }

    parameters {
        string(name: 'BRANCH', defaultValue: 'main', description: '构建分支')
        booleanParam(name: 'SKIP_TEST', defaultValue: false, description: '跳过测试阶段')
    }

    stages {
        stage('代码检出') {
            steps {
                script {
                    stageCallback('checkout', 'running')
                }
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.BRANCH}"]],
                    userRemoteConfigs: [[
                        url: env.GIT_URL ?: scm.userRemoteConfigs[0].url,
                        credentialsId: env.GIT_CREDENTIAL
                    ]]
                ])
                script {
                    SHORT_COMMIT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    FULL_IMAGE = "${HARBOR_URL}/${HARBOR_PROJECT}/${IMAGE_NAME}:${params.BRANCH}-${SHORT_COMMIT}-${TIMESTAMP}"
                    echo "镜像地址: ${FULL_IMAGE}"
                }
            }
            post {
                success { script { stageCallback('checkout', 'success') } }
                failure { script { stageCallback('checkout', 'failed') } }
            }
        }

        stage('构建') {
            steps {
                script {
                    stageCallback('build', 'running')
                }
                // 根据项目类型选择构建方式
                sh '''
                    # Java Maven 项目
                    # mvn clean package -DskipTests
                    
                    # Node.js 项目
                    # npm install && npm run build
                    
                    # Go 项目
                    # go build -o app .
                    
                    echo "构建完成"
                '''
            }
            post {
                success { script { stageCallback('build', 'success') } }
                failure { script { stageCallback('build', 'failed') } }
            }
        }

        stage('测试') {
            when {
                expression { return !params.SKIP_TEST }
            }
            steps {
                script {
                    stageCallback('test', 'running')
                }
                sh '''
                    # 执行测试
                    # mvn test
                    # npm test
                    # go test ./...
                    
                    echo "测试通过"
                '''
            }
            post {
                success { script { stageCallback('test', 'success') } }
                failure { script { stageCallback('test', 'failed') } }
            }
        }

        stage('推送镜像') {
            steps {
                script {
                    stageCallback('push', 'running')
                }
                
                // 构建 Docker 镜像
                sh """
                    nerdctl build -t ${FULL_IMAGE} .
                """
                
                // 推送到 Harbor
                withCredentials([usernamePassword(
                    credentialsId: 'harbor-credentials',
                    usernameVariable: 'HARBOR_USER',
                    passwordVariable: 'HARBOR_PASS'
                )]) {
                    sh """
                        nerdctl login ${HARBOR_URL} -u ${HARBOR_USER} -p ${HARBOR_PASS}
                        nerdctl push ${FULL_IMAGE}
                    """
                }
                
                // 获取镜像 digest
                script {
                    IMAGE_DIGEST = sh(
                        script: "nerdctl inspect ${FULL_IMAGE} --format '{{.RepoDigests}}' | grep -oP 'sha256:[a-f0-9]+'",
                        returnStdout: true
                    ).trim()
                    echo "镜像 Digest: ${IMAGE_DIGEST}"
                }
            }
            post {
                success { script { stageCallback('push', 'success') } }
                failure { script { stageCallback('push', 'failed') } }
            }
        }
    }

    post {
        success {
            script {
                echo "========== 构建成功 =========="
                
                // 回调平台：构建成功
                def callbackBody = [
                    job_name     : env.JOB_NAME,
                    build_number : env.BUILD_NUMBER.toInteger(),
                    pipeline_id  : env.PIPELINE_ID.toLong(),
                    status       : 'success',
                    git_commit   : SHORT_COMMIT,
                    git_branch   : params.BRANCH,
                    image_url    : FULL_IMAGE,
                    image_digest : IMAGE_DIGEST,
                    duration_sec : currentBuild.duration.intdiv(1000),
                    build_url    : env.BUILD_URL
                ]
                
                platformCallback('/api/v1/k8s/cicd/pipeline/callback', callbackBody)
            }
        }
        
        failure {
            script {
                echo "========== 构建失败 =========="
                
                // 回调平台：构建失败
                def callbackBody = [
                    job_name     : env.JOB_NAME,
                    build_number : env.BUILD_NUMBER.toInteger(),
                    pipeline_id  : env.PIPELINE_ID.toLong(),
                    status       : 'failed',
                    git_commit   : SHORT_COMMIT ?: '',
                    git_branch   : params.BRANCH,
                    duration_sec : currentBuild.duration.intdiv(1000),
                    build_url    : env.BUILD_URL,
                    error_message: currentBuild.description ?: '构建失败'
                ]
                
                platformCallback('/api/v1/k8s/cicd/pipeline/callback', callbackBody)
            }
        }
        
        aborted {
            script {
                echo "========== 构建中止 =========="
                
                def callbackBody = [
                    job_name     : env.JOB_NAME,
                    build_number : env.BUILD_NUMBER.toInteger(),
                    pipeline_id  : env.PIPELINE_ID.toLong(),
                    status       : 'aborted',
                    git_branch   : params.BRANCH,
                    duration_sec : currentBuild.duration.intdiv(1000),
                    build_url    : env.BUILD_URL
                ]
                
                platformCallback('/api/v1/k8s/cicd/pipeline/callback', callbackBody)
            }
        }
        
        always {
            // 清理工作空间（可选）
            cleanWs(cleanWhenSuccess: false)
        }
    }
}

// ==================== 辅助函数 ====================

/**
 * 阶段回调：实时更新阶段状态到平台
 */
def stageCallback(String stageType, String status) {
    try {
        def body = [
            job_name     : env.JOB_NAME,
            build_number : env.BUILD_NUMBER.toInteger(),
            pipeline_id  : env.PIPELINE_ID.toLong(),
            stage_type   : stageType,
            status       : status
        ]
        
        // 计算 HMAC 签名（可选）
        def signature = ''
        withCredentials([string(credentialsId: 'hmac-secret', variable: 'HMAC_SECRET')]) {
            if (HMAC_SECRET) {
                def data = "${env.JOB_NAME}:${env.BUILD_NUMBER}:${stageType}"
                signature = sh(
                    script: "echo -n '${data}' | openssl dgst -sha256 -hmac '${HMAC_SECRET}' | awk '{print \$2}'",
                    returnStdout: true
                ).trim()
            }
        }
        
        def headers = [[name: 'Content-Type', value: 'application/json']]
        if (signature) {
            headers.add([name: 'X-Signature', value: signature])
        }
        
        httpRequest(
            url: "${PLATFORM_URL}/api/v1/k8s/cicd/stage/callback",
            httpMode: 'POST',
            contentType: 'APPLICATION_JSON',
            requestBody: groovy.json.JsonOutput.toJson(body),
            customHeaders: headers,
            validResponseCodes: '200:299',
            timeout: 30,
            quiet: true
        )
        
        echo "[阶段回调] ${stageType} -> ${status}"
    } catch (Exception e) {
        echo "[阶段回调] 失败（不影响构建）: ${e.message}"
    }
}

/**
 * 平台回调：发送构建结果
 */
def platformCallback(String path, Map body) {
    try {
        // 计算 HMAC 签名（可选）
        def signature = ''
        withCredentials([string(credentialsId: 'hmac-secret', variable: 'HMAC_SECRET')]) {
            if (HMAC_SECRET) {
                def data = "${body.job_name}:${body.build_number}:${body.status}"
                signature = sh(
                    script: "echo -n '${data}' | openssl dgst -sha256 -hmac '${HMAC_SECRET}' | awk '{print \$2}'",
                    returnStdout: true
                ).trim()
            }
        }
        
        def headers = [[name: 'Content-Type', value: 'application/json']]
        if (signature) {
            headers.add([name: 'X-Signature', value: signature])
        }
        
        def response = httpRequest(
            url: "${PLATFORM_URL}${path}",
            httpMode: 'POST',
            contentType: 'APPLICATION_JSON',
            requestBody: groovy.json.JsonOutput.toJson(body),
            customHeaders: headers,
            validResponseCodes: '200:299',
            timeout: 60
        )
        
        echo "[平台回调] 成功: ${response.status}"
    } catch (Exception e) {
        echo "[平台回调] 失败: ${e.message}"
    }
}
```

### 9.2 配置说明

#### 必须修改的变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `HARBOR_URL` | Harbor 地址 | `harbor.example.com` |
| `HARBOR_PROJECT` | Harbor 项目名 | `myproject` |
| `IMAGE_NAME` | 镜像名称 | `myapp` |
| `PLATFORM_URL` | 平台地址 | `http://192.168.1.100:8080` |
| `PIPELINE_ID` | 流水线ID | 创建流水线后从平台获取 |
| `GIT_CREDENTIAL` | Git凭据ID | `gitee-id` |

#### Jenkins 凭据配置

| 凭据ID | 类型 | 用途 |
|--------|------|------|
| `gitee-id` | Username/Password | Git 仓库认证 |
| `harbor-credentials` | Username/Password | Harbor 镜像仓库认证 |
| `hmac-secret` | Secret text | HMAC 签名密钥（可选） |

### 9.3 回调时序图

```
代码检出 开始 → 回调 checkout/running
代码检出 完成 → 回调 checkout/success
    ↓
构建 开始 → 回调 build/running
构建 完成 → 回调 build/success
    ↓
测试 开始 → 回调 test/running
测试 完成 → 回调 test/success
    ↓
推送镜像 开始 → 回调 push/running
推送镜像 完成 → 回调 push/success
    ↓
全部完成 → 回调 /pipeline/callback (status=success)
    ↓
平台创建审批/部署阶段 → 钉钉通知
```

---

## 十、功能启用

在创建/编辑流水线时配置：

| 配置项 | 效果 |
|--------|------|
| 勾选「需要审批」 | 构建成功后进入审批阶段，等待人工操作 |
| 勾选「自动部署」 | 审批通过后（或无需审批时）自动执行部署 |
| 选择「目标集群」 | 指定部署的 K8s 集群 |
| 选择「目标命名空间」 | 指定部署的命名空间 |
| 选择「工作负载类型」 | Deployment / StatefulSet / DaemonSet |
| 填写「工作负载名称」 | 要更新镜像的工作负载 |
| 填写「容器名称」 | 要更新的容器（可选，默认第一个） |
