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

// ArtifactController 制品库控制器
type ArtifactController struct{}

func NewArtifactController() *ArtifactController {
	return &ArtifactController{}
}

// List 制品列表
// @Summary 获取制品列表
// @Tags CICD Artifact
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param pipeline_id query int false "流水线ID"
// @Param artifact_type query string false "制品类型：jar/binary/dist/wheel/image"
// @Param language_type query string false "语言类型：go/java/frontend/python"
// @Param status query string false "状态：ready/expired"
// @Router /api/v1/k8s/cicd/artifact/list [get]
func (c *ArtifactController) List(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	pipelineID, _ := strconv.ParseInt(ctx.Query("pipeline_id"), 10, 64)
	artifactType := ctx.Query("artifact_type")
	languageType := ctx.Query("language_type")
	status := ctx.Query("status")

	svc := services.NewServices()
	list, total, err := svc.ArtifactList(ctx.Request.Context(), pipelineID, artifactType, languageType, status, page, pageSize)
	if err != nil {
		global.Logger.Error("ArtifactList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// Detail 制品详情
// @Summary 获取制品详情
// @Tags CICD Artifact
// @Produce json
// @Param id query int true "制品ID"
// @Router /api/v1/k8s/cicd/artifact/detail [get]
func (c *ArtifactController) Detail(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的制品ID"))
		return
	}

	svc := services.NewServices()
	artifact, err := svc.ArtifactDetail(ctx.Request.Context(), id)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"artifact": artifact})
}

// Upload 上传制品
// @Summary 上传制品文件
// @Tags CICD Artifact
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "制品文件"
// @Param pipeline_id formData int false "流水线ID"
// @Param run_id formData int false "运行记录ID"
// @Param version formData string false "版本号"
// @Param language_type formData string false "语言类型"
// @Param artifact_type formData string false "制品类型"
// @Router /api/v1/k8s/cicd/artifact/upload [post]
func (c *ArtifactController) Upload(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("请上传文件"))
		return
	}
	defer file.Close()

	// 解析表单参数
	pipelineID, _ := strconv.ParseInt(ctx.PostForm("pipeline_id"), 10, 64)
	runID, _ := strconv.ParseInt(ctx.PostForm("run_id"), 10, 64)
	buildNumber, _ := strconv.Atoi(ctx.PostForm("build_number"))

	artifact := &models.CicdArtifact{
		PipelineID:   pipelineID,
		RunID:        runID,
		BuildNumber:  buildNumber,
		Version:      ctx.PostForm("version"),
		LanguageType: ctx.PostForm("language_type"),
		ArtifactType: ctx.PostForm("artifact_type"),
		GitRepo:      ctx.PostForm("git_repo"),
		GitBranch:    ctx.PostForm("git_branch"),
		GitCommit:    ctx.PostForm("git_commit"),
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	result, err := svc.ArtifactUpload(ctx.Request.Context(), file, header, artifact, userID)
	if err != nil {
		global.Logger.Error("ArtifactUpload error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message":  "上传成功",
		"artifact": result,
	})
}

// Download 下载制品
// @Summary 下载制品文件
// @Tags CICD Artifact
// @Produce octet-stream
// @Param id query int true "制品ID"
// @Param token query string false "认证Token（用于 window.open 下载）"
// @Router /api/v1/k8s/cicd/artifact/download [get]
func (c *ArtifactController) Download(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的制品ID"))
		return
	}

	svc := services.NewServices()
	artifact, err := svc.ArtifactDownload(ctx.Request.Context(), id)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	fileName := filepath.Base(artifact.FilePath)
	if artifact.Name != "" {
		fileName = artifact.Name
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Cache-Control", "no-cache")
	ctx.File(artifact.FilePath)
}

// Delete 删除制品
// @Summary 删除制品
// @Tags CICD Artifact
// @Accept json
// @Produce json
// @Router /api/v1/k8s/cicd/artifact/delete [post]
func (c *ArtifactController) Delete(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	var req struct {
		ID int64 `json:"id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.ArtifactDelete(ctx.Request.Context(), req.ID); err != nil {
		global.Logger.Error("ArtifactDelete error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "删除成功"})
}

// Stats 制品统计
// @Summary 获取制品统计数据
// @Tags CICD Artifact
// @Produce json
// @Param pipeline_id query int false "流水线ID（不传则统计全部）"
// @Router /api/v1/k8s/cicd/artifact/stats [get]
func (c *ArtifactController) Stats(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	pipelineID, _ := strconv.ParseInt(ctx.Query("pipeline_id"), 10, 64)

	svc := services.NewServices()
	stats, err := svc.ArtifactStats(ctx.Request.Context(), pipelineID)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"stats": stats})
}

// Update 更新制品信息
// @Summary 更新制品信息
// @Tags CICD Artifact
// @Accept json
// @Produce json
// @Router /api/v1/k8s/cicd/artifact/update [post]
func (c *ArtifactController) Update(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	var req struct {
		ID           int64  `json:"id" binding:"required"`
		Name         string `json:"name"`
		Version      string `json:"version"`
		ArtifactType string `json:"artifact_type"`
		Status       string `json:"status"`
		ImageRepo    string `json:"image_repo"`
		ImageTag     string `json:"image_tag"`
		ImageDigest  string `json:"image_digest"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Version != "" {
		updates["version"] = req.Version
	}
	if req.ArtifactType != "" {
		updates["artifact_type"] = req.ArtifactType
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.ImageRepo != "" {
		updates["image_repo"] = req.ImageRepo
	}
	if req.ImageTag != "" {
		updates["image_tag"] = req.ImageTag
	}
	if req.ImageDigest != "" {
		updates["image_digest"] = req.ImageDigest
	}

	if len(updates) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无可更新的字段"))
		return
	}

	svc := services.NewServices()
	if err := svc.ArtifactUpdate(ctx.Request.Context(), req.ID, updates); err != nil {
		global.Logger.Error("ArtifactUpdate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "更新成功"})
}

// CreateRecord 创建制品记录（无需上传文件，用于镜像类型的制品记录）
// @Summary 创建制品记录
// @Tags CICD Artifact
// @Accept json
// @Produce json
// @Router /api/v1/k8s/cicd/artifact/create [post]
func (c *ArtifactController) CreateRecord(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	var req struct {
		PipelineID   int64  `json:"pipeline_id"`
		RunID        int64  `json:"run_id"`
		BuildNumber  int    `json:"build_number"`
		Name         string `json:"name" binding:"required"`
		ArtifactType string `json:"artifact_type"`
		Version      string `json:"version"`
		LanguageType string `json:"language_type"`
		ImageRepo    string `json:"image_repo"`
		ImageTag     string `json:"image_tag"`
		ImageDigest  string `json:"image_digest"`
		GitRepo      string `json:"git_repo"`
		GitBranch    string `json:"git_branch"`
		GitCommit    string `json:"git_commit"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	artifact := &models.CicdArtifact{
		PipelineID:   req.PipelineID,
		RunID:        req.RunID,
		BuildNumber:  req.BuildNumber,
		Name:         req.Name,
		ArtifactType: req.ArtifactType,
		Version:      req.Version,
		LanguageType: req.LanguageType,
		ImageRepo:    req.ImageRepo,
		ImageTag:     req.ImageTag,
		ImageDigest:  req.ImageDigest,
		GitRepo:      req.GitRepo,
		GitBranch:    req.GitBranch,
		GitCommit:    req.GitCommit,
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	result, err := svc.ArtifactCreateRecord(ctx.Request.Context(), artifact, userID)
	if err != nil {
		global.Logger.Error("ArtifactCreateRecord error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message":  "创建成功",
		"artifact": result,
	})
}

// BatchDelete 批量删除制品
// @Summary 批量删除制品
// @Tags CICD Artifact
// @Accept json
// @Produce json
// @Router /api/v1/k8s/cicd/artifact/batch-delete [post]
func (c *ArtifactController) BatchDelete(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	if len(req.IDs) == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("删除ID列表不能为空"))
		return
	}
	if len(req.IDs) > 100 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("每次最多删除100条"))
		return
	}

	svc := services.NewServices()
	affected, err := svc.ArtifactBatchDelete(ctx.Request.Context(), req.IDs)
	if err != nil {
		global.Logger.Error("ArtifactBatchDelete error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message":  "批量删除成功",
		"affected": affected,
	})
}

// AttachFile 为已有制品补传/替换文件
// @Summary 为已有制品补传文件
// @Tags CICD Artifact
// @Accept multipart/form-data
// @Produce json
// @Param id formData int true "制品ID"
// @Param file formData file true "制品文件"
// @Router /api/v1/k8s/cicd/artifact/attach [post]
func (c *ArtifactController) AttachFile(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.PostForm("id"), 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的制品ID"))
		return
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("请上传文件"))
		return
	}
	defer file.Close()

	svc := services.NewServices()
	result, err := svc.ArtifactAttachFile(ctx.Request.Context(), id, file, header)
	if err != nil {
		global.Logger.Error("ArtifactAttachFile error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message":  "文件上传成功",
		"artifact": result,
	})
}

// ListByRunID 获取某次运行的制品列表
// @Summary 获取流水线某次运行产出的制品
// @Tags CICD Artifact
// @Produce json
// @Param run_id query int true "运行记录ID"
// @Router /api/v1/k8s/cicd/artifact/by-run [get]
func (c *ArtifactController) ListByRunID(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	runID, err := strconv.ParseInt(ctx.Query("run_id"), 10, 64)
	if err != nil || runID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的运行记录ID"))
		return
	}

	svc := services.NewServices()
	list, err := svc.ArtifactListByRunID(ctx.Request.Context(), runID)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"list": list, "total": len(list)})
}
