package services

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/cronjob"
	"sigs.k8s.io/yaml"
)

// KubeCronJobCreate 创建 CronJob
func (s *Services) KubeCronJobCreate(ctx context.Context, cli *K8sClients, req *requests.KubeCronJobCreateRequest) (*batchv1.CronJob, error) {
	cronJobObj, err := cronjob.CreateCronJob(ctx, cli.Kube, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("cronjob %s/%s already exists", req.Namespace, req.Name)
			return nil, fmt.Errorf("cronjob %q already exists in namespace %q", req.Name, req.Namespace)
		}
		return nil, fmt.Errorf("create cronjob failed: %w", err)
	}

	global.Logger.Infof("cronjob %s/%s created successfully", req.Namespace, cronJobObj.Name)
	return cronJobObj, nil
}

// KubeCronJobList 获取 CronJob 列表
func (s *Services) KubeCronJobList(
	ctx context.Context,
	cli *K8sClients,
	param *requests.KubeCronJobListRequest,
) (*cronjob.CronJobListResult, error) {

	result, err := cronjob.GetCronJobList(ctx, cli.Kube, param.Name, param.Namespace, param.Page, param.Limit, param.LabelSelector)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// KubeCronJobDetail 获取 CronJob 详情
func (s *Services) KubeCronJobDetail(ctx context.Context, cli *K8sClients, param *requests.KubeCronJobDetailRequest) (
	*batchv1.CronJob, []batchv1.Job, error) {
	return cronjob.GetCronJobDetail(ctx, cli.Kube, param.Name, param.Namespace)
}

// KubeCronJobDelete 删除 cronjob 资源
func (s *Services) KubeCronJobDelete(ctx context.Context, cli *K8sClients, param *requests.KubeCronJobDeleteRequest) error {
	return cronjob.DeleteCronJob(ctx, cli.Kube, param.Name, param.Namespace)
}

// KubeCronJobSuspend 暂停和恢复
func (s *Services) KubeCronJobSuspend(ctx context.Context, cli *K8sClients, param *requests.KubeCronJobSuspendRequest) error {
	return cronjob.SetCronJobSuspend(ctx, cli.Kube, param.Namespace, param.Name, param.Suspend)
}

// KubeCronJobCreateFromYaml 从 YAML 创建 CronJob
func (s *Services) KubeCronJobCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*batchv1.CronJob, error) {
	// 1. 解析 YAML 到 CronJob 对象
	cronJobObj := &batchv1.CronJob{}
	if err := yaml.Unmarshal([]byte(yamlContent), cronJobObj); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证基本字段
	if cronJobObj.Kind != "CronJob" {
		return nil, fmt.Errorf("YAML kind must be 'CronJob', got: %s", cronJobObj.Kind)
	}
	if cronJobObj.APIVersion != "batch/v1" {
		return nil, fmt.Errorf("YAML apiVersion must be 'batch/v1', got: %s", cronJobObj.APIVersion)
	}
	if cronJobObj.Name == "" {
		return nil, fmt.Errorf("CronJob name is required")
	}
	if cronJobObj.Namespace == "" {
		cronJobObj.Namespace = "default"
	}

	// 3. 创建 CronJob
	createdCronJob, err := cli.Kube.BatchV1().CronJobs(cronJobObj.Namespace).Create(ctx, cronJobObj, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("CronJob %q already exists in namespace %q", cronJobObj.Name, cronJobObj.Namespace)
		}
		return nil, fmt.Errorf("failed to create CronJob: %w", err)
	}

	global.Logger.Infof("CronJob %s/%s created from YAML successfully", createdCronJob.Namespace, createdCronJob.Name)
	return createdCronJob, nil
}

// KubeCronJobUpdateFromYaml 从 YAML 更新 CronJob
func (s *Services) KubeCronJobUpdateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*batchv1.CronJob, error) {
	// 1. 解析 YAML 到 CronJob 对象
	cronJobObj := &batchv1.CronJob{}
	if err := yaml.Unmarshal([]byte(yamlContent), cronJobObj); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 2. 验证基本字段
	if cronJobObj.Kind != "CronJob" {
		return nil, fmt.Errorf("YAML kind must be 'CronJob', got: %s", cronJobObj.Kind)
	}
	if cronJobObj.APIVersion != "batch/v1" {
		return nil, fmt.Errorf("YAML apiVersion must be 'batch/v1', got: %s", cronJobObj.APIVersion)
	}
	if cronJobObj.Name == "" {
		return nil, fmt.Errorf("CronJob name is required")
	}
	if cronJobObj.Namespace == "" {
		cronJobObj.Namespace = "default"
	}

	// 3. 获取现有 CronJob
	existing, err := cli.Kube.BatchV1().CronJobs(cronJobObj.Namespace).Get(ctx, cronJobObj.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("CronJob %q not found in namespace %q", cronJobObj.Name, cronJobObj.Namespace)
		}
		return nil, fmt.Errorf("failed to get CronJob: %w", err)
	}

	// 4. 保留 resourceVersion 以支持乐观锁更新
	cronJobObj.ResourceVersion = existing.ResourceVersion

	// 5. 更新 CronJob
	updatedCronJob, err := cli.Kube.BatchV1().CronJobs(cronJobObj.Namespace).Update(ctx, cronJobObj, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update CronJob: %w", err)
	}

	global.Logger.Infof("CronJob %s/%s updated from YAML successfully", updatedCronJob.Namespace, updatedCronJob.Name)
	return updatedCronJob, nil
}

// KubeCronJobTrigger 手动触发 CronJob（立即创建一个 Job）
func (s *Services) KubeCronJobTrigger(ctx context.Context, cli *K8sClients, param *requests.KubeCronJobTriggerRequest) (*batchv1.Job, error) {
	// 1. 获取 CronJob
	cj, err := cli.Kube.BatchV1().CronJobs(param.Namespace).Get(ctx, param.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get CronJob: %w", err)
	}

	// 2. 系统标签（继承 CronJob 标签 + 添加触发标识）
	jobLabels := map[string]string{
		"app.k8soperation.io/name":       cj.Name,
		"app.k8soperation.io/managed-by": "k8soperation",
		"app.k8soperation.io/cronjob":    cj.Name,
		"app.k8soperation.io/created-by": "cronjob-trigger",
	}

	// 合并 CronJob JobTemplate 中的自定义标签
	for k, v := range cj.Spec.JobTemplate.Labels {
		if _, exists := jobLabels[k]; !exists {
			jobLabels[k] = v
		}
	}

	// 3. 从 CronJob 的 JobTemplate 创建 Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cj.Name + "-manual-",
			Namespace:    cj.Namespace,
			Labels:       jobLabels,
			Annotations: map[string]string{
				"cronjob.kubernetes.io/instantiate": "manual",
				"app.k8soperation.io/trigger-time":  metav1.Now().Format("2006-01-02 15:04:05"),
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "batch/v1",
					Kind:       "CronJob",
					Name:       cj.Name,
					UID:        cj.UID,
				},
			},
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}

	// 确保 Pod 模板也继承标签
	if job.Spec.Template.Labels == nil {
		job.Spec.Template.Labels = make(map[string]string)
	}
	for k, v := range jobLabels {
		job.Spec.Template.Labels[k] = v
	}

	// 4. 创建 Job
	createdJob, err := cli.Kube.BatchV1().Jobs(param.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Job: %w", err)
	}

	global.Logger.Infof("CronJob %s/%s triggered manually, created Job: %s", cj.Namespace, cj.Name, createdJob.Name)
	return createdJob, nil
}
