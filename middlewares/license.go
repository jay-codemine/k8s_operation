package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/license"
)

// licenseWhitelist License 闸门放行的路径前缀
// 未授权时仅允许：登录相关（激活需管理员登录）、License 状态/激活接口、健康探针
var licenseWhitelist = []string{
	"/api/v1/auth/",             // 登录/登出/刷新
	"/api/v1/platform/license",  // License 状态查询/激活
	"/api/v1/helloworld",        // 健康检查
}

// LicenseGate License 授权闸门中间件（挂在 /api/v1 最外层）
// 平台未激活或授权过期时，除白名单外所有接口返回 110001/110002，
// 前端拦截该错误码跳转激活页
func LicenseGate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 白名单直接放行
		path := ctx.Request.URL.Path
		for _, prefix := range licenseWhitelist {
			if strings.HasPrefix(path, prefix) {
				ctx.Next()
				return
			}
		}

		ok, reason := license.Valid()
		if ok {
			ctx.Next()
			return
		}

		rsp := response.NewResponse(ctx)
		if strings.Contains(reason, "过期") {
			rsp.ToErrorResponse(errorcode.LicenseExpired.WithDetails(reason))
		} else {
			rsp.ToErrorResponse(errorcode.LicenseRequired.WithDetails(reason))
		}
		ctx.Abort()
	}
}
