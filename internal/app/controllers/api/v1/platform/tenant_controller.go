package platform

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"k8soperation/global"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/models"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

type TenantController struct{}

func NewTenantController() *TenantController { return &TenantController{} }

// requireSuperAdmin 仅超级管理员可管理租户，无权限时返回 false（已自动响应错误）
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

	var tenants []models.Tenant
	db := global.DB.Where("is_del = 0")

	// 普通用户只能看到自己所属的租户
	// 注意：auth 中间件写入的是 uint32，gin 的 GetUint 只断言 uint（64位），会静默得到 0
	if !ctx.GetBool("is_super_admin") {
		if v, exists := ctx.Get("tenant_id"); exists {
			if tid, ok := v.(uint32); ok {
				db = db.Where("id = ?", tid)
			}
		}
	}

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

	// code 唯一（含已软删记录，uk_code 唯一索引不含 is_del）
	var cnt int64
	global.DB.Model(&models.Tenant{}).Where("code = ?", req.Code).Count(&cnt)
	if cnt > 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("租户编码已存在"))
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
	// 租户记录和它的默认角色必须同生共死：只建了 tenant 行、没建 super_admin 角色的
	// 租户是个不可用的半成品（该租户所有用户的权限判定都会返回 false）
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tenant).Error; err != nil {
			return err
		}
		return dao.TenantSeedRBAC(tx, tenant.ID)
	})
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
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

	// 记录必须存在
	var exist int64
	global.DB.Model(&models.Tenant{}).Where("id = ? AND is_del = 0", id).Count(&exist)
	if exist == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("租户不存在"))
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

	// code 与其他租户冲突时拒绝（uk_code 唯一索引不含 is_del）
	if req.Code != "" {
		var cnt int64
		global.DB.Model(&models.Tenant{}).Where("code = ? AND id <> ?", req.Code, id).Count(&cnt)
		if cnt > 0 {
			rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("租户编码已存在"))
			return
		}
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
	if !requireSuperAdmin(ctx) {
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的租户ID"))
		return
	}
	if id == 1 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("不能删除默认租户"))
		return
	}

	// 记录必须存在
	var exist int64
	global.DB.Model(&models.Tenant{}).Where("id = ? AND is_del = 0", id).Count(&exist)
	if exist == 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("租户不存在"))
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
