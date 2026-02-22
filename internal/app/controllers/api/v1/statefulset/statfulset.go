package statefulset

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
	stsbuilder "k8soperation/pkg/k8s/statefulset"
	"k8soperation/pkg/utils"
	"k8soperation/pkg/valid"
	"time"
)

type KubeStatefulSetController struct{}

func NewKubeStatefulSetController() *KubeStatefulSetController {
	return &KubeStatefulSetController{}
}

// Create godoc
// @Summary 创建 StatefulSet（可选创建 Service）
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeStatefulSetCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/create [post]
func (c *KubeStatefulSetController) Create(ctx *gin.Context) {
	req := requests.NewKubeStatefulSetCreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidKubeStatefulSetCreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	service := services.NewServices()
	sts, svc, err := service.KubeStatefulSetCreateService(ctx.Request.Context(), cli, req)
	// fulSetCreateService(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetCreate error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": "创建成功",
		"service": svc,
		"result":  sts,
	})
}

// CreateFromYaml godoc
// @Summary 从 YAML 创建 StatefulSet
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/create-from-yaml [post]
func (c *KubeStatefulSetController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	service := services.NewServices()
	sts, err := service.KubeStatefulSetCreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetCreateFromYaml error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"message":   "StatefulSet 创建成功",
		"name":      sts.Name,
		"namespace": sts.Namespace,
	})
}

// List godoc
// @Summary 获取 StatefulSet 列表
// @Description 分页、模糊查询 StatefulSet 列表
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param name query string false "StatefulSet 名称关键字（模糊匹配）"
// @Param page query int false "页码（默认 1）"
// @Param limit query int false "每页数量（默认 10）"
// @Success 200 {object} string "查询成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/list [get]
func (c *KubeStatefulSetController) List(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetListRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	sts, total, err := svc.KubeStatefulSetList(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("获取StatefulSet列表失败", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorStatefulSetQueryFail.WithDetails(err.Error()))
		return
	}

	// 使用转换函数，包含状态信息
	list := stsbuilder.BuildStatefulSetListResponse(sts)
	resp.SuccessList(list, total)
}

// Detail godoc
// @Summary 获取 StatefulSet 详情
// @Description 根据命名空间和名称获取单个 StatefulSet 的详细信息
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "StatefulSet 名称"
// @Success 200 {object} string "获取详情成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/detail [get]
func (c *KubeStatefulSetController) Detail(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetDetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	service := services.NewServices()
	sts, err := service.KubeStatefulSetDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetDetail error", zap.Error(err))
		return // ← 修复：错误时必须 return
	}

	r.Success(gin.H{
		"message": "获取详情成功",
		"result":  sts,
	})
}

// Scale godoc
// @Summary 扩缩容 StatefulSet（修改副本数）
// @Description 通过 Patch 局部更新 .spec.replicas，K8s 将按策略有序创建/删除 Pod
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeStatefulSetScaleRequest true "扩缩容参数（namespace、name、scale_num）"
// @Success 200 {object} map[string]interface{} "扩缩容成功，返回修改前后及当前副本信息"
// @Failure 400 {object} errorcode.Error
// @Failure 404 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/scale [put]
func (c *KubeStatefulSetController) Scale(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetScaleRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetScaleRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	service := services.NewServices()
	sts, err := service.KubeStatefulSetPatchReplicas(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err) // 交给中间件
		return
	}

	// 扩缩容成功后的返回（sts 为 *appv1.StatefulSet）
	r.Success(gin.H{
		"namespace": sts.Namespace,
		"name":      sts.Name,
		"replicas":  utils.ValueOrZero(sts.Spec.Replicas),
		"ready":     sts.Status.ReadyReplicas,
		"updated":   sts.Status.UpdatedReplicas,
		"rv":        sts.ResourceVersion,
		"status": fmt.Sprintf(
			"扩缩容成功，目标副本数：%d，当前就绪：%d/%d",
			utils.ValueOrZero(sts.Spec.Replicas),
			sts.Status.ReadyReplicas,
			utils.ValueOrZero(sts.Spec.Replicas),
		),
	})

}

// UpdateImage godoc
// @Summary 更新 StatefulSet 容器镜像（Patch 局部更新）
// @Description 仅修改 .spec.template.spec.containers[*].image，不影响其它字段；根据 UpdateStrategy 触发滚动更新
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeStatefulSetUpdateImageRequest true "更新镜像参数（namespace、name、container、image）"
// @Success 200 {object} map[string]interface{} "更新成功，返回资源版本与副本进度"
// @Failure 400 {object} errorcode.Error
// @Failure 404 {object} errorcode.Error
// @Failure 409 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/update_image [put]
func (c *KubeStatefulSetController) UpdateImage(ctx *gin.Context) {
	// 1) 绑定请求体
	param := new(requests.KubeStatefulSetUpdateImageRequest)
	r := response.NewResponse(ctx)

	// 2) 参数校验（StatefulSet 的校验器）
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetUpdateImageRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 3) 调用服务
	svc := services.NewServices()
	sts, err := svc.KubeStatefulSetPatchImage(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err) // 交给中间件统一处理
		return
	}

	// 4) 成功返回（回显关键信息）
	r.Success(gin.H{
		"namespace": sts.Namespace,
		"name":      sts.Name,
		"container": param.Container,
		"image":     param.Image,
		"ready":     sts.Status.ReadyReplicas,             // 就绪副本
		"replicas":  utils.ValueOrZero(sts.Spec.Replicas), // 期望副本
		"rv":        sts.ResourceVersion,                  // 资源版本
		"status": fmt.Sprintf("修改镜像成功，当前就绪：%d/%d",
			sts.Status.ReadyReplicas, utils.ValueOrZero(sts.Spec.Replicas)),
	})
}

// Restart godoc
// @Summary 重启 StatefulSet（触发滚动重启）
// @Description 在 .spec.template.metadata.annotations 写入 `kubectl.kubernetes.io/restartedAt` 时间戳，从而触发 StatefulSet 的滚动更新；等价于 `kubectl rollout restart sts <name>`。
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeStatefulSetRestartRequest true "重启参数（namespace、name）"
// @Success 200 {object} map[string]interface{} "重启成功，返回命名空间、名称与触发时间等信息"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 404 {object} errorcode.Error "StatefulSet 未找到"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/statefulset/restart [post]
func (c *KubeStatefulSetController) Restart(ctx *gin.Context) {
	ts := time.Now().Format(time.RFC3339)

	param := requests.NewKubeStatefulSetRestartRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetRestartRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	service := services.NewServices()
	sts, err := service.KubeStatefulSetRestart(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetRestart error", zap.Error(err))
		return // ← 修复：错误时必须 return
	}
	r.Success(gin.H{
		"namespace":   sts.Namespace,
		"name":        sts.Name,
		"restartedAt": ts,
		"status":      "StatefulSet 滚动重启已触发",
	})
}

// Delete godoc
// @Summary     删除 StatefulSet
// @Description 前台级联删除 StatefulSet（先删 Pod/ControllerRevision，再删 StatefulSet 本体）；成功返回命名空间、名称与状态信息
// @Tags        K8s StatefulSet 管理
// @Accept      json
// @Produce     json
// @Param       namespace query string true  "命名空间"
// @Param       name      query string true  "StatefulSet 名称"
// @Success     200 {object} map[string]interface{} "示例: {\"namespace\":\"default\",\"name\":\"web\",\"status\":\"StatefulSet 删除成功\"}"
// @Failure     400 {object} errorcode.Error "参数错误"
// @Failure     404 {object} errorcode.Error "StatefulSet 未找到"
// @Failure     500 {object} errorcode.Error "内部错误"
// @Router      /api/v1/k8s/statefulset/delete [delete]
// @Security    BearerAuth
func (c *KubeStatefulSetController) Delete(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetDeleteRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetDeleteRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	service := services.NewServices()
	err := service.KubeStatefulSetDelete(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetDelete error", zap.Error(err))
		return // ← 修复：错误时必须 return
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"status":    "StatefulSet 删除成功",
	})
}

// Delete godoc
// @Summary     删除 Service
// @Description 根据命名空间与名称删除 Service；成功仅表示删除请求已受理。
// @Tags        K8s StatefulSet 管理
// @Accept      json
// @Produce     json
// @Param       body  body  map[string]string  true  "删除参数（namespace、name）"
// @Success     200 {object} map[string]interface{} "删除成功"
// @Failure     400 {object} errorcode.Error "参数错误"
// @Failure     404 {object} errorcode.Error "Service 未找到"
// @Failure     500 {object} errorcode.Error "内部错误"
// @Router      /api/v1/k8s/statefulset/delete_svc [delete]
func (c *KubeStatefulSetController) DeleteService(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetDeleteRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetDeleteRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	service := services.NewServices()
	err := service.KubeStatefulSetDeleteService(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetDeleteService error", zap.Error(err))
		return // ← 修复：错误时必须 return
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"status":    "Service 删除成功",
	})
}

// PodList godoc
// @Summary 获取 StatefulSet 对应的 Pod 列表
// @Tags K8s StatefulSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "StatefulSet 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/sts_pods [get]
func (c *KubeStatefulSetController) PodList(ctx *gin.Context) {
	param := requests.NewKubeCommonRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.VaildKubeCommonRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	pods, err := svc.KubeStatefulSetGetPod(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetGetPod error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"pods":    pods,
		"message": "获取 Pod 列表成功",
	})
}

// EventList godoc
// @Summary 获取事件列表（支持全局或按对象筛选）
// @Tags K8s StatefulSet 管理
// @Produce json
// @Param namespace query string false "命名空间"
// @Param kind query string false "资源类型"
// @Param name query string false "资源名称"
// @Success 200 {object} response.Response "事件列表加载完成"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/events [post]
func (c *KubeStatefulSetController) EventList(ctx *gin.Context) {
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
		global.Logger.Error("service.KubeStatefulSetGetEvent error", zap.Error(err))
		return // ← 修复：错误时必须 return
	}
	r.Success(gin.H{
		"events":  items,
		"next":    next,
		"message": "已获取到最新的事件记录",
	})
}

// History godoc
// @Summary 获取 StatefulSet 历史版本（ControllerRevision 列表）
// @Tags K8s StatefulSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "StatefulSet 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/history [get]
func (c *KubeStatefulSetController) History(ctx *gin.Context) {
	param := requests.NewKubeCommonRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.VaildKubeCommonRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	revList, err := svc.KubeStatefulSetHistory(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetHistory error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"list":    revList,
		"message": "获取历史版本成功",
	})
}

// Rollback godoc
// @Summary 回滚 StatefulSet
// @Description 将 StatefulSet 回滚到指定的历史版本（ControllerRevision）
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param request body requests.KubeStatefulSetRollbackRequest true "回滚参数（namespace、name、revision_name）"
// @Success 200 {object} string "StatefulSet 回滚成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/statefulset/rollback [post]
func (c *KubeStatefulSetController) Rollback(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetRollbackRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetRollbackRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	_, err := svc.KubeStatefulSetRollback(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeStatefulSetRollback error", zap.Error(err))
		return
	}
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "StatefulSet 回滚成功",
	})
}

// Yaml godoc
// @Summary 获取 StatefulSet 的 YAML 配置
// @Description 获取指定 StatefulSet 的 YAML 格式配置
// @Tags K8s StatefulSet 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "StatefulSet 名称"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/statefulset/yaml [get]
func (c *KubeStatefulSetController) Yaml(ctx *gin.Context) {
	param := requests.NewKubeStatefulSetDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeStatefulSetDetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	yamlStr, err := stsbuilder.GetYaml(ctx.Request.Context(), cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Error("获取 StatefulSet YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorStatefulSetQueryFail.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"yaml":      yamlStr,
	})
}

// ApplyYaml godoc
// @Summary 应用 StatefulSet YAML 配置
// @Description 应用修改后的 YAML 配置到 StatefulSet
// @Tags K8s StatefulSet 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeApplyYamlRequest true "YAML内容"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/statefulset/apply_yaml [put]
func (c *KubeStatefulSetController) ApplyYaml(ctx *gin.Context) {
	param := requests.NewKubeApplyYamlRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeApplyYamlRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	_, err := stsbuilder.ApplyYaml(ctx.Request.Context(), cli.Kube, param.Namespace, param.Name, param.Yaml)
	if err != nil {
		global.Logger.Error("应用 StatefulSet YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ErrorStatefulSetQueryFail.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "YAML 应用成功",
	})
}
