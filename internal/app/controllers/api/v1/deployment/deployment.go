package deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/k8s/deployment"
	"k8soperation/pkg/utils"
	"k8soperation/pkg/valid"
)

type KubeDeploymentController struct{}

func NewKubeDeploymentController() *KubeDeploymentController {
	return &KubeDeploymentController{}
}

// List godoc
// @Summary 获取 K8s Deployment 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s Deployment 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "Deployment 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/deployment/list [get]
// controller/deployment.go
func (c *KubeDeploymentController) List(ctx *gin.Context) {
	param := requests.NewKubeDeploymentListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentListRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	deployments, total, err := svc.KubeDeploymentList(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("获取Deployment列表失败", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorK8sDeploymentListFail.WithDetails(err.Error()))
		return
	}

	// 使用新的转换函数，包含状态信息
	list := deployment.BuildDeploymentListResponse(deployments)
	resp.SuccessList(list, total)
}

// Detail godoc
// @Summary 获取 Deployment 详情
// @Tags K8s Deployment 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/detail [get]
func (c *KubeDeploymentController) Detail(ctx *gin.Context) {
	param := requests.NewKubeDeploymentDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentDetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	dp, err := svc.KubeDeploymentDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDeploymentDetail error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"message": "获取 Deployment 详情成功",
		"data":    dp,
	})
}

// Create godoc
// @Summary 创建 Deployment（可选创建 Service）
// @Tags K8s Deployment 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDeploymentCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/create [post]
func (c *KubeDeploymentController) Create(ctx *gin.Context) {
	req := requests.NewKubeDeploymentCreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeDeploymentCreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	dp, svcObj, err := svc.KubeDeploymentCreate(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDeploymentCreate error", zap.Error(err))
		return
	}
	r.Success(deployment.BuildDeploymentResponse(dp, svcObj, req))
}

// CreateFromYaml godoc
// @Summary 从 YAML 创建 Deployment
// @Tags K8s Deployment 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/create-from-yaml [post]
func (c *KubeDeploymentController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	dp, err := svc.KubeDeploymentCreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDeploymentCreateFromYaml error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"message":   "Deployment 创建成功",
		"name":      dp.Name,
		"namespace": dp.Namespace,
	})
}

// Delete godoc
// @Summary 删除 Deployment
// @Tags K8s Deployment 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/delete [delete]
func (c *KubeDeploymentController) Delete(ctx *gin.Context) {
	param := requests.NewKubeDeploymentDeleteRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentDeleteRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	if err := svc.KubeDeploymentDelete(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Error("service.KubeDeploymentDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}
	// 成功时“加点提示”
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "删除成功",
	})
}

// Scale godoc
// @Summary 扩缩容（修改副本数）
// @Tags K8s Deployment 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDeploymentScaleRequest true "扩缩容参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/scale [put]
func (c *KubeDeploymentController) Scale(ctx *gin.Context) {
	param := requests.NewKubeDeploymentScaleRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentScaleRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	// 注意：传入的是 context.Context
	dep, err := svc.KubeUpdateDeploymentReplicas(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err) // 交给中间件
		return
	}

	r.Success(gin.H{
		"namespace": dep.Namespace,
		"name":      dep.Name,
		"replicas":  utils.ValueOrZero(dep.Spec.Replicas),
		"updated":   dep.Status.UpdatedReplicas,
		"available": dep.Status.AvailableReplicas,
		"rv":        dep.ResourceVersion,
		"status":    fmt.Sprintf("修改副本数成功，当前副本数：%d", utils.ValueOrZero(dep.Spec.Replicas)),
	})
}

// UpdateImage godoc
// @Summary 更新镜像（触发滚动升级）
// @Tags K8s Deployment 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDeploymentUpdateImageRequest true "更新镜像参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/update-image [put]
func (c *KubeDeploymentController) UpdateImage(ctx *gin.Context) {
	param := requests.NewKubeDeploymentUpdateImageRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentUpdateImageRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	dp, err := svc.KubeUpdateDeploymentImage(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeUpdateDeploymentImage error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"namespace": dp.Namespace,
		"name":      dp.Name,
		"message":   "更新镜像成功，触发滚动升级",
	})
}

// Patch godoc
// @Summary Patch 模板（JSONPatch / StrategicMergePatch）
// @Tags K8s Deployment 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDeploymentUpdateRequest true "Patch 内容"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/patch [put]
func (c *KubeDeploymentController) PatchTemplate(ctx *gin.Context) {
	param := requests.NewKubeDeploymentUpdateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentUpdateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	dp, err := svc.KubeUpdateDeploymentTemplate(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeUpdateDeploymentTemplate error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"namespace": dp.Namespace,
		"name":      dp.Name,
		"message":   "更新成功",
	})
}

// Rollback godoc
// @Summary 回滚到指定 ReplicaSet
// @Tags K8s Deployment 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDeploymentRollbackRequest true "回滚参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/rollback [post]
func (c *KubeDeploymentController) Rollback(ctx *gin.Context) {
	param := requests.NewKubeDeploymentRollbackRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentRollbackRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	_, err := svc.KubeDeploymentRollback(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDeploymentRollback error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": "回滚成功",
	})
}

// Restart godoc
// @Summary 滚动重启 Deployment
// @Tags K8s Deployment 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeDeploymentRestartRequest true "重启参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/restart [post]
func (c *KubeDeploymentController) Restart(ctx *gin.Context) {
	param := requests.NewKubeDeploymentRestartRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentRestartRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	if err := svc.KubeDeploymentRestart(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDeploymentRestart error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"message": "滚动重启成功",
	})
}

// Pods godoc
// @Summary 获取 Deployment 对应的 Pod 列表
// @Tags K8s Deployment 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/pods.js [get]
func (c *KubeDeploymentController) DeploymentPodList(ctx *gin.Context) {
	param := requests.NewKubeCommonRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.VaildKubeCommonRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	pods, err := svc.KubeDeploymentGetPod(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDeploymentGetPod error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"pods":    pods,
		"message": "获取 Pod 列表成功",
	})
}

// DeleteService godoc
// @Summary 删除 Deployment 对应的 Service
// @Description 删除指定命名空间下，与 Deployment 同名的 Service 资源
// @Tags K8s Deployment 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment 名称"
// @Success 200 {object} response.Response "Service 删除成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "服务器内部错误"
// @Router /api/v1/k8s/deployment/delete_service [delete]
func (c *KubeDeploymentController) DeleteService(ctx *gin.Context) {
	param := requests.NewKubeDeploymentDeleteRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentDeleteRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	if err := svc.KubeDeploymentDeleteService(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDeploymentDeleteService error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "Service 删除成功",
	})
}

// EventList godoc
// @Summary 获取事件列表（支持全局或按对象筛选）
// @Description 按命名空间、资源类型/名称、事件类型(Type)、原因(Reason)等过滤；支持最近N秒窗口与分页（continue游标）。
// @Tags K8s Event 管理
// @Produce json
// @Param namespace     query string false "命名空间（为空=全局）"
// @Param kind          query string false "资源类型（如 Pod/Deployment/Node）"
// @Param name          query string false "资源名称"
// @Param type          query string false "事件类型（Normal | Warning）"
// @Param reason        query string false "事件原因（如 FailedScheduling/BackOff）"
// @Param limit         query int    false "返回条数限制（默认50，最大500）"
// @Param continue      query string false "K8s 分页游标（下一页传回上次返回的 next）"
// @Param since_seconds query int    false "最近N秒的事件（客户端二次过滤）"
// @Success 200 {object} response.Response "事件列表加载完成"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "服务器内部错误"
// @Router /api/v1/k8s/deployment/events [post]
func (c *KubeDeploymentController) EventList(ctx *gin.Context) {
	param := requests.NewKubeEventListRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeEventListRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	items, next, err := svc.KubeEventList(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDeploymentGetEvent error", zap.Error(err))
	}
	r.Success(gin.H{
		"events":  items,         // 事件记录
		"next":    next,          //下一页的起始时间
		"message": "已获取到最新的事件记录", // 返回信息
	})
}

// History godoc
// @Summary 获取 Deployment 历史版本（ReplicaSet 列表）
// @Tags K8s Deployment 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/deployment/history [get]
func (c *KubeDeploymentController) History(ctx *gin.Context) {
	param := requests.NewKubeCommonRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.VaildKubeCommonRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	rsList, err := svc.KubeDeploymentHistory(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeDeploymentHistory error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"list":    rsList,
		"message": "获取历史版本成功",
	})
}

// Yaml godoc
// @Summary 获取 Deployment 的 YAML 配置
// @Description 获取指定 Deployment 的 YAML 格式配置
// @Tags K8s Deployment 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Deployment 名称"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/deployment/yaml [get]
func (c *KubeDeploymentController) Yaml(ctx *gin.Context) {
	param := requests.NewKubeDeploymentDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeDeploymentDetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	yamlStr, err := deployment.GetYaml(ctx.Request.Context(), cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Error("获取 Deployment YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sDeploymentListFail.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"yaml":      yamlStr,
	})
}

// ApplyYaml godoc
// @Summary 应用 Deployment YAML 配置
// @Description 应用修改后的 YAML 配置到 Deployment
// @Tags K8s Deployment 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeApplyYamlRequest true "YAML内容"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/deployment/apply_yaml [put]
func (c *KubeDeploymentController) ApplyYaml(ctx *gin.Context) {
	param := requests.NewKubeApplyYamlRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeApplyYamlRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	_, err := deployment.ApplyYaml(ctx.Request.Context(), cli.Kube, param.Namespace, param.Name, param.Yaml)
	if err != nil {
		global.Logger.Error("应用 Deployment YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorK8sDeploymentListFail.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "YAML 应用成功",
	})
}
