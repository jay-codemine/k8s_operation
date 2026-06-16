package autoscaler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/k8s/hpa"
	"k8soperation/pkg/k8s/vpa"
	"k8soperation/pkg/valid"
)

/* ====================================================================
 * HPA Controller
 * ==================================================================== */

type KubeHPAController struct{}

func NewKubeHPAController() *KubeHPAController { return &KubeHPAController{} }

// List godoc
// @Summary 获取 HPA 列表
// @Tags K8s HPA 弹性扩缩容
// @Produce json
// @Param namespace query string false "命名空间"
// @Param name query string false "HPA 名称(模糊)"
// @Param page query int false "页码"
// @Param limit query int false "每页数量"
// @Router /api/v1/k8s/hpa/list [get]
func (c *KubeHPAController) List(ctx *gin.Context) {
	param := requests.NewKubeHPAListRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeHPAListRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	list, total, err := svc.KubeHPAList(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("KubeHPAList error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sHPAListFail.WithDetails(err.Error()))
		return
	}
	r.SuccessList(list, total)
}

// Detail godoc
// @Summary 获取 HPA 详情
// @Tags K8s HPA 弹性扩缩容
// @Router /api/v1/k8s/hpa/detail [get]
func (c *KubeHPAController) Detail(ctx *gin.Context) {
	param := requests.NewKubeHPADetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeHPADetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	obj, err := svc.KubeHPADetail(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("KubeHPADetail error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sHPADetailFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{
		"item": hpa.BuildHPAItem(obj),
		"raw":  obj,
	})
}

// Create godoc
// @Summary 创建 HPA
// @Tags K8s HPA 弹性扩缩容
// @Accept json
// @Param body body requests.KubeHPACreateRequest true "创建参数"
// @Router /api/v1/k8s/hpa/create [post]
func (c *KubeHPAController) Create(ctx *gin.Context) {
	req := requests.NewKubeHPACreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeHPACreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	obj, err := svc.KubeHPACreate(ctx.Request.Context(), cli, req)
	if err != nil {
		global.Logger.Error("KubeHPACreate error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sHPACreateFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{
		"message": "HPA 创建成功",
		"item":    hpa.BuildHPAItem(obj),
	})
}

// Update godoc
// @Summary 更新 HPA
// @Tags K8s HPA 弹性扩缩容
// @Router /api/v1/k8s/hpa/update [post]
func (c *KubeHPAController) Update(ctx *gin.Context) {
	req := requests.NewKubeHPAUpdateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeHPAUpdateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	obj, err := svc.KubeHPAUpdate(ctx.Request.Context(), cli, req)
	if err != nil {
		global.Logger.Error("KubeHPAUpdate error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sHPAUpdateFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{
		"message": "HPA 更新成功",
		"item":    hpa.BuildHPAItem(obj),
	})
}

// Delete godoc
// @Summary 删除 HPA
// @Tags K8s HPA 弹性扩缩容
// @Router /api/v1/k8s/hpa/delete [delete]
func (c *KubeHPAController) Delete(ctx *gin.Context) {
	param := requests.NewKubeHPADetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeHPADetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	if err := svc.KubeHPADelete(ctx.Request.Context(), cli, param.Namespace, param.Name); err != nil {
		global.Logger.Error("KubeHPADelete error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sHPADeleteFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{"message": "HPA 删除成功", "namespace": param.Namespace, "name": param.Name})
}

// Scale godoc
// @Summary 单独修改 HPA 副本数（min/max）
// @Tags K8s HPA 弹性扩缩容
// @Router /api/v1/k8s/hpa/scale [post]
func (c *KubeHPAController) Scale(ctx *gin.Context) {
	req := requests.NewKubeHPAScaleRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeHPAScaleRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	obj, err := svc.KubeHPAScale(ctx.Request.Context(), cli, req)
	if err != nil {
		global.Logger.Error("KubeHPAScale error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sHPAScaleFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{
		"message": "HPA 副本数已更新",
		"item":    hpa.BuildHPAItem(obj),
	})
}

// CreateFromYaml godoc
// @Summary 通过 YAML 创建 HPA
// @Tags K8s HPA 弹性扩缩容
// @Accept json
// @Param body body requests.YamlCreateRequest true "YAML 内容"
// @Router /api/v1/k8s/hpa/create-from-yaml [post]
func (c *KubeHPAController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	obj, err := svc.KubeHPACreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		global.Logger.Error("KubeHPACreateFromYaml error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sHPACreateFromYamlFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{
		"message":   "HPA 创建成功",
		"name":      obj.Name,
		"namespace": obj.Namespace,
		"item":      hpa.BuildHPAItem(obj),
	})
}

// BatchScale godoc
// @Summary 批量扩容/缩容 HPA（618 促销场景）
// @Tags K8s HPA 弹性扩缩容
// @Accept json
// @Param body body requests.KubeHPABatchScaleRequest true "批量扩缩容"
// @Router /api/v1/k8s/hpa/batch-scale [post]
func (c *KubeHPAController) BatchScale(ctx *gin.Context) {
	req := requests.NewKubeHPABatchScaleRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeHPABatchScaleRequest); !ok {
		return
	}
	if len(req.Items) == 0 {
		r.ToErrorResponse(errorcode.InvalidParams.WithDetails("items 不能为空"))
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	results, successCnt, failCnt := svc.KubeHPABatchScale(ctx.Request.Context(), cli, req)
	r.Success(gin.H{
		"message":      "批量扩缩容完成",
		"total":        len(results),
		"success":      successCnt,
		"fail":         failCnt,
		"results":      results,
	})
}

// BatchStatus godoc
// @Summary 批量查询 HPA 当前状态（统一数据，验证扩缩容是否成功）
// @Tags K8s HPA 弹性扩缩容
// @Router /api/v1/k8s/hpa/batch-status [post]
func (c *KubeHPAController) BatchStatus(ctx *gin.Context) {
	req := requests.NewKubeHPABatchScaleRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeHPABatchScaleRequest); !ok {
		return
	}
	if len(req.Items) == 0 {
		r.ToErrorResponse(errorcode.InvalidParams.WithDetails("items 不能为空"))
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	results := svc.KubeHPABatchStatus(ctx.Request.Context(), cli, req.Items)

	successCnt, failCnt := 0, 0
	for _, x := range results {
		if x.Success {
			successCnt++
		} else {
			failCnt++
		}
	}
	r.Success(gin.H{
		"total":   len(results),
		"success": successCnt,
		"fail":    failCnt,
		"results": results,
	})
}

/* ====================================================================
 * VPA Controller
 * ==================================================================== */

type KubeVPAController struct{}

func NewKubeVPAController() *KubeVPAController { return &KubeVPAController{} }

// Available godoc
// @Summary 检测 VPA Operator 是否在集群中安装
// @Tags K8s VPA 垂直扩缩容
// @Router /api/v1/k8s/vpa/available [get]
func (c *KubeVPAController) Available(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	available := svc.KubeVPAAvailable(ctx.Request.Context(), cli)
	r.Success(gin.H{"available": available})
}

// List godoc
// @Summary 获取 VPA 列表
// @Tags K8s VPA 垂直扩缩容
// @Router /api/v1/k8s/vpa/list [get]
func (c *KubeVPAController) List(ctx *gin.Context) {
	param := requests.NewKubeVPAListRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeVPAListRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	if !svc.KubeVPAAvailable(ctx.Request.Context(), cli) {
		r.ToErrorResponse(errorcode.ErrorK8sVPANotInstalled)
		return
	}
	list, total, err := svc.KubeVPAList(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("KubeVPAList error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sVPAListFail.WithDetails(err.Error()))
		return
	}
	r.SuccessList(list, total)
}

// Detail godoc
// @Summary 获取 VPA 详情
// @Tags K8s VPA 垂直扩缩容
// @Router /api/v1/k8s/vpa/detail [get]
func (c *KubeVPAController) Detail(ctx *gin.Context) {
	param := requests.NewKubeVPADetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeVPADetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	obj, err := svc.KubeVPADetail(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("KubeVPADetail error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sVPADetailFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{
		"item": vpa.BuildVPAItem(obj),
		"raw":  obj.Object,
	})
}

// Create godoc
// @Summary 创建 VPA
// @Tags K8s VPA 垂直扩缩容
// @Router /api/v1/k8s/vpa/create [post]
func (c *KubeVPAController) Create(ctx *gin.Context) {
	req := requests.NewKubeVPACreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeVPACreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	if !svc.KubeVPAAvailable(ctx.Request.Context(), cli) {
		r.ToErrorResponse(errorcode.ErrorK8sVPANotInstalled)
		return
	}
	obj, err := svc.KubeVPACreate(ctx.Request.Context(), cli, req)
	if err != nil {
		global.Logger.Error("KubeVPACreate error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sVPACreateFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{
		"message": "VPA 创建成功",
		"item":    vpa.BuildVPAItem(obj),
	})
}

// Update godoc
// @Summary 更新 VPA
// @Tags K8s VPA 垂直扩缩容
// @Router /api/v1/k8s/vpa/update [post]
func (c *KubeVPAController) Update(ctx *gin.Context) {
	req := requests.NewKubeVPAUpdateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeVPAUpdateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	obj, err := svc.KubeVPAUpdate(ctx.Request.Context(), cli, req)
	if err != nil {
		global.Logger.Error("KubeVPAUpdate error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sVPAUpdateFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{
		"message": "VPA 更新成功",
		"item":    vpa.BuildVPAItem(obj),
	})
}

// Delete godoc
// @Summary 删除 VPA
// @Tags K8s VPA 垂直扩缩容
// @Router /api/v1/k8s/vpa/delete [delete]
func (c *KubeVPAController) Delete(ctx *gin.Context) {
	param := requests.NewKubeVPADetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeVPADetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	if err := svc.KubeVPADelete(ctx.Request.Context(), cli, param.Namespace, param.Name); err != nil {
		global.Logger.Error("KubeVPADelete error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sVPADeleteFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{"message": "VPA 删除成功", "namespace": param.Namespace, "name": param.Name})
}

// CreateFromYaml godoc
// @Summary 通过 YAML 创建 VPA
// @Tags K8s VPA 垂直扩缩容
// @Accept json
// @Param body body requests.YamlCreateRequest true "YAML 内容"
// @Router /api/v1/k8s/vpa/create-from-yaml [post]
func (c *KubeVPAController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	if cli == nil {
		return
	}
	svc := services.NewServices()
	if !svc.KubeVPAAvailable(ctx.Request.Context(), cli) {
		r.ToErrorResponse(errorcode.ErrorK8sVPANotInstalled)
		return
	}
	obj, err := svc.KubeVPACreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		global.Logger.Error("KubeVPACreateFromYaml error", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sVPACreateFromYamlFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{
		"message":   "VPA 创建成功",
		"name":      obj.GetName(),
		"namespace": obj.GetNamespace(),
		"item":      vpa.BuildVPAItem(obj),
	})
}
