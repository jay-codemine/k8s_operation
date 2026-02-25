package node

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	nodepkg "k8soperation/pkg/k8s/node"
	"k8soperation/pkg/valid"
)

type KubeNodeController struct {
}

func NewKubeNodeController() *KubeNodeController {
	return &KubeNodeController{}
}

// @Summary 获取 Node 列表
// @Description 支持分页、名称模糊查询（Node 为集群级资源，无 namespace 参数）
// @Tags K8s Node 管理
// @Produce json
// @Param name  query string false "Node 名称(模糊匹配)" maxlength(100)
// @Param page  query int    true  "页码(从1开始)"
// @Param limit query int    true  "每页数量(默认20)"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/node/list [get]
func (ctl *KubeNodeController) List(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeNodeListRequest()

	// 若 valid.Validate 内部已完成 ShouldBindQuery，这里无需再绑定
	// 否则可手动绑定：_ = ctx.ShouldBindQuery(param)
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeListRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	items, total, err := svc.KubeNodeList(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNodeList error", zap.Error(err))
		return
	}

	// 格式化返回数据，计算 Pod 数量
	type NodeItem struct {
		Name              string            `json:"name"`
		Status            string            `json:"status"`
		IP                string            `json:"ip"`
		Role              string            `json:"role"`
		Version           string            `json:"version"`
		PodCount          int               `json:"pod_count"`
		Labels            map[string]string `json:"labels"`
		Unschedulable     bool              `json:"unschedulable"`
		CreationTimestamp interface{}       `json:"creation_timestamp"`
		// 资源容量
		CPUCapacity    string `json:"cpu_capacity"`
		MemoryCapacity string `json:"memory_capacity"`
	}

	result := make([]NodeItem, 0, len(items))
	for _, node := range items {
		// 解析状态
		status := "Unknown"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				if cond.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		// 获取内部 IP
		ip := "-"
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				ip = addr.Address
				break
			}
		}

		// 获取角色
		role := "Worker"
		if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			role = "Control Plane"
		} else if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			role = "Control Plane"
		}

		// 统计该节点上的 Pod 数量
		podCount := 0
		podList, err := cli.Kube.CoreV1().Pods("").List(ctx.Request.Context(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", node.Name),
		})
		if err == nil {
			podCount = len(podList.Items)
		}

		result = append(result, NodeItem{
			Name:              node.Name,
			Status:            status,
			IP:                ip,
			Role:              role,
			Version:           node.Status.NodeInfo.KubeletVersion,
			PodCount:          podCount,
			Labels:            node.Labels,
			Unschedulable:     node.Spec.Unschedulable,
			CreationTimestamp: node.CreationTimestamp,
			CPUCapacity:       node.Status.Capacity.Cpu().String(),
			MemoryCapacity:    node.Status.Capacity.Memory().String(),
		})
	}

	r.SuccessList(result, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 Node 列表成功，共 %d 条", total),
	})
}

// Detail godoc
// @Summary 获取 Node 详情
// @Description 查询指定 Node 的详情（Node 为集群级资源，无 namespace 参数）
// @Tags K8s Node 管理
// @Produce json
// @Param name query string true "Node 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/detail [get]
func (c *KubeNodeController) Detail(ctx *gin.Context) {
	// 1) 构造参数
	param := requests.NewKubeNodeDetailRequest()

	// 2) 响应封装
	r := response.NewResponse(ctx)

	// 3) 参数校验（若 valid.Validate 不含绑定，请加上：_ = ctx.ShouldBindQuery(param)）
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 4) 调用 Service
	svc := services.NewServices()
	nodeObj, err := svc.KubeNodeDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNodeDetail error", zap.Error(err))
		return
	}

	// 5) 整理返回字段
	ready := "Unknown"
	for _, cond := range nodeObj.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				ready = "True"
			} else {
				ready = "False"
			}
			break
		}
	}

	r.Success(gin.H{
		"message":       fmt.Sprintf("获取 Node %s 详情成功", nodeObj.Name),
		"name":          nodeObj.Name,
		"labels":        nodeObj.Labels,
		"taints":        nodeObj.Spec.Taints,
		"unschedulable": nodeObj.Spec.Unschedulable,
		"capacity":      nodeObj.Status.Capacity,
		"allocatable":   nodeObj.Status.Allocatable,
		"addresses":     nodeObj.Status.Addresses,
		"ready":         ready,
		"created_at":    nodeObj.CreationTimestamp,
		// 新增：节点条件
		"conditions": nodepkg.GetNodeConditions(nodeObj),
		// 新增：系统信息
		"system_info": nodepkg.GetNodeSystemInfo(nodeObj),
	})
}

// @Summary 获取指定 Node 上的 Pod 列表
// @Description 支持分页、名称模糊查询（Pod 属于命名空间级资源，但此处跨命名空间按 Node 过滤）
// @Tags K8s Node 管理
// @Produce json
// @Param nodeName query string true  "Node 名称"
// @Param name     query string false "Pod 名称(模糊匹配)" maxlength(100)
// @Param page     query int    true  "页码(从1开始)"
// @Param limit    query int    true  "每页数量(默认20)"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/node/pods.js [get]
func (ctl *KubeNodeController) ListPods(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeNodePodsRequest()

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodePodsRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service 层
	svc := services.NewServices()
	items, err := svc.KubeNodePods(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNodePods error", zap.Error(err))
		return
	}

	// 成功响应
	r.SuccessList(items, gin.H{
		"total":   len(items),
		"message": fmt.Sprintf("获取 Node[%s] 上的 Pod 列表成功，共 %d 条", param.Name, len(items)),
	})
}

// @Summary 获取 Node 指标（CPU/内存使用率）
// @Description name 为空则返回全量节点指标；填写则返回指定节点。
// @Tags K8s Node 管理
// @Produce json
// @Param name query string false "Node 名称（为空=全局）"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/node/metrics [get]
func (ctl *KubeNodeController) Metrics(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := &requests.KubeNodeMetricsRequest{}

	// 参数绑定与校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeMetricsRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 service
	svc := services.NewServices()
	items, total, err := svc.KubeNodeMetricsList(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Errorf("获取 Node 指标失败")
		global.Logger.Error("service.KubeNodeMetricsList error", zap.Error(err))
		return
	}

	// 响应信息
	msg := "获取全量 Node 指标成功"
	if param.Name != "" {
		msg = fmt.Sprintf("获取 Node[%s] 指标成功", param.Name)
	}

	r.SuccessList(items, gin.H{
		"total":   total,
		"message": msg,
	})
}

// Cordon godoc
// @Summary 标记 Node 是否可调度（cordon / uncordon）
// @Description 通过设置 spec.unschedulable 实现 cordon / uncordon
// @Tags K8s Node 管理
// @Accept json
// @Produce json
// @Param data body requests.KubeNodeCordonRequest true "Node 调度控制参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/cordon [post]
func (c *KubeNodeController) Cordon(ctx *gin.Context) {
	// 1) 构造参数
	param := requests.NewKubeNodeCordonRequest()

	// 2) 响应封装
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeCordonRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 4) 调用 Service
	svc := services.NewServices()
	if err := svc.KubeNodeCordon(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNodeCordon error", zap.Error(err))
		return
	}

	// 5) 成功返回
	r.Success(gin.H{
		"message":       fmt.Sprintf("设置 Node %s unschedulable=%v 成功", param.NodeName, param.Unschedulable),
		"nodeName":      param.NodeName,
		"unschedulable": param.Unschedulable,
	})
}

// Drain godoc
// @Summary 驱逐节点上的可驱逐 Pod（drain）
// @Description cordon 节点并驱逐其上非 DaemonSet/非静态 Pod（维护/下线常用）
// @Tags K8s Node 管理
// @Accept json
// @Produce json
// @Param data body requests.KubeNodeDrainRequest true "Node 驱逐参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/drain [post]
func (c *KubeNodeController) Drain(ctx *gin.Context) {
	// 1) 构造参数
	param := requests.NewKubeNodeDrainRequest()

	// 2) 响应封装
	r := response.NewResponse(ctx)

	// 3) 绑定 + 校验
	_ = ctx.ShouldBindJSON(param)
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeDrainRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 4) 调用 Service
	svc := services.NewServices()
	if err := svc.KubeNodeDrain(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNodeDrain error", zap.Error(err))
		return
	}

	// 5) 成功返回
	r.Success(gin.H{
		"message":  fmt.Sprintf("节点 %s drain 成功（已 cordon 并驱逐可驱逐 Pod）", param.NodeName),
		"nodeName": param.NodeName,
	})
}

// PatchLabels godoc
// @Summary 修改节点标签
// @Description 添加或删除节点标签
// @Tags K8s Node 管理
// @Accept json
// @Produce json
// @Param data body requests.KubeNodeLabelPatchRequest true "标签修改参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/labels [patch]
func (c *KubeNodeController) PatchLabels(ctx *gin.Context) {
	param := requests.NewKubeNodeLabelPatchRequest()
	r := response.NewResponse(ctx)

	// valid.Validate 内部会处理 JSON 绑定，不需要在这里再调用 ShouldBindJSON
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeLabelPatchRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()

	if err := svc.KubeNodePatchLabels(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNodePatchLabels error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("节点 %s 标签修改成功", param.Name),
		"name":    param.Name,
	})
}

// PatchTaints godoc
// @Summary 修改节点污点
// @Description 添加或删除节点污点
// @Tags K8s Node 管理
// @Accept json
// @Produce json
// @Param data body requests.KubeNodeTaintPatchRequest true "污点修改参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/taints [patch]
func (c *KubeNodeController) PatchTaints(ctx *gin.Context) {
	param := requests.NewKubeNodeTaintPatchRequest()
	r := response.NewResponse(ctx)

	// valid.Validate 内部会处理 JSON 绑定，不需要在这里再调用 ShouldBindJSON
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeTaintPatchRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()

	if err := svc.KubeNodePatchTaints(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNodePatchTaints error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("节点 %s 污点修改成功", param.Name),
		"name":    param.Name,
	})
}

// Events godoc
// @Summary 获取节点事件
// @Description 查询与节点相关的 Events
// @Tags K8s Node 管理
// @Produce json
// @Param name query string true "Node 名称"
// @Param limit query int false "最大返回条数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/events [get]
func (c *KubeNodeController) Events(ctx *gin.Context) {
	param := requests.NewKubeNodeEventsRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeEventsRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()

	events, err := svc.KubeNodeEvents(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNodeEvents error", zap.Error(err))
		return
	}

	r.SuccessList(events, gin.H{
		"total":   len(events),
		"message": fmt.Sprintf("获取节点 %s 事件成功", param.Name),
	})
}

// Yaml godoc
// @Summary 获取 Node YAML 配置
// @Description 获取指定 Node 的 YAML 配置
// @Tags K8s Node 管理
// @Produce json
// @Param name query string true "Node 名称"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/yaml [get]
func (c *KubeNodeController) Yaml(ctx *gin.Context) {
	param := requests.NewKubeNodeDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeNodeDetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	yamlStr, err := nodepkg.GetYaml(ctx.Request.Context(), cli.Kube, param.Name)
	if err != nil {
		global.Logger.Error("获取 Node YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"name": param.Name,
		"yaml": yamlStr,
	})
}

// ApplyYaml godoc
// @Summary 应用 Node YAML 配置
// @Description 应用修改后的 YAML 配置到 Node
// @Tags K8s Node 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeApplyYamlClusterRequest true "YAML内容"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/node/apply_yaml [put]
func (c *KubeNodeController) ApplyYaml(ctx *gin.Context) {
	param := requests.NewKubeApplyYamlClusterRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeApplyYamlClusterRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	_, err := nodepkg.ApplyYaml(ctx.Request.Context(), cli.Kube, param.Name, param.Yaml)
	if err != nil {
		global.Logger.Error("应用 Node YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"name":    param.Name,
		"message": "YAML 应用成功",
	})
}
