package v1

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

// AppStoreController 应用商城控制器
type AppStoreController struct {
	factory *services.ClusterClientFactory
}

func NewAppStoreController() *AppStoreController {
	return &AppStoreController{}
}

func NewAppStoreControllerWithFactory(factory *services.ClusterClientFactory) *AppStoreController {
	return &AppStoreController{factory: factory}
}

// List 应用列表
// @Summary 获取应用商城列表
// @Description 分页获取应用列表，支持分类筛选和关键词搜索
// @Tags 应用商城
// @Produce json
// @Param category query string false "分类"
// @Param keyword query string false "搜索关键词"
// @Param status query int false "状态"
// @Param featured query int false "推荐"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/list [get]
func (c *AppStoreController) List(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req models.AppStoreListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		global.Logger.Error("解析应用列表请求失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	list, total, err := svc.AppStoreList(ctx.Request.Context(), &req)
	if err != nil {
		global.Logger.Error("获取应用列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreListFailed)
		return
	}

	resp.SuccessList(list, total)
}

// Detail 应用详情
// @Summary 获取应用详情
// @Description 根据 ID 获取应用详情
// @Tags 应用商城
// @Produce json
// @Param id path int true "应用ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/detail/{id} [get]
func (c *AppStoreController) Detail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的应用ID"))
		return
	}

	svc := services.NewServices()
	app, err := svc.AppStoreDetail(ctx.Request.Context(), uint32(id))
	if err != nil {
		global.Logger.Error("获取应用详情失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreNotFound)
		return
	}

	resp.Success(app)
}

// Create 创建应用
// @Summary 创建应用
// @Description 管理员创建新应用
// @Tags 应用商城
// @Accept json
// @Produce json
// @Param body body models.AppStoreCreateRequest true "应用信息"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/create [post]
func (c *AppStoreController) Create(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req models.AppStoreCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		global.Logger.Error("解析创建请求失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	app, err := svc.AppStoreCreate(ctx.Request.Context(), &req)
	if err != nil {
		global.Logger.Error("创建应用失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreCreateFailed.WithDetails(err.Error()))
		return
	}

	resp.Success(app)
}

// Update 更新应用
// @Summary 更新应用
// @Description 管理员更新应用信息
// @Tags 应用商城
// @Accept json
// @Produce json
// @Param body body models.AppStoreUpdateRequest true "应用信息"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/update [put]
func (c *AppStoreController) Update(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req models.AppStoreUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		global.Logger.Error("解析更新请求失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.AppStoreUpdate(ctx.Request.Context(), &req); err != nil {
		global.Logger.Error("更新应用失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreUpdateFailed.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "更新成功"})
}

// Delete 删除应用
// @Summary 删除应用
// @Description 管理员删除应用（软删除）
// @Tags 应用商城
// @Produce json
// @Param id path int true "应用ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/delete/{id} [delete]
func (c *AppStoreController) Delete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的应用ID"))
		return
	}

	svc := services.NewServices()
	if err := svc.AppStoreDelete(ctx.Request.Context(), uint32(id)); err != nil {
		global.Logger.Error("删除应用失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreDeleteFailed.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "删除成功"})
}

// Categories 获取分类列表
// @Summary 获取分类列表
// @Description 获取所有分类及其应用计数
// @Tags 应用商城
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/categories [get]
func (c *AppStoreController) Categories(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	svc := services.NewServices()
	categories, err := svc.AppStoreCategories(ctx.Request.Context())
	if err != nil {
		global.Logger.Error("获取分类列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreListFailed)
		return
	}

	resp.Success(categories)
}

// Install 安装应用到集群
// @Summary 安装应用
// @Description 将应用安装到指定集群的命名空间
// @Tags 应用商城
// @Accept json
// @Produce json
// @Param body body models.AppStoreInstallRequest true "安装信息"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/install [post]
func (c *AppStoreController) Install(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if c.factory == nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails("集群工厂未初始化"))
		return
	}

	var req models.AppStoreInstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		global.Logger.Error("解析安装请求失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	// 获取当前操作用户
	operator, _ := ctx.Get("username")
	operatorStr, _ := operator.(string)
	if operatorStr == "" {
		operatorStr = "admin"
	}

	svc := services.NewServices()
	install, err := svc.AppStoreInstall(ctx.Request.Context(), c.factory, &req, operatorStr)
	if err != nil {
		global.Logger.Error("安装应用失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreInstallFailed.WithDetails(err.Error()))
		return
	}

	resp.Success(install)
}

// Uninstall 卸载应用
// @Summary 卸载应用
// @Description 从集群卸载已安装的应用
// @Tags 应用商城
// @Produce json
// @Param id path int true "安装记录ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/uninstall/{id} [post]
func (c *AppStoreController) Uninstall(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if c.factory == nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails("集群工厂未初始化"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的安装记录ID"))
		return
	}

	svc := services.NewServices()
	if err := svc.AppStoreUninstall(ctx.Request.Context(), c.factory, uint32(id)); err != nil {
		global.Logger.Error("卸载应用失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreUninstallFailed.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "卸载成功"})
}

// InstallList 安装记录列表
// @Summary 获取安装记录列表
// @Description 分页获取应用安装记录
// @Tags 应用商城
// @Produce json
// @Param app_id query int false "应用ID"
// @Param cluster_id query int false "集群ID"
// @Param status query int false "状态"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/installs [get]
func (c *AppStoreController) InstallList(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req models.AppStoreInstallListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	list, total, err := svc.AppStoreInstallList(ctx.Request.Context(), &req)
	if err != nil {
		global.Logger.Error("获取安装记录失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreListFailed)
		return
	}

	resp.SuccessList(list, total)
}

// InstallDetail 安装记录详情
// @Summary 获取安装记录详情
// @Tags 应用商城
// @Produce json
// @Param id path int true "安装记录ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/installs/{id} [get]
func (c *AppStoreController) InstallDetail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的安装记录ID"))
		return
	}

	svc := services.NewServices()
	install, err := svc.AppStoreInstallDetail(ctx.Request.Context(), uint32(id))
	if err != nil {
		resp.ToErrorResponse(errorcode.AppStoreInstallNotFound)
		return
	}

	resp.Success(install)
}

// InstallStatus 安装实时状态（查询 K8s 集群真实 Pod/Deployment/Service 状态）
// @Summary 获取安装实时状态
// @Description 实时查询已安装应用在 K8s 集群中的 Deployment/Pod/Service 状态
// @Tags 应用商城
// @Produce json
// @Param id path int true "安装记录ID"
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/platform/appstore/installs/{id}/status [get]
func (c *AppStoreController) InstallStatus(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if c.factory == nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails("集群工厂未初始化"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的安装记录ID"))
		return
	}

	svc := services.NewServices()
	status, err := svc.AppStoreInstallStatus(ctx.Request.Context(), c.factory, uint32(id))
	if err != nil {
		global.Logger.Error("查询安装状态失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreInstallNotFound.WithDetails(err.Error()))
		return
	}

	resp.Success(status)
}

// InstallUpdate 编辑安装（更新 Deployment 参数）
func (c *AppStoreController) InstallUpdate(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	if c.factory == nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails("集群工厂未初始化"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的安装记录ID"))
		return
	}

	var req models.AppStoreInstallUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.AppStoreInstallUpdate(ctx.Request.Context(), c.factory, uint32(id), &req); err != nil {
		global.Logger.Error("更新安装失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "更新成功"})
}

// ============================================================
// 组件管理 API
// ============================================================

// ComponentList 获取应用的组件列表
func (c *AppStoreController) ComponentList(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的应用ID"))
		return
	}

	svc := services.NewServices()
	list, err := svc.AppStoreComponentList(ctx.Request.Context(), uint32(id))
	if err != nil {
		global.Logger.Error("获取组件列表失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AppStoreListFailed.WithDetails(err.Error()))
		return
	}

	resp.Success(list)
}

// ComponentCreate 创建组件
func (c *AppStoreController) ComponentCreate(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req models.AppStoreComponentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	comp, err := svc.AppStoreComponentCreate(ctx.Request.Context(), &req)
	if err != nil {
		global.Logger.Error("创建组件失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(comp)
}

// ComponentUpdate 更新组件
func (c *AppStoreController) ComponentUpdate(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req models.AppStoreComponentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.AppStoreComponentUpdate(ctx.Request.Context(), &req); err != nil {
		global.Logger.Error("更新组件失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "更新成功"})
}

// ComponentDelete 删除组件
func (c *AppStoreController) ComponentDelete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	idStr := ctx.Param("comp_id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的组件ID"))
		return
	}

	svc := services.NewServices()
	if err := svc.AppStoreComponentDelete(ctx.Request.Context(), uint32(id)); err != nil {
		global.Logger.Error("删除组件失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "删除成功"})
}

// ComponentBatchDelete 批量删除组件
func (c *AppStoreController) ComponentBatchDelete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req models.AppStoreComponentBatchDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.AppStoreComponentBatchDelete(ctx.Request.Context(), req.IDs); err != nil {
		global.Logger.Error("批量删除组件失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "批量删除成功"})
}

// ComponentSort 批量更新组件排序
func (c *AppStoreController) ComponentSort(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	var req models.AppStoreComponentSortRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.AppStoreComponentSort(ctx.Request.Context(), &req); err != nil {
		global.Logger.Error("更新组件排序失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "排序更新成功"})
}
