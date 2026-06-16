package statefulset

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

func PatchScaleReplicasStatefulSet(ctx context.Context, Kube kubernetes.Interface, namespace, name string, replicas int32,
) (*appv1.StatefulSet, error) {
	patch := map[string]interface{}{
		"spec": map[string]interface{}{"replicas": replicas},
	}
	bs, _ := json.Marshal(patch)

	sts, err := Kube.AppsV1().
		StatefulSets(namespace).
		Patch(ctx, name, types.MergePatchType, bs, metav1.PatchOptions{})
	if err != nil {
		switch {
		case errors.IsNotFound(err):
			global.Logger.Warn("statefulset not found",
				zap.String("namespace", namespace), zap.String("name", name))
			return nil, fmt.Errorf("statefulset %s/%s not found", namespace, name)
		case errors.IsForbidden(err):
			return nil, fmt.Errorf("forbidden to patch statefulset %s/%s: %w", namespace, name, err)
		default:
			global.Logger.Error("patch statefulset replicas failed",
				zap.String("namespace", namespace), zap.String("name", name), zap.Error(err))
			return nil, fmt.Errorf("patch statefulset %s/%s replicas=%d failed: %w", namespace, name, replicas, err)
		}
	}

	global.Logger.Info("scaled statefulset",
		zap.String("namespace", namespace),
		zap.String("name", name),
		zap.Int32("replicas", replicas),
	)
	return sts, nil
}
