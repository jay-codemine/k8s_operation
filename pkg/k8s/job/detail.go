package job

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"go.uber.org/zap"
	"k8soperation/global"
)

// GetJobDetail 获取 Job 详情
func GetJobDetail(ctx context.Context, Kube kubernetes.Interface, name, namespace string) (*batchv1.Job, error) {
	job, err := Kube.BatchV1().
		Jobs(namespace).
		Get(ctx, name, metav1.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Error("job not found",
				zap.String("namespace", namespace),
				zap.String("name", name),
			)
			return nil, fmt.Errorf("job %s/%s not found", namespace, name)
		}

		global.Logger.Error("get job failed",
			zap.String("namespace", namespace),
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, err
	}

	return job, nil
}
