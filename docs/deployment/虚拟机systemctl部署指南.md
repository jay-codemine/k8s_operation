# 创建目录
mkdir -p /opt/k8soperation/{bin,configs,storage/logs,storage/artifacts}
mkdir -p /var/www/k8soperation

# 创建运行用户
useradd -r -s /bin/false -d /opt/k8soperation k8sop# K8sOperation 虚拟机部署指南（systemctl 管理）

> 本文档详细描述如何在 Linux 虚拟机上以 systemd 方式部署 K8sOperation 前后端服务，实现开机自启、日志管理和服务治理。

---

## 一、环境要求

| 组件 | 版本要求 | 说明 |
|------|---------|------|
| 操作系统 | CentOS 7+/Ubuntu 20.04+/Debian 11+ | 需要 systemd |
| Go | 1.24+ | 编译后端（编译机，目标机不需要） |
| Node.js | 20.19+ 或 22.12+ | 编译前端（编译机，目标机不需要） |
| Nginx | 1.20+ | 托管前端静态文件 + API 反代 |
| MySQL | 8.0+ | 数据存储 |
| Redis | 7.0+ | Session/缓存 |

### 服务器推荐配置

- CPU：2 核+
- 内存：4GB+
- 磁盘：40GB+
- 网络：开放 80（前端）、8080（后端 API）端口

---

## 二、基础依赖安装

### 2.1 安装 MySQL 8.0

```bash
# CentOS/RHEL
yum install -y https://dev.mysql.com/get/mysql80-community-release-el7-11.noarch.rpm
yum install -y mysql-community-server

# Ubuntu/Debian
apt update && apt install -y mysql-server

# 启动并开机自启
systemctl start mysqld
systemctl enable mysqld
```

**初始化数据库：**

```bash
# 登录 MySQL（首次安装查看临时密码：grep 'temporary password' /var/log/mysqld.log）
mysql -u root -p

# 执行以下 SQL
CREATE DATABASE IF NOT EXISTS `k8s-platform` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER USER 'root'@'localhost' IDENTIFIED BY 'YourStrongPassword123!';
FLUSH PRIVILEGES;
```

**导入初始化表结构：**

```bash
mysql -u root -p'YourStrongPassword123!' k8s-platform < docs/sql/k8s_platform_full_init.sql
```

### 2.2 安装 Redis 7

```bash
# CentOS/RHEL
yum install -y epel-release
yum install -y redis

# Ubuntu/Debian
apt install -y redis-server
```

**配置 Redis 密码：**

```bash
# 编辑 /etc/redis.conf 或 /etc/redis/redis.conf
# 找到并修改（或添加）：
requirepass YourRedisPassword123!
bind 127.0.0.1
appendonly yes
maxmemory 256mb
maxmemory-policy allkeys-lru
```

```bash
systemctl start redis
systemctl enable redis

# 验证连接
redis-cli -a 'YourRedisPassword123!' ping
```

### 2.3 安装 Nginx

```bash
# CentOS/RHEL
yum install -y nginx

# Ubuntu/Debian
apt install -y nginx

systemctl start nginx
systemctl enable nginx
```

---

## 三、编译构建

> 可在本地开发机或 CI 服务器编译，产物传到目标服务器。

### 3.1 编译后端（Go）

```bash
# 在项目根目录
cd /path/to/k8s_operation-main

# 设置环境变量（交叉编译到 Linux amd64）
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export GOPROXY=https://goproxy.cn,direct

# 编译
go build -trimpath -ldflags="-s -w" -o bin/k8soperation ./cmd/k8soperation
```

产物：`bin/k8soperation`（静态链接二进制，约 30-50MB）

### 3.2 编译前端（Vue3）

```bash
cd k8s-web

# 安装依赖
npm ci --registry=https://registry.npmmirror.com

# 构建生产包
npm run build
```

产物：`k8s-web/dist/` 目录（静态 HTML/JS/CSS）

---

## 四、部署目录结构

在目标服务器创建标准化部署目录：

```bash
# 创建部署目录
sudo mkdir -p /opt/k8soperation/{bin,configs,storage/logs,storage/artifacts}
sudo mkdir -p /opt/k8soperation/configs/{jenkins-templates,dockerfile-templates}
sudo mkdir -p /var/www/k8soperation

# 创建运行用户（非 root）
sudo useradd -r -s /bin/false -d /opt/k8soperation k8sop
sudo chown -R k8sop:k8sop /opt/k8soperation
```

最终目录结构：

```
/opt/k8soperation/              # 后端主目录
├── bin/
│   └── k8soperation            # Go 二进制
├── configs/
│   ├── config.yaml             # 主配置文件
│   ├── k8s.yaml                # K8s 集群配置（可选）
│   ├── jenkins-templates/      # Jenkins 流水线模板
│   └── dockerfile-templates/   # Dockerfile 模板
└── storage/
    ├── logs/                   # 运行日志
    └── artifacts/              # 构建产物

/var/www/k8soperation/          # 前端静态文件（Nginx 托管）
└── dist/
    ├── index.html
    ├── assets/
    └── ...
```

---

## 五、传输文件到服务器

```bash
# 在开发机上执行（替换 YOUR_SERVER_IP）
SERVER=root@YOUR_SERVER_IP

# 传输后端二进制
scp bin/k8soperation ${SERVER}:/opt/k8soperation/bin/

# 传输配置文件
scp configs/config.yaml ${SERVER}:/opt/k8soperation/configs/
scp configs/k8s.yaml ${SERVER}:/opt/k8soperation/configs/  # 如有
scp -r configs/jenkins-templates/ ${SERVER}:/opt/k8soperation/configs/
scp -r configs/dockerfile-templates/ ${SERVER}:/opt/k8soperation/configs/

# 传输前端产物
scp -r k8s-web/dist/ ${SERVER}:/var/www/k8soperation/

# 传输 SQL 初始化文件
scp docs/sql/k8s_platform_full_init.sql ${SERVER}:/tmp/
```

---

## 六、服务器端配置

### 6.1 配置后端（config.yaml）

在服务器上编辑 `/opt/k8soperation/configs/config.yaml`：

```yaml
Server:
  RunMode: release
  Port: 8080
  ReadTimeout: 3600
  WriteTimeout: 3600
  IdleTimeout: 300
  ShutdownTimeout: 300

Database:
  DBType: mysql
  Username: root
  Password: "YourStrongPassword123!"
  Host: 127.0.0.1
  Port: 3306
  DBName: k8s-platform
  Charset: utf8mb4
  ParseTime: true
  MaxIdleConns: 10
  MaxOpenConns: 100
  MaxLifeSeconds: 300

Cache:
  Type: redis
  Name: sk_sid
  Address: 127.0.0.1:6379
  Addresses: []
  Username: "default"
  Password: "YourRedisPassword123!"
  MaxConnect: 10
  Network: tcp
  Secret: "k8smana-prod-secret-change-me"

App:
  LogLevel: info
  TIMEZONE: "Asia/Shanghai"
  LogType: single
  LogFileName: /opt/k8soperation/storage/logs/app.log
  LogMaxSize: 100
  LogMaxBackup: 10
  LogMaxAge: 30
  LogCompress: true
  BusinessLogFileName: /opt/k8soperation/storage/logs/biz.log
  MirrorBusinessToSystem: false
  JWTMaxRefreshTime: 86400
  JWTSigningKey: "your-production-jwt-key-at-least-32chars!"
  JWTExpireTime: 120000
  AppName: "k8operation"
  GlobalKubeConfigPath: /opt/k8soperation/configs/k8s.yaml
  DefaultClusterID: 1
  AutoInitK8s: true
  AllowEmptyStart: true

PodLog:
  EnableStreaming: false
  TailDefault: 500
  TailMax: 5000
  LimitBytes: 2097152
  Timestamps: false
  Previous: false

Security:
  KubeConfigEncryptKey: "your-32-char-encryption-key-here!"
  PasswordBcryptCost: 10
  AutoEncryptLegacyData: false

ErrorCode:
  AllowOverride: false

ClusterClient:
  TTL: 30m
  TTLJitter: 3m

Pod:
  eviction:
    default_grace_seconds: 30

Node:
  drain:
    max_grace_seconds: 300
    ignore_daemon_sets: true
    delete_empty_dir: false

Jenkins:
  URL: ""
  Username: ""
  APIToken: ""
  TriggerTimeout: 60
  CallbackURL: ""
  PlatformURL: ""
  HMACSecret: ""
  PollInterval: 15
  MaxBuildTime: 30
  DingTalkWebhook: ""

Monitoring:
  Enabled: false
  PrometheusURL: ""
  QueryTimeout: 30

AIAssistant:
  Enabled: false
  APIKey: ""
  BaseURL: ""
  Model: "gpt-4o"
  MaxTokens: 2048
  Temperature: 0.7
  SystemPrompt: ""
  ApprovalExpire: 30
  MaxHistoryRound: 20

LDAP:
  Enabled: false
```

> **重要**：请修改以下字段为实际值：
> - `Database.Password`
> - `Cache.Password` / `Cache.Secret`
> - `App.JWTSigningKey`
> - `Security.KubeConfigEncryptKey`（必须 32 字符以上）

### 6.2 配置参数必要性说明

以下详细说明每个配置段/字段是否为启动必需，帮助你快速完成最小化部署。

#### 启动必需配置（缺少会导致服务无法启动）

| 配置段 | 必要性 | 说明 |
|--------|--------|------|
| **Server** | ★★★ 必需 | 缺少则启动直接报错。Port 决定监听端口 |
| **Database** | ★★★ 必需 | MySQL 连接信息，启动时会 Ping 测试连通性，连不上直接失败 |
| **Cache** | ★★★ 必需 | Redis 连接信息，Session 初始化依赖 Redis，连不上直接失败 |
| **App** | ★★★ 必需 | 应用核心配置（JWT、日志路径等），缺少则启动报错 |
| **PodLog** | ★★★ 必需 | Pod 日志配置段，缺少则启动报错（可保留默认值） |
| **Pod** | ★★★ 必需 | Pod 驱逐配置段，缺少则启动报错（可保留默认值） |
| **Node** | ★★★ 必需 | Node 排水配置段，缺少则启动报错（可保留默认值） |
| **ErrorCode** | ★★★ 必需 | 错误码配置段，缺少则启动报错（可保留默认值） |
| **ClusterClient** | ★★★ 必需 | K8s Client TTL 配置，缺少则启动报错（可保留默认值） |

#### 可选配置（缺少不影响服务启动）

| 配置段 | 必要性 | 缺省行为 |
|--------|--------|----------|
| **Security** | ☆ 可选 | 缺少时使用内置默认密钥，**但生产环境强烈建议配置** |
| **Jenkins** | ☆ 可选 | 缺少或 URL 为空时，CI/CD 功能不可用，其他正常 |
| **Monitoring** | ☆ 可选 | 缺少或 Enabled=false 时，监控功能不可用 |
| **AIAssistant** | ☆ 可选 | 缺少或 Enabled=false 时，AI 助手功能不可用 |
| **LDAP** | ☆ 可选 | 缺少或 Enabled=false 时，使用本地账号认证 |
| **PlatformSettings** | ☆ 可选 | 缺少时使用程序内置默认值 |

#### 各配置段字段详细说明

##### Server（必需）

| 字段 | 必填 | 默认建议值 | 说明 |
|------|------|-----------|------|
| RunMode | 是 | `release` | 生产用 release，开发用 debug |
| Port | 是 | `8080` | API 监听端口 |
| ReadTimeout | 是 | `3600` | 读超时（秒），WebSocket 需要较大值 |
| WriteTimeout | 是 | `3600` | 写超时（秒） |
| IdleTimeout | 是 | `300` | 空闲超时（秒） |
| ShutdownTimeout | 是 | `300` | 优雅关闭超时（秒） |

##### Database（必需 - 必须连通真实 MySQL）

| 字段 | 必填 | 说明 |
|------|------|------|
| DBType | 是 | 固定 `mysql` |
| Username | 是 | 数据库用户名 |
| **Password** | **是** | **数据库密码（必须修改为实际值）** |
| **Host** | **是** | **数据库地址（必须修改为实际值）** |
| **Port** | **是** | **数据库端口（默认 3306）** |
| **DBName** | **是** | **数据库名（需提前创建）** |
| Charset | 是 | 建议 `utf8mb4` |
| ParseTime | 是 | 固定 `true` |
| MaxIdleConns | 是 | 连接池空闲数，建议 5-10 |
| MaxOpenConns | 是 | 连接池最大数，建议 50-100 |
| MaxLifeSeconds | 是 | 连接最大存活（秒） |

##### Cache（必需 - 必须连通真实 Redis）

| 字段 | 必填 | 说明 |
|------|------|------|
| Type | 是 | 固定 `redis` |
| Name | 是 | Session 名称前缀 |
| **Address** | **是** | **Redis 地址（host:port）** |
| Addresses | 否 | Cluster 模式节点列表，单节点留 `[]` |
| **Username** | **是** | **不能为空！Redis 6+ 用 `"default"`** |
| **Password** | **是** | **不能为空！对应 Redis 的 requirepass** |
| MaxConnect | 是 | Session Store 连接池大小 |
| Network | 是 | 固定 `tcp` |
| **Secret** | **是** | **Session Cookie 加密密钥，建议随机字符串** |

> ⚠️ **特别注意**：`Username` 和 `Password` 字段代码中强制校验非空，为空会直接导致启动失败并报错 `redis username is empty` 或 `redis password is empty`。

##### App（必需）

| 字段 | 必填 | 可用默认值 | 说明 |
|------|------|-----------|------|
| LogLevel | 是 | `info` | 日志级别：debug/info/warn/error |
| TIMEZONE | 是 | `Asia/Shanghai` | 时区 |
| LogType | 是 | `single` | 日志类型 |
| LogFileName | 是 | 见模板 | 日志文件路径 |
| LogMaxSize | 是 | `100` | 单日志文件最大 MB |
| LogMaxBackup | 是 | `10` | 保留备份数 |
| LogMaxAge | 是 | `30` | 保留天数 |
| LogCompress | 是 | `true` | 是否压缩归档 |
| BusinessLogFileName | 是 | 见模板 | 业务日志路径 |
| MirrorBusinessToSystem | 否 | `false` | 业务日志是否镜像到系统日志 |
| **JWTSigningKey** | **是** | **必须修改** | **JWT 签名密钥，至少 32 字符** |
| JWTExpireTime | 是 | `120000` | JWT 过期时间（秒） |
| JWTMaxRefreshTime | 是 | `86400` | JWT 最大刷新时间（秒） |
| AppName | 是 | `k8operation` | 应用名 |
| GlobalKubeConfigPath | 否 | `configs/k8s.yaml` | 本地 kubeconfig 路径（可不存在） |
| DefaultClusterID | 否 | `1` | 默认集群 ID |
| AutoInitK8s | 否 | `true` | 是否自动初始化 K8s |
| **AllowEmptyStart** | **建议** | `true` | **设为 true 允许无 K8s 集群时启动** |

##### PodLog / Pod / Node / ErrorCode / ClusterClient（必需但可用默认值）

这些配置段在 YAML 文件中**必须存在**，否则启动报错。但内部字段可以保持示例中的默认值，无需修改。

```yaml
# 以下配置段直接复制即可，无需修改
PodLog:
  EnableStreaming: false
  TailDefault: 500
  TailMax: 5000
  LimitBytes: 2097152
  Timestamps: false
  Previous: false

ErrorCode:
  AllowOverride: false

ClusterClient:
  TTL: 30m
  TTLJitter: 3m

Pod:
  eviction:
    default_grace_seconds: 30

Node:
  drain:
    max_grace_seconds: 300
    ignore_daemon_sets: true
    delete_empty_dir: false
```

##### Security（可选但强烈建议）

| 字段 | 必填 | 说明 |
|------|------|------|
| KubeConfigEncryptKey | 否 | KubeConfig 加密密钥，缺省用内置默认值。**创建集群时必须有此配置** |
| PasswordBcryptCost | 否 | BCrypt 加密强度，默认 10 |
| AutoEncryptLegacyData | 否 | 是否自动加密历史明文数据 |

> ⚠️ 虽然 Security 段缺少不会阻止启动，但如果不配置 `KubeConfigEncryptKey`，**通过界面创建/添加 K8s 集群时会使用内置默认密钥**，存在安全风险。生产环境必须配置。

##### Jenkins（可选）

| 字段 | 条件必填 | 说明 |
|------|---------|------|
| URL | 启用时必填 | Jenkins 服务地址，为空则 CI/CD 功能整体禁用 |
| Username | 启用时必填 | Jenkins 用户名 |
| APIToken | 启用时必填 | Jenkins API Token |
| CallbackURL | 启用时需要 | 后端回调地址（Jenkins 构建完回调） |
| PlatformURL | 启用时需要 | 前端地址（通知链接） |
| 其他字段 | 否 | 可保留默认值 |

> 💡 如果暂时不需要 CI/CD 功能，**整个 Jenkins 段可以留空或设 URL 为空字符串**。

##### Monitoring（可选）

| 字段 | 条件必填 | 说明 |
|------|---------|------|
| Enabled | 否 | `false` 则监控功能完全禁用 |
| PrometheusURL | 启用时必填 | Prometheus 查询地址 |
| QueryTimeout | 否 | 查询超时，默认 30 秒 |

> 💡 不接 Prometheus 就设 `Enabled: false`，不影响任何其他功能。

##### AIAssistant（可选）

| 字段 | 条件必填 | 说明 |
|------|---------|------|
| Enabled | 否 | `false` 则 AI 助手完全禁用 |
| APIKey | 启用时必填 | OpenAI 或兼容 API 的 Key |
| BaseURL | 否 | 自定义 API 地址（国内代理） |
| Model | 否 | 默认 `gpt-4o` |
| 其他字段 | 否 | 可保留默认值 |

> 💡 AI 功能可后续开启，初始部署可设 `Enabled: false`。

##### LDAP（可选）

| 字段 | 条件必填 | 说明 |
|------|---------|------|
| Enabled | 否 | `false` 则使用本地账号认证 |
| Host/Port/BindDN 等 | 启用时必填 | LDAP 服务器连接信息 |

> 💡 不需要 LDAP 统一认证就设 `Enabled: false`，使用内置账号体系即可。

---

### 6.3 最小化配置模板（快速启动）

如果只需要最快启动服务进行验证，以下是**最小必需配置**：

```yaml
Server:
  RunMode: release
  Port: 8080
  ReadTimeout: 3600
  WriteTimeout: 3600
  IdleTimeout: 300
  ShutdownTimeout: 300

Database:
  DBType: mysql
  Username: root
  Password: "your_mysql_password"
  Host: 127.0.0.1
  Port: 3306
  DBName: k8s-platform
  Charset: utf8mb4
  ParseTime: true
  MaxIdleConns: 5
  MaxOpenConns: 100
  MaxLifeSeconds: 300

Cache:
  Type: redis
  Name: sk_sid
  Address: 127.0.0.1:6379
  Addresses: []
  Username: "default"
  Password: "your_redis_password"
  MaxConnect: 10
  Network: tcp
  Secret: "random-session-secret"

App:
  LogLevel: info
  TIMEZONE: "Asia/Shanghai"
  LogType: single
  LogFileName: /opt/k8soperation/storage/logs/app.log
  LogMaxSize: 100
  LogMaxBackup: 10
  LogMaxAge: 30
  LogCompress: true
  BusinessLogFileName: /opt/k8soperation/storage/logs/biz.log
  MirrorBusinessToSystem: false
  JWTMaxRefreshTime: 86400
  JWTSigningKey: "change-me-to-random-32-char-key!"
  JWTExpireTime: 120000
  AppName: "k8operation"
  GlobalKubeConfigPath: /opt/k8soperation/configs/k8s.yaml
  DefaultClusterID: 1
  AutoInitK8s: true
  AllowEmptyStart: true

PodLog:
  EnableStreaming: false
  TailDefault: 500
  TailMax: 5000
  LimitBytes: 2097152
  Timestamps: false
  Previous: false

Security:
  KubeConfigEncryptKey: "change-me-32-char-encrypt-key!!"
  PasswordBcryptCost: 10
  AutoEncryptLegacyData: false

ErrorCode:
  AllowOverride: false

ClusterClient:
  TTL: 30m
  TTLJitter: 3m

Pod:
  eviction:
    default_grace_seconds: 30

Node:
  drain:
    max_grace_seconds: 300
    ignore_daemon_sets: true
    delete_empty_dir: false

# === 以下全部可选，不配不影响启动 ===

Jenkins:
  URL: ""

Monitoring:
  Enabled: false

AIAssistant:
  Enabled: false

LDAP:
  Enabled: false
```

> **总结：最小启动只需要**：
> 1. 一个可连通的 MySQL（已建库 `k8s-platform`）
> 2. 一个可连通的 Redis（配置了密码）
> 3. 以上配置文件（约 60 行）
>
> K8s 集群、Jenkins、Prometheus、AI、LDAP 都可以后续通过界面或修改配置添加。

### 6.4 设置文件权限

```bash
# 设置二进制可执行
sudo chmod +x /opt/k8soperation/bin/k8soperation

# 设置目录归属
sudo chown -R k8sop:k8sop /opt/k8soperation
sudo chown -R nginx:nginx /var/www/k8soperation

# 配置文件严格权限（仅 owner 可读）
sudo chmod 600 /opt/k8soperation/configs/config.yaml
```

---

## 七、Systemd 服务配置

### 7.1 后端服务（k8soperation.service）

创建文件 `/etc/systemd/system/k8soperation.service`：

```ini
[Unit]
Description=K8sOperation Backend API Server
Documentation=https://gitee.com/jay-kim/k8s_operation
After=network.target mysql.service redis.service
Wants=mysql.service redis.service

[Service]
Type=simple
User=k8sop
Group=k8sop

# 工作目录
WorkingDirectory=/opt/k8soperation

# 环境变量
Environment=GIN_MODE=release
Environment=APP_CONFIG=/opt/k8soperation/configs/config.yaml

# 启动命令
ExecStart=/opt/k8soperation/bin/k8soperation

# 优雅停止（发送 SIGTERM，程序内部 graceful shutdown）
ExecStop=/bin/kill -TERM $MAINPID
TimeoutStopSec=30

# 自动重启策略
Restart=on-failure
RestartSec=5
StartLimitIntervalSec=60
StartLimitBurst=3

# 资源限制
LimitNOFILE=65536
LimitNPROC=65536

# 日志输出到 journald
StandardOutput=journal
StandardError=journal
SyslogIdentifier=k8soperation

[Install]
WantedBy=multi-user.target
```

### 7.2 启用并启动后端服务

```bash
# 重载 systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start k8soperation

# 开机自启
sudo systemctl enable k8soperation

# 查看状态
sudo systemctl status k8soperation

# 查看日志
sudo journalctl -u k8soperation -f
```

---

## 八、Nginx 配置（前端 + API 反代）

### 8.1 创建 Nginx 站点配置

创建文件 `/etc/nginx/conf.d/k8soperation.conf`：

```nginx
# ============================================================
# K8sOperation - Nginx 配置
# 前端静态文件 + 后端 API 反向代理
# ============================================================

upstream k8sop_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name _;  # 替换为实际域名，如 devops.example.com

    root /var/www/k8soperation/dist;
    index index.html;

    # ==================== Gzip 压缩 ====================
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/javascript
        application/json
        application/javascript
        application/xml
        image/svg+xml;

    # ==================== 静态资源缓存 ====================
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # ==================== API 反向代理 ====================
    location /api/ {
        proxy_pass http://k8sop_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 支持（容器终端、实时日志等）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;

        # 大文件上传支持
        client_max_body_size 200m;
    }

    # ==================== Swagger 文档代理 ====================
    location /swagger/ {
        proxy_pass http://k8sop_backend;
        proxy_set_header Host $host;
    }

    # ==================== 健康检查 ====================
    location /health {
        access_log off;
        return 200 "ok";
        add_header Content-Type text/plain;
    }

    # ==================== Vue Router History 模式 ====================
    location / {
        try_files $uri $uri/ /index.html;
    }

    # ==================== 安全头 ====================
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
```

### 8.2 验证并启动 Nginx

```bash
# 测试配置
sudo nginx -t

# 重载配置
sudo systemctl reload nginx

# 确认开机自启
sudo systemctl enable nginx
```

---

## 九、防火墙配置

```bash
# firewalld（CentOS/RHEL）
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp  # 可选，直接访问 API
sudo firewall-cmd --reload

# ufw（Ubuntu/Debian）
sudo ufw allow 80/tcp
sudo ufw allow 8080/tcp  # 可选
sudo ufw reload
```

---

## 十、服务管理命令

### 10.1 后端服务管理

```bash
# 启动
sudo systemctl start k8soperation

# 停止
sudo systemctl stop k8soperation

# 重启
sudo systemctl restart k8soperation

# 查看状态
sudo systemctl status k8soperation

# 查看实时日志
sudo journalctl -u k8soperation -f

# 查看最近 100 行日志
sudo journalctl -u k8soperation -n 100

# 查看今天的日志
sudo journalctl -u k8soperation --since today
```

### 10.2 前端（Nginx）管理

```bash
# 重载配置（不中断连接）
sudo systemctl reload nginx

# 完全重启
sudo systemctl restart nginx

# 查看状态
sudo systemctl status nginx
```

### 10.3 全部服务一键管理

```bash
# 一键停止所有服务
sudo systemctl stop k8soperation nginx

# 一键启动所有服务
sudo systemctl start mysql redis nginx k8soperation

# 查看所有相关服务状态
sudo systemctl status mysql redis nginx k8soperation --no-pager
```

---

## 十一、版本升级流程

### 11.1 后端升级

```bash
# 1. 在开发机编译新版本
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/k8soperation ./cmd/k8soperation

# 2. 传输到服务器
scp bin/k8soperation root@YOUR_SERVER_IP:/opt/k8soperation/bin/k8soperation.new

# 3. 在服务器上执行替换并重启
ssh root@YOUR_SERVER_IP << 'EOF'
cd /opt/k8soperation/bin
mv k8soperation k8soperation.bak
mv k8soperation.new k8soperation
chmod +x k8soperation
chown k8sop:k8sop k8soperation
systemctl restart k8soperation
systemctl status k8soperation --no-pager
EOF
```

### 11.2 前端升级

```bash
# 1. 在开发机编译
cd k8s-web && npm run build

# 2. 传输并替换
scp -r k8s-web/dist/ root@YOUR_SERVER_IP:/var/www/k8soperation/dist.new

# 3. 在服务器上执行替换
ssh root@YOUR_SERVER_IP << 'EOF'
cd /var/www/k8soperation
rm -rf dist.bak
mv dist dist.bak
mv dist.new dist
chown -R nginx:nginx dist
# Nginx 无需重启，静态文件直接生效
EOF
```

### 11.3 回滚

```bash
# 后端回滚
ssh root@YOUR_SERVER_IP << 'EOF'
cd /opt/k8soperation/bin
mv k8soperation k8soperation.failed
mv k8soperation.bak k8soperation
systemctl restart k8soperation
EOF

# 前端回滚
ssh root@YOUR_SERVER_IP << 'EOF'
cd /var/www/k8soperation
rm -rf dist
mv dist.bak dist
EOF
```

---

## 十二、日志管理

### 12.1 日志文件位置

| 日志 | 路径 | 说明 |
|------|------|------|
| 后端系统日志 | `/opt/k8soperation/storage/logs/app.log` | 应用主日志 |
| 后端业务日志 | `/opt/k8soperation/storage/logs/biz.log` | 业务操作日志 |
| systemd 日志 | `journalctl -u k8soperation` | 服务启停/异常日志 |
| Nginx 访问日志 | `/var/log/nginx/access.log` | HTTP 访问日志 |
| Nginx 错误日志 | `/var/log/nginx/error.log` | Nginx 错误日志 |

### 12.2 日志轮转

后端日志由程序内置 lumberjack 自动轮转（配置在 config.yaml 的 App 段）。

Nginx 日志使用系统 logrotate，通常已自动配置。可确认：

```bash
cat /etc/logrotate.d/nginx
```

---

## 十三、健康检查与监控

### 13.1 健康检查端点

```bash
# 后端存活检查
curl http://127.0.0.1:8080/healthz/live

# 前端健康检查（通过 Nginx）
curl http://127.0.0.1/health
```

### 13.2 简单监控脚本

创建 `/opt/k8soperation/scripts/healthcheck.sh`：

```bash
#!/bin/bash
# K8sOperation 健康检查脚本

BACKEND_URL="http://127.0.0.1:8080/healthz/live"
FRONTEND_URL="http://127.0.0.1/health"

check_service() {
    local name=$1
    local url=$2
    local response=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 "$url")
    if [ "$response" = "200" ]; then
        echo "[OK] $name is healthy"
    else
        echo "[FAIL] $name is unhealthy (HTTP $response)"
        return 1
    fi
}

echo "===== K8sOperation Health Check ====="
echo "Time: $(date)"
echo ""
check_service "Backend API" "$BACKEND_URL"
check_service "Frontend Web" "$FRONTEND_URL"
echo ""
echo "===== Service Status ====="
systemctl is-active --quiet k8soperation && echo "[OK] k8soperation service active" || echo "[FAIL] k8soperation service inactive"
systemctl is-active --quiet nginx && echo "[OK] nginx service active" || echo "[FAIL] nginx service inactive"
systemctl is-active --quiet mysqld 2>/dev/null || systemctl is-active --quiet mysql 2>/dev/null && echo "[OK] mysql service active" || echo "[FAIL] mysql service inactive"
systemctl is-active --quiet redis 2>/dev/null || systemctl is-active --quiet redis-server 2>/dev/null && echo "[OK] redis service active" || echo "[FAIL] redis service inactive"
```

添加到 crontab 定时检查：

```bash
chmod +x /opt/k8soperation/scripts/healthcheck.sh

# 每 5 分钟检查一次
echo "*/5 * * * * /opt/k8soperation/scripts/healthcheck.sh >> /opt/k8soperation/storage/logs/healthcheck.log 2>&1" | crontab -
```

---

## 十四、一键部署脚本

创建 `/opt/k8soperation/scripts/deploy.sh`（在开发机执行）：

```bash
#!/bin/bash
# ============================================================
# K8sOperation 一键部署脚本
# 用法: ./deploy.sh <server_ip> [ssh_user]
# 示例: ./deploy.sh 1.117.227.207 root
# ============================================================

set -e

SERVER_IP=${1:?"用法: $0 <server_ip> [ssh_user]"}
SSH_USER=${2:-root}
SSH_TARGET="${SSH_USER}@${SERVER_IP}"

REMOTE_APP_DIR="/opt/k8soperation"
REMOTE_WEB_DIR="/var/www/k8soperation"

echo "===== K8sOperation 部署 ====="
echo "目标服务器: ${SSH_TARGET}"
echo ""

# --- 1. 编译后端 ---
echo "[1/5] 编译后端..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/k8soperation ./cmd/k8soperation
echo "      后端编译完成: bin/k8soperation"

# --- 2. 编译前端 ---
echo "[2/5] 编译前端..."
cd k8s-web
npm ci --registry=https://registry.npmmirror.com --silent
npm run build
cd ..
echo "      前端编译完成: k8s-web/dist/"

# --- 3. 传输文件 ---
echo "[3/5] 传输文件到服务器..."
scp bin/k8soperation ${SSH_TARGET}:${REMOTE_APP_DIR}/bin/k8soperation.new
scp -r configs/jenkins-templates/ ${SSH_TARGET}:${REMOTE_APP_DIR}/configs/
scp -r configs/dockerfile-templates/ ${SSH_TARGET}:${REMOTE_APP_DIR}/configs/
scp -r k8s-web/dist ${SSH_TARGET}:${REMOTE_WEB_DIR}/dist.new
echo "      文件传输完成"

# --- 4. 远程替换并重启 ---
echo "[4/5] 远程部署..."
ssh ${SSH_TARGET} << 'REMOTE_EOF'
set -e

# 替换后端
cd /opt/k8soperation/bin
[ -f k8soperation ] && mv k8soperation k8soperation.bak
mv k8soperation.new k8soperation
chmod +x k8soperation
chown k8sop:k8sop k8soperation

# 替换前端
cd /var/www/k8soperation
[ -d dist ] && rm -rf dist.bak && mv dist dist.bak
mv dist.new dist
chown -R nginx:nginx dist

# 重启后端
systemctl restart k8soperation
REMOTE_EOF
echo "      远程部署完成"

# --- 5. 验证 ---
echo "[5/5] 等待服务启动..."
sleep 3
ssh ${SSH_TARGET} "systemctl status k8soperation --no-pager && curl -sf http://127.0.0.1:8080/healthz/live && echo ' Backend OK'"
echo ""
echo "===== 部署成功！====="
echo "访问地址: http://${SERVER_IP}"
```

---

## 十五、常见问题排查

### Q1: 后端启动失败 - "redis username is empty"

**原因**：`config.yaml` 中 `Cache.Username` 为空  
**解决**：设置为 `"default"`（Redis 6+ ACL 默认用户）

### Q2: 创建集群失败 - "global crypto service not initialized"

**原因**：`config.yaml` 缺少 `Security` 配置段  
**解决**：添加 Security 配置并确保 `KubeConfigEncryptKey` 非空且至少 16 字符

### Q3: 后端启动失败 - "bind: address already in use"

**原因**：8080 端口被占用  
**解决**：
```bash
# 查看占用进程
lsof -i:8080
# 或杀死进程后重启
fuser -k 8080/tcp
systemctl start k8soperation
```

### Q4: Nginx 502 Bad Gateway

**原因**：后端服务未启动或崩溃  
**解决**：
```bash
systemctl status k8soperation
journalctl -u k8soperation -n 50
# 确认后端正常后重试
```

### Q5: Redis 认证失败 - "WRONGPASS"

**原因**：Redis 版本 < 6 不支持 ACL，或密码配置不匹配  
**解决**：
- 确认 Redis 版本 `redis-server --version`
- 确认 `requirepass` 与 `config.yaml` 中 `Cache.Password` 一致
- Redis 6+ 且未配置 ACL 时，Username 用 `"default"`

### Q6: 权限问题 - "permission denied"

**解决**：
```bash
sudo chown -R k8sop:k8sop /opt/k8soperation
sudo chmod +x /opt/k8soperation/bin/k8soperation
sudo chmod 600 /opt/k8soperation/configs/config.yaml
```

---

## 十六、HTTPS 配置（可选）

如需启用 HTTPS，使用 Let's Encrypt：

```bash
# 安装 certbot
sudo yum install -y certbot python3-certbot-nginx  # CentOS
# 或
sudo apt install -y certbot python3-certbot-nginx  # Ubuntu

# 自动获取证书并配置 Nginx
sudo certbot --nginx -d devops.example.com

# 自动续期（certbot 默认已配置）
sudo systemctl enable certbot-renew.timer
```

---

## 附录：完整部署清单

```
□ 1. 安装 MySQL 8.0 并初始化数据库
□ 2. 安装 Redis 7 并配置密码
□ 3. 安装 Nginx
□ 4. 创建部署目录和运行用户
□ 5. 编译后端二进制
□ 6. 编译前端静态文件
□ 7. 传输文件到服务器
□ 8. 配置 config.yaml（修改数据库/Redis/密钥）
□ 9. 创建 systemd service 文件
□ 10. 启动并启用后端服务
□ 11. 配置 Nginx 站点
□ 12. 测试并重载 Nginx
□ 13. 配置防火墙规则
□ 14. 验证服务健康
□ 15. （可选）配置 HTTPS
□ 16. （可选）配置定时健康检查
```
