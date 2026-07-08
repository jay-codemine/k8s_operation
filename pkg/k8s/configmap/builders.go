package configmap

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/internal/app/requests"
)

// BuildConfigMapFromReq 根据请求构建 ConfigMap 对象（对应 BuildSecretFromReq）
func BuildConfigMapFromReq(req *requests.KubeConfigMapCreateRequest) (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		// 避免 nil map：后续赋值更安全
		Data:       map[string]string{},
		BinaryData: map[string][]byte{},
	}

	// 拷贝数据（不做业务校验，保持“构造器”纯粹）
	for k, v := range req.Data {
		cm.Data[k] = v
	}
	for k, v := range req.BinaryData {
		cm.BinaryData[k] = v
	}

	// 可选：轻量的健壮性检查（保留或删除均可）
	for k := range cm.Data {
		if k == "" {
			return nil, fmt.Errorf("data 中存在空键名")
		}
	}
	for k := range cm.BinaryData {
		if k == "" {
			return nil, fmt.Errorf("binaryData 中存在空键名")
		}
	}

	return cm, nil
}
