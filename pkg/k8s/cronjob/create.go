package cronjob

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

func CreateCronJob(ctx context.Context, Kube kubernetes.Interface, req *requests.KubeCronJobCreateRequest) (*batchv1.CronJob, error) {
	// 1. 构建 CronJob 对象（建议写个 BuildCronJobFromCreateReq 辅助函数）
	cronJob := BuildCronJobFromCreateReq(req)

	// 2. 调用 K8s API 创建 CronJob
	created, err := Kube.BatchV1().
		CronJobs(req.Namespace).
		Create(ctx, cronJob, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("cronjob %q already exists in namespace %q", req.Name, req.Namespace)
		}
		global.Logger.Warnf("create cronjob failed: %v", err)
		return nil, err
	}

	global.Logger.Infof("cronjob %q created successfully", created.Name)
	return created, nil
}
