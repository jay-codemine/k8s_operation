# 外部 MySQL 与 Redis Cluster 接入配置指南

> **适用场景**：将平台从本地开发环境（host.docker.internal）切换到外部生产级 MySQL 和 Redis Cluster  
> **核心结论**：无需修改任何应用代码，仅通过 ConfigMap + Secret 配置即可切换

---

## 目录

1. [当前架构 Redis 模式支持](#1-当前架构-redis-模式支持)
2. [认证模式兼容性（用户名/密码可选）](#2-认证模式兼容性)
3. [外部 MySQL 接入配置](#3-外部-mysql-接入配置)
4. [外部 Redis Cluster 接入配置](#4-外部-redis-cluster-接入配置)
5. [修改文件清单总览](#5-修改文件清单总览)
6. [完整配置示例](#6-完整配置示例)
7. [部署操作步骤](#7-部署操作步骤)
8. [验证与排查](#8-验证与排查)

---

## 1. 当前架构 Redis 模式支持

### 1.1 双模式自动切换

平台已**原生支持** Redis 单节点和 Cluster 集群两种模式，切换逻辑零代码改动：

```
┌─────────────────────────────────────────────────────┐
│            配置判断逻辑（redis_client.go）             │
│                                                     │
│   Addresses 数组非空？                               │
│       ├── YES → redis.NewClusterClient（集群模式）    │
│       └── NO  → redis.NewClient（单节点模式）         │
└─────────────────────────────────────────────────────┘
```

**代码实现**（`initialize/redis_client.go`）：

```go
// 判断是否为 Cluster 模式（Addresses 非空则使用 Cluster）
if len(global.CacheSetting.Addresses) > 0 {
    // Redis Cluster 模式
    clusterCli := redis.NewClusterClient(&redis.ClusterOptions{
        Addrs:    global.CacheSetting.Addresses,  // 集群节点列表
        Password: global.CacheSetting.Password,   // 空则跳过认证
        Username: global.CacheSetting.Username,   // 空则跳过 ACL
    })
    global.RedisCli = clusterCli
} else {
    // 单节点模式
    rdb := redis.NewClient(&redis.Options{
        Addr:     global.CacheSetting.Address,    // 单节点地址
        Username: global.CacheSetting.Username,
        Password: global.CacheSetting.Password,
    })
    global.RedisCli = rdb
}
```

### 1.2 关键设计

| 组件 | 实现 | 说明 |
|------|------|------|
| `global.RedisCli` | `redis.UniversalClient` 接口 | 天然兼容单节点和集群 |
| `CacheSettingS` 结构体 | 同时包含 `Address` + `Addresses` | 配置驱动模式选择 |
| Session Store | Cluster 模式自动取 `Addresses[0]` | 无需额外配置 |
| go-redis/v9 | 支持自动 MOVED/ASK 重定向 | Cluster 槽位迁移透明处理 |

### 1.3 当前默认配置

```yaml
# 当前使用：单节点模式
Cache:
  Address: "host.docker.internal:6380"   # 单节点地址
  Addresses: []                           # 空 = 单节点模式
```

---

## 2. 认证模式兼容性

### 2.1 结论：用户名和密码均可为空

**代码中唯一的非空校验是 `Address`（地址），对 Username 和 Password 没有任何强制校验。**

go-redis 底层库行为：

```go
// go-redis 内部连接逻辑：
if password != "" {
    if username != "" {
        // Redis 6+ ACL 模式: AUTH username password
        conn.AuthACL(ctx, username, password)
    } else {
        // Redis 5 传统模式: AUTH password
        conn.Auth(ctx, password)
    }
}
// password 为空 → 完全跳过 AUTH 步骤，直接连接
```

### 2.2 支持的认证组合

| 场景 | Username | Password | 是否支持 | Redis 版本 |
|------|----------|----------|----------|-----------|
| **无密码 Redis** | `""` (空) | `""` (空) | ✅ 支持 | 所有版本 |
| **仅密码认证** | `""` (空) | `"yourpass"` | ✅ 支持 | Redis 5 及以下 |
| **ACL 用户名+密码** | `"default"` 或自定义 | `"yourpass"` | ✅ 支持 | Redis 6+ |
| **仅用户名无密码** | `"user"` | `""` (空) | ⚠️ 不推荐 | Redis 不支持此模式 |

### 2.3 Session Store 兼容性

```go
// session.go 中的注释：
// 单节点和集群模式均允许 Username/Password 为空（无密码 Redis）
// 当 Username/Password 为空时，底层库会自动跳过 AUTH 认证
```

Session Store（`gin-contrib/sessions/redis`）同样支持空用户名/密码，内部会检测到空值后跳过认证步骤。

### 2.4 各场景配置示例

#### 无密码 Redis（开发/测试环境）

```yaml
# Secret
REDIS_PASSWORD: ""
CACHE_USERNAME: ""
```

#### 仅密码 Redis（Redis 5 传统模式）

```yaml
# Secret
REDIS_PASSWORD: "your_redis_password"
CACHE_USERNAME: ""
```

#### ACL 认证（Redis 6+ 生产环境推荐）

```yaml
# Secret
REDIS_PASSWORD: "your_acl_password"
CACHE_USERNAME: "default"          # 或自定义 ACL 用户名
```

---

## 3. 外部 MySQL 接入配置

### 3.1 需要修改的文件

**仅需修改 `deploy/backend/secret.yaml`**（ConfigMap 模板无需改动）

### 3.2 Secret 修改项

```yaml
# deploy/backend/secret.yaml
stringData:
  # ========== 外部 MySQL ==========
  DB_TYPE: "mysql"
  DB_USERNAME: "your_mysql_user"           # ← 替换为外部用户
  DB_PASSWORD: "your_mysql_password"       # ← 替换为外部密码
  DB_HOST: "mysql.example.com"             # ← 替换为外部地址（IP/域名）
  DB_PORT: "3306"                          # ← 替换为外部端口
  DB_NAME: "k8s-platform"                  # ← 确保数据库已创建
  DB_CHARSET: "utf8mb4"
  DB_PARSE_TIME: "true"
  DB_MAX_IDLE_CONNS: "10"                  # 建议：生产环境 10-20
  DB_MAX_OPEN_CONNS: "100"                 # 建议：生产环境 50-200
  DB_MAX_LIFE_SECONDS: "300s"
```

### 3.3 ConfigMap 不需要修改

ConfigMap 中 Database 部分全部使用 `${ENV_VAR}` 占位符，配置值完全由 Secret 注入：

```yaml
# deploy/backend/configmap.yaml（无需修改）
Database:
  DBType: "${DB_TYPE}"
  Username: "${DB_USERNAME}"
  Password: "${DB_PASSWORD}"
  Host: "${DB_HOST}"
  Port: "${DB_PORT}"
  DBName: "${DB_NAME}"
  ...
```

### 3.4 MySQL 集群（主从/MGR）

如果使用 MySQL 集群，只需将 `DB_HOST` 指向：
- **主从模式**：填写主节点地址（写入走主）
- **MGR（InnoDB Cluster）**：填写 MySQL Router 代理地址
- **云 RDS**：填写云厂商提供的连接地址

---

## 4. 外部 Redis Cluster 接入配置

### 4.1 需要修改的文件

| 文件 | 修改内容 | 是否必须 |
|------|----------|----------|
| `deploy/backend/secret.yaml` | Redis 密码、用户名 | ✅ 必须 |
| `deploy/backend/configmap.yaml` | `Addresses` 集群节点列表 | ✅ 必须 |

### 4.2 Secret 修改项

```yaml
# deploy/backend/secret.yaml
stringData:
  # ========== 外部 Redis Cluster ==========
  CACHE_TYPE: "redis"
  CACHE_NAME: "sk_sid"
  REDIS_PASSWORD: "your_cluster_password"   # ← 集群密码（无密码则留空 ""）
  REDIS_ADDRESS: ""                         # ← Cluster 模式建议留空
  CACHE_USERNAME: ""                        # ← ACL 用户（无则留空 ""）
  CACHE_MAX_CONNECT: "20"                   # ← 集群建议加大至 20-50
  CACHE_NETWORK: "tcp"
  CACHE_SECRET: "k8smana"                   # ← Session Cookie 加密密钥
```

### 4.3 ConfigMap 修改项

```yaml
# deploy/backend/configmap.yaml（Cache 部分）
Cache:
  Type: "${CACHE_TYPE}"
  Name: "${CACHE_NAME}"
  Address: "${REDIS_ADDRESS}"
  Addresses:                                # ← 改为集群节点列表
    - "redis-node1.example.com:6379"
    - "redis-node2.example.com:6379"
    - "redis-node3.example.com:6379"
    - "redis-node4.example.com:6379"
    - "redis-node5.example.com:6379"
    - "redis-node6.example.com:6379"
  Username: "${CACHE_USERNAME}"
  Password: "${REDIS_PASSWORD}"
  MaxConnect: "${CACHE_MAX_CONNECT}"
  Network: "${CACHE_NETWORK}"
  Secret: "${CACHE_SECRET}"
```

### 4.4 特殊说明

| 注意项 | 说明 |
|--------|------|
| **Address 字段** | Cluster 模式下可留空，Session Store 会自动取 `Addresses[0]` |
| **所有节点相同密码** | Redis Cluster 要求所有节点使用相同的认证密码 |
| **网络可达性** | Pod 必须能访问所有 Cluster 节点（包括 MOVED 重定向后的目标节点） |
| **不能只暴露代理** | go-redis ClusterClient 需直连所有节点，不能仅通过 Proxy |

### 4.5 使用 Redis Proxy 的方案

如果 Redis Cluster 前有代理（Twemproxy / Redis Proxy / 云代理），可以使用**单节点模式**连代理：

```yaml
# 通过代理访问 Redis Cluster（当作单节点用）
Cache:
  Address: "redis-proxy.example.com:6379"   # ← 代理地址
  Addresses: []                              # ← 留空 = 单节点模式
```

---

## 5. 修改文件清单总览

### 5.1 一图看清

```
deploy/backend/
├── secret.yaml          ← ✅ 必改（地址、端口、用户名、密码）
├── configmap.yaml       ← ⚠️ 按需改（Redis Cluster 节点列表）
├── namespace.yaml       ← ❌ 无需修改
├── pv.yaml              ← ❌ 无需修改
├── pvc.yaml             ← ❌ 无需修改
├── service.yaml         ← ❌ 无需修改
├── deployment.yaml      ← ❌ 无需修改
└── kustomization.yaml   ← ❌ 无需修改
```

### 5.2 详细清单

| # | 文件 | 修改字段 | 场景 |
|---|------|----------|------|
| 1 | `deploy/backend/secret.yaml` | `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD` | 外部 MySQL |
| 2 | `deploy/backend/secret.yaml` | `REDIS_ADDRESS`, `REDIS_PASSWORD`, `CACHE_USERNAME` | 外部 Redis |
| 3 | `deploy/backend/configmap.yaml` | `Cache.Addresses` 数组 | Redis Cluster 模式 |
| 4 | `deploy/backend/configmap.yaml` | 无需修改 Database 部分 | 外部 MySQL |

### 5.3 不同场景修改对照

| 场景 | secret.yaml | configmap.yaml | 代码改动 |
|------|-------------|----------------|----------|
| 外部 MySQL 单实例 | 改 DB_HOST/PORT/PASSWORD | 不需要 | 无 |
| 外部 MySQL 集群(主从/MGR) | 改 DB_HOST 指向 Router | 不需要 | 无 |
| 外部 Redis 单节点 | 改 REDIS_ADDRESS/PASSWORD | 不需要 | 无 |
| 外部 Redis 单节点(无密码) | REDIS_PASSWORD="" | 不需要 | 无 |
| **外部 Redis Cluster(有密码)** | 改 REDIS_PASSWORD | 改 Addresses 列表 | 无 |
| **外部 Redis Cluster(无密码)** | REDIS_PASSWORD="" | 改 Addresses 列表 | 无 |
| Redis Proxy 代理 | 改 REDIS_ADDRESS 为代理地址 | 不需要 | 无 |

---

## 6. 完整配置示例

### 6.1 外部 MySQL + Redis Cluster（有密码）

**`deploy/backend/secret.yaml`**：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: k8soperation-secret
  namespace: k8soperation
type: Opaque
stringData:
  # ===== 外部 MySQL（生产） =====
  DB_TYPE: "mysql"
  DB_USERNAME: "k8s_app"
  DB_PASSWORD: "Pr0d_MySQL@2024!"
  DB_HOST: "mysql-master.prod.internal"
  DB_PORT: "3306"
  DB_NAME: "k8s-platform"
  DB_CHARSET: "utf8mb4"
  DB_PARSE_TIME: "true"
  DB_MAX_IDLE_CONNS: "20"
  DB_MAX_OPEN_CONNS: "200"
  DB_MAX_LIFE_SECONDS: "300s"

  # ===== 外部 Redis Cluster（生产） =====
  CACHE_TYPE: "redis"
  CACHE_NAME: "sk_sid"
  REDIS_PASSWORD: "Pr0d_Redis@2024!"
  REDIS_ADDRESS: ""                        # Cluster 模式留空
  CACHE_USERNAME: "default"                # Redis 6+ ACL
  CACHE_MAX_CONNECT: "30"
  CACHE_NETWORK: "tcp"
  CACHE_SECRET: "random_session_key_2024"

  # ===== 其他配置保持不变 =====
  JWT_SIGNING_KEY: "eoNB0%bv5M7995F1"
  JENKINS_URL: "http://jenkins.devops.svc.cluster.local:8080/"
  JENKINS_USERNAME: "ops-dev"
  JENKINS_API_TOKEN: "1189c2297408abc5543cdecbdb5427e050"
  HMAC_SECRET: "changeme"
  KUBECONFIG_ENCRYPT_KEY: "changemp"
  DINGTALK_WEBHOOK: ""
  PLATFORM_FRONTEND_URL: "http://localhost:5173"
```

**`deploy/backend/configmap.yaml`（Cache 部分修改）**：

```yaml
    Cache:
      Type: "${CACHE_TYPE}"
      Name: "${CACHE_NAME}"
      Address: "${REDIS_ADDRESS}"
      Addresses:
        - "10.0.1.101:6379"
        - "10.0.1.102:6379"
        - "10.0.1.103:6379"
        - "10.0.1.104:6379"
        - "10.0.1.105:6379"
        - "10.0.1.106:6379"
      Username: "${CACHE_USERNAME}"
      Password: "${REDIS_PASSWORD}"
      MaxConnect: "${CACHE_MAX_CONNECT}"
      Network: "${CACHE_NETWORK}"
      Secret: "${CACHE_SECRET}"
```

### 6.2 外部 MySQL + Redis Cluster（无密码）

**`deploy/backend/secret.yaml`（Redis 部分）**：

```yaml
  # ===== 外部 Redis Cluster（无密码） =====
  CACHE_TYPE: "redis"
  CACHE_NAME: "sk_sid"
  REDIS_PASSWORD: ""                       # ← 空字符串 = 不认证
  REDIS_ADDRESS: ""                        # ← Cluster 模式留空
  CACHE_USERNAME: ""                       # ← 空字符串 = 不用 ACL
  CACHE_MAX_CONNECT: "30"
  CACHE_NETWORK: "tcp"
  CACHE_SECRET: "random_session_key_2024"
```

ConfigMap 的 `Addresses` 同样填写集群节点列表即可。

### 6.3 外部 Redis 单节点（无密码）

**`deploy/backend/secret.yaml`**：

```yaml
  REDIS_PASSWORD: ""                       # ← 无密码
  REDIS_ADDRESS: "redis.prod.internal:6379" # ← 单节点地址
  CACHE_USERNAME: ""                       # ← 无用户名
```

**ConfigMap 无需修改**（`Addresses: []` 保持为空即走单节点模式）。

---

## 7. 部署操作步骤

### 7.1 切换到外部中间件

```bash
# Step 1: 修改 Secret（敏感配置）
vim deploy/backend/secret.yaml
# 修改 DB_HOST、DB_PORT、DB_PASSWORD、REDIS_PASSWORD 等

# Step 2: 修改 ConfigMap（仅 Redis Cluster 需要）
vim deploy/backend/configmap.yaml
# 修改 Cache.Addresses 数组为集群节点列表

# Step 3: 应用变更
kubectl apply -k deploy/backend/

# Step 4: 重启后端 Pod（使新配置生效）
kubectl -n k8soperation rollout restart deployment/k8soperation

# Step 5: 等待就绪
kubectl -n k8soperation rollout status deployment/k8soperation

# Step 6: 查看日志确认连接
kubectl -n k8soperation logs -f deploy/k8soperation --tail=30
```

### 7.2 预期日志输出

**Redis Cluster 模式**：
```
[Redis] Cluster mode enabled, nodes: [10.0.1.101:6379 10.0.1.102:6379 10.0.1.103:6379 ...]
[Session] Cluster 模式: 使用节点 10.0.1.101:6379 作为 Session Store
[Session] Redis Session Store 初始化成功 (address=10.0.1.101:6379, cluster=true)
```

**Redis 单节点模式**：
```
[Session] Redis Session Store 初始化成功 (address=redis.prod.internal:6379, cluster=false)
```

---

## 8. 验证与排查

### 8.1 验证连接

```bash
# 验证后端健康
kubectl -n k8soperation exec deploy/k8soperation -- wget -qO- http://localhost:8080/healthz/ready
# 返回 200 表示 DB + Redis 均已就绪

# 查看启动日志
kubectl -n k8soperation logs deploy/k8soperation | grep -E "Redis|Session|DB|connect"
```

### 8.2 常见错误

| 错误信息 | 原因 | 解决方案 |
|----------|------|----------|
| `redis cluster ping failed` | 集群节点不可达 | 检查网络策略/防火墙，确保 Pod 能访问所有节点 |
| `redis ping failed` | 单节点连接失败 | 检查地址和端口，确认 Redis 运行中 |
| `redis address is empty` | 未配置地址 | 单节点模式必须填 `REDIS_ADDRESS`；Cluster 模式必须填 `Addresses` |
| `connect db failed` | MySQL 连接失败 | 检查 DB_HOST、网络可达性、账号权限 |
| `WRONGPASS` / `AUTH failed` | 密码错误 | 确认 Secret 中密码与 Redis 实际密码一致 |
| `NOAUTH` | 需要认证但未提供密码 | 在 Secret 中填写 REDIS_PASSWORD |

### 8.3 网络排查

```bash
# 从 Pod 内测试 MySQL 连通性
kubectl -n k8soperation exec deploy/k8soperation -- \
  sh -c "wget -qO- --timeout=3 http://mysql.example.com:3306 2>&1 || echo 'port reachable'"

# 从 Pod 内测试 Redis 连通性
kubectl -n k8soperation exec deploy/k8soperation -- \
  sh -c "echo PING | nc -w 3 redis-node1.example.com 6379"
```

### 8.4 配置注入机制说明

```
┌────────────┐     注入 ENV      ┌────────────────┐     os.ExpandEnv()     ┌───────────────┐
│   Secret   │ ─────────────────→│  Pod 环境变量   │ ────────────────────→  │ 最终 config   │
│ (敏感值)    │                   │  ${DB_HOST}=..│                        │ Host: 实际值   │
└────────────┘                   └────────────────┘                        └───────────────┘
                                         ↑
┌────────────┐     挂载文件       │
│  ConfigMap │ ─────────────────→│  /app/configs/config.yaml（含 ${} 占位符）
│ (模板)      │                   
└────────────┘                   
```

**流程**：
1. Secret 中的 key-value 通过 Deployment `env` 字段注入为容器环境变量
2. ConfigMap 中的 `config.yaml` 以文件形式挂载到 `/app/configs/config.yaml`
3. 应用启动时通过 `os.ExpandEnv()` 将配置文件中的 `${ENV_VAR}` 替换为实际环境变量值
4. Viper 读取展开后的配置文件，填充到 `global.CacheSetting`、`global.DatabaseSetting` 等结构体

---

## 附录：配置结构体定义

```go
// pkg/setting/section.go

// CacheSettingS 缓存配置
type CacheSettingS struct {
    Type       string   // 缓存类型：固定 "redis"
    Name       string   // Session 名称前缀
    Address    string   // 单节点地址 "host:port"
    Addresses  []string // Cluster 节点列表 ["host1:6379", "host2:6379"]
    Username   string   // ACL 用户名（可空）
    Password   string   // 密码（可空）
    MaxConnect int      // 连接池大小
    Network    string   // 网络类型 "tcp"
    Secret     string   // Session Cookie 加密密钥
}
```

**判断规则**：
- `Addresses` 非空 → `redis.NewClusterClient` → 集群模式
- `Addresses` 为空 → `redis.NewClient` → 单节点模式
- `Password` 为空 → 跳过 AUTH → 无密码连接
- `Username` 为空 → 不使用 ACL → 传统密码认证或无认证
