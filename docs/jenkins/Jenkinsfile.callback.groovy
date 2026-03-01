// ============================================================
// Jenkinsfile 回调模板（生产级）
// 功能：构建完成后回调平台，更新状态
// ============================================================
// 安全设计：
//   HMAC 密钥不通过参数传递，而是使用 Jenkins Credentials 管理
//   平台端和 Jenkins 端各自配置相同的密钥，避免网络传输泄露
// ============================================================

pipeline {
    agent any
    
    // ==================== 参数定义 ====================
    // 这些参数由平台触发构建时传入（不含敏感信息）
    parameters {
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'PIPELINE_ID', defaultValue: '', description: '流水线ID（平台传入）')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
    }
    
    environment {
        // Harbor 配置
        HARBOR_URL = 'harbor.example.com'
        HARBOR_PROJECT = 'my-project'
        IMAGE_NAME = 'my-app'
        
        // 构建产物（用于回调）
        BUILD_IMAGE = ''
        IMAGE_DIGEST = ''
        
        // HMAC 签名密钥（从 Jenkins Credentials 读取，不通过参数传递）
        // 需要在 Jenkins 中创建 Secret text 类型的凭据，ID 为 'hmac-secret'
        // 该密钥需与平台 config.yaml 中的 Jenkins.HMACSecret 保持一致
        HMAC_SECRET = credentials('hmac-secret')
    }
    
    stages {
        // ==================== 代码检出 ====================
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "${params.GIT_BRANCH}"]],
                    extensions: [],
                    userRemoteConfigs: [[
                        url: "${params.GIT_REPO}",
                        credentialsId: 'gitee-id'  // Jenkins 凭据ID
                    ]]
                ])
            }
        }
        
        // ==================== 构建镜像 ====================
        stage('Build Image') {
            steps {
                script {
                    def imageTag = "${env.BUILD_NUMBER}"
                    env.BUILD_IMAGE = "${HARBOR_URL}/${HARBOR_PROJECT}/${IMAGE_NAME}:${imageTag}"
                    
                    // 使用 nerdctl 构建（containerd 环境）
                    sh """
                        nerdctl build -t ${env.BUILD_IMAGE} .
                    """
                }
            }
        }
        
        // ==================== 推送镜像 ====================
        stage('Push Image') {
            steps {
                script {
                    // 推送到 Harbor
                    sh """
                        nerdctl push ${env.BUILD_IMAGE}
                    """
                    
                    // 获取镜像 digest
                    env.IMAGE_DIGEST = sh(
                        script: "nerdctl inspect ${env.BUILD_IMAGE} --format '{{.RepoDigests}}'",
                        returnStdout: true
                    ).trim()
                }
            }
        }
        
        // ==================== 部署（可选） ====================
        stage('Deploy') {
            when {
                expression { return env.DEPLOY_ENABLED == 'true' }
            }
            steps {
                script {
                    // kubectl 部署逻辑
                    echo "Deploying ${env.BUILD_IMAGE}..."
                }
            }
        }
    }
    
    // ==================== 构建后回调 ====================
    post {
        always {
            script {
                // 只有配置了回调地址才执行
                if (params.PLATFORM_CALLBACK_URL) {
                    sendCallback(currentBuild.result ?: 'FAILURE')
                }
            }
        }
    }
}

// ==================== 回调函数 ====================
def sendCallback(String buildResult) {
    def status = buildResult == 'SUCCESS' ? 'SUCCESS' : 
                 buildResult == 'ABORTED' ? 'ABORTED' : 'FAILURE'
    
    // 计算 HMAC 签名（使用 Credentials 中的密钥，而非参数传入）
    def signature = ''
    if (env.HMAC_SECRET) {
        def signData = "${env.JOB_NAME}${env.BUILD_NUMBER}${status}"
        signature = computeHMAC(env.HMAC_SECRET, signData)
    }
    
    // 构建回调数据
    def callbackData = [
        job_name: env.JOB_NAME,
        build_number: env.BUILD_NUMBER.toInteger(),
        status: status,
        pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID.toLong() : 0,
        image: env.BUILD_IMAGE ?: '',
        image_digest: env.IMAGE_DIGEST ?: '',
        duration: currentBuild.duration ? (currentBuild.duration / 1000).toInteger() : 0,
        message: currentBuild.description ?: ''
    ]
    
    try {
        def response = httpRequest(
            url: params.PLATFORM_CALLBACK_URL,
            httpMode: 'POST',
            contentType: 'APPLICATION_JSON',
            requestBody: groovy.json.JsonOutput.toJson(callbackData),
            customHeaders: [
                [name: 'X-Signature', value: signature]
            ],
            validResponseCodes: '200',
            timeout: 30
        )
        echo "[回调] 成功: ${response.status}"
    } catch (Exception e) {
        echo "[回调] 失败: ${e.message}"
        // 回调失败不影响构建结果，平台会通过轮询兜底
    }
}

// ==================== HMAC-SHA256 签名 ====================
def computeHMAC(String secret, String data) {
    import javax.crypto.Mac
    import javax.crypto.spec.SecretKeySpec
    
    def mac = Mac.getInstance('HmacSHA256')
    mac.init(new SecretKeySpec(secret.bytes, 'HmacSHA256'))
    def hash = mac.doFinal(data.bytes)
    return hash.encodeHex().toString()
}
