// ==============================================================================
// K8s Operation Platform - Go 项目通用构建模板
// ==============================================================================
// 设计理念：一个模板服务 100+ Go 项目，所有项目差异通过参数传入
// 回调协议：与平台后端 StageCallbackRequest / PipelineCallbackRequest 完全对齐
//
// ======================== Jenkins Job 配置方式 ========================
// 推荐使用 "Pipeline script from SCM"（版本化管理，自动同步更新）：
//   1. Jenkins → New Item → Pipeline → 命名为 k8s-builder-go
//   2. Pipeline → Definition: Pipeline script from SCM
//   3. SCM: Git → Repository URL: 平台仓库地址
//   4. Script Path: configs/jenkins-templates/go-pipeline.groovy
// ==============================================================================

pipeline {
    agent any

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

        // SonarQube 代码质量扫描参数
        booleanParam(name: 'ENABLE_SONAR', defaultValue: false, description: '启用 SonarQube 代码质量扫描')
        string(name: 'SONAR_PROJECT_KEY', defaultValue: '', description: 'SonarQube 项目 Key（空则使用 Job 名称）')
        string(name: 'SONAR_PROJECT_NAME', defaultValue: '', description: 'SonarQube 项目名称（空则使用 Job 名称）')
        string(name: 'SONAR_SOURCES', defaultValue: '.', description: '源代码目录')
        string(name: 'SONAR_EXCLUSIONS', defaultValue: '**/vendor/**,**/*_test.go,**/test/**', description: '排除扫描的文件模式')
        booleanParam(name: 'SONAR_QUALITY_GATE', defaultValue: true, description: '启用质量门禁检查（不通过则构建失败）')

        // 制品上传参数
        booleanParam(name: 'ENABLE_ARTIFACT_UPLOAD', defaultValue: false, description: '启用制品上传到平台制品库')
    }

    environment {
        GOROOT     = "/usr/local/go"
        GOPATH     = "/var/lib/jenkins/go"
        GOMODCACHE = "/var/lib/jenkins/go/pkg/mod"
        GOCACHE    = "/var/lib/jenkins/.cache/go-build"
        PATH       = "/usr/local/go/bin:${env.GOPATH}/bin:${env.PATH}"

        REGISTRY_CREDS = credentials('harbor-registry')
        HMAC_SECRET    = credentials('hmac-secret')
        GOPROXY        = 'https://goproxy.cn,direct'
        CGO_ENABLED    = '0'
        GOOS           = 'linux'
        GOARCH         = 'amd64'
        // BuildKit 层缓存目录（跨构建持久复用，二次构建仅重建变化层）
        BUILDKIT_CACHE = '/var/lib/jenkins/.buildkit-cache'
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

                    checkout([
                        $class: 'GitSCM',
                        branches: [[name: "*/${targetBranch}"]],
                        extensions: [
                            [$class: 'CleanBeforeCheckout'],
                            [$class: 'LocalBranch', localBranch: targetBranch],
                            [$class: 'CloneOption', depth: 1, shallow: true, noTags: true, timeout: 10]
                        ],
                        userRemoteConfigs: [[
                            url: params.GIT_REPO,
                            credentialsId: params.GIT_CREDENTIAL_ID ?: 'gitee-id'
                        ]]
                    ])

                    env.TARGET_BRANCH = targetBranch
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
                        : "${env.GIT_BRANCH_NAME}-${env.GIT_COMMIT_SHORT}-${env.BUILD_TS}"

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
            post {
                success { script { stageCallback('dependencies', 'success') } }
                failure { script { stageCallback('dependencies', 'failed') } }
            }
        }

        stage('Compile Check') {
            steps {
                echo "=== 编译检查（直接产出最终二进制，Build Image 复用，避免重复编译） ==="
                script {
                    if (!fileExists('go.mod')) { echo "跳过编译检查"; return }
                    def appName = params.GIT_REPO?.split('/')?.getAt(-1)?.replace('.git', '') ?: 'server'
                    env.APP_NAME = appName
                    env.BINARY_PATH = "bin/${appName}"
                    sh """
                        set -e
                        mkdir -p bin
                        go build -ldflags="-s -w -X main.Version=${env.FINAL_TAG} -X main.GitCommit=${env.GIT_COMMIT_FULL}" -o ${env.BINARY_PATH} ./cmd/... || \\
                        go build -ldflags="-s -w -X main.Version=${env.FINAL_TAG} -X main.GitCommit=${env.GIT_COMMIT_FULL}" -o ${env.BINARY_PATH} .
                    """
                    echo "[编译] 二进制产物: ${env.BINARY_PATH}"
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
            post {
                success { script { stageCallback('test', 'success') } }
                failure { script { stageCallback('test', 'failed') } }
            }
        }

        stage('Lint') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 代码检查 ==="
                script {
                    def hasLint = sh(script: "which golangci-lint >/dev/null 2>&1 && echo yes || echo no", returnStdout: true).trim()
                    if (hasLint == 'yes') { sh 'golangci-lint run --timeout=5m || true' }
                    else { sh 'go vet ./...' }
                }
            }
            post {
                success { script { stageCallback('lint', 'success') } }
                failure { script { stageCallback('lint', 'failed') } }
            }
        }

        // ==================== SonarQube 代码质量扫描（性能优化版） ====================
        // 优化要点：
        //   1. -Dsonar.scm.disabled=true       → 禁用 git blame（节省 60%+ 时间）
        //   2. -Dsonar.qualitygate.wait=false  → 扫描阶段不阻塞，由 Quality Gate 阶段异步等待
        //   3. -Dsonar.threads=4               → 启用 4 线程并行分析（多核 CPU 加速 30-50%）
        //   4. exclusions 追加 bin/build       → 避免扫描编译产物目录
        stage('SonarQube Analysis') {
            when { expression { return params.ENABLE_SONAR } }
            steps {
                echo "=== SonarQube 代码质量扫描（轻量模式） ==="
                script {
                    def projectKey  = params.SONAR_PROJECT_KEY?.trim()  ?: env.JOB_NAME.replaceAll('/', '_')
                    def projectName = params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME
                    def sources     = params.SONAR_SOURCES?.trim()      ?: '.'
                    def exclusions  = params.SONAR_EXCLUSIONS?.trim()   ?: '**/vendor/**,**/*_test.go'

                    // 使用 SonarQube Scanner CLI（Go 项目不用 Maven）
                    withSonarQubeEnv('SonarQube') {
                        sh """
                            sonar-scanner \\
                                -Dsonar.projectKey=${projectKey} \\
                                -Dsonar.projectName=${projectName} \\
                                -Dsonar.projectVersion=${env.FINAL_TAG} \\
                                -Dsonar.sources=${sources} \\
                                -Dsonar.exclusions=${exclusions},**/bin/**,**/build/** \\
                                -Dsonar.go.coverage.reportPaths=coverage.out \\
                                -Dsonar.scm.disabled=true \\
                                -Dsonar.qualitygate.wait=false \\
                                -Dsonar.threads=4 \\
                                -Dsonar.links.ci=${env.BUILD_URL}
                        """
                    }
                    echo "[SonarQube] 扫描已提交，等待质量门禁..."
                }
            }
            post {
                success { script { stageCallback('sonar', 'success') } }
                failure { script { stageCallback('sonar', 'failed') } }
            }
        }

        // ==================== SonarQube 质量门禁检查 ====================
        stage('Quality Gate') {
            when { expression { return params.ENABLE_SONAR && params.SONAR_QUALITY_GATE } }
            steps {
                echo "=== 质量门禁检查 ==="
                script {
                    // webhookSecretId: '' + abortPipeline: false → 与 Java 模板对齐，由脚本控制失败行为
                    def qg = waitForQualityGate(webhookSecretId: '', abortPipeline: false)
                    env.SONAR_QUALITY_GATE_STATUS = qg.status
                    if (qg.status != 'OK') {
                        echo "[Quality Gate] 状态: ${qg.status}"
                        echo "[Quality Gate] 代码质量未达标，请登录 SonarQube 查看详细报告"
                        sonarReportCallback(qg.status)
                        error("SonarQube Quality Gate 未通过: ${qg.status}")
                    }
                    echo "[Quality Gate] ✅ 通过！状态: ${qg.status}"
                    sonarReportCallback(qg.status)
                }
            }
            post {
                success { script { stageCallback('quality_gate', 'success') } }
                failure { script { stageCallback('quality_gate', 'failed') } }
            }
        }

        // ==================== 制品上传（可选，gzip 压缩加速传输） ====================
        stage('Upload Artifact') {
            when { expression { return params.ENABLE_ARTIFACT_UPLOAD && params.PLATFORM_CALLBACK_URL?.trim() } }
            steps {
                echo "=== 上传制品到平台制品库（gzip 压缩加速） ==="
                script {
                    def binaryPath = env.BINARY_PATH ?: "bin/${env.APP_NAME ?: 'server'}"
                    if (!fileExists(binaryPath)) { echo "[制品上传] 二进制文件不存在: ${binaryPath}，跳过"; return }
        
                    // gzip 压缩二进制（Go 二进制压缩率 60-70%，大幅减少传输时间）
                    def gzPath = "${binaryPath}.gz"
                    sh "gzip -1 -c ${binaryPath} > ${gzPath}"
                    def origSize = sh(script: "stat -c%s ${binaryPath} 2>/dev/null || stat -f%z ${binaryPath}", returnStdout: true).trim()
                    def gzSize = sh(script: "stat -c%s ${gzPath} 2>/dev/null || stat -f%z ${gzPath}", returnStdout: true).trim()
                    echo "[制品上传] 原始: ${origSize} bytes → 压缩: ${gzSize} bytes"
        
                    // 构造上传 URL：从回调地址推导制品上传接口
                    def uploadUrl = params.PLATFORM_CALLBACK_URL
                        .replace('/pipeline/callback', '/artifact/upload')
                        .replace('/stage/callback', '/artifact/upload')
        
                    // curl 上传（multipart/form-data，携带构建元数据）
                    def curlStatus = sh(script: """
                        set -e
                        curl -s -w '%{http_code}' -o /tmp/artifact_resp.json \\
                            -X POST '${uploadUrl}' \\
                            -F 'file=@${gzPath}' \\
                            -F 'pipeline_id=${params.PIPELINE_ID ?: 0}' \\
                            -F 'run_id=${params.RUN_ID ?: 0}' \\
                            -F 'build_number=${env.BUILD_NUMBER}' \\
                            -F 'version=${env.FINAL_TAG}' \\
                            -F 'language_type=go' \\
                            -F 'artifact_type=binary' \\
                            -F 'git_repo=${params.GIT_REPO}' \\
                            -F 'git_branch=${env.GIT_BRANCH_NAME}' \\
                            -F 'git_commit=${env.GIT_COMMIT_SHORT}' \\
                            --connect-timeout 10 \\
                            --max-time 120 \\
                            --tcp-nodelay \\
                            -H "Expect:" \\
                            --retry 1
                    """, returnStdout: true).trim()
        
                    if (curlStatus.endsWith('200')) {
                        echo "[制品上传] 上传成功"
                    } else {
                        echo "[制品上传] 上传返回: HTTP ${curlStatus}（非致命，不影响构建）"
                    }
                    sh "rm -f ${gzPath} /tmp/artifact_resp.json 2>/dev/null || true"
                }
            }
            post {
                success { script { stageCallback('upload_artifact', 'success') } }
                failure { script { stageCallback('upload_artifact', 'failed') } }
            }
        }
        
        // ==================== Docker 镜像构建（复用 Compile Check 产出的二进制） ====================
        
        stage('Build Image') {
            steps {
                echo "=== 构建 Docker 镜像（复用 Compile Check 产出的二进制） ==="
                script {
                    def appName = env.APP_NAME ?: (params.GIT_REPO?.split('/')?.getAt(-1)?.replace('.git', '') ?: 'server')
                    env.APP_NAME = appName
                    if (!env.BINARY_PATH) { env.BINARY_PATH = "bin/${appName}" }
        
                    def dockerfile = params.DOCKERFILE_PATH?.trim()

                    // 优先级：1) 参数指定路径 → 2) 项目自带 Dockerfile → 3) 自动生成
                    // __PLATFORM_GENERATE__ 为平台哨兵值，表示强制使用平台生成
                    def forceGenerate = (dockerfile == '__PLATFORM_GENERATE__')
                    if (!dockerfile || forceGenerate) {
                        if (!forceGenerate && fileExists('Dockerfile')) {
                            dockerfile = 'Dockerfile'
                            echo "[Build Image] 检测到项目自带 Dockerfile，直接使用"
                        } else {
                            dockerfile = '.Dockerfile.runtime'
                            writeFile file: dockerfile, text: """\
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && \\
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \\
    addgroup -S app && adduser -S app -G app
WORKDIR /app
RUN mkdir -p /app/storage/logs /app/configs
COPY bin/${appName} /app/${appName}
RUN chmod +x /app/${appName} && chown -R app:app /app
USER app
ENV GIN_MODE=release
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \\
    CMD wget -qO- http://127.0.0.1:8080/healthz/live || exit 1
ENTRYPOINT ["/app/${appName}"]
"""
                            echo "[Build Image] ${forceGenerate ? '强制' : '项目无 Dockerfile，'}已自动生成纯运行时 Dockerfile"
                        }
                    }

                    // 使用 BuildKit 本地层缓存：首次全量构建，后续仅重建变化层
                    def cacheDir = "${env.BUILDKIT_CACHE}/${env.JOB_NAME}".replaceAll('[^a-zA-Z0-9/_.-]', '_')
                    sh """
                        set -e
                        mkdir -p ${cacheDir}
                        nerdctl build \\
                            -t ${env.FULL_IMAGE} \\
                            -f ${dockerfile} \\
                            --cache-from type=local,src=${cacheDir} \\
                            --cache-to type=local,dest=${cacheDir},mode=max \\
                            --label git.commit=${env.GIT_COMMIT_FULL} \\
                            --label git.branch=${env.GIT_BRANCH_NAME} \\
                            --label build.number=${env.BUILD_NUMBER} \\
                            --label build.timestamp=${env.BUILD_TS} \\
                            --label build.mode=platform-compile \\
                            .
                    """
                }
            }
            post {
                success { script { stageCallback('build', 'success') } }
                failure { script { stageCallback('build', 'failed') } }
            }
        }

        stage('Push Image') {
            steps {
                echo "=== 推送镜像 ==="
                script {
                    def registryHost = params.IMAGE_REPO.split('/')[0]
                    sh """
                        set -e
                        echo \${REGISTRY_CREDS_PSW} | nerdctl login -u \${REGISTRY_CREDS_USR} --password-stdin ${registryHost}
                        nerdctl push ${env.FULL_IMAGE}
                    """
                    env.IMAGE_DIGEST = sh(
                        script: "nerdctl inspect ${env.FULL_IMAGE} --format '{{range .RepoDigests}}{{println .}}{{end}}' 2>/dev/null | grep -oE 'sha256:[a-f0-9]+' | head -1 || echo ''",
                        returnStdout: true
                    ).trim()
                    env.IMAGE_WITH_DIGEST = env.IMAGE_DIGEST
                        ? "${params.IMAGE_REPO}@${env.IMAGE_DIGEST}"
                        : env.FULL_IMAGE
                    echo "Digest: ${env.IMAGE_DIGEST ?: '未获取到'}"
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
                def msg = params.ENABLE_SONAR
                    ? "Go 项目构建成功 | SonarQube: ${env.SONAR_QUALITY_GATE_STATUS ?: 'SKIPPED'}"
                    : 'Go 项目构建成功'
                callbackPlatform('SUCCESS', msg)
            }
        }
        failure { script { callbackPlatform('FAILURE', 'Go 项目构建失败') } }
        aborted { script { callbackPlatform('ABORTED', '构建中止') } }
        always {
            sh "nerdctl rmi ${env.FULL_IMAGE} || true"
            // Go 缓存已在 workspace 外（GOMODCACHE/GOCACHE），只清源码
            sh 'rm -rf bin .git coverage.out 2>/dev/null || true'
        }
    }
}

// ==================== 阶段级回调（与平台 StageCallbackRequest 对齐） ====================
def stageCallback(String stageType, String status) {
    if (!params.PLATFORM_CALLBACK_URL?.trim()) return
    try {
        def payload = [
            job_name     : env.JOB_NAME,
            build_number : env.BUILD_NUMBER as Integer,
            pipeline_id  : params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
            stage_type   : stageType,
            status       : status
        ]
        def body = groovy.json.JsonOutput.toJson(payload)
        def stageUrl = params.PLATFORM_CALLBACK_URL.replace('/pipeline/callback', '/stage/callback')
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
    if (!params.PLATFORM_CALLBACK_URL?.trim()) { echo "未配置回调地址"; return }
    def payload = [
        job_name          : env.JOB_NAME,
        build_number      : env.BUILD_NUMBER as Integer,
        status            : status,
        pipeline_id       : params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
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
    httpRequest(url: params.PLATFORM_CALLBACK_URL, httpMode: 'POST', contentType: 'APPLICATION_JSON',
        requestBody: body, customHeaders: headers, validResponseCodes: '200:299', consoleLogResponseBody: true)
}

// ==================== SonarQube 指标回传平台 ====================
def sonarReportCallback(String qualityGateStatus) {
    if (!params.PLATFORM_CALLBACK_URL?.trim()) return
    try {
        def projectKey = params.SONAR_PROJECT_KEY?.trim() ?: env.JOB_NAME.replaceAll('/', '_')
        def sonarUrl = params.PLATFORM_CALLBACK_URL.replace('/pipeline/callback', '/pipeline/sonar-callback')

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

// SonarQube rating 数值转字母: 1.0=A, 2.0=B, 3.0=C, 4.0=D, 5.0=E
def ratingToLetter(Double rating) {
    if (rating <= 1.0) return 'A'
    if (rating <= 2.0) return 'B'
    if (rating <= 3.0) return 'C'
    if (rating <= 4.0) return 'D'
    return 'E'
}

// ==================== HMAC-SHA256（openssl 版，避免 Sandbox 拦截） ====================
def hmacSha256(String secret, String data) {
    def result = ''
    withEnv(["SIGN_SECRET=${secret}", "SIGN_DATA=${data}"]) {
        result = sh(script: 'set +x && printf "%s" "$SIGN_DATA" | openssl dgst -sha256 -hmac "$SIGN_SECRET" | awk \'{print $2}\'', returnStdout: true).trim()
    }
    return result
}
