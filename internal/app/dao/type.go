package dao

type K8sClusterWithPlain struct {
	ID          uint32
	ClusterName string
	KubeConfig  string
	ClusterVer  string
	Status      uint8
	ModifiedAt  uint64
}
