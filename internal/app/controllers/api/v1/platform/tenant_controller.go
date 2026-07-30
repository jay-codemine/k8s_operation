package platform

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

type TenantController struct{}

func NewTenantController() *TenantController { return &TenantController{} }

// List 获取租户列表（超级管理员看全部，普通用户只看自己的）
func (c *TenantController) List(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	var tenants []models.Tenant
	db := global.DB.Where("is_del = 0")

	if err := db.Find(&tenants).Error; err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}
	if tenants == nil {
		tenants = []models.Tenant{}
	}
	rsp.Success(gin.H{"items": tenants})
}

// Create 创建租户
func (c *TenantController) Create(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	var req struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Name == "" || req.Code == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("name 和 code 为必填"))
		return
	}

	now := uint32(time.Now().Unix())
	tenant := models.Tenant{
		Name:   req.Name,
		Code:   req.Code,
		Status: 1,
		Base: &models.Base{
			TenantID:   1,
			CreatedAt:  now,
			ModifiedAt: now,
		},
	}
	if err := global.DB.Create(&tenant).Error; err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}
	rsp.Success(tenant)
}

// Update 更新租户
func (c *TenantController) Update(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
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

	values := map[string]interface{}{"modified_at": uint32(time.Now().Unix())}
	if req.Name != "" {
		values["name"] = req.Name
	}
	if req.Code != "" {
		values["code"] = req.Code
	}
	if req.Status != nil {
		values["status"] = *req.Status
	}

	if err := global.DB.Model(&models.Tenant{}).Where("id = ? AND is_del = 0", id).Updates(values).Error; err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}
	rsp.Success(nil)
}

// Delete 删除租户（软删除）
func (c *TenantController) Delete(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的租户ID"))
		return
	}
	if id == 1 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("不能删除默认租户"))
		return
	}

	now := uint32(time.Now().Unix())
	if err := global.DB.Model(&models.Tenant{}).Where("id = ? AND is_del = 0", id).
		Updates(map[string]interface{}{"is_del": 1, "deleted_at": now, "modified_at": now}).Error; err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}
	rsp.Success(nil)
}
