// Jenkinsfile - Vue + Nginx 前端项目 CI/CD
// =============================================================
// 架构：
//   Jenkins（虚拟机 + containerd）
//     -> checkout(平台传入 repo/branch)
//     -> npm install & build
//     -> nerdctl build
//     -> push
//     -> 回调平台
//
// 修复点：HMAC 使用 openssl 计算，避免 Groovy Sandbox 拦截
// =============================================================

pipeline {
    agent any

    options {
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '20'))

        // 禁用默认 SCM checkout（我们自己拉）
        skipDefaultCheckout(true)
    }

    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址（必填）')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')

        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库地址（必填）')
        string(name: 'IMAGE_TAG', defaultValue: '', description: '镜像标签（空自动生成）')
        string(name: 'DOCKERFILE_PATH', defaultValue: 'Dockerfile', description: 'Dockerfile 路径')

        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'PIPELINE_ID', defaultValue: '', description: '平台流水线ID')
    }

    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
        HMAC_SECRET    = credentials('hmac-secret')
        NPM_REGISTRY   = 'https://registry.npmmirror.com'
    }

    stages {

        // ==================== 1. 拉取代码 ====================
        stage('Checkout') {
            steps {
                echo "=== 拉取代码 ==="

                script {
                    if (!params.GIT_REPO?.trim()) {
                        error("GIT_REPO 不能为空")
                    }
                    if (!params.IMAGE_REPO?.trim()) {
                        error("IMAGE_REPO 不能为空")
                    }
                }

                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.GIT_BRANCH}"]],
                    userRemoteConfigs: [[
                        url: params.GIT_REPO,
                        credentialsId: 'gitee-id'
                    ]],
                    extensions: [
                        [$class: 'CleanBeforeCheckout'],
                        [$class: 'PruneStaleBranch']
                    ]
                ])

                script {
                    env.GIT_COMMIT_SHORT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_FULL  = sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_MSG   = sh(script: 'git log -1 --pretty=%B | head -1', returnStdout: true).trim()

                    env.GIT_BRANCH_NAME  = params.GIT_BRANCH.replaceAll('/', '-')
                    env.BUILD_TS = sh(script: 'date +%Y%m%d%H%M%S', returnStdout: true).trim()

                    env.FINAL_TAG = params.IMAGE_TAG?.trim()
                        ? params.IMAGE_TAG.trim()
                        : "${env.GIT_BRANCH_NAME}-${env.GIT_COMMIT_SHORT}-${env.BUILD_TS}"

                    env.FULL_IMAGE = "${params.IMAGE_REPO}:${env.FINAL_TAG}"

                    echo "Commit: ${env.GIT_COMMIT_SHORT}"
                    echo "Image: ${env.FULL_IMAGE}"
                }
            }
        }

        // ==================== 2. 前端构建 ====================
        stage('Build Frontend') {
            steps {
                echo "=== 编译前端 ==="

                script {
                    if (!fileExists('package.json')) {
                        echo "未检测到 package.json，跳过"
                        return
                    }

                    sh """
                        set -e
                        npm config set registry ${NPM_REGISTRY}
                        npm ci --prefer-offline || npm install
                        npm run build
                        test -d dist
                    """
                }
            }
        }

        // ==================== 3. 构建镜像 ====================
        stage('Build Image') {
            steps {
                echo "=== 构建镜像 ==="

                sh """
                    set -e
                    nerdctl build \
                        -t ${env.FULL_IMAGE} \
                        -f ${params.DOCKERFILE_PATH} \
                        --label git.commit=${env.GIT_COMMIT_FULL} \
                        --label git.branch=${env.GIT_BRANCH_NAME} \
                        --label build.number=${env.BUILD_NUMBER} \
                        --label build.timestamp=${env.BUILD_TS} \
                        .
                """
            }
        }

        // ==================== 4. 推送镜像 ====================
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
                        script: """
                            nerdctl inspect ${env.FULL_IMAGE} \
                            --format '{{range .RepoDigests}}{{println .}}{{end}}' \
                            2>/dev/null | grep -oE 'sha256:[a-f0-9]+' | head -1 || echo ''
                        """,
                        returnStdout: true
                    ).trim()

                    echo "Digest: ${env.IMAGE_DIGEST ?: '未获取到'}"
                }
            }
        }
    }

    post {
        success {
            script { callbackPlatform('SUCCESS', '构建成功') }
        }
        failure {
            script { callbackPlatform('FAILURE', '构建失败') }
        }
        aborted {
            script { callbackPlatform('ABORTED', '构建中止') }
        }
        always {
            sh "nerdctl rmi ${env.FULL_IMAGE} || true"
            cleanWs()
        }
    }
}

// ==================== 回调函数 ====================
def callbackPlatform(String status, String message) {

    if (!params.PLATFORM_CALLBACK_URL?.trim()) {
        echo "未配置回调地址"
        return
    }

    def payload = [
        job_name     : env.JOB_NAME,
        build_number : env.BUILD_NUMBER as Integer,
        status       : status,
        pipeline_id  : params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
        image        : env.FULL_IMAGE ?: '',
        image_digest : env.IMAGE_DIGEST ?: '',
        duration     : currentBuild.duration ? (currentBuild.duration / 1000) as Integer : 0,
        message      : message
    ]

    def body = groovy.json.JsonOutput.toJson(payload)

    def signature = ''
    if (env.HMAC_SECRET?.trim()) {
        signature = hmacSha256(env.HMAC_SECRET, body)
    }

    def headers = signature ? [[name: 'X-Signature', value: signature]] : []

    httpRequest(
        url: params.PLATFORM_CALLBACK_URL,
        httpMode: 'POST',
        contentType: 'APPLICATION_JSON',
        requestBody: body,
        customHeaders: headers,
        validResponseCodes: '200:299',
        consoleLogResponseBody: true
    )
}

// ==================== HMAC（openssl版，不触发Sandbox） ====================
def hmacSha256(String secret, String data) {
    return sh(
        script: """
            printf '%s' '${data.replace("'", "'\\\\''")}' \
            | openssl dgst -sha256 -hmac '${secret}' \
            | awk '{print \$2}'
        """,
        returnStdout: true
    ).trim()
}