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

// ImageRegistryController 镜像仓库控制器
type ImageRegistryController struct{}

func NewImageRegistryController() *ImageRegistryController {
	return &ImageRegistryController{}
}

// List godoc
// @Summary 获取镜像仓库列表
// @Description 分页获取镜像仓库列表
// @Tags 镜像管理
// @Produce json
// @Security ApiKeyAuth
// @Param keyword query string false "搜索关键词"
// @Param type query string false "仓库类型"
// @Param page query int true "页码"
// @Param page_size query int true "每页数量"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/registry/list [get]
func (c *ImageRegistryController) List(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req services.RegistryListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewImageRegistryService()
	list, total, err := svc.List(&req)
	if err != nil {
		global.Logger.Error("获取镜像仓库列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorGetImageRegistryListFail)
		return
	}

	resp.SuccessList(list, int(total))
}

// ListAll godoc
// @Summary 获取所有镜像仓库
// @Description 获取所有镜像仓库（不分页，用于下拉选择）
// @Tags 镜像管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} string "成功"
// @Router /api/v1/image/registry/all [get]
func (c *ImageRegistryController) ListAll(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	svc := services.NewImageRegistryService()
	list, err := svc.ListAll()
	if err != nil {
		global.Logger.Error("获取镜像仓库列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorGetImageRegistryListFail)
		return
	}

	resp.SuccessList(list, len(list))
}

// Detail godoc
// @Summary 获取镜像仓库详情
// @Description 根据ID获取镜像仓库详情
// @Tags 镜像管理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "仓库ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/registry/detail [get]
func (c *ImageRegistryController) Detail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的仓库ID"))
		return
	}

	svc := services.NewImageRegistryService()
	registry, err := svc.GetByID(id)
	if err != nil {
		global.Logger.Error("获取镜像仓库详情失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorImageRegistryNotFound)
		return
	}

	resp.Success(registry)
}

// Create godoc
// @Summary 创建镜像仓库
// @Description 创建新的镜像仓库
// @Tags 镜像管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body services.RegistryCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/registry/create [post]
func (c *ImageRegistryController) Create(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req services.RegistryCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	// 获取当前用户ID
	userID := int64(0)
	if uid, exists := ctx.Get("user_id"); exists {
		userID = uid.(int64)
	}

	svc := services.NewImageRegistryService()
	registry, err := svc.Create(&req, userID)
	if err != nil {
		global.Logger.Error("创建镜像仓库失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorCreateImageRegistryFail.WithDetails(err.Error()))
		return
	}

	resp.Success(registry)
}

// Update godoc
// @Summary 更新镜像仓库
// @Description 更新镜像仓库信息
// @Tags 镜像管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param body body services.RegistryUpdateRequest true "更新参数"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/registry/update [post]
func (c *ImageRegistryController) Update(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req services.RegistryUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewImageRegistryService()
	registry, err := svc.Update(&req)
	if err != nil {
		global.Logger.Error("更新镜像仓库失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorUpdateImageRegistryFail.WithDetails(err.Error()))
		return
	}

	resp.Success(registry)
}

// Delete godoc
// @Summary 删除镜像仓库
// @Description 删除镜像仓库
// @Tags 镜像管理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "仓库ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/registry/delete [post]
func (c *ImageRegistryController) Delete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的仓库ID"))
		return
	}

	svc := services.NewImageRegistryService()
	if err := svc.Delete(id); err != nil {
		global.Logger.Error("删除镜像仓库失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorDeleteImageRegistryFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "删除成功"})
}

// CheckConnection godoc
// @Summary 检测仓库连接
// @Description 检测镜像仓库连接状态
// @Tags 镜像管理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "仓库ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/registry/check [post]
func (c *ImageRegistryController) CheckConnection(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的仓库ID"))
		return
	}

	svc := services.NewImageRegistryService()
	if err := svc.CheckConnection(id); err != nil {
		global.Logger.Error("检测仓库连接失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorCheckImageRegistryFail)
		return
	}

	// 获取最新状态
	registry, _ := svc.GetByID(id)
	resp.Success(registry)
}

// Stats godoc
// @Summary 获取仓库统计
// @Description 获取镜像仓库统计信息
// @Tags 镜像管理
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} string "成功"
// @Router /api/v1/image/registry/stats [get]
func (c *ImageRegistryController) Stats(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	svc := services.NewImageRegistryService()
	stats, err := svc.GetStats()
	if err != nil {
		global.Logger.Error("获取仓库统计失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorGetImageRegistryStatsFail)
		return
	}

	resp.Success(stats)
}

// SetDefault godoc
// @Summary 设置默认仓库
// @Description 设置默认镜像仓库
// @Tags 镜像管理
// @Produce json
// @Security ApiKeyAuth
// @Param id query int true "仓库ID"
// @Success 200 {object} string "成功"
// @Router /api/v1/image/registry/default [post]
func (c *ImageRegistryController) SetDefault(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的仓库ID"))
		return
	}

	svc := services.NewImageRegistryService()
	if err := svc.SetDefault(id); err != nil {
		global.Logger.Error("设置默认仓库失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ErrorSetDefaultRegistryFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "设置成功"})
}
