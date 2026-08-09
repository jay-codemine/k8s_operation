package tenant

import "k8soperation/pkg/db"

// Tenant 租户（组织/团队隔离单元）
type Tenant struct {
	ID     uint32 `gorm:"primary_key" json:"id"`
	Name   string `gorm:"column:name;size:128;not null" json:"name" description:"租户名称"`
	Code   string `gorm:"column:code;size:64;uniqueIndex;not null" json:"code" description:"租户编码，用于 API/路由标识"`
	Status int8   `gorm:"column:status;default:1" json:"status" description:"状态:1启用,0禁用"`
	*db.Base
}

func (Tenant) TableName() string { return "tenant" }

// AggregateID 实现 domain.AggregateRoot 接口
func (t Tenant) AggregateID() int64 { return int64(t.ID) }

// DefaultTenantID 默认租户 ID
const DefaultTenantID uint32 = 1
