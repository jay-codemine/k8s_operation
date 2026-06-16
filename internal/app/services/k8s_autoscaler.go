package services

import (
	"context"
	"fmt"
	"sync"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/hpa"
	"k8soperation/pkg/k8s/vpa"
)

/* ====================================================================
 * HPA 业务服务
 * ==================================================================== */

// KubeHPAList 列表（分页在内存做，HPA 一般数量有限）
func (s *Services) KubeHPAList(ctx context.Context, cli *K8sClients, param *requests.KubeHPAListRequest) ([]hpa.HPAItem, int64, error) {
	all, err := hpa.List(ctx, cli.Kube, param.Namespace, param.Name)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))

	page, limit := param.Page, param.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	start := (page - 1) * limit
	end := start + limit
	if start > len(all) {
		start = len(all)
	}
	if end > len(all) {
		end = len(all)
	}
	return hpa.BuildHPAItemList(all[start:end]), total, nil
}

// KubeHPADetail 详情
func (s *Services) KubeHPADetail(ctx context.Context, cli *K8sClients, param *requests.KubeHPADetailRequest) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	return hpa.Get(ctx, cli.Kube, param.Namespace, param.Name)
}

// KubeHPACreate 创建
func (s *Services) KubeHPACreate(ctx context.Context, cli *K8sClients, req *requests.KubeHPACreateRequest) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	return hpa.Create(ctx, cli.Kube, buildHPAOptions(req))
}

// KubeHPACreateFromYaml 使用 YAML 创建 HPA
func (s *Services) KubeHPACreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	return hpa.CreateFromYaml(ctx, cli.Kube, yamlContent)
}

// KubeHPAUpdate 更新
func (s *Services) KubeHPAUpdate(ctx context.Context, cli *K8sClients, req *requests.KubeHPAUpdateRequest) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	return hpa.Update(ctx, cli.Kube, buildHPAOptions(req))
}

// KubeHPADelete 删除
func (s *Services) KubeHPADelete(ctx context.Context, cli *K8sClients, namespace, name string) error {
	return hpa.Delete(ctx, cli.Kube, namespace, name)
}

// KubeHPAScale 单独修改 min/max（一次性扩缩容）
func (s *Services) KubeHPAScale(ctx context.Context, cli *K8sClients, req *requests.KubeHPAScaleRequest) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	if req.MinReplicas > req.MaxReplicas {
		return nil, fmt.Errorf("min_replicas (%d) must be <= max_replicas (%d)", req.MinReplicas, req.MaxReplicas)
	}
	return hpa.PatchMinMax(ctx, cli.Kube, req.Namespace, req.Name, req.MinReplicas, req.MaxReplicas)
}

// HPABatchResult 批量项结果
type HPABatchResult struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	Success         bool   `json:"success"`
	Message         string `json:"message,omitempty"`
	MinReplicas     int32  `json:"min_replicas"`
	MaxReplicas     int32  `json:"max_replicas"`
	CurrentReplicas int32  `json:"current_replicas"`
	DesiredReplicas int32  `json:"desired_replicas"`
}

// KubeHPABatchScale 批量扩缩容（618 促销场景），并发执行
func (s *Services) KubeHPABatchScale(ctx context.Context, cli *K8sClients, req *requests.KubeHPABatchScaleRequest) ([]HPABatchResult, int, int) {
	results := make([]HPABatchResult, len(req.Items))
	var wg sync.WaitGroup
	// 限并发，避免对 K8s API Server 压力过大
	sem := make(chan struct{}, 10)

	for i, item := range req.Items {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, it requests.KubeHPABatchItem) {
			defer wg.Done()
			defer func() { <-sem }()
			r := HPABatchResult{
				Namespace:   it.Namespace,
				Name:        it.Name,
				MinReplicas: it.MinReplicas,
				MaxReplicas: it.MaxReplicas,
			}
			if it.MinReplicas > it.MaxReplicas {
				r.Message = fmt.Sprintf("min(%d) > max(%d)", it.MinReplicas, it.MaxReplicas)
				results[idx] = r
				return
			}
			updated, err := hpa.PatchMinMax(ctx, cli.Kube, it.Namespace, it.Name, it.MinReplicas, it.MaxReplicas)
			if err != nil {
				r.Message = err.Error()
				results[idx] = r
				return
			}
			r.Success = true
			r.Message = "OK"
			r.CurrentReplicas = updated.Status.CurrentReplicas
			r.DesiredReplicas = updated.Status.DesiredReplicas
			results[idx] = r
		}(i, item)
	}
	wg.Wait()

	successCnt, failCnt := 0, 0
	for _, r := range results {
		if r.Success {
			successCnt++
		} else {
			failCnt++
		}
	}
	return results, successCnt, failCnt
}

// KubeHPABatchStatus 批量查询当前状态（统一数据：用于扩缩容后查看是否成功）
func (s *Services) KubeHPABatchStatus(ctx context.Context, cli *K8sClients, items []requests.KubeHPABatchItem) []HPABatchResult {
	results := make([]HPABatchResult, len(items))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	for i, item := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, it requests.KubeHPABatchItem) {
			defer wg.Done()
			defer func() { <-sem }()
			r := HPABatchResult{
				Namespace: it.Namespace,
				Name:      it.Name,
			}
			cur, err := hpa.Get(ctx, cli.Kube, it.Namespace, it.Name)
			if err != nil {
				r.Message = err.Error()
				results[idx] = r
				return
			}
			r.Success = true
			r.Message = "OK"
			if cur.Spec.MinReplicas != nil {
				r.MinReplicas = *cur.Spec.MinReplicas
			}
			r.MaxReplicas = cur.Spec.MaxReplicas
			r.CurrentReplicas = cur.Status.CurrentReplicas
			r.DesiredReplicas = cur.Status.DesiredReplicas
			results[idx] = r
		}(i, item)
	}
	wg.Wait()
	return results
}

func buildHPAOptions(r *requests.KubeHPACreateRequest) *hpa.CreateOptions {
	return &hpa.CreateOptions{
		Namespace:     r.Namespace,
		Name:          r.Name,
		TargetKind:    r.TargetKind,
		TargetName:    r.TargetName,
		TargetAPIVer:  r.TargetAPIVer,
		MinReplicas:   r.MinReplicas,
		MaxReplicas:   r.MaxReplicas,
		CPUTargetUtil: r.CPUTargetUtil,
		MemTargetUtil: r.MemTargetUtil,
		Labels:        r.Labels,
		Annotations:   r.Annotations,
		ScaleUpStab:   r.ScaleUpStab,
		ScaleDownStab: r.ScaleDownStab,
	}
}

/* ====================================================================
 * VPA 业务服务
 * ==================================================================== */

// KubeVPAAvailable 判断 VPA Operator 是否安装
func (s *Services) KubeVPAAvailable(ctx context.Context, cli *K8sClients) bool {
	return vpa.IsAvailable(ctx, cli.Dynamic)
}

// KubeVPAList 列表
func (s *Services) KubeVPAList(ctx context.Context, cli *K8sClients, param *requests.KubeVPAListRequest) ([]vpa.VPAItem, int64, error) {
	all, err := vpa.List(ctx, cli.Dynamic, param.Namespace, param.Name)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))
	page, limit := param.Page, param.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	start := (page - 1) * limit
	end := start + limit
	if start > len(all) {
		start = len(all)
	}
	if end > len(all) {
		end = len(all)
	}
	return vpa.BuildVPAItemList(all[start:end]), total, nil
}

// KubeVPADetail 详情
func (s *Services) KubeVPADetail(ctx context.Context, cli *K8sClients, param *requests.KubeVPADetailRequest) (*unstructured.Unstructured, error) {
	return vpa.Get(ctx, cli.Dynamic, param.Namespace, param.Name)
}

// KubeVPACreate 创建
func (s *Services) KubeVPACreate(ctx context.Context, cli *K8sClients, req *requests.KubeVPACreateRequest) (*unstructured.Unstructured, error) {
	return vpa.Create(ctx, cli.Dynamic, buildVPAOptions(req))
}

// KubeVPACreateFromYaml 使用 YAML 创建 VPA
func (s *Services) KubeVPACreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*unstructured.Unstructured, error) {
	return vpa.CreateFromYaml(ctx, cli.Dynamic, yamlContent)
}

// KubeVPAUpdate 更新
func (s *Services) KubeVPAUpdate(ctx context.Context, cli *K8sClients, req *requests.KubeVPAUpdateRequest) (*unstructured.Unstructured, error) {
	return vpa.Update(ctx, cli.Dynamic, buildVPAOptions(req))
}

// KubeVPADelete 删除
func (s *Services) KubeVPADelete(ctx context.Context, cli *K8sClients, namespace, name string) error {
	return vpa.Delete(ctx, cli.Dynamic, namespace, name)
}

func buildVPAOptions(r *requests.KubeVPACreateRequest) *vpa.CreateOptions {
	return &vpa.CreateOptions{
		Namespace:     r.Namespace,
		Name:          r.Name,
		TargetKind:    r.TargetKind,
		TargetName:    r.TargetName,
		TargetAPIVer:  r.TargetAPIVer,
		UpdateMode:    r.UpdateMode,
		ContainerName: r.ContainerName,
		ControlledRes: r.ControlledRes,
		MinAllowedCPU: r.MinAllowedCPU,
		MinAllowedMem: r.MinAllowedMem,
		MaxAllowedCPU: r.MaxAllowedCPU,
		MaxAllowedMem: r.MaxAllowedMem,
		Labels:        r.Labels,
	}
}
