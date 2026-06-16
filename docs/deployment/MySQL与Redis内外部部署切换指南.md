# MySQL 与 Redis 内外部部署切换指南

> 本文档详细说明：K8s 集群部署时，如何在「集群内部署 MySQL/Redis」和「使用外部 MySQL/Redis」之间切换。

## 推荐方式：直接使用对应子目录

项目已提供 **三套独立部署配置**，可根据场景一键选用：

```bash
# 方案 A：后端 + K8s 内 MySQL/Redis（开发/测试）
kubectl apply -k deploy/backend/

# 方案 B：外部 MySQL + Redis Cluster（生产推荐）
kubectl apply -k deploy/external/

# 方案 C：前端
kubectl apply -k deploy/frontend/

# 方案 D：前后端一起（K8s 内 MySQL/Redis）
kubectl apply -k deploy/
```

### 目录结构

```
deploy/
├── kustomization.yaml      ← 总编排入口（前后端一起）
├── backend/                ← K8s 内 MySQL + Redis 单节点模式
│   ├── configmap.yaml      # Address: redis:6379（单节点）
│   ├── middleware.yaml     # MySQL + Redis Deployment
│   └── ...
├── frontend/               ← 前端（Nginx + Vue3）
│   └── ...
└── external/               ← 外部 MySQL + Redis Cluster 模式（生产推荐）
    ├── configmap.yaml      # Addresses: [多节点]（Cluster 模式）
    ├── secret.yaml         # 外部密码
    └── ...                 # 无 middleware.yaml（不部署数据库）
```

---

> 如果需要在同一套配置内手动切换，以下是详细说明（改 3 个文件）。

---

## 目录

- [方案对比](#方案对比)
- [方案一：外部数据库（推荐生产环境）](#方案一外部数据库推荐生产环境)
- [方案二：集群内部署（开发/测试环境）](#方案二集群内部署开发测试环境)
- [需要修改的文件清单](#需要修改的文件清单)
- [切换步骤详解](#切换步骤详解)
- [配置模板速查](#配置模板速查)
- [验证方法](#验证方法)
- [常见问题](#常见问题)

---

## 方案对比

| 维度 | 外部（集群外） | 内部（集群内） |
|------|---------------|---------------|
| 适用环境 | **生产、预发** | 开发、测试 |
| 数据安全 | 云 RDS 自动备份、主从 | 依赖 PVC，需自行备份 |
| 高可用 | 云厂商保障 | 单节点，需自建主从 |
| 运维成本 | 低（托管） | 高（自维护） |
| 网络延迟 | 同 VPC < 1ms | Pod 内 < 0.1ms |
| 费用 | 有 RDS/Redis 费用 | 共享集群资源 |

---

## 方案一：外部数据库（推荐生产环境）

### 架构图

```
┌──── K8s Cluster ────────────────┐
│                                 │
│  Pod: k8soperation              │
│    ├── 读取 ConfigMap 配置       │
│    └── 通过内网地址连接 ─────────┼──────┐
│                                 │      │
└─────────────────────────────────┘      │
                                         ▼
                               ┌──────────────────┐
                               │ 外部 MySQL (RDS) │
                               │ 10.0.1.100:3306  │
                               └──────────────────┘
                               ┌──────────────────┐
                               │ 外部 Redis       │
                               │ 10.0.1.100:6379  │
                               └──────────────────┘
```

### 需要改的地方

| # | 文件 | 改什么 |
|---|------|--------|
| 1 | `deploy/backend/configmap.yaml` | Database.Host → 外部地址；Cache.Address/Addresses → 外部地址 |
| 2 | `deploy/backend/secret.yaml` | DB_PASSWORD → 外部数据库密码；REDIS_PASSWORD → 外部 Redis 密码 |
| 3 | `deploy/backend/kustomization.yaml` | 确认 `middleware.yaml` 保持注释（不部署内部数据库） |

> **更简单的方式**：直接使用 `kubectl apply -k deploy/external/`，无需手动改文件。

### 具体修改

#### 1. `deploy/backend/configmap.yaml`

```yaml
    Database:
      DBType: mysql
      Username: root                              # 改为外部数据库用户名
      Password: "${DB_PASSWORD}"
      Host: rm-bp1xxx.mysql.rds.aliyuncs.com     # ← 改这里：外部 MySQL 地址
      Port: "3306"
      DBName: k8s-platform
      ...

    Cache:
      Type: redis
      Name: sk_sid
      Address: ""                                 # 单节点留空（Cluster 模式用 Addresses）
      Addresses:                                   # ← Redis Cluster 节点列表
        - 192.168.1.201:6379
        - 192.168.1.202:6379
        - 192.168.1.203:6379
        - 192.168.1.204:6379
        - 192.168.1.205:6379
        - 192.168.1.206:6379
      Username: ""
      Password: "${REDIS_PASSWORD}"
      ...
```

> **Redis 模式说明**：
> - 单节点：填 `Address`，`Addresses` 留空
> - Cluster 集群：`Address` 留空，`Addresses` 填多个节点地址

#### 2. `deploy/backend/secret.yaml`（改密码）

```bash
# 生成外部 MySQL 密码的 base64
echo -n "your-rds-password" | base64
# 输出: eW91ci1yZHMtcGFzc3dvcmQ=

# 生成外部 Redis 密码的 base64
echo -n "your-redis-password" | base64
# 输出: eW91ci1yZWRpcy1wYXNzd29yZA==
```

```yaml
data:
  DB_PASSWORD: "eW91ci1yZHMtcGFzc3dvcmQ="       # ← 替换为外部 MySQL 密码
  REDIS_PASSWORD: "eW91ci1yZWRpcy1wYXNzd29yZA==" # ← 替换为外部 Redis 密码
```

#### 3. `deploy/backend/kustomization.yaml`（保持注释）

```yaml
resources:
  - namespace.yaml
  - secret.yaml
  - configmap.yaml
  - pvc.yaml
  - service.yaml
  - deployment.yaml
  # - middleware.yaml   # ← 保持注释！不部署内部 MySQL/Redis
```

---

## 方案二：集群内部署（开发/测试环境）

### 架构图

```
┌──── K8s Cluster ─────────────────────────────────────────┐
│                                                           │
│  Pod: k8soperation                                        │
│    └── 通过 K8s Service 名称连接                           │
│              │                                            │
│    ┌─────────┴──────────┐                                 │
│    ▼                    ▼                                 │
│  ┌──────────────┐  ┌──────────────┐                      │
│  │ Pod: mysql   │  │ Pod: redis   │                      │
│  │ Svc: mysql   │  │ Svc: redis   │                      │
│  │ PVC: 10Gi   │  │ PVC: 2Gi    │                      │
│  └──────────────┘  └──────────────┘                      │
│                                                           │
└───────────────────────────────────────────────────────────┘
```

### 需要改的地方

| # | 文件 | 改什么 |
|---|------|--------|
| 1 | `deploy/backend/configmap.yaml` | Database.Host → `mysql`；Cache.Address → `redis:6379` |
| 2 | `deploy/backend/secret.yaml` | DB_PASSWORD / REDIS_PASSWORD → 内部数据库密码 |
| 3 | `deploy/backend/kustomization.yaml` | 取消注释 `- middleware.yaml` |

### 具体修改

#### 1. `deploy/backend/configmap.yaml`（只改 2 行）

```yaml
    Database:
      DBType: mysql
      Username: root
      Password: "${DB_PASSWORD}"
      Host: mysql                           # ← 改这里：K8s Service 名称
      Port: "3306"
      DBName: k8s-platform
      ...

    Cache:
      Type: redis
      Name: sk_sid
      Address: redis:6379                   # ← 改这里：K8s Service 名称:端口（单节点模式）
      Addresses: []                          # 留空（不使用 Cluster 模式）
      Username: ""
      Password: "${REDIS_PASSWORD}"
      ...
```

> **说明**：K8s 内同 namespace 可直接用 Service 名称（如 `mysql`）访问，也可用完整形式 `mysql.k8soperation.svc.cluster.local`

#### 2. `deploy/secret.yaml`（设置内部密码）

```bash
# 设置一个内部数据库密码（MySQL 和 Redis 会用这个密码创建）
echo -n "k8s-internal-2024" | base64
# 输出: azhzLWludGVybmFsLTIwMjQ=
```

```yaml
data:
  DB_PASSWORD: "azhzLWludGVybmFsLTIwMjQ="       # MySQL root 密码
  REDIS_PASSWORD: "azhzLWludGVybmFsLTIwMjQ="    # Redis requirepass
```

> **重要**：`middleware.yaml` 中 MySQL/Redis 会读取同一个 Secret，密码自动同步。

#### 3. `deploy/kustomization.yaml`（取消注释）

```yaml
resources:
  - namespace.yaml
  - secret.yaml
  - configmap.yaml
  - pvc.yaml
  - service.yaml
  - deployment.yaml
  - middleware.yaml   # ← 取消注释！部署内部 MySQL + Redis
```

#### 4. 初始化数据库

内部 MySQL 启动后，还需要导入初始化 SQL：

```bash
# 等待 MySQL Pod 就绪
kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=mysql -n k8soperation --timeout=120s

# 拷贝 SQL 文件到 Pod
kubectl cp docs/sql/k8s_platform_full_init.sql k8soperation/$(kubectl get pod -n k8soperation -l app.kubernetes.io/component=mysql -o jsonpath='{.items[0].metadata.name}'):/tmp/init.sql

# 执行初始化
kubectl exec -n k8soperation deploy/mysql -- mysql -u root -p"k8s-internal-2024" --default-character-set=utf8mb4 -e "source /tmp/init.sql"
```

---

## 需要修改的文件清单

### 一图看清

```
deploy/
├── kustomization.yaml          ← 总编排入口
├── backend/                    ← K8s 内 MySQL + Redis 模式
│   ├── configmap.yaml          ← 改 Host/Address/Addresses（数据库连接地址）
│   ├── secret.yaml             ← 改 DB_PASSWORD / REDIS_PASSWORD（密码）
│   ├── kustomization.yaml      ← 注释/取消注释 middleware.yaml
│   ├── middleware.yaml         （内部部署时使用）
│   ├── namespace.yaml          （无需改）
│   ├── pvc.yaml                （无需改）
│   ├── service.yaml            （无需改）
│   └── deployment.yaml         （无需改）
├── frontend/                   ← 前端部署
└── external/                   ← ★ 外部 MySQL + Redis Cluster 模式（生产推荐）
    ├── configmap.yaml          # 已配置外部地址 + Redis Cluster
    ├── secret.yaml             # 外部密码
    └── ...                     # 一键部署：kubectl apply -k deploy/external/
```

---

## 切换步骤详解

### 从「外部」切换到「内部」

```bash
# Step 1: 修改 configmap.yaml
sed -i 's|Host: .*#.*MySQL.*|Host: mysql|' deploy/backend/configmap.yaml
sed -i 's|Address: .*#.*Redis.*|Address: redis:6379|' deploy/backend/configmap.yaml

# Step 2: 取消注释 middleware.yaml
sed -i 's|# - middleware.yaml|- middleware.yaml|' deploy/backend/kustomization.yaml

# Step 3: 应用变更
kubectl apply -k deploy/backend/

# Step 4: 等待 MySQL 就绪后初始化数据
kubectl wait --for=condition=ready pod -l app.kubernetes.io/component=mysql -n k8soperation --timeout=120s
# ... 然后执行 SQL 初始化
```

### 从「内部」切换到「外部」

**推荐方式：直接使用 external 目录**
```bash
kubectl apply -k deploy/external/
```

**手动切换方式：**
```bash
# Step 1: 修改 deploy/backend/configmap.yaml 的地址
# 将 Host: mysql 改为外部地址
# 将 Address: redis:6379 改为空，Addresses 填入 Redis Cluster 节点

# Step 2: 修改 deploy/backend/secret.yaml 的密码（base64 编码的外部密码）

# Step 3: 注释 middleware.yaml
sed -i 's|^  - middleware.yaml|  # - middleware.yaml|' deploy/backend/kustomization.yaml

# Step 4: 应用变更（会更新配置，Pod 自动重启）
kubectl apply -k deploy/backend/

# Step 5: 删除集群内已部署的 MySQL/Redis（可选）
kubectl delete deployment mysql redis -n k8soperation
kubectl delete pvc mysql-data redis-data -n k8soperation
```

---

## 配置模板速查

### 外部模板（各云厂商 RDS 地址格式）

| 云厂商 | MySQL 地址示例 | Redis 地址示例 |
|--------|---------------|---------------|
| 阿里云 | `rm-bp1xxx.mysql.rds.aliyuncs.com` | `r-bp1xxx.redis.rds.aliyuncs.com:6379` |
| 腾讯云 | `cdb-xxx.sql.tencentcdb.com` | `crs-xxx.redis.ap-guangzhou.cdb.myqcloud.com:6379` |
| 华为云 | `xxx.rds.cn-north-4.myhuaweicloud.com` | `xxx.dcs.cn-north-4.myhuaweicloud.com:6379` |
| AWS | `xxx.rds.amazonaws.com` | `xxx.cache.amazonaws.com:6379` |
| 自建 | `10.0.1.100` 或 `db.internal.company.com` | `10.0.1.101:6379` |

### 内部模板（K8s Service DNS）

| 服务 | 简短形式（同 namespace） | 完整形式（跨 namespace） |
|------|------------------------|------------------------|
| MySQL | `mysql` | `mysql.k8soperation.svc.cluster.local` |
| Redis | `redis:6379` | `redis.k8soperation.svc.cluster.local:6379` |

---

## 验证方法

### 验证外部数据库连通性

```bash
# 在集群内创建临时 Pod 测试连接
kubectl run test-mysql --rm -it --image=mysql:8.0 -n k8soperation -- \
  mysql -h rm-bp1xxx.mysql.rds.aliyuncs.com -u root -p"your-password" -e "SELECT 1"

kubectl run test-redis --rm -it --image=redis:7-alpine -n k8soperation -- \
  redis-cli -h r-bp1xxx.redis.rds.aliyuncs.com -a "your-password" ping
```

### 验证内部数据库连通性

```bash
# 测试内部 MySQL
kubectl run test-mysql --rm -it --image=mysql:8.0 -n k8soperation -- \
  mysql -h mysql -u root -p"k8s-internal-2024" -e "SHOW DATABASES"

# 测试内部 Redis
kubectl run test-redis --rm -it --image=redis:7-alpine -n k8soperation -- \
  redis-cli -h redis -a "k8s-internal-2024" ping
```

### 验证应用连接状态

```bash
# 查看 Pod 是否 Running
kubectl get pods -n k8soperation

# 查看就绪探针（readinessProbe 会检查 DB 连接）
kubectl describe pod -n k8soperation -l app.kubernetes.io/name=k8soperation | grep -A5 "Readiness"

# 查看应用日志确认连接成功
kubectl logs -n k8soperation -l app.kubernetes.io/name=k8soperation | grep -i "database\|redis\|connected"
```

---

## 常见问题

### Q1: 外部 MySQL 连不上

**排查步骤：**
```bash
# 1. 确认网络可达（从 Pod 内 telnet）
kubectl exec -n k8soperation deploy/k8soperation -- wget -qO- --timeout=3 http://mysql-host:3306 || echo "port check done"

# 2. 确认安全组/防火墙放行了 K8s Node CIDR
# 3. 阿里云 RDS 需要将 K8s Node IP 加入白名单
# 4. 确认用户名密码正确（base64 编码别搞错）
echo "Y2hhbmdlbWU=" | base64 -d    # 解码看看是不是期望的密码
```

### Q2: 内部 MySQL Pod 一直 CrashLoopBackOff

```bash
# 查看 MySQL 日志
kubectl logs -n k8soperation deploy/mysql

# 常见原因：
# 1. PVC 挂载失败 → kubectl describe pvc mysql-data -n k8soperation
# 2. 内存不足 → 适当调大 resources.limits.memory
# 3. 残留旧数据 → 删除 PVC 重建: kubectl delete pvc mysql-data -n k8soperation
```

### Q3: 切换后应用 Pod 不自动重启

ConfigMap 修改后，Pod 不会自动重启，需要手动触发：

```bash
# 方式一：重启 Deployment
kubectl rollout restart deployment/k8soperation -n k8soperation

# 方式二：删掉旧 Pod 让 Deployment 自动拉起新的
kubectl delete pod -n k8soperation -l app.kubernetes.io/name=k8soperation
```

### Q4: 外部数据库需要 SSL 连接

在 `configmap.yaml` 中 Database 段添加：
```yaml
    Database:
      ...
      TLS: true                    # 启用 TLS
      TLSCACert: "/app/certs/ca.pem"  # CA 证书路径
```

然后通过 Secret 或 ConfigMap 挂载 CA 证书文件。

### Q5: Redis Cluster 模式

项目已原生支持 Redis Cluster，无需额外配置：

```yaml
# deploy/external/configmap.yaml 中的配置
Cache:
  Type: redis
  Address: ""                     # 留空
  Addresses:                       # 填写所有 Cluster 节点
    - 192.168.1.201:6379
    - 192.168.1.202:6379
    - 192.168.1.203:6379
    - 192.168.1.204:6379
    - 192.168.1.205:6379
    - 192.168.1.206:6379
```

- `Addresses` 非空时自动启用 `redis.NewClusterClient`
- `Addresses` 为空时使用 `Address` 单节点模式 `redis.NewClient`
- 直接使用 `kubectl apply -k deploy/external/` 即可部署 Redis Cluster 模式

---

## 总结：一张表搞定切换

| 场景 | 推荐命令 | 说明 |
|------|---------|------|
| **外部 MySQL + Redis Cluster** | `kubectl apply -k deploy/external/` | 生产推荐，已预配置 Cluster 模式 |
| **K8s 内 MySQL + Redis** | `kubectl apply -k deploy/backend/` | 开发测试用，取消注释 middleware.yaml |
| **前端** | `kubectl apply -k deploy/frontend/` | 独立前端部署 |
| **全部** | `kubectl apply -k deploy/` | 前后端 + K8s 内中间件 |

### 手动切换对照表

| 切换方向 | backend/configmap.yaml | backend/secret.yaml | backend/kustomization.yaml |
|---------|----------------|-------------|-------------------|
| **用外部** | Host → 外部IP/域名<br>Addresses → Redis Cluster 节点列表 | 填外部密码(base64) | 注释 `middleware.yaml` |
| **用内部** | Host → `mysql`<br>Address → `redis:6379`<br>Addresses → 留空 | 填内部密码(base64) | 取消注释 `middleware.yaml` |

**核心原则：生产环境直接用 `deploy/external/`，免改文件一步到位。**
