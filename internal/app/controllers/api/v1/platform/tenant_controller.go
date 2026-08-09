package platform

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
)

type TenantController struct{}

func NewTenantController() *TenantController { return &TenantController{} }

func requireSuperAdmin(ctx *gin.Context) bool {
	if !ctx.GetBool("is_super_admin") {
		rsp := response.NewResponse(ctx)
		rsp.ToErrorResponse(errorcode.ErrorRBACAccessDenied.WithDetails("仅超级管理员可管理租户"))
		return false
	}
	return true
}

// List 获取租户列表（超级管理员看全部，普通用户只看自己的）
func (c *TenantController) List(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	svc := middlewares.NewServicesFromContext(ctx)

	tenants, err := svc.TenantList(ctx.Request.Context(), ctx.GetBool("is_super_admin"), getTenantID(ctx))
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}
	rsp.Success(gin.H{"items": tenants})
}

// Create 创建租户
func (c *TenantController) Create(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	if !requireSuperAdmin(ctx) {
		return
	}

	var req struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Name == "" || req.Code == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("name 和 code 为必填"))
		return
	}

	tenant, err := middlewares.NewServicesFromContext(ctx).TenantCreate(ctx.Request.Context(), req.Name, req.Code)
	if err != nil {
		if err == services.ErrTenantCodeExists {
			rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		} else {
			rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		}
		return
	}
	rsp.Success(tenant)
}

// Update 更新租户
func (c *TenantController) Update(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	if !requireSuperAdmin(ctx) {
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的租户ID"))
		return
	}

	var req struct {
		Name   string `json:"name"`
		Code   string `json:"code"`
		Status *int8  `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	err = middlewares.NewServicesFromContext(ctx).TenantUpdate(ctx.Request.Context(), uint32(id), req.Name, req.Code, req.Status)
	if err != nil {
		switch err {
		case services.ErrTenantNotFound:
			rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		case services.ErrTenantCodeExists:
			rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		default:
			rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		}
		return
	}
	rsp.Success(nil)
}

// Delete 删除租户（软删除）
func (c *TenantController) Delete(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	if !requireSuperAdmin(ctx) {
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的租户ID"))
		return
	}

	err = middlewares.NewServicesFromContext(ctx).TenantDelete(ctx.Request.Context(), uint32(id))
	if err != nil {
		switch err {
		case services.ErrTenantDefaultCannotDelete, services.ErrTenantNotFound:
			rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		default:
			rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		}
		return
	}
	rsp.Success(nil)
}

func getTenantID(ctx *gin.Context) uint32 {
	if v, exists := ctx.Get("tenant_id"); exists {
		if tid, ok := v.(uint32); ok {
			return tid
		}
	}
	return 0
}
