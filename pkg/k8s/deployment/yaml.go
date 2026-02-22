// pkg/k8s/deployment/yaml.go
package deployment

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// GetYaml 获取 Deployment 的 YAML 配置
func GetYaml(ctx context.Context, client kubernetes.Interface, namespace, name string) (string, error) {
	deploy, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// 清理不需要的字段
	cleanDeploymentForYaml(deploy)

	// 转换为 YAML
	yamlBytes, err := yaml.Marshal(deploy)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// ApplyYaml 应用 Deployment YAML 配置
func ApplyYaml(ctx context.Context, client kubernetes.Interface, namespace, name, yamlContent string) (*appv1.Deployment, error) {
	// 解析 YAML
	var deploy appv1.Deployment
	if err := yaml.Unmarshal([]byte(yamlContent), &deploy); err != nil {
		return nil, err
	}

	// 确保 namespace 和 name 匹配
	deploy.Namespace = namespace
	deploy.Name = name

	// 获取现有资源以保留必要的字段
	existing, err := client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// 保留资源版本以支持更新
	deploy.ResourceVersion = existing.ResourceVersion

	// 更新 Deployment
	updated, err := client.AppsV1().Deployments(namespace).Update(ctx, &deploy, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

// cleanDeploymentForYaml 清理 Deployment 对象中的运行时字段
func cleanDeploymentForYaml(deploy *appv1.Deployment) {
	// 清理 TypeMeta
	deploy.APIVersion = "apps/v1"
	deploy.Kind = "Deployment"

	// 清理 ObjectMeta 中的运行时字段
	deploy.ManagedFields = nil
	deploy.UID = ""
	deploy.ResourceVersion = ""
	deploy.Generation = 0
	deploy.CreationTimestamp = metav1.Time{}
	deploy.DeletionTimestamp = nil
	deploy.DeletionGracePeriodSeconds = nil
	deploy.SelfLink = ""

	// 清理 Status
	deploy.Status = appv1.DeploymentStatus{}
}
