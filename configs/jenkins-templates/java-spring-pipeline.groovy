// ==============================================================================
// K8s Operation Platform - Java/Spring Boot 通用构建模板
// ==============================================================================
// 设计理念：一个模板服务 100+ Java 项目，所有项目差异通过参数传入
// 支持框架：Spring Boot, Spring Cloud, 普通 Maven 项目
// 代码扫描：SonarQube 质量门禁集成（可选）
//
// ======================== Jenkins Job 配置方式 ========================
// 推荐使用 "Pipeline script from SCM"（版本化管理，自动同步更新）：
//
//   1. Jenkins → New Item → Pipeline → 命名为 k8s-builder-java
//   2. Pipeline 区域 → Definition 选择: Pipeline script from SCM
//   3. SCM 选择: Git
//   4. Repository URL: 填写平台仓库地址（如 https://gitee.com/your-org/k8s_operation.git）
//   5. Credentials: 选择 Git 凭证（如 gitee-id）
//   6. Branch Specifier: */main
//   7. Script Path: configs/jenkins-templates/java-spring-pipeline.groovy
//   8. 保存即可，后续模板更新自动生效
//
// 备选方式（直接粘贴脚本，不推荐）：
//   Pipeline → Definition: Pipeline script → 粘贴本文件内容
//
// ==============================================================================

pipeline {
    agent any

    // ==================== 工具配置（从 Jenkins 全局工具读取） ====================
    // Jenkins → Manage Jenkins → Tools 中配置：
    //   Maven: 名称填 "Maven-3.9"，指向本地安装路径（如 /opt/apache-maven-3.9.9）
    //   JDK:   名称填 "JDK-21"，指向本地安装路径（如 /usr/lib/jvm/java-21）
    tools {
        maven 'Maven-3.9'
        jdk   'JDK-21'
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
        string(name: 'JAVA_VERSION', defaultValue: '17', description: 'Java 版本')
        string(name: 'MAVEN_GOALS', defaultValue: 'clean package -DskipTests -B', description: 'Maven 构建命令')
        string(name: 'GIT_CREDENTIAL_ID', defaultValue: 'gitee-id', description: 'Git 凭证ID')

        // SonarQube 代码质量扫描参数
        booleanParam(name: 'ENABLE_SONAR', defaultValue: true, description: '启用 SonarQube 代码质量扫描')
        string(name: 'SONAR_PROJECT_KEY', defaultValue: '', description: 'SonarQube 项目 Key（空则使用 Job 名称）')
        string(name: 'SONAR_PROJECT_NAME', defaultValue: '', description: 'SonarQube 项目名称（空则使用 Job 名称）')
        string(name: 'SONAR_SOURCES', defaultValue: 'src/main/java', description: '源代码目录')
        string(name: 'SONAR_JAVA_BINARIES', defaultValue: 'target/classes', description: 'Java 编译输出目录')
        string(name: 'SONAR_EXCLUSIONS', defaultValue: '**/test/**,**/generated/**', description: '排除扫描的文件模式')
        booleanParam(name: 'SONAR_QUALITY_GATE', defaultValue: true, description: '启用质量门禁检查（不通过则构建失败）')

        // 制品上传参数
        booleanParam(name: 'ENABLE_ARTIFACT_UPLOAD', defaultValue: true, description: '启用制品上传到平台制品库')
    }

    environment {
        REGISTRY_CREDS   = credentials('harbor-registry')
        HMAC_SECRET      = credentials('hmac-secret')
        MAVEN_OPTS       = '-Xmx1024m -Xms512m -XX:+TieredCompilation -XX:TieredStopAtLevel=1'
        // Maven 本地仓库放在 workspace 外部，跨构建持久缓存（首次下载后后续秒级复用）
        MAVEN_LOCAL_REPO = '/var/lib/jenkins/.m2/repository'
        // BuildKit 层缓存目录（跨构建持久复用，二次构建仅重建变化层）
        BUILDKIT_CACHE   = '/var/lib/jenkins/.buildkit-cache'
    }

    stages {

        stage('Clean Workspace') {
            steps {
                // 选择性清理：保留外部缓存目录，只清源码区
                sh 'find . -mindepth 1 -maxdepth 1 ! -name ".m2" | xargs rm -rf 2>/dev/null || true'
                script {
                    // 语言类型交叉校验：防止自定义 Job 配错脚本
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

        // ==================== 构建环境检测 ====================
        stage('Setup Build Tools') {
            steps {
                echo "=== 检测构建环境 ==="
                script {
                    // 检测 Java
                    def javaVer = sh(script: 'java -version 2>&1 | head -1', returnStdout: true).trim()
                    echo "[Setup] Java: ${javaVer}"

                    // 检测 Maven
                    def mvnVer = sh(script: 'mvn --version 2>&1 | head -1', returnStatus: true)
                    if (mvnVer != 0) {
                        error("""
=== Maven 未找到 ===
MAVEN_HOME=${MAVEN_HOME}

请在 Jenkins 服务器上安装 Maven：
  cd /opt && wget https://mirrors.aliyun.com/apache/maven/maven-3/3.9.9/binaries/apache-maven-3.9.9-bin.tar.gz && tar xzf apache-maven-3.9.9-bin.tar.gz

安装后确认模板中 MAVEN_HOME 路径与实际安装路径一致。
""")
                    }
                    sh 'mvn --version | head -2'

                    // 生成阿里云 Maven 镜像 settings.xml（加速依赖下载）
                    def settingsFile = "${env.WORKSPACE}/.m2/settings.xml"
                    sh "mkdir -p ${env.WORKSPACE}/.m2"
                    writeFile file: settingsFile, text: """\
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
"""
                    env.MVN_SETTINGS = settingsFile
                    echo "[Setup] 已配置阿里云 Maven 镜像加速"
                }
            }
            post {
                success { script { stageCallback('dependencies', 'success') } }
                failure { script { stageCallback('dependencies', 'failed') } }
            }
        }

        // ==================== 编译 + 打包（一步到位产出 JAR，供制品上传和 Docker 构建复用） ====================
        stage('Compile & Package') {
            steps {
                echo "=== Maven 编译 & 打包（产出 target/*.jar） ==="
                sh "mvn package -DskipTests -B -T 1C -s ${MVN_SETTINGS} -Dmaven.repo.local=${MAVEN_LOCAL_REPO}"
                archiveArtifacts artifacts: '**/target/*.jar', fingerprint: true, allowEmptyArchive: true
                script {
                    // 验证 JAR 产出
                    def jarFile = sh(script: "find target -maxdepth 1 -name '*.jar' ! -name '*-sources.jar' ! -name '*-javadoc.jar' | head -1", returnStdout: true).trim()
                    if (jarFile) {
                        def jarSize = sh(script: "stat -c%s ${jarFile} 2>/dev/null || stat -f%z ${jarFile}", returnStdout: true).trim()
                        echo "[Compile & Package] ✅ 产出: ${jarFile} (${jarSize} bytes)"
                        env.JAR_FILE = jarFile
                    } else {
                        error("Maven package 未产出 JAR 文件，请检查 pom.xml 配置")
                    }
                }
            }
            post {
                success { script { stageCallback('compile', 'success'); stageCallback('build_binary', 'success') } }
                failure { script { stageCallback('compile', 'failed'); stageCallback('build_binary', 'failed') } }
            }
        }

        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                echo "=== 单元测试 ==="
                sh "mvn test -B -T 1C -s ${MVN_SETTINGS} -Dmaven.repo.local=${MAVEN_LOCAL_REPO} -Dsurefire.useFile=false"
            }
            post {
                success { script { stageCallback('test', 'success') } }
                failure { script { stageCallback('test', 'failed') } }
                always { junit allowEmptyResults: true, testResults: '**/target/surefire-reports/*.xml' }
            }
        }

        // ==================== SonarQube 代码质量扫描（性能优化版） ====================
        // 优化要点：
        //   1. -Dsonar.scm.disabled=true  → 禁用 git blame 逐行追溯（节省 60%+ 时间）
        //   2. -Dmaven.repo.local         → 复用已缓存的本地仓库（避免重复下载）
        //   3. -DskipTests               → 跳过测试编译（扫描阶段不需要）
        //   4. -Dsonar.qualitygate.wait=false → 不在 mvn 内等待门禁（由后续阶段处理）
        //   5. -Dmaven.main.skip=true    → 跳过主代码编译（compile 阶段已产出 class）
        //   6. -Dsonar.threads=4         → 启用 4 线程并行分析（多核 CPU 加速 30-50%）
        //   7. exclusions 追加 target/build → 避免扫描编译产物目录
        stage('SonarQube Analysis') {
            when { expression { return params.ENABLE_SONAR } }
            steps {
                script {
                    try {
                        echo "=== SonarQube 代码质量扫描（轻量模式） ==="
                        def projectKey  = params.SONAR_PROJECT_KEY?.trim()  ?: env.JOB_NAME.replaceAll('/', '_')
                        def projectName = params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME
                        def sources     = params.SONAR_SOURCES?.trim()      ?: 'src/main/java'
                        def binaries    = params.SONAR_JAVA_BINARIES?.trim() ?: 'target/classes'
                        def exclusions  = params.SONAR_EXCLUSIONS?.trim()   ?: '**/test/**,**/generated/**'

                        withSonarQubeEnv('SonarQube') {
                            sh """
                                mvn sonar:sonar \\
                                    -s ${env.MVN_SETTINGS} \\
                                    -Dmaven.repo.local=${MAVEN_LOCAL_REPO} \\
                                    -Dmaven.main.skip=true \\
                                    -DskipTests \\
                                    -Dsonar.projectKey=${projectKey} \\
                                    -Dsonar.projectName=${projectName} \\
                                    -Dsonar.projectVersion=${env.FINAL_TAG} \\
                                    -Dsonar.sources=${sources} \\
                                    -Dsonar.java.binaries=${binaries} \\
                                    -Dsonar.exclusions=${exclusions},**/target/**,**/build/** \\
                                    -Dsonar.scm.disabled=true \\
                                    -Dsonar.qualitygate.wait=false \\
                                    -Dsonar.threads=4 \\
                                    -Dsonar.links.ci=${env.BUILD_URL} \\
                                    -B
                            """
                        }
                        echo "[SonarQube] 扫描已提交，等待质量门禁..."
                        stageCallback('sonar', 'success')
                    } catch (e) {
                        echo "[SonarQube] ❌ 扫描失败: ${e.message}"
                        echo "[SonarQube] 常见原因: 1) SonarQube 服务未启动  2) Jenkins 与 SonarQube 网络不通  3) SonarQube Token 过期"
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
                }
            }
            steps {
                echo "=== 质量门禁检查 ==="
                script {
                    // 等待 SonarQube 分析完成（最多 2 分钟，轻量扫描通常 30s 内返回）
                    def qg = waitForQualityGate(webhookSecretId: '', abortPipeline: false)
                    env.SONAR_QUALITY_GATE_STATUS = qg.status
                    if (qg.status != 'OK') {
                        echo "[Quality Gate] 状态: ${qg.status}"
                        echo "[Quality Gate] 代码质量未达标，请登录 SonarQube 查看详细报告"
                        // 即使未通过也回传报告数据，让平台展示具体问题
                        sonarReportCallback(qg.status)
                        error("SonarQube Quality Gate 未通过: ${qg.status}")
                    }
                    echo "[Quality Gate] ✅ 通过！状态: ${qg.status}"
                    // 质量门禁通过后，将 SonarQube 指标数据回传平台
                    sonarReportCallback(qg.status)
                }
            }
            post {
                success { script { stageCallback('quality_gate', 'success') } }
                failure { script { stageCallback('quality_gate', 'failed') } }
            }
        }

        // ==================== 制品上传（可选，JAR 直接上传，无需压缩） ====================
        stage('Upload Artifact') {
            when { expression { return params.ENABLE_ARTIFACT_UPLOAD && params.PLATFORM_CALLBACK_URL?.trim() } }
            steps {
                echo "=== 上传制品到平台制品库 ==="
                script {
                    // 查找 JAR 文件（Compile & Package 阶段已产出）
                    def jarFile = env.JAR_FILE ?: sh(script: "find target -maxdepth 1 -name '*.jar' ! -name '*-sources.jar' ! -name '*-javadoc.jar' | head -1", returnStdout: true).trim()
                    if (!jarFile) { error("[制品上传] 未找到 JAR 文件，Compile & Package 阶段可能异常") }

                    def fileSize = sh(script: "stat -c%s ${jarFile} 2>/dev/null || stat -f%z ${jarFile}", returnStdout: true).trim()
                    echo "[制品上传] 上传文件: ${jarFile} (${fileSize} bytes)"

                    def uploadUrl = params.PLATFORM_CALLBACK_URL
                        .replace('/pipeline/callback', '/artifact/upload')
                        .replace('/stage/callback', '/artifact/upload')

                    def curlStatus = sh(script: """
                        set -e
                        curl -s -w '%{http_code}' -o /tmp/artifact_resp.json \\
                            -X POST '${uploadUrl}' \\
                            -F 'file=@${jarFile}' \\
                            -F 'pipeline_id=${params.PIPELINE_ID ?: 0}' \\
                            -F 'run_id=${params.RUN_ID ?: 0}' \\
                            -F 'build_number=${env.BUILD_NUMBER}' \\
                            -F 'version=${env.FINAL_TAG}' \\
                            -F 'language_type=java' \\
                            -F 'artifact_type=jar' \\
                            -F 'git_repo=${params.GIT_REPO}' \\
                            -F 'git_branch=${env.GIT_BRANCH_NAME}' \\
                            -F 'git_commit=${env.GIT_COMMIT_SHORT}' \\
                            --connect-timeout 10 \\
                            --max-time 120 \\
                            --tcp-nodelay \\
                            -H "Expect:" \\
                            --retry 1
                    """, returnStdout: true).trim()

                    def httpCode = curlStatus[-3..-1]
                    def respBody = sh(script: "cat /tmp/artifact_resp.json 2>/dev/null || echo '{}'", returnStdout: true).trim()
                    sh "rm -f /tmp/artifact_resp.json 2>/dev/null || true"
                    if (httpCode == '200') {
                        echo "[制品上传] ✅ 上传成功，平台制品库已入库"
                        echo "[制品上传] 响应: ${respBody}"
                    } else {
                        echo "[制品上传] ❌ 上传失败 HTTP ${httpCode}"
                        echo "[制品上传] 响应内容: ${respBody}"
                        echo "[制品上传] 上传地址: ${uploadUrl}"
                        error("制品上传失败: HTTP ${httpCode}")
                    }
                }
            }
            post {
                success { script { stageCallback('upload_artifact', 'success') } }
                failure { script { stageCallback('upload_artifact', 'failed') } }
            }
        }

        // ==================== Docker 镜像构建（复用 Compile & Package 阶段已打包的 JAR） ====================
        stage('Build Image') {
            steps {
                echo "=== 构建 Docker 镜像（复用 Compile & Package 产出的 JAR） ==="
                script {
                    // JAR 已在 Compile & Package 阶段产出，此处直接构建镜像
                    def jarFile = env.JAR_FILE ?: sh(script: "find target -maxdepth 1 -name '*.jar' ! -name '*-sources.jar' ! -name '*-javadoc.jar' | head -1", returnStdout: true).trim()
                    if (!jarFile) { error("target/ 下未找到 JAR 文件，请检查 Compile & Package 阶段") }
                    echo "[Build Image] 使用 JAR: ${jarFile}"

                    def dockerfile = params.DOCKERFILE_PATH?.trim()
                    def javaVersion = params.JAVA_VERSION ?: '17'

                    // 优先级：1) 参数指定路径 → 2) 项目自带 Dockerfile → 3) 自动生成
                    // __PLATFORM_GENERATE__ 为平台哨兵值，表示强制使用平台生成
                    def forceGenerate = (dockerfile == '__PLATFORM_GENERATE__')
                    if (!dockerfile || forceGenerate) {
                        // 智能检测模式：检查项目是否自带 Dockerfile（非强制生成时）
                        if (!forceGenerate && fileExists('Dockerfile')) {
                            dockerfile = 'Dockerfile'
                            echo "[Build Image] 检测到项目自带 Dockerfile，直接使用"
                        } else {
                            // 项目无 Dockerfile，自动生成纯运行时版本（阿里云镜像源）
                            dockerfile = '.Dockerfile.runtime'
                            writeFile file: dockerfile, text: """\
FROM registry.cn-hangzhou.aliyuncs.com/k8s-gos/java:${javaVersion}-jre-alpine
ENV TZ=Asia/Shanghai
WORKDIR /app
RUN addgroup -S appgroup && adduser -S -G appgroup appuser
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app
COPY target/*.jar /app/app.jar
USER appuser
EXPOSE 8080
ENV JAVA_OPTS="\\
-XX:MaxRAMPercentage=75.0 \\
-XX:+UseG1GC \\
-XX:+HeapDumpOnOutOfMemoryError \\
-XX:HeapDumpPath=/app/logs \\
-Djava.security.egd=file:/dev/./urandom"
ENTRYPOINT ["sh", "-c", "exec java \$JAVA_OPTS -jar /app/app.jar"]
"""
                            echo "[Build Image] 项目无 Dockerfile，已自动生成纯运行时版本（阿里云 JRE 镜像）"
                        }
                    }

                    // 使用 BuildKit 本地层缓存：首次全量构建，后续仅重建变化层
                    def cacheDir = "${env.BUILDKIT_CACHE}/${env.JOB_NAME}".replaceAll('[^a-zA-Z0-9/_.-]', '_')
                    sh """
                        set -e
                        mkdir -p ${cacheDir}
                        nerdctl build \\
                            -t ${env.FULL_IMAGE} \\
                            -f ${dockerfile} \\
                            --cache-from type=local,src=${cacheDir} \\
                            --cache-to type=local,dest=${cacheDir},mode=max \\
                            --build-arg JAVA_VERSION=${javaVersion} \\
                            --label git.commit=${env.GIT_COMMIT_FULL} \\
                            --label git.branch=${env.GIT_BRANCH_NAME} \\
                            --label build.number=${env.BUILD_NUMBER} \\
                            --label artifact.version=${env.FINAL_TAG} \\
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
                def msg
                if (!params.ENABLE_SONAR) {
                    msg = 'Java 项目构建成功'
                } else if (env.SONAR_ANALYSIS_FAILED == 'true') {
                    msg = "Java 项目构建成功 | SonarQube: UNAVAILABLE（扫描阶段连接失败，请检查 SonarQube 服务状态）"
                } else {
                    msg = "Java 项目构建成功 | SonarQube: ${env.SONAR_QUALITY_GATE_STATUS ?: 'SKIPPED'}"
                }
                callbackPlatform('SUCCESS', msg)
            }
        }
        failure { script { callbackPlatform('FAILURE', 'Java 项目构建失败') } }
        aborted { script { callbackPlatform('ABORTED', '构建中止') } }
        always {
            sh "nerdctl rmi ${env.FULL_IMAGE} || true"
            // 清理源码和 target，但保留外部 Maven 缓存
            sh 'rm -rf target src .git 2>/dev/null || true'
        }
    }
}

// ==================== 回调函数（与 Go 模板完全一致） ====================
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

// ==================== SonarQube 指标回传平台 ====================
// 调用平台 sonar-callback 接口，将扫描结果持久化到运行记录中
def sonarReportCallback(String qualityGateStatus) {
    if (!params.PLATFORM_CALLBACK_URL?.trim()) return
    try {
        def projectKey = params.SONAR_PROJECT_KEY?.trim() ?: env.JOB_NAME.replaceAll('/', '_')
        def sonarUrl = params.PLATFORM_CALLBACK_URL.replace('/pipeline/callback', '/pipeline/sonar-callback')

        // 通过 SonarQube Web API 获取指标数据
        def metrics = [:]
        withSonarQubeEnv('SonarQube') {
            def apiUrl = "${env.SONAR_HOST_URL}/api/measures/component?component=${projectKey}&metricKeys=bugs,vulnerabilities,code_smells,coverage,duplicated_lines_density,ncloc,security_hotspots,reliability_rating,security_rating,sqale_rating"
            def resp = httpRequest(url: apiUrl, httpMode: 'GET', validResponseCodes: '200', quiet: true)
            def json = readJSON text: resp.content
            json.component?.measures?.each { m ->
                metrics[m.metric] = m.value
            }
        }

        def payload = [
            pipeline_id:          params.PIPELINE_ID ? params.PIPELINE_ID as Long : 0,
            project_key:          projectKey,
            project_name:         params.SONAR_PROJECT_NAME?.trim() ?: env.JOB_NAME,
            quality_gate:         qualityGateStatus,
            dashboard_url:        "${env.SONAR_HOST_URL}/dashboard?id=${projectKey}",
            bugs:                 (metrics.bugs ?: '0') as Integer,
            vulnerabilities:      (metrics.vulnerabilities ?: '0') as Integer,
            code_smells:          (metrics.code_smells ?: '0') as Integer,
            coverage:             (metrics.coverage ?: '0.0') as Double,
            duplications:         (metrics['duplicated_lines_density'] ?: '0.0') as Double,
            lines_of_code:        (metrics.ncloc ?: '0') as Integer,
            security_hotspots:    (metrics.security_hotspots ?: '0') as Integer,
            reliability_rating:   ratingToLetter((metrics.reliability_rating ?: '1') as Double),
            security_rating:      ratingToLetter((metrics.security_rating ?: '1') as Double),
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

// SonarQube rating 数值转字母: 1.0=A, 2.0=B, 3.0=C, 4.0=D, 5.0=E
def ratingToLetter(Double rating) {
    if (rating <= 1.0) return 'A'
    if (rating <= 2.0) return 'B'
    if (rating <= 3.0) return 'C'
    if (rating <= 4.0) return 'D'
    return 'E'
}

def hmacSha256(String secret, String data) {
    def result = ''
    withEnv(["SIGN_SECRET=${secret}", "SIGN_DATA=${data}"]) {
        result = sh(script: 'set +x && printf "%s" "$SIGN_DATA" | openssl dgst -sha256 -hmac "$SIGN_SECRET" | awk \'{print $2}\'', returnStdout: true).trim()
    }
    return result
}
