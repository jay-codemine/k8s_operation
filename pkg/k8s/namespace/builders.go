package namespace

import (
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
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


func BuildNamespaceLabelPatch(addLabels map[string]string, removeLabels []string) ([]byte, error) {
	labels := map[string]interface{}{}

	for key, val := range addLabels {
		labels[key] = val
	}

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
