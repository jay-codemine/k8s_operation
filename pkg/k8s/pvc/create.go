package pvc

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

// CreatePersistentVolumeClaim 创建 PVC（命名空间级）
// 支持两种模式：
// 1. YAML 模式：如果 req.YamlContent 不为空，则从 YAML 创建
// 2. 表单模式：否则从表单字段构建 PVC 对象
func CreatePersistentVolumeClaim(ctx context.Context, Kube kubernetes.Interface, req *requests.KubePVCCreateRequest) (*corev1.PersistentVolumeClaim, error) {
	var pvc *corev1.PersistentVolumeClaim
	var err error
	
	// 判断是否使用 YAML 模式
	if req.YamlContent != "" {
		// YAML 模式：从 YAML 解析
		pvc, err = decodePVCFromYAML(req.YamlContent)
		if err != nil {
			return nil, fmt.Errorf("解析 YAML 失败: %w", err)
		}
	} else {
		// 表单模式：从请求字段构建
		pvc = BuildPVCFromReq(req)
	}

	// 2) 调用 Kubernetes API 创建
	namespace := pvc.Namespace
	if namespace == "" {
		namespace = req.Namespace // 预防 YAML 中没有 namespace
		if namespace == "" {
			namespace = "default"
		}
		pvc.Namespace = namespace
	}
	
	created, err := Kube.CoreV1().
		PersistentVolumeClaims(namespace).
		Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("PersistentVolumeClaim %q already exists in namespace %q", pvc.Name, namespace)
		}
		global.Logger.Errorf("create PersistentVolumeClaim failed: %v", err)
		return nil, err
	}

	global.Logger.Infof("PersistentVolumeClaim %q created successfully in namespace %q", created.Name, created.Namespace)
	return created, nil
}

// decodePVCFromYAML 从 YAML 字符串解析 PVC 对象
func decodePVCFromYAML(yamlContent string) (*corev1.PersistentVolumeClaim, error) {
	decode := serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yamlContent), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("YAML 解析错误: %w", err)
	}

	pvc, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		return nil, fmt.Errorf("YAML 内容不是有效的 PersistentVolumeClaim 对象，实际类型: %T", obj)
	}

	return pvc, nil
}
