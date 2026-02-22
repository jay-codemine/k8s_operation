package storageclass

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8soperation/global"
)

func GetStorageClassDetail(ctx context.Context, Kube kubernetes.Interface, name string) (*storagev1.StorageClass, error) {
	sc, err := Kube.StorageV1().
		StorageClasses().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			global.Logger.Error("storageclass not found", zap.String("name", name))
			return nil, fmt.Errorf("storageclass %s not found", name)
		}
		global.Logger.Error("get storageclass failed", zap.String("name", name), zap.Error(err))
		return nil, err
	}
	return sc, nil
}
