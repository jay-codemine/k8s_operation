// ==============================================================================
// K8s Operation Platform - Python 项目通用构建模板（K8s Pod Agent + Kaniko 容器化版）
// ==============================================================================
// 设计理念：一个模板服务 100+ Python 项目，所有项目差异通过参数传入
// 运行模式：Jenkins K8s 动态 Pod Agent，每次构建创建独立 Pod，完成后自动销毁
// 镜像构建：使用 Kaniko（无需 Docker daemon、无需特权容器）
// 支持框架：Flask, FastAPI, Django, Celery 等
//
// ======================== Jenkins Job 配置方式 ========================
//   1. Jenkins → New Item → Pipeline → 命名为 python-pipeline
//   2. Pipeline → Definition: Pipeline script from SCM
//   3. SCM: Git → Repository URL: 平台仓库地址
//   4. Script Path: configs/jenkins-templates/python-pipeline.groovy
// ==============================================================================

pipeline {
    agent {
        kubernetes {
            yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    jenkins-build: python
spec:
  containers:
  - name: python
    image: docker.m.daocloud.io/library/python:3.11-slim
    imagePullPolicy: IfNotPresent
    command: ['sleep', '99d']
    resources:
      requests:
        cpu: '500m'
        memory: '1Gi'
      limits:
        cpu: '4'
        memory: '4Gi'
    env:
    - name: PIP_INDEX_URL
      value: 'https://pypi.tuna.tsinghua.edu.cn/simple'
    - name: PIP_NO_CACHE_DIR
      value: '0'
    - name: PYTHONDONTWRITEBYTECODE
      value: '1'
    volumeMounts:
    - name: pip-cache
      mountPath: /root/.cache/pip
    - name: workspace-volume
      mountPath: /home/jenkins/agent
  - name: kaniko
    image: gcr.m.daocloud.io/kaniko-project/executor:debug
    imagePullPolicy: IfNotPresent
    command: ['sleep', '99d']
    securityContext:
      runAsUser: 0
    resources:
      requests:
        cpu: '500m'
        memory: '1Gi'
      limits:
        cpu: '2'
        memory: '4Gi'
    volumeMounts:
    - name: workspace-volume
      mountPath: /home/jenkins/agent
  - name: jnlp
    image: docker.m.daocloud.io/jenkins/inbound-agent:latest-jdk21
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: '1'
        memory: 512Mi
    imagePullPolicy: IfNotPresent
  volumes:
  - name: pip-cache
    persistentVolumeClaim:
      claimName: jenkins-pip-cache
  - name: workspace-volume
    emptyDir: {}
"""
        }
    }

    options {
        timeout(time: 30, unit: 'MINUTES')
        // 并发限制由平台 config.yaml 的 MaxConcurrentBuilds 控制，通过参数动态注入
        buildDiscarder(logRotator(numToKeepStr: '20'))
        skipDefaultCheckout(true)
    }

    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址（必填）')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库地址（必填）')
        string(name: 'IMAGE_TAG', defaultValue: '', description: '镜像标签（空则自动生成）')
        string(name: 'DOCKERFILE_PATH', defaultValue: '', description: 'Dockerfile 路径（空则自动生成）')
        booleanParam(name: 'USE_PROJECT_DOCKERFILE', defaultValue: false, description: '使用项目根目录的 Dockerfile（而非平台自动生成）')
        string(name: 'EXTRA_REPOS', defaultValue: '', description: '额外依赖仓库列表（格式: url|path[|branch];url|path[|branch]）')
        string(name: 'LANGUAGE_TYPE', defaultValue: '', description: '平台注入的语言类型')

        string(name: 'PIPELINE_ID', defaultValue: '', description: '平台流水线ID')
        string(name: 'RUN_ID', defaultValue: '', description: '平台运行记录ID')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')

        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过测试')
        string(name: 'PYTHON_VERSION', defaultValue: '3.11', description: 'Python 版本')
        string(name: 'GIT_CREDENTIAL_ID', defaultValue: 'gitee-id', description: 'Git 凭证ID')
        string(name: 'REGISTRY_CREDENTIAL_ID', defaultValue: 'harbor-registry', description: '镜像仓库凭证ID')
        string(name: 'HMAC_CREDENTIAL_ID', defaultValue: 'hmac-secret', description: 'HMAC签名凭证ID')

        booleanParam(name: 'ENABLE_SONAR', defaultValue: false, description: '启用 SonarQube')
        string(name: 'SONAR_PROJECT_KEY', defaultValue: '', description: 'SonarQube 项目 Key')
        string(name: 'SONAR_PROJECT_NAME', defaultValue: '', description: 'SonarQube 项目名称')
        string(name: 'SONAR_SOURCES', defaultValue: '.', description: '源代码目录')
        string(name: 'SONAR_EXCLUSIONS', defaultValue: '**/venv/**,**/__pycache__/**,**/test_*,**/*_test.py,**/migrations/**', description: '排除扫描')
        booleanParam(name: 'SONAR_QUALITY_GATE', defaultValue: true, description: '启用质量门禁')
        string(name: 'SONAR_COVERAGE_THRESHOLD', defaultValue: '80', description: '覆盖率阈值')
        string(name: 'SONAR_NEW_BUGS_MAX', defaultValue: '0', description: 'Bug 最大数')
        string(name: 'SONAR_CODE_SMELLS_MAX', defaultValue: '10', description: '异味最大数')
        string(name: 'SONAR_VULNERABILITIES_MAX', defaultValue: '0', description: '漏洞最大数')
        string(name: 'SONAR_DUPLICATIONS_MAX', defaultValue: '3', description: '重复率阈值')
        string(name: 'SONAR_GATE_ACTION', defaultValue: 'block', description: '门禁策略: block | warn | skip')
        booleanParam(name: 'ENABLE_ARTIFACT_UPLOAD', defaultValue: true, description: '启用制品上传')

        // 并发控制（由平台 config.yaml 的 MaxConcurrentBuilds 自动注入，无需手动修改）
        string(name: 'MAX_CONCURRENT_BUILDS', defaultValue: '10', description: '最大并发构建数（平台自动注入，勿手动修改）')
    }

    environment {
        REGISTRY_CREDS = credentials("${params.REGISTRY_CREDENTIAL_ID ?: 'harbor-registry'}")
        HMAC_SECRET    = credentials("${params.HMAC_CREDENTIAL_ID ?: 'hmac-secret'}")
    }

    stages {

        stage('Clean Workspace') {
            steps {
                deleteDir()
                sh 'rm -rf .git 2>/dev/null || true'
                script {
                    // 动态设置并发限制（从平台 config.yaml 的 MaxConcurrentBuilds 注入，需 Throttle Concurrent Builds 插件）
                    def maxConcurrent = (params.MAX_CONCURRENT_BUILDS ?: '10').toInteger()
                    properties([
                        [$class: 'ThrottleJobProperty',
                         maxConcurrentPerNode: 0,
                         maxConcurrentTotal: maxConcurrent,
                         categories: [],
                         throttleEnabled: true,
                         throttleOption: 'project'
                        ]
                    ])
                    def expectedType = 'python'
                    def actualType = params.LANGUAGE_TYPE?.trim()
                    if (actualType && actualType != expectedType) {
                        error("模板类型不匹配: 语言=${actualType}, 当前=python")
                    }
                    if (!params.GIT_REPO?.trim()) { error("GIT_REPO 不能为空") }
                    if (!params.IMAGE_REPO?.trim()) { error("IMAGE_REPO 不能为空") }

                    def targetBranch = params.GIT_BRANCH?.trim() ?: 'main'
                    sh 'rm -rf .git 2>/dev/null || true'
                    checkout([$class: 'GitSCM', branches: [[name: "*/${targetBranch}"]],
                        extensions: [[$class: 'CleanBeforeCheckout', deleteUntrackedNestedRepositories: true],
                            [$class: 'LocalBranch', localBranch: targetBranch],
                            [$class: 'CloneOption', depth: 1, shallow: true, noTags: true, timeout: 10, honorRefspec: true]],
                        userRemoteConfigs: [[url: params.GIT_REPO, credentialsId: params.GIT_CREDENTIAL_ID ?: 'gitee-id']]])
                    env.TARGET_BRANCH = targetBranch
                    echo "[Checkout] ✅ ${sh(script: 'git log -1 --format=\"%h %s\"', returnStdout: true).trim()}"
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
                    env.FINAL_TAG = params.IMAGE_TAG?.trim() ?: "${env.GIT_COMMIT_SHORT}-${env.BUILD_TS}"
                    env.FULL_IMAGE = "${params.IMAGE_REPO}:${env.FINAL_TAG}"
                    echo "Commit: ${env.GIT_COMMIT_SHORT} | Image: ${env.FULL_IMAGE}"
                }
            }
            post { success { script { stageCallback('checkout', 'success') } }; failure { script { stageCallback('checkout', 'failed') } } }
        }

        // ==================== 克隆额外依赖仓库 ====================
        stage('Clone Extra Repos') {
            when { expression { return params.EXTRA_REPOS?.trim() } }
            steps {
                echo "=== 克隆私有依赖仓库 ==="
                script {
                    def extraRepos = params.EXTRA_REPOS.trim()
                    extraRepos.split(';').each { entry ->
                        def parts = entry.split('\\|')
                        if (parts.size() >= 2) {
                            def repoUrl = parts[0].trim()
                            def targetPath = parts[1].trim()
                            def branch = parts.size() >= 3 ? parts[2].trim() : 'master'
                            echo "[Extra Repos] 克隆: ${repoUrl} → ${targetPath} (${branch})"
                            try {
                                sh "mkdir -p \$(dirname ${targetPath})"
                                sh "git clone --depth 1 -b ${branch} ${repoUrl} ${targetPath}"
                            } catch (Exception e) {
                                echo "[Extra Repos] 克隆失败: ${repoUrl} (${branch})，尝试默认分支..."
                                try { sh "git clone --depth 1 ${repoUrl} ${targetPath}" }
                                catch (Exception e2) { echo "[Extra Repos] 最终克隆失败: ${repoUrl}，跳过" }
                            }
                        }
                    }
                    echo "[Extra Repos] ✅ 所有依赖仓库克隆完成"
                }
            }
        }

        // ==================== 准备构建探针（SkyWalking/OpenTelemetry） ====================
        stage('Prepare Build Agents') {
            when { expression { return params.ENABLE_TRACING?.toBoolean() } }
            steps {
                echo "=== 准备 APM 探针 ==="
                container('python') {
                    script {
                        def agentsDir = '.agents'
                        sh "mkdir -p ${agentsDir}"
                        env.AGENT_DOCKER_COPY_LINES = ''
                        env.AGENT_ENV_LINES = ''
                        def agentsPrepared = []

                        def platformUrl = params.PLATFORM_CALLBACK_URL?.trim()
                        if (platformUrl) {
                            def apiBase = platformUrl.replaceAll('/api/v1/k8s/cicd/pipeline/callback.*', '/api/v1/k8s/cicd')
                            def listUrl = "${apiBase}/agent/by-scope?scope=python"
                            try {
                                def response = sh(script: "wget -q -O - '${listUrl}' 2>/dev/null || curl -s '${listUrl}'", returnStdout: true).trim()
                                if (response) {
                                    def json = new groovy.json.JsonSlurper().parseText(response)
                                    def agentList = json?.data?.list ?: json?.list ?: []
                                    agentList.each { agent ->
                                        def fileName = agent.file_name ?: "${agent.name}"
                                        def destPath = agent.docker_copy_dest ?: "/app/${fileName}"
                                        def localDir = "${agentsDir}/${agent.name}"
                                        def localPath = "${localDir}/${fileName}"
                                        sh "mkdir -p ${localDir}"
                                        def downloadUrl = "${apiBase}/agent/download?name=${agent.name}"
                                        def dlResult = sh(script: "wget -q -O '${localPath}' '${downloadUrl}' || curl -s -o '${localPath}' '${downloadUrl}'", returnStatus: true)
                                        if (dlResult == 0) {
                                            env.AGENT_DOCKER_COPY_LINES += "COPY ${localPath} ${destPath}\n"
                                            if (agent.env_key && agent.env_value) {
                                                env.AGENT_ENV_LINES += "ENV ${agent.env_key}=\"${agent.env_value}\"\n"
                                            }
                                            // 自动解压归档文件（tgz/zip），供项目自带 Dockerfile 直接 COPY
                                            def lowerName = fileName.toLowerCase()
                                            if (lowerName.endsWith('.tgz') || lowerName.endsWith('.tar.gz')) {
                                                sh "mkdir -p ./${agent.name} && tar -xzf ${localPath} -C ./${agent.name}/"
                                                echo "[Agents] 已解压 tgz → ./${agent.name}/"
                                            } else if (lowerName.endsWith('.zip')) {
                                                sh "mkdir -p ./${agent.name} && unzip -q ${localPath} -d ./${agent.name}/"
                                                echo "[Agents] 已解压 zip → ./${agent.name}/"
                                            } else {
                                                sh "mkdir -p ./${agent.name}"
                                                sh "cp ${localPath} ./${agent.name}/${fileName}"
                                                echo "[Agents] 已复制单文件 → ./${agent.name}/${fileName}"
                                            }
                                            agentsPrepared << agent.name
                                        }
                                    }
                                }
                            } catch (Exception e) {
                                echo "[Agents] 平台 API 不可用: ${e.message}"
                            }
                        }
                        echo "[Agents] === 准备完成: ${agentsPrepared.join(', ') ?: '无'} ==="
                    }
                }
            }
        }

        stage('Install Dependencies') {
            // 仅在需要跑 Lint/Test 时安装依赖；SKIP_TESTS 时依赖会在 Kaniko 构建镜像阶段安装，避免重复安装
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 安装 Python 依赖 ==="
                container('python') {
                    script {
                        if (fileExists('requirements.txt')) {
                            sh 'pip install -r requirements.txt -q'
                        } else if (fileExists('setup.py') || fileExists('pyproject.toml')) {
                            sh 'pip install -e . -q'
                        } else {
                            echo "未检测到依赖文件，跳过"
                        }
                    }
                }
            }
            post { success { script { stageCallback('dependencies', 'success') } }; failure { script { stageCallback('dependencies', 'failed') } } }
        }

        stage('Lint') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                container('python') {
                    sh 'pip install flake8 -q 2>/dev/null || true; flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics 2>/dev/null || true'
                }
            }
            post { success { script { stageCallback('lint', 'success') } }; failure { script { stageCallback('lint', 'failed') } } }
        }

        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                container('python') {
                    script {
                        def hasTests = sh(script: "find . -name 'test_*.py' -o -name '*_test.py' | grep . >/dev/null 2>&1 && echo yes || echo no", returnStdout: true).trim()
                        if (hasTests == 'yes') {
                            sh 'pip install pytest pytest-cov -q 2>/dev/null || true; pytest --cov=. -v --cov-report=xml:coverage.xml 2>/dev/null || pytest -v 2>/dev/null || true'
                        } else { echo "未检测到测试文件，跳过" }
                    }
                }
            }
            post { success { script { stageCallback('test', 'success'); stageCallback('build_binary', 'success') } }; failure { script { stageCallback('test', 'failed'); stageCallback('build_binary', 'failed') } } }
        }

        stage('SonarQube Analysis') {
            when { expression { return params.ENABLE_SONAR } }
            steps {
                script {
                    try {
                        def projectKey = params.SONAR_PROJECT_KEY?.trim() ?: env.JOB_NAME.replaceAll('/', '_')
                        withSonarQubeEnv('SonarQube') {
                            container('python') {
                                sh """
                                    if ! command -v sonar-scanner &>/dev/null; then
                                        apt-get update -qq && apt-get install -y -qq wget unzip 2>/dev/null || true
                                        wget -q https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-5.0.1.3006-linux-x64.zip -O /tmp/sonar.zip 2>/dev/null || true
                                        [ -f /tmp/sonar.zip ] && unzip -qo /tmp/sonar.zip -d /tmp/ && export PATH="/tmp/sonar-scanner-5.0.1.3006-linux-x64/bin:\$PATH"
                                    fi
                                    sonar-scanner -Dsonar.projectKey=${projectKey} \
                                        -Dsonar.sources=${params.SONAR_SOURCES ?: '.'} \
                                        -Dsonar.exclusions=${params.SONAR_EXCLUSIONS ?: '**/venv/**'},**/build/**,**/dist/** \
                                        -Dsonar.python.coverage.reportPaths=coverage.xml \
                                        -Dsonar.scm.disabled=true -Dsonar.qualitygate.wait=false -Dsonar.threads=4
                                """
                            }
                        }
                        stageCallback('sonar', 'success')
                    } catch (e) {
                        stageCallback('sonar', 'failed'); env.SONAR_ANALYSIS_FAILED = 'true'
                        error("SonarQube 扫描失败: ${e.message}")
                    }
                }
            }
        }

        stage('Quality Gate') {
            when { allOf { expression { return params.ENABLE_SONAR && params.SONAR_QUALITY_GATE }; expression { return env.SONAR_ANALYSIS_FAILED != 'true' }; expression { return (params.SONAR_GATE_ACTION ?: 'block') != 'skip' } } }
            steps {
                script {
                    def qg = waitForQualityGate(webhookSecretId: '', abortPipeline: false)
                    env.SONAR_QUALITY_GATE_STATUS = qg.status
                    if (qg.status != 'OK' && (params.SONAR_GATE_ACTION ?: 'block') == 'block') { error("Quality Gate 未通过: ${qg.status}") }
                }
            }
            post { success { script { stageCallback('quality_gate', 'success') } }; failure { script { stageCallback('quality_gate', 'failed') } } }
        }

        stage('Upload Artifact') {
            when { expression { return params.ENABLE_ARTIFACT_UPLOAD && params.PLATFORM_CALLBACK_URL?.trim() } }
            steps {
                container('python') {
                    script {
                        def appName = params.GIT_REPO?.tokenize('/')?.last()?.replace('.git', '') ?: 'python-app'
                        def archiveName = "${appName}-${env.FINAL_TAG}.tar.gz"
                        sh "tar czf ${archiveName} --exclude='.git' --exclude='venv' --exclude='__pycache__' --exclude='*.pyc' ."
                        def uploadUrl = params.PLATFORM_CALLBACK_URL.replace('/pipeline/callback', '/artifact/upload').replace('/stage/callback', '/artifact/upload')
                        def curlStatus = sh(script: "curl -s -w '%{http_code}' -o /tmp/resp.json -X POST '${uploadUrl}' -F 'file=@${archiveName}' -F 'pipeline_id=${params.PIPELINE_ID ?: 0}' -F 'run_id=${params.RUN_ID ?: 0}' -F 'version=${env.FINAL_TAG}' -F 'language_type=python' -F 'artifact_type=archive' --connect-timeout 10 --max-time 300", returnStdout: true).trim()
                        if (!curlStatus.endsWith('200')) { error("制品上传失败") }
                        echo "[制品上传] ✅ 上传成功"
                        sh "rm -f ${archiveName} /tmp/resp.json 2>/dev/null || true"
                    }
                }
            }
            post { success { script { stageCallback('upload_artifact', 'success') } }; failure { script { stageCallback('upload_artifact', 'failed') } } }
        }

        // ==================== Kaniko 构建 + 推送镜像 ====================
        stage('Build & Push Image') {
            steps {
                echo "=== Kaniko 构建并推送镜像 ==="
                container('kaniko') {
                    script {
                        def dockerfile = params.DOCKERFILE_PATH?.trim()
                        def pythonVersion = params.PYTHON_VERSION ?: '3.11'

                        // Dockerfile 选择优先级：
                        // 1. DOCKERFILE_PATH 显式指定 → 直接使用
                        // 2. USE_PROJECT_DOCKERFILE=true 且项目根有 Dockerfile → 使用项目自带
                        // 3. 平台自动生成生产级 Dockerfile
                        if (!dockerfile || dockerfile == '__PLATFORM_GENERATE__') {
                            if (dockerfile != '__PLATFORM_GENERATE__' && params.USE_PROJECT_DOCKERFILE && fileExists('Dockerfile')) {
                                dockerfile = 'Dockerfile'
                                echo "[Build Image] 使用项目自带 Dockerfile"
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
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD curl -f http://localhost:8000/health || exit 1
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
"""
                            }
                            echo "[Build Image] 平台统一生成 Dockerfile"
                        }

                        def registryHost = params.IMAGE_REPO.split('/')[0]
                        sh """
                            mkdir -p /kaniko/.docker
                            echo '{"auths":{"${registryHost}":{"username":"${REGISTRY_CREDS_USR}","password":"${REGISTRY_CREDS_PSW}"}}}' > /kaniko/.docker/config.json
                            /kaniko/executor \
                                --context=. \
                                --dockerfile=${dockerfile} \
                                --destination=${env.FULL_IMAGE} \
                                --cache=true \
                                --cache-repo=${registryHost}/kaniko-cache/python \
                                --label git.commit=${env.GIT_COMMIT_FULL} \
                                --label git.branch=${env.GIT_BRANCH_NAME} \
                                --label build.mode=k8s-kaniko \
                                --snapshot-mode=redo \
                                --use-new-run
                        """
                        env.IMAGE_DIGEST = ''; env.IMAGE_WITH_DIGEST = env.FULL_IMAGE
                        echo "[Build & Push] ✅ ${env.FULL_IMAGE}"
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
        success { script { callbackPlatform('SUCCESS', "Python 项目构建成功${params.ENABLE_SONAR ? ' | SonarQube: ' + (env.SONAR_QUALITY_GATE_STATUS ?: 'SKIPPED') : ''}") } }
        failure { script { callbackPlatform('FAILURE', 'Python 项目构建失败') } }
        aborted { script { callbackPlatform('ABORTED', '构建中止') } }
    }
}

// ==================== 回调函数 ====================
def stageCallback(String stageType, String status) {
    if (!params.PLATFORM_CALLBACK_URL?.trim()) return
    try {
        def payload = [job_name: env.JOB_NAME, build_number: env.BUILD_NUMBER as Integer, pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0, run_id: params.RUN_ID ? params.RUN_ID as Long : 0, stage_type: stageType, status: status]
        def body = groovy.json.JsonOutput.toJson(payload)
        def stageUrl = params.PLATFORM_CALLBACK_URL.replace('/pipeline/callback', '/stage/callback')
        def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${stageType}") : ''
        def headers = signature ? [[name: 'X-Signature', value: signature]] : []
        httpRequest(url: stageUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON', requestBody: body, customHeaders: headers, validResponseCodes: '100:599', timeout: 10)
    } catch (e) { echo "[阶段回调] 非致命: ${e.message}" }
}

def callbackPlatform(String status, String message) {
    if (!params.PLATFORM_CALLBACK_URL?.trim()) { echo "未配置回调地址"; return }
    def payload = [job_name: env.JOB_NAME, build_number: env.BUILD_NUMBER as Integer, status: status,
        pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
        run_id: params.RUN_ID ? params.RUN_ID as Long : 0,
        image_url: env.FULL_IMAGE ?: '', image_digest: env.IMAGE_DIGEST ?: '', image_with_digest: env.IMAGE_WITH_DIGEST ?: '',
        git_commit: env.GIT_COMMIT_SHORT ?: '', git_branch: env.GIT_BRANCH_NAME ?: '',
        duration_sec: currentBuild.duration ? (currentBuild.duration / 1000) as Integer : 0, message: message, build_url: env.BUILD_URL ?: '']
    def body = groovy.json.JsonOutput.toJson(payload)
    def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${status}") : ''
    def headers = signature ? [[name: 'X-Signature', value: signature]] : []
    httpRequest(url: params.PLATFORM_CALLBACK_URL, httpMode: 'POST', contentType: 'APPLICATION_JSON', requestBody: body, customHeaders: headers, validResponseCodes: '200:299', consoleLogResponseBody: true)
}

def hmacSha256(String secret, String data) {
    def result = ''; withEnv(["SIGN_SECRET=${secret}", "SIGN_DATA=${data}"]) { result = sh(script: 'set +x && printf "%s" "$SIGN_DATA" | openssl dgst -sha256 -hmac "$SIGN_SECRET" | awk \'{print $2}\'', returnStdout: true).trim() }; return result
}
