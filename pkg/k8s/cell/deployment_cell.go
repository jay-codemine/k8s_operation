package cell

import (
	appv1 "k8s.io/api/apps/v1"
	"time"
)

type DeploymentCell appv1.Deployment

// DeploymentCell 结构体的 GetName 方法
// 该方法用于获取 DeploymentCell 的名称
// 返回值: string - 返回 DeploymentCell 的名称字段
func (d *DeploymentCell) GetName() string {
	return d.Name
}

func (d *DeploymentCell) GetCreation() time.Time {
	return d.CreationTimestamp.Time
}
