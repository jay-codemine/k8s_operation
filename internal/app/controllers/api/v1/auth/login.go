package auth

import (
	"encoding/json"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/jwt"
	"k8soperation/pkg/utils"
	"k8soperation/pkg/valid"
	"time"
)

// Create godoc
// @Summary 用户登录
// @Description 用户登录
// @Tags 认证管理
// @Produce json
// @Param body body requests.AuthLoginRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/auth/login [post]
// AuthController 结构体的登录处理方法
// 处理用户登录请求，验证用户信息并返回token
func (u *AuthController) Login(ctx *gin.Context) {
	// 创建登录请求参数结构体
	param := requests.NewAuthLoginRequest()
	// 创建响应结构体
	response := response.NewResponse(ctx)

	// 验证请求参数是否合法，如果不合法则直接返回
	if ok := valid.Validate(ctx, param, requests.ValidAuthLoginRequest); !ok {
		return
	}
	// 创建服务层实例
	svc := services.NewServices()

	// 调用服务层的用户登录方法
	user, err := svc.UserLogin(param)

	// 处理登录错误情况
	if err != nil {
		// 记录登录失败的错误日志
		global.Logger.Error("用户登录失败,", zap.String("error", err.Error()))
		// 返回登录失败的错误响应
		response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
		return
	}

	// 用户不存在
	if user == nil {
		global.Logger.Error("用户登录失败,用户不存在")
		response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
		return
	}

	// 验证用户密码是否正确
	if user.Password != param.Password {
		// 记录密码错误的日志
		global.Logger.Error("用户登录失败,密码错误")
		// 返回登录失败的错误响应
		response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
		return
	}

	// 生成 JWT token（必须接收两个返回值）
	mgr := jwt.NewManager()
	token, err := mgr.IssueToken(cast.ToString(user.ID), user.Username)
	if err != nil {
		global.Logger.Error("颁发 token 失败", zap.String("error", err.Error()))
		response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
		return
	}

	// 构造安全的用户返回体（别把密码等敏感字段直接回给前端）
	respUser := gin.H{
		"id":       user.ID,
		"username": user.Username,
		// 需要的话补充其他非敏感字段
	}

	// ====== 存入 Session ======
	sessionInfo := models.LoginSessionInfo{
		Username:  param.Username,
		Token:     token,
		LoginTime: time.Now(),
	}

	sessionBty, err := json.Marshal(sessionInfo)
	if err != nil {
		global.Logger.Error("序列化 session 失败", zap.String("error", err.Error()))
		response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
		return
	}

	session := sessions.Default(ctx)
	session.Set(utils.EncodeMD5(user.Username), string(sessionBty))
	if err := session.Save(); err != nil {
		global.Logger.Error("session 保存失败", zap.String("error", err.Error()))
		response.ToErrorResponse(errorcode.ServerError)
		return
	}

	// ====== 最终返回 ======
	response.Success(gin.H{
		"user":  respUser,
		"token": token,
	})
}
