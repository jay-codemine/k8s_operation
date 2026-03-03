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
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/common"
	"k8soperation/pkg/k8s/daemonset"
	"strings"
)

// 创建 DaemonSet（可选同时创建 Service）
func (s *Services) KubeDaemonSetCreate(ctx context.Context, cli *K8sClients, req *requests.KubeDaemonSetCreateRequest) (*appv1.DaemonSet, *corev1.Service, error) {
	// 1) 创建 DaemonSet
	ds, err := daemonset.CreateDaemonSet(ctx, cli.Kube, req)
	if err != nil {
		return nil, nil, fmt.Errorf("create daemonset failed: %w", err)
	}

	// 2) 按需创建 Service
	var svcObj *corev1.Service
	if req.IsCreateService {
		svcObj, err = daemonset.CreateServiceFromDaemonSet(ctx, cli.Kube, req)
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
					return ds, exist, nil
				}
				// Get 失败才继续回滚
			}

			// Service 真失败 → 回滚删除 DaemonSet（带传播策略，清理 Pods）
			pol := metav1.DeletePropagationForeground // 或 Background
			if delErr := cli.Kube.AppsV1().
				DaemonSets(req.Namespace).
				Delete(ctx, ds.Name, metav1.DeleteOptions{PropagationPolicy: &pol}); delErr != nil {
				global.Logger.Errorf("rollback delete daemonset %s/%s failed: %v", req.Namespace, ds.Name, delErr)
			}
			return nil, nil, fmt.Errorf("create service failed: %w", err)
		}
	}

	return ds, svcObj, nil
}

func (s *Services) KubeDaemonSetList(ctx context.Context, cli *K8sClients,
	param *requests.KubeDaemonSetListRequest,
) ([]appv1.DaemonSet, int64, error) {

	daemonsets, total, err := daemonset.GetDaemonSetList(
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

	return daemonsets, total, nil
}

func (s *Services) KubeDaemonSetDetail(ctx context.Context, cli *K8sClients, param *requests.KubeDaemonSetDetailRequest) (*appv1.DaemonSet, error) {
	return daemonset.GetDaemonSetDetail(ctx, cli.Kube, param.Name, param.Namespace)
}

// 删除 DaemonSet
func (s *Services) KubeDaemonSetDelete(ctx context.Context, cli *K8sClients, param *requests.KubeDaemonSetDeleteRequest) error {
	return daemonset.DeleteDaemonSet(ctx, cli.Kube, param.Name, param.Namespace)
}

// 删除 DaemonSet 对应的 Service（如果有）
func (s *Services) KubeDaemonSetDeleteService(ctx context.Context, cli *K8sClients, param *requests.KubeDaemonSetDeleteRequest) error {
	return daemonset.DeleteDaemonSetService(ctx, cli.Kube, param.Name, param.Namespace)
}

// 更新 DaemonSet 的镜像
func (s *Services) KubeDaemonSetUpdateImage(ctx context.Context, cli *K8sClients, param *requests.KubeDaemonSetUpdateImageRequest) (*appv1.DaemonSet, error) {
	return daemonset.PatchUpdateDaemonSetImage(ctx, cli.Kube, param.Namespace, param.Name, param.Container, param.Image)
}

// 重启 DaemonSet
func (s *Services) KubeDaemonSetRestart(ctx context.Context, cli *K8sClients, param *requests.KubeDaemonSetRestartRequest) error {
	return daemonset.RestartDaemonSet(ctx, cli.Kube, param.Namespace, param.Name)
}

// 回滚到指定版本
func (s *Services) KubeDaemonSetRollback(ctx context.Context, cli *K8sClients, param *requests.KubeDaemonSetRollbackRequest) (*appv1.DaemonSet, error) {
	return daemonset.RollbackDaemonSet(ctx, cli.Kube, param.Name, param.Namespace, param.RevisionName)
}

// 获取 DaemonSet 关联的 Pods
func (s *Services) KubeDaemonSetPods(ctx context.Context, cli *K8sClients, param *requests.KubeDaemonSetPodsRequest) ([]corev1.Pod, error) {
	return daemonset.GetDaemonSetPods(ctx, cli.Kube, param.Namespace, param.Name)
}

// 获取 DaemonSet 历史版本（ControllerRevision）
func (s *Services) KubeDaemonSetHistory(ctx context.Context, cli *K8sClients, param *requests.KubeDaemonSetHistoryRequest) ([]daemonset.RevisionItem, error) {
	return daemonset.GetDaemonSetHistory(ctx, cli.Kube, param.Namespace, param.Name)
}

// KubeDaemonSetCreateFromYaml 从 YAML 创建 DaemonSet（支持多资源：ConfigMap/Secret+DaemonSet）
func (s *Services) KubeDaemonSetCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*appv1.DaemonSet, error) {
	// 1. 使用多资源解析器解析 YAML
	parser := common.NewMultiYAMLParser(yamlContent)
	if err := parser.Parse(); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证主资源（DaemonSet）是否存在且唯一
	mainResource, err := parser.ValidateMainResource(common.ResourceTypeDaemonSet)
	if err != nil {
		return nil, err
	}

	// 3. 按依赖顺序创建所有资源（ConfigMap/Secret/PVC -> Service -> DaemonSet）
	created, err := common.CreateResourcesInOrder(ctx, cli.Kube, parser)
	if err != nil {
		return nil, fmt.Errorf("failed to create resources: %w", err)
	}

	// 4. 创建主资源 DaemonSet
	ds := &appv1.DaemonSet{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(mainResource.Raw.Object, ds); err != nil {
		return nil, fmt.Errorf("failed to convert to DaemonSet: %w", err)
	}

	// 确保 namespace 不为空（如果 YAML 中未指定，使用 ParsedResource 中的 namespace）
	if ds.Namespace == "" {
		ds.Namespace = mainResource.Namespace
	}

	createdDs, err := cli.Kube.AppsV1().DaemonSets(ds.Namespace).Create(ctx, ds, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("DaemonSet %q already exists in namespace %q", ds.Name, ds.Namespace)
		}
		return nil, fmt.Errorf("failed to create DaemonSet: %w", err)
	}

	// 5. 记录创建的所有资源
	global.Logger.Infof("Multi-resource YAML created successfully:")
	for resType, names := range created {
		for _, name := range names {
			global.Logger.Infof("  - %s: %s/%s", resType, ds.Namespace, name)
		}
	}
	global.Logger.Infof("  - DaemonSet: %s/%s", createdDs.Namespace, createdDs.Name)

	return createdDs, nil
}
