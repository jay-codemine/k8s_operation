package statefulset

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"time"
)

func RestartStatefulSet(ctx context.Context, Kube kubernetes.Interface, namespace, name string) (*appv1.StatefulSet, error) {
	// 定义用于记录重启时间的注释键名
	const restartedAtAnno = "kubectl.kubernetes.io/restartedAt"
	// 获取当前时间并格式化为RFC3339标准格式
	ts := time.Now().Format(time.RFC3339)

	// 构建用于更新StatefulSet的patch数据
	// 该patch会在pod模板的metadata中添加一个包含当前时间的注释
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]string{
						restartedAtAnno: ts, // 设置重启时间注释
					},
				},
			},
		},
	}

	// 将patch数据转换为JSON格式
	b, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("marshal patch failed: %w", err) // 如果JSON转换失败，返回错误
	}

	// 使用StrategicMergePatch类型对StatefulSet进行patch操作
	sts, err := Kube.AppsV1().
		StatefulSets(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, b, metav1.PatchOptions{})
	if err != nil {
		// 如果patch操作失败，记录错误日志并返回错误
		global.Logger.Error("restart statefulset (patch) failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	// 如果操作成功，记录信息日志
	global.Logger.Info("restart statefulset triggered",
		zap.String("namespace", namespace),
		zap.String("name", name),
		zap.String("restartedAt", ts),
	)
	return sts, nil // 当前实现中返回nil
}
