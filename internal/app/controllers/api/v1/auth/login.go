package auth

import (
	"encoding/json"
	"fmt"
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
	ldapclient "k8soperation/pkg/ldap"
	"k8soperation/pkg/utils"
	"k8soperation/pkg/valid"
	"time"
)

// Login godoc
// @Summary 用户登录
// @Description 用户登录（支持本地/LDAP 双模式）
// @Tags 认证管理
// @Produce json
// @Param body body requests.AuthLoginRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/auth/login [post]
func (u *AuthController) Login(ctx *gin.Context) {
	param := requests.NewAuthLoginRequest()
	response := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidAuthLoginRequest); !ok {
		return
	}

	svc := services.NewServices()

	var user *models.User
	var err error
	authMethod := "local" // 认证方式标记

	// ====== LDAP 登录尝试 ======
	if ldapclient.IsEnabled() {
		ldapUser, isLDAP, ldapErr := svc.LDAPLogin(param.Username, param.Password)
		if ldapErr == nil && ldapUser != nil {
			// LDAP 认证成功
			user = ldapUser
			authMethod = "ldap"
		} else if isLDAP && ldapErr != nil && ldapErr.Error() != "LDAP_AUTH_FAILED" {
			// LDAP 认证失败且不回退
			global.Logger.Error("LDAP 登录失败", zap.String("error", ldapErr.Error()))
			response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
			return
		}
		// 如果 LDAP 认证失败且允许回退，继续本地认证
	}

	// ====== 本地登录 ======
	if user == nil {
		user, err = svc.UserLogin(param)
		if err != nil {
			global.Logger.Error("用户登录失败", zap.String("error", err.Error()))
			response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
			return
		}
		if user == nil {
			global.Logger.Error("用户登录失败,用户不存在")
			response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
			return
		}

		// 验证密码（本地模式）
		matched, needMigrate := utils.CheckPasswordSmart(user.Password, param.Password)
		if !matched {
			global.Logger.Error("用户登录失败,密码错误")
			response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
			return
		}

		// 自动迁移旧密码
		if needMigrate {
			go func() {
				if err := svc.MigrateUserPassword(user.ID, param.Password); err != nil {
					global.Logger.Warn("密码迁移失败", zap.Uint32("user_id", user.ID), zap.Error(err))
				}
			}()
		}
	}

	// ====== 统一检查用户状态 ======
	if user.Status == 0 {
		global.Logger.Error("用户登录失败,账号已禁用", zap.String("username", user.Username))
		response.ToErrorResponse(errorcode.ErrorUserDisabled)
		return
	}

	// ====== 签发 JWT ======
	mgr := jwt.NewManager()
	token, err := mgr.IssueToken(cast.ToString(user.ID), user.Username, user.TenantID)
	if err != nil {
		global.Logger.Error("颁发 token 失败", zap.String("error", err.Error()))
		response.ToErrorResponse(errorcode.ErrorAuthLoginFail)
		return
	}

	// ====== 构造响应 ======
	respUser := gin.H{
		"id":          user.ID,
		"username":    user.Username,
		"auth_method": authMethod,
	}

	// ====== 存入 Session（支持同一账户多人并发登录） ======
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

	// 使用 userID + 时间戳 作为 session key，确保同一账户多处登录不会互相覆盖
	sessionKey := fmt.Sprintf("login:%d:%d", user.ID, time.Now().UnixNano())
	session := sessions.Default(ctx)
	session.Set(sessionKey, string(sessionBty))
	if err := session.Save(); err != nil {
		// Session 保存失败不中断登录（JWT 为主认证方式，Session 仅为辅助）
		// Redis Cluster 模式下 gin-contrib/sessions 不支持 MOVED 重定向
		global.Logger.Warn("session 保存失败(不影响登录)", zap.String("error", err.Error()))
	}

	global.Logger.Info("用户登录成功",
		zap.String("username", user.Username),
		zap.String("auth_method", authMethod))

	// ====== 返回 ======
	response.Success(gin.H{
		"user":  respUser,
		"token": token,
	})
}
