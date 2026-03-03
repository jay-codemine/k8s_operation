package daemonset

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	dspkg "k8soperation/pkg/k8s/daemonset"
	"k8soperation/pkg/valid"
)

type KubeDaemonSetController struct{}

func NewKubeDaemonSetController() *KubeDaemonSetController {
	return &KubeDaemonSetController{}
}

// Create godoc
// @Summary 创建 DaemonSet（可选创建 Service）
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDaemonSetCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/create [post]
func (c *KubeDaemonSetController) Create(ctx *gin.Context) {
	// 创建请求体结构
	req := requests.NewKubeDaemonSetCreateRequest()

	// 响应封装器
	r := response.NewResponse(ctx)

	// 参数校验（通用 valid.Validate）
	if ok := valid.Validate(ctx, req, requests.ValidKubeDaemonSetCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service 层
	svc := services.NewServices()
	ds, svcObj, err := svc.KubeDaemonSetCreate(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetCreate error", zap.Error(err))
		return
	}

	// 成功响应（使用 daemonset 包的构建函数）
	r.Success(dspkg.BuildDaemonSetResponse(ds, svcObj, req))
}

// CreateFromYaml godoc
// @Summary 从 YAML 创建 DaemonSet
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/create-from-yaml [post]
func (c *KubeDaemonSetController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	ds, err := svc.KubeDaemonSetCreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetCreateFromYaml error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"message":   "DaemonSet 创建成功",
		"name":      ds.Name,
		"namespace": ds.Namespace,
	})
}

// List godoc
// @Summary 获取 K8s DaemonSet 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "DaemonSet 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/daemonset/list [get]
func (c *KubeDaemonSetController) List(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetListRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	daemonsets, total, err := svc.KubeDaemonSetList(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("获取DaemonSet列表失败", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorDaemonSetQueryFail.WithDetails(err.Error()))
		return
	}

	// 转换为前端友好格式
	list := dspkg.BuildDaemonSetListResponse(daemonsets)
	resp.SuccessList(list, total)
}

// Detail godoc
// @Summary 获取 DaemonSet 详情
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/detail [get]
func (c *KubeDaemonSetController) Detail(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetDetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	ds, err := svc.KubeDaemonSetDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetDetail error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"message": "获取 DaemonSet 详情成功",
		"data":    ds,
	})
}

// Delete godoc
// @Summary 删除 DaemonSet
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/delete [delete]
func (c *KubeDaemonSetController) Delete(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetDeleteRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetDeleteRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	if err := svc.KubeDaemonSetDelete(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Error("service.KubeDaemonSetDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	// 成功返回
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "DaemonSet 删除成功",
	})
}

// DeleteService godoc
// @Summary 删除 DaemonSet 对应的 Service
// @Description 删除指定命名空间下，与 DaemonSet 同名的 Service 资源
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} response.Response "Service 删除成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "服务器内部错误"
// @Router /api/v1/k8s/daemonset/delete_service [delete]
func (c *KubeDaemonSetController) DeleteService(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetDeleteRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetDeleteRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	if err := svc.KubeDaemonSetDeleteService(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Error("service.KubeDaemonSetDeleteService error", zap.Error(err))
		ctx.Error(err)
		return
	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "DaemonSet Service 删除成功",
	})
}

// UpdateImage godoc
// @Summary 更新 DaemonSet 容器镜像
// @Description 修改指定命名空间下 DaemonSet 的容器镜像（支持滚动更新）
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDaemonSetUpdateImageRequest true "更新镜像参数"
// @Success 200 {object} string "更新成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "服务器内部错误"
// @Router /api/v1/k8s/daemonset/update_image [put]
func (c *KubeDaemonSetController) UpdateImage(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetUpdateImageRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetUpdateImageRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	ds, err := svc.KubeDaemonSetUpdateImage(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetUpdateImage error", zap.Error(err))
		return

	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "更新 DaemonSet 镜像成功",
		"data":      ds,
	})
}

// Restart godoc
// @Summary 重启 DaemonSet
// @Description 触发指定命名空间下 DaemonSet 的滚动重启（等价于 kubectl rollout restart ds <name>）
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} string "DaemonSet 重启成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/restart [post]
func (c *KubeDaemonSetController) Restart(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetRestartRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetRestartRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	err := svc.KubeDaemonSetRestart(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetRestart error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "DaemonSet 重启成功",
	})
}

// Rollback godoc
// @Summary 回滚 DaemonSet
// @Description 将 DaemonSet 回滚到指定的历史版本（ControllerRevision）。不传可在服务端实现“回滚到上一个版本”的兜底策略（可选）。
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param request body requests.KubeDaemonSetRollbackRequest true "回滚参数（namespace、name、revision_name）"
// @Success 200 {object} string "DaemonSet 回滚成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/rollback [post]
func (c *KubeDaemonSetController) Rollback(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetRollbackRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetRollbackRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	_, err := svc.KubeDaemonSetRollback(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetRollback error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "DaemonSet 回滚成功",
	})
}

// Pods godoc
// @Summary 获取 DaemonSet 关联的 Pod 列表
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/ds_pods [get]
func (c *KubeDaemonSetController) Pods(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetPodsRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetPodsRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	pods, err := svc.KubeDaemonSetPods(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetPods error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"pods": pods,
	})
}

// History godoc
// @Summary 获取 DaemonSet 历史版本
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/history [get]
func (c *KubeDaemonSetController) History(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetHistoryRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetHistoryRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	history, err := svc.KubeDaemonSetHistory(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDaemonSetHistory error", zap.Error(err))
		return
	}
	r.Success(history)
}

// Events godoc
// @Summary 获取 DaemonSet 相关事件
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeEventListRequest true "事件查询参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/daemonset/events [post]
func (c *KubeDaemonSetController) Events(ctx *gin.Context) {
	param := requests.NewKubeEventListRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeEventListRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	events, _, err := svc.KubeEventList(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeEventList error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"events": events,
	})
}

// Yaml godoc
// @Summary 获取 DaemonSet YAML 配置
// @Description 获取指定 DaemonSet 的 YAML 配置
// @Tags K8s DaemonSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "DaemonSet 名称"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/daemonset/yaml [get]
func (c *KubeDaemonSetController) Yaml(ctx *gin.Context) {
	param := requests.NewKubeDaemonSetDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDaemonSetDetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	yamlStr, err := dspkg.GetYaml(ctx.Request.Context(), cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Error("获取 DaemonSet YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorDaemonSetQueryFail.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"yaml":      yamlStr,
	})
}

// ApplyYaml godoc
// @Summary 应用 DaemonSet YAML 配置
// @Description 应用修改后的 YAML 配置到 DaemonSet
// @Tags K8s DaemonSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeApplyYamlRequest true "YAML内容"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/daemonset/apply_yaml [put]
func (c *KubeDaemonSetController) ApplyYaml(ctx *gin.Context) {
	param := requests.NewKubeApplyYamlRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeApplyYamlRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	_, err := dspkg.ApplyYaml(ctx.Request.Context(), cli.Kube, param.Namespace, param.Name, param.Yaml)
	if err != nil {
		global.Logger.Error("应用 DaemonSet YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorDaemonSetQueryFail.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "YAML 应用成功",
	})
}
