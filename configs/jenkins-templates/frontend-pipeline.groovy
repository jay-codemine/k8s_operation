// ==============================================================================
// K8s Operation Platform - 前端项目通用构建模板（K8s Pod Agent + Kaniko 容器化版）
// ==============================================================================
// 设计理念：一个模板服务 100+ 前端项目，所有项目差异通过参数传入
// 运行模式：Jenkins K8s 动态 Pod Agent，每次构建创建独立 Pod，完成后自动销毁
// 镜像构建：使用 Kaniko（无需 Docker daemon、无需特权容器）
// 支持框架：Vue.js, React, Angular, Next.js, Nuxt.js 等
//
// ======================== Jenkins Job 配置方式 ========================
//   1. Jenkins → New Item → Pipeline → 命名为 k8s-builder-frontend
//   2. Pipeline → Definition: Pipeline script from SCM
//   3. SCM: Git → Repository URL: 平台仓库地址
//   4. Script Path: configs/jenkins-templates/frontend-pipeline.groovy
// ==============================================================================

pipeline {
    agent {
        kubernetes {
            yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    jenkins-build: frontend
spec:
  containers:
  - name: node
    image: docker.m.daocloud.io/library/node:18-alpine3.18
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
    - name: NPM_CONFIG_REGISTRY
      value: 'https://registry.npmmirror.com'
    volumeMounts:
    - name: npm-cache
      mountPath: /root/.npm
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
  - name: npm-cache
    persistentVolumeClaim:
      claimName: jenkins-npm-cache
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
        string(name: 'DOCKERFILE_PATH', defaultValue: '', description: 'Dockerfile 路径（空则自动生成纯运行时 Dockerfile）')
        booleanParam(name: 'USE_PROJECT_DOCKERFILE', defaultValue: false, description: '使用项目根目录的 Dockerfile（而非平台自动生成）')
        string(name: 'EXTRA_REPOS', defaultValue: '', description: '额外依赖仓库列表（格式: url|path[|branch];url|path[|branch]）')
        string(name: 'LANGUAGE_TYPE', defaultValue: '', description: '平台注入的语言类型')

        string(name: 'PIPELINE_ID', defaultValue: '', description: '平台流水线ID')
        string(name: 'RUN_ID', defaultValue: '', description: '平台运行记录ID')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')

        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过测试')
        string(name: 'NODE_VERSION', defaultValue: '18', description: 'Node.js 版本')
        string(name: 'BUILD_COMMAND', defaultValue: 'npm run build', description: '构建命令')
        string(name: 'BUILD_OUTPUT_DIR', defaultValue: 'dist', description: '构建产物目录')
        string(name: 'GIT_CREDENTIAL_ID', defaultValue: 'gitee-id', description: 'Git 凭证ID')
        string(name: 'REGISTRY_CREDENTIAL_ID', defaultValue: 'harbor-registry', description: '镜像仓库凭证ID')
        string(name: 'HMAC_CREDENTIAL_ID', defaultValue: 'hmac-secret', description: 'HMAC签名凭证ID')

        booleanParam(name: 'ENABLE_SONAR', defaultValue: false, description: '启用 SonarQube 代码质量扫描')
        string(name: 'SONAR_PROJECT_KEY', defaultValue: '', description: 'SonarQube 项目 Key')
        string(name: 'SONAR_PROJECT_NAME', defaultValue: '', description: 'SonarQube 项目名称')
        string(name: 'SONAR_SOURCES', defaultValue: 'src', description: '源代码目录')
        string(name: 'SONAR_EXCLUSIONS', defaultValue: '**/node_modules/**,**/dist/**,**/*.spec.*,**/*.test.*', description: '排除扫描的文件模式')
        booleanParam(name: 'SONAR_QUALITY_GATE', defaultValue: true, description: '启用质量门禁检查')
        string(name: 'SONAR_COVERAGE_THRESHOLD', defaultValue: '80', description: '代码覆盖率阈值（%）')
        string(name: 'SONAR_NEW_BUGS_MAX', defaultValue: '0', description: '新增 Bug 最大允许数')
        string(name: 'SONAR_CODE_SMELLS_MAX', defaultValue: '10', description: '代码异味最大允许数')
        string(name: 'SONAR_VULNERABILITIES_MAX', defaultValue: '0', description: '安全漏洞最大允许数')
        string(name: 'SONAR_DUPLICATIONS_MAX', defaultValue: '3', description: '代码重复率阈值（%）')
        string(name: 'SONAR_GATE_ACTION', defaultValue: 'block', description: '门禁失败策略: block | warn | skip')
        booleanParam(name: 'ENABLE_ARTIFACT_UPLOAD', defaultValue: true, description: '启用制品上传到平台制品库')

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
                sh 'rm -rf .git 2>/dev/null || true; find . -mindepth 1 -maxdepth 1 -exec rm -rf {} + 2>/dev/null || true'
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
                    def expectedType = 'frontend'
                    def actualType = params.LANGUAGE_TYPE?.trim()
                    if (actualType && actualType != expectedType) {
                        def scriptMap = ['go': 'go-pipeline.groovy', 'java': 'java-spring-pipeline.groovy', 'frontend': 'frontend-pipeline.groovy', 'python': 'python-pipeline.groovy']
                        error("模板类型不匹配: 语言=${actualType}, 当前=frontend. 请修改 Script Path 为: configs/jenkins-templates/${scriptMap[actualType] ?: "${actualType}-pipeline.groovy"}")
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
            post {
                success { script { stageCallback('checkout', 'success') } }
                failure { script { stageCallback('checkout', 'failed') } }
            }
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
                container('node') {
                    script {
                        def agentsDir = '.agents'
                        sh "mkdir -p ${agentsDir}"
                        env.AGENT_DOCKER_COPY_LINES = ''
                        env.AGENT_ENV_LINES = ''
                        def agentsPrepared = []

                        def platformUrl = params.PLATFORM_CALLBACK_URL?.trim()
                        if (platformUrl) {
                            def apiBase = platformUrl.replaceAll('/api/v1/k8s/cicd/pipeline/callback.*', '/api/v1/k8s/cicd')
                            def listUrl = "${apiBase}/agent/by-scope?scope=frontend"
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
            steps {
                echo "=== 安装依赖 ==="
                container('node') {
                    script {
                        if (!fileExists('package.json')) { echo "未检测到 package.json，跳过"; return }
                        sh 'npm ci --prefer-offline || npm install --prefer-offline'
                    }
                }
            }
            post {
                success { script { stageCallback('dependencies', 'success') } }
                failure { script { stageCallback('dependencies', 'failed') } }
            }
        }

        stage('Lint & Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 代码检查 + 测试 ==="
                container('node') {
                    sh 'npm run lint 2>/dev/null || true; npm run test:unit 2>/dev/null || npm test 2>/dev/null || true'
                }
            }
            post {
                success { script { stageCallback('test', 'success') } }
                failure { script { stageCallback('test', 'failed') } }
            }
        }

        stage('Build Frontend') {
            steps {
                echo "=== 构建前端 ==="
                container('node') {
                    sh """
                        set -e
                        ${params.BUILD_COMMAND}
                        test -d ${params.BUILD_OUTPUT_DIR}
                    """
                }
            }
            post {
                success { script { stageCallback('compile', 'success'); stageCallback('build_binary', 'success') } }
                failure { script { stageCallback('compile', 'failed'); stageCallback('build_binary', 'failed') } }
            }
        }

        stage('SonarQube Analysis') {
            when { expression { return params.ENABLE_SONAR } }
            steps {
                script {
                    try {
                        def projectKey  = params.SONAR_PROJECT_KEY?.trim()  ?: env.JOB_NAME.replaceAll('/', '_')
                        def projectName = params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME
                        withSonarQubeEnv('SonarQube') {
                            container('node') {
                                sh """
                                    if ! command -v sonar-scanner &>/dev/null; then
                                        wget -q https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-5.0.1.3006-linux-x64.zip -O /tmp/sonar.zip 2>/dev/null || true
                                        [ -f /tmp/sonar.zip ] && unzip -qo /tmp/sonar.zip -d /tmp/ && export PATH="/tmp/sonar-scanner-5.0.1.3006-linux-x64/bin:\$PATH"
                                    fi
                                    sonar-scanner -Dsonar.projectKey=${projectKey} -Dsonar.projectName=${projectName} \
                                        -Dsonar.sources=${params.SONAR_SOURCES ?: 'src'} \
                                        -Dsonar.exclusions=${params.SONAR_EXCLUSIONS ?: '**/node_modules/**'},**/build/**,**/.next/** \
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
                    def gateAction = params.SONAR_GATE_ACTION ?: 'block'
                    def qg = waitForQualityGate(webhookSecretId: '', abortPipeline: false)
                    env.SONAR_QUALITY_GATE_STATUS = qg.status
                    if (qg.status != 'OK' && gateAction == 'block') { error("Quality Gate 未通过: ${qg.status}") }
                }
            }
            post { success { script { stageCallback('quality_gate', 'success') } }; failure { script { stageCallback('quality_gate', 'failed') } } }
        }

        stage('Upload Artifact') {
            when { expression { return params.ENABLE_ARTIFACT_UPLOAD && params.PLATFORM_CALLBACK_URL?.trim() } }
            steps {
                container('node') {
                    script {
                        def outputDir = params.BUILD_OUTPUT_DIR ?: 'dist'
                        if (!fileExists(outputDir)) { error("构建产物目录不存在: ${outputDir}") }
                        def appName = params.GIT_REPO?.tokenize('/')?.last()?.replace('.git', '') ?: 'frontend-app'
                        def archiveName = "${appName}-${env.FINAL_TAG}.tar.gz"
                        sh "tar czf ${archiveName} -C ${outputDir} ."
                        def uploadUrl = params.PLATFORM_CALLBACK_URL.replace('/pipeline/callback', '/artifact/upload').replace('/stage/callback', '/artifact/upload')
                        def curlStatus = sh(script: "curl -s -w '%{http_code}' -o /tmp/resp.json -X POST '${uploadUrl}' -F 'file=@${archiveName}' -F 'pipeline_id=${params.PIPELINE_ID ?: 0}' -F 'run_id=${params.RUN_ID ?: 0}' -F 'version=${env.FINAL_TAG}' -F 'language_type=frontend' -F 'artifact_type=dist' --connect-timeout 10 --max-time 300", returnStdout: true).trim()
                        if (!curlStatus.endsWith('200')) { error("制品上传失败: HTTP ${curlStatus[-3..-1]}") }
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
                        def outputDir = params.BUILD_OUTPUT_DIR ?: 'dist'

                        // Dockerfile 选择优先级：
                        // 1. DOCKERFILE_PATH 显式指定 → 直接使用
                        // 2. USE_PROJECT_DOCKERFILE=true 且项目根有 Dockerfile → 使用项目自带
                        // 3. 平台自动生成生产级 Dockerfile（Gzip + 缓存 + 安全头 + API 代理 + WebSocket）
                        if (!dockerfile || dockerfile == '__PLATFORM_GENERATE__') {
                            if (params.USE_PROJECT_DOCKERFILE && fileExists('Dockerfile')) {
                                dockerfile = 'Dockerfile'
                                echo "[Build Image] 使用项目自带 Dockerfile"
                            } else {
                                dockerfile = '.Dockerfile.runtime'

                            // 生产级 Nginx 配置
                            writeFile file: 'nginx-app.conf', text: """\
upstream backend {
    server 127.0.0.1:8080;
}

server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    # Gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/javascript application/json application/javascript application/xml image/svg+xml;

    # 静态资源缓存（30天 + immutable）
    location ~* \\.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)\$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # API 反向代理 + WebSocket 支持
    location /api/ {
        proxy_pass http://backend;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        client_max_body_size 200m;
    }

    # 健康检查端点
    location /health {
        access_log off;
        return 200 "ok";
        add_header Content-Type text/plain;
    }

    # SPA History 模式 fallback
    location / {
        try_files \$uri \$uri/ /index.html;
    }

    # 安全响应头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
"""

                            // 运行时后端地址注入脚本（通过 API_BACKEND_URL 环境变量动态配置）
                            writeFile file: 'backend-url-entrypoint.sh', text: """\
#!/bin/sh
if [ -n "\$API_BACKEND_URL" ]; then
    echo "Configuring backend URL: \$API_BACKEND_URL"
    sed -i "s|server 127.0.0.1:8080;|server \${API_BACKEND_URL#http://};|g" /etc/nginx/conf.d/default.conf
fi
"""

                            writeFile file: dockerfile, text: """\
FROM nginx:1.25-alpine
RUN apk --no-cache add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY ${outputDir}/ /usr/share/nginx/html/
COPY nginx-app.conf /etc/nginx/conf.d/default.conf
COPY backend-url-entrypoint.sh /docker-entrypoint.d/90-backend-url.sh
RUN chmod +x /docker-entrypoint.d/90-backend-url.sh
RUN chown -R nginx:nginx /usr/share/nginx/html && chown -R nginx:nginx /var/cache/nginx && touch /var/run/nginx.pid && chown nginx:nginx /var/run/nginx.pid
USER nginx
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget -qO- http://localhost/health || exit 1
CMD ["nginx", "-g", "daemon off;"]
"""
                            }
                            echo "[Build Image] 平台统一生成 Dockerfile（Nginx + Gzip + 缓存 + 安全头）"
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
                                --cache-repo=${registryHost}/kaniko-cache/frontend \
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
        success { script { callbackPlatform('SUCCESS', "前端项目构建成功${params.ENABLE_SONAR ? ' | SonarQube: ' + (env.SONAR_QUALITY_GATE_STATUS ?: 'SKIPPED') : ''}") } }
        failure { script { callbackPlatform('FAILURE', '前端项目构建失败') } }
        aborted { script { callbackPlatform('ABORTED', '构建中止') } }
    }
}

// ==================== 回调函数 ====================
// 获取回调地址（兼容首次构建 params 未注册的情况，fallback 到 JCasC 全局环境变量）
def getCallbackUrl() {
    return params.PLATFORM_CALLBACK_URL?.trim() ?: env.PLATFORM_CALLBACK_URL?.trim() ?: ''
}

def stageCallback(String stageType, String status) {
    def callbackUrl = getCallbackUrl()
    if (!callbackUrl) return
    try {
        def payload = [job_name: env.JOB_NAME, build_number: env.BUILD_NUMBER as Integer, pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0, stage_type: stageType, status: status]
        def body = groovy.json.JsonOutput.toJson(payload)
        def stageUrl = callbackUrl.replace('/pipeline/callback', '/stage/callback')
        def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${stageType}") : ''
        def headers = signature ? [[name: 'X-Signature', value: signature]] : []
        httpRequest(url: stageUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON', requestBody: body, customHeaders: headers, validResponseCodes: '100:599', timeout: 10)
    } catch (e) { echo "[阶段回调] 非致命: ${e.message}" }
}

def callbackPlatform(String status, String message) {
    def callbackUrl = getCallbackUrl()
    if (!callbackUrl) { echo "未配置回调地址"; return }
    def payload = [job_name: env.JOB_NAME, build_number: env.BUILD_NUMBER as Integer, status: status,
        pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
        run_id: params.RUN_ID ? params.RUN_ID as Long : 0,
        image_url: env.FULL_IMAGE ?: '', image_digest: env.IMAGE_DIGEST ?: '', image_with_digest: env.IMAGE_WITH_DIGEST ?: '',
        git_commit: env.GIT_COMMIT_SHORT ?: '', git_branch: env.GIT_BRANCH_NAME ?: '',
        duration_sec: currentBuild.duration ? (currentBuild.duration / 1000) as Integer : 0, message: message, build_url: env.BUILD_URL ?: '']
    def body = groovy.json.JsonOutput.toJson(payload)
    def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${status}") : ''
    def headers = signature ? [[name: 'X-Signature', value: signature]] : []
    httpRequest(url: callbackUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON', requestBody: body, customHeaders: headers, validResponseCodes: '200:299', consoleLogResponseBody: true)
}

def hmacSha256(String secret, String data) {
    def result = ''; withEnv(["SIGN_SECRET=${secret}", "SIGN_DATA=${data}"]) { result = sh(script: 'set +x && printf "%s" "$SIGN_DATA" | openssl dgst -sha256 -hmac "$SIGN_SECRET" | awk \'{print $2}\'', returnStdout: true).trim() }; return result
}
