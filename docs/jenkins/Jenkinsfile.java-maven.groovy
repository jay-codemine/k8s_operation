// =============================================================================
// Jenkinsfile - Java Maven 项目完整 CICD 流水线
// =============================================================================
// 功能：代码拉取 → 编译 → 测试 → 打包 → 构建镜像 → 推送仓库 → 部署 K8s → 回调平台
// 适用：Spring Boot / Spring Cloud 项目
// =============================================================================

pipeline {
    agent {
        kubernetes {
            yaml '''
apiVersion: v1
kind: Pod
metadata:
  labels:
    jenkins-build: java-maven
spec:
  containers:
  # Maven 构建容器
  - name: maven
    image: maven:3.9-eclipse-temurin-17
    command: ['cat']
    tty: true
    volumeMounts:
    - name: maven-cache
      mountPath: /root/.m2
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "2"
        memory: "4Gi"
  
  # BuildKit 镜像构建容器（适配 containerd）
  - name: buildkit
    image: moby/buildkit:latest
    command: ['cat']
    tty: true
    securityContext:
      privileged: true
    volumeMounts:
    - name: buildkit-socket
      mountPath: /run/buildkit
  
  # kubectl 部署容器
  - name: kubectl
    image: bitnami/kubectl:latest
    command: ['cat']
    tty: true
  
  volumes:
  - name: maven-cache
    persistentVolumeClaim:
      claimName: maven-cache-pvc  # 可选：Maven 缓存 PVC
  - name: buildkit-socket
    emptyDir: {}
'''
        }
    }
    
    // ==================== 平台传入的参数 ====================
    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库地址 (如 harbor.example.com/library/my-app)')
        string(name: 'IMAGE_TAG', defaultValue: 'latest', description: '镜像标签')
        string(name: 'NAMESPACE', defaultValue: 'default', description: 'K8s 命名空间')
        string(name: 'DEPLOYMENT_NAME', defaultValue: '', description: 'Deployment 名称')
        string(name: 'CONTAINER_NAME', defaultValue: '', description: '容器名称')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'PLATFORM_TOKEN', defaultValue: '', description: '平台认证Token')
        string(name: 'PIPELINE_ID', defaultValue: '', description: '流水线ID')
        string(name: 'BUILD_ID', defaultValue: '', description: '发布单关联的构建ID')
        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '是否跳过测试')
        booleanParam(name: 'SKIP_DEPLOY', defaultValue: false, description: '是否跳过部署（仅构建镜像）')
    }
    
    environment {
        // 镜像仓库凭据（从 Jenkins Credentials 获取）
        REGISTRY_CREDS = credentials('harbor-registry')
        // K8s kubeconfig（从 Jenkins Credentials 获取）
        KUBECONFIG_CREDS = credentials('k8s-kubeconfig')
        // 构建产物
        JAR_FILE = ''
        IMAGE_DIGEST = ''
    }
    
    stages {
        // ==================== 阶段1：代码检出 ====================
        stage('Checkout') {
            steps {
                echo "📦 [1/7] 拉取代码: ${params.GIT_REPO} @ ${params.GIT_BRANCH}"
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.GIT_BRANCH}"]],
                    userRemoteConfigs: [[url: params.GIT_REPO]]
                ])
                
                script {
                    env.GIT_COMMIT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_MSG = sh(script: 'git log -1 --pretty=%B', returnStdout: true).trim()
                    echo "Git Commit: ${env.GIT_COMMIT}"
                    echo "Commit Message: ${env.GIT_COMMIT_MSG}"
                }
            }
        }
        
        // ==================== 阶段2：编译 ====================
        stage('Compile') {
            steps {
                container('maven') {
                    echo "🔨 [2/7] Maven 编译..."
                    sh '''
                        mvn clean compile -DskipTests \
                            -Dmaven.repo.local=/root/.m2/repository \
                            --batch-mode --errors
                    '''
                }
            }
        }
        
        // ==================== 阶段3：单元测试 ====================
        stage('Test') {
            when {
                expression { return !params.SKIP_TESTS }
            }
            steps {
                container('maven') {
                    echo "🧪 [3/7] 运行单元测试..."
                    sh '''
                        mvn test \
                            -Dmaven.repo.local=/root/.m2/repository \
                            --batch-mode --errors
                    '''
                }
            }
            post {
                always {
                    // 收集测试报告
                    junit allowEmptyResults: true, testResults: '**/target/surefire-reports/*.xml'
                }
            }
        }
        
        // ==================== 阶段4：打包 ====================
        stage('Package') {
            steps {
                container('maven') {
                    echo "📦 [4/7] Maven 打包..."
                    sh '''
                        mvn package -DskipTests \
                            -Dmaven.repo.local=/root/.m2/repository \
                            --batch-mode --errors
                    '''
                    
                    script {
                        // 查找生成的 JAR 文件
                        env.JAR_FILE = sh(
                            script: 'find target -name "*.jar" -not -name "*-sources.jar" -not -name "*-javadoc.jar" | head -1',
                            returnStdout: true
                        ).trim()
                        echo "打包产物: ${env.JAR_FILE}"
                    }
                }
            }
        }
        
        // ==================== 阶段5：构建镜像 ====================
        stage('Build Image') {
            steps {
                container('buildkit') {
                    echo "🐳 [5/7] 构建 Docker 镜像: ${params.IMAGE_REPO}:${params.IMAGE_TAG}"
                    
                    sh '''
                        # 启动 buildkitd
                        buildkitd --addr unix:///run/buildkit/buildkitd.sock &
                        sleep 3
                        
                        # 登录镜像仓库
                        echo "${REGISTRY_CREDS_PSW}" | buildctl \
                            --addr unix:///run/buildkit/buildkitd.sock \
                            build \
                            --frontend dockerfile.v0 \
                            --local context=. \
                            --local dockerfile=. \
                            --output type=image,name=${IMAGE_REPO}:${IMAGE_TAG},push=true \
                            --export-cache type=inline \
                            --import-cache type=registry,ref=${IMAGE_REPO}:buildcache \
                            --metadata-file /tmp/metadata.json
                        
                        # 提取镜像 digest
                        cat /tmp/metadata.json
                    '''
                    
                    script {
                        // 获取镜像 digest（用于回调平台）
                        env.IMAGE_DIGEST = sh(
                            script: "cat /tmp/metadata.json | grep -o 'sha256:[a-f0-9]*' | head -1 || echo ''",
                            returnStdout: true
                        ).trim()
                        echo "镜像 Digest: ${env.IMAGE_DIGEST}"
                    }
                }
            }
        }
        
        // ==================== 阶段6：推送制品库（可选）====================
        stage('Push to Artifact Repo') {
            when {
                expression { return false }  // 默认跳过，按需开启
            }
            steps {
                container('maven') {
                    echo "📤 [6/7] 推送到制品库 (Nexus/Artifactory)..."
                    sh '''
                        mvn deploy -DskipTests \
                            -Dmaven.repo.local=/root/.m2/repository \
                            --batch-mode
                    '''
                }
            }
        }
        
        // ==================== 阶段7：部署到 K8s ====================
        stage('Deploy to K8s') {
            when {
                expression { return !params.SKIP_DEPLOY && params.DEPLOYMENT_NAME != '' }
            }
            steps {
                container('kubectl') {
                    echo "🚀 [7/7] 部署到 K8s: ${params.NAMESPACE}/${params.DEPLOYMENT_NAME}"
                    
                    writeFile file: '/tmp/kubeconfig', text: env.KUBECONFIG_CREDS
                    
                    sh """
                        export KUBECONFIG=/tmp/kubeconfig
                        
                        # 更新 Deployment 镜像
                        kubectl set image deployment/${params.DEPLOYMENT_NAME} \
                            ${params.CONTAINER_NAME}=${params.IMAGE_REPO}:${params.IMAGE_TAG} \
                            -n ${params.NAMESPACE}
                        
                        # 等待 rollout 完成
                        kubectl rollout status deployment/${params.DEPLOYMENT_NAME} \
                            -n ${params.NAMESPACE} \
                            --timeout=300s
                        
                        # 输出部署结果
                        echo "========== 部署结果 =========="
                        kubectl get deployment ${params.DEPLOYMENT_NAME} -n ${params.NAMESPACE} -o wide
                        kubectl get pods -n ${params.NAMESPACE} -l app=${params.DEPLOYMENT_NAME} -o wide
                    """
                }
            }
        }
    }
    
    // ==================== 构建后处理 ====================
    post {
        success {
            echo "✅ 构建部署成功！"
            script {
                callbackPlatform('SUCCESS', '构建部署成功')
            }
        }
        failure {
            echo "❌ 构建部署失败！"
            script {
                callbackPlatform('FAILURE', "构建失败: ${currentBuild.result}")
            }
        }
        aborted {
            echo "⏹️ 构建被中止！"
            script {
                callbackPlatform('ABORTED', '构建被中止')
            }
        }
        always {
            // 清理工作区
            cleanWs()
        }
    }
}

// ==================== 回调平台函数 ====================
def callbackPlatform(String status, String message) {
    if (!params.PLATFORM_CALLBACK_URL) {
        echo "未配置回调地址，跳过回调"
        return
    }
    
    def payload = """{
        "job_name": "${env.JOB_NAME}",
        "build_number": ${env.BUILD_NUMBER},
        "build_id": ${params.BUILD_ID ?: 0},
        "pipeline_id": ${params.PIPELINE_ID ?: 0},
        "status": "${status}",
        "duration": ${currentBuild.duration / 1000},
        "message": "${message}",
        "git_commit": "${env.GIT_COMMIT}",
        "git_branch": "${params.GIT_BRANCH}",
        "image_repo": "${params.IMAGE_REPO}",
        "image_tag": "${params.IMAGE_TAG}",
        "image_digest": "${env.IMAGE_DIGEST ?: ''}"
    }"""
    
    try {
        httpRequest(
            url: params.PLATFORM_CALLBACK_URL,
            httpMode: 'POST',
            contentType: 'APPLICATION_JSON',
            customHeaders: [
                [name: 'Authorization', value: "Bearer ${params.PLATFORM_TOKEN}"],
                [name: 'Content-Type', value: 'application/json']
            ],
            requestBody: payload,
            validResponseCodes: '200:299'
        )
        echo "回调平台成功"
    } catch (Exception e) {
        echo "回调平台失败: ${e.message}"
    }
}
