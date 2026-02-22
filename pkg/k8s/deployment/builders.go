package deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/dataselect"
	"k8soperation/pkg/k8s/probe"
)

// 描述注解的键名常量
const (
	DescriptionAnnotationKey = "description"
	SystemLabelKeyApp        = "system.k8soperation/app"
)

// BuildObjectMeta 根据 KubeDeploymentCreateRequest 构建对象元数据
// 1) 元数据构建函数
func BuildObjectMeta(req *requests.KubeDeploymentCreateRequest, labels map[string]string) metav1.ObjectMeta {
	annotations := map[string]string{}
	// 如果有描述信息，则添加到注解中
	// 检查请求中的描述是否不为空
	if req.Description != nil && *req.Description != "" {
		// 如果描述不为空，则将其添加到annotations映射中，键为DescriptionAnnotationKey
		annotations[DescriptionAnnotationKey] = *req.Description
	}
	// 返回一个metav1.ObjectMeta对象，其中包含名称、命名空间、标签和注解
	return metav1.ObjectMeta{
		Name:        req.Name,      // 设置对象名称为请求中的名称
		Namespace:   req.Namespace, // 设置对象命名空间为请求中的命名空间
		Labels:      labels,        // 设置标签为之前构建的labels映射
		Annotations: annotations,   // 设置注解为之前构建的annotations映射
	}
}

// BuildContainerFromCreateReq 从创建请求构建容器配置
// 2) 容器构建函数
func BuildContainerFromCreateReq(req *requests.KubeDeploymentCreateRequest) corev1.Container {
	/*
	 * 创建并配置一个Kubernetes容器对象
	 * 包含容器的基本配置、安全设置、资源限制、环境变量、命令参数、端口映射和健康检查等
	 */
	c := corev1.Container{
		Name:  req.Name,           // 容器名称：指定容器的名称
		Image: req.ContainerImage, // 容器镜像：指定容器使用的镜像
		SecurityContext: &corev1.SecurityContext{ // 安全上下文配置：定义容器的安全运行设置
			Privileged: &req.RunAsPrivileged, // 是否以特权模式运行：设置容器是否具有特权权限
		},
		Resources: corev1.ResourceRequirements{ // 资源需求配置：定义容器的资源限制和请求
			Requests: map[corev1.ResourceName]resource.Quantity{}, // 初始化资源请求：创建一个空的资源请求映射
		},
		Env: dataselect.ConvertEnvVarSpec(req.Variables), // 环境变量配置：将请求中的变量转换为容器环境变量
	}

	// command/args - 设置容器执行的命令和参数
	if req.ContainerCommand != nil && *req.ContainerCommand != "" {
		// 如果请求中包含容器命令且不为空，则将其设置为容器执行的命令
		c.Command = []string{*req.ContainerCommand}
	}
	if req.ContainerCommandArgs != nil && *req.ContainerCommandArgs != "" {
		// 如果请求中包含容器命令参数且不为空，则将其分割为字符串列表并设置为容器参数
		c.Args = strings.Fields(*req.ContainerCommandArgs)
	}

	// resources (Requests，可按需补 Limits) - 设置容器资源请求和限制
	// 优先使用新的 Resources 配置（Rancher/Kuboard 风格）
	if req.Resources != nil {
		// CPU Requests
		if req.Resources.CPURequest != nil && *req.Resources.CPURequest != "" {
			if q, err := resource.ParseQuantity(*req.Resources.CPURequest); err == nil {
				c.Resources.Requests[corev1.ResourceCPU] = q
			}
		}
		// CPU Limits
		if req.Resources.CPULimit != nil && *req.Resources.CPULimit != "" {
			if q, err := resource.ParseQuantity(*req.Resources.CPULimit); err == nil {
				if c.Resources.Limits == nil {
					c.Resources.Limits = make(map[corev1.ResourceName]resource.Quantity)
				}
				c.Resources.Limits[corev1.ResourceCPU] = q
			}
		}
		// Memory Requests
		if req.Resources.MemoryRequest != nil && *req.Resources.MemoryRequest != "" {
			if q, err := resource.ParseQuantity(*req.Resources.MemoryRequest); err == nil {
				c.Resources.Requests[corev1.ResourceMemory] = q
			}
		}
		// Memory Limits
		if req.Resources.MemoryLimit != nil && *req.Resources.MemoryLimit != "" {
			if q, err := resource.ParseQuantity(*req.Resources.MemoryLimit); err == nil {
				if c.Resources.Limits == nil {
					c.Resources.Limits = make(map[corev1.ResourceName]resource.Quantity)
				}
				c.Resources.Limits[corev1.ResourceMemory] = q
			}
		}
	} else {
		// 向下兼容：使用旧的 MemoryRequirement/CpuRequirement
		if req.MemoryRequirement != nil && *req.MemoryRequirement != "" {
			// 如果请求中包含内存需求且不为空，则尝试解析内存数量
			if q, err := resource.ParseQuantity(*req.MemoryRequirement); err == nil {
				// 如果解析成功，将内存请求量添加到容器的资源请求中
				c.Resources.Requests[corev1.ResourceMemory] = q
			}
		}
		// 检查CPU需求是否已设置且不为空
		if req.CpuRequirement != nil && *req.CpuRequirement != "" {
			// 尝试将CPU需求字符串解析为资源数量
			if q, err := resource.ParseQuantity(*req.CpuRequirement); err == nil {
				// 如果解析成功，将CPU请求量设置到容器的资源请求中
				c.Resources.Requests[corev1.ResourceCPU] = q
			}
		}
	}

	// ports
	// 检查请求中的端口验证列表是否大于0
	if len(req.PortMappings) > 0 {
		// 如果端口验证列表不为空，则将转换后的端口信息赋值给容器的端口属性
		c.Ports = probe.ConvertContainerPorts(req.PortMappings)
	}

	// probes - 优先使用新的 Probes 配置（支持 Startup Probe）
	if req.Probes != nil {
		// Readiness Probe - 就绪探针
		if req.Probes.EnableReadiness {
			c.ReadinessProbe = probe.BuildProbe(req.Probes.ReadinessProbe)
		}
		// Liveness Probe - 存活探针
		if req.Probes.EnableLiveness {
			c.LivenessProbe = probe.BuildProbe(req.Probes.LivenessProbe)
		}
		// Startup Probe - 启动探针（K8s 1.16+）
		if req.Probes.EnableStartup {
			c.StartupProbe = probe.BuildProbe(req.Probes.StartupProbe)
		}
	} else {
		// 向下兼容：使用旧的 IsReadinessEnable/IsLivenessEnable
		// 如果启用了就绪探针（Readiness Probe），则构建就绪探针并赋值给容器配置
		if req.IsReadinessEnable {
			c.ReadinessProbe = probe.BuildProbe(req.ReadinessProbe)
		}
		// 如果启用了存活探针（Liveness Probe），则构建存活探针并赋值给容器配置
		if req.IsLivenessEnable {
			c.LivenessProbe = probe.BuildProbe(req.LivenessProbe)
		}
	}

	// 返回配置好的容器配置
	return c
}

// 3) PodSpec
func BuildPodSpec(req *requests.KubeDeploymentCreateRequest, containers []corev1.Container) corev1.PodSpec {
	// 创建PodSpec对象，其中包含容器列表
	ps := corev1.PodSpec{Containers: containers}
	// 检查请求中是否包含镜像拉取密钥，且密钥名称不为空
	if req.ImagePullSecret != nil && *req.ImagePullSecret != "" {
		// 如果存在镜像拉取密钥，则将其添加到PodSpec的ImagePullSecrets字段中
		// 创建一个LocalObjectReference对象，名称为请求中指定的密钥名称
		ps.ImagePullSecrets = []corev1.LocalObjectReference{{Name: *req.ImagePullSecret}}
	}

	// 应用调度规则
	applySchedulingPolicy(&ps, req)

	// 返回配置好的PodSpec对象
	return ps
}

// applySchedulingPolicy 根据调度策略配置 PodSpec 的调度规则
func applySchedulingPolicy(ps *corev1.PodSpec, req *requests.KubeDeploymentCreateRequest) {
	// 1) 应用容忍配置（所有策略都可以有容忍）
	if len(req.Tolerations) > 0 {
		ps.Tolerations = req.Tolerations
	}

	// 2) 根据策略类型配置调度规则
	switch req.SchedulingPolicy {
	case requests.SchedulingPolicySpread:
		// 分散调度：使用 Pod 反亲和性，尽量将副本调度到不同节点
		ps.Affinity = buildSpreadAffinity(req.Name)

	case requests.SchedulingPolicyPack:
		// 集中调度：使用 Pod 亲和性，尽量将副本调度到同一节点
		ps.Affinity = buildPackAffinity(req.Name)

	case requests.SchedulingPolicyCustom:
		// 自定义规则：使用用户提供的配置
		if len(req.NodeSelector) > 0 {
			ps.NodeSelector = req.NodeSelector
		}
		
		// 构建完整的 Affinity（合并用户配置和简化格式）
		affinity := buildCustomAffinity(req)
		if affinity != nil {
			ps.Affinity = affinity
		}

	default:
		// 默认规则：不设置额外的调度规则
		// 如果用户传了 NodeSelector 也应用
		if len(req.NodeSelector) > 0 {
			ps.NodeSelector = req.NodeSelector
		}
	}

	// 3) 应用拓扑分布约束（所有策略都可以配置）
	if len(req.TopologySpreadConfigs) > 0 {
		ps.TopologySpreadConstraints = buildTopologySpreadConstraints(req)
	}
}

// buildCustomAffinity 根据用户配置构建自定义亲和性
func buildCustomAffinity(req *requests.KubeDeploymentCreateRequest) *corev1.Affinity {
	// 如果用户直接传了原生 Affinity，优先使用
	if req.Affinity != nil {
		return req.Affinity
	}

	// 如果没有简化格式的规则，返回 nil
	if len(req.NodeAffinityRules) == 0 {
		return nil
	}

	// 从简化格式构建 NodeAffinity
	affinity := &corev1.Affinity{}
	nodeAffinity := buildNodeAffinityFromRules(req.NodeAffinityRules)
	if nodeAffinity != nil {
		affinity.NodeAffinity = nodeAffinity
	}

	return affinity
}

// buildNodeAffinityFromRules 从简化规则构建 NodeAffinity
func buildNodeAffinityFromRules(rules []requests.NodeAffinityRule) *corev1.NodeAffinity {
	if len(rules) == 0 {
		return nil
	}

	var requiredTerms []corev1.NodeSelectorTerm
	var preferredTerms []corev1.PreferredSchedulingTerm

	for _, rule := range rules {
		if rule.Key == "" {
			continue
		}

		// 构建 NodeSelectorRequirement
		requirement := corev1.NodeSelectorRequirement{
			Key:      rule.Key,
			Operator: parseNodeSelectorOperator(rule.Operator),
		}

		// 处理值（逗号分隔转数组）
		if rule.Values != "" && rule.Operator != "Exists" && rule.Operator != "DoesNotExist" {
			requirement.Values = splitValues(rule.Values)
		}

		if rule.Required {
			// 硬性要求
			requiredTerms = append(requiredTerms, corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{requirement},
			})
		} else {
			// 软性偏好
			weight := rule.Weight
			if weight <= 0 {
				weight = 50 // 默认权重
			}
			if weight > 100 {
				weight = 100
			}
			preferredTerms = append(preferredTerms, corev1.PreferredSchedulingTerm{
				Weight: weight,
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{requirement},
				},
			})
		}
	}

	nodeAffinity := &corev1.NodeAffinity{}

	if len(requiredTerms) > 0 {
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{
			NodeSelectorTerms: requiredTerms,
		}
	}

	if len(preferredTerms) > 0 {
		nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferredTerms
	}

	if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil &&
		len(nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
		return nil
	}

	return nodeAffinity
}

// buildTopologySpreadConstraints 构建拓扑分布约束
func buildTopologySpreadConstraints(req *requests.KubeDeploymentCreateRequest) []corev1.TopologySpreadConstraint {
	var constraints []corev1.TopologySpreadConstraint

	for _, cfg := range req.TopologySpreadConfigs {
		if cfg.TopologyKey == "" {
			continue
		}

		maxSkew := cfg.MaxSkew
		if maxSkew <= 0 {
			maxSkew = 1 // 默认最大偏差为 1
		}

		whenUnsatisfiable := corev1.ScheduleAnyway // 默认软性约束
		if cfg.WhenUnsatisfiable == "DoNotSchedule" {
			whenUnsatisfiable = corev1.DoNotSchedule
		}

		constraint := corev1.TopologySpreadConstraint{
			TopologyKey:       cfg.TopologyKey,
			MaxSkew:           maxSkew,
			WhenUnsatisfiable: whenUnsatisfiable,
		}

		// 构建标签选择器（默认使用 Deployment 名称）
		if cfg.LabelSelector != "" {
			constraint.LabelSelector = parseLabelSelector(cfg.LabelSelector)
		} else {
			// 默认选择本 Deployment 的 Pod
			constraint.LabelSelector = &metav1.LabelSelector{
				MatchLabels: map[string]string{
					SystemLabelKeyApp: req.Name,
				},
			}
		}

		constraints = append(constraints, constraint)
	}

	return constraints
}

// parseNodeSelectorOperator 解析节点选择器操作符
func parseNodeSelectorOperator(op string) corev1.NodeSelectorOperator {
	switch op {
	case "In":
		return corev1.NodeSelectorOpIn
	case "NotIn":
		return corev1.NodeSelectorOpNotIn
	case "Exists":
		return corev1.NodeSelectorOpExists
	case "DoesNotExist":
		return corev1.NodeSelectorOpDoesNotExist
	case "Gt":
		return corev1.NodeSelectorOpGt
	case "Lt":
		return corev1.NodeSelectorOpLt
	default:
		return corev1.NodeSelectorOpIn
	}
}

// splitValues 将逗号分隔的值转换为数组
func splitValues(values string) []string {
	var result []string
	for _, v := range strings.Split(values, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			result = append(result, v)
		}
	}
	return result
}

// parseLabelSelector 解析标签选择器字符串 (key=value)
func parseLabelSelector(selector string) *metav1.LabelSelector {
	labels := make(map[string]string)
	for _, pair := range strings.Split(selector, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			labels[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	if len(labels) == 0 {
		return nil
	}
	return &metav1.LabelSelector{MatchLabels: labels}
}

// buildSpreadAffinity 构建分散调度的反亲和性配置
// 效果：尽可能将 Pod 副本调度到不同的节点上
func buildSpreadAffinity(appName string) *corev1.Affinity {
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			// 使用软性反亲和（preferred），避免因节点数不足导致调度失败
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								SystemLabelKeyApp: appName,
							},
						},
						TopologyKey: "kubernetes.io/hostname", // 按节点分散
					},
				},
			},
		},
	}
}

// buildPackAffinity 构建集中调度的亲和性配置
// 效果：尽可能将 Pod 副本调度到同一节点上
func buildPackAffinity(appName string) *corev1.Affinity {
	return &corev1.Affinity{
		PodAffinity: &corev1.PodAffinity{
			// 使用软性亲和（preferred），避免单节点资源不足导致调度失败
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								SystemLabelKeyApp: appName,
							},
						},
						TopologyKey: "kubernetes.io/hostname", // 按节点聚合
					},
				},
			},
		},
	}
}

// 统一合并标签：写入系统关键标签，防止被覆盖
func mergedLabels(raw map[string]string, appName string) map[string]string {
	out := map[string]string{SystemLabelKeyApp: appName} // 关键标签
	for k, v := range raw {
		switch k {
		case SystemLabelKeyApp:
			// 不允许用户覆盖关键/系统性标签
			continue
		default:
			out[k] = v
		}
	}
	return out
}

// 关键 selector（建议只放最小集合，避免未来因 selector 不可变而难以演进）
func requiredSelector(appName string) map[string]string {
	return map[string]string{SystemLabelKeyApp: appName}
}

func BuildDeploymentFromCreateReq(req *requests.KubeDeploymentCreateRequest) *appv1.Deployment {
	// 1) 用户传入的 labels -> 规范化（可能为空）
	userLabels := dataselect.GetLabelsMap(req.Labels)
	// 2) 合并系统关键标签
	labels := mergedLabels(userLabels, req.Name)
	// 3) 最小 selector（与 Pod 模版/Service 必须一致）
	selector := requiredSelector(req.Name)

	// 4) 元数据
	dpMeta := BuildObjectMeta(req, labels)

	// 5) PodTemplate（Namespace 会从 Deployment 继承）
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels, // 包含关键标签 + 自定义标签
			Annotations: dpMeta.Annotations,
		},
		Spec: BuildPodSpec(req, []corev1.Container{
			BuildContainerFromCreateReq(req),
		}),
	}

	// 6) 副本数兜底
	replicas := req.Replicas
	if replicas == 0 {
		replicas = 1
	}

	// 7) 历史版本保留数量（默认10个）
	revisionHistoryLimit := int32(10) // K8s 默认值
	if req.RevisionHistoryLimit != nil {
		revisionHistoryLimit = *req.RevisionHistoryLimit
	}

	// 8) 滚动更新策略（显式设置，符合 K8s 最佳实践）
	// maxSurge: 滚动更新时最多可以超出期望副本数的 Pod 数量（默认 25%）
	// maxUnavailable: 滚动更新时最多允许不可用的 Pod 数量（默认 25%）
	maxSurge := intstr.FromInt(1)       // 每次最多多创建 1 个 Pod
	maxUnavailable := intstr.FromInt(0) // 确保滚动更新期间服务不中断

	return &appv1.Deployment{
		ObjectMeta: dpMeta,
		Spec: appv1.DeploymentSpec{
			Replicas: &replicas,
			// selector 不可变，尽量只放最小关键标签，避免未来改动困难
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: podTemplate,
			// 显式设置滚动更新策略
			Strategy: appv1.DeploymentStrategy{
				Type: appv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appv1.RollingUpdateDeployment{
					MaxSurge:       &maxSurge,       // 每次最多多创建 1 个 Pod
					MaxUnavailable: &maxUnavailable, // 滚动更新期间保持服务可用
				},
			},
			RevisionHistoryLimit: &revisionHistoryLimit, // 历史版本保留数量
		},
	}
}

// BuildDeploymentResponse 用于统一构造创建 Deployment 成功后的返回体
func BuildDeploymentResponse(dp *appv1.Deployment, svc *corev1.Service, req *requests.KubeDeploymentCreateRequest) gin.H {
	// 创建一个响应数据结构，使用gin.H来构建一个嵌套的map结构
	// gin.H是Go语言中gin框架提供的快捷方式，本质上是一个map[string]interface{}
	resp := gin.H{
		// 部署(deployment)相关信息
		"deployment": gin.H{
			// 部署名称
			"name": dp.Name,
			// 部署所在的命名空间
			"namespace": dp.Namespace,
			// 部署的副本数量
			"replicas": dp.Spec.Replicas,
			// 容器镜像
			"image": req.ContainerImage,
			// 部署的标签
			"labels": dp.Labels,
			// 部署的唯一标识符
			"uid": string(dp.UID),
			// 资源版本号，用于跟踪资源的变更
			"resourceVersion": dp.ResourceVersion,
		},
	}

	// 检查服务(service)是否为空
	if svc != nil {
		// 创建一个包含服务数据的字典(map)
		svcData := gin.H{
			"created":  true,                  // 标记服务已创建
			"name":     svc.Name,              // 服务名称
			"type":     string(svc.Spec.Type), // 服务类型，转换为字符串
			"ports":    svc.Spec.Ports,        // 服务端口信息
			"selector": svc.Spec.Selector,     // 服务选择器
		}
		// 检查服务是否分配了集群IP
		if ip := svc.Spec.ClusterIP; ip != "" {
			svcData["clusterIP"] = ip // 添加集群IP到服务数据
		}
		// 检查负载均衡器是否有入站流量
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			svcData["ingress"] = svc.Status.LoadBalancer.Ingress // 添加负载均衡器入站信息
		}
		resp["service"] = svcData // 将服务数据添加到响应中
	}

	return resp // 返回包含服务数据的响应
}

// DeploymentListItem 列表项响应结构
type DeploymentListItem struct {
	Name              string                      `json:"name"`
	Namespace         string                      `json:"namespace"`
	Status            string                      `json:"status"`             // Running/Updating/Failed
	StatusReason      string                      `json:"status_reason"`      // 状态原因
	Replicas          int32                       `json:"replicas"`           // 期望副本数
	ReadyReplicas     int32                       `json:"ready_replicas"`     // 就绪副本数
	AvailableReplicas int32                       `json:"available_replicas"` // 可用副本数
	UpdatedReplicas   int32                       `json:"updated_replicas"`   // 已更新副本数
	Image             string                      `json:"image"`
	Images            []string                    `json:"images"`
	Containers        []string                    `json:"containers"`
	Selector          map[string]string           `json:"selector"`
	UpdateStrategy    string                      `json:"update_strategy"`
	CreatedAt         string                      `json:"created_at"`
	Conditions        []appv1.DeploymentCondition `json:"conditions"` // 原始 conditions
}

// BuildDeploymentListResponse 将 Deployment 列表转换为响应格式
func BuildDeploymentListResponse(deployments []appv1.Deployment) []DeploymentListItem {
	result := make([]DeploymentListItem, 0, len(deployments))

	for _, dp := range deployments {
		// 提取容器信息
		containers := []string{}
		images := []string{}
		var firstImage string

		if len(dp.Spec.Template.Spec.Containers) > 0 {
			for _, c := range dp.Spec.Template.Spec.Containers {
				containers = append(containers, c.Name)
				images = append(images, c.Image)
			}
			firstImage = dp.Spec.Template.Spec.Containers[0].Image
		}

		// 获取副本数
		replicas := int32(0)
		if dp.Spec.Replicas != nil {
			replicas = *dp.Spec.Replicas
		}

		// 获取状态
		status, reason := getDeploymentStatus(&dp)

		item := DeploymentListItem{
			Name:              dp.Name,
			Namespace:         dp.Namespace,
			Status:            status,
			StatusReason:      reason,
			Replicas:          replicas,
			ReadyReplicas:     dp.Status.ReadyReplicas,
			AvailableReplicas: dp.Status.AvailableReplicas,
			UpdatedReplicas:   dp.Status.UpdatedReplicas,
			Image:             firstImage,
			Images:            images,
			Containers:        containers,
			Selector:          dp.Spec.Selector.MatchLabels,
			UpdateStrategy:    string(dp.Spec.Strategy.Type),
			CreatedAt:         dp.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Conditions:        dp.Status.Conditions,
		}

		result = append(result, item)
	}

	return result
}

// getDeploymentStatus 根据 Deployment 的 status 判断状态
func getDeploymentStatus(dp *appv1.Deployment) (status, reason string) {
	// 默认状态
	status = "Unknown"
	reason = ""

	// 获取期望副本数
	desiredReplicas := int32(0)
	if dp.Spec.Replicas != nil {
		desiredReplicas = *dp.Spec.Replicas
	}

	readyReplicas := dp.Status.ReadyReplicas
	updatedReplicas := dp.Status.UpdatedReplicas
	availableReplicas := dp.Status.AvailableReplicas

	// 检查 conditions
	var availableCond, progressingCond *appv1.DeploymentCondition
	for i := range dp.Status.Conditions {
		cond := &dp.Status.Conditions[i]
		switch cond.Type {
		case appv1.DeploymentAvailable:
			availableCond = cond
		case appv1.DeploymentProgressing:
			progressingCond = cond
		}
	}

	// 1. 检查是否处于失败状态
	if availableCond != nil && availableCond.Status == corev1.ConditionFalse {
		status = "Failed"
		reason = availableCond.Reason + ": " + availableCond.Message
		return
	}

	// 2. 检查是否处于进度暂停或失败
	if progressingCond != nil {
		if progressingCond.Status == corev1.ConditionFalse {
			status = "Failed"
			reason = progressingCond.Reason + ": " + progressingCond.Message
			return
		}
		// ProgressDeadlineExceeded 表示更新超时
		if progressingCond.Reason == "ProgressDeadlineExceeded" {
			status = "Failed"
			reason = "更新超时: " + progressingCond.Message
			return
		}
	}

	// 3. 期望副本数为 0（停服状态）
	if desiredReplicas == 0 {
		if readyReplicas == 0 {
			status = "Stopped"
			reason = "副本数为 0"
		} else {
			status = "Updating"
			reason = "正在缩容到 0"
		}
		return
	}

	// 4. 检查是否正在更新
	if updatedReplicas < desiredReplicas {
		status = "Updating"
		reason = fmt.Sprintf("正在滚动更新 (%d/%d 已更新)", updatedReplicas, desiredReplicas)
		return
	}

	// 5. 检查就绪状态
	if readyReplicas < desiredReplicas {
		status = "Updating"
		if readyReplicas == 0 {
			status = "Failed"
			reason = fmt.Sprintf("0/%d Pod 就绪", desiredReplicas)
		} else {
			reason = fmt.Sprintf("%d/%d Pod 就绪", readyReplicas, desiredReplicas)
		}
		return
	}

	// 6. 检查可用性
	if availableReplicas < desiredReplicas {
		status = "Updating"
		reason = fmt.Sprintf("%d/%d Pod 可用", availableReplicas, desiredReplicas)
		return
	}

	// 7. 一切正常
	if readyReplicas == desiredReplicas && availableReplicas == desiredReplicas {
		status = "Running"
		reason = "所有副本就绪"
		return
	}

	// 兜底
	status = "Unknown"
	reason = "状态未知"
	return
}

// 根据 Deployment 的创建请求构建 Service（默认 ClusterIP）
// 注意：Service 的 selector 必须与 Deployment 的 selector 一致（至少包含关键标签）
func BuildServiceFromDeploymentReq(req *requests.KubeDeploymentCreateRequest) *corev1.Service {
	// 1) 用户 labels 规范化 + 合并关键标签（用于 Service 自身的 Labels）
	userLabels := dataselect.GetLabelsMap(req.Labels)
	labels := mergedLabels(userLabels, req.Name)

	// 2) 与 Deployment 对齐的最小 selector
	selector := requiredSelector(req.Name)

	// 3) Service 名称兜底
	name := req.ServiceName
	if name == "" {
		name = req.Name
	}

	// 4) Service 类型/端口兜底
	svcType := corev1.ServiceTypeClusterIP
	if req.ServiceType != "" {
		svcType = corev1.ServiceType(req.ServiceType)
	}
	ports := ConvertServicePorts(req.PortMappings) // 你的现有转换逻辑
	// 可选：如果没有端口映射且你需要兜底，可在此处根据 req.Port 构造一个 ServicePort

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: req.Namespace,
			Labels:    labels, // Service 自身标签：关键 + 自定义
		},
		Spec: corev1.ServiceSpec{
			Type:     svcType,
			Selector: selector, // 关键：与 Deployment 的 selector 保持一致
			Ports:    ports,
		},
	}
}

// ---- 端口工具：Service ----

// ConvertServicePorts 将请求里的端口映射转换为 ServicePort 列表
func ConvertServicePorts(ports []requests.PortMapping) []corev1.ServicePort {
	// 创建一个新的ServicePort切片，初始容量为ports切片的长度，以避免后续扩容
	out := make([]corev1.ServicePort, 0, len(ports))
	// 遍历输入的ports切片
	for _, p := range ports {
		// 检查端口和目标端口是否合法（大于0）
		if p.Port <= 0 || p.TargetPort <= 0 {
			continue // 如果端口不合法，跳过当前循环
		}
		// 将合法的端口信息添加到out切片中
		out = append(out, corev1.ServicePort{
			Name:       buildPortName(p.Protocol, p.Port), // 合法可读的端口名
			Port:       p.Port,                            // Service 对外端口
			TargetPort: intstr.FromInt32(p.TargetPort),    // 指向容器端口
			Protocol:   parseProtocol(p.Protocol),         // 解析并设置协议类型
		})
	}
	// 返回处理后的ServicePort切片
	return out
}

// 解析协议字符串到 corev1.Protocol（默认 TCP）
func parseProtocol(s string) corev1.Protocol {
	// 使用 switch 语句对输入字符串 s 进行处理，将其转换为大写并去除前后空格
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "UDP":
		// 如果匹配到 "UDP"，则返回 corev1.ProtocolUDP
		return corev1.ProtocolUDP
	case "SCTP":
		// 如果匹配到 "SCTP"，则返回 corev1.ProtocolSCTP
		return corev1.ProtocolSCTP
	default:
		// 如果不匹配任何已知协议，则默认返回 corev1.ProtocolTCP
		return corev1.ProtocolTCP
	}
}

// 给端口生成一个可读的名称（可选）
// 例: "http-80"、"tcp-8080"
func buildPortName(proto string, port int32) string {
	// 将协议字符串转换为小写并去除前后空格
	p := strings.ToLower(strings.TrimSpace(proto))
	// 如果处理后的协议字符串为空，则使用默认值"tcp"
	if p == "" {
		p = "tcp"
	}
	// 返回格式化后的字符串，格式为"协议-端口号"
	return fmt.Sprintf("%s-%d", p, port)
}
