package appconfig

import (
	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

// KubeAppConfigController AppConfig 控制器
type KubeAppConfigController struct{}

func NewKubeAppConfigController() *KubeAppConfigController {
	return &KubeAppConfigController{}
}

/* -------------------- Create -------------------- */

// Create godoc
// @Summary 创建 AppConfig
// @Tags Operator AppConfig
// @Accept json
// @Produce json
// @Param body body requests.KubeAppConfigCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/appconfig/create [post]
func (c *KubeAppConfigController) Create(ctx *gin.Context) {
	param := requests.NewKubeAppConfigCreateRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeAppConfigCreateRequest); !ok {
		return
	}

	svc := services.NewServices()
	data, err := svc.KubeAppConfigCreate(ctx, param)
	if err != nil {
		ctx.Error(err)
		return
	}

	r.Success(data)
}

/* -------------------- Update -------------------- */

// Update godoc
// @Summary 更新 AppConfig
// @Tags Operator AppConfig
// @Accept json
// @Produce json
// @Param body body requests.KubeAppConfigUpdateRequest true "更新参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/appconfig/update [put]
func (c *KubeAppConfigController) Update(ctx *gin.Context) {
	param := requests.NewKubeAppConfigUpdateRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeAppConfigUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	data, err := svc.KubeAppConfigUpdate(ctx, param)
	if err != nil {
		ctx.Error(err)
		return
	}

	r.Success(data)
}

/* -------------------- Detail -------------------- */

// Detail godoc
// @Summary 获取 AppConfig 详情
// @Tags Operator AppConfig
// @Accept json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param app_name query string true "App 名称"
// @Param cluster_id query int false "集群 ID"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/appconfig/detail [get]
func (c *KubeAppConfigController) Detail(ctx *gin.Context) {
	param := requests.NewKubeAppConfigDetailRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeAppConfigDetailRequest); !ok {
		return
	}

	svc := services.NewServices()
	data, err := svc.KubeAppConfigGet(ctx, param)
	if err != nil {
		ctx.Error(err)
		return
	}

	r.Success(data)
}

/* -------------------- List -------------------- */

// List godoc
// @Summary 获取 AppConfig 列表
// @Tags Operator AppConfig
// @Accept json
// @Produce json
// @Param namespace query string false "命名空间"
// @Param cluster_id query int false "集群 ID"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/appconfig/list [get]
func (c *KubeAppConfigController) List(ctx *gin.Context) {
	param := requests.NewKubeAppConfigListRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeAppConfigListRequest); !ok {
		return
	}

	svc := services.NewServices()
	data, err := svc.KubeAppConfigList(ctx, param)
	if err != nil {
		ctx.Error(err)
		return
	}

	r.Success(data)
}

/* -------------------- Delete -------------------- */

// Delete godoc
// @Summary 删除 AppConfig
// @Tags Operator AppConfig
// @Accept json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param app_name query string true "App 名称"
// @Param cluster_id query int false "集群 ID"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/appconfig/delete [delete]
func (c *KubeAppConfigController) Delete(ctx *gin.Context) {
	param := requests.NewKubeAppConfigDeleteRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeAppConfigDeleteRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.KubeAppConfigDelete(ctx, param); err != nil {
		ctx.Error(err)
		return
	}

	r.Success("deleted")
}
