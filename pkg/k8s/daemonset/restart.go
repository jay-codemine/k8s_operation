package daemonset

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"time"
)

// RestartDaemonSet 触发 DaemonSet 滚动重启（等价于：kubectl rollout restart ds <name> -n <ns>）
func RestartDaemonSet(ctx context.Context, Kube kubernetes.Interface, namespace, name string) error {
	const restartedAtAnno = "kubectl.kubernetes.io/restartedAt"
	ts := time.Now().Format(time.RFC3339)

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						restartedAtAnno: ts,
					},
				},
			},
		},
	}

	b, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal patch failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = Kube.AppsV1().
		DaemonSets(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, b, metav1.PatchOptions{})
	if err != nil {
		global.Logger.Error("restart daemonset (patch) failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return err
	}

	global.Logger.Info("restart daemonset triggered",
		zap.String("namespace", namespace),
		zap.String("name", name),
		zap.String("restartedAt", ts),
	)
	return nil
}
