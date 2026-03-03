package services

import (
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/common"
	"k8soperation/pkg/k8s/pod"
)

// KubePodEvict 驱逐单个 Pod
func (s *Services) KubePodEvict(ctx context.Context, cli *K8sClients, param *requests.KubePodEvictRequest) error {
	// 默认：-1 表示不显式指定，让 K8s 用 Pod 自己的 terminationGracePeriodSeconds
	graceSeconds := int64(-1)
	if global.PodSetting != nil {
		if global.PodSetting.Eviction.DefaultGraceSeconds >= 0 {
			graceSeconds = global.PodSetting.Eviction.DefaultGraceSeconds
		}
	}
	return pod.EvictOnePod(ctx, cli.Kube, param.Namespace, param.PodName, graceSeconds)
}

func (s *Services) KubePodCreate(
	ctx context.Context,
	cli *K8sClients,
	param *requests.KubePodCreateRequest,
) (*corev1.Pod, error) {

	return pod.CreatePod(ctx, cli.Kube, param)
}

// KubePodCreateFromYaml 从 YAML 创建 Pod（支持多资源：ConfigMap/Secret/PVC+Pod）
func (s *Services) KubePodCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*corev1.Pod, error) {
	// 1. 使用多资源解析器解析 YAML
	parser := common.NewMultiYAMLParser(yamlContent)
	if err := parser.Parse(); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证主资源（Pod）是否存在且唯一
	mainResource, err := parser.ValidateMainResource(common.ResourceTypePod)
	if err != nil {
		return nil, err
	}

	// 3. 按依赖顺序创建所有资源（ConfigMap/Secret/PVC -> Pod）
	created, err := common.CreateResourcesInOrder(ctx, cli.Kube, parser)
	if err != nil {
		return nil, fmt.Errorf("failed to create resources: %w", err)
	}

	// 4. 创建主资源 Pod
	podObj := &corev1.Pod{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(mainResource.Raw.Object, podObj); err != nil {
		return nil, fmt.Errorf("failed to convert to Pod: %w", err)
	}

	// 确保 namespace 不为空（如果 YAML 中未指定，使用 ParsedResource 中的 namespace）
	if podObj.Namespace == "" {
		podObj.Namespace = mainResource.Namespace
	}

	createdPod, err := cli.Kube.CoreV1().Pods(podObj.Namespace).Create(ctx, podObj, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("Pod %q already exists in namespace %q", podObj.Name, podObj.Namespace)
		}
		return nil, fmt.Errorf("failed to create Pod: %w", err)
	}

	// 5. 记录创建的所有资源
	global.Logger.Infof("Multi-resource YAML created successfully:")
	for resType, names := range created {
		for _, name := range names {
			global.Logger.Infof("  - %s: %s/%s", resType, podObj.Namespace, name)
		}
	}
	global.Logger.Infof("  - Pod: %s/%s", createdPod.Namespace, createdPod.Name)

	return createdPod, nil
}

// PodList 获取Pod列表
func (s *Services) KubePodList(ctx context.Context, cli *K8sClients,
	param *requests.KubePodListRequest,
) ([]corev1.Pod, int64, error) {

	pods, total, err := pod.GetPodList(ctx, cli.Kube, param.Name, param.Namespace, param.Page, param.Limit)
	if err != nil {
		return nil, 0, err
	}
	return pods, total, nil
}

// PodDelete 从Pod列表中删除Pod
// services/pod_service.go
func (s *Services) KubePodDelete(ctx context.Context, cli *K8sClients, param *requests.KubePodDeleteRequest) error {
	if err := pod.DeletePod(ctx, cli.Kube, param.Namespace, param.Name, param.GraceSeconds, param.Force); err != nil {
		global.Logger.Errorf("删除 Pod 失败 ns=%s name=%s : %v", param.Namespace, param.Name, err)
		return err
	}

	var g int64 = -1
	if param.GraceSeconds != nil {
		g = *param.GraceSeconds
	}
	global.Logger.Infof("删除 Pod 已提交 ns=%s name=%s force=%v grace=%d", param.Namespace, param.Name, param.Force, g)
	return nil
}

// KubePodUpdate PodUpdate 更新Pod
func (s *Services) KubePodUpdate(ctx context.Context, cli *K8sClients, param *requests.KubePodUpdateRequest) error {
	if err := pod.UpdatePod(ctx, cli.Kube, param.Namespace, param.Name, param.Content); err != nil {
		global.Logger.Errorf("UpdatePod error: %v", err)
		return err
	}
	global.Logger.Infof("UpdatePod success")
	return nil
}

func (s *Services) PatchPodImage(ctx context.Context, cli *K8sClients, param *requests.PatchPodImageRequest) error {
	if err := pod.PatchPodImage(
		ctx,
		cli.Kube,
		param.Namespace,
		param.Name,
		param.Container,
		param.NewImage,
	); err != nil {
		global.Logger.Errorf("PatchPodImage error: %v", err)
		return err
	}

	global.Logger.Infof(
		"PatchPodImage success: ns=%s pod=%s container=%s image=%s",
		param.Namespace, param.Name, param.Container, param.NewImage,
	)
	return nil
}

// KubePodDetail PodDetail 获取单个 Pod
func (s *Services) KubePodDetail(ctx context.Context, cli *K8sClients, param *requests.KubePodDetailRequest) (*corev1.Pod, error) {
	p, err := pod.GetPodDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)
	return p, nil
}

// GetContainerNames 获取容器名称列表
func (s *Services) GetContainerNames(ctx context.Context, cli *K8sClients, param *requests.KubePodDetailRequest) ([]string, error) {
	// 先获取Pod详情
	p, err := pod.GetPodDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)

	// 获取容器名称列表
	containersNames := common.GetContainerNames(&p.Spec)
	return containersNames, nil
}

// GetInitContainerNames 获取Init容器名称列表
func (s *Services) GetInitContainerNames(ctx context.Context, cli *K8sClients, param *requests.KubeCommonRequest) ([]string, error) {
	// 先获取Pod详情
	p, err := pod.GetPodDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)

	// 获取Init容器
	initContainerNames := common.GetInitContainerNames(&p.Spec)
	return initContainerNames, nil
}

// GetContainerImages 获取容器镜像名称列表
func (s *Services) GetContainerImages(ctx context.Context, cli *K8sClients, param *requests.KubePodDetailRequest) ([]string, error) {
	// 先获取Pod详情
	p, err := pod.GetPodDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)

	// 获取容器镜像名称列表
	containerImages := common.GetContainerImages(&p.Spec)
	return containerImages, nil
}

// GetInitContainerImages 获取Init容器镜像名称列表
func (s *Services) GetInitContainerImages(ctx context.Context, cli *K8sClients, param *requests.KubeCommonRequest) ([]string, error) {
	// 先获取Pod详情
	p, err := pod.GetPodDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}
	global.Logger.Infof("GetPodDetail success: %s/%s", param.Namespace, param.Name)

	// 获取Init容器镜像名称列表
	initContainerImages := common.GetInitContainerImages(&p.Spec)
	return initContainerImages, nil
}

// 获取所有容器名称（常规 + Init）
func (s *Services) KubePodAllContainerNames(ctx context.Context, cli *K8sClients, param *requests.KubePodDetailRequest) ([]string, error) {
	// 1. 获取 Pod 对象
	p, err := pod.GetPodDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}

	// 2. 从 PodSpec 中提取容器名称
	names := common.GetAllContainerNames(&p.Spec)

	global.Logger.Infof("GetAllContainerNames success: %s/%s -> %v", param.Namespace, param.Name, names)
	return names, nil
}

// 获取所有容器镜像（常规 + Init）
func (s *Services) KubePodAllContainerImages(ctx context.Context, cli *K8sClients, param *requests.KubePodDetailRequest) ([]string, error) {
	// 1. 获取 Pod 对象
	p, err := pod.GetPodDetail(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Errorf("GetPodDetail error: %v", err)
		return nil, err
	}

	// 2. 从 PodSpec 中提取容器镜像
	images := common.GetAllContainerImages(&p.Spec)

	global.Logger.Infof("GetAllContainerImages success: %s/%s -> %v", param.Namespace, param.Name, images)
	return images, nil
}

func (s *Services) KubePodLog(ctx context.Context, cli *K8sClients, name, namespace, container string,
	tail *int64,
) (string, error) {
	return pod.GetPodLog(ctx, cli.Kube, name, namespace, container, tail, false)
}

func (s *Services) KubePodLogStream(
	ctx context.Context,
	cli *K8sClients,
	name, namespace, container string,
	tail *int64,
) (io.ReadCloser, error) {
	return pod.GetPodLogStream(ctx, cli.Kube, name, namespace, container, tail)
}

// KubePodMetrics 获取单个 Pod 的资源使用情况
func (s *Services) KubePodMetrics(ctx context.Context, cli *K8sClients, namespace, podName string) (*pod.PodMetrics, error) {
	return pod.GetPodMetrics(ctx, cli.Metrics, namespace, podName)
}

// KubePodsMetrics 批量获取多个 Pod 的资源使用情况
func (s *Services) KubePodsMetrics(ctx context.Context, cli *K8sClients, namespace string) (map[string]*pod.PodMetrics, error) {
	return pod.GetPodsMetrics(ctx, cli.Metrics, namespace)
}

// KubePodPatchLabels 修改 Pod 标签
func (s *Services) KubePodPatchLabels(ctx context.Context, cli *K8sClients, param *requests.KubePodLabelPatchRequest) error {
	err := pod.PatchLabels(ctx, cli.Kube, param.Namespace, param.Name, param.Add, param.Remove)
	if err != nil {
		global.Logger.Errorf("patch Pod labels failed: ns=%s name=%s err=%v", param.Namespace, param.Name, err)
		return err
	}
	global.Logger.Infof("patch Pod labels success: ns=%s name=%s", param.Namespace, param.Name)
	return nil
}
