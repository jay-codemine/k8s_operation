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

type TemplateController struct {
}

func NewTemplateController() *TemplateController {
	return &TemplateController{}
}

// Create godoc
// @Summary 创建流水线模板
// @Description 创建新的流水线模板，可用于快速创建流水线
// @Tags CICD Template
// @Accept json
// @Produce json
// @Param body body requests.TemplateCreateRequest true "创建参数"
// @Success 200 {object} map[string]any "返回 template_id"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/k8s/cicd/template/create [post]
func (c *TemplateController) Create(ctx *gin.Context) {
	param := requests.NewTemplateCreateRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidTemplateCreateRequest); !ok {
		return
	}

	userID := ctx.GetInt64("user_id")

	svc := services.NewServices()
	id, err := svc.TemplateCreate(ctx.Request.Context(), param, userID)
	if err != nil {
		global.Logger.Error("TemplateCreate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"template_id": id})
}

// Detail godoc
// @Summary 获取流水线模板详情
// @Description 获取模板的完整配置信息
// @Tags CICD Template
// @Produce json
// @Param id query int true "模板ID"
// @Success 200 {object} map[string]any "返回模板详情"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/k8s/cicd/template/detail [get]
func (c *TemplateController) Detail(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	idStr := ctx.Query("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的模板ID"))
		return
	}

	svc := services.NewServices()
	template, err := svc.TemplateDetail(ctx.Request.Context(), id)
	if err != nil {
		global.Logger.Error("TemplateDetail error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"template": template})
}

// List godoc
// @Summary 获取流水线模板列表
// @Description 分页查询模板列表，支持关键字和类型筛选
// @Tags CICD Template
// @Produce json
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页数量，默认10"
// @Param keyword query string false "关键字（名称/描述）"
// @Param type query string false "类型筛选（frontend/backend/microservice/database/custom）"
// @Success 200 {object} map[string]any "返回列表和总数"
// @Router /api/v1/k8s/cicd/template/list [get]
func (c *TemplateController) List(ctx *gin.Context) {
	param := requests.NewTemplateListRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidTemplateListRequest); !ok {
		return
	}

	svc := services.NewServices()
	list, total, err := svc.TemplateList(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("TemplateList error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.SuccessList(list, total)
}

// Update godoc
// @Summary 更新流水线模板
// @Description 更新模板配置信息
// @Tags CICD Template
// @Accept json
// @Produce json
// @Param body body requests.TemplateUpdateRequest true "更新参数"
// @Success 200 {object} map[string]any "更新成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/k8s/cicd/template/update [post]
func (c *TemplateController) Update(ctx *gin.Context) {
	param := requests.NewTemplateUpdateRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidTemplateUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	err := svc.TemplateUpdate(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("TemplateUpdate error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "更新成功"})
}

// Delete godoc
// @Summary 删除流水线模板
// @Description 软删除模板
// @Tags CICD Template
// @Accept json
// @Produce json
// @Param body body requests.TemplateIDRequest true "删除参数"
// @Success 200 {object} map[string]any "删除成功"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "内部错误"
// @Router /api/v1/k8s/cicd/template/delete [post]
func (c *TemplateController) Delete(ctx *gin.Context) {
	param := &requests.TemplateIDRequest{}
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidTemplateIDRequest); !ok {
		return
	}

	svc := services.NewServices()
	err := svc.TemplateDelete(ctx.Request.Context(), param.ID)
	if err != nil {
		global.Logger.Error("TemplateDelete error", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	rsp.Success(gin.H{"message": "删除成功"})
}
