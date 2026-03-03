package user

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type UserController struct {
}

func NewUserController() *UserController {
	return &UserController{}
}

// Create godoc
// @Summary 创建用户
// @Description 创建用户
// @Tags 用户管理
// @Produce json
// @Param body body requests.UserCreateRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/user/create [post]
func (c *UserController) Create(ctx *gin.Context) {
	// 创建一个新的用户创建请求参数对象
	param := requests.NewUserUserCreateRequest()
	// 创建一个新的响应对象，用于返回处理结果
	resp := response.NewResponse(ctx)

	// 验证请求参数，如果验证失败则直接返回
	if ok := valid.Validate(ctx, param, requests.VaildUserCreateRequest); !ok {
		return
	}

	// 创建一个新的服务对象
	svc := services.NewServices()
	// 调用服务层的用户创建方法，如果创建失败则记录错误并返回错误响应
	if err := svc.UserCreate(param); err != nil {
		global.Logger.Error("创建用户失败,", zap.String("error", err.Error())) // 记录错误日志
		resp.ToErrorResponse(errorcode.ErrorUserCreateFail)              // 返回创建失败的错误响应
		return
	}
	// 如果创建成功，返回成功响应
	resp.Success("创建用户成功")
}

// @Summary 删除用户
// @Description 删除用户
// @Tags 用户管理
// @Produce json
// @Param body body requests.CommonIdRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/user/delete [post]
func (c *UserController) Delete(ctx *gin.Context) {
	param := requests.NewCommonIdRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidCommonIdRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.UserDelete(param); err != nil {
		global.Logger.Error("删除用户失败,", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorUserDeleteFail)
		return
	}

	resp.Success(gin.H{
		"msg": "用户删除成功",
	})
}

// @Summary 更新用户
// @Description 更新用户
// @Tags 用户管理
// @Produce json
// @Param body body requests.UserUpdateRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/user/update [post]
func (u *UserController) Update(ctx *gin.Context) {
	param := requests.NewUserUpdateRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidUserUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.UserUpdate(param); err != nil {
		global.Logger.Error("更新用户失败,", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorUserUpdateFail)
		return
	}

	resp.Success(gin.H{
		"msg": "用户更新成功",
	})
}

// Create godoc
// @Summary 列出用户
// @Description 列出用户
// @Tags 用户管理
// @Produce json
// @Security ApiKeyAuth
// @Param username query string false "用户名" maxlength(100)
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/user/list [get]
func (c *UserController) List(ctx *gin.Context) {
	param := requests.NewUserListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidUserListRequest); !ok {
		return
	}

	svc := services.NewServices()
	users, err := svc.UserList(param)
	if err != nil {
		global.Logger.Error("获取用户列表失败,", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorUserListFail)
		return
	}

	resp.SuccessList(users, len(users))
}
