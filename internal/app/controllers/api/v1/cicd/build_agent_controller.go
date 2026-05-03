package cicd

import (
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

type BuildAgentController struct{}

func NewBuildAgentController() *BuildAgentController {
	return &BuildAgentController{}
}

// List 探针列表
// @Summary 获取构建探针列表
// @Tags CICD BuildAgent
// @Produce json
// @Param category query string false "分类：observability/diagnostics/security/custom"
// @Param scope query string false "适用语言：java/go/python/all"
// @Param status query string false "状态：active/inactive"
// @Param keyword query string false "关键词搜索"
// @Router /api/v1/k8s/cicd/agent/list [get]
func (c *BuildAgentController) List(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	category := ctx.Query("category")
	scope := ctx.Query("scope")
	status := ctx.Query("status")
	keyword := ctx.Query("keyword")

	svc := services.NewServices()
	list, total, err := svc.BuildAgentList(ctx.Request.Context(), category, scope, status, keyword, page, pageSize)
	if err != nil {
		global.Logger.Error("BuildAgentList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// Detail 探针详情
// @Summary 获取构建探针详情
// @Tags CICD BuildAgent
// @Produce json
// @Param id query int true "探针ID"
// @Router /api/v1/k8s/cicd/agent/detail [get]
func (c *BuildAgentController) Detail(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的探针ID"))
		return
	}

	svc := services.NewServices()
	agent, err := svc.BuildAgentDetail(ctx.Request.Context(), id)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"agent": agent})
}

// Upload 上传探针
// @Summary 上传构建探针文件
// @Tags CICD BuildAgent
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "探针文件（如 opentelemetry-javaagent.jar）"
// @Param name formData string true "探针名称（唯一标识）"
// @Param display_name formData string true "显示名称"
// @Param category formData string true "分类"
// @Param scope formData string true "适用语言"
// @Param version formData string false "版本号"
// @Router /api/v1/k8s/cicd/agent/upload [post]
func (c *BuildAgentController) Upload(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("请上传文件"))
		return
	}
	defer file.Close()

	name := ctx.PostForm("name")
	if name == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("探针名称不能为空"))
		return
	}

	agent := &models.CicdBuildAgent{
		Name:           name,
		DisplayName:    ctx.PostForm("display_name"),
		Description:    ctx.PostForm("description"),
		Category:       ctx.PostForm("category"),
		Scope:          ctx.PostForm("scope"),
		Version:        ctx.PostForm("version"),
		DownloadURL:    ctx.PostForm("download_url"),
		DocURL:         ctx.PostForm("doc_url"),
		Icon:           ctx.PostForm("icon"),
		DockerCopyDest: ctx.PostForm("docker_copy_dest"),
		EnvKey:         ctx.PostForm("env_key"),
		EnvValue:       ctx.PostForm("env_value"),
	}

	// 默认值
	if agent.Category == "" {
		agent.Category = models.AgentCategoryObservability
	}
	if agent.Scope == "" {
		agent.Scope = models.AgentScopeJava
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	result, err := svc.BuildAgentUpload(ctx.Request.Context(), file, header, agent, userID)
	if err != nil {
		global.Logger.Error("BuildAgentUpload error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message": "上传成功",
		"agent":   result,
	})
}

// Update 更新探针信息
// @Summary 更新构建探针信息
// @Tags CICD BuildAgent
// @Accept json
// @Produce json
// @Router /api/v1/k8s/cicd/agent/update [post]
func (c *BuildAgentController) Update(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	var req struct {
		ID             int64  `json:"id" binding:"required"`
		DisplayName    string `json:"display_name"`
		Description    string `json:"description"`
		Category       string `json:"category"`
		Scope          string `json:"scope"`
		Version        string `json:"version"`
		DockerCopyDest string `json:"docker_copy_dest"`
		EnvKey         string `json:"env_key"`
		EnvValue       string `json:"env_value"`
		DownloadURL    string `json:"download_url"`
		DocURL         string `json:"doc_url"`
		Icon           string `json:"icon"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Scope != "" {
		updates["scope"] = req.Scope
	}
	if req.Version != "" {
		updates["version"] = req.Version
	}
	if req.DockerCopyDest != "" {
		updates["docker_copy_dest"] = req.DockerCopyDest
	}
	if req.EnvKey != "" {
		updates["env_key"] = req.EnvKey
	}
	if req.EnvValue != "" {
		updates["env_value"] = req.EnvValue
	}
	if req.DownloadURL != "" {
		updates["download_url"] = req.DownloadURL
	}
	if req.DocURL != "" {
		updates["doc_url"] = req.DocURL
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}

	if len(updates) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无可更新的字段"))
		return
	}

	svc := services.NewServices()
	if err := svc.BuildAgentUpdate(ctx.Request.Context(), req.ID, updates); err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "更新成功"})
}

// ToggleStatus 切换探针启用/停用
// @Summary 切换构建探针启用状态
// @Tags CICD BuildAgent
// @Accept json
// @Produce json
// @Router /api/v1/k8s/cicd/agent/toggle [post]
func (c *BuildAgentController) ToggleStatus(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	var req struct {
		ID int64 `json:"id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	agent, err := svc.BuildAgentToggleStatus(ctx.Request.Context(), req.ID)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "状态切换成功", "agent": agent})
}

// Delete 删除探针
// @Summary 删除构建探针
// @Tags CICD BuildAgent
// @Accept json
// @Produce json
// @Router /api/v1/k8s/cicd/agent/delete [post]
func (c *BuildAgentController) Delete(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	var req struct {
		ID int64 `json:"id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.BuildAgentDelete(ctx.Request.Context(), req.ID); err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "删除成功"})
}

// Download 下载探针文件
// @Summary 下载构建探针文件
// @Tags CICD BuildAgent
// @Produce octet-stream
// @Param id query int false "探针ID（二选一）"
// @Param name query string false "探针名称（二选一，流水线构建使用）"
// @Router /api/v1/k8s/cicd/agent/download [get]
func (c *BuildAgentController) Download(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	svc := services.NewServices()
	var agent *models.CicdBuildAgent
	var err error

	// 支持按 ID 或按名称下载
	name := ctx.Query("name")
	if name != "" {
		agent, err = svc.BuildAgentDownloadByName(ctx.Request.Context(), name)
	} else {
		id, parseErr := strconv.ParseInt(ctx.Query("id"), 10, 64)
		if parseErr != nil || id <= 0 {
			rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("请提供探针 ID 或名称"))
			return
		}
		agent, err = svc.BuildAgentDownload(ctx.Request.Context(), id)
	}

	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	fileName := filepath.Base(agent.FilePath)
	if agent.FileName != "" {
		fileName = agent.FileName
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("X-Agent-Version", agent.Version)
	ctx.Header("X-Agent-Sha256", agent.Sha256)
	ctx.File(agent.FilePath)
}

// ListByScope 获取指定语言的已启用探针（流水线构建时使用）
// @Summary 获取指定语言的已启用探针
// @Tags CICD BuildAgent
// @Produce json
// @Param scope query string true "语言类型：java/go/python"
// @Router /api/v1/k8s/cicd/agent/by-scope [get]
func (c *BuildAgentController) ListByScope(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	scope := ctx.Query("scope")
	if scope == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("请指定语言类型"))
		return
	}

	svc := services.NewServices()
	list, err := svc.BuildAgentListByScope(ctx.Request.Context(), scope)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"list": list, "total": len(list)})
}
