package platform

import (
	"github.com/gin-gonic/gin"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

// TenantController 租户控制器
type TenantController struct{}

func NewTenantController() *TenantController { return &TenantController{} }

// List 获取租户列表（超级管理员看全部，普通用户只看自己的）
func (c *TenantController) List(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	tid, _ := ctx.Get("tenant_id")
	uid, _ := ctx.Get("user_id")

	var tenants []models.Tenant
	db := global.DB

	// 普通用户只看自己的租户
	if tid != nil && uid != nil {
		if tidVal, ok := tid.(uint32); ok && tidVal > 0 {
			if uidVal, ok := uid.(int64); ok && uidVal > 0 {
				db = db.Where("id = ?", tidVal)
			}
		}
	}

	if err := db.Where("status = 1").Find(&tenants).Error; err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	if tenants == nil {
		tenants = []models.Tenant{}
	}

	rsp.Success(gin.H{"items": tenants})
}
