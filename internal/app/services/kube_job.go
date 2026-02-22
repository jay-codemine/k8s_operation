package services

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/job"
	"sigs.k8s.io/yaml"
)

// KubeJobCreate 仅创建 Job（不创建 Service）
func (s *Services) KubeJobCreate(ctx context.Context, cli *K8sClients, req *requests.KubeJobCreateRequest) (*batchv1.Job, error) {
	jobObj, err := job.CreateJob(ctx, cli.Kube, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("job %s/%s already exists", req.Namespace, req.Name)
			return nil, fmt.Errorf("job %q already exists in namespace %q", req.Name, req.Namespace)
		}
		return nil, fmt.Errorf("create job failed: %w", err)
	}

	global.Logger.Infof("job %s/%s created successfully", req.Namespace, jobObj.Name)
	return jobObj, nil
}

// listJob 列出 Job
func (s *Services) KubeJobList(
	ctx context.Context,
	cli *K8sClients,
	param *requests.KubeJobListRequest,
) ([]batchv1.Job, int64, error) {

	// 构建标签选择器
	labelSelector := param.LabelSelector
	
	// 如果指定了 CronJob 名称，添加 CronJob 标签筛选
	if param.CronJob != "" {
		cronjobLabel := "app.k8soperation.io/cronjob=" + param.CronJob
		if labelSelector != "" {
			labelSelector += "," + cronjobLabel
		} else {
			labelSelector = cronjobLabel
		}
	}

	jobs, total, err := job.GetJobList(ctx, cli.Kube, param.Name, param.Namespace, param.Page, param.Limit, labelSelector)
	if err != nil {
		return nil, 0, err
	}
	return jobs, total, nil
}

// 获取 Job 详情
func (s *Services) KubeJobDetail(ctx context.Context, cli *K8sClients, param *requests.KubeJobDetailRequest) (*batchv1.Job, error) {
	return job.GetJobDetail(ctx, cli.Kube, param.Name, param.Namespace)
}

// KubeJobDelete 删除 Job
func (s *Services) KubeJobDelete(ctx context.Context, cli *K8sClients, param *requests.KubeJobDeleteRequest) error {
	return job.DeleteJob(ctx, cli.Kube, param.Name, param.Namespace)
}

// KubeSuspendJob 控制 Job 暂停或恢复
func (s *Services) KubeJobSuspend(ctx context.Context, cli *K8sClients, param *requests.KubeJobSuspendRequest) error {
	return job.SetJobSuspend(ctx, cli.Kube, param.Namespace, param.Name, param.Suspend)
}

// KubeJobRestart 重跑 Job（基于旧 Job 模板创建一个新名字的 Job）
func (s *Services) KubeJobRestart(ctx context.Context, cli *K8sClients, param *requests.KubeJobRestartRequest) (*batchv1.Job, error) {
	return job.RestartJob(ctx, cli.Kube, param.Namespace, param.Name)
}

// KubeJobUpdateImage 更新 Job 镜像
func (s *Services) KubeJobUpdateImage(ctx context.Context, cli *K8sClients, param *requests.KubeJobUpdateImageRequest) (*batchv1.Job, error) {
	return job.PatchJobImage(ctx, cli.Kube, param.Namespace, param.Name, param.Container, param.Image)
}

// KubeJobCreateFromYaml 从 YAML 创建 Job
func (s *Services) KubeJobCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*batchv1.Job, error) {
	// 1. 解析 YAML 到 Job 对象
	jobObj := &batchv1.Job{}
	if err := yaml.Unmarshal([]byte(yamlContent), jobObj); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证基本字段
	if jobObj.Kind != "Job" {
		return nil, fmt.Errorf("YAML kind must be 'Job', got: %s", jobObj.Kind)
	}
	if jobObj.APIVersion != "batch/v1" {
		return nil, fmt.Errorf("YAML apiVersion must be 'batch/v1', got: %s", jobObj.APIVersion)
	}
	if jobObj.Name == "" {
		return nil, fmt.Errorf("Job name is required")
	}
	if jobObj.Namespace == "" {
		jobObj.Namespace = "default"
	}

	// 3. 创建 Job
	createdJob, err := cli.Kube.BatchV1().Jobs(jobObj.Namespace).Create(ctx, jobObj, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("Job %q already exists in namespace %q", jobObj.Name, jobObj.Namespace)
		}
		return nil, fmt.Errorf("failed to create Job: %w", err)
	}

	global.Logger.Infof("Job %s/%s created from YAML successfully", createdJob.Namespace, createdJob.Name)
	return createdJob, nil
}

// KubeJobGetYaml 获取 Job 的 YAML
func (s *Services) KubeJobGetYaml(ctx context.Context, cli *K8sClients, namespace, name string) (string, error) {
	jobObj, err := cli.Kube.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get Job: %w", err)
	}

	// 清理不必要的字段
	jobObj.ManagedFields = nil
	jobObj.Status = batchv1.JobStatus{}

	yamlBytes, err := yaml.Marshal(jobObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Job to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// KubeJobApplyYaml 应用 YAML 更新 Job
func (s *Services) KubeJobApplyYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*batchv1.Job, error) {
	// 1. 解析 YAML 到 Job 对象
	jobObj := &batchv1.Job{}
	if err := yaml.Unmarshal([]byte(yamlContent), jobObj); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证基本字段
	if jobObj.Kind != "Job" {
		return nil, fmt.Errorf("YAML kind must be 'Job', got: %s", jobObj.Kind)
	}
	if jobObj.Name == "" {
		return nil, fmt.Errorf("Job name is required")
	}
	if jobObj.Namespace == "" {
		jobObj.Namespace = "default"
	}

	// 3. 更新 Job
	updatedJob, err := cli.Kube.BatchV1().Jobs(jobObj.Namespace).Update(ctx, jobObj, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update Job: %w", err)
	}

	global.Logger.Infof("Job %s/%s updated from YAML successfully", updatedJob.Namespace, updatedJob.Name)
	return updatedJob, nil
}
