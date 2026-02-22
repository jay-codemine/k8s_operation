package namespace

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
)

// BuildNamespaceFromReq 根据请求构造 Namespace 对象
func BuildNamespaceFromReq(req *requests.KubeNamespaceCreateRequest) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
	}
}

// BuildQuotaForNamespace 构造 ResourceQuota（CPU/内存/Pods）
func BuildQuotaForNamespace(nsName, cpu, mem, pods string) *corev1.ResourceQuota {
	return &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-quota",
			Namespace: nsName,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceLimitsCPU:    resource.MustParse(cpu),
				corev1.ResourceLimitsMemory: resource.MustParse(mem),
				corev1.ResourcePods:         resource.MustParse(pods),
			},
		},
	}
}

// BuildLimitRangeForNamespace 构造默认 LimitRange
func BuildLimitRangeForNamespace(nsName string) *corev1.LimitRange {
	/*
	 * 返回一个 LimitRange 对象的指针
	 * 该对象定义了命名空间中默认的容器资源限制
	 */
	return &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default-limits", // LimitRange 的名称
			Namespace: nsName,           // 所属的命名空间
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem{
				{
					Type: corev1.LimitTypeContainer, // 限制类型为容器级别
					Default: corev1.ResourceList{ // 默认资源限制
						corev1.ResourceCPU:    resource.MustParse("500m"), // 默认 CPU 限制为 500m
						corev1.ResourceMemory: resource.MustParse("1Gi"),  // 默认内存限制为 1Gi
					},
					DefaultRequest: corev1.ResourceList{ // 默认资源请求量
						corev1.ResourceCPU:    resource.MustParse("200m"),  // 默认 CPU 请求为 200m
						corev1.ResourceMemory: resource.MustParse("256Mi"), // 默认内存请求为 256Mi
					},
				},
			},
		},
	}
}

func BuildNamespaceLabelPatch(addLabels map[string]string, removeLabels []string) ([]byte, error) {
	labels := map[string]interface{}{}

	// 添加/更新 labels
	for key, val := range addLabels {
		labels[key] = val
	}

	// 删除 labels（K8s 用 null 表示删除）
	for _, key := range removeLabels {
		labels[key] = nil
	}

	patch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": labels,
		},
	}

	return json.Marshal(patch)
}
