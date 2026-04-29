package kube_cicd

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/controllers/api/v1/cicd"
)

type CicdRouter struct {
	releaseCtrl     *cicd.CicdReleaseController
	pipelineCtrl    *cicd.PipelineController
	gitCtrl         *cicd.GitController
	environmentCtrl *cicd.EnvironmentController
	approvalCtrl    *cicd.ApprovalController
	stageCtrl       *cicd.StageController
	templateCtrl    *cicd.TemplateController
	resourceCtrl    *cicd.ResourceController
	artifactCtrl    *cicd.ArtifactController
}

func NewCicdRouter() *CicdRouter {
	return &CicdRouter{
		releaseCtrl:     cicd.NewCicdReleaseController(),
		pipelineCtrl:    cicd.NewPipelineController(),
		gitCtrl:         cicd.NewGitController(),
		environmentCtrl: cicd.NewEnvironmentController(),
		approvalCtrl:    cicd.NewApprovalController(),
		stageCtrl:       cicd.NewStageController(),
		templateCtrl:    cicd.NewTemplateController(),
		resourceCtrl:    cicd.NewResourceController(),
		artifactCtrl:    cicd.NewArtifactController(),
	}
}

func (r *CicdRouter) Inject(rg *gin.RouterGroup) {
	// ==================== 流水线管理 ====================
	// /api/v1/k8s/cicd/pipeline/...
	pipeline := rg.Group("/pipeline")
	{
		pipeline.GET("/list", r.pipelineCtrl.List)         // 获取流水线列表
		pipeline.GET("/detail", r.pipelineCtrl.Detail)     // 获取流水线详情
		pipeline.POST("/create", r.pipelineCtrl.Create)    // 创建流水线
		pipeline.POST("/batch-create", r.pipelineCtrl.BatchCreate) // 批量创建流水线（导入多个项目）
		pipeline.POST("/update", r.pipelineCtrl.Update)    // 更新流水线
		pipeline.POST("/delete", r.pipelineCtrl.Delete)    // 删除流水线
		pipeline.POST("/run", r.pipelineCtrl.Run)          // 运行流水线（触发 Jenkins 构建）
		pipeline.POST("/stop", r.pipelineCtrl.Stop)        // 停止流水线
		pipeline.POST("/batch-run", r.pipelineCtrl.BatchRun)   // 批量运行流水线
		pipeline.POST("/batch-stop", r.pipelineCtrl.BatchStop) // 批量停止流水线
		pipeline.GET("/logs", r.pipelineCtrl.Logs)         // 获取构建日志
		pipeline.GET("/status", r.pipelineCtrl.Status)     // 获取实时状态
		pipeline.GET("/stages", r.pipelineCtrl.Stages)     // 获取阶段数据（Jenkins Pipeline）
		pipeline.GET("/history", r.pipelineCtrl.History)   // 获取运行历史
		pipeline.GET("/template-verify", r.pipelineCtrl.TemplateVerify)     // 模板化发布验证
		pipeline.GET("/template-simulate", r.pipelineCtrl.TemplateSimulate) // 模拟模板化发布流程
		pipeline.GET("/sonar-report", r.pipelineCtrl.SonarReport)          // SonarQube 代码质量报告
		// sonar-callback 已移至 cicd_callback_router.go（公开接口，跳过 JWT，Jenkins 回调无需认证）
		// callback 已移至 cicd_callback_router.go（公开接口，跳过 JWT）
	}

	// ==================== 发布单管理 ====================
	// /api/v1/k8s/cicd/release/...
	release := rg.Group("/release")
	{
		release.POST("/create", r.releaseCtrl.Create)     // 创建发布单
		release.GET("/detail", r.releaseCtrl.Detail)      // 发布单详情
		release.GET("/list", r.releaseCtrl.List)          // 发布单列表
		release.GET("/stats", r.releaseCtrl.Stats)        // 发布单统计
		release.POST("/update", r.releaseCtrl.Update)     // 编辑发布单
		release.POST("/delete", r.releaseCtrl.Delete)     // 删除发布单
		release.POST("/cancel", r.releaseCtrl.Cancel)     // 取消发布（智能判断回滚/取消）
		release.POST("/rollback", r.releaseCtrl.Rollback) // 回滚发布
		release.POST("/retry", r.releaseCtrl.Retry)       // 重试发布
		release.GET("/tasks", r.releaseCtrl.Tasks)        // 获取发布单下的任务列表
		release.POST("/batch-retry", r.releaseCtrl.BatchRetry)       // 批量重新发布
		release.POST("/batch-rollback", r.releaseCtrl.BatchRollback) // 批量回滚
		release.POST("/sync-from-pipeline", r.releaseCtrl.SyncFromPipeline) // 同步流水线记录
	}

	// ==================== 回调接口 ====================
	// 回调接口已移至 cicd_callback_router.go（公开接口，跳过JWT）

	// ==================== Git 仓库操作 ====================
	// /api/v1/k8s/cicd/git/...
	git := rg.Group("/git")
	{
		git.POST("/branches", r.gitCtrl.GetBranches) // 获取远程分支列表
		git.POST("/validate", r.gitCtrl.ValidateRepo) // 验证仓库连接
	}

	// ==================== 环境管理 ====================
	// /api/v1/k8s/cicd/environment/...
	environment := rg.Group("/environment")
	{
		environment.GET("/list", r.environmentCtrl.List)       // 获取环境列表
		environment.GET("/detail", r.environmentCtrl.Detail)   // 获取环境详情
		environment.POST("/create", r.environmentCtrl.Create)  // 创建环境
		environment.POST("/update", r.environmentCtrl.Update)  // 更新环境
		environment.POST("/delete", r.environmentCtrl.Delete)  // 删除环境
	}

	// ==================== 审批流程 ====================
	// /api/v1/k8s/cicd/approval/...
	approval := rg.Group("/approval")
	{
		approval.GET("/list", r.approvalCtrl.List)       // 获取审批列表
		approval.GET("/detail", r.approvalCtrl.Detail)   // 获取审批详情
		approval.GET("/pending", r.approvalCtrl.Pending) // 获取待审批列表
		approval.GET("/stats", r.approvalCtrl.Stats)     // 获取审批统计
		approval.POST("/create", r.approvalCtrl.Create)  // 创建审批申请
		approval.POST("/update", r.approvalCtrl.Update)  // 更新审批记录
		approval.POST("/delete", r.approvalCtrl.Delete)  // 删除审批记录
		approval.POST("/action", r.approvalCtrl.Action)  // 审批操作
	}

	// ==================== 流水线阶段 ====================
	// /api/v1/k8s/cicd/stage/...
	stage := rg.Group("/stage")
	{
		stage.GET("/list", r.stageCtrl.GetStages)        // 获取运行阶段列表
		stage.GET("/logs", r.stageCtrl.GetStageLogs)     // 获取阶段日志
		stage.POST("/approve", r.stageCtrl.ApproveStage) // 审批阶段
		stage.POST("/deploy", r.stageCtrl.DeployStage)   // 执行部署阶段
		stage.POST("/cancel", r.stageCtrl.CancelDeploy)  // 取消部署（智能判断）
		stage.POST("/rollback", r.stageCtrl.RollbackDeploy) // 回滚到指定版本
		stage.GET("/history", r.stageCtrl.GetDeployHistory) // 获取历史版本列表
		// callback 已移至 cicd_callback_router.go（公开接口，跳过JWT）
	}

	// ==================== 流水线模板 ====================
	// /api/v1/k8s/cicd/template/...
	template := rg.Group("/template")
	{
		template.GET("/list", r.templateCtrl.List)       // 获取模板列表
		template.GET("/detail", r.templateCtrl.Detail)   // 获取模板详情
		template.POST("/create", r.templateCtrl.Create)  // 创建模板
		template.POST("/update", r.templateCtrl.Update)  // 更新模板
		template.POST("/delete", r.templateCtrl.Delete)  // 删除模板
	}

	// ==================== 资源配置管理 ====================
	// /api/v1/k8s/cicd/resource/...
	resource := rg.Group("/resource")
	{
		// 资源模板
		resource.GET("/templates", r.resourceCtrl.TemplateList)           // 获取资源模板列表
		resource.GET("/template/default", r.resourceCtrl.TemplateDefault) // 获取默认模板
		resource.GET("/template/:id", r.resourceCtrl.TemplateDetail)      // 获取模板详情
		resource.POST("/template", r.resourceCtrl.TemplateCreate)         // 创建模板
		resource.PUT("/template/:id", r.resourceCtrl.TemplateUpdate)      // 更新模板
		resource.DELETE("/template/:id", r.resourceCtrl.TemplateDelete)   // 删除模板

		// 环境规则
		resource.GET("/rules", r.resourceCtrl.RuleList)     // 获取规则列表
		resource.PUT("/rule/:id", r.resourceCtrl.RuleUpdate) // 更新规则

		// 资源校验
		resource.POST("/validate", r.resourceCtrl.Validate) // 校验资源配置

		// 发布审批
		resource.GET("/approvals", r.resourceCtrl.ApprovalList)              // 审批列表
		resource.GET("/approval/:id", r.resourceCtrl.ApprovalDetail)         // 审批详情
		resource.PUT("/approval/:id/approve", r.resourceCtrl.ApprovalApprove) // 通过审批
		resource.PUT("/approval/:id/reject", r.resourceCtrl.ApprovalReject)   // 拒绝审批
	}

	// ==================== 制品库管理 ====================
	// /api/v1/k8s/cicd/artifact/...
	artifact := rg.Group("/artifact")
	{
		artifact.GET("/list", r.artifactCtrl.List)                 // 制品列表（分页 + 筛选）
		artifact.GET("/detail", r.artifactCtrl.Detail)             // 制品详情
		artifact.GET("/by-run", r.artifactCtrl.ListByRunID)        // 某次运行的制品列表
		artifact.POST("/create", r.artifactCtrl.CreateRecord)      // 创建制品记录（镜像类型 / 无文件）
		// upload 已移至 cicd_callback_router.go（公开接口，Jenkins 回调无需 JWT，手动上传也可用）
		artifact.POST("/attach", r.artifactCtrl.AttachFile)        // 为已有制品补传/替换文件
		artifact.POST("/update", r.artifactCtrl.Update)            // 更新制品信息
		artifact.GET("/download", r.artifactCtrl.Download)         // 下载制品文件
		artifact.POST("/delete", r.artifactCtrl.Delete)            // 删除制品
		artifact.POST("/batch-delete", r.artifactCtrl.BatchDelete) // 批量删除制品
		artifact.GET("/stats", r.artifactCtrl.Stats)               // 制品统计（按类型分组）
	}
}
