# Jenkins 构建缓存与 SonarQube 性能优化文档

> **适用范围**：K8s Operation Platform 全部 4 个 CI/CD 流水线模板  
> **模板路径**：`configs/jenkins-templates/`  
> **更新日期**：2026-04-24

---

## 一、优化概览

### 1.1 优化目标

在**不影响生产环境、不影响新代码提交**的前提下，通过缓存复用和流程精简，显著加速 CI/CD 构建速度。

### 1.2 优化效果（预估）

| 语言模板 | 优化前 | 优化后 | 加速比 |
|---------|--------|--------|--------|
| Java/Spring Boot | ~8 分钟 | ~2-3 分钟 | **60-70%** |
| 前端（Vue/React） | ~5 分钟 | ~1-2 分钟 | **60-70%** |
| Go | ~3 分钟 | ~1.5 分钟 | **50%** |
| Python | ~3 分钟 | ~1.5 分钟 | **50%** |

### 1.3 核心原则

- **源码每次拉最新**：Git checkout 每次获取最新代码，保证构建结果与代码一致
- **依赖缓存持久化**：依赖包由 lock 文件版本锁定，缓存只复用未变化的依赖
- **选择性清理**：只清理源码和构建产物，保留外部缓存目录
- **增量编译**：避免重复 `clean`，复用上次编译产物

---

## 二、缓存架构设计

### 2.1 缓存路径规划

所有缓存目录统一放在 Jenkins 全局路径下（`/var/lib/jenkins/`），不在 workspace 内部，确保 `deleteDir()` 或 `cleanWs()` 不会误删缓存。

| 语言 | 缓存类型 | 缓存路径 | 说明 |
|------|---------|---------|------|
| Java | Maven 本地仓库 | `/var/lib/jenkins/.m2/repository` | 所有 Java 项目共享 |
| 前端 | npm 缓存 | `/var/lib/jenkins/.npm-cache` | 所有前端项目共享 |
| Go | Go Module 缓存 | `/var/lib/jenkins/go/pkg/mod` | Go 生态默认路径 |
| Go | Go 编译缓存 | `/var/lib/jenkins/.cache/go-build` | Go 生态默认路径 |
| Python | pip 缓存 | 系统默认（`~/.cache/pip`） | pip 自带缓存机制 |

### 2.2 缓存安全性保证

```
新代码提交 → Git checkout 拉最新 → 依赖缓存命中(如版本未变) → 增量编译 → 构建镜像
                                  → 依赖缓存未命中(如版本变了) → 下载新依赖+更新缓存 → 全量编译
```

- **Maven**：依赖版本由 `pom.xml` 锁定，缓存中的 jar 是按 GAV 坐标存储的，版本变化会自动下载新版本
- **npm**：依赖版本由 `package-lock.json` 锁定，`npm ci` 严格按 lock 文件安装
- **Go**：依赖版本由 `go.sum` 校验，`go mod download` 自动处理增量更新
- **pip**：依赖版本由 `requirements.txt` 指定，pip 自动判断是否需要重新下载

---

## 三、各模板优化详情

### 3.1 Java/Spring Boot 模板

**模板文件**：`configs/jenkins-templates/java-spring-pipeline.groovy`

#### 3.1.1 Maven 缓存持久化

```groovy
environment {
    MAVEN_OPTS       = '-Xmx1024m -Xms512m -XX:+TieredCompilation -XX:TieredStopAtLevel=1'
    // Maven 本地仓库放在 workspace 外部，跨构建持久缓存
    MAVEN_LOCAL_REPO = '/var/lib/jenkins/.m2/repository'
}
```

**变更说明**：
- **优化前**：`MAVEN_LOCAL_REPO = '${env.WORKSPACE}/.m2/repository'`（每次 `deleteDir()` 被清除）
- **优化后**：迁移到 `/var/lib/jenkins/.m2/repository`（跨构建持久保留）
- **效果**：首次构建下载依赖后，后续构建秒级复用（节省 2-3 分钟）

#### 3.1.2 选择性工作区清理

```groovy
stage('Clean Workspace') {
    steps {
        // 选择性清理：保留外部缓存目录，只清源码区
        sh 'find . -mindepth 1 -maxdepth 1 ! -name ".m2" | xargs rm -rf 2>/dev/null || true'
        // ... checkout ...
    }
}
```

**变更说明**：
- **优化前**：`deleteDir()` 全量清空工作区（包括缓存）
- **优化后**：`find` 命令选择性清理，排除 `.m2` 目录

#### 3.1.3 增量编译

```groovy
stage('Compile') {
    steps {
        echo "=== Maven 编译（增量模式） ==="
        sh "mvn compile -DskipTests -B -T 1C -s ${MVN_SETTINGS} -Dmaven.repo.local=${MAVEN_LOCAL_REPO}"
    }
}

stage('Package') {
    steps {
        echo "=== 打包（跳过 clean，复用已编译 class） ==="
        sh "mvn package -DskipTests -B -T 1C -s ${MVN_SETTINGS} -Dmaven.repo.local=${MAVEN_LOCAL_REPO}"
    }
}
```

**变更说明**：
- **优化前**：`mvn clean compile` + `mvn ${params.MAVEN_GOALS}`（重复 clean，双重编译）
- **优化后**：`mvn compile`（增量）+ `mvn package`（复用已编译 class）
- **关键参数**：`-T 1C` 启用 Maven 并行构建（每个 CPU 核心一个线程）

#### 3.1.4 构建后选择性清理

```groovy
post {
    always {
        sh "nerdctl rmi ${env.FULL_IMAGE} || true"
        // 清理源码和 target，但保留外部 Maven 缓存
        sh 'rm -rf target src .git 2>/dev/null || true'
    }
}
```

**变更说明**：
- **优化前**：`cleanWs()` 全量清空（包括 Maven 缓存）
- **优化后**：只清理 `target`、`src`、`.git`，Maven 缓存不受影响

---

### 3.2 前端模板（Vue/React/Angular）

**模板文件**：`configs/jenkins-templates/frontend-pipeline.groovy`

#### 3.2.1 npm 缓存持久化

```groovy
environment {
    NPM_REGISTRY  = 'https://registry.npmmirror.com'
    // npm 缓存放 workspace 外，跨构建持久复用
    NPM_CACHE_DIR = '/var/lib/jenkins/.npm-cache'
}
```

#### 3.2.2 依赖安装使用缓存

```groovy
stage('Install Dependencies') {
    steps {
        sh """
            set -e
            npm config set registry ${NPM_REGISTRY}
            npm config set cache ${NPM_CACHE_DIR}
            npm ci --prefer-offline --cache ${NPM_CACHE_DIR} || \
            npm install --prefer-offline --cache ${NPM_CACHE_DIR}
        """
    }
}
```

**关键参数**：
- `--prefer-offline`：优先使用本地缓存，缓存未命中才走网络
- `--cache ${NPM_CACHE_DIR}`：指定外部缓存目录
- `npm ci` 优先：严格按 `package-lock.json` 安装，保证一致性

#### 3.2.3 构建后选择性清理

```groovy
post {
    always {
        sh "nerdctl rmi ${env.FULL_IMAGE} || true"
        sh 'rm -rf node_modules dist .git 2>/dev/null || true'
    }
}
```

---

### 3.3 Go 模板

**模板文件**：`configs/jenkins-templates/go-pipeline.groovy`

#### 3.3.1 Go 缓存（天然外部化）

Go 生态的缓存默认就在 workspace 外部，无需额外迁移：

```groovy
environment {
    GOPATH     = "/var/lib/jenkins/go"
    GOMODCACHE = "/var/lib/jenkins/go/pkg/mod"      // Module 缓存
    GOCACHE    = "/var/lib/jenkins/.cache/go-build"  // 编译缓存
    GOPROXY    = 'https://goproxy.cn,direct'         // 国内代理加速
}
```

#### 3.3.2 构建后选择性清理

```groovy
post {
    always {
        sh "nerdctl rmi ${env.FULL_IMAGE} || true"
        sh 'rm -rf bin .git coverage.out 2>/dev/null || true'
    }
}
```

**变更说明**：
- **优化前**：`cleanWs()` 全量清空
- **优化后**：只清理 `bin`、`.git`、`coverage.out`，Go Module 缓存不受影响

---

### 3.4 Python 模板

**模板文件**：`configs/jenkins-templates/python-pipeline.groovy`

#### 3.4.1 pip 国内镜像加速

```groovy
environment {
    PIP_INDEX_URL = 'https://pypi.tuna.tsinghua.edu.cn/simple'
}
```

#### 3.4.2 构建后选择性清理

```groovy
post {
    always {
        sh "nerdctl rmi ${env.FULL_IMAGE} || true"
        sh 'rm -rf .git venv __pycache__ coverage.xml 2>/dev/null || true'
    }
}
```

---

## 四、SonarQube 扫描性能优化

### 4.1 优化背景

SonarQube 默认扫描模式下，`git blame` 逐行追溯（SCM Analysis）占用 **60%+** 的扫描时间。在 CI 场景中，这个信息并非必须。

### 4.2 核心优化项

| 优化项 | 参数 | 节省时间 | 影响 |
|-------|------|---------|------|
| 禁用 Git Blame | `-Dsonar.scm.disabled=true` | ~60% | 无法显示每行代码作者信息，不影响缺陷检测 |
| 复用 Maven 缓存 | `-Dmaven.repo.local=${MAVEN_LOCAL_REPO}` | ~20% | 避免 sonar 插件重复下载依赖 |
| 跳过测试编译 | `-DskipTests` | ~10% | 扫描阶段无需执行测试 |
| 分离质量门禁等待 | `-Dsonar.qualitygate.wait=false` | ~10% | 由独立 `waitForQualityGate()` 阶段处理 |
| 移除非必要参数 | 移除 `branch.name`/`links.scm` | 微小 | 减少开销 |

### 4.3 Java 模板 SonarQube 配置示例

```groovy
stage('SonarQube Analysis') {
    when { expression { return params.ENABLE_SONAR } }
    steps {
        script {
            withSonarQubeEnv('SonarQube') {
                sh """
                    mvn sonar:sonar \\
                        -s ${env.MVN_SETTINGS} \\
                        -Dmaven.repo.local=${MAVEN_LOCAL_REPO} \\
                        -DskipTests \\
                        -Dsonar.projectKey=${projectKey} \\
                        -Dsonar.projectName=${projectName} \\
                        -Dsonar.projectVersion=${env.FINAL_TAG} \\
                        -Dsonar.sources=${sources} \\
                        -Dsonar.java.binaries=${binaries} \\
                        -Dsonar.exclusions=${exclusions} \\
                        -Dsonar.scm.disabled=true \\
                        -Dsonar.qualitygate.wait=false \\
                        -Dsonar.links.ci=${env.BUILD_URL} \\
                        -B
                """
            }
        }
    }
}
```

### 4.4 前端/Go/Python 模板 SonarQube 配置

这三个模板使用 `sonar-scanner` 命令行工具（非 Maven 插件），统一添加：

```bash
sonar-scanner \
    -Dsonar.scm.disabled=true \
    -Dsonar.links.ci=${env.BUILD_URL}
    # ... 其他项目参数
```

### 4.5 质量门禁等待优化

```groovy
stage('Quality Gate') {
    steps {
        script {
            // 等待 SonarQube 分析完成（最多 2 分钟，轻量扫描通常 30s 内返回）
            def qg = waitForQualityGate(webhookSecretId: '', abortPipeline: false)
            // ... 质量门禁判断逻辑
        }
    }
}
```

**关键配置**：
- `abortPipeline: false`：门禁不通过时不直接中断，由脚本逻辑控制行为
- `webhookSecretId: ''`：兼容无 webhook secret 的场景

---

## 五、SonarQube 服务端配置要求

### 5.1 Jenkins 系统配置

路径：`Jenkins → Manage Jenkins → System → SonarQube servers`

| 配置项 | 值 | 注意事项 |
|-------|------|---------|
| Name | `SonarQube` | **不能有前后空格**，必须与模板中 `withSonarQubeEnv('SonarQube')` 严格匹配 |
| Server URL | `http://<IP>:30090/` | 确保 Jenkins 服务器可访问 |
| Server authentication token | `sonar-token` | 需在 Jenkins Credentials 中预先创建（类型：Secret text） |

### 5.2 SonarQube Webhook 配置

路径：`SonarQube → Administration → Configuration → Webhooks`

| 配置项 | 值 |
|-------|------|
| Name | `Jenkins` |
| URL | `http://<Jenkins-IP>:8080/sonarqube-webhook/` |

> **说明**：Webhook 用于将质量门禁结果推送回 Jenkins，`waitForQualityGate()` 依赖此配置。

---

## 六、Jenkins 前置准备

### 6.1 缓存目录初始化

首次使用前，需在 Jenkins 服务器上创建缓存目录并设置权限：

```bash
# Maven 缓存目录
sudo mkdir -p /var/lib/jenkins/.m2/repository
sudo chown -R jenkins:jenkins /var/lib/jenkins/.m2

# npm 缓存目录
sudo mkdir -p /var/lib/jenkins/.npm-cache
sudo chown -R jenkins:jenkins /var/lib/jenkins/.npm-cache

# Go 缓存目录（通常已存在）
sudo mkdir -p /var/lib/jenkins/go/pkg/mod
sudo mkdir -p /var/lib/jenkins/.cache/go-build
sudo chown -R jenkins:jenkins /var/lib/jenkins/go /var/lib/jenkins/.cache
```

### 6.2 全局工具配置

路径：`Jenkins → Manage Jenkins → Tools`

| 工具 | 名称 | 路径 |
|------|-----|------|
| Maven | `Maven-3.9` | `/opt/apache-maven-3.9.9`（或实际安装路径） |
| JDK | `JDK-21` | `/usr/lib/jvm/java-21`（或实际安装路径） |

### 6.3 凭证配置

路径：`Jenkins → Manage Jenkins → Credentials`

| 凭证 ID | 类型 | 说明 |
|---------|------|------|
| `harbor-registry` | Username with password | 镜像仓库登录凭证 |
| `hmac-secret` | Secret text | 平台回调签名密钥 |
| `sonar-token` | Secret text | SonarQube 认证 Token |
| `gitee-id` | Username with password | Git 仓库拉取凭证 |

---

## 七、优化前后对比

### 7.1 工作区清理策略对比

| 阶段 | 优化前 | 优化后 |
|------|-------|--------|
| 构建前清理 | `deleteDir()` 全量删除 | 选择性清理（`find + xargs`），保留缓存 |
| 构建后清理 | `cleanWs()` 全量删除 | 只删源码/产物（`rm -rf target src .git`） |

### 7.2 Maven 生命周期对比（Java）

| 阶段 | 优化前 | 优化后 |
|------|-------|--------|
| Compile | `mvn clean compile` | `mvn compile`（增量，不 clean） |
| Package | `mvn ${params.MAVEN_GOALS}` | `mvn package`（复用 compile 结果） |
| 缓存位置 | `${WORKSPACE}/.m2/repository` | `/var/lib/jenkins/.m2/repository` |

### 7.3 SonarQube 扫描对比

| 项目 | 优化前 | 优化后 |
|------|-------|--------|
| SCM 分析 | 启用（git blame 全量追溯） | 禁用（`-Dsonar.scm.disabled=true`） |
| 门禁等待 | mvn 内同步等待 | 独立阶段异步等待 |
| 预估耗时 | 5-8 分钟 | 1-2 分钟 |

---

## 八、常见问题

### Q1：缓存会不会导致依赖版本不一致？

**不会。** 依赖版本由各语言的 lock 文件严格控制：
- Java：`pom.xml` 中的 `<version>` 指定精确版本
- 前端：`package-lock.json` 锁定每个包的精确版本
- Go：`go.sum` 校验每个 module 的哈希
- Python：`requirements.txt` 中使用 `==` 指定精确版本

缓存只是避免重复下载，不会改变安装版本。

### Q2：禁用 SonarQube SCM 分析有什么影响？

仅影响 SonarQube 界面上"每行代码的作者信息"展示，**不影响**：
- Bug 检测
- 漏洞检测
- Code Smell 检测
- 覆盖率分析
- 质量门禁判断

### Q3：缓存占用磁盘空间怎么办？

建议定期清理（可通过 Jenkins 定时任务）：

```bash
# 每月清理一次 Maven 缓存中 60 天未使用的依赖
find /var/lib/jenkins/.m2/repository -atime +60 -type f -name '*.jar' -delete

# 清理 npm 缓存
npm cache clean --force --cache /var/lib/jenkins/.npm-cache

# 清理 Go 编译缓存
GOCACHE=/var/lib/jenkins/.cache/go-build go clean -cache
```

### Q4：新增 Jenkins Agent 节点需要做什么？

1. 按照「六、Jenkins 前置准备」创建缓存目录并设置权限
2. 安装对应语言运行时（Java/Node/Go/Python）
3. 确保 Agent 用户对缓存目录有读写权限

### Q5：SonarQube `withSonarQubeEnv('SonarQube')` 报错 "does not match any configured installation"？

检查 Jenkins 系统配置中 SonarQube 服务器名称是否完全匹配 `SonarQube`，注意**不能有前后空格**。

---

## 九、模板文件清单

| 文件 | 语言 | 行数 | 关键优化 |
|------|------|------|---------|
| `java-spring-pipeline.groovy` | Java/Spring Boot | ~546 | Maven 缓存持久化 + 增量编译 + SonarQube 优化 |
| `frontend-pipeline.groovy` | Vue/React/Angular | ~454 | npm 缓存持久化 + SonarQube 优化 |
| `go-pipeline.groovy` | Go | ~556 | 天然外部缓存 + SonarQube 优化 |
| `python-pipeline.groovy` | Python | ~456 | pip 镜像加速 + SonarQube 优化 |
