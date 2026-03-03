package cicd

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type CicdReleaseController struct {
}

func NewCicdReleaseController() *CicdReleaseController {
	return &CicdReleaseController{}
}

// @Summary 创建 CICD 发布单
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseCreateRequest true "创建参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/cicd/release/create [post]
func (c *CicdReleaseController) Create(ctx *gin.Context) {
	param := requests.NewCicdReleaseCreateRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseCreateRequest); !ok {
		return
	}

	// 获取当前用户 ID
	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	id, err := svc.CicdReleaseCreate(ctx.Request.Context(), param, userID)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("CicdReleaseCreate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseCreateFail.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"release_id": id})
}

// Detail godoc
// @Summary 获取发布单详情
// @Tags CICD Release
// @Produce json
// @Param id query int true "发布单ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/detail [get]
func (c *CicdReleaseController) Detail(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的发布单ID"))
		return
	}

	svc := services.NewServices()
	rel, tasks, err := svc.CicdReleaseDetail(ctx.Request.Context(), id)
	if err != nil {
		global.Logger.Error("CicdReleaseDetail error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"release": rel,
		"tasks":   tasks,
	})
}

// List godoc
// @Summary 获取发布单列表
// @Tags CICD Release
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param app_name query string false "应用名称"
// @Param status query string false "状态"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/list [get]
func (c *CicdReleaseController) List(ctx *gin.Context) {
	param := requests.NewCicdReleaseListRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseListRequest); !ok {
		return
	}

	svc := services.NewServices()
	list, total, err := svc.CicdReleaseList(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("CicdReleaseList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// Cancel godoc
// @Summary 取消发布单
// @Description 智能取消：已部署成功/运行中的会触发回滚，未部署的直接取消
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseIDRequest true "取消参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/cancel [post]
func (c *CicdReleaseController) Cancel(ctx *gin.Context) {
	param := &requests.CicdReleaseIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseIDRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	result, err := svc.CicdReleaseCancel(ctx.Request.Context(), param.ID, userID)
	if err != nil {
		global.Logger.Error("CicdReleaseCancel error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseCancelFail.WithDetails(err.Error()))
		return
	}

	// 根据操作类型返回不同的消息
	if result.Action == "rollback" {
		rsp.Success(gin.H{
			"message":            "发布单已触发回滚",
			"action":             result.Action,
			"rollback_release_id": result.RollbackReleaseID,
		})
	} else {
		rsp.Success(gin.H{
			"message": "取消成功",
			"action":  result.Action,
		})
	}
}

// Rollback godoc
// @Summary 回滚发布单
// @Description 将已部署的工作负载回滚到上一个版本，会创建新的发布单执行回滚操作
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseIDRequest true "回滚参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/rollback [post]
func (c *CicdReleaseController) Rollback(ctx *gin.Context) {
	param := &requests.CicdReleaseIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseIDRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	newID, err := svc.CicdReleaseRollback(ctx.Request.Context(), param.ID, userID)
	if err != nil {
		global.Logger.Error("CicdReleaseRollback error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseRollbackFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{
		"message":        "回滚成功",
		"rollback_release_id": newID,
	})
}

// Retry godoc
// @Summary 重试发布单
// @Tags CICD Release
// @Accept json
// @Produce json
// @Param body body requests.CicdReleaseIDRequest true "重试参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/retry [post]
func (c *CicdReleaseController) Retry(ctx *gin.Context) {
	param := &requests.CicdReleaseIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCicdReleaseIDRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")
	svc := services.NewServices()
	newID, err := svc.CicdReleaseRetry(ctx.Request.Context(), param.ID, userID)
	if err != nil {
		global.Logger.Error("CicdReleaseRetry error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseRetryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"release_id": newID})
}

// Tasks godoc
// @Summary 获取发布单下的任务列表
// @Tags CICD Release
// @Produce json
// @Param release_id query int true "发布单ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/release/tasks [get]
func (c *CicdReleaseController) Tasks(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("release_id")
	releaseID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || releaseID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的发布单ID"))
		return
	}

	svc := services.NewServices()
	tasks, err := svc.CicdTasksByRelease(ctx.Request.Context(), releaseID)
	if err != nil {
		global.Logger.Error("CicdTasksByRelease error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdReleaseQueryFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"tasks": tasks})
}

// BuildCallback godoc
// @Summary Jenkins 构建回调
// @Tags CICD Callback
// @Accept json
// @Produce json
// @Param body body requests.CicdBuildCallbackRequest true "回调参数"
// @Success 200 {object} map[string]any
// @Router /api/v1/k8s/cicd/callback/build [post]
func (c *CicdReleaseController) BuildCallback(ctx *gin.Context) {
	param := &requests.CicdBuildCallbackRequest{}
	rsp := response.NewResponse(ctx)

	if err := ctx.ShouldBindJSON(param); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if err := svc.CicdBuildCallback(ctx.Request.Context(), param); err != nil {
		global.Logger.Error("CicdBuildCallback error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorCicdBuildCallbackFail.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "回调处理成功"})
}
