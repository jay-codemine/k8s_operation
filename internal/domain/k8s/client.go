package k8s

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

// K8sClients 聚合一套与某个集群绑定的客户端
type K8sClients struct {
	Config       *rest.Config
	Kube         *kubernetes.Clientset
	Dynamic      dynamic.Interface
	Metrics      *metricsclient.Clientset
	SupportsEvV1 bool
}

// TuneRESTConfig 调优 REST 配置
func TuneRESTConfig(cfg *rest.Config, connTimeoutSec int) {
	cfg.UserAgent = "k8soperation/1.0"
	cfg.QPS = 50
	cfg.Burst = 100

	timeout := 30
	if connTimeoutSec > 0 {
		timeout = connTimeoutSec
	}
	cfg.Timeout = time.Duration(timeout) * time.Second

	cfg.TLSClientConfig.Insecure = true
	cfg.TLSClientConfig.CAData = nil
	cfg.TLSClientConfig.CAFile = ""
}

// BuildClients 从 REST 配置构建 K8s 客户端集
func BuildClients(cfg *rest.Config, connTimeoutSec int) (*K8sClients, error) {
	TuneRESTConfig(cfg, connTimeoutSec)

	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create kube client: %w", err)
	}

	if _, err := kube.Discovery().ServerVersion(); err != nil {
		return nil, fmt.Errorf("API Server connectivity check failed: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create dynamic client: %w", err)
	}

	var mc *metricsclient.Clientset
	if m, mErr := metricsclient.NewForConfig(cfg); mErr == nil {
		mc = m
	}

	supports := false
	if _, err := kube.Discovery().ServerResourcesForGroupVersion("events.k8s.io/v1"); err == nil {
		supports = true
	}

	return &K8sClients{
		Config:       cfg,
		Kube:         kube,
		Dynamic:      dynClient,
		Metrics:      mc,
		SupportsEvV1: supports,
	}, nil
}

// BuildClientsFromKubeconfig 从 kubeconfig 字符串构建客户端（不经过数据库）
func BuildClientsFromKubeconfig(kubeConfigPlain string, connTimeoutSec int) (*K8sClients, error) {
	plain := strings.TrimSpace(kubeConfigPlain)
	if plain == "" {
		return nil, fmt.Errorf("empty kubeconfig")
	}

	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(plain))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig failed: %w", err)
	}

	return BuildClients(cfg, connTimeoutSec)
}
