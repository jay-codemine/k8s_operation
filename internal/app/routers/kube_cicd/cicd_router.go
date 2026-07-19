package kube_cicd

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/controllers/api/v1/cicd"
	"k8soperation/middlewares"
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
	agentCtrl       *cicd.BuildAgentController
	onboardCtrl     *cicd.QuickOnboardController
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
		agentCtrl:       cicd.NewBuildAgentController(),
		onboardCtrl:     cicd.NewQuickOnboardController(),
	}
}

func (r *CicdRouter) Inject(rg *gin.RouterGroup) {
	// ==================== 流水线管理 ====================
	// /api/v1/k8s/cicd/pipeline/...
	pipeline := rg.Group("/pipeline")
	{
		pipeline.GET("/list", r.pipelineCtrl.List)
		pipeline.GET("/detail", r.pipelineCtrl.Detail)
		pipeline.GET("/check-name", r.pipelineCtrl.CheckName)
		pipeline.POST("/create", middlewares.RequireCICDPermission("cicd:pipeline:create"), r.pipelineCtrl.Create)
		pipeline.POST("/batch-create", middlewares.RequireCICDPermission("cicd:pipeline:create"), r.pipelineCtrl.BatchCreate)
		pipeline.POST("/update", middlewares.RequireCICDPermission("cicd:pipeline:edit"), r.pipelineCtrl.Update)
		pipeline.POST("/delete", middlewares.RequireCICDPermission("cicd:pipeline:delete"), r.pipelineCtrl.Delete)
		pipeline.POST("/run", middlewares.RequireCICDPermission("cicd:pipeline:run"), r.pipelineCtrl.Run)
		pipeline.POST("/stop", middlewares.RequireCICDPermission("cicd:build:cancel"), r.pipelineCtrl.Stop)
		pipeline.POST("/batch-run", middlewares.RequireCICDPermission("cicd:pipeline:run"), r.pipelineCtrl.BatchRun)
		pipeline.POST("/batch-stop", middlewares.RequireCICDPermission("cicd:build:cancel"), r.pipelineCtrl.BatchStop)
		pipeline.GET("/logs", r.pipelineCtrl.Logs)
		pipeline.GET("/status", r.pipelineCtrl.Status)
		pipeline.GET("/stages", r.pipelineCtrl.Stages)
		pipeline.GET("/history", r.pipelineCtrl.History)
		pipeline.GET("/build-records", r.pipelineCtrl.BuildRecords)
		pipeline.GET("/build-records/export", r.pipelineCtrl.ExportBuildRecords)
		pipeline.GET("/build-stats", r.pipelineCtrl.BuildStats)
		pipeline.GET("/template-verify", r.pipelineCtrl.TemplateVerify)
		pipeline.GET("/template-simulate", r.pipelineCtrl.TemplateSimulate)
		pipeline.GET("/sonar-report", r.pipelineCtrl.SonarReport)
		pipeline.GET("/deploy-silence-status", r.pipelineCtrl.DeploySilenceStatus)
		pipeline.GET("/jenkins-config", r.pipelineCtrl.JenkinsConfig)
		pipeline.GET("/discover", r.pipelineCtrl.Discover)
		pipeline.GET("/discover-workloads", r.pipelineCtrl.DiscoverWorkloads)
		pipeline.GET("/discover-apps", r.pipelineCtrl.DiscoverApplications)
	}

	// ==================== 发布单管理 ====================
	// ==================== 金丝雀部署 ====================
	canaryCtrl := cicd.NewCanaryDeployController()
	canary := rg.Group("/canary")
	{
		canary.POST("/promote", middlewares.RequireCICDPermission("cicd:deploy:dev"), canaryCtrl.Promote)
		canary.POST("/rollback", middlewares.RequireCICDPermission("cicd:deploy:rollback"), canaryCtrl.Rollback)
		canary.GET("/status", canaryCtrl.Status)
		canary.POST("/traffic-split", middlewares.RequireCICDPermission("cicd:deploy:dev"), canaryCtrl.SetTrafficSplit)
	}
	// /api/v1/k8s/cicd/release/...
	release := rg.Group("/release")
	{
		release.POST("/create", middlewares.RequireCICDPermission("cicd:deploy:dev"), r.releaseCtrl.Create)
		release.GET("/detail", r.releaseCtrl.Detail)
		release.GET("/list", r.releaseCtrl.List)
		release.GET("/stats", r.releaseCtrl.Stats)
		release.POST("/update", middlewares.RequireCICDPermission("cicd:pipeline:edit"), r.releaseCtrl.Update)
		release.POST("/delete", middlewares.RequireCICDPermission("cicd:pipeline:delete"), r.releaseCtrl.Delete)
		release.POST("/cancel", middlewares.RequireCICDPermission("cicd:deploy:rollback"), r.releaseCtrl.Cancel)
		release.POST("/rollback", middlewares.RequireCICDPermission("cicd:deploy:rollback"), r.releaseCtrl.Rollback)
		release.POST("/retry", middlewares.RequireCICDPermission("cicd:deploy:dev"), r.releaseCtrl.Retry)
		release.GET("/tasks", r.releaseCtrl.Tasks)
		release.POST("/batch-retry", middlewares.RequireCICDPermission("cicd:deploy:dev"), r.releaseCtrl.BatchRetry)
		release.POST("/batch-rollback", middlewares.RequireCICDPermission("cicd:deploy:rollback"), r.releaseCtrl.BatchRollback)
		release.POST("/batch-cancel", middlewares.RequireCICDPermission("cicd:deploy:rollback"), r.releaseCtrl.BatchCancel)
		release.POST("/sync-from-pipeline", middlewares.RequireCICDPermission("cicd:pipeline:run"), r.releaseCtrl.SyncFromPipeline)
		release.GET("/history", r.releaseCtrl.History)
		release.GET("/stats-enhanced", r.releaseCtrl.StatsEnhanced)
	}

	// ==================== 回调接口 ====================
	// 回调接口已移至 cicd_callback_router.go（公开接口，跳过JWT）

	// ==================== Git 仓库操作 ====================
	// /api/v1/k8s/cicd/git/...
	git := rg.Group("/git")
	{
		git.POST("/branches", r.gitCtrl.GetBranches)
		git.POST("/validate", r.gitCtrl.ValidateRepo)
	}

	// ==================== 环境管理 ====================
	// /api/v1/k8s/cicd/environment/...
	environment := rg.Group("/environment")
	{
		environment.GET("/list", r.environmentCtrl.List)
		environment.GET("/detail", r.environmentCtrl.Detail)
		environment.POST("/create", middlewares.RequireCICDPermission("cicd:environment:manage"), r.environmentCtrl.Create)
		environment.POST("/update", middlewares.RequireCICDPermission("cicd:environment:manage"), r.environmentCtrl.Update)
		environment.POST("/delete", middlewares.RequireCICDPermission("cicd:environment:manage"), r.environmentCtrl.Delete)
	}

	// ==================== 审批流程 ====================
	// /api/v1/k8s/cicd/approval/...
	approval := rg.Group("/approval")
	{
		approval.GET("/list", r.approvalCtrl.List)
		approval.GET("/detail", r.approvalCtrl.Detail)
		approval.GET("/pending", r.approvalCtrl.Pending)
		approval.GET("/stats", r.approvalCtrl.Stats)
		approval.POST("/create", middlewares.RequireCICDPermission("cicd:deploy:dev"), r.approvalCtrl.Create)
		approval.POST("/update", middlewares.RequireCICDPermission("cicd:pipeline:edit"), r.approvalCtrl.Update)
		approval.POST("/delete", middlewares.RequireCICDPermission("cicd:approval:manage"), r.approvalCtrl.Delete)
		approval.POST("/action", middlewares.RequireCICDPermission("cicd:approval:action"), r.approvalCtrl.Action)
		approval.POST("/batch-action", middlewares.RequireCICDPermission("cicd:approval:action"), r.approvalCtrl.BatchAction)
	}

	// ==================== 流水线阶段 ====================
	// /api/v1/k8s/cicd/stage/...
	stage := rg.Group("/stage")
	{
		stage.GET("/list", r.stageCtrl.GetStages)
		stage.GET("/logs", r.stageCtrl.GetStageLogs)
		stage.POST("/approve", middlewares.RequireCICDPermission("cicd:approval:action"), r.stageCtrl.ApproveStage)
		stage.POST("/deploy", middlewares.RequireCICDPermission("cicd:deploy:dev"), r.stageCtrl.DeployStage)
		stage.POST("/cancel", middlewares.RequireCICDPermission("cicd:deploy:rollback"), r.stageCtrl.CancelDeploy)
		stage.POST("/rollback", middlewares.RequireCICDPermission("cicd:deploy:rollback"), r.stageCtrl.RollbackDeploy)
		stage.GET("/history", r.stageCtrl.GetDeployHistory)
		stage.GET("/deploy-status", r.stageCtrl.DeployStatus)
	}

	// ==================== 流水线模板 ====================
	// /api/v1/k8s/cicd/template/...
	template := rg.Group("/template")
	{
		template.GET("/list", r.templateCtrl.List)
		template.GET("/detail", r.templateCtrl.Detail)
		template.POST("/create", middlewares.RequireCICDPermission("cicd:template:manage"), r.templateCtrl.Create)
		template.POST("/update", middlewares.RequireCICDPermission("cicd:template:manage"), r.templateCtrl.Update)
		template.POST("/delete", middlewares.RequireCICDPermission("cicd:template:manage"), r.templateCtrl.Delete)
	}

	// ==================== 资源配置管理 ====================
	// /api/v1/k8s/cicd/resource/...
	resource := rg.Group("/resource")
	{
		// 资源模板
		resource.GET("/templates", r.resourceCtrl.TemplateList)
		resource.GET("/template/default", r.resourceCtrl.TemplateDefault)
		resource.GET("/template/:id", r.resourceCtrl.TemplateDetail)
		resource.POST("/template", middlewares.RequireCICDPermission("cicd:resource:manage"), r.resourceCtrl.TemplateCreate)
		resource.PUT("/template/:id", middlewares.RequireCICDPermission("cicd:resource:manage"), r.resourceCtrl.TemplateUpdate)
		resource.DELETE("/template/:id", middlewares.RequireCICDPermission("cicd:resource:manage"), r.resourceCtrl.TemplateDelete)

		// 环境规则
		resource.GET("/rules", r.resourceCtrl.RuleList)
		resource.PUT("/rule/:id", middlewares.RequireCICDPermission("cicd:resource:manage"), r.resourceCtrl.RuleUpdate)

		// 资源校验
		resource.POST("/validate", r.resourceCtrl.Validate)

		// 发布审批
		resource.GET("/approvals", r.resourceCtrl.ApprovalList)
		resource.GET("/approval/:id", r.resourceCtrl.ApprovalDetail)
		resource.PUT("/approval/:id/approve", middlewares.RequireCICDPermission("cicd:approval:action"), r.resourceCtrl.ApprovalApprove)
		resource.PUT("/approval/:id/reject", middlewares.RequireCICDPermission("cicd:approval:action"), r.resourceCtrl.ApprovalReject)
	}

	// ==================== 制品库管理 ====================
	// /api/v1/k8s/cicd/artifact/...
	artifact := rg.Group("/artifact")
	{
		artifact.GET("/list", r.artifactCtrl.List)
		artifact.GET("/detail", r.artifactCtrl.Detail)
		artifact.GET("/by-run", r.artifactCtrl.ListByRunID)
		artifact.POST("/create", middlewares.RequireCICDPermission("cicd:artifact:upload"), r.artifactCtrl.CreateRecord)
		artifact.POST("/attach", middlewares.RequireCICDPermission("cicd:artifact:upload"), r.artifactCtrl.AttachFile)
		artifact.POST("/update", middlewares.RequireCICDPermission("cicd:artifact:upload"), r.artifactCtrl.Update)
		artifact.GET("/download", r.artifactCtrl.Download)
		artifact.POST("/delete", middlewares.RequireCICDPermission("cicd:artifact:delete"), r.artifactCtrl.Delete)
		artifact.POST("/batch-delete", middlewares.RequireCICDPermission("cicd:artifact:delete"), r.artifactCtrl.BatchDelete)
		artifact.GET("/stats", r.artifactCtrl.Stats)
	}

	// ==================== 构建探针管理 ====================
	// /api/v1/k8s/cicd/agent/...
	agent := rg.Group("/agent")
	{
		agent.GET("/list", r.agentCtrl.List)
		agent.GET("/detail", r.agentCtrl.Detail)
		agent.POST("/upload", middlewares.RequireCICDPermission("cicd:agent:manage"), r.agentCtrl.Upload)
		agent.POST("/update", middlewares.RequireCICDPermission("cicd:agent:manage"), r.agentCtrl.Update)
		agent.POST("/toggle", middlewares.RequireCICDPermission("cicd:agent:manage"), r.agentCtrl.ToggleStatus)
		agent.POST("/delete", middlewares.RequireCICDPermission("cicd:agent:manage"), r.agentCtrl.Delete)
		agent.GET("/download", r.agentCtrl.Download)
		agent.GET("/by-scope", r.agentCtrl.ListByScope)
	}

	// ==================== GitOps 管理（认证接口） ====================
	// /api/v1/k8s/cicd/gitops/...
	gitOpsCtrl := cicd.NewGitOpsController()
	gitops := rg.Group("/gitops")
	{
		gitops.GET("/app-status", gitOpsCtrl.GetAppStatus)
		gitops.GET("/sync-history", gitOpsCtrl.GetSyncHistory)
		gitops.POST("/sync", middlewares.RequireCICDPermission("cicd:pipeline:run"), gitOpsCtrl.TriggerSync)
	}

	// ==================== 快速接入 ====================
	rg.POST("/quick-onboard", middlewares.RequireCICDPermission("cicd:pipeline:create"), r.onboardCtrl.Onboard)
}
