package db

// Base 可嵌入的数据库基础字段
type Base struct {
	ID         uint32 `gorm:"primary_key" json:"id" description:"自增ID"`
	TenantID   uint32 `gorm:"column:tenant_id;default:1;index" json:"tenant_id" description:"租户ID,1=默认租户"`
	CreatedAt  uint32 `json:"created_at" description:"创建时间"`
	ModifiedAt uint32 `json:"modified_at" description:"修改时间"`
	DeletedAt  uint32 `json:"deleted_at" description:"删除时间"`
	IsDel      uint8  `json:"is_del" description:"是否删除,1 表示删除,0表示未删除"`
}
