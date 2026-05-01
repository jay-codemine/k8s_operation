// ==============================================================================
// K8s Operation Platform - Python 项目通用构建模板
// ==============================================================================
// 设计理念：一个模板服务 100+ Python 项目，所有项目差异通过参数传入
// 支持框架：Flask, FastAPI, Django, Celery 等
//
// ======================== Jenkins Job 配置方式 ========================
// 推荐使用 "Pipeline script from SCM"（版本化管理，自动同步更新）：
//   1. Jenkins → New Item → Pipeline → 命名为 k8s-builder-python
//   2. Pipeline → Definition: Pipeline script from SCM
//   3. SCM: Git → Repository URL: 平台仓库地址
//   4. Script Path: configs/jenkins-templates/python-pipeline.groovy
// ==============================================================================

pipeline {
    agent any

    options {
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '20'))
        skipDefaultCheckout(true)
    }

    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址（必填）')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库地址（必填）')
        string(name: 'IMAGE_TAG', defaultValue: '', description: '镜像标签（空则自动生成）')
        string(name: 'DOCKERFILE_PATH', defaultValue: '', description: 'Dockerfile 路径（空则自动生成纯运行时 Dockerfile）')
        string(name: 'LANGUAGE_TYPE', defaultValue: '', description: '平台注入的语言类型（用于交叉校验，不要手动修改）')

        string(name: 'PIPELINE_ID', defaultValue: '', description: '平台流水线ID')
        string(name: 'RUN_ID', defaultValue: '', description: '平台运行记录ID')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')

        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过测试')
        string(name: 'PYTHON_VERSION', defaultValue: '3.11', description: 'Python 版本')
        string(name: 'GIT_CREDENTIAL_ID', defaultValue: 'gitee-id', description: 'Git 凭证ID')

        // SonarQube 代码质量扫描参数
        booleanParam(name: 'ENABLE_SONAR', defaultValue: false, description: '启用 SonarQube 代码质量扫描')
        string(name: 'SONAR_PROJECT_KEY', defaultValue: '', description: 'SonarQube 项目 Key（空则使用 Job 名称）')
        string(name: 'SONAR_PROJECT_NAME', defaultValue: '', description: 'SonarQube 项目名称（空则使用 Job 名称）')
        string(name: 'SONAR_SOURCES', defaultValue: '.', description: '源代码目录')
        string(name: 'SONAR_EXCLUSIONS', defaultValue: '**/venv/**,**/__pycache__/**,**/test_*,**/*_test.py,**/migrations/**', description: '排除扫描的文件模式')
        booleanParam(name: 'SONAR_QUALITY_GATE', defaultValue: true, description: '启用质量门禁检查（不通过则构建失败）')

        // 制品上传参数
        booleanParam(name: 'ENABLE_ARTIFACT_UPLOAD', defaultValue: true, description: '启用制品上传到平台制品库')
    }

    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
        HMAC_SECRET    = credentials('hmac-secret')
        PIP_INDEX_URL  = 'https://pypi.tuna.tsinghua.edu.cn/simple'
        // BuildKit 层缓存目录（跨构建持久复用，二次构建仅重建变化层）
        BUILDKIT_CACHE = '/var/lib/jenkins/.buildkit-cache'
    }

    stages {

        stage('Clean Workspace') {
            steps {
                deleteDir()
                // 强制清除 .git（防止 deleteDir 未完全清理导致浅克隆残留）
                sh 'rm -rf .git 2>/dev/null || true'
                script {
                    // 语言类型交叉校验：防止自定义 Job 配错脚本
                    def expectedType = 'python'
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
当前模板类型: ${expectedType} (python-pipeline.groovy)

解决方案（二选一）:
  1. 修改 Jenkins Job 的 Script Path 为: configs/jenkins-templates/${correctScript}
  2. 在平台将 Jenkins Job 名称留空，使用自动匹配
""")
                    }

                    if (!params.GIT_REPO?.trim()) { error("GIT_REPO 不能为空") }
                    if (!params.IMAGE_REPO?.trim()) { error("IMAGE_REPO 不能为空") }

                    def targetBranch = params.GIT_BRANCH?.trim() ?: 'main'

                    // 强制清除 .git（防止浅克隆残留导致 fetch 拉不到最新代码）
                    sh 'rm -rf .git 2>/dev/null || true'

                    checkout([
                        $class: 'GitSCM',
                        branches: [[name: "*/${targetBranch}"]],
                        extensions: [
                            [$class: 'CleanBeforeCheckout', deleteUntrackedNestedRepositories: true],
                            [$class: 'LocalBranch', localBranch: targetBranch],
                            [$class: 'CloneOption', depth: 1, shallow: true, noTags: true, timeout: 10, honorRefspec: true]
                        ],
                        userRemoteConfigs: [[url: params.GIT_REPO, credentialsId: params.GIT_CREDENTIAL_ID ?: 'gitee-id']]
                    ])
                    env.TARGET_BRANCH = targetBranch

                    // 验证拉取的是最新代码
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
                    env.GIT_BRANCH_NAME  = (env.TARGET_BRANCH ?: 'main').replaceAll('/', '-')
                    env.BUILD_TS = sh(script: 'date +%Y%m%d%H%M%S', returnStdout: true).trim()
                    env.FINAL_TAG = params.IMAGE_TAG?.trim() ?: "${env.GIT_BRANCH_NAME}-${env.GIT_COMMIT_SHORT}-${env.BUILD_TS}"
                    env.FULL_IMAGE = "${params.IMAGE_REPO}:${env.FINAL_TAG}"
                    echo "Commit: ${env.GIT_COMMIT_SHORT} | Image: ${env.FULL_IMAGE}"
                }
            }
            post {
                success { script { stageCallback('checkout', 'success') } }
                failure { script { stageCallback('checkout', 'failed') } }
            }
        }

        stage('Install Dependencies') {
            steps {
                echo "=== 安装 Python 依赖 ==="
                script {
                    if (fileExists('requirements.txt')) {
                        sh 'pip install -r requirements.txt -q --index-url ${PIP_INDEX_URL}'
                    } else if (fileExists('setup.py') || fileExists('pyproject.toml')) {
                        sh 'pip install -e . -q --index-url ${PIP_INDEX_URL}'
                    } else {
                        echo "未检测到依赖文件，跳过"
                    }
                }
            }
            post {
                success { script { stageCallback('dependencies', 'success') } }
                failure { script { stageCallback('dependencies', 'failed') } }
            }
        }

        stage('Lint') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 代码检查 ==="
                sh '''
                    pip install flake8 -q 2>/dev/null || true
                    flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics 2>/dev/null || true
                '''
            }
            post {
                success { script { stageCallback('lint', 'success') } }
                failure { script { stageCallback('lint', 'failed') } }
            }
        }

        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 单元测试 ==="
                script {
                    def hasTests = sh(script: "find . -name 'test_*.py' -o -name '*_test.py' | grep . >/dev/null 2>&1 && echo yes || echo no", returnStdout: true).trim()
                    if (hasTests == 'yes') {
                        sh '''
                            pip install pytest pytest-cov -q 2>/dev/null || true
                            pytest --cov=. tests/ -v --cov-report=xml:coverage.xml 2>/dev/null || pytest -v 2>/dev/null || true
                        '''
                    } else {
                        echo "未检测到测试文件，跳过"
                    }
                }
            }
            post {
                success { script { stageCallback('test', 'success'); stageCallback('build_binary', 'success') } }
                failure { script { stageCallback('test', 'failed'); stageCallback('build_binary', 'failed') } }
            }
        }

        // ==================== SonarQube 代码质量扫描（性能优化版） ====================
        // 优化要点：
        //   1. -Dsonar.scm.disabled=true       → 禁用 git blame（节省 60%+ 时间）
        //   2. -Dsonar.qualitygate.wait=false  → 扫描阶段不阻塞，由 Quality Gate 阶段异步等待
        //   3. -Dsonar.threads=4               → 启用 4 线程并行分析（多核 CPU 加速 30-50%）
        //   4. exclusions 追加 build/dist/.pytest_cache/.tox → 避免扫描编译产物目录
        stage('SonarQube Analysis') {
            when { expression { return params.ENABLE_SONAR } }
            steps {
                script {
                    try {
                        echo "=== SonarQube 代码质量扫描（轻量模式） ==="
                        def projectKey  = params.SONAR_PROJECT_KEY?.trim()  ?: env.JOB_NAME.replaceAll('/', '_')
                        def projectName = params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME
                        def sources     = params.SONAR_SOURCES?.trim()      ?: '.'
                        def exclusions  = params.SONAR_EXCLUSIONS?.trim()   ?: '**/venv/**,**/__pycache__/**'

                        withSonarQubeEnv('SonarQube') {
                            sh """
                                sonar-scanner \\
                                    -Dsonar.projectKey=${projectKey} \\
                                    -Dsonar.projectName=${projectName} \\
                                    -Dsonar.projectVersion=${env.FINAL_TAG} \\
                                    -Dsonar.sources=${sources} \\
                                    -Dsonar.exclusions=${exclusions},**/build/**,**/dist/**,**/.pytest_cache/**,**/.tox/** \\
                                    -Dsonar.python.coverage.reportPaths=coverage.xml \\
                                    -Dsonar.scm.disabled=true \\
                                    -Dsonar.qualitygate.wait=false \\
                                    -Dsonar.threads=4 \\
                                    -Dsonar.links.ci=${env.BUILD_URL}
                            """
                        }
                        echo "[SonarQube] 扫描已提交，等待质量门禁..."
                        stageCallback('sonar', 'success')
                    } catch (e) {
                        echo "[SonarQube] ❌ 扫描失败: ${e.message}"
                        echo "[SonarQube] 常见原因: 1) SonarQube 服务未启动  2) Jenkins 与 SonarQube 网络不通  3) SonarQube Token 过期"
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
                }
            }
            steps {
                echo "=== 质量门禁检查 ==="
                script {
                    // webhookSecretId: '' + abortPipeline: false → 与 Java 模板对齐，由脚本控制失败行为
                    def qg = waitForQualityGate(webhookSecretId: '', abortPipeline: false)
                    env.SONAR_QUALITY_GATE_STATUS = qg.status
                    if (qg.status != 'OK') {
                        echo "[Quality Gate] 状态: ${qg.status}"
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

        // ==================== 制品归档（Quality Gate 之后、Build Image 之前，失败即终止流水线） ====================
        stage('Upload Artifact') {
            when { expression { return params.ENABLE_ARTIFACT_UPLOAD && params.PLATFORM_CALLBACK_URL?.trim() } }
            steps {
                echo "=== 上传制品到平台制品库（tar.gz 打包加速） ==="
                script {
                        def appName = params.GIT_REPO?.split('/')?.getAt(-1)?.replace('.git', '') ?: 'python-app'
                        def archiveName = "${appName}-${env.FINAL_TAG}.tar.gz"
                        sh "tar czf ${archiveName} --exclude='.git' --exclude='venv' --exclude='__pycache__' --exclude='*.pyc' --exclude='.Dockerfile.runtime' ."
                        def fileSize = sh(script: "stat -c%s ${archiveName} 2>/dev/null || stat -f%z ${archiveName}", returnStdout: true).trim()
                        echo "[制品上传] 上传文件: ${archiveName} (${fileSize} bytes)"

                        def uploadUrl = params.PLATFORM_CALLBACK_URL
                            .replace('/pipeline/callback', '/artifact/upload')
                            .replace('/stage/callback', '/artifact/upload')

                        def curlStatus = sh(script: """
                            set -e
                            curl -s -w '%{http_code}' -o /tmp/artifact_resp.json \\
                                -X POST '${uploadUrl}' \\
                                -F 'file=@${archiveName}' \\
                                -F 'pipeline_id=${params.PIPELINE_ID ?: 0}' \\
                                -F 'run_id=${params.RUN_ID ?: 0}' \\
                                -F 'build_number=${env.BUILD_NUMBER}' \\
                                -F 'version=${env.FINAL_TAG}' \\
                                -F 'language_type=python' \\
                                -F 'artifact_type=archive' \\
                                -F 'git_repo=${params.GIT_REPO}' \\
                                -F 'git_branch=${env.GIT_BRANCH_NAME}' \\
                                -F 'git_commit=${env.GIT_COMMIT_SHORT}' \\
                                --connect-timeout 10 \\
                                --max-time 300 \\
                                --tcp-nodelay \\
                                -H "Expect:" \\
                                --retry 2 --retry-delay 5
                        """, returnStdout: true).trim()

                        if (curlStatus.endsWith('200')) {
                            echo "[制品上传] ✅ 上传成功"
                        } else {
                            echo "[制品上传] ❌ 上传失败: HTTP ${curlStatus[-3..-1]}"
                            def respBody = sh(script: "cat /tmp/artifact_resp.json 2>/dev/null || echo '{}'", returnStdout: true).trim()
                            echo "[制品上传] 响应内容: ${respBody}"
                            echo "[制品上传] 上传地址: ${uploadUrl}"
                            error("制品上传失败: HTTP ${curlStatus[-3..-1]}")
                        }
                        sh "rm -f ${archiveName} /tmp/artifact_resp.json 2>/dev/null || true"
                }
            }
            post {
                success { script { stageCallback('upload_artifact', 'success') } }
                failure { script { stageCallback('upload_artifact', 'failed') } }
            }
        }

        stage('Build Image') {
            steps {
                echo "=== 构建镜像（纯运行时，打包 Python 应用） ==="
                script {
                    def dockerfile = params.DOCKERFILE_PATH?.trim()
                    def pythonVersion = params.PYTHON_VERSION ?: '3.11'

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
FROM python:${pythonVersion}-slim
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 TZ=Asia/Shanghai
ENV PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple PIP_NO_CACHE_DIR=1
RUN apt-get update && apt-get install -y --no-install-recommends curl tzdata && \\
    ln -snf /usr/share/zoneinfo/\$TZ /etc/localtime && \\
    apt-get clean && rm -rf /var/lib/apt/lists/*
RUN groupadd -r app && useradd -r -g app app
WORKDIR /app
COPY requirements.txt* ./
RUN if [ -f requirements.txt ]; then pip install -r requirements.txt; fi
COPY . .
RUN chown -R app:app /app
USER app
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \\
    CMD curl -f http://localhost:8000/health || exit 1
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
"""
                            echo "[Build Image] ${forceGenerate ? '强制' : '项目无 Dockerfile，'}已自动生成纯运行时 Dockerfile"
                        }
                    }

                    // 使用 BuildKit 本地层缓存 + Dockerfile 内容哈希防止缓存过期
                    def cacheDir = "${env.BUILDKIT_CACHE}/${env.JOB_NAME}".replaceAll('[^a-zA-Z0-9/_.-]', '_')
                    def dfHash = sh(script: "md5sum ${dockerfile} | awk '{print \$1}'", returnStdout: true).trim()
                    def cacheHashFile = "${cacheDir}/.dockerfile_hash"
                    def oldHash = sh(script: "cat ${cacheHashFile} 2>/dev/null || echo ''", returnStdout: true).trim()
                    def cacheArgs = ''
                    if (dfHash != oldHash) {
                        echo "[Build Image] Dockerfile 内容已变化（${oldHash ?: '无缓存'} → ${dfHash}），清除旧缓存"
                        sh "rm -rf ${cacheDir} 2>/dev/null || true"
                    } else {
                        echo "[Build Image] Dockerfile 未变化，复用 BuildKit 层缓存"
                        cacheArgs = "--cache-from type=local,src=${cacheDir}"
                    }
                    sh """
                        set -e
                        mkdir -p ${cacheDir}
                        echo '${dfHash}' > ${cacheHashFile}
                        nerdctl build \\
                            -t ${env.FULL_IMAGE} \\
                            -f ${dockerfile} \\
                            ${cacheArgs} \\
                            --cache-to type=local,dest=${cacheDir},mode=max \\
                            --label git.commit=${env.GIT_COMMIT_FULL} \\
                            --label git.branch=${env.GIT_BRANCH_NAME} \\
                            --label build.number=${env.BUILD_NUMBER} \\
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
                    env.IMAGE_WITH_DIGEST = env.IMAGE_DIGEST ? "${params.IMAGE_REPO}@${env.IMAGE_DIGEST}" : env.FULL_IMAGE
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
                def msg
                if (!params.ENABLE_SONAR) {
                    msg = 'Python 项目构建成功'
                } else if (env.SONAR_ANALYSIS_FAILED == 'true') {
                    msg = "Python 项目构建成功 | SonarQube: UNAVAILABLE（扫描阶段连接失败，请检查 SonarQube 服务状态）"
                } else {
                    msg = "Python 项目构建成功 | SonarQube: ${env.SONAR_QUALITY_GATE_STATUS ?: 'SKIPPED'}"
                }
                callbackPlatform('SUCCESS', msg)
            }
        }
        failure { script { callbackPlatform('FAILURE', 'Python 项目构建失败') } }
        aborted { script { callbackPlatform('ABORTED', '构建中止') } }
        always {
            sh "nerdctl rmi ${env.FULL_IMAGE} || true"
            sh 'rm -rf .git venv __pycache__ coverage.xml 2>/dev/null || true'
        }
    }
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

def ratingToLetter(Double rating) {
    if (rating <= 1.0) return 'A'
    if (rating <= 2.0) return 'B'
    if (rating <= 3.0) return 'C'
    if (rating <= 4.0) return 'D'
    return 'E'
}

// ==================== 统一回调函数 ====================
def stageCallback(String stageType, String status) {
    if (!params.PLATFORM_CALLBACK_URL?.trim()) return
    try {
        def payload = [job_name: env.JOB_NAME, build_number: env.BUILD_NUMBER as Integer,
            pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0, stage_type: stageType, status: status]
        def body = groovy.json.JsonOutput.toJson(payload)
        def stageUrl = params.PLATFORM_CALLBACK_URL.replace('/pipeline/callback', '/stage/callback')
        def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${stageType}") : ''
        def headers = signature ? [[name: 'X-Signature', value: signature]] : []
        httpRequest(url: stageUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON',
            requestBody: body, customHeaders: headers, validResponseCodes: '100:599', timeout: 10)
    } catch (e) { echo "[阶段回调] 非致命: ${e.message}" }
}

def callbackPlatform(String status, String message) {
    if (!params.PLATFORM_CALLBACK_URL?.trim()) { echo "未配置回调地址"; return }
    def payload = [job_name: env.JOB_NAME, build_number: env.BUILD_NUMBER as Integer, status: status,
        pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
        image_url: env.FULL_IMAGE ?: '', image_digest: env.IMAGE_DIGEST ?: '',
        image_with_digest: env.IMAGE_WITH_DIGEST ?: '', git_commit: env.GIT_COMMIT_SHORT ?: '',
        git_branch: env.GIT_BRANCH_NAME ?: '',
        duration_sec: currentBuild.duration ? (currentBuild.duration / 1000) as Integer : 0,
        message: message, build_url: env.BUILD_URL ?: '']
    def body = groovy.json.JsonOutput.toJson(payload)
    def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${status}") : ''
    def headers = signature ? [[name: 'X-Signature', value: signature]] : []
    httpRequest(url: params.PLATFORM_CALLBACK_URL, httpMode: 'POST', contentType: 'APPLICATION_JSON',
        requestBody: body, customHeaders: headers, validResponseCodes: '200:299', consoleLogResponseBody: true)
}

def hmacSha256(String secret, String data) {
    def result = ''
    withEnv(["SIGN_SECRET=${secret}", "SIGN_DATA=${data}"]) {
        result = sh(script: 'set +x && printf "%s" "$SIGN_DATA" | openssl dgst -sha256 -hmac "$SIGN_SECRET" | awk \'{print $2}\'', returnStdout: true).trim()
    }
    return result
}
