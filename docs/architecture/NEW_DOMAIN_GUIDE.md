# 新增领域完整实现指南——以 CMDB 为例

## 第一步：建模

CMDB 管理服务器资产：主机、IP、机房、操作系统。

### 实体关系
```
Asset (资产聚合根)
  ├── ID, Hostname, IP, OS, CPU, Memory, Disk
  ├── Status (online/offline/maintenance)
  └── BelongsTo → IDC (机房)
```

---

## 第二步：创建文件（共需 7 个）

### 文件 1: `internal/domain/cmdb/models.go`

```go
package cmdb

// AssetStatus 资产状态
type AssetStatus string

const (
    AssetStatusOnline      AssetStatus = "online"
    AssetStatusOffline     AssetStatus = "offline"
    AssetStatusMaintenance AssetStatus = "maintenance"
)

// Asset 资产实体
type Asset struct {
    ID         uint32      `gorm:"primaryKey" json:"id"`
    TenantID   uint32      `gorm:"column:tenant_id;default:1;index" json:"tenant_id"`
    Hostname   string      `gorm:"size:128;not null" json:"hostname"`
    IP         string      `gorm:"size:45" json:"ip"`
    OSType     string      `gorm:"size:50" json:"os_type"`
    OSVersion  string      `gorm:"size:50" json:"os_version"`
    CPU        int         `gorm:"default:0" json:"cpu"`
    Memory     int64       `gorm:"default:0" json:"memory"`
    Disk       int64       `gorm:"default:0" json:"disk"`
    IDCID      uint32      `gorm:"column:idc_id;index" json:"idc_id"`
    Status     AssetStatus `gorm:"size:20;default:online" json:"status"`
    Tags       string      `gorm:"type:text" json:"tags"`
    IsDel      uint8       `gorm:"column:is_del;default:0" json:"-"`
    CreatedAt  uint64      `json:"created_at"`
    ModifiedAt uint64      `json:"modified_at"`
}

func (Asset) TableName() string { return "cmdb_asset" }
```

### 文件 2: `internal/domain/cmdb/repository.go`

```go
package cmdb

import "context"

// AssetRepository 资产仓储接口
type AssetRepository interface {
    Save(ctx context.Context, a *Asset) error
    FindByID(ctx context.Context, id uint32) (*Asset, error)
    Update(ctx context.Context, id uint32, values map[string]interface{}) error
    Delete(ctx context.Context, id uint32) error
    Query(ctx context.Context, keyword string, status AssetStatus, page, limit int) ([]*Asset, int64, error)
}

// 仓储接口定义了 CMDB 域需要的数据操作。
// 它不依赖 GORM——实现在 infra 层。
```

### 文件 3: `internal/domain/cmdb/service.go`

```go
package cmdb

import (
    "context"
    "fmt"
    "time"

    "k8soperation/internal/domain/events"
    "k8soperation/pkg/logger"
)

// AssetService 资产领域服务
type AssetService struct {
    repo      AssetRepository
    logger    *logger.Logger
    publisher events.EventPublisher
}

func NewAssetService(repo AssetRepository, logger *logger.Logger, publisher events.EventPublisher) *AssetService {
    return &AssetService{repo: repo, logger: logger, publisher: publisher}
}

func (s *AssetService) Create(ctx context.Context, req *CreateAssetRequest) (*Asset, error) {
    // 1. 值对象验证（见 values.go）
    hostname, err := NewHostname(req.Hostname)
    if err != nil {
        return nil, err
    }
    ip, err := NewIPAddress(req.IP)
    if err != nil {
        return nil, err
    }
    // 2. 领域规则：检查主机名唯一
    if exists, _ := s.repo.FindByHostname(ctx, req.Hostname); exists != nil {
        return nil, fmt.Errorf("主机名 %s 已存在", req.Hostname)
    }
    // 3. 落库
    now := uint64(time.Now().Unix())
    asset := &Asset{
        Hostname: hostname.String(), IP: ip.String(),
        OSType: req.OSType, OSVersion: req.OSVersion,
        CPU: req.CPU, Memory: req.Memory, Disk: req.Disk,
        IDCID: req.IDCID, Status: AssetStatusOnline,
        CreatedAt: now, ModifiedAt: now,
    }
    if err := s.repo.Save(ctx, asset); err != nil {
        s.logger.Error("创建资产失败", zap.Error(err))
        return nil, err
    }
    // 4. 领域事件
    s.publish(NewAssetCreated(asset.ID, asset.Hostname))
    return asset, nil
}

func (s *AssetService) Update(ctx context.Context, id uint32, req *UpdateAssetRequest) error {
    _, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return fmt.Errorf("资产不存在")
    }
    values := map[string]interface{}{
        "hostname": req.Hostname, "ip": req.IP,
        "os_type": req.OSType, "status": req.Status,
        "modified_at": uint64(time.Now().Unix()),
    }
    return s.repo.Update(ctx, id, values)
}

func (s *AssetService) Query(ctx context.Context, keyword string, status AssetStatus, page, limit int) ([]*Asset, int64, error) {
    return s.repo.Query(ctx, keyword, status, page, limit)
}

func (s *AssetService) GetByID(ctx context.Context, id uint32) (*Asset, error) {
    return s.repo.FindByID(ctx, id)
}

func (s *AssetService) Delete(ctx context.Context, id uint32) error {
    return s.repo.Delete(ctx, id)
}

func (s *AssetService) publish(event events.DomainEvent) {
    if s.publisher == nil { return }
    defer func() { _ = recover() }()
    s.publisher.Publish(event)
}

// CreateAssetRequest / UpdateAssetRequest 请求 DTO
type CreateAssetRequest struct {
    Hostname, IP, OSType, OSVersion string
    CPU, Memory, Disk               int
    IDCID                           uint32
}
type UpdateAssetRequest struct {
    Hostname, IP, OSType, OSVersion string
    Status                          AssetStatus
}
```

### 文件 4: `internal/domain/cmdb/values.go`

```go
package cmdb

import (
    "fmt"
    "net"
    "strings"
)

// Hostname 主机名值对象
type Hostname struct{ val string }

func NewHostname(s string) (Hostname, error) {
    s = strings.TrimSpace(s)
    if s == "" {
        return Hostname{}, fmt.Errorf("主机名不能为空")
    }
    if len(s) > 128 {
        return Hostname{}, fmt.Errorf("主机名最长 128 字符")
    }
    return Hostname{val: s}, nil
}
func (h Hostname) String() string { return h.val }

// IPAddress IP 地址值对象
type IPAddress struct{ val string }

func NewIPAddress(s string) (IPAddress, error) {
    s = strings.TrimSpace(s)
    if s == "" {
        return IPAddress{}, nil // IP 可选
    }
    if net.ParseIP(s) == nil {
        return IPAddress{}, fmt.Errorf("无效的 IP 地址: %s", s)
    }
    return IPAddress{val: s}, nil
}
func (ip IPAddress) String() string { return ip.val }
```

### 文件 5: `internal/domain/cmdb/events.go`

```go
package cmdb

import "k8soperation/internal/domain/events"

type AssetCreated struct {
    events.BaseEvent
    AssetID  uint32
    Hostname string
}

func NewAssetCreated(id uint32, hostname string) AssetCreated {
    return AssetCreated{
        BaseEvent: events.NewBaseEvent("cmdb.asset.created"),
        AssetID: id, Hostname: hostname,
    }
}
```

### 文件 6: `internal/infra/persistence/cmdb_repo.go`

```go
package persistence

import (
    "context"
    "gorm.io/gorm"
    "k8soperation/internal/domain/cmdb"
)

type assetRepo struct{ db *gorm.DB }

func NewAssetRepository(db *gorm.DB) cmdb.AssetRepository { return &assetRepo{db: db} }

func (r *assetRepo) Save(ctx context.Context, a *cmdb.Asset) error {
    return r.db.WithContext(ctx).Create(a).Error
}
func (r *assetRepo) FindByID(ctx context.Context, id uint32) (*cmdb.Asset, error) {
    var a cmdb.Asset
    err := r.db.WithContext(ctx).Where("id = ? AND is_del = 0", id).First(&a).Error
    return &a, err
}
func (r *assetRepo) Update(ctx context.Context, id uint32, values map[string]interface{}) error {
    return r.db.WithContext(ctx).Model(&cmdb.Asset{}).Where("id = ?", id).Updates(values).Error
}
func (r *assetRepo) Delete(ctx context.Context, id uint32) error {
    return r.db.WithContext(ctx).Model(&cmdb.Asset{}).Where("id = ?", id).Update("is_del", 1).Error
}
func (r *assetRepo) Query(ctx context.Context, keyword string, status cmdb.AssetStatus, page, limit int) ([]*cmdb.Asset, int64, error) {
    db := r.db.WithContext(ctx).Model(&cmdb.Asset{}).Where("is_del = 0")
    if keyword != "" {
        db = db.Where("hostname LIKE ? OR ip LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
    }
    if status != "" {
        db = db.Where("status = ?", status)
    }
    var total int64; db.Count(&total)
    var list []*cmdb.Asset
    if page <= 0 { page = 1 }
    if limit <= 0 { limit = 20 }
    db.Order("id DESC").Offset((page-1)*limit).Limit(limit).Find(&list)
    return list, total, nil
}

// 额外领域查询（不在接口中，但在 repo 中可用）
func (r *assetRepo) FindByHostname(ctx context.Context, hostname string) (*cmdb.Asset, error) {
    var a cmdb.Asset
    err := r.db.WithContext(ctx).Where("hostname = ? AND is_del = 0", hostname).First(&a).Error
    return &a, err
}
```

### 文件 7: `internal/app/services/cmdb.go`

```go
package services

import (
    "k8soperation/internal/app/models"
    "k8soperation/internal/domain/cmdb"
    "k8soperation/internal/infra/persistence"
)

func (s *Services) cmdbSvc() *cmdb.AssetService {
    return cmdb.NewAssetService(
        persistence.NewAssetRepository(s.db),
        s.logger,
        s.eventBus,
    )
}

// AssetCreate 创建资产
func (s *Services) AssetCreate(req *cmdb.CreateAssetRequest) (*models.Asset, error) {
    return s.cmdbSvc().Create(context.Background(), req)
}

// AssetUpdate 更新资产
func (s *Services) AssetUpdate(id uint32, req *cmdb.UpdateAssetRequest) error {
    return s.cmdbSvc().Update(context.Background(), id, req)
}

// AssetQuery 查询资产列表
func (s *Services) AssetQuery(keyword string, status cmdb.AssetStatus, page, limit int) ([]*models.Asset, int64, error) {
    return s.cmdbSvc().Query(context.Background(), keyword, status, page, limit)
}

// AssetGetByID 获取资产详情
func (s *Services) AssetGetByID(id uint32) (*models.Asset, error) {
    return s.cmdbSvc().GetByID(context.Background(), id)
}

// AssetDelete 删除资产
func (s *Services) AssetDelete(id uint32) error {
    return s.cmdbSvc().Delete(context.Background(), id)
}
```

### 文件 8: `internal/app/models/cmdb.go`（防腐层）

```go
package models

import dm "k8soperation/internal/domain/cmdb"

type Asset = dm.Asset
type AssetStatus = dm.AssetStatus
```

### 文件 9: 路由 + Controller（省略，复用现有 K8s CRUD 路由模式）

---

## 第三步：注册事件处理器

```go
// bootstrap/event_handlers.go 增加:
bus.Subscribe("cmdb.asset.created", func(e events.DomainEvent) {
    global.Logger.Infof("[DomainEvent] CMDB 资产创建: %s", e.EventName())
})
```

---

## 第四步：数据库建表

```sql
CREATE TABLE cmdb_asset (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    tenant_id INT UNSIGNED DEFAULT 1,
    hostname VARCHAR(128) NOT NULL,
    ip VARCHAR(45),
    os_type VARCHAR(50),
    os_version VARCHAR(50),
    cpu INT DEFAULT 0,
    memory BIGINT DEFAULT 0,
    disk BIGINT DEFAULT 0,
    idc_id INT UNSIGNED,
    status VARCHAR(20) DEFAULT 'online',
    tags TEXT,
    is_del TINYINT DEFAULT 0,
    created_at BIGINT UNSIGNED,
    modified_at BIGINT UNSIGNED,
    INDEX idx_tenant (tenant_id),
    INDEX idx_hostname (hostname)
);
```

---

## 完整新增域检查清单

```
□ 1. domain/{name}/models.go        ← 实体
□ 2. domain/{name}/repository.go    ← 仓储接口
□ 3. domain/{name}/service.go       ← 领域服务
□ 4. domain/{name}/values.go        ← 值对象 ★推荐
□ 5. domain/{name}/events.go        ← 领域事件 ★推荐
□ 6. domain/{name}/aggregate.go     ← 聚合根 (如需)
□ 7. infra/persistence/{name}_repo.go ← GORM 实现
□ 8. app/services/{name}.go         ← Services hook
□ 9. app/models/{name}.go           ← 类型别名
□ 10. app/controllers/...            ← HTTP 控制器
□ 11. app/routers/...                ← 路由注册
□ 12. bootstrap/event_handlers.go   ← 事件处理器
□ 13. DB DDL                        ← 建表语句
```

## 时间估算

| 复杂度 | 文件数 | 时间 |
|--------|:---:|------|
| 简单（如 audit，1 实体） | 7 个 | 30 分钟 |
| 中等（如 cmdb，1-3 实体） | 9 个 | 1 小时 |
| 复杂（如 cicd，10+ 实体） | 15+ 个 | 3-4 小时 |

---

## 设计模式速查

```
❌ 禁止:  domain 层 import global
❌ 禁止:  domain 层暴露 DB()
❌ 禁止:  Controller 直接 global.DB
❌ 禁止:  实体上挂 Create(tx *gorm.DB)

✅ 必须:  Repository 接口在 domain 定义
✅ 必须:  GORM 实现在 infra/persistence
✅ 必须:  Services hook 在 app/services
✅ 必须:  值对象工厂 + 自验证
✅ 推荐:  领域事件解耦跨域
✅ 推荐:  聚合根 + 状态机
```
