package secret

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

type KubeSecretController struct{}

func NewKubeSecretController() *KubeSecretController {
	return &KubeSecretController{}
}

// @Summary     创建 Secret
// @Description 创建 Kubernetes Secret（支持 Opaque / TLS / DockerConfigJson / BasicAuth / SSHAuth）
// @Tags        K8s Secret 管理
// @Accept      json
// @Produce     json
// @Param       body  body  requests.KubeSecretCreateRequest  true  "Secret 创建参数"
// @Success     200   {object} response.Response "成功"
// @Failure     400   {object} errorcode.Error   "请求参数错误"
// @Failure     500   {object} errorcode.Error   "内部错误"
// @Router      /api/v1/k8s/secret/create [post]
func (ctl *KubeSecretController) Create(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	req := requests.NewKubeSecretCreateRequest()

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeSecretCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service
	svc := services.NewServices()
	sec, err := svc.KubeCreateSecret(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCreateSecret error", zap.Error(err))
		return
	}

	// 成功响应
	r.Success(gin.H{
		"name":       sec.Name,
		"namespace":  sec.Namespace,
		"type":       sec.Type,
		"data_keys":  lo.Keys(sec.Data), // 返回数据键名（避免输出敏感值）
		"created_at": sec.CreationTimestamp,
	})
}

// List godoc
// @Summary 获取 K8s Secret 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s Secret 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "Secret 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/secret/list [get]
func (c *KubeSecretController) List(ctx *gin.Context) {
	// 构造请求参数结构体
	param := requests.NewKubeSecretListRequest()
	// 创建响应器
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeSecretListRequest); !ok {
		return // 校验失败时，valid 已自动返回错误响应
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service 层
	svc := services.NewServices()
	secrets, total, err := svc.KubeSecretList(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeSecretList error", zap.Error(err))
		return
	}

	// 转换为前端期望的格式
	var list []gin.H
	for _, sec := range secrets {
		list = append(list, gin.H{
			"name":       sec.Name,
			"namespace":  sec.Namespace,
			"type":       string(sec.Type),
			"data_count": len(sec.Data),
			"labels":     sec.Labels,
			"created_at": sec.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	// 返回成功响应
	r.SuccessList(list, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 Secret 列表成功，共 %d 条数据", total),
	})
}

// Detail godoc
// @Summary 获取 Secret 详情
// @Tags K8s Secret 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Secret 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/secret/detail [get]
func (c *KubeSecretController) Detail(ctx *gin.Context) {
	// 构造请求参数
	param := requests.NewKubeSecretDetailRequest()

	// 构造统一响应器
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeSecretDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用业务逻辑层
	svc := services.NewServices()
	secretDetail, err := svc.KubeSecretDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeSecretDetail error", zap.Error(err))
		return
	}

	// 返回成功响应（注意：不直接返回 Data 内容，避免泄漏敏感信息）
	r.Success(gin.H{
		"message":    "获取 Secret 详情成功",
		"name":       secretDetail.Name,
		"namespace":  secretDetail.Namespace,
		"type":       secretDetail.Type,
		"data_keys":  lo.Keys(secretDetail.Data), // 仅返回 key 列表
		"created_at": secretDetail.CreationTimestamp,
	})
}

// Delete godoc
// @Summary 删除 Secret
// @Tags K8s Secret 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Secret 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/secret/delete [delete]
func (c *KubeSecretController) Delete(ctx *gin.Context) {
	param := requests.NewKubeSecretDeleteRequest()
	r := response.NewResponse(ctx)

	// 参数校验（通用 Valid）
	if ok := valid.Validate(ctx, param, requests.ValidKubeSecretDeleteRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	// 调用服务层
	svc := services.NewServices()
	if err := svc.KubeSecretDelete(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Error("service.KubeSecretDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	// 成功响应
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "Secret 删除成功",
	})
}

// @Summary Patch Secret（StrategicMergePatch）
// @Tags K8s Secret 管理
// @Accept application/strategic-merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Secret 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/secret/patch [patch]
func (c *KubeSecretController) Patch(ctx *gin.Context) {
	param := requests.NewKubeSecretUpdateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	out, err := svc.KubeSecretPatch(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeSecretPatch error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": "Secret StrategicMergePatch 成功",
		"data":    out,
	})
}

// @Summary Patch Secret（JSON Merge Patch – 覆盖式）
// @Tags K8s Secret 管理
// @Accept application/merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Secret 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/secret/patch-json [post]
func (c *KubeSecretController) PatchJSON(ctx *gin.Context) {
	param := requests.NewKubeSecretUpdateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	out, err := svc.KubeSecretUpdate(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeSecretPatchJSON error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": "Secret JSON Merge Patch 成功",
		"data":    out,
	})
}

// @Summary     Base64 解码 Secret 数据
// @Description 接收 Base64 字符串或整个 Secret.data 对象，返回明文
// @Tags        K8s Secret 管理
// @Accept      json
// @Produce     json
// @Param       body  body  requests.KubeSecretDecodeRequest  true  "要解码的数据"
// @Success     200   {object} response.Response "成功"
// @Failure     400   {object} errorcode.Error   "请求参数错误"
// @Failure     500   {object} errorcode.Error   "内部错误"
// @Router      /api/v1/k8s/secret/decode [post]
func (c *KubeSecretController) Decode(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	req := requests.NewKubeSecretDecodeRequest()

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeSecretDecodeRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	decoded, err := svc.KubeSecretDecode(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeSecretDecode error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": "Base64 解码成功",
		"data":    decoded,
	})
}

// CreateFromYaml godoc
// @Summary 从 YAML 创建 Secret
// @Tags K8s Secret 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/secret/create-from-yaml [post]
func (c *KubeSecretController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	sec, err := svc.KubeSecretCreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeSecretCreateFromYaml error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"message":   "Secret 创建成功",
		"name":      sec.Name,
		"namespace": sec.Namespace,
	})
}

// Yaml godoc
// @Summary 获取 Secret YAML
// @Tags K8s Secret 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Secret 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/secret/yaml [get]
func (c *KubeSecretController) Yaml(ctx *gin.Context) {
	param := requests.NewKubeSecretDetailRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeSecretDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	yamlStr, err := svc.KubeSecretYaml(ctx.Request.Context(), cli, param.Namespace, param.Name)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeSecretYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"yaml": yamlStr,
	})
}

// ApplyYaml godoc
// @Summary 从 YAML 创建或更新 Secret
// @Tags K8s Secret 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeConfigMapApplyYamlRequest true "YAML 内容"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/secret/apply-yaml [put]
func (c *KubeSecretController) ApplyYaml(ctx *gin.Context) {
	param := requests.NewKubeConfigMapApplyYamlRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeConfigMapApplyYamlRequest); !ok {
		return
	}

	global.Logger.Info("[Secret ApplyYaml] Parsed param",
		zap.String("yaml_length", fmt.Sprintf("%d", len(param.Yaml))),
	)

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	sec, err := svc.KubeSecretApplyYaml(ctx.Request.Context(), cli, param.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeSecretApplyYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "Secret YAML 应用成功",
		"name":      sec.Name,
		"namespace": sec.Namespace,
	})
}
