// =============================================================================
// Jenkinsfile - Python 项目完整 CICD 流水线
// =============================================================================
// 功能：代码拉取 → 安装依赖 → 测试 → 构建镜像 → 部署 K8s → 回调平台
// 适用：Flask / Django / FastAPI 项目
// =============================================================================

pipeline {
    agent {
        kubernetes {
            yaml '''
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: python
    image: python:3.11-slim
    command: ['cat']
    tty: true
    volumeMounts:
    - name: pip-cache
      mountPath: /root/.cache/pip
  - name: buildkit
    image: moby/buildkit:latest
    command: ['cat']
    tty: true
    securityContext:
      privileged: true
  - name: kubectl
    image: bitnami/kubectl:latest
    command: ['cat']
    tty: true
  volumes:
  - name: pip-cache
    emptyDir: {}
'''
        }
    }
    
    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库地址')
        string(name: 'IMAGE_TAG', defaultValue: 'latest', description: '镜像标签')
        string(name: 'NAMESPACE', defaultValue: 'default', description: 'K8s 命名空间')
        string(name: 'DEPLOYMENT_NAME', defaultValue: '', description: 'Deployment 名称')
        string(name: 'CONTAINER_NAME', defaultValue: '', description: '容器名称')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'PLATFORM_TOKEN', defaultValue: '', description: '平台认证Token')
        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '是否跳过测试')
    }
    
    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
        KUBECONFIG_CREDS = credentials('k8s-kubeconfig')
    }
    
    stages {
        stage('Checkout') {
            steps {
                echo "📦 拉取代码: ${params.GIT_REPO}"
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.GIT_BRANCH}"]],
                    userRemoteConfigs: [[url: params.GIT_REPO]]
                ])
                script {
                    env.GIT_COMMIT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                }
            }
        }
        
        stage('Install Dependencies') {
            steps {
                container('python') {
                    echo "📥 安装依赖..."
                    sh '''
                        pip install --upgrade pip -i https://pypi.tuna.tsinghua.edu.cn/simple
                        pip install -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple
                    '''
                }
            }
        }
        
        stage('Lint') {
            steps {
                container('python') {
                    echo "🔍 代码检查..."
                    sh '''
                        pip install flake8 -i https://pypi.tuna.tsinghua.edu.cn/simple
                        flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics || true
                    '''
                }
            }
        }
        
        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                container('python') {
                    echo "🧪 运行测试..."
                    sh '''
                        pip install pytest pytest-cov -i https://pypi.tuna.tsinghua.edu.cn/simple
                        pytest --cov=. --cov-report=xml tests/ || true
                    '''
                }
            }
        }
        
        stage('Build Image') {
            steps {
                container('buildkit') {
                    echo "🐳 构建镜像: ${params.IMAGE_REPO}:${params.IMAGE_TAG}"
                    sh '''
                        buildctl build \
                            --frontend dockerfile.v0 \
                            --local context=. \
                            --local dockerfile=. \
                            --output type=image,name=${IMAGE_REPO}:${IMAGE_TAG},push=true
                    '''
                }
            }
        }
        
        stage('Deploy') {
            when { expression { return params.DEPLOYMENT_NAME != '' } }
            steps {
                container('kubectl') {
                    echo "🚀 部署到 K8s..."
                    writeFile file: '/tmp/kubeconfig', text: env.KUBECONFIG_CREDS
                    sh """
                        export KUBECONFIG=/tmp/kubeconfig
                        kubectl set image deployment/${params.DEPLOYMENT_NAME} \
                            ${params.CONTAINER_NAME}=${params.IMAGE_REPO}:${params.IMAGE_TAG} \
                            -n ${params.NAMESPACE}
                        kubectl rollout status deployment/${params.DEPLOYMENT_NAME} \
                            -n ${params.NAMESPACE} --timeout=300s
                    """
                }
            }
        }
    }
    
    post {
        success {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        customHeaders: [[name: 'Authorization', value: "Bearer ${params.PLATFORM_TOKEN}"]],
                        requestBody: """{
                            "job_name": "${env.JOB_NAME}",
                            "build_number": ${env.BUILD_NUMBER},
                            "status": "SUCCESS",
                            "image": "${params.IMAGE_REPO}:${params.IMAGE_TAG}",
                            "git_commit": "${env.GIT_COMMIT}"
                        }"""
                    )
                }
            }
        }
        failure {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        customHeaders: [[name: 'Authorization', value: "Bearer ${params.PLATFORM_TOKEN}"]],
                        requestBody: """{"job_name": "${env.JOB_NAME}", "build_number": ${env.BUILD_NUMBER}, "status": "FAILURE"}"""
                    )
                }
            }
        }
    }
}
