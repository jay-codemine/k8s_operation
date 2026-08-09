package models

import dm "k8soperation/internal/domain/tenant"

// Tenant 租户（领域定义在 domain/tenant）
type Tenant = dm.Tenant

// DefaultTenantID 默认租户 ID
const DefaultTenantID = dm.DefaultTenantID
