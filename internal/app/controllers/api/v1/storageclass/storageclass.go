package storageclass

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/k8s/storageclass"
	"k8soperation/pkg/valid"
)

type KubeStorageClassController struct{}

func NewKubeStorageClassController() *KubeStorageClassController {
	return &KubeStorageClassController{}
}

// @Summary     创建 StorageClass
// @Description 创建 Kubernetes StorageClass（支持指定 Provisioner、ReclaimPolicy、VolumeBindingMode 等参数）
// @Tags        K8s StorageClass 管理
// @Accept      json
// @Produce     json
// @Param       body  body  requests.KubeStorageClassCreateRequest  true  "StorageClass 创建参数"
// @Success     200   {object} response.Response "成功"
// @Failure     400   {object} errorcode.Error   "请求参数错误"
// @Failure     500   {object} errorcode.Error   "内部错误"
// @Router      /api/v1/k8s/storageclass/create [post]
func (ctl *KubeStorageClassController) Create(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	req := requests.NewKubeStorageClassCreateRequest()

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeStorageClassCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service
	svc := services.NewServices()
	sc, err := svc.KubeCreateStorageClass(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCreateStorageClass error", zap.Error(err))
		return
	}

	// 成功响应
	r.Success(gin.H{
		"name":                sc.Name,
		"provisioner":         sc.Provisioner,
		"reclaim_policy":      sc.ReclaimPolicy,
		"allow_expansion":     sc.AllowVolumeExpansion,
		"volume_binding_mode": sc.VolumeBindingMode,
		"parameters":          sc.Parameters,
		"mount_options":       sc.MountOptions,
		"annotations":         sc.Annotations,
		"created_at":          sc.CreationTimestamp,
	})
}

// List godoc
// @Summary 获取 K8s StorageClass 列表
// @Description 支持分页、名称模糊查询
// @Tags K8s StorageClass 管理
// @Produce json
// @Param name  query string false "StorageClass 名称(模糊匹配)" maxlength(100)
// @Param page  query int    true  "页码 (从1开始)"
// @Param limit query int    true  "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/storageclass/list [get]
func (c *KubeStorageClassController) List(ctx *gin.Context) {
	// 1. 构造请求参数结构体
	param := requests.NewKubeStorageClassListRequest()

	// 2. 创建响应器
	r := response.NewResponse(ctx)

	// 3. 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeStorageClassListRequest); !ok {
		return // 校验失败时，valid 已自动返回错误响应
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 4. 调用 Service 层
	svc := services.NewServices()
	storageClasses, total, err := svc.KubeStorageClassList(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStorageClassList error", zap.Error(err))
		return
	}

	// 5. 转换为前端友好格式
	list := storageclass.BuildStorageClassListResponse(storageClasses)
	r.SuccessList(list, total)
}

// Detail godoc
// @Summary 获取 StorageClass 详情
// @Tags    K8s StorageClass 管理
// @Produce json
// @Param   name query string true "StorageClass 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router  /api/v1/k8s/storageclass/detail [get]
func (c *KubeStorageClassController) Detail(ctx *gin.Context) {
	// 参数
	param := requests.NewKubeStorageClassDetailRequest()

	// 响应器
	r := response.NewResponse(ctx)

	// 校验（内部应当绑定 query：name）
	if ok := valid.Validate(ctx, param, requests.ValidKubeStorageClassDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service
	svc := services.NewServices()
	sc, err := svc.KubeStorageClassDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStorageClassDetail error", zap.Error(err))
		return
	}

	// 成功响应
	r.Success(gin.H{
		"message": "获取 StorageClass 详情成功",
		"data":    sc,
	})
}

// Delete godoc
// @Summary 删除 StorageClass
// @Tags    K8s StorageClass 管理
// @Produce json
// @Param   name query string true "StorageClass 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router  /api/v1/k8s/storageclass/delete [delete]
func (c *KubeStorageClassController) Delete(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeStorageClassDeleteRequest()

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeStorageClassDeleteRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	if err := svc.KubeStorageClassDelete(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Error("service.KubeStorageClassDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	r.Success(gin.H{
		"name":    param.Name,
		"message": "StorageClass 删除成功",
	})
}

// GetYaml godoc
// @Summary 获取 StorageClass YAML
// @Tags K8s StorageClass 管理
// @Produce json
// @Param name query string true "StorageClass 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/storageclass/yaml [get]
func (c *KubeStorageClassController) GetYaml(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeStorageClassDetailRequest()

	if ok := valid.Validate(ctx, param, requests.ValidKubeStorageClassDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	yamlContent, err := svc.KubeStorageClassGetYaml(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("service.KubeStorageClassGetYaml error", zap.Error(err))
		ctx.Error(err)
		return
	}

	r.Success(gin.H{
		"yaml":    yamlContent,
		"message": "获取 YAML 成功",
	})
}

// CreateFromYaml godoc
// @Summary 从 YAML 创建 StorageClass
// @Tags K8s StorageClass 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 内容"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/storageclass/create-from-yaml [post]
func (c *KubeStorageClassController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	sc, err := svc.KubeStorageClassCreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		global.Logger.Error("service.KubeStorageClassCreateFromYaml error", zap.Error(err))
		ctx.Error(err)
		return
	}

	r.Success(storageclass.BuildStorageClassResponse(sc))
}

// ApplyYaml godoc
// @Summary 应用 StorageClass YAML（创建/更新）
// @Tags K8s StorageClass 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 内容"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/storageclass/apply-yaml [post]
func (c *KubeStorageClassController) ApplyYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	sc, err := svc.KubeStorageClassApplyYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		global.Logger.Error("service.KubeStorageClassApplyYaml error", zap.Error(err))
		ctx.Error(err)
		return
	}

	r.Success(gin.H{
		"message": "StorageClass 应用成功",
		"name":    sc.Name,
	})
}
