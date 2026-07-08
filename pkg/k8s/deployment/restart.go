package deployment

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

// RestartDeployment 给 Deployment 打注解触发滚动重启（等同于 kubectl rollout restart）
func RestartDeployment(ctx context.Context, Kube kubernetes.Interface, namespace, name string) error {
	// 官方约定的注解 key（kubectl rollout restart 使用的就是这个）
	const restartedAtAnno = "kubectl.kubernetes.io/restartedAt"

	// 获取当前时间（RFC3339 / ISO8601 格式字符串，例如 2025-10-02T12:34:56+08:00）
	ts := time.Now().Format(time.RFC3339)

	// 构造 patch 数据：只修改 Deployment 的 podTemplate.metadata.annotations
	// 注意：只要 podTemplate 发生变化，Deployment 控制器就会触发滚动更新
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]string{
						restartedAtAnno: ts, // 设置重启时间戳
					},
				},
			},
		},
	}

	// 将 map 转换为 JSON（用于 Patch 请求）
	b, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal patch failed: %w", err)
	}

	// 调用 Kubernetes API，使用 StrategicMergePatch 更新 Deployment
	// - StrategicMergePatchType 能按容器 name 等字段智能合并，比 MergePatchType 更安全
	_, err = Kube.AppsV1().
		Deployments(namespace).
		Patch(ctx, name, types.StrategicMergePatchType, b, metav1.PatchOptions{})
	if err != nil {
		// 失败时记录日志并返回错误
		global.Logger.Error("restart deployment (patch) failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return err
	}

	// 成功时记录一条日志，说明 Deployment 已触发滚动重启
	global.Logger.Info("restart deployment triggered",
		zap.String("namespace", namespace),
		zap.String("name", name),
		zap.String("restartedAt", ts),
	)

	return nil
}
