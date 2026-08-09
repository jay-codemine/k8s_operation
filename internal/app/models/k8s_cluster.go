package models

import dm "k8soperation/internal/domain/k8s"

// ClusterStatus 集群状态定义（领域定义在 domain/k8s）
type ClusterStatus = dm.ClusterStatus

const (
	ClusterStatusOK      = dm.ClusterStatusOK
	ClusterStatusBad     = dm.ClusterStatusBad
	ClusterStatusPending = dm.ClusterStatusPending
)

// K8sCluster 集群（领域定义在 domain/k8s）
type K8sCluster = dm.Cluster
