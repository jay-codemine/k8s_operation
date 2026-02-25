// pkg/k8s/statefulset/builders.go
package statefulset

import (
	"fmt"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/dataselect"
	"k8soperation/pkg/k8s/probe"
)

// 描述注解与系统关键标签
const (
	DescriptionAnnotationKey = "description"
	SystemLabelKeyApp        = "system.k8soperation/app"
)

// BuildObjectMeta 根据请求构建对象元数据
func BuildObjectMeta(req *requests.KubeStatefulSetCreateRequest, labels map[string]string) metav1.ObjectMeta {
	annotations := map[string]string{}
	if req.Description != nil && *req.Description != "" {
		annotations[DescriptionAnnotationKey] = *req.Description
	}
	return metav1.ObjectMeta{
		Name:        req.Name,
		Namespace:   req.Namespace,
		Labels:      labels,
		Annotations: annotations,
	}
}

// BuildContainerFromCreateReq 从请求构建容器配置
func BuildContainerFromCreateReq(req *requests.KubeStatefulSetCreateRequest) corev1.Container {
	c := corev1.Container{
		Name:  req.Name,
		Image: req.ContainerImage,
		// 更稳妥的指针写法
		SecurityContext: &corev1.SecurityContext{
			Privileged: pointer.Bool(req.RunAsPrivileged),
		},
		Resources: corev1.ResourceRequirements{
			Requests: map[corev1.ResourceName]resource.Quantity{},
		},
		Env: dataselect.ConvertEnvVarSpec(req.Variables),
		// 建议显式设置，避免 latest 带来的不确定性；如需更细策略可由 req 控制
		ImagePullPolicy: corev1.PullIfNotPresent,
	}

	// Command / Args
	if req.ContainerCommand != nil && *req.ContainerCommand != "" {
		c.Command = []string{*req.ContainerCommand}
	}
	if req.ContainerCommandArgs != nil && *req.ContainerCommandArgs != "" {
		c.Args = strings.Fields(*req.ContainerCommandArgs)
	}

	// Resources
	if req.MemoryRequirement != nil && *req.MemoryRequirement != "" {
		if q, err := resource.ParseQuantity(*req.MemoryRequirement); err == nil {
			c.Resources.Requests[corev1.ResourceMemory] = q
		}
	}
	if req.CpuRequirement != nil && *req.CpuRequirement != "" {
		if q, err := resource.ParseQuantity(*req.CpuRequirement); err == nil {
			c.Resources.Requests[corev1.ResourceCPU] = q
		}
	}

	// Ports
	if len(req.PortMappings) > 0 {
		c.Ports = probe.ConvertContainerPorts(req.PortMappings)
	}

	// Probes
	if req.IsReadinessEnable {
		c.ReadinessProbe = probe.BuildProbe(req.ReadinessProbe)
	}
	if req.IsLivenessEnable {
		c.LivenessProbe = probe.BuildProbe(req.LivenessProbe)
	}

	return c
}

// BuildPodSpec 构建 PodSpec
func BuildPodSpec(req *requests.KubeStatefulSetCreateRequest, containers []corev1.Container) corev1.PodSpec {
	ps := corev1.PodSpec{Containers: containers}
	if req.ImagePullSecret != nil && *req.ImagePullSecret != "" {
		ps.ImagePullSecrets = []corev1.LocalObjectReference{{Name: *req.ImagePullSecret}}
	}
	return ps
}

// 合并关键标签
func mergedLabels(raw map[string]string, appName string) map[string]string {
	out := map[string]string{SystemLabelKeyApp: appName}
	for k, v := range raw {
		if k == SystemLabelKeyApp {
			continue
		}
		out[k] = v
	}
	return out
}

// 构建 selector（StatefulSet 与 Deployment 同理）
func requiredSelector(appName string) map[string]string {
	return map[string]string{SystemLabelKeyApp: appName}
}

// BuildStatefulSetFromCreateReq 构造 StatefulSet 资源对象
func BuildStatefulSetFromCreateReq(req *requests.KubeStatefulSetCreateRequest) *appv1.StatefulSet {
	userLabels := dataselect.GetLabelsMap(req.Labels)
	labels := mergedLabels(userLabels, req.Name)
	selector := requiredSelector(req.Name)
	stsMeta := BuildObjectMeta(req, labels)

	// 副本数
	var replicas int32 = 1
	if req.Replicas > 0 {
		replicas = int32(req.Replicas)
	}

	// 1) 先构造容器
	container := BuildContainerFromCreateReq(req)

	// 2) 为容器追加卷挂载：与 PVC 模板名称一一对应
	for _, t := range req.VolumeClaimTemplates {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			continue
		}
		mp := strings.TrimSpace(t.MountPath)
		if mp == "" {
			mp = "/data_storage" // 默认挂载点，可按需修改
		}
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      name, // 必须与 volumeClaimTemplates[].metadata.name 一致
			MountPath: mp,
		})
	}

	// 3) 用补过挂载的容器构造 Pod 模板
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: stsMeta.Annotations,
		},
		Spec: BuildPodSpec(req, []corev1.Container{container}),
	}

	// 4) 构造 PVC 模板（会用于自动创建 data-<sts>-<ordinal> PVC）
	volumeClaims := BuildPVCs(req.VolumeClaimTemplates)

	// 5) 滚动更新策略（显式设置，支持 partition 金丝雀发布）
	// partition: 指定从哪个序号开始更新（例如 partition=2，则只更新 pod-2、pod-3...）
	// 默认为 0，表示全量更新所有 Pod
	updateStrategy := appv1.StatefulSetUpdateStrategy{
		Type: appv1.RollingUpdateStatefulSetStrategyType,
	}
	if req.UpdatePartition != nil {
		updateStrategy.RollingUpdate = &appv1.RollingUpdateStatefulSetStrategy{
			Partition: req.UpdatePartition,
		}
	}

	// 6) 组装 StatefulSet
	return &appv1.StatefulSet{
		ObjectMeta: stsMeta,
		Spec: appv1.StatefulSetSpec{
			ServiceName:          req.ServiceName,
			Replicas:             pointer.Int32(replicas),
			Selector:             &metav1.LabelSelector{MatchLabels: selector},
			Template:             podTemplate,
			VolumeClaimTemplates: volumeClaims,
			UpdateStrategy:       updateStrategy, // 显式设置滚动更新策略
			// PVC 自动删除策略：删除 StatefulSet 时自动删除 PVC（但保留 PV）
			// WhenDeleted=Delete: 删除 StatefulSet 时自动删除 PVC
			// WhenScaled=Retain: 缩容时保留 PVC，避免数据丢失
			// 注意：PVC 删除不会删除 PV（如果 PV 的 ReclaimPolicy=Retain）
			PersistentVolumeClaimRetentionPolicy: &appv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appv1.DeletePersistentVolumeClaimRetentionPolicyType, // 删除时自动清理 PVC
				WhenScaled:  appv1.RetainPersistentVolumeClaimRetentionPolicyType,  // 缩容时保留 PVC
			},
		},
	}
}

// BuildPVCs 将请求中的 PVC 模板转换为 K8s 规范的 VolumeClaimTemplates
func BuildPVCs(templates []requests.VolumeClaimTemplate) []corev1.PersistentVolumeClaim {
	out := make([]corev1.PersistentVolumeClaim, 0, len(templates))
	for _, t := range templates {
		if t.StorageSize == "" || t.Name == "" {
			continue
		}
		q, err := resource.ParseQuantity(t.StorageSize)
		if err != nil || q.Sign() <= 0 {
			continue
		}

		// 访问模式（安全映射，非法值回退到 RWO）
		accessMode := toAccessMode(t.AccessMode)

		var sc *string
		if strings.TrimSpace(t.StorageClass) != "" {
			v := t.StorageClass
			sc = &v
		}

		out = append(out, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: t.Name,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{accessMode},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: q,
					},
					// 不设置 Limits，避免把卷大小“锁死”；支持后续在线扩容
				},
				StorageClassName: sc,
			},
		})
	}
	return out
}

// toAccessMode 将字符串安全映射为 AccessMode
func toAccessMode(s string) corev1.PersistentVolumeAccessMode {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "RWO", "READWRITEONCE":
		return corev1.ReadWriteOnce
	case "RWX", "READWRITEMANY":
		return corev1.ReadWriteMany
	case "ROX", "READONLYMANY":
		return corev1.ReadOnlyMany
	default:
		return corev1.ReadWriteOnce
	}
}

// BuildStatefulSetResponse 用于返回创建结果
func BuildStatefulSetResponse(sts *appv1.StatefulSet, svc *corev1.Service, req *requests.KubeStatefulSetCreateRequest) gin.H {
	rep := int32(0)
	if sts.Spec.Replicas != nil {
		rep = *sts.Spec.Replicas
	}

	resp := gin.H{
		"stateful": gin.H{
			"name":            sts.Name,
			"namespace":       sts.Namespace,
			"replicas":        rep,
			"image":           req.ContainerImage,
			"labels":          sts.Labels,
			"uid":             string(sts.UID),
			"resourceVersion": sts.ResourceVersion,
			"serviceName":     req.ServiceName,
			"volumeClaimCount": func() int {
				if sts.Spec.VolumeClaimTemplates == nil {
					return 0
				}
				return len(sts.Spec.VolumeClaimTemplates)
			}(),
		},
	}

	if svc != nil {
		svcData := gin.H{
			"created":  true,
			"name":     svc.Name,
			"type":     string(svc.Spec.Type),
			"ports":    svc.Spec.Ports,
			"selector": svc.Spec.Selector,
		}
		if ip := svc.Spec.ClusterIP; ip != "" {
			svcData["clusterIP"] = ip
		}
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			svcData["ingress"] = svc.Status.LoadBalancer.Ingress
		}
		resp["service"] = svcData
	}

	return resp
}

// BuildServiceFromStatefulSetReq 根据 StatefulSet 请求构建 Headless Service（clusterIP=None）
func BuildServiceFromStatefulSetReq(req *requests.KubeStatefulSetCreateRequest) *corev1.Service {
	userLabels := dataselect.GetLabelsMap(req.Labels)
	labels := mergedLabels(userLabels, req.Name)
	selector := requiredSelector(req.Name)

	name := req.ServiceName
	if name == "" {
		name = req.Name
	}

	ports := ConvertServicePorts(req.PortMappings)
	meta := BuildObjectMeta(req, labels)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   req.Namespace,
			Labels:      labels,
			Annotations: meta.Annotations,
		},
		Spec: corev1.ServiceSpec{
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
			Selector:  selector,
			Ports:     ports,
		},
	}
}

// ConvertServicePorts 将端口映射转换为 ServicePort
func ConvertServicePorts(ports []requests.PortMapping) []corev1.ServicePort {
	out := make([]corev1.ServicePort, 0, len(ports))
	for _, p := range ports {
		if p.Port <= 0 || p.TargetPort <= 0 {
			continue
		}
		out = append(out, corev1.ServicePort{
			Name:       buildPortName(p.Protocol, p.Port),
			Port:       p.Port,
			TargetPort: intstr.FromInt32(p.TargetPort), // 兼容性更好
			Protocol:   parseProtocol(p.Protocol),
		})
	}
	return out
}

func parseProtocol(s string) corev1.Protocol {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "UDP":
		return corev1.ProtocolUDP
	case "SCTP":
		return corev1.ProtocolSCTP
	default:
		return corev1.ProtocolTCP
	}
}

func buildPortName(proto string, port int32) string {
	p := strings.ToLower(strings.TrimSpace(proto))
	if p == "" {
		p = "tcp"
	}
	return fmt.Sprintf("%s-%d", p, port)
}

// =========================
// StatefulSet 列表响应转换
// =========================

// StatefulSetListItem 列表项响应结构（前端友好格式）
type StatefulSetListItem struct {
	Name            string                       `json:"name"`
	Namespace       string                       `json:"namespace"`
	Status          string                       `json:"status"`           // Running/Updating/Failed/Stopped
	StatusReason    string                       `json:"status_reason"`    // 状态原因
	Replicas        int32                        `json:"replicas"`         // 期望副本数
	ReadyReplicas   int32                        `json:"ready_replicas"`   // 就绪副本数
	CurrentReplicas int32                        `json:"current_replicas"` // 当前副本数
	UpdatedReplicas int32                        `json:"updated_replicas"` // 已更新副本数
	Image           string                       `json:"image"`
	Images          []string                     `json:"images"`
	Containers      []string                     `json:"containers"`
	Selector        map[string]string            `json:"selector"`
	UpdateStrategy  string                       `json:"update_strategy"`
	ServiceName     string                       `json:"service_name"`
	CreatedAt       string                       `json:"created_at"`
	Conditions      []appv1.StatefulSetCondition `json:"conditions"` // 原始 conditions
}

// BuildStatefulSetListResponse 将 StatefulSet 列表转换为响应格式
func BuildStatefulSetListResponse(statefulsets []appv1.StatefulSet) []StatefulSetListItem {
	result := make([]StatefulSetListItem, 0, len(statefulsets))

	for _, sts := range statefulsets {
		// 提取容器信息
		containers := []string{}
		images := []string{}
		var firstImage string

		if len(sts.Spec.Template.Spec.Containers) > 0 {
			for _, c := range sts.Spec.Template.Spec.Containers {
				containers = append(containers, c.Name)
				images = append(images, c.Image)
			}
			firstImage = sts.Spec.Template.Spec.Containers[0].Image
		}

		// 获取副本数
		replicas := int32(0)
		if sts.Spec.Replicas != nil {
			replicas = *sts.Spec.Replicas
		}

		// 获取选择器
		selector := map[string]string{}
		if sts.Spec.Selector != nil && sts.Spec.Selector.MatchLabels != nil {
			selector = sts.Spec.Selector.MatchLabels
		}

		// 获取状态
		status, reason := getStatefulSetStatus(&sts)

		item := StatefulSetListItem{
			Name:            sts.Name,
			Namespace:       sts.Namespace,
			Status:          status,
			StatusReason:    reason,
			Replicas:        replicas,
			ReadyReplicas:   sts.Status.ReadyReplicas,
			CurrentReplicas: sts.Status.CurrentReplicas,
			UpdatedReplicas: sts.Status.UpdatedReplicas,
			Image:           firstImage,
			Images:          images,
			Containers:      containers,
			Selector:        selector,
			UpdateStrategy:  string(sts.Spec.UpdateStrategy.Type),
			ServiceName:     sts.Spec.ServiceName,
			CreatedAt:       sts.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Conditions:      sts.Status.Conditions,
		}

		result = append(result, item)
	}

	return result
}

// getStatefulSetStatus 根据 StatefulSet 的 status 判断状态
func getStatefulSetStatus(sts *appv1.StatefulSet) (status, reason string) {
	// 默认状态
	status = "Unknown"
	reason = ""

	// 获取期望副本数
	desiredReplicas := int32(0)
	if sts.Spec.Replicas != nil {
		desiredReplicas = *sts.Spec.Replicas
	}

	readyReplicas := sts.Status.ReadyReplicas
	updatedReplicas := sts.Status.UpdatedReplicas
	currentReplicas := sts.Status.CurrentReplicas

	// 1. 期望副本数为 0（停服状态）
	if desiredReplicas == 0 {
		if readyReplicas == 0 && currentReplicas == 0 {
			status = "Stopped"
			reason = "副本数为 0"
		} else {
			status = "Updating"
			reason = "正在缩容到 0"
		}
		return
	}

	// 2. 检查是否正在更新（updatedReplicas 小于 desiredReplicas）
	if updatedReplicas < desiredReplicas {
		status = "Updating"
		reason = fmt.Sprintf("正在滚动更新 (%d/%d 已更新)", updatedReplicas, desiredReplicas)
		return
	}

	// 3. 检查就绪状态
	if readyReplicas < desiredReplicas {
		if readyReplicas == 0 && currentReplicas == 0 {
			status = "Pending"
			reason = fmt.Sprintf("等待 Pod 创建 (0/%d)", desiredReplicas)
		} else if readyReplicas == 0 {
			status = "Failed"
			reason = fmt.Sprintf("0/%d Pod 就绪", desiredReplicas)
		} else {
			status = "Updating"
			reason = fmt.Sprintf("%d/%d Pod 就绪", readyReplicas, desiredReplicas)
		}
		return
	}

	// 4. 一切正常
	if readyReplicas == desiredReplicas {
		status = "Running"
		reason = "所有副本就绪"
		return
	}

	// 兜底
	status = "Unknown"
	reason = "状态未知"
	return
}
