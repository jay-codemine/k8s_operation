package v1

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/models"
	"k8soperation/middlewares"
)

// requirePlatformWrite 校验平台级写权限：超管或平台管理员（ScopePlatform + AccessLevelAdmin）
func requirePlatformWrite(ctx *gin.Context) bool {
	if ctx.GetBool("is_super_admin") {
		return true
	}
	if uid, ok := ctx.Get("user_id"); ok {
		if id, ok2 := uid.(int64); ok2 && id > 0 {
			return middlewares.NewServicesFromContext(ctx).CheckScopePermission(id, models.ScopePlatform, models.AccessLevelAdmin)
		}
	}
	return false
}
