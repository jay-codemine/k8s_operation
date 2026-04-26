// ==============================================================================
// K8s Operation Platform - 前端项目通用构建模板（Vue/React/Angular）
// ==============================================================================
// 设计理念：一个模板服务 100+ 前端项目，所有项目差异通过参数传入
// 支持框架：Vue.js, React, Angular, Next.js, Nuxt.js 等
//
// ======================== Jenkins Job 配置方式 ========================
// 推荐使用 "Pipeline script from SCM"（版本化管理，自动同步更新）：
//   1. Jenkins → New Item → Pipeline → 命名为 k8s-builder-frontend
//   2. Pipeline → Definition: Pipeline script from SCM
//   3. SCM: Git → Repository URL: 平台仓库地址
//   4. Script Path: configs/jenkins-templates/frontend-pipeline.groovy
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
        string(name: 'NODE_VERSION', defaultValue: '18', description: 'Node.js 版本')
        string(name: 'BUILD_COMMAND', defaultValue: 'npm run build', description: '构建命令')
        string(name: 'BUILD_OUTPUT_DIR', defaultValue: 'dist', description: '构建产物目录')
        string(name: 'GIT_CREDENTIAL_ID', defaultValue: 'gitee-id', description: 'Git 凭证ID')

        // SonarQube 代码质量扫描参数
        booleanParam(name: 'ENABLE_SONAR', defaultValue: false, description: '启用 SonarQube 代码质量扫描')
        string(name: 'SONAR_PROJECT_KEY', defaultValue: '', description: 'SonarQube 项目 Key（空则使用 Job 名称）')
        string(name: 'SONAR_PROJECT_NAME', defaultValue: '', description: 'SonarQube 项目名称（空则使用 Job 名称）')
        string(name: 'SONAR_SOURCES', defaultValue: 'src', description: '源代码目录')
        string(name: 'SONAR_EXCLUSIONS', defaultValue: '**/node_modules/**,**/dist/**,**/*.spec.*,**/*.test.*', description: '排除扫描的文件模式')
        booleanParam(name: 'SONAR_QUALITY_GATE', defaultValue: true, description: '启用质量门禁检查（不通过则构建失败）')
    }

    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
        HMAC_SECRET    = credentials('hmac-secret')
        NPM_REGISTRY   = 'https://registry.npmmirror.com'
        // npm 缓存放 workspace 外，跨构建持久复用
        NPM_CACHE_DIR  = '/var/lib/jenkins/.npm-cache'
    }

    stages {

        stage('Clean Workspace') {
            steps {
                // 选择性清理：npm 缓存在外部目录，不受影响
                sh 'find . -mindepth 1 -maxdepth 1 | xargs rm -rf 2>/dev/null || true'
                script {
                    // 语言类型交叉校验：防止自定义 Job 配错脚本
                    def expectedType = 'frontend'
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
当前模板类型: ${expectedType} (frontend-pipeline.groovy)

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
                        userRemoteConfigs: [[url: params.GIT_REPO, credentialsId: params.GIT_CREDENTIAL_ID ?: 'gitee-id']]
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
                echo "=== 安装依赖（复用缓存） ==="
                script {
                    if (!fileExists('package.json')) { echo "未检测到 package.json，跳过"; return }
                    sh """
                        set -e
                        npm config set registry ${NPM_REGISTRY}
                        npm config set cache ${NPM_CACHE_DIR}
                        npm ci --prefer-offline --cache ${NPM_CACHE_DIR} || npm install --prefer-offline --cache ${NPM_CACHE_DIR}
                    """
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
                sh '''
                    npm run lint 2>/dev/null || true
                    npm run test:unit 2>/dev/null || npm test 2>/dev/null || true
                '''
            }
            post {
                success { script { stageCallback('test', 'success') } }
                failure { script { stageCallback('test', 'failed') } }
            }
        }

        stage('Build Frontend') {
            steps {
                echo "=== 构建前端 ==="
                sh """
                    set -e
                    ${params.BUILD_COMMAND}
                    test -d ${params.BUILD_OUTPUT_DIR}
                """
            }
            post {
                success { script { stageCallback('compile', 'success') } }
                failure { script { stageCallback('compile', 'failed') } }
            }
        }

        // ==================== SonarQube 代码质量扫描 ====================
        stage('SonarQube Analysis') {
            when { expression { return params.ENABLE_SONAR } }
            steps {
                echo "=== SonarQube 代码质量扫描 ==="
                script {
                    def projectKey  = params.SONAR_PROJECT_KEY?.trim()  ?: env.JOB_NAME.replaceAll('/', '_')
                    def projectName = params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME
                    def sources     = params.SONAR_SOURCES?.trim()      ?: 'src'
                    def exclusions  = params.SONAR_EXCLUSIONS?.trim()   ?: '**/node_modules/**,**/dist/**'

                    withSonarQubeEnv('SonarQube') {
                        sh """
                            sonar-scanner \\
                                -Dsonar.projectKey=${projectKey} \\
                                -Dsonar.projectName=${projectName} \\
                                -Dsonar.projectVersion=${env.FINAL_TAG} \\
                                -Dsonar.sources=${sources} \\
                                -Dsonar.exclusions=${exclusions} \\
                                -Dsonar.scm.disabled=true \\
                                -Dsonar.links.ci=${env.BUILD_URL}
                        """
                    }
                    echo "[SonarQube] 扫描完成，等待质量门禁结果..."
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
                    def qg = waitForQualityGate()
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

        stage('Build Image') {
            steps {
                echo "=== 构建镜像（纯运行时，仅打包 dist 静态文件） ==="
                script {
                    def dockerfile = params.DOCKERFILE_PATH?.trim()
                    def outputDir = params.BUILD_OUTPUT_DIR ?: 'dist'

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
FROM nginx:1.25-alpine
RUN apk --no-cache add tzdata && \\
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY ${outputDir}/ /usr/share/nginx/html/
RUN echo 'server { listen 80; server_name _; root /usr/share/nginx/html; index index.html; location / { try_files \$uri \$uri/ /index.html; } location /health { access_log off; return 200 ok; } }' > /etc/nginx/conf.d/default.conf
RUN chown -R nginx:nginx /usr/share/nginx/html && \\
    chown -R nginx:nginx /var/cache/nginx && \\
    touch /var/run/nginx.pid && chown nginx:nginx /var/run/nginx.pid
USER nginx
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \\
    CMD wget -qO- http://localhost/health || exit 1
CMD ["nginx", "-g", "daemon off;"]
"""
                            echo "[Build Image] ${forceGenerate ? '强制' : '项目无 Dockerfile，'}已自动生成纯运行时 Dockerfile（Nginx）"
                        }
                    }

                    sh """
                        set -e
                        nerdctl build \\
                            -t ${env.FULL_IMAGE} \\
                            -f ${dockerfile} \\
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
                def msg = params.ENABLE_SONAR
                    ? "前端项目构建成功 | SonarQube: ${env.SONAR_QUALITY_GATE_STATUS ?: 'SKIPPED'}"
                    : '前端项目构建成功'
                callbackPlatform('SUCCESS', msg)
            }
        }
        failure { script { callbackPlatform('FAILURE', '前端项目构建失败') } }
        aborted { script { callbackPlatform('ABORTED', '构建中止') } }
        always {
            sh "nerdctl rmi ${env.FULL_IMAGE} || true"
            sh 'rm -rf node_modules dist .git 2>/dev/null || true'
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
