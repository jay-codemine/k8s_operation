package dao

import (
	"context"
	"errors"
	"fmt"
	"k8soperation/internal/app/requests"

	"gorm.io/gorm"

	"k8soperation/internal/app/models"
	"k8soperation/pkg/utils"
)

// K8sClusterCreate：入参 kubeConfigPlain 是明文，落库转 base64
func (d *Dao) KubeClusterCreate(ctx context.Context, clusterName, clusterVersion, kubeConfigPlain string) error {
	kubeB64, err := utils.EncodeKubeconfigBase64(kubeConfigPlain)
	if err != nil {
		return err
	}

	now := utils.NowUnix()
	kc := models.K8sCluster{
		ClusterName:    clusterName,
		ClusterVersion: clusterVersion,
		KubeConfig:     kubeB64,

		// 新建默认 Pending
		Status:      models.ClusterStatusPending,
		LastCheckAt: 0,
		LastError:   "",

		CreatedAt:  now,
		ModifiedAt: now,
		IsDel:      0,
	}
	return kc.Create(d.db.WithContext(ctx))
}

func (d *Dao) KubeClusterGetByName(ctx context.Context, clusterName string) (*models.K8sCluster, error) {
	kc := models.K8sCluster{ClusterName: clusterName}
	return kc.GetByName(d.db.WithContext(ctx))
}

func (d *Dao) KubeClusterMarkCheckResult(ctx context.Context, id uint32, status uint8, checkAt uint64, lastErr string) error {
	values := map[string]interface{}{
		"status":        status,
		"last_check_at": checkAt,
		"last_error":    lastErr,
		"modified_at":   utils.NowUnix(),
	}
	kc := models.K8sCluster{ID: id}
	return kc.Update(d.db.WithContext(ctx), values)
}

// Update：不再接收 status
// hasKC 表示本次是否真的更新 kubeconfig
func (d *Dao) KubeClusterUpdate(
	ctx context.Context,
	id uint32,
	clusterName, clusterVersion, kubeConfigPlain string,
	hasKC bool,
) error {
	now := utils.NowUnix()

	values := map[string]interface{}{
		"cluster_name":    clusterName,
		"cluster_version": clusterVersion,
		"modified_at":     now,
	}

	if hasKC {
		kubeB64, err := utils.EncodeKubeconfigBase64(kubeConfigPlain)
		if err != nil {
			return err
		}
		values["kube_config"] = kubeB64

		// 只要 kubeconfig 变了，就回到 Pending（等待 init 探测）
		values["status"] = models.ClusterStatusPending
		values["last_check_at"] = uint32(0)
		values["last_error"] = ""
	}

	kc := models.K8sCluster{ID: id}
	return kc.Update(d.db.WithContext(ctx), values)
}

func (d *Dao) KubeClusterList(ctx context.Context, req *requests.K8sClusterListRequest) (list []*models.K8sCluster, total int64, err error) {
	kc := &models.K8sCluster{ClusterName: req.ClusterName}
	return kc.List(d.db.WithContext(ctx), req.Page, req.Limit)
}

func (d *Dao) KubeClusterDelete(ctx context.Context, clusterId uint32) error {
	kc := &models.K8sCluster{ID: clusterId}
	return kc.Delete(d.db.WithContext(ctx))
}

func (d *Dao) UpdateClusterHealth(ctx context.Context, clusterID uint32, status uint8, lastErr string) error {
	kc := &models.K8sCluster{ID: clusterID}
	return kc.UpdateHealth(d.db.WithContext(ctx), status, lastErr)
}

// KubeClusterGetByID：返回明文 kubeconfig（DAO 负责解码）
func (d *Dao) KubeClusterGetByID(ctx context.Context, id uint32) (*K8sClusterWithPlain, error) {
	var m models.K8sCluster
	cluster, err := m.GetByID(d.db.WithContext(ctx), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("cluster not found: %d", id)
		}
		return nil, err
	}

	plain, err := utils.DecodeKubeconfigBase64(cluster.KubeConfig)
	if err != nil {
		return nil, err
	}

	return &K8sClusterWithPlain{
		ID:          cluster.ID,
		ClusterName: cluster.ClusterName,
		KubeConfig:  plain,
		ClusterVer:  cluster.ClusterVersion,
		Status:      cluster.Status,
		ModifiedAt:  cluster.ModifiedAt,
	}, nil
}
