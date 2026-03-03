package services

import (
	"context"
	"fmt"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/pkg/k8s/common"
	"k8soperation/pkg/k8s/event"
	"strings"

	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/deployment"
)

// 列表
// services/deployment.go
func (s *Services) KubeDeploymentList(ctx context.Context, cli *K8sClients,
	param *requests.KubeDeploymentListRequest,
) ([]appv1.Deployment, int64, error) {

	deployments, total, err := deployment.GetDeploymentList(
		ctx,
		cli.Kube,
		param.Name,
		param.Namespace,
		param.Page,
		param.Limit,
	)
	if err != nil {
		return nil, 0, err
	}
	return deployments, total, nil
}

// 删除
func (s *Services) KubeDeploymentDelete(ctx context.Context, cli *K8sClients, param *requests.KubeDeploymentDeleteRequest) error {
	return deployment.DeleteDeployment(ctx, cli.Kube, param.Name, param.Namespace)
}

// 删除 Service
func (s *Services) KubeDeploymentDeleteService(ctx context.Context, cli *K8sClients, param *requests.KubeDeploymentDeleteRequest) error {
	return deployment.DeleteService(ctx, cli.Kube, param.Name, param.Namespace)
}

// 扩缩容（改副本数）
func (s *Services) KubeUpdateDeploymentReplicas(ctx context.Context, cli *K8sClients, param *requests.KubeDeploymentScaleRequest) (*appv1.Deployment, error) {
	return deployment.PatchDeploymentReplicas(ctx, cli.Kube, param.Namespace, param.Name, param.ScaleNum)
}

// 更新镜像
func (s *Services) KubeUpdateDeploymentImage(ctx context.Context, cli *K8sClients, param *requests.KubeDeploymentUpdateImageRequest) (*appv1.Deployment, error) {
	return deployment.PatchDeploymentImage(ctx, cli.Kube, param.Namespace, param.Name, param.Container, param.Image)
}

// Patch 模板（content 一般是 JSON Patch / StrategicMergePatch）
// 如果你传的是字符串，转成 []byte 再下发
func (s *Services) KubeUpdateDeploymentTemplate(ctx context.Context, cli *K8sClients, param *requests.KubeDeploymentUpdateRequest) (*appv1.Deployment, error) {
	return deployment.PatchDeployment(ctx, cli.Kube, param.Namespace, param.Name, []byte(param.Content))
}

// 回滚到指定 RS —— 和你的 DTO 对应
func (s *Services) KubeDeploymentRollback(ctx context.Context, cli *K8sClients, param *requests.KubeDeploymentRollbackRequest) (*appv1.Deployment, error) {
	return deployment.RollbackDeployment(ctx, cli.Kube, param.Name, param.Namespace, param.ReplicaSet)
}

// 重启 Deployment
func (s *Services) KubeDeploymentRestart(ctx context.Context, cli *K8sClients, param *requests.KubeDeploymentRestartRequest) error {
	return deployment.RestartDeployment(ctx, cli.Kube, param.Namespace, param.Name)
}

// 获取 Deployment 详情
func (s *Services) KubeDeploymentDetail(ctx context.Context, cli *K8sClients, param *requests.KubeDeploymentDetailRequest) (*appv1.Deployment, error) {
	return deployment.GetDeploymentDetail(ctx, cli.Kube, param.Name, param.Namespace)
}

// 创建 Deployment
func (s *Services) KubeDeploymentCreate(ctx context.Context, cli *K8sClients, req *requests.KubeDeploymentCreateRequest) (*appv1.Deployment, *corev1.Service, error) {
	// 1) 创建 Deployment
	dp, err := deployment.CreateDeployment(ctx, cli.Kube, req)
	if err != nil {
		// 用 %w 包装，便于上层中间件用 errors.Is / apierrors.* 继续识别
		return nil, nil, fmt.Errorf("create deployment failed: %w", err)
	}

	// 2) 按需创建 Service
	var svcObj *corev1.Service
	if req.IsCreateService {
		svcObj, err = deployment.CreateServiceFromDeployment(ctx, cli.Kube, req)
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				svcName := strings.TrimSpace(req.ServiceName)
				if svcName == "" {
					svcName = req.Name
				}
				if exist, gerr := cli.Kube.CoreV1().
					Services(req.Namespace).
					Get(ctx, svcName, metav1.GetOptions{}); gerr == nil {
					global.Logger.Infof("service %s/%s already exists, reuse it", req.Namespace, svcName)
					return dp, exist, nil
				}
				// Get 失败才回滚
			}

			// 真失败 → 回滚 Deployment（带传播策略，清理 RS/Pods）
			pol := metav1.DeletePropagationForeground // 或 Background
			if delErr := cli.Kube.AppsV1().
				Deployments(req.Namespace).
				Delete(ctx, dp.Name, metav1.DeleteOptions{PropagationPolicy: &pol}); delErr != nil {
				global.Logger.Errorf("rollback delete deployment %s/%s failed: %v", req.Namespace, dp.Name, delErr)
			}
			return nil, nil, fmt.Errorf("create service failed: %w", err)
		}

	}

	return dp, svcObj, nil
}

// 获取 Deployment 对应的 Pod 列表（原始 Pod 对象）
func (s *Services) KubeDeploymentGetPod(ctx context.Context, cli *K8sClients, param *requests.KubeCommonRequest) ([]corev1.Pod, error) {
	return deployment.GetPodByDeployment(ctx, cli.Kube, param.Namespace, param.Name)
}

// 获取 Deployment 对应的 事件
func (s *Services) KubeEventList(ctx context.Context, cli *K8sClients, param *requests.KubeEventListRequest) ([]models.EventItem, string, error) {
	return event.ListEvents(ctx, cli.Kube, param)
}

// 获取 Deployment 历史版本（ReplicaSet 列表）
func (s *Services) KubeDeploymentHistory(ctx context.Context, cli *K8sClients, param *requests.KubeCommonRequest) ([]appv1.ReplicaSet, error) {
	return deployment.GetDeploymentReplicaSet(ctx, cli.Kube, param.Namespace, param.Name)
}

// KubeDeploymentCreateFromYaml 从 YAML 创建 Deployment（支持多资源：PVC+Deployment）
func (s *Services) KubeDeploymentCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*appv1.Deployment, error) {
	// 1. 使用多资源解析器解析 YAML
	parser := common.NewMultiYAMLParser(yamlContent)
	if err := parser.Parse(); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证主资源（Deployment）是否存在且唯一
	mainResource, err := parser.ValidateMainResource(common.ResourceTypeDeployment)
	if err != nil {
		return nil, err
	}

	// 3. 按依赖顺序创建所有资源（PVC -> Service -> Deployment）
	created, err := common.CreateResourcesInOrder(ctx, cli.Kube, parser)
	if err != nil {
		return nil, fmt.Errorf("failed to create resources: %w", err)
	}

	// 4. 创建主资源 Deployment
	dp := &appv1.Deployment{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(mainResource.Raw.Object, dp); err != nil {
		return nil, fmt.Errorf("failed to convert to Deployment: %w", err)
	}

	// 确保 namespace 不为空（如果 YAML 中未指定，使用 ParsedResource 中的 namespace）
	if dp.Namespace == "" {
		dp.Namespace = mainResource.Namespace
	}

	createdDp, err := cli.Kube.AppsV1().Deployments(dp.Namespace).Create(ctx, dp, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("Deployment %q already exists in namespace %q", dp.Name, dp.Namespace)
		}
		return nil, fmt.Errorf("failed to create Deployment: %w", err)
	}

	// 5. 记录创建的所有资源
	global.Logger.Infof("Multi-resource YAML created successfully:")
	for resType, names := range created {
		for _, name := range names {
			global.Logger.Infof("  - %s: %s/%s", resType, dp.Namespace, name)
		}
	}
	global.Logger.Infof("  - Deployment: %s/%s", createdDp.Namespace, createdDp.Name)

	return createdDp, nil
}
