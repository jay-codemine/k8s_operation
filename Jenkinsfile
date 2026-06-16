// ==================== 通用 Pipeline 分发器 ====================
// 根据平台传递的 LANGUAGE_TYPE 参数，自动加载对应语言的构建模板
// 支持的类型：go / java / frontend / python
// 若未传递 LANGUAGE_TYPE 或值为空，默认使用 go 模板（向下兼容）
//
// 用法：所有 Jenkins Job 的 Script Path 保持默认 "Jenkinsfile" 即可
//       平台触发构建时会自动注入 LANGUAGE_TYPE 参数
//       分发器根据参数值动态加载 configs/jenkins-templates/ 下对应的模板
// ==================================================================

def templateMap = [
    'go'      : 'configs/jenkins-templates/go-pipeline.groovy',
    'java'    : 'configs/jenkins-templates/java-spring-pipeline.groovy',
    'frontend': 'configs/jenkins-templates/frontend-pipeline.groovy',
    'python'  : 'configs/jenkins-templates/python-pipeline.groovy'
]

node {
    // 先 checkout 代码仓库，确保模板文件在 workspace 中可用
    checkout scm

    // 从平台注入的参数中获取语言类型
    def langType = params.LANGUAGE_TYPE?.trim() ?: 'go'
    def templateFile = templateMap[langType]

    if (!templateFile) {
        error("""
=== 不支持的语言类型 ===
LANGUAGE_TYPE: ${langType}
支持的类型: ${templateMap.keySet().join(', ')}

请在平台「语言/框架类型」中选择正确的类型。
""")
    }

    echo "=========================================="
    echo "  K8s Platform - Pipeline 模板分发器"
    echo "  语言类型:  ${langType}"
    echo "  加载模板:  ${templateFile}"
    echo "=========================================="

    // 动态加载对应语言的 Pipeline 模板
    load templateFile
}
