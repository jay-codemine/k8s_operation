// ==============================================================================
// K8s Operation Platform - Java/Spring Boot 通用构建模板（K8s Pod Agent + Kaniko 容器化版）
// ==============================================================================
// 设计理念：一个模板服务 100+ Java 项目，所有项目差异通过参数传入
// 运行模式：Jenkins K8s 动态 Pod Agent，每次构建创建独立 Pod，完成后自动销毁
// 镜像构建：使用 Kaniko（无需 Docker daemon、无需特权容器）
// 支持框架：Spring Boot, Spring Cloud, 普通 Maven 项目
//
// ======================== Jenkins Job 配置方式 ========================
// 推荐使用 "Pipeline script from SCM"（版本化管理，自动同步更新）：
//   1. Jenkins → New Item → Pipeline → 命名为 k8s-builder-java
//   2. Pipeline → Definition: Pipeline script from SCM
//   3. SCM: Git → Repository URL: 平台仓库地址
//   4. Script Path: configs/jenkins-templates/java-spring-pipeline.groovy
//
// ======================== K8s 环境要求 ========================
// Jenkins 需安装 Kubernetes Plugin，并配置 K8s Cloud：
//   Manage Jenkins → Clouds → Kubernetes → 配置 K8s API 地址 + 命名空间
// ==============================================================================

pipeline {
    agent {
        kubernetes {
            yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    jenkins-build: java
    java-version: "${params.JAVA_VERSION ?: '17'}"
spec:
  imagePullSecrets:
  - name: aliyun-registry
  containers:
  - name: maven
    image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/maven:3.9-eclipse-temurin-${params.JAVA_VERSION ?: '17'}
    imagePullPolicy: Always
    command: ['sleep', '99d']
    resources:
      requests:
        cpu: '500m'
        memory: '1Gi'
      limits:
        cpu: '2'
        memory: '4Gi'
    env:
    - name: MAVEN_OPTS
      value: '-Xmx1024m -Xms512m -XX:+TieredCompilation -XX:TieredStopAtLevel=1'
    volumeMounts:
    - name: maven-cache
      mountPath: /root/.m2/repository
    - name: workspace-volume
      mountPath: /home/jenkins/agent
  - name: kaniko
    image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/kaniko-executor:debug
    imagePullPolicy: Always
    command: ['sleep', '99d']
    resources:
      requests:
        cpu: '200m'
        memory: '256Mi'
      limits:
        cpu: '1'
        memory: '2Gi'
    volumeMounts:
    - name: workspace-volume
      mountPath: /home/jenkins/agent
  - name: jnlp
    image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/inbound-agent:latest
    imagePullPolicy: Always
  volumes:
  - name: maven-cache
    persistentVolumeClaim:
      claimName: jenkins-maven-cache
  - name: workspace-volume
    emptyDir: {}
"""
        }
    }

    options {
        timeout(time: 45, unit: 'MINUTES')
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

        // 构建参数
        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过单元测试')
        choice(name: 'JAVA_VERSION', choices: ['17', '21', '11', '8'], description: 'Java 版本（同时决定构建 JDK 和运行时镜像：8/11/17/21）')
        string(name: 'MAVEN_GOALS', defaultValue: 'clean package -DskipTests -B', description: 'Maven 构建命令')
        string(name: 'BUILD_DIR', defaultValue: '', description: '构建目录（pom.xml 所在路径，留空则自动检测。支持：根目录、子目录如 backend/、多模块如 services/user-service/）')
        string(name: 'GIT_CREDENTIAL_ID', defaultValue: 'gitee-id', description: 'Git 凭证ID')
        string(name: 'REGISTRY_CREDENTIAL_ID', defaultValue: 'harbor-registry', description: '镜像仓库凭证ID')
        string(name: 'HMAC_CREDENTIAL_ID', defaultValue: 'hmac-secret', description: 'HMAC签名凭证ID')

        // SonarQube 代码质量扫描参数
        booleanParam(name: 'ENABLE_SONAR', defaultValue: true, description: '启用 SonarQube 代码质量扫描')
        string(name: 'SONAR_PROJECT_KEY', defaultValue: '', description: 'SonarQube 项目 Key（空则使用 Job 名称）')
        string(name: 'SONAR_PROJECT_NAME', defaultValue: '', description: 'SonarQube 项目名称（空则使用 Job 名称）')
        string(name: 'SONAR_SOURCES', defaultValue: 'src/main/java', description: '源代码目录')
        string(name: 'SONAR_JAVA_BINARIES', defaultValue: 'target/classes', description: 'Java 编译输出目录')
        string(name: 'SONAR_EXCLUSIONS', defaultValue: '**/test/**,**/generated/**', description: '排除扫描的文件模式')
        booleanParam(name: 'SONAR_QUALITY_GATE', defaultValue: true, description: '启用质量门禁检查（不通过则构建失败）')

        // 平台注入的质量门禁阈值
        string(name: 'SONAR_COVERAGE_THRESHOLD', defaultValue: '80', description: '代码覆盖率阈值（%）')
        string(name: 'SONAR_NEW_BUGS_MAX', defaultValue: '0', description: '新增 Bug 最大允许数')
        string(name: 'SONAR_CODE_SMELLS_MAX', defaultValue: '10', description: '代码异味最大允许数')
        string(name: 'SONAR_VULNERABILITIES_MAX', defaultValue: '0', description: '安全漏洞最大允许数')
        string(name: 'SONAR_DUPLICATIONS_MAX', defaultValue: '3', description: '代码重复率阈值（%）')
        string(name: 'SONAR_GATE_ACTION', defaultValue: 'block', description: '门禁失败策略: block(阻断) | warn(告警) | skip(跳过)')

        // 制品上传参数
        booleanParam(name: 'ENABLE_ARTIFACT_UPLOAD', defaultValue: true, description: '启用制品上传到平台制品库')
    }

    environment {
        REGISTRY_CREDS = credentials("${params.REGISTRY_CREDENTIAL_ID ?: 'harbor-registry'}")
        HMAC_SECRET    = credentials("${params.HMAC_CREDENTIAL_ID ?: 'hmac-secret'}")
    }

    stages {

        stage('Clean Workspace') {
            steps {
                sh '''
                    rm -rf .git 2>/dev/null || true
                    find . -mindepth 1 -maxdepth 1 ! -name ".m2" -exec rm -rf {} + 2>/dev/null || true
                '''
                script {
                    // 语言类型交叉校验
                    def expectedType = 'java'
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
当前模板类型: ${expectedType} (java-spring-pipeline.groovy)

解决方案（二选一）:
  1. 修改 Jenkins Job 的 Script Path 为: configs/jenkins-templates/${correctScript}
  2. 在平台将 Jenkins Job 名称留空，使用自动匹配
""")
                    }

                    if (!params.GIT_REPO?.trim()) { error("GIT_REPO 不能为空") }
                    if (!params.IMAGE_REPO?.trim()) { error("IMAGE_REPO 不能为空") }

                    def targetBranch = params.GIT_BRANCH?.trim() ?: 'main'
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

        // ==================== Maven 编译 + 打包 ====================
        stage('Setup & Compile') {
            steps {
                echo "=== Maven 编译 & 打包 ==="
                container('maven') {
                    script {
                        // 生成阿里云 Maven 镜像 settings.xml（直接用 sh 写入，避免 JNLP 权限问题）
                        def settingsFile = "${env.WORKSPACE}/.m2/settings.xml"
                        sh """
                            mkdir -p ${env.WORKSPACE}/.m2
                            cat > ${env.WORKSPACE}/.m2/settings.xml << 'SETTINGS_EOF'
<?xml version="1.0" encoding="UTF-8"?>
<settings>
  <mirrors>
    <mirror>
      <id>aliyun</id>
      <name>Aliyun Maven Mirror</name>
      <url>https://maven.aliyun.com/repository/public</url>
      <mirrorOf>central</mirrorOf>
    </mirror>
  </mirrors>
</settings>
SETTINGS_EOF
                        """
                        env.MVN_SETTINGS = settingsFile

                        sh 'java -version 2>&1 | head -1'
                        sh 'mvn --version | head -2'

                        // ==================== 智能检测 pom.xml 位置 ====================
                        def buildDir = params.BUILD_DIR?.trim()
                        if (buildDir) {
                            // 用户手动指定了构建目录
                            def pomPath = "${buildDir}/pom.xml"
                            if (!fileExists(pomPath)) {
                                def allPoms = sh(script: "find . -name pom.xml -not -path '*/target/*' -not -path '*/.m2/*' | sort", returnStdout: true).trim()
                                error("BUILD_DIR='${buildDir}' 下找不到 pom.xml\n项目中检测到的 pom.xml:\n${allPoms}\n请修改 BUILD_DIR 为正确路径")
                            }
                            echo "[Build] 使用用户指定目录: ${buildDir}"
                        } else {
                            // 自动检测 pom.xml 位置
                            if (fileExists('pom.xml')) {
                                buildDir = '.'
                                echo '[Build] ✅ 根目录找到 pom.xml'
                            } else {
                                // 搜索整个项目中的 pom.xml（不限深度，排除编译产物和缓存目录）
                                def pomSearch = sh(script: "find . -name pom.xml -not -path '*/target/*' -not -path '*/.m2/*' -not -path '*/node_modules/*' | sort", returnStdout: true).trim()
                                if (!pomSearch) {
                                    error("项目中未找到任何 pom.xml，请确认 Git 仓库和分支是否正确，或在流水线配置中设置 BUILD_DIR")
                                }
                                def pomList = pomSearch.split('\n').collect { it.trim() }.findAll { it }
                                echo "[Build] 检测到 ${pomList.size()} 个 pom.xml:\n${pomSearch}"

                                if (pomList.size() == 1) {
                                    buildDir = pomList[0].replace('/pom.xml', '').replaceAll('^\\./','') ?: '.'
                                } else {
                                    // 多个 pom.xml：优先找包含 spring-boot-maven-plugin 的（可运行模块）
                                    def bootPom = ''
                                    for (pom in pomList) {
                                        def hasBootPlugin = sh(script: "grep -l 'spring-boot-maven-plugin' '${pom}' 2>/dev/null || true", returnStdout: true).trim()
                                        if (hasBootPlugin) { bootPom = pom; break }
                                    }
                                    if (bootPom) {
                                        buildDir = bootPom.replace('/pom.xml', '').replaceAll('^\\./','') ?: '.'
                                        echo "[Build] ✅ 自动检测到 Spring Boot 模块: ${buildDir}"
                                    } else {
                                        // 取最浅层级的 pom.xml
                                        buildDir = pomList[0].replace('/pom.xml', '').replaceAll('^\\./','') ?: '.'
                                        echo "[Build] 使用最浅层级 pom.xml 所在目录: ${buildDir}"
                                    }
                                }
                            }
                        }
                        env.BUILD_DIR = buildDir
                        echo "[Build] 最终构建目录: ${buildDir}"

                        // ==================== 执行 Maven 构建 ====================
                        if (buildDir == '.') {
                            sh "mvn package -DskipTests -B -T 1C -s ${env.MVN_SETTINGS}"
                        } else {
                            sh "mvn package -DskipTests -B -T 1C -s ${env.MVN_SETTINGS} -f ${buildDir}/pom.xml"
                        }
                        archiveArtifacts artifacts: '**/target/*.jar', fingerprint: true, allowEmptyArchive: true

                        // ==================== 智能查找 JAR 产出 ====================
                        def searchBase = (buildDir == '.') ? 'target' : "${buildDir}/target"
                        def jarFile = sh(script: "find ${searchBase} -maxdepth 1 -name '*.jar' ! -name '*-sources.jar' ! -name '*-javadoc.jar' ! -name '*-plain.jar' 2>/dev/null | head -1", returnStdout: true).trim()

                        if (!jarFile) {
                            // 多模块项目：递归搜索所有 target 目录
                            echo "[Build] 在 ${searchBase} 未直接找到 JAR，尝试递归搜索..."
                            jarFile = sh(script: "find . -path '*/target/*.jar' ! -name '*-sources.jar' ! -name '*-javadoc.jar' ! -name '*-plain.jar' ! -path '*/.m2/*' -type f | sort -t/ -k3 | head -1", returnStdout: true).trim()
                        }

                        if (jarFile) {
                            def jarSize = sh(script: "stat -c%s ${jarFile} 2>/dev/null || stat -f%z ${jarFile}", returnStdout: true).trim()
                            echo "[Compile & Package] ✅ 产出: ${jarFile} (${jarSize} bytes)"
                            env.JAR_FILE = jarFile
                        } else {
                            sh "echo '=== 诊断信息 ===' && find . -path '*/target/*.jar' -type f 2>/dev/null || echo '无任何 JAR 文件'"
                            error("Maven 构建未产出 JAR 文件。构建目录: ${buildDir}\n建议: 1.确认pom.xml有<packaging>jar 2.多模块项目请设置BUILD_DIR为可运行模块路径 3.本地mvn package验证")                        }
                    }
                }
            }
            post {
                success { script { stageCallback('dependencies', 'success'); stageCallback('compile', 'success'); stageCallback('build_binary', 'success') } }
                failure { script { stageCallback('dependencies', 'failed'); stageCallback('compile', 'failed'); stageCallback('build_binary', 'failed') } }
            }
        }

        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 单元测试 ==="
                container('maven') {
                    sh "mvn test -B -T 1C -s ${env.MVN_SETTINGS} -Dsurefire.useFile=false"
                }
            }
            post {
                success { script { stageCallback('test', 'success') } }
                failure { script { stageCallback('test', 'failed') } }
                always { junit allowEmptyResults: true, testResults: '**/target/surefire-reports/*.xml' }
            }
        }

        // ==================== SonarQube 代码质量扫描 ====================
        stage('SonarQube Analysis') {
            when { expression { return params.ENABLE_SONAR } }
            steps {
                script {
                    try {
                        echo "=== SonarQube 代码质量扫描 ==="
                        def projectKey  = params.SONAR_PROJECT_KEY?.trim()  ?: env.JOB_NAME.replaceAll('/', '_')
                        def projectName = params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME
                        def sources     = params.SONAR_SOURCES?.trim()      ?: 'src/main/java'
                        def binaries    = params.SONAR_JAVA_BINARIES?.trim() ?: 'target/classes'
                        def exclusions  = params.SONAR_EXCLUSIONS?.trim()   ?: '**/test/**,**/generated/**'

                        withSonarQubeEnv('SonarQube') {
                            container('maven') {
                                sh """
                                    mvn sonar:sonar \
                                        -s ${env.MVN_SETTINGS} \
                                        -Dmaven.main.skip=true \
                                        -DskipTests \
                                        -Dsonar.projectKey=${projectKey} \
                                        -Dsonar.projectName=${projectName} \
                                        -Dsonar.projectVersion=${env.FINAL_TAG} \
                                        -Dsonar.sources=${sources} \
                                        -Dsonar.java.binaries=${binaries} \
                                        -Dsonar.exclusions=${exclusions},**/target/**,**/build/** \
                                        -Dsonar.scm.disabled=true \
                                        -Dsonar.qualitygate.wait=false \
                                        -Dsonar.threads=4 \
                                        -Dsonar.links.ci=${env.BUILD_URL} \
                                        -B
                                """
                            }
                        }
                        echo "[SonarQube] 扫描已提交，等待质量门禁..."
                        stageCallback('sonar', 'success')
                    } catch (e) {
                        echo "[SonarQube] ❌ 扫描失败: ${e.message}"
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
                    expression { return (params.SONAR_GATE_ACTION ?: 'block') != 'skip' }
                }
            }
            steps {
                echo "=== 质量门禁检查（策略: ${params.SONAR_GATE_ACTION ?: 'block'}） ==="
                script {
                    def gateAction = params.SONAR_GATE_ACTION ?: 'block'
                    def qg = waitForQualityGate(webhookSecretId: '', abortPipeline: false)
                    env.SONAR_QUALITY_GATE_STATUS = qg.status

                    def metricsReport = checkPlatformThresholds()
                    env.SONAR_METRICS_REPORT = metricsReport

                    if (qg.status != 'OK') {
                        echo "[Quality Gate] ❌ SonarQube 门禁状态: ${qg.status}"
                        echo "[Quality Gate] 平台阈值检查:\n${metricsReport}"
                        sonarReportCallback(qg.status)
                        if (gateAction == 'block') {
                            error("SonarQube Quality Gate 未通过: ${qg.status}")
                        } else {
                            echo "[Quality Gate] ⚠️ 门禁未通过但策略为 warn，继续构建"
                            currentBuild.result = 'UNSTABLE'
                        }
                    } else {
                        echo "[Quality Gate] ✅ 通过！状态: ${qg.status}"
                        sonarReportCallback(qg.status)
                    }
                }
            }
            post {
                success { script { stageCallback('quality_gate', 'success') } }
                failure { script { stageCallback('quality_gate', 'failed') } }
            }
        }

        // ==================== 制品归档 ====================
        stage('Upload Artifact') {
            when { expression { return params.ENABLE_ARTIFACT_UPLOAD && params.PLATFORM_CALLBACK_URL?.trim() } }
            steps {
                echo "=== 上传制品到平台制品库 ==="
                container('maven') {
                    script {
                        def jarFile = env.JAR_FILE ?: sh(script: "find . -path '*/target/*.jar' ! -name '*-sources.jar' ! -name '*-javadoc.jar' ! -name '*-plain.jar' ! -path '*/.m2/*' -type f | head -1", returnStdout: true).trim()
                        if (!jarFile) { error("[制品上传] 未找到 JAR 文件") }

                        def gzPath = "${jarFile}.gz"
                        sh "gzip -1 -c ${jarFile} > ${gzPath}"

                        def uploadUrl = params.PLATFORM_CALLBACK_URL
                            .replace('/pipeline/callback', '/artifact/upload')
                            .replace('/stage/callback', '/artifact/upload')

                        def curlStatus = sh(script: """
                            set -e
                            curl -s -w '%{http_code}' -o /tmp/artifact_resp.json \
                                -X POST '${uploadUrl}' \
                                -F 'file=@${gzPath}' \
                                -F 'pipeline_id=${params.PIPELINE_ID ?: 0}' \
                                -F 'run_id=${params.RUN_ID ?: 0}' \
                                -F 'build_number=${env.BUILD_NUMBER}' \
                                -F 'version=${env.FINAL_TAG}' \
                                -F 'language_type=java' \
                                -F 'artifact_type=jar' \
                                -F 'git_repo=${params.GIT_REPO}' \
                                -F 'git_branch=${env.GIT_BRANCH_NAME}' \
                                -F 'git_commit=${env.GIT_COMMIT_SHORT}' \
                                --connect-timeout 10 --max-time 300 --retry 2 --retry-delay 5
                        """, returnStdout: true).trim()

                        if (curlStatus.endsWith('200')) {
                            echo "[制品上传] ✅ 上传成功"
                        } else {
                            echo "[制品上传] ❌ 上传失败: HTTP ${curlStatus[-3..-1]}"
                            error("制品上传失败")
                        }
                        sh "rm -f ${gzPath} /tmp/artifact_resp.json 2>/dev/null || true"
                    }
                }
            }
            post {
                success { script { stageCallback('upload_artifact', 'success') } }
                failure { script { stageCallback('upload_artifact', 'failed') } }
            }
        }

        // ==================== 准备构建探针（从平台拉取） ====================
        stage('Prepare Build Agents') {
            steps {
                echo "=== 准备构建探针（自动拉取平台已启用 Agent） ==="
                container('maven') {
                    script {
                        def agentsDir = '.agents'
                        sh "mkdir -p ${agentsDir}"
                        env.AGENT_DOCKER_COPY_LINES = ''
                        env.AGENT_ENV_LINES = ''
                        env.AGENT_JAVA_OPTS = ''
                        def agentsPrepared = []

                        def platformUrl = params.PLATFORM_CALLBACK_URL?.trim()
                        def platformAvailable = false

                        if (platformUrl) {
                            def apiBase = platformUrl.replaceAll('/api/v1/k8s/cicd/pipeline/callback.*', '/api/v1/k8s/cicd')
                            def listUrl = "${apiBase}/agent/by-scope?scope=java"
                            try {
                                def response = sh(script: "wget -q -O - '${listUrl}' 2>/dev/null || curl -s '${listUrl}'", returnStdout: true).trim()
                                if (response) {
                                    def json = new groovy.json.JsonSlurper().parseText(response)
                                    def agentList = json?.data?.list ?: json?.list ?: []
                                    agentList.each { agent ->
                                        def agentName = agent.name
                                        def fileName = agent.file_name ?: "${agentName}.jar"
                                        def destPath = agent.docker_copy_dest ?: "/app/${fileName}"
                                        def localDir = "${agentsDir}/${agentName}"
                                        def localPath = "${localDir}/${fileName}"
                                        sh "mkdir -p ${localDir}"
                                        def downloadUrl = "${apiBase}/agent/download?name=${agentName}"
                                        def dlResult = sh(script: "wget -q -O '${localPath}' '${downloadUrl}' || curl -s -o '${localPath}' '${downloadUrl}'", returnStatus: true)
                                        if (dlResult == 0) {
                                            env.AGENT_DOCKER_COPY_LINES += "COPY ${localPath} ${destPath}\n"
                                            if (agent.env_key && agent.env_value) {
                                                env.AGENT_ENV_LINES += "ENV ${agent.env_key}=\"${agent.env_value}\"\n"
                                                if (agent.env_value?.contains('-javaagent:')) {
                                                    env.AGENT_JAVA_OPTS += "${agent.env_value} "
                                                }
                                            }
                                            agentsPrepared << agentName
                                            platformAvailable = true
                                        }
                                    }
                                }
                            } catch (Exception e) {
                                echo "[Agents] ⚠️ 平台 API 不可用: ${e.message}"
                            }
                        }

                        if (!platformAvailable) {
                            echo "[Agents] 平台不可用，跳过探针注入"
                            env.AGENT_JAVA_OPTS = ''
                        }
                        echo "[Agents] === 准备完成: ${agentsPrepared.join(', ') ?: '无'} ==="
                    }
                }
            }
            post {
                success { script { stageCallback('prepare_agents', 'success') } }
                failure { script { stageCallback('prepare_agents', 'failed') } }
            }
        }

        // ==================== Kaniko 构建 + 推送镜像 ====================
        stage('Build & Push Image') {
            steps {
                echo "=== Kaniko 构建并推送镜像（无需 Docker daemon） ==="
                container('kaniko') {
                    script {
                        def jarFile = env.JAR_FILE ?: sh(script: "find . -path '*/target/*.jar' ! -name '*-sources.jar' ! -name '*-javadoc.jar' ! -name '*-plain.jar' ! -path '*/.m2/*' -type f | head -1", returnStdout: true).trim()
                        if (!jarFile) { error("未找到 JAR 文件，请确认 Maven 构建是否成功") }

                        def dockerfile = params.DOCKERFILE_PATH?.trim()
                        def javaVersion = params.JAVA_VERSION ?: '17'

                        def forceGenerate = (dockerfile == '__PLATFORM_GENERATE__')
                        if (!dockerfile || forceGenerate) {
                            if (fileExists('Dockerfile')) {
                                dockerfile = 'Dockerfile'
                                echo "[Build Image] 检测到项目自带 Dockerfile，优先使用"
                            } else {
                                dockerfile = '.Dockerfile.runtime'
                                def agentCopyLines = env.AGENT_DOCKER_COPY_LINES ?: ''
                                def agentEnvLines = env.AGENT_ENV_LINES ?: ''
                                def agentJavaOpts = env.AGENT_JAVA_OPTS?.trim() ?: ''

                                def otelOptsValue = agentJavaOpts ? """\
${agentJavaOpts} \
-Dotel.service.name=java-app \
-Dotel.traces.exporter=otlp \
-Dotel.metrics.exporter=none \
-Dotel.logs.exporter=none \
-Dotel.exporter.otlp.endpoint=http://otel-collector-monitoring.svc.cluster.local:4318""" : ''

                                def dockerfileContent = """\
FROM registry.cn-hangzhou.aliyuncs.com/k8s-gos/java:${javaVersion}-jre-alpine
ENV TZ=Asia/Shanghai
WORKDIR /app
RUN addgroup -S appgroup && adduser -S -G appgroup appuser
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app
${agentCopyLines}COPY ${jarFile} /app/app.jar
USER appuser
EXPOSE 8080
${agentEnvLines}${otelOptsValue ? "ENV OTEL_OPTS=\"${otelOptsValue}\"\n" : ''}ENV JAVA_OPTS="\\
-XX:MaxRAMPercentage=75.0 \\
-XX:+UseG1GC \\
-XX:+HeapDumpOnOutOfMemoryError \\
-XX:HeapDumpPath=/app/logs \\
-Xlog:gc*:file=/app/logs/gc.log:time,uptime,level \\
-Djava.security.egd=file:/dev/./urandom"
ENTRYPOINT ["sh", "-c", "exec java \$OTEL_OPTS \$JAVA_OPTS -jar /app/app.jar"]
"""
                                sh """
cat > ${dockerfile} << 'DOCKERFILE_EOF'
${dockerfileContent}
DOCKERFILE_EOF
"""
                                echo "[Build Image] 自动生成 Dockerfile（注入 ${agentCopyLines.count('COPY')} 个探针）"
                            }
                        }

                        // 配置镜像仓库认证（writeFile 避免 shell 特殊字符 + sandbox 限制）
                        def registryHost = params.IMAGE_REPO.split('/')[0]
                        def dockerConfigJson = groovy.json.JsonOutput.toJson([
                            auths: [(registryHost): [username: env.REGISTRY_CREDS_USR, password: env.REGISTRY_CREDS_PSW]]
                        ])
                        writeFile file: '.docker-config.json', text: dockerConfigJson
                        sh """
                            mkdir -p /kaniko/.docker
                            cp .docker-config.json /kaniko/.docker/config.json
                            rm -f .docker-config.json
                            echo '[Kaniko Auth] registry=${registryHost} user=${REGISTRY_CREDS_USR}'
                        """

                        // Kaniko 构建 + 推送
                        sh """
                            /kaniko/executor \
                                --context=. \
                                --dockerfile=${dockerfile} \
                                --destination=${env.FULL_IMAGE} \
                                --customPlatform=linux/amd64 \
                                --cache=true \
                                --cache-repo=${registryHost}/kaniko-cache/java \
                                --build-arg JAVA_VERSION=${javaVersion} \
                                --label git.commit=${env.GIT_COMMIT_FULL} \
                                --label git.branch=${env.GIT_BRANCH_NAME} \
                                --label build.number=${env.BUILD_NUMBER} \
                                --label artifact.version=${env.FINAL_TAG} \
                                --label build.mode=k8s-kaniko \
                                --snapshot-mode=redo \
                                --use-new-run
                        """

                        env.IMAGE_DIGEST = ''
                        env.IMAGE_WITH_DIGEST = env.FULL_IMAGE
                        echo "[Build & Push] ✅ 镜像已构建并推送: ${env.FULL_IMAGE}"
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
        success {
            script {
                def msg
                if (!params.ENABLE_SONAR) {
                    msg = 'Java 项目构建成功'
                } else if (env.SONAR_ANALYSIS_FAILED == 'true') {
                    msg = "Java 项目构建失败 | SonarQube: UNAVAILABLE（扫描阶段连接失败，请检查 SonarQube 服务状态）"
                    callbackPlatform('FAILURE', msg)
                    return
                } else {
                    msg = "Java 项目构建成功 | SonarQube: ${env.SONAR_QUALITY_GATE_STATUS ?: 'SKIPPED'}"
                }
                callbackPlatform('SUCCESS', msg)
            }
        }
        failure { script { callbackPlatform('FAILURE', 'Java 项目构建失败') } }
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
        def payload = [job_name: env.JOB_NAME, build_number: env.BUILD_NUMBER as Integer,
            pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0, stage_type: stageType, status: status]
        def body = groovy.json.JsonOutput.toJson(payload)
        def stageUrl = callbackUrl.replace('/pipeline/callback', '/stage/callback')
        def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${stageType}") : ''
        def headers = signature ? [[name: 'X-Signature', value: signature]] : []
        httpRequest(url: stageUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON',
            requestBody: body, customHeaders: headers, validResponseCodes: '100:599', timeout: 10)
    } catch (e) { echo "[阶段回调] 非致命: ${e.message}" }
}

def callbackPlatform(String status, String message) {
    def callbackUrl = getCallbackUrl()
    if (!callbackUrl) { echo "未配置回调地址"; return }
    def payload = [job_name: env.JOB_NAME, build_number: env.BUILD_NUMBER as Integer, status: status,
        pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
        run_id: params.RUN_ID ? params.RUN_ID as Long : 0,
        image_url: env.FULL_IMAGE ?: '', image_digest: env.IMAGE_DIGEST ?: '',
        image_with_digest: env.IMAGE_WITH_DIGEST ?: '', git_commit: env.GIT_COMMIT_SHORT ?: '',
        git_branch: env.GIT_BRANCH_NAME ?: '',
        duration_sec: currentBuild.duration ? (currentBuild.duration / 1000) as Integer : 0,
        message: message, build_url: env.BUILD_URL ?: '']
    def body = groovy.json.JsonOutput.toJson(payload)
    def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:${status}") : ''
    def headers = signature ? [[name: 'X-Signature', value: signature]] : []
    httpRequest(url: callbackUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON',
        requestBody: body, customHeaders: headers, validResponseCodes: '200:299', consoleLogResponseBody: true)
}

def sonarReportCallback(String qualityGateStatus) {
    def callbackUrl = getCallbackUrl()
    if (!callbackUrl) return
    try {
        def projectKey = params.SONAR_PROJECT_KEY?.trim() ?: env.JOB_NAME.replaceAll('/', '_')
        def sonarUrl = callbackUrl.replace('/pipeline/callback', '/pipeline/sonar-callback')
        def metrics = [:]
        withSonarQubeEnv('SonarQube') {
            def apiUrl = "${env.SONAR_HOST_URL}/api/measures/component?component=${projectKey}&metricKeys=bugs,vulnerabilities,code_smells,coverage,duplicated_lines_density,ncloc,security_hotspots,reliability_rating,security_rating,sqale_rating"
            def resp = httpRequest(url: apiUrl, httpMode: 'GET', validResponseCodes: '200', quiet: true)
            def json = readJSON text: resp.content
            json.component?.measures?.each { m -> metrics[m.metric] = m.value }
        }
        def payload = [pipeline_id: params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
            project_key: projectKey, project_name: params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME,
            quality_gate: qualityGateStatus, dashboard_url: "${env.SONAR_HOST_URL}/dashboard?id=${projectKey}",
            bugs: (metrics.bugs ?: '0') as Integer, vulnerabilities: (metrics.vulnerabilities ?: '0') as Integer,
            code_smells: (metrics.code_smells ?: '0') as Integer, coverage: (metrics.coverage ?: '0.0') as Double,
            duplications: (metrics['duplicated_lines_density'] ?: '0.0') as Double,
            lines_of_code: (metrics.ncloc ?: '0') as Integer, security_hotspots: (metrics.security_hotspots ?: '0') as Integer,
            reliability_rating: ratingToLetter((metrics.reliability_rating ?: '1') as Double),
            security_rating: ratingToLetter((metrics.security_rating ?: '1') as Double),
            maintainability_rating: ratingToLetter((metrics.sqale_rating ?: '1') as Double)]
        def body = groovy.json.JsonOutput.toJson(payload)
        def signature = env.HMAC_SECRET?.trim() ? hmacSha256(env.HMAC_SECRET, "${env.JOB_NAME}:${env.BUILD_NUMBER}:sonar") : ''
        def headers = signature ? [[name: 'X-Signature', value: signature]] : []
        httpRequest(url: sonarUrl, httpMode: 'POST', contentType: 'APPLICATION_JSON',
            requestBody: body, customHeaders: headers, validResponseCodes: '100:599', timeout: 15)
    } catch (e) { echo "[SonarQube] 指标回传非致命错误: ${e.message}" }
}

def checkPlatformThresholds() {
    def report = []
    try {
        def projectKey = params.SONAR_PROJECT_KEY?.trim() ?: env.JOB_NAME.replaceAll('/', '_')
        def metrics = [:]
        withSonarQubeEnv('SonarQube') {
            def apiUrl = "${env.SONAR_HOST_URL}/api/measures/component?component=${projectKey}&metricKeys=coverage,bugs,code_smells,vulnerabilities,duplicated_lines_density"
            def resp = httpRequest(url: apiUrl, httpMode: 'GET', validResponseCodes: '200', quiet: true)
            def json = readJSON text: resp.content
            json.component?.measures?.each { m -> metrics[m.metric] = m.value }
        }
        def actualCoverage = (metrics.coverage ?: '0') as Double
        def thresholdCoverage = (params.SONAR_COVERAGE_THRESHOLD ?: '80') as Double
        report << "  覆盖率: ${actualCoverage}% (${actualCoverage >= thresholdCoverage ? '✅' : '❌'} 阈值: ≥${thresholdCoverage}%)"
        def actualBugs = (metrics.bugs ?: '0') as Integer; def maxBugs = (params.SONAR_NEW_BUGS_MAX ?: '0') as Integer
        report << "  Bug: ${actualBugs} (${actualBugs <= maxBugs ? '✅' : '❌'} 阈值: ≤${maxBugs})"
        def actualSmells = (metrics.code_smells ?: '0') as Integer; def maxSmells = (params.SONAR_CODE_SMELLS_MAX ?: '10') as Integer
        report << "  异味: ${actualSmells} (${actualSmells <= maxSmells ? '✅' : '❌'} 阈值: ≤${maxSmells})"
        def actualVulns = (metrics.vulnerabilities ?: '0') as Integer; def maxVulns = (params.SONAR_VULNERABILITIES_MAX ?: '0') as Integer
        report << "  漏洞: ${actualVulns} (${actualVulns <= maxVulns ? '✅' : '❌'} 阈值: ≤${maxVulns})"
        def actualDup = (metrics['duplicated_lines_density'] ?: '0') as Double; def maxDup = (params.SONAR_DUPLICATIONS_MAX ?: '3') as Double
        report << "  重复率: ${actualDup}% (${actualDup <= maxDup ? '✅' : '❌'} 阈值: ≤${maxDup}%)"
        def allPass = (actualCoverage >= thresholdCoverage) && (actualBugs <= maxBugs) && (actualSmells <= maxSmells) && (actualVulns <= maxVulns) && (actualDup <= maxDup)
        report.add(0, allPass ? "✅ 平台阈值全部通过" : "❌ 部分指标未达平台阈值")
    } catch (e) { report << "  ⚠️ 指标查询失败: ${e.message}" }
    return report.join('\n')
}

def ratingToLetter(Double rating) { if (rating <= 1.0) return 'A'; if (rating <= 2.0) return 'B'; if (rating <= 3.0) return 'C'; if (rating <= 4.0) return 'D'; return 'E' }

def hmacSha256(String secret, String data) {
    def result = ''
    withEnv(["SIGN_SECRET=${secret}", "SIGN_DATA=${data}"]) {
        result = sh(script: 'set +x && printf "%s" "$SIGN_DATA" | openssl dgst -sha256 -hmac "$SIGN_SECRET" | awk \'{print $2}\'', returnStdout: true).trim()
    }
    return result
}
