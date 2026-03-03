package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/utils"
	"strings"
	"time"
)

func (s *Services) K8sClusterCreate(ctx context.Context, param *requests.K8sClusterCreateRequest) error {
	return s.dao.KubeClusterCreate(ctx, param.ClusterName, param.ClusterVersion, param.KubeConfig)
}

func (s *Services) K8sClusterUpdate(ctx context.Context, param *requests.K8sClusterUpdateRequest) error {
	// Update：不再接收 status
	// kubeconfig 为空时：不覆盖 kube_config；同时也不强制改 status
	// kubeconfig 非空时：覆盖 kube_config，并把 status 置 Pending(2)，清空 last_error/last_check_at
	kubeConfigPlain := strings.TrimSpace(param.KubeConfig)
	hasKC := kubeConfigPlain != ""
	return s.dao.KubeClusterUpdate(ctx,
		param.ID,
		param.ClusterName,
		param.ClusterVersion,
		kubeConfigPlain,
		hasKC,
	)
}

func (s *Services) K8sClusterDelete(ctx context.Context, param *requests.K8sClusterDeleteRequest) error {
	return s.dao.KubeClusterDelete(ctx, param.ID)
}

func (s *Services) K8sClusterList(ctx context.Context, param *requests.K8sClusterListRequest,
) (list []*models.K8sCluster, total int64, err error) {
	return s.dao.KubeClusterList(ctx, param)
}

func (s *Services) K8sClusterInit(ctx context.Context, param *requests.K8sClusterInitRequest) (*K8sClients, error) {
	now := utils.NowUnix()

	// 1) 取 kubeconfig（DAO 返回明文）
	cfg, err := s.mustFromDB(ctx, param.ID)
	if err != nil {
		_ = s.dao.KubeClusterMarkCheckResult(ctx, param.ID, models.ClusterStatusBad, now, err.Error())
		return nil, errorcode.ErrorClusterInitFailed
	}

	// 2) build clients
	clients, err := s.buildClients(cfg)
	if err != nil {
		_ = s.dao.KubeClusterMarkCheckResult(ctx, param.ID, models.ClusterStatusBad, now, err.Error())
		return nil, errorcode.ErrorClusterInitFailed
	}

	// 3) 成功：写库 status=OK，清空 last_error
	if err := s.dao.KubeClusterMarkCheckResult(ctx, param.ID, models.ClusterStatusOK, now, ""); err != nil {
		// 写库失败不影响“已连通”的事实，但你可以选择返回错误
		global.Logger.Warn("mark check result failed", zap.Uint32("cluster_id", param.ID), zap.Error(err))
	}

	return clients, nil
}

// mustFromDB：DAO 已返回明文 kubeconfig（不要在 service 再 base64 decode）
func (s *Services) mustFromDB(ctx context.Context, clusterID uint32) (*rest.Config, error) {
	kc, err := s.dao.KubeClusterGetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	plain := strings.TrimSpace(kc.KubeConfig)
	if plain == "" {
		return nil, fmt.Errorf("empty kubeconfig in DB, cluster=%d", clusterID)
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(plain))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig failed: %w", err)
	}

	global.Logger.Info("init from DB kubeconfig (plain)", zap.Uint32("cluster_id", clusterID))
	return cfg, nil
}

func (s *Services) buildClients(cfg *rest.Config) (*K8sClients, error) {
	tuneRESTConfig(cfg)

	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create kube client: %w", err)
	}

	// metrics 非硬依赖：失败只告警
	var mc *metricsclient.Clientset
	if m, mErr := metricsclient.NewForConfig(cfg); mErr != nil {
		global.Logger.Warn("init MetricsClient failed", zap.Error(mErr))
	} else {
		mc = m
	}

	// 探测 events.k8s.io/v1
	supports := false
	if _, err := kube.Discovery().ServerResourcesForGroupVersion("events.k8s.io/v1"); err == nil {
		supports = true
	}

	return &K8sClients{
		Config:       cfg,
		Kube:         kube,
		Metrics:      mc,
		SupportsEvV1: supports,
	}, nil
}

func tuneRESTConfig(cfg *rest.Config) {
	cfg.UserAgent = "k8soperation/1.0"
	cfg.QPS = 50
	cfg.Burst = 100
	cfg.Timeout = 30 * time.Second
}

// BuildClientsFromKubeconfig 从 kubeconfig 字符串直接构建客户端（不经过数据库）
// 用于启动时本地 kubeconfig 回退场景
func BuildClientsFromKubeconfig(kubeConfigPlain string) (*K8sClients, error) {
	plain := strings.TrimSpace(kubeConfigPlain)
	if plain == "" {
		return nil, fmt.Errorf("empty kubeconfig")
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(plain))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig failed: %w", err)
	}

	tuneRESTConfig(cfg)

	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create kube client: %w", err)
	}

	// metrics 非硬依赖
	var mc *metricsclient.Clientset
	if m, mErr := metricsclient.NewForConfig(cfg); mErr != nil {
		global.Logger.Warn("init MetricsClient failed", zap.Error(mErr))
	} else {
		mc = m
	}

	// 探测 events.k8s.io/v1
	supports := false
	if _, err := kube.Discovery().ServerResourcesForGroupVersion("events.k8s.io/v1"); err == nil {
		supports = true
	}

	return &K8sClients{
		Config:       cfg,
		Kube:         kube,
		Metrics:      mc,
		SupportsEvV1: supports,
	}, nil
}
