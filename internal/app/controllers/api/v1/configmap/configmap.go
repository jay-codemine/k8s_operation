package configmap

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type KubeConfigMapController struct {
}

func NewKubeConfigMapController() *KubeConfigMapController {
	return &KubeConfigMapController{}
}

// @Summary     创建 ConfigMap
// @Description 创建 Kubernetes ConfigMap（支持 data 与 binaryData）
// @Tags        K8s ConfigMap 管理
// @Accept      json
// @Produce     json
// @Param       body  body  requests.KubeConfigMapCreateRequest  true  "ConfigMap 创建参数"
// @Success     200   {object} response.Response "成功"
// @Failure     400   {object} errorcode.Error   "请求参数错误"
// @Failure     500   {object} errorcode.Error   "内部错误"
// @Router      /api/v1/k8s/configmap/create [post]
func (ctl *KubeConfigMapController) Create(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	req := requests.NewKubeConfigMapCreateRequest()

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeConfigMapCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service
	svc := services.NewServices()
	cm, err := svc.KubeCreateConfigMap(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCreateConfigMap error", zap.Error(err))
		return
	}

	// 成功响应
	r.Success(gin.H{
		"name":        cm.Name,
		"namespace":   cm.Namespace,
		"data_keys":   lo.Keys(cm.Data),       // 返回 data 键名（避免返回内容）
		"binary_keys": lo.Keys(cm.BinaryData), // 返回 binaryData 键名
		"created_at":  cm.CreationTimestamp,
	})
}

// List godoc
// @Summary 获取 K8s ConfigMap 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s ConfigMap 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "ConfigMap 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/configmap/list [get]
func (c *KubeConfigMapController) List(ctx *gin.Context) {
	// 构造请求参数结构体
	param := requests.NewKubeConfigMapListRequest()

	// 创建响应器
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeConfigMapListRequest); !ok {
		return // 校验失败时，valid 已自动返回错误响应
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service 层
	svc := services.NewServices()
	configMaps, total, err := svc.KubeConfigMapList(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeConfigMapList error", zap.Error(err))
		return
	}

	// 返回成功响应
	r.SuccessList(configMaps, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 ConfigMap 列表成功，共 %d 条数据", total),
	})
}

// Detail godoc
// @Summary 获取 ConfigMap 详情
// @Tags K8s ConfigMap 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "ConfigMap 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/configmap/detail [get]
func (c *KubeConfigMapController) Detail(ctx *gin.Context) {
	// 构造请求参数
	param := requests.NewKubeConfigMapDetailRequest()

	// 构造统一响应器
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeConfigMapDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用业务逻辑层
	svc := services.NewServices()
	cm, err := svc.KubeConfigMapDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeConfigMapDetail error", zap.Error(err))
		return
	}

	// 返回成功响应
	r.Success(gin.H{
		"message": "获取 ConfigMap 详情成功",
		"data":    cm,
	})
}

// Delete godoc
// @Summary 删除 ConfigMap
// @Tags    K8s ConfigMap 管理
// @Produce json
// @Param   namespace query string true "命名空间"
// @Param   name      query string true "ConfigMap 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router  /api/v1/k8s/configmap/delete [delete]
func (c *KubeConfigMapController) Delete(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeConfigMapDeleteRequest()

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeConfigMapDeleteRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用服务层
	svc := services.NewServices()
	if err := svc.KubeConfigMapDelete(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Error("service.KubeConfigMapDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	// 成功响应
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "ConfigMap 删除成功",
	})
}

// @Summary Patch ConfigMap（StrategicMergePatch）
// @Tags K8s ConfigMap 管理
// @Accept application/strategic-merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "ConfigMap 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/configmap/patch [patch]
func (c *KubeConfigMapController) Patch(ctx *gin.Context) {
	param := requests.NewKubeConfigMapUpdateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	out, err := svc.KubeConfigMapPatch(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeConfigMapPatch error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": "ConfigMap StrategicMergePatch 成功",
		"data":    out,
	})
}

// @Summary Patch ConfigMap（JSON Merge Patch – 覆盖式）
// @Tags K8s ConfigMap 管理
// @Accept application/merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "ConfigMap 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/configmap/patch-json [post]
func (c *KubeConfigMapController) PatchJSON(ctx *gin.Context) {
	param := requests.NewKubeConfigMapUpdateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	out, err := svc.KubeConfigMapUpdate(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeConfigMapPatchJSON error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": "ConfigMap JSON Merge Patch 成功",
		"data":    out,
	})
}

// @Summary 获取 ConfigMap YAML
// @Tags K8s ConfigMap 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "ConfigMap 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/configmap/yaml [get]
func (c *KubeConfigMapController) Yaml(ctx *gin.Context) {
	param := requests.NewKubeConfigMapDetailRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeConfigMapDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	yamlStr, err := svc.KubeConfigMapYaml(ctx.Request.Context(), cli, param.Namespace, param.Name)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeConfigMapYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"yaml": yamlStr,
	})
}

// @Summary 更新 ConfigMap data 字段
// @Tags K8s ConfigMap 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeConfigMapUpdateDataRequest true "ConfigMap data 更新参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/configmap/update-data [put]
func (c *KubeConfigMapController) UpdateData(ctx *gin.Context) {
	param := requests.NewKubeConfigMapUpdateDataRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeConfigMapUpdateDataRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	cm, err := svc.KubeConfigMapUpdateData(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeConfigMapUpdateData error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "ConfigMap data 更新成功",
		"name":      cm.Name,
		"namespace": cm.Namespace,
	})
}

// @Summary 从 YAML 创建或更新 ConfigMap
// @Tags K8s ConfigMap 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeConfigMapApplyYamlRequest true "YAML 内容"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/configmap/apply-yaml [post]
func (c *KubeConfigMapController) ApplyYaml(ctx *gin.Context) {
	param := requests.NewKubeConfigMapApplyYamlRequest()
	r := response.NewResponse(ctx)
	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeConfigMapApplyYamlRequest); !ok {
		return
	}

	// 打印解析后的参数
	global.Logger.Info("[ConfigMap ApplyYaml] Parsed param",
		zap.String("yaml_length", fmt.Sprintf("%d", len(param.Yaml))),
		zap.String("yaml_preview", param.Yaml[:min(100, len(param.Yaml))]),
	)

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	cm, err := svc.KubeConfigMapApplyYaml(ctx.Request.Context(), cli, param.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeConfigMapApplyYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "ConfigMap YAML 应用成功",
		"name":      cm.Name,
		"namespace": cm.Namespace,
	})
}

// @Summary 解析多资源 YAML
// @Tags K8s 多资源管理
// @Accept json
// @Produce json
// @Param body body requests.KubeMultiResourceApplyYamlRequest true "多资源 YAML 内容"
// @Success 200 {object} requests.MultiResourceParsedResult "解析结果"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/multi-resource/parse-yaml [post]
func (c *KubeConfigMapController) ParseMultiYaml(ctx *gin.Context) {
	param := requests.NewKubeMultiResourceApplyYamlRequest()
	r := response.NewResponse(ctx)
	
	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeMultiResourceApplyYamlRequest); !ok {
		return
	}

	global.Logger.Info("[MultiResource ParseYaml] Parsed param",
		zap.String("yaml_length", fmt.Sprintf("%d", len(param.Yaml))),
		zap.String("yaml_preview", param.Yaml[:min(100, len(param.Yaml))]),
	)

	svc := services.NewServices()
	result, err := svc.KubeMultiResourceParseYaml(ctx.Request.Context(), param.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeMultiResourceParseYaml error", zap.Error(err))
		return
	}

	r.Success(result)
}

// @Summary 应用多资源 YAML
// @Tags K8s 多资源管理
// @Accept json
// @Produce json
// @Param body body requests.KubeMultiResourceApplyYamlRequest true "多资源 YAML 内容"
// @Success 200 {object} requests.MultiResourceCreateResult "创建结果"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/multi-resource/apply-yaml [post]
func (c *KubeConfigMapController) ApplyMultiYaml(ctx *gin.Context) {
	param := requests.NewKubeMultiResourceApplyYamlRequest()
	r := response.NewResponse(ctx)
	
	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeMultiResourceApplyYamlRequest); !ok {
		return
	}

	global.Logger.Info("[MultiResource ApplyYaml] Parsed param",
		zap.String("yaml_length", fmt.Sprintf("%d", len(param.Yaml))),
		zap.String("yaml_preview", param.Yaml[:min(100, len(param.Yaml))]),
	)

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	result, err := svc.KubeMultiResourceApplyYaml(ctx.Request.Context(), cli, param.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeMultiResourceApplyYaml error", zap.Error(err))
		return
	}

	r.Success(result)
}
