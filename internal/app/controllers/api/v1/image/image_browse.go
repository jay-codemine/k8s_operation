package image

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

// ImageBrowseController 镜像浏览控制器
type ImageBrowseController struct{}

func NewImageBrowseController() *ImageBrowseController {
	return &ImageBrowseController{}
}

// ListRepositories godoc
// @Summary 列出仓库中的镜像项目
// @Description 获取指定镜像仓库中的所有镜像项目/命名空间
// @Tags 镜像浏览
// @Produce json
// @Security ApiKeyAuth
// @Param registry_id query int true "仓库ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/browse/repositories [get]
func (c *ImageBrowseController) ListRepositories(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	registryID, err := strconv.ParseInt(ctx.Query("registry_id"), 10, 64)
	if err != nil || registryID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的仓库ID"))
		return
	}

	svc := services.NewImageService()
	repos, err := svc.ListRepositories(registryID)
	if err != nil {
		global.Logger.Error("获取镜像项目列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorListImagesFail.WithDetails(err.Error()))
		return
	}

	resp.SuccessList(repos, len(repos))
}

// ListTags godoc
// @Summary 列出镜像的所有标签
// @Description 获取指定镜像的所有标签/版本
// @Tags 镜像浏览
// @Produce json
// @Security ApiKeyAuth
// @Param registry_id query int true "仓库ID"
// @Param repository query string true "镜像名称"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/browse/tags [get]
func (c *ImageBrowseController) ListTags(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	registryID, err := strconv.ParseInt(ctx.Query("registry_id"), 10, 64)
	if err != nil || registryID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的仓库ID"))
		return
	}

	repository := ctx.Query("repository")
	if repository == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("镜像名称不能为空"))
		return
	}

	svc := services.NewImageService()
	tags, err := svc.ListTags(registryID, repository)
	if err != nil {
		global.Logger.Error("获取镜像标签失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorGetImageTagsFail.WithDetails(err.Error()))
		return
	}

	resp.SuccessList(tags, len(tags))
}

// GetImageDetail godoc
// @Summary 获取镜像详情
// @Description 获取指定镜像标签的详细信息
// @Tags 镜像浏览
// @Produce json
// @Security ApiKeyAuth
// @Param registry_id query int true "仓库ID"
// @Param repository query string true "镜像名称"
// @Param tag query string true "标签"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/browse/detail [get]
func (c *ImageBrowseController) GetImageDetail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	registryID, err := strconv.ParseInt(ctx.Query("registry_id"), 10, 64)
	if err != nil || registryID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的仓库ID"))
		return
	}

	repository := ctx.Query("repository")
	tag := ctx.Query("tag")
	if repository == "" || tag == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("镜像名称和标签不能为空"))
		return
	}

	svc := services.NewImageService()
	detail, err := svc.GetImageDetail(registryID, repository, tag)
	if err != nil {
		global.Logger.Error("获取镜像详情失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorGetImageTagsFail.WithDetails(err.Error()))
		return
	}

	resp.Success(detail)
}

// DeleteTag godoc
// @Summary 删除镜像标签
// @Description 删除指定镜像的标签
// @Tags 镜像浏览
// @Produce json
// @Security ApiKeyAuth
// @Param registry_id query int true "仓库ID"
// @Param repository query string true "镜像名称"
// @Param tag query string true "标签"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/browse/delete [post]
func (c *ImageBrowseController) DeleteTag(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	registryID, err := strconv.ParseInt(ctx.Query("registry_id"), 10, 64)
	if err != nil || registryID <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的仓库ID"))
		return
	}

	repository := ctx.Query("repository")
	tag := ctx.Query("tag")
	if repository == "" || tag == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("镜像名称和标签不能为空"))
		return
	}

	svc := services.NewImageService()
	if err := svc.DeleteTag(registryID, repository, tag); err != nil {
		global.Logger.Error("删除镜像标签失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorDeleteImageRegistryFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "删除成功"})
}

// ========================================
// 清理策略控制器
// ========================================

// CleanupPolicyController 清理策略控制器
type CleanupPolicyController struct{}

func NewCleanupPolicyController() *CleanupPolicyController {
	return &CleanupPolicyController{}
}

// List godoc
// @Summary 获取清理策略列表
// @Description 分页获取清理策略列表
// @Tags 镜像清理
// @Produce json
// @Security ApiKeyAuth
// @Param registry_id query int false "仓库ID（可选）"
// @Param keyword query string false "搜索关键词"
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/cleanup/list [get]
func (c *CleanupPolicyController) List(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	registryID, _ := strconv.ParseInt(ctx.Query("registry_id"), 10, 64)
	keyword := ctx.Query("keyword")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	svc := services.NewImageService()
	list, total, err := svc.ListCleanupPolicies(registryID, keyword, page, pageSize)
	if err != nil {
		global.Logger.Error("获取清理策略列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.SuccessList(list, int(total))
}

// Create godoc
// @Summary 创建清理策略
// @Description 创建新的镜像清理策略
// @Tags 镜像清理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body services.CleanupPolicyRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/cleanup/create [post]
func (c *CleanupPolicyController) Create(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req services.CleanupPolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	userID := int64(0)
	if uid, exists := ctx.Get("user_id"); exists {
		userID = uid.(int64)
	}

	svc := services.NewImageService()
	policy, err := svc.CreateCleanupPolicy(&req, userID)
	if err != nil {
		global.Logger.Error("创建清理策略失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(policy)
}

// Update godoc
// @Summary 更新清理策略
// @Description 更新镜像清理策略
// @Tags 镜像清理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body services.CleanupPolicyRequest true "更新参数"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/cleanup/update [post]
func (c *CleanupPolicyController) Update(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req services.CleanupPolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewImageService()
	policy, err := svc.UpdateCleanupPolicy(&req)
	if err != nil {
		global.Logger.Error("更新清理策略失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(policy)
}

// Delete godoc
// @Summary 删除清理策略
// @Description 删除镜像清理策略
// @Tags 镜像清理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "策略ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/cleanup/delete [post]
func (c *CleanupPolicyController) Delete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的策略ID"))
		return
	}

	svc := services.NewImageService()
	if err := svc.DeleteCleanupPolicy(id); err != nil {
		global.Logger.Error("删除清理策略失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "删除成功"})
}

// Toggle godoc
// @Summary 启用/禁用清理策略
// @Description 切换清理策略的启用状态
// @Tags 镜像清理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "策略ID"
// @Param enabled query bool true "是否启用"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/cleanup/toggle [post]
func (c *CleanupPolicyController) Toggle(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的策略ID"))
		return
	}

	enabled := ctx.Query("enabled") == "true"

	svc := services.NewImageService()
	if err := svc.ToggleCleanupPolicy(id, enabled); err != nil {
		global.Logger.Error("切换策略状态失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "操作成功"})
}

// Run godoc
// @Summary 手动执行清理策略
// @Description 手动触发一次清理任务
// @Tags 镜像清理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "策略ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/cleanup/run [post]
func (c *CleanupPolicyController) Run(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	id, err := strconv.ParseInt(ctx.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的策略ID"))
		return
	}

	svc := services.NewImageService()
	log, err := svc.RunCleanupPolicy(id)
	if err != nil {
		global.Logger.Error("执行清理策略失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(log)
}

// Logs godoc
// @Summary 获取清理日志
// @Description 获取清理任务执行日志
// @Tags 镜像清理
// @Produce json
// @Security ApiKeyAuth
// @Param policy_id query int false "策略ID（可选）"
// @Param limit query int false "返回数量（默认20）"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/cleanup/logs [get]
func (c *CleanupPolicyController) Logs(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	policyID, _ := strconv.ParseInt(ctx.Query("policy_id"), 10, 64)
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	svc := services.NewImageService()
	logs, err := svc.GetCleanupLogs(policyID, limit)
	if err != nil {
		global.Logger.Error("获取清理日志失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.SuccessList(logs, len(logs))
}
