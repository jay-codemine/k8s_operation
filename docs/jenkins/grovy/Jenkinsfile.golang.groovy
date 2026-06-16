 比较// Jenkinsfile - Go 项目（containerd + nerdctl + digest 部署）
// =============================================================
// 架构：Jenkins 虚拟机 -> nerdctl 构建 -> 推送 Harbor -> 回调平台
// 镜像标签：自动生成 {branch}-{commit}-{timestamp}，支持手动覆盖
// 部署方式：使用 image@digest 确保一致性
// =============================================================

pipeline {
    agent any
    
    options {
        timeout(time: 30, unit: 'MINUTES')
        disableConcurrentBuilds()
        buildDiscarder(logRotator(numToKeepStr: '20'))
    }
    
    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库（如 harbor.example.com/project/app）')
        string(name: 'IMAGE_TAG', defaultValue: '', description: '镜像标签（空则自动生成：branch-commit-timestamp）')
        string(name: 'DOCKERFILE_PATH', defaultValue: 'Dockerfile', description: 'Dockerfile 路径')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'PIPELINE_ID', defaultValue: '', description: '流水线 ID')
        string(name: 'RELEASE_ID', defaultValue: '', description: '发布单 ID')
        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过测试')
    }
    
    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
        GOPROXY = 'https://goproxy.cn,direct'
        CGO_ENABLED = '0'
    }
    
    stages {
        stage('Checkout') {
            steps {
                echo "=== 阶段1: 拉取代码 ==="
                
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.GIT_BRANCH}"]],
                    userRemoteConfigs: [[url: params.GIT_REPO]],
                    extensions: [[$class: 'CleanBeforeCheckout']]
                ])
                
                script {
                    // 获取 Git 信息
                    env.GIT_COMMIT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_FULL = sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_MSG = sh(script: 'git log -1 --pretty=%B | head -1', returnStdout: true).trim()
                    env.GIT_BRANCH_SAFE = params.GIT_BRANCH.replaceAll('/', '-')
                    
                    // 生成时间戳
                    env.BUILD_TS = sh(script: 'date +%Y%m%d%H%M%S', returnStdout: true).trim()
                    
                    // 自动生成镜像标签：branch-commit-timestamp
                    env.FINAL_TAG = params.IMAGE_TAG?.trim() ?: "${env.GIT_BRANCH_SAFE}-${env.GIT_COMMIT}-${env.BUILD_TS}"
                    env.FULL_IMAGE = "${params.IMAGE_REPO}:${env.FINAL_TAG}"
                    
                    echo "Git Commit: ${env.GIT_COMMIT}"
                    echo "Image: ${env.FULL_IMAGE}"
                }
            }
        }
        
        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 阶段2: 单元测试 ==="
                sh 'go test -v -cover ./...'
            }
        }
        
        stage('Build Image') {
            steps {
                echo "=== 阶段3: 构建镜像 ==="
                
                script {
                    sh """
                        nerdctl build \\
                            -t ${env.FULL_IMAGE} \\
                            -f ${params.DOCKERFILE_PATH} \\
                            --label git.commit=${env.GIT_COMMIT_FULL} \\
                            --label git.branch=${env.GIT_BRANCH_SAFE} \\
                            --label build.number=${env.BUILD_NUMBER} \\
                            --label build.timestamp=${env.BUILD_TS} \\
                            .
                    """
                }
            }
        }
        
        stage('Push Image') {
            steps {
                echo "=== 阶段4: 推送镜像到 Harbor ==="
                
                script {
                    def registryHost = params.IMAGE_REPO.split('/')[0]
                    
                    // 登录 Harbor
                    sh """
                        echo \${REGISTRY_CREDS_PSW} | nerdctl login -u \${REGISTRY_CREDS_USR} --password-stdin ${registryHost}
                    """
                    
                    // 推送镜像
                    sh "nerdctl push ${env.FULL_IMAGE}"
                    
                    // 获取镜像 Digest
                    env.IMAGE_DIGEST = sh(
                        script: """nerdctl inspect ${env.FULL_IMAGE} --format '{{range .RepoDigests}}{{.}}{{end}}' 2>/dev/null | grep -oE 'sha256:[a-f0-9]+' | head -1 || echo ''""",
                        returnStdout: true
                    ).trim()
                    
                    // 构建带 digest 的完整镜像地址（用于部署）
                    env.IMAGE_WITH_DIGEST = env.IMAGE_DIGEST ? "${params.IMAGE_REPO}@${env.IMAGE_DIGEST}" : env.FULL_IMAGE
                    
                    echo "镜像推送成功"
                    echo "Image Tag: ${env.FULL_IMAGE}"
                    echo "Image Digest: ${env.IMAGE_DIGEST}"
                    echo "Deploy Image: ${env.IMAGE_WITH_DIGEST}"
                }
            }
        }
    }
    
    post {
        success {
            echo "=== 构建成功，回调平台触发部署 ==="
            script {
                callbackPlatform('SUCCESS', 'Go 项目镜像构建推送成功')
            }
        }
        failure {
            echo "=== 构建失败，回调平台 ==="
            script {
                callbackPlatform('FAILURE', "Go 项目构建失败: ${currentBuild.result}")
            }
        }
        aborted {
            echo "=== 构建中止，回调平台 ==="
            script {
                callbackPlatform('ABORTED', '构建被用户中止')
            }
        }
        always {
            sh "nerdctl rmi ${env.FULL_IMAGE} || true"
            cleanWs()
        }
    }
}

// 回调平台函数
def callbackPlatform(String status, String message) {
    if (!params.PLATFORM_CALLBACK_URL) {
        echo "未配置回调地址，跳过回调"
        return
    }
    
    def payload = [
        // Jenkins 构建信息
        job_name: env.JOB_NAME,
        build_number: env.BUILD_NUMBER as Integer,
        status: status,
        duration: (currentBuild.duration / 1000) as Integer,
        message: message,
        
        // Git 信息
        git_repo: params.GIT_REPO,
        git_branch: params.GIT_BRANCH,
        git_commit: env.GIT_COMMIT ?: '',
        git_commit_full: env.GIT_COMMIT_FULL ?: '',
        git_commit_msg: env.GIT_COMMIT_MSG ?: '',
        
        // 镜像信息（核心）
        image_repo: params.IMAGE_REPO ?: '',
        image_tag: env.FINAL_TAG ?: '',
        image: env.FULL_IMAGE ?: '',
        image_digest: env.IMAGE_DIGEST ?: '',
        image_with_digest: env.IMAGE_WITH_DIGEST ?: '',  // 用于部署
        
        // 平台关联信息
        pipeline_id: params.PIPELINE_ID ?: '',
        release_id: params.RELEASE_ID ?: '',
        
        // 构建时间戳
        build_timestamp: env.BUILD_TS ?: ''
    ]
    
    try {
        httpRequest(
            url: params.PLATFORM_CALLBACK_URL,
            httpMode: 'POST',
            contentType: 'APPLICATION_JSON',
            requestBody: groovy.json.JsonOutput.toJson(payload),
            validResponseCodes: '200:299',
            consoleLogResponseBody: true
        )
        echo "平台回调成功"
    } catch (Exception e) {
        echo "平台回调失败: ${e.message}"
    }
}
