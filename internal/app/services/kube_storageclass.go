package services

import (
	"context"
	"fmt"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/pkg/k8s/storageclass"
)

func (s *Services) KubeCreateStorageClass(ctx context.Context, cli *K8sClients, req *requests.KubeStorageClassCreateRequest) (*storagev1.StorageClass, error) {
	sc, err := storageclass.CreateStorageClass(ctx, cli.Kube, req)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			global.Logger.Warnf("storageclass %s already exists", req.Name)
			return nil, fmt.Errorf("storageclass %q already exists", req.Name)
		}
		return nil, fmt.Errorf("create storageclass failed: %w", err)
	}
	global.Logger.Infof("storageclass %q created successfully", sc.Name)
	return sc, nil
}

func (s *Services) KubeStorageClassList(ctx context.Context, cli *K8sClients, param *requests.KubeStorageClassListRequest) ([]storagev1.StorageClass, int, error) {
	return storageclass.GetStorageClassList(ctx, cli.Kube, param.Name, param.Page, param.Limit)
}

func (s *Services) KubeStorageClassDetail(ctx context.Context, cli *K8sClients, param *requests.KubeStorageClassDetailRequest) (*storagev1.StorageClass, error) {
	return storageclass.GetStorageClassDetail(ctx, cli.Kube, param.Name)
}

func (s *Services) KubeStorageClassDelete(ctx context.Context, cli *K8sClients, param *requests.KubeStorageClassDeleteRequest) error {
	return storageclass.DeleteStorageClass(ctx, cli.Kube, param.Name)
}

// YAML 相关方法
func (s *Services) KubeStorageClassGetYaml(ctx context.Context, cli *K8sClients, param *requests.KubeStorageClassDetailRequest) (string, error) {
	return storageclass.GetStorageClassYaml(ctx, cli.Kube, param.Name)
}

func (s *Services) KubeStorageClassCreateFromYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*storagev1.StorageClass, error) {
	return storageclass.CreateStorageClassFromYaml(ctx, cli.Kube, yamlContent)
}

func (s *Services) KubeStorageClassApplyYaml(ctx context.Context, cli *K8sClients, yamlContent string) (*storagev1.StorageClass, error) {
	return storageclass.ApplyStorageClassYaml(ctx, cli.Kube, yamlContent)
}
