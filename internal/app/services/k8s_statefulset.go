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
	"k8soperation/pkg/k8s/statefulset"
	"time"
)

func (s *Services) KubeStatefulSetCreate(ctx context.Context, cli *K8sClients, req *requests.KubeStatefulSetCreateRequest) (*appv1.StatefulSet, error) {
	sts, err := statefulset.CreateStatefulSet(ctx, cli.Kube, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("statefulset %q already exists in namespace %q", req.Name, req.Namespace)
		}
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("namespace %q not found: %w", req.Namespace, err)
		}
		// 其他错误：原样抛出（可保留一条 warn 日志，避免重复打 warn+error）
		global.Logger.Warnf("[StatefulSet] create failed %s/%s: %v", req.Namespace, req.Name, err)
		return nil, err
	}

	global.Logger.Infof("[StatefulSet] %s/%s created successfully", req.Namespace, req.Name)
	return sts, nil
}

func (s *Services) KubeStatefulSetCreateService(ctx context.Context, cli *K8sClients, req *requests.KubeStatefulSetCreateRequest,
) (*appv1.StatefulSet, *corev1.Service, error) {
	return statefulset.CreateStatefulSetWithService(ctx, cli.Kube, req)
}

func (s *Services) KubeStatefulSetList(
	ctx context.Context,
	cli *K8sClients,
	param *requests.KubeStatefulSetListRequest,
) ([]appv1.StatefulSet, int64, error) {

	sts, total, err := statefulset.GetStatefulSetList(
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

	return sts, total, nil
}

func (s *Services) KubeStatefulSetDetail(ctx context.Context, cli *K8sClients, param *requests.KubeStatefulSetDetailRequest) (*appv1.StatefulSet, error) {
	return statefulset.GetStatefulSetDetail(ctx, cli.Kube, param.Namespace, param.Name)
}

func (s *Services) KubeStatefulSetPatchReplicas(ctx context.Context, cli *K8sClients, param *requests.KubeStatefulSetScaleRequest) (*appv1.StatefulSet, error) {
	return statefulset.PatchScaleReplicasStatefulSet(ctx, cli.Kube, param.Namespace, param.Name, param.ScaleNum)
}

func (s *Services) KubeStatefulSetPatchImage(ctx context.Context, cli *K8sClients, param *requests.KubeStatefulSetUpdateImageRequest) (*appv1.StatefulSet, error) {
	return statefulset.PatchImageStatefulSet(ctx, cli.Kube, param.Namespace, param.Name, param.Container, param.Image)
}

func (s *Services) KubeStatefulSetRestart(ctx context.Context, cli *K8sClients, param *requests.KubeStatefulSetRestartRequest) (*appv1.StatefulSet, error) {
	return statefulset.RestartStatefulSet(ctx, cli.Kube, param.Namespace, param.Name)
}

func (s *Services) KubeStatefulSetDelete(ctx context.Context, cli *K8sClients, param *requests.KubeStatefulSetDeleteRequest) error {
	timeout := 30 * time.Second
	// 删除 StatefulSet（会自动触发 Pod 和 PVC 的删除，但保留 PV 和底层存储）
	// 自动删除逻辑：
	//   1. StatefulSet 的 PersistentVolumeClaimRetentionPolicy.WhenDeleted=Delete
	//   2. 删除 StatefulSet 时，Kubernetes 会自动删除关联的 PVC
	//   3. PVC 删除后，PV 会根据 ReclaimPolicy 保留（Retain）或删除（Delete）
	// 注意：确保 PV 的 ReclaimPolicy=Retain 以保留底层数据
	return statefulset.DeleteStatefulSet(ctx, cli.Kube, param.Name, param.Namespace, timeout)
}

func (s *Services) KubeStatefulSetDeleteService(ctx context.Context, cli *K8sClients, param *requests.KubeStatefulSetDeleteRequest) error {
	return statefulset.DeleteStatefulSetService(ctx, cli.Kube, param.Namespace, param.Name)
}

// KubeStatefulSetGetPod 获取 StatefulSet 对应的 Pod 列表
func (s *Services) KubeStatefulSetGetPod(ctx context.Context, cli *K8sClients, param *requests.KubeCommonRequest) ([]corev1.Pod, error) {
	return statefulset.GetPodByStatefulSet(ctx, cli.Kube, param.Namespace, param.Name)
}

// KubeStatefulSetHistory 获取 StatefulSet 历史版本（ControllerRevision 列表）
func (s *Services) KubeStatefulSetHistory(ctx context.Context, cli *K8sClients, param *requests.KubeCommonRequest) ([]statefulset.ControllerRevisionItem, error) {
	return statefulset.GetStatefulSetHistory(ctx, cli.Kube, param.Namespace, param.Name)
}

// KubeStatefulSetRollback 回滚 StatefulSet 到指定版本
func (s *Services) KubeStatefulSetRollback(ctx context.Context, cli *K8sClients, param *requests.KubeStatefulSetRollbackRequest) (*appv1.StatefulSet, error) {
	return statefulset.RollbackStatefulSet(ctx, cli.Kube, param.Name, param.Namespace, param.RevisionName)
}

// KubeStatefulSetCreateFromYaml 从 YAML 创建 StatefulSet（支持多资源：PVC+StatefulSet）
func (s *Services) KubeStatefulSetCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*appv1.StatefulSet, error) {
	// 1. 使用多资源解析器解析 YAML
	parser := common.NewMultiYAMLParser(yamlContent)
	if err := parser.Parse(); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证主资源（StatefulSet）是否存在且唯一
	mainResource, err := parser.ValidateMainResource(common.ResourceTypeStatefulSet)
	if err != nil {
		return nil, err
	}

	// 3. 按依赖顺序创建所有资源（PVC -> Service -> StatefulSet）
	created, err := common.CreateResourcesInOrder(ctx, cli.Kube, parser)
	if err != nil {
		return nil, fmt.Errorf("failed to create resources: %w", err)
	}

	// 4. 创建主资源 StatefulSet
	sts := &appv1.StatefulSet{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(mainResource.Raw.Object, sts); err != nil {
		return nil, fmt.Errorf("failed to convert to StatefulSet: %w", err)
	}

	// 确保 namespace 不为空（如果 YAML 中未指定，使用 ParsedResource 中的 namespace）
	if sts.Namespace == "" {
		sts.Namespace = mainResource.Namespace
	}

	createdSts, err := cli.Kube.AppsV1().StatefulSets(sts.Namespace).Create(ctx, sts, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("StatefulSet %q already exists in namespace %q", sts.Name, sts.Namespace)
		}
		return nil, fmt.Errorf("failed to create StatefulSet: %w", err)
	}

	// 5. 记录创建的所有资源
	global.Logger.Infof("Multi-resource YAML created successfully:")
	for resType, names := range created {
		for _, name := range names {
			global.Logger.Infof("  - %s: %s/%s", resType, sts.Namespace, name)
		}
	}
	global.Logger.Infof("  - StatefulSet: %s/%s", createdSts.Namespace, createdSts.Name)

	return createdSts, nil
}
