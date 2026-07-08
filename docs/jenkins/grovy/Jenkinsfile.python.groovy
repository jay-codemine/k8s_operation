// Jenkinsfile - Python 项目（containerd + nerdctl + digest 部署）
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
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库')
        string(name: 'IMAGE_TAG', defaultValue: '', description: '镜像标签（空则自动生成）')
        string(name: 'DOCKERFILE_PATH', defaultValue: 'Dockerfile', description: 'Dockerfile 路径')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'PIPELINE_ID', defaultValue: '', description: '流水线 ID')
        string(name: 'RELEASE_ID', defaultValue: '', description: '发布单 ID')
        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过测试')
    }
    
    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
        PIP_INDEX_URL = 'https://pypi.tuna.tsinghua.edu.cn/simple'
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.GIT_BRANCH}"]],
                    userRemoteConfigs: [[url: params.GIT_REPO]],
                    extensions: [[$class: 'CleanBeforeCheckout']]
                ])
                script {
                    env.GIT_COMMIT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_FULL = sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
                    env.GIT_COMMIT_MSG = sh(script: 'git log -1 --pretty=%B | head -1', returnStdout: true).trim()
                    env.GIT_BRANCH_SAFE = params.GIT_BRANCH.replaceAll('/', '-')
                    env.BUILD_TS = sh(script: 'date +%Y%m%d%H%M%S', returnStdout: true).trim()
                    
                    env.FINAL_TAG = params.IMAGE_TAG?.trim() ?: "${env.GIT_BRANCH_SAFE}-${env.GIT_COMMIT}-${env.BUILD_TS}"
                    env.FULL_IMAGE = "${params.IMAGE_REPO}:${env.FINAL_TAG}"
                    echo "Image: ${env.FULL_IMAGE}"
                }
            }
        }
        
        stage('Lint') {
            steps {
                sh '''
                    pip install flake8 -q
                    flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics || true
                '''
            }
        }
        
        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                sh '''
                    pip install -r requirements.txt -q
                    pip install pytest pytest-cov -q
                    pytest --cov=. tests/ || true
                '''
            }
        }
        
        stage('Build & Push Image') {
            steps {
                script {
                    def registryHost = params.IMAGE_REPO.split('/')[0]
                    sh """
                        echo \${REGISTRY_CREDS_PSW} | nerdctl login -u \${REGISTRY_CREDS_USR} --password-stdin ${registryHost}
                        nerdctl build -t ${env.FULL_IMAGE} -f ${params.DOCKERFILE_PATH} \\
                            --label git.commit=${env.GIT_COMMIT_FULL} \\
                            --label build.timestamp=${env.BUILD_TS} .
                        nerdctl push ${env.FULL_IMAGE}
                    """
                    
                    env.IMAGE_DIGEST = sh(
                        script: """nerdctl inspect ${env.FULL_IMAGE} --format '{{range .RepoDigests}}{{.}}{{end}}' 2>/dev/null | grep -oE 'sha256:[a-f0-9]+' | head -1 || echo ''""",
                        returnStdout: true
                    ).trim()
                    
                    env.IMAGE_WITH_DIGEST = env.IMAGE_DIGEST ? "${params.IMAGE_REPO}@${env.IMAGE_DIGEST}" : env.FULL_IMAGE
                    echo "Deploy Image: ${env.IMAGE_WITH_DIGEST}"
                }
            }
        }
    }
    
    post {
        success { script { callbackPlatform('SUCCESS', 'Python 项目构建成功') } }
        failure { script { callbackPlatform('FAILURE', 'Python 项目构建失败') } }
        aborted { script { callbackPlatform('ABORTED', '构建被中止') } }
        always {
            sh "nerdctl rmi ${env.FULL_IMAGE} || true"
            cleanWs()
        }
    }
}

def callbackPlatform(String status, String message) {
    if (!params.PLATFORM_CALLBACK_URL) return
    
    def payload = [
        job_name: env.JOB_NAME, build_number: env.BUILD_NUMBER as Integer,
        status: status, message: message,
        git_commit: env.GIT_COMMIT ?: '', git_commit_full: env.GIT_COMMIT_FULL ?: '',
        git_commit_msg: env.GIT_COMMIT_MSG ?: '',
        image: env.FULL_IMAGE ?: '', image_tag: env.FINAL_TAG ?: '',
        image_digest: env.IMAGE_DIGEST ?: '', image_with_digest: env.IMAGE_WITH_DIGEST ?: '',
        pipeline_id: params.PIPELINE_ID ?: '', release_id: params.RELEASE_ID ?: '',
        build_timestamp: env.BUILD_TS ?: ''
    ]
    
    try {
        httpRequest(url: params.PLATFORM_CALLBACK_URL, httpMode: 'POST',
            contentType: 'APPLICATION_JSON', requestBody: groovy.json.JsonOutput.toJson(payload),
            validResponseCodes: '200:299')
    } catch (Exception e) { echo "回调失败: ${e.message}" }
}
