// ==============================================================================
// K8s Operation Platform - Go 项目通用构建模板（K8s Pod Agent + Kaniko 容器化版）
// ==============================================================================
// 设计理念：一个模板服务 100+ Go 项目，所有项目差异通过参数传入
// 运行模式：Jenkins K8s 动态 Pod Agent，每次构建创建独立 Pod，完成后自动销毁
// 镜像构建：使用 Kaniko（无需 Docker daemon、无需特权容器）
//
// ======================== Jenkins Job 配置方式 ========================
// 推荐使用 "Pipeline script from SCM"（版本化管理，自动同步更新）：
//   1. Jenkins → New Item → Pipeline → 命名为 k8s-builder-go
//   2. Pipeline → Definition: Pipeline script from SCM
//   3. SCM: Git → Repository URL: 平台仓库地址
//   4. Script Path: configs/jenkins-templates/go-pipeline.groovy
//
// ======================== K8s 环境要求 ========================
// Jenkins 需安装 Kubernetes Plugin，并配置 K8s Cloud：
//   Manage Jenkins → Clouds → Kubernetes → 配置 K8s API 地址 + 命名空间
// ==============================================================================

pipeline {
    agent {
        kubernetes {
            yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    jenkins-build: go
spec:
  containers:
  - name: golang
    image: swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/golang:1.24.6-bullseye
    imagePullPolicy: Always
    command: ['sleep', '99d']
    resources:
      requests:
        cpu: '500m'
        memory: '512Mi'
      limits:
        cpu: '2'
        memory: '2Gi'
    env:
    - name: GOPROXY
      value: 'https://goproxy.cn,direct'
    - name: CGO_ENABLED
      value: '0'
    - name: GOOS
      value: 'linux'
    - name: GOARCH
      value: 'amd64'
    volumeMounts:
    - name: go-cache
      mountPath: /go/pkg/mod
    - name: workspace-volume
      mountPath: /home/jenkins/agent
  - name: kaniko
    image: swr.cn-north-4.myhuaweicloud.com/ddn-k8s/gcr.io/kaniko-project/executor:debug
    imagePullPolicy: Always
    command: ['sleep', '99d']
    resources:
      requests:
        cpu: '200m'
        memory: '256Mi'
      limits:
        cpu: '1'
        memory: '1Gi'
    volumeMounts:
    - name: workspace-volume
      mountPath: /home/jenkins/agent
  - name: jnlp
    image: swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/jenkins/inbound-agent:latest-jdk21-linuxarm64
    imagePullPolicy: Always
  volumes:
  - name: go-cache
    persistentVolumeClaim:
      claimName: jenkins-go-cache
  - name: workspace-volume
    emptyDir: {}
"""
        }
    }

    options {
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '20'))
        skipDefaultCheckout(true)
    }

    // ==================== 通用参数（平台自动填充） ====================
    parameters {
        // 必填 - 平台传入
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址（必填）')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库地址（必填，如 harbor.example.com/myproject/myapp）')
        string(name: 'IMAGE_TAG', defaultValue: '', description: '镜像标签（空则自动生成 branch-commit-timestamp）')
        string(name: 'DOCKERFILE_PATH', defaultValue: '', description: 'Dockerfile 路径（空则自动生成纯运行时 Dockerfile）')
        string(name: 'LANGUAGE_TYPE', defaultValue: '', description: '平台注入的语言类型（用于交叉校验，不要手动修改）')

        // 平台回调
        string(name: 'PIPELINE_ID', defaultValue: '', description: '平台流水线ID')
        string(name: 'RUN_ID', defaultValue: '', description: '平台运行记录ID')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')

        // 可选参数
        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过单元测试')
        string(name: 'GO_VERSION', defaultValue: '1.24', description: 'Go 版本')
        string(name: 'GIT_CREDENTIAL_ID', defaultValue: 'gitee-id', description: 'Git 凭证ID')
        string(name: 'REGISTRY_CREDENTIAL_ID', defaultValue: 'harbor-registry', description: '镜像仓库凭证ID')
        string(name: 'HMAC_CREDENTIAL_ID', defaultValue: 'hmac-secret', description: 'HMAC签名凭证ID')

        // SonarQube 代码质量扫描参数
        booleanParam(name: 'ENABLE_SONAR', defaultValue: false, description: '启用 SonarQube 代码质量扫描')
        string(name: 'SONAR_PROJECT_KEY', defaultValue: '', description: 'SonarQube 项目 Key（空则使用 Job 名称）')
        string(name: 'SONAR_PROJECT_NAME', defaultValue: '', description: 'SonarQube 项目名称（空则使用 Job 名称）')
        string(name: 'SONAR_SOURCES', defaultValue: '.', description: '源代码目录')
        string(name: 'SONAR_EXCLUSIONS', defaultValue: '**/vendor/**,**/*_test.go,**/test/**', description: '排除扫描的文件模式')
        booleanParam(name: 'SONAR_QUALITY_GATE', defaultValue: true, description: '启用质量门禁检查（不通过则构建失败）')

        // 平台注入的质量门禁阈值（由平台 UI 配置，自动传入）
        string(name: 'SONAR_COVERAGE_THRESHOLD', defaultValue: '80', description: '代码覆盖率阈值（%）')
        string(name: 'SONAR_NEW_BUGS_MAX', defaultValue: '0', description: '新增 Bug 最大允许数')
        string(name: 'SONAR_CODE_SMELLS_MAX', defaultValue: '10', description: '代码异味最大允许数')
        string(name: 'SONAR_VULNERABILITIES_MAX', defaultValue: '0', description: '安全漏洞最大允许数')
        string(name: 'SONAR_DUPLICATIONS_MAX', defaultValue: '3', description: '代码重复率阈值（%）')
        string(name: 'SONAR_GATE_ACTION', defaultValue: 'block', description: '门禁失败策略: block(阻断) | warn(告警) | skip(跳过)')

        // 制品上传参数
        booleanParam(name: 'ENABLE_ARTIFACT_UPLOAD', defaultValue: true, description: '启用制品上传到平台制品库')
    }

    environment {
        REGISTRY_CREDS = credentials("${params.REGISTRY_CREDENTIAL_ID ?: 'harbor-registry'}")
        HMAC_SECRET    = credentials("${params.HMAC_CREDENTIAL_ID ?: 'hmac-secret'}")
    }

    stages {

        stage('Clean Workspace') {
            steps {
                echo "=== 清理工作空间 + 拉取代码 ==="
                deleteDir()

                script {
                    // 语言类型交叉校验：防止自定义 Job 配错脚本
                    def expectedType = 'go'
                    def actualType = params.LANGUAGE_TYPE?.trim()
                    if (actualType && actualType != expectedType) {
                        def scriptMap = [
                            'go': 'go-pipeline.groovy',
                            'java': 'java-spring-pipeline.groovy',
                            'frontend': 'frontend-pipeline.groovy',
                            'python': 'python-pipeline.groovy'
                        ]
                        def correctScript = scriptMap[actualType] ?: "${actualType}-pipeline.groovy"
                        error("""
=== 模板类型不匹配 ===
平台配置语言类型: ${actualType}
当前模板类型: ${expectedType} (go-pipeline.groovy)

解决方案（二选一）:
  1. 修改 Jenkins Job 的 Script Path 为: configs/jenkins-templates/${correctScript}
  2. 在平台将 Jenkins Job 名称留空，使用自动匹配
""")
                    }

                    if (!params.GIT_REPO?.trim()) { error("GIT_REPO 不能为空") }
                    if (!params.IMAGE_REPO?.trim()) { error("IMAGE_REPO 不能为空") }

                    def targetBranch = params.GIT_BRANCH?.trim() ?: 'main'

                    sh 'rm -rf .git 2>/dev/null || true'

                    checkout([
                        $class: 'GitSCM',
                        branches: [[name: "*/${targetBranch}"]],
                        extensions: [
                            [$class: 'CleanBeforeCheckout', deleteUntrackedNestedRepositories: true],
                            [$class: 'LocalBranch', localBranch: targetBranch],
                            [$class: 'CloneOption', depth: 1, shallow: true, noTags: true, timeout: 10, honorRefspec: true]
                        ],
                        userRemoteConfigs: [[
                            url: params.GIT_REPO,
                            credentialsId: params.GIT_CREDENTIAL_ID ?: 'gitee-id'
                        ]]
                    ])

                    env.TARGET_BRANCH = targetBranch

                    def latestCommit = sh(script: 'git log -1 --format="%h %s (%ci)"', returnStdout: true).trim()
                    echo "[Checkout] ✅ 最新提交: ${latestCommit}"
                    echo "[Checkout] 分支: ${targetBranch} | 仓库: ${params.GIT_REPO}"
                }
            }
        }

        stage('Checkout Info') {
            steps {
                script {
                    env.GIT_COMMIT_SHORT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_FULL  = sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_MSG   = sh(script: 'git log -1 --pretty=%B | head -1', returnStdout: true).trim()
                    env.GIT_BRANCH_NAME  = (env.TARGET_BRANCH ?: 'main').replaceAll('/', '-')
                    env.BUILD_TS = sh(script: 'date +%Y%m%d%H%M%S', returnStdout: true).trim()

                    env.FINAL_TAG = params.IMAGE_TAG?.trim()
                        ? params.IMAGE_TAG.trim()
                        : "${env.GIT_COMMIT_SHORT}-${env.BUILD_TS}"

                    env.FULL_IMAGE = "${params.IMAGE_REPO}:${env.FINAL_TAG}"

                    echo "Commit : ${env.GIT_COMMIT_SHORT}"
                    echo "Branch : ${env.GIT_BRANCH_NAME}"
                    echo "Image  : ${env.FULL_IMAGE}"
                }
            }
            post {
                success { script { stageCallback('checkout', 'success') } }
                failure { script { stageCallback('checkout', 'failed') } }
            }
        }

        stage('Dependencies') {
            steps {
                echo "=== 下载依赖 ==="
                container('golang') {
                    script {
                        if (!fileExists('go.mod')) {
                            echo "未检测到 go.mod，跳过"
                            return
                        }
                        sh '''
                            set -e
                            go version
                            go mod download
                            go mod verify
                        '''
                    }
                }
            }
            post {
                success { script { stageCallback('dependencies', 'success') } }
                failure { script { stageCallback('dependencies', 'failed') } }
            }
        }

        stage('Compile Check') {
            steps {
                echo "=== 编译检查（直接产出最终二进制，Build Image 复用，避免重复编译） ==="
                container('golang') {
                    script {
                        if (!fileExists('go.mod')) { echo "跳过编译检查"; return }
                        def appName = params.GIT_REPO?.tokenize('/')?.last()?.replace('.git', '') ?: 'server'
                        env.APP_NAME = appName
                        env.BINARY_PATH = "bin/${appName}"
                        sh """
                            set -e
                            mkdir -p bin
                            go build -buildvcs=false -ldflags="-s -w -X main.Version=${env.FINAL_TAG} -X main.GitCommit=${env.GIT_COMMIT_FULL}" -o ${env.BINARY_PATH} ./cmd/... || \
                            go build -buildvcs=false -ldflags="-s -w -X main.Version=${env.FINAL_TAG} -X main.GitCommit=${env.GIT_COMMIT_FULL}" -o ${env.BINARY_PATH} .
                        """
                        echo "[编译] 二进制产物: ${env.BINARY_PATH}"
                    }
                }
            }
            post {
                success { script { stageCallback('compile', 'success'); stageCallback('build_binary', 'success') } }
                failure { script { stageCallback('compile', 'failed'); stageCallback('build_binary', 'failed') } }
            }
        }

        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 单元测试 ==="
                container('golang') {
                    script {
                        if (!fileExists('go.mod')) { echo "跳过测试"; return }
                        def hasTests = sh(script: "find . -name '*_test.go' | grep . >/dev/null 2>&1 && echo yes || echo no", returnStdout: true).trim()
                        if (hasTests != 'yes') { echo "无测试文件"; return }
                        sh '''
                            set -e
                            go test -v -coverprofile=coverage.out ./...
                            go tool cover -func=coverage.out | tail -1
                        '''
                    }
                }
            }
            post {
                success { script { stageCallback('test', 'success') } }
                failure { script { stageCallback('test', 'failed') } }
            }
        }

        stage('Lint') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 代码检查 ==="
                container('golang') {
                    script {
                        sh 'go vet ./...'
                    }
                }
            }
            post {
                success { script { stageCallback('lint', 'success') } }
                failure { script { stageCallback('lint', 'failed') } }
            }
        }

        // ==================== SonarQube 代码质量扫描 ====================
        stage('SonarQube Analysis') {
            when { expression { return params.ENABLE_SONAR } }
            steps {
                script {
                    try {
                        echo "=== SonarQube 代码质量扫描 ==="
                        def projectKey  = params.SONAR_PROJECT_KEY?.trim()  ?: env.JOB_NAME.replaceAll('/', '_')
                        def projectName = params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME
                        def sources     = params.SONAR_SOURCES?.trim()      ?: '.'
                        def exclusions  = params.SONAR_EXCLUSIONS?.trim()   ?: '**/vendor/**,**/*_test.go'

                        withSonarQubeEnv('SonarQube') {
                            container('golang') {
                                sh """
                                    # 下载 sonar-scanner（如果不存在）
                                    if ! command -v sonar-scanner &>/dev/null; then
                                        wget -q https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-5.0.1.3006-linux-x64.zip -O /tmp/sonar.zip || true
                                        if [ -f /tmp/sonar.zip ]; then
                                            unzip -qo /tmp/sonar.zip -d /tmp/ && export PATH="/tmp/sonar-scanner-5.0.1.3006-linux-x64/bin:\$PATH"
                                        fi
                                    fi
                                    sonar-scanner \
                                        -Dsonar.projectKey=${projectKey} \
                                        -Dsonar.projectName=${projectName} \
                                        -Dsonar.projectVersion=${env.FINAL_TAG} \
                                        -Dsonar.sources=${sources} \
                                        -Dsonar.exclusions=${exclusions},**/bin/**,**/build/** \
                                        -Dsonar.go.coverage.reportPaths=coverage.out \
                                        -Dsonar.scm.disabled=true \
                                        -Dsonar.qualitygate.wait=false \
                                        -Dsonar.threads=4 \
                                        -Dsonar.links.ci=${env.BUILD_URL}
                                """
                            }
                        }
                        echo "[SonarQube] 扫描已提交，等待质量门禁..."
                        stageCallback('sonar', 'success')
                    } catch (e) {
                        echo "[SonarQube] ❌ 扫描失败: ${e.message}"
                        stageCallback('sonar', 'failed')
                        env.SONAR_ANALYSIS_FAILED = 'true'
                        error("SonarQube 扫描失败: ${e.message}")
                    }
                }
            }
        }

        // ==================== SonarQube 质量门禁检查 ====================
        stage('Quality Gate') {
            when {
                allOf {
                    expression { return params.ENABLE_SONAR && params.SONAR_QUALITY_GATE }
                    expression { return env.SONAR_ANALYSIS_FAILED != 'true' }
                    expression { return (params.SONAR_GATE_ACTION ?: 'block') != 'skip' }
                }
            }
            steps {
                echo "=== 质量门禁检查（策略: ${params.SONAR_GATE_ACTION ?: 'block'}） ==="
                script {
                    def gateAction = params.SONAR_GATE_ACTION ?: 'block'
                    def qg = waitForQualityGate(webhookSecretId: '', abortPipeline: false)
                    env.SONAR_QUALITY_GATE_STATUS = qg.status

                    def metricsReport = checkPlatformThresholds()
                    env.SONAR_METRICS_REPORT = metricsReport

                    if (qg.status != 'OK') {
                        echo "[Quality Gate] ❌ SonarQube 门禁状态: ${qg.status}"
                        echo "[Quality Gate] 平台阈值检查:\n${metricsReport}"
                        sonarReportCallback(qg.status)

                        if (gateAction == 'block') {
                            error("SonarQube Quality Gate 未通过: ${qg.status}")
                        } else {
                            echo "[Quality Gate] ⚠️ 门禁未通过但策略为 warn，继续构建"
                            currentBuild.result = 'UNSTABLE'
                        }
                    } else {
                        echo "[Quality Gate] ✅ 通过！状态: ${qg.status}"
                        echo "[Quality Gate] 平台阈值检查:\n${metricsReport}"
                        sonarReportCallback(qg.status)
                    }
                }
            }
            post {
                success { script { stageCallback('quality_gate', 'success') } }
                failure { script { stageCallback('quality_gate', 'failed') } }
            }
        }

        // ==================== 制品归档 ====================
        stage('Upload Artifact') {
            when { expression { return params.ENABLE_ARTIFACT_UPLOAD && params.PLATFORM_CALLBACK_URL?.trim() } }
            steps {
                echo "=== 上传制品到平台制品库（gzip 压缩加速） ==="
                container('golang') {
                    script {
                        def binaryPath = env.BINARY_PATH ?: "bin/${env.APP_NAME ?: 'server'}"
                        if (!fileExists(binaryPath)) { error("[制品上传] 二进制文件不存在: ${binaryPath}") }

                        def gzPath = "${binaryPath}.gz"
                        sh "gzip -1 -c ${binaryPath} > ${gzPath}"
                        def origSize = sh(script: "stat -c%s ${binaryPath} 2>/dev/null || stat -f%z ${binaryPath}", returnStdout: true).trim()
                        def gzSize = sh(script: "stat -c%s ${gzPath} 2>/dev/null || stat -f%z ${gzPath}", returnStdout: true).trim()
                        echo "[制品上传] 原始: ${origSize} bytes → 压缩: ${gzSize} bytes"

                        def uploadUrl = params.PLATFORM_CALLBACK_URL
                            .replace('/pipeline/callback', '/artifact/upload')
                            .replace('/stage/callback', '/artifact/upload')

                        def curlStatus = sh(script: """
                            set -e
                            curl -s -w '%{http_code}' -o /tmp/artifact_resp.json \
                                -X POST '${uploadUrl}' \
                                -F 'file=@${gzPath}' \
                                -F 'pipeline_id=${params.PIPELINE_ID ?: 0}' \
                                -F 'run_id=${params.RUN_ID ?: 0}' \
                                -F 'build_number=${env.BUILD_NUMBER}' \
                                -F 'version=${env.FINAL_TAG}' \
                                -F 'language_type=go' \
                                -F 'artifact_type=binary' \
                                -F 'git_repo=${params.GIT_REPO}' \
                                -F 'git_branch=${env.GIT_BRANCH_NAME}' \
                                -F 'git_commit=${env.GIT_COMMIT_SHORT}' \
                                --connect-timeout 10 \
                                --max-time 300 \
                                --retry 2 --retry-delay 5
                        """, returnStdout: true).trim()

                        if (curlStatus.endsWith('200')) {
                            echo "[制品上传] ✅ 上传成功"
                        } else {
                            echo "[制品上传] ❌ 上传失败: HTTP ${curlStatus[-3..-1]}"
                            def respBody = sh(script: "cat /tmp/artifact_resp.json 2>/dev/null || echo '{}'", returnStdout: true).trim()
                            echo "[制品上传] 响应内容: ${respBody}"
                            error("制品上传失败: HTTP ${curlStatus[-3..-1]}")
                        }
                        sh "rm -f ${gzPath} /tmp/artifact_resp.json 2>/dev/null || true"
                    }
                }
            }
            post {
                success { script { stageCallback('upload_artifact', 'success') } }
                failure { script { stageCallback('upload_artifact', 'failed') } }
            }
        }

        // ==================== Kaniko 构建 + 推送镜像（合并为一步） ====================
        stage('Build & Push Image') {
            steps {
                echo "=== Kaniko 构建并推送镜像（无需 Docker daemon） ==="
                container('kaniko') {
                    script {
                        def appName = env.APP_NAME ?: (params.GIT_REPO?.tokenize('/')?.last()?.replace('.git', '') ?: 'server')
                        env.APP_NAME = appName
                        if (!env.BINARY_PATH) { env.BINARY_PATH = "bin/${appName}" }

                        def dockerfile = params.DOCKERFILE_PATH?.trim()

                        // 优先级：1) 参数指定路径 → 2) 项目自带 Dockerfile → 3) 自动生成
                        def forceGenerate = (dockerfile == '__PLATFORM_GENERATE__')
                        if (!dockerfile || forceGenerate) {
                            if (fileExists('Dockerfile')) {
                                dockerfile = 'Dockerfile'
                                echo "[Build Image] 检测到项目自带 Dockerfile，优先使用"
                            } else {
                                dockerfile = '.Dockerfile.runtime'
                                writeFile file: dockerfile, text: """\
FROM registry.cn-hangzhou.aliyuncs.com/k8s-gos/alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    addgroup -S app && adduser -S app -G app
WORKDIR /app
RUN mkdir -p /app/storage/logs /app/configs
COPY bin/${appName} /app/${appName}
RUN chmod +x /app/${appName} && chown -R app:app /app
USER app
ENV GIN_MODE=release
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz/live || exit 1
ENTRYPOINT ["/app/${appName}"]
"""
                                echo "[Build Image] ${forceGenerate ? '强制' : '项目无 Dockerfile，'}已自动生成纯运行时 Dockerfile"
                            }
                        }

                        // 配置镜像仓库认证
                        def registryHost = params.IMAGE_REPO.split('/')[0]
                        sh """
                            mkdir -p /kaniko/.docker
                            echo '{"auths":{"${registryHost}":{"username":"${REGISTRY_CREDS_USR}","password":"${REGISTRY_CREDS_PSW}"}}}' > /kaniko/.docker/config.json
                        """

                        // Kaniko 构建 + 推送（一步完成）
                        sh """
                            /kaniko/executor \
                                --context=. \
                                --dockerfile=${dockerfile} \
                                --destination=${env.FULL_IMAGE} \
                                --cache=true \
                                --cache-repo=${registryHost}/kaniko-cache/go \
                                --label git.commit=${env.GIT_COMMIT_FULL} \
                                --label git.branch=${env.GIT_BRANCH_NAME} \
                                --label build.number=${env.BUILD_NUMBER} \
                                --label build.timestamp=${env.BUILD_TS} \
                                --label build.mode=k8s-kaniko \
                                --snapshot-mode=redo \
                                --use-new-run
                        """

                        // Kaniko 完成后镜像已推送，获取 digest
                        env.IMAGE_DIGEST = ''
                        env.IMAGE_WITH_DIGEST = env.FULL_IMAGE
                        echo "[Build & Push] ✅ 镜像已构建并推送: ${env.FULL_IMAGE}"
                    }
                }
            }
            post {
                success { script { stageCallback('build', 'success'); stageCallback('push', 'success') } }
                failure { script { stageCallback('build', 'failed'); stageCallback('push', 'failed') } }
            }
        }

    }

    post {
        success {
            script {
                def msg
                if (!params.ENABLE_SONAR) {
                    msg = 'Go 项目构建成功'
                } else if (env.SONAR_ANALYSIS_FAILED == 'true') {
                    msg = "Go 项目构建失败 | SonarQube: UNAVAILABLE（扫描阶段连接失败，请检查 SonarQube 服务状态）"
                    callbackPlatform('FAILURE', msg)
                    return
                } else {
                    msg = "Go 项目构建成功 | SonarQube: ${env.SONAR_QUALITY_GATE_STATUS ?: 'SKIPPED'}"
                }
                callbackPlatform('SUCCESS', msg)
            }
        }
        failure { script { callbackPlatform('FAILURE', 'Go 项目构建失败') } }
        aborted { script { callbackPlatform('ABORTED', '构建中止') } }
    }
}

// ==================== 阶段级回调（与平台 StageCallbackRequest 对齐） ====================
// 获取回调地址（兼容首次构建 params 未注册的情况，fallback 到 JCasC 全局环境变量）
def getCallbackUrl() {
    return params.PLATFORM_CALLBACK_URL?.trim() ?: env.PLATFORM_CALLBACK_URL?.trim() ?: ''
}

def stageCallback(String stageType, String status) {
    def callbackUrl = getCallbackUrl()
    if (!callbackUrl) return
    try {
        def payload = [
            job_name     : env.JOB_NAME,
            build_number : env.BUILD_NUMBER as Integer,
            pipeline_id  : params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
            stage_type   : stageType,
            status       : status
        ]
        def body = groovy.json.JsonOutput.toJson(payload)
        def stageUrl = callbackUrl.replace('/pipeline/callback', '/stage/callback')
        def signature = ''
        if (env.HMAC_SECRET?.trim()) {
            signature = hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${stageType}")
        }
        def headers = signature ? [[name: 'X-Signature', value: signature]] : []
        httpRequest(url: stageUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON',
            requestBody: body, customHeaders: headers, validResponseCodes: '100:599', timeout: 10)
        echo "[阶段回调] ${stageType} -> ${status}"
    } catch (e) { echo "[阶段回调] 非致命: ${e.message}" }
}

// ==================== 最终回调（与平台 PipelineCallbackRequest 对齐） ====================
def callbackPlatform(String status, String message) {
    def callbackUrl = getCallbackUrl()
    if (!callbackUrl) { echo "未配置回调地址"; return }
    def payload = [
        job_name          : env.JOB_NAME,
        build_number      : env.BUILD_NUMBER as Integer,
        status            : status,
        pipeline_id       : params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
        run_id            : params.RUN_ID ? params.RUN_ID as Long : 0,
        image_url         : env.FULL_IMAGE ?: '',
        image_digest      : env.IMAGE_DIGEST ?: '',
        image_with_digest : env.IMAGE_WITH_DIGEST ?: '',
        git_commit        : env.GIT_COMMIT_SHORT ?: '',
        git_branch        : env.GIT_BRANCH_NAME ?: '',
        duration_sec      : currentBuild.duration ? (currentBuild.duration / 1000) as Integer : 0,
        message           : message,
        build_url         : env.BUILD_URL ?: ''
    ]
    def body = groovy.json.JsonOutput.toJson(payload)
    def signature = ''
    if (env.HMAC_SECRET?.trim()) {
        signature = hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${status}")
    }
    def headers = signature ? [[name: 'X-Signature', value: signature]] : []
    httpRequest(url: callbackUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON',
        requestBody: body, customHeaders: headers, validResponseCodes: '200:299', consoleLogResponseBody: true)
}

// ==================== SonarQube 指标回传平台 ====================
def sonarReportCallback(String qualityGateStatus) {
    def callbackUrl = getCallbackUrl()
    if (!callbackUrl) return
    try {
        def projectKey = params.SONAR_PROJECT_KEY?.trim() ?: env.JOB_NAME.replaceAll('/', '_')
        def sonarUrl = callbackUrl.replace('/pipeline/callback', '/pipeline/sonar-callback')

        def metrics = [:]
        withSonarQubeEnv('SonarQube') {
            def apiUrl = "${env.SONAR_HOST_URL}/api/measures/component?component=${projectKey}&metricKeys=bugs,vulnerabilities,code_smells,coverage,duplicated_lines_density,ncloc,security_hotspots,reliability_rating,security_rating,sqale_rating"
            def resp = httpRequest(url: apiUrl, httpMode: 'GET', validResponseCodes: '200', quiet: true)
            def json = readJSON text: resp.content
            json.component?.measures?.each { m -> metrics[m.metric] = m.value }
        }

        def payload = [
            pipeline_id:            params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
            project_key:            projectKey,
            project_name:           params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME,
            quality_gate:           qualityGateStatus,
            dashboard_url:          "${env.SONAR_HOST_URL}/dashboard?id=${projectKey}",
            bugs:                   (metrics.bugs ?: '0') as Integer,
            vulnerabilities:        (metrics.vulnerabilities ?: '0') as Integer,
            code_smells:            (metrics.code_smells ?: '0') as Integer,
            coverage:               (metrics.coverage ?: '0.0') as Double,
            duplications:           (metrics['duplicated_lines_density'] ?: '0.0') as Double,
            lines_of_code:          (metrics.ncloc ?: '0') as Integer,
            security_hotspots:      (metrics.security_hotspots ?: '0') as Integer,
            reliability_rating:     ratingToLetter((metrics.reliability_rating ?: '1') as Double),
            security_rating:        ratingToLetter((metrics.security_rating ?: '1') as Double),
            maintainability_rating: ratingToLetter((metrics.sqale_rating ?: '1') as Double)
        ]

        def body = groovy.json.JsonOutput.toJson(payload)
        def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:sonar") : ''
        def headers = signature ? [[name: 'X-Signature', value: signature]] : []
        httpRequest(url: sonarUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON',
            requestBody: body, customHeaders: headers, validResponseCodes: '100:599', timeout: 15)
        echo "[SonarQube] 指标数据已回传平台"
    } catch (e) {
        echo "[SonarQube] 指标回传非致命错误: ${e.message}"
    }
}

// ==================== 平台阈值检查 ====================
def checkPlatformThresholds() {
    def report = []
    try {
        def projectKey = params.SONAR_PROJECT_KEY?.trim() ?: env.JOB_NAME.replaceAll('/', '_')
        def metrics = [:]
        withSonarQubeEnv('SonarQube') {
            def apiUrl = "${env.SONAR_HOST_URL}/api/measures/component?component=${projectKey}&metricKeys=coverage,bugs,code_smells,vulnerabilities,duplicated_lines_density"
            def resp = httpRequest(url: apiUrl, httpMode: 'GET', validResponseCodes: '200', quiet: true)
            def json = readJSON text: resp.content
            json.component?.measures?.each { m -> metrics[m.metric] = m.value }
        }
        def actualCoverage = (metrics.coverage ?: '0') as Double
        def thresholdCoverage = (params.SONAR_COVERAGE_THRESHOLD ?: '80') as Double
        report << "  覆盖率: ${actualCoverage}% (${actualCoverage >= thresholdCoverage ? '✅' : '❌'} 阈值: ≥${thresholdCoverage}%)"
        def actualBugs = (metrics.bugs ?: '0') as Integer
        def maxBugs = (params.SONAR_NEW_BUGS_MAX ?: '0') as Integer
        report << "  Bug: ${actualBugs} (${actualBugs <= maxBugs ? '✅' : '❌'} 阈值: ≤${maxBugs})"
        def actualSmells = (metrics.code_smells ?: '0') as Integer
        def maxSmells = (params.SONAR_CODE_SMELLS_MAX ?: '10') as Integer
        report << "  异味: ${actualSmells} (${actualSmells <= maxSmells ? '✅' : '❌'} 阈值: ≤${maxSmells})"
        def actualVulns = (metrics.vulnerabilities ?: '0') as Integer
        def maxVulns = (params.SONAR_VULNERABILITIES_MAX ?: '0') as Integer
        report << "  漏洞: ${actualVulns} (${actualVulns <= maxVulns ? '✅' : '❌'} 阈值: ≤${maxVulns})"
        def actualDup = (metrics['duplicated_lines_density'] ?: '0') as Double
        def maxDup = (params.SONAR_DUPLICATIONS_MAX ?: '3') as Double
        report << "  重复率: ${actualDup}% (${actualDup <= maxDup ? '✅' : '❌'} 阈值: ≤${maxDup}%)"
        def allPass = (actualCoverage >= thresholdCoverage) && (actualBugs <= maxBugs) && (actualSmells <= maxSmells) && (actualVulns <= maxVulns) && (actualDup <= maxDup)
        report.add(0, allPass ? "✅ 平台阈值全部通过" : "❌ 部分指标未达平台阈值")
    } catch (e) {
        report << "  ⚠️ 指标查询失败: ${e.message}"
    }
    return report.join('\n')
}

def ratingToLetter(Double rating) {
    if (rating <= 1.0) return 'A'
    if (rating <= 2.0) return 'B'
    if (rating <= 3.0) return 'C'
    if (rating <= 4.0) return 'D'
    return 'E'
}

// ==================== HMAC-SHA256 ====================
def hmacSha256(String secret, String data) {
    def result = ''
    withEnv(["SIGN_SECRET=${secret}", "SIGN_DATA=${data}"]) {
        result = sh(script: 'set +x && printf "%s" "$SIGN_DATA" | openssl dgst -sha256 -hmac "$SIGN_SECRET" | awk \'{print $2}\'', returnStdout: true).trim()
    }
    return result
}
