# 📄 K8sOperation 完整部署文档（一站式指南）

> 本文档是 **K8sOperation 平台**（Go + Vue3 + K8s 多集群管理 + AI 助手 + CI/CD）的**完整端到端部署指南**，覆盖从「环境准备 → 数据库初始化 → 后端启动 → 前端启动 → Docker / Kubernetes 部署 → 验证 → 排错」的全流程，**新用户照着抄一遍即可跑起来**。

> 🎯 想 5 分钟跑起来？请直接看根目录 [`QUICK_START.md`](../QUICK_START.md)
> 🎯 想直接部署到生产 K8s 集群？请看 [`docs/K8s部署指南.md`](./K8s部署指南.md)
> 🎯 本文档 = **完整版**，三种模式都覆盖、含所有细节与排错

---

## 📑 目录

- [一、平台与组件总览](#一平台与组件总览)
- [二、环境要求](#二环境要求)
- [三、第一次部署（强烈推荐：一键脚本）](#三第一次部署强烈推荐一键脚本)
- [四、手动分步部署](#四手动分步部署)
  - [4.1 拉取代码](#41-拉取代码)
  - [4.2 准备 MySQL 与 Redis](#42-准备-mysql-与-redis)
  - [4.3 初始化数据库（唯一脚本）](#43-初始化数据库唯一脚本)
  - [4.4 准备配置文件](#44-准备配置文件)
  - [4.5 编译并启动后端](#45-编译并启动后端)
  - [4.6 启动前端](#46-启动前端)
- [五、Docker 单机部署](#五docker-单机部署)
- [六、Kubernetes 生产部署](#六kubernetes-生产部署)
- [七、可选组件接入](#七可选组件接入)
  - [7.1 Jenkins（CI/CD）](#71-jenkinscicd)
  - [7.2 Prometheus / Loki（监控与日志）](#72-prometheus--loki监控与日志)
  - [7.3 SonarQube（代码扫描）](#73-sonarqube代码扫描)
  - [7.4 AI 大模型](#74-ai-大模型)
- [八、部署后验证](#八部署后验证)
- [九、常见问题排错](#九常见问题排错)
- [十、运维操作清单](#十运维操作清单)
- [十一、目录与脚本对照表](#十一目录与脚本对照表)

---

## 一、平台与组件总览

```
┌────────────────────────────────────────────────────────────────────┐
│                      K8sOperation 部署组件视图                      │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  浏览器 ─▶ k8s-web (Vue3 + Vite, :5173/80)                          │
│              │ /api/v1/*                                            │
│              ▼                                                      │
│  k8soperation 后端 (Go, :8080)                                      │
│   ├─▶ MySQL  :3306  ── 主数据存储（50+ 张表）                       │
│   ├─▶ Redis  :6379  ── Session / 缓存 / 消息队列                    │
│   ├─▶ K8s API ────── 多集群管理（client-go）                       │
│   ├─▶ Jenkins ────── （可选）CI/CD 引擎                            │
│   ├─▶ Prometheus ──  （可选）监控指标                              │
│   ├─▶ Loki ────────  （可选）日志检索                              │
│   ├─▶ SonarQube ───  （可选）代码扫描                              │
│   └─▶ LLM (OpenAI 等) （可选）AI 智能助手                          │
└────────────────────────────────────────────────────────────────────┘
```

**最小可运行组合**：MySQL + Redis + 后端 + 前端 → 即可登录 / 管理 / RBAC / 制品库等核心功能可用。
K8s / Jenkins / Prometheus / Loki / AI **均为可选模块**，不配置时**对应功能优雅降级**，不影响主流程。

---

## 二、环境要求

### 2.1 必备依赖

| 组件 | 最低版本 | 推荐版本 | 用途 |
|------|---------|----------|------|
| **Go** | 1.21+ | **1.24.x** | 后端编译 |
| **Node.js** | 18+ | **20.x** | 前端构建（仅前端开发/构建机器需要） |
| **MySQL** | 5.7+ | **8.0.33+** | 主数据库 |
| **Redis** | 5.0+ | **7.x** | 会话 / 缓存 / 队列 |

### 2.2 可选依赖

| 组件 | 用途 | 不装时影响 |
|------|------|-----------|
| Docker 20+ | 容器化部署 | 仅影响 Docker/K8s 路径 |
| kubectl 1.25+ | 管理 K8s 集群 | 不影响后端启动，影响"集群管理"模块 |
| Jenkins 2.346+ | CI/CD 流水线 | 影响"流水线发布"模块 |
| Prometheus / VictoriaMetrics | 监控指标 | 影响"监控总览"模块 |
| Loki + Promtail | 日志聚合 | 影响"日志探索"模块 |
| SonarQube 9+ | 代码质量扫描 | 影响"代码扫描"模块 |
| Harbor / 阿里云 ACR | 镜像仓库 | 影响"镜像管理"模块 |

### 2.3 端口占用清单

| 端口 | 进程 | 是否可改 |
|------|------|---------|
| 8080 | Go 后端 | `configs/config.yaml → Server.Port` |
| 5173 | Vite 前端开发 | `k8s-web/vite.config.js` |
| 80 / 443 | Nginx 前端生产 | Nginx 配置 |
| 3306 | MySQL | 改后端 `Database.Port` |
| 6379 | Redis | 改后端 `Cache.Address` |
| 8081 | Swagger UI | `make swagger-ui` 自定义 |

### 2.4 默认凭据（**部署后必须改**）

| 项 | 默认值 |
|----|--------|
| 平台管理员 | `admin` / `admin123` |
| MySQL root | `root` / `admin123` |
| Redis | （无用户） / `admin123` |
| JWT Signing Key | `configs/config.yaml.example` 中给出的占位值 |

---

## 三、第一次部署（强烈推荐：一键脚本）

项目内置跨平台一键部署脚本，**自动完成「环境检查 → 库表初始化 → 配置生成 → 编译 → 启动」**。

### 3.1 Linux / macOS

```bash
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation

chmod +x scripts/*.sh

# 仅检查环境（不做任何修改），可选
bash scripts/check-env.sh

# 一键启动
bash scripts/quick-start.sh
```

### 3.2 Windows（PowerShell）

```powershell
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation

# 仅检查环境
powershell -ExecutionPolicy Bypass -File scripts\check-env.ps1

# 一键启动
powershell -ExecutionPolicy Bypass -File scripts\quick-start.ps1
```

### 3.3 自定义连接信息

通过环境变量覆盖默认值：

```bash
# Linux / macOS
DB_HOST=10.0.0.100 DB_USER=root DB_PASS=YourPwd \
REDIS_HOST=10.0.0.100 REDIS_PASS=YourPwd \
bash scripts/quick-start.sh
```

```powershell
# Windows
$env:DB_HOST="10.0.0.100"; $env:DB_PASS="YourPwd"
$env:REDIS_HOST="10.0.0.100"; $env:REDIS_PASS="YourPwd"
.\scripts\quick-start.ps1
```

完整环境变量见 [脚本工具一览](#1112-环境变量一览)。

---

## 四、手动分步部署

适合需要精细控制每一步、或一键脚本失败时手动排错。

### 4.1 拉取代码

```bash
# Gitee 主仓库
git clone https://gitee.com/jay-kim/k8s_operation.git
# 或 GitHub 镜像仓库
git clone https://github.com/jay-codemine/k8s_operation.git

cd k8s_operation
```

### 4.2 准备 MySQL 与 Redis

#### 方式 A：Docker（最快，推荐开发使用）

```bash
# MySQL 8.0
docker run -d --name k8s-mysql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=admin123 \
  -v k8s-mysql-data:/var/lib/mysql \
  mysql:8.0 \
  --default-authentication-plugin=mysql_native_password \
  --character-set-server=utf8mb4 \
  --collation-server=utf8mb4_unicode_ci

# Redis 7
docker run -d --name k8s-redis \
  -p 6379:6379 \
  redis:7 redis-server --requirepass admin123
```

#### 方式 B：本地安装

```bash
# Ubuntu / Debian
sudo apt install -y mysql-server redis-server
sudo systemctl enable --now mysql redis-server

# CentOS / RHEL
sudo yum install -y mysql-server redis
sudo systemctl enable --now mysqld redis

# macOS
brew install mysql redis
brew services start mysql
brew services start redis
```

> ⚠️ MySQL 必须使用 **utf8mb4 + utf8mb4_unicode_ci**，否则中文备注、emoji 等会乱码。

### 4.3 初始化数据库（唯一脚本）

> 🎯 **整个项目只需要一个 SQL 脚本**：`docs/sql/k8s_platform_full_init.sql`
> 该脚本同时支持「全新部署」和「存量升级」（基于 `information_schema` 的幂等 ALTER 兜底，**可重复执行**）。

#### 方式 A：使用项目脚本（推荐）

```bash
# Linux / macOS
bash scripts/init-db.sh

# Windows
powershell -ExecutionPolicy Bypass -File scripts\init-db.ps1
```

支持「安全追加」和「删除重建」两种模式，交互式选择。

#### 方式 B：一行命令

```bash
mysql -h 127.0.0.1 -P 3306 -u root -padmin123 \
  --default-character-set=utf8mb4 \
  < docs/sql/k8s_platform_full_init.sql
```

#### 方式 C：进入 MySQL 交互模式

```sql
-- 1. 创建数据库（脚本会自动 CREATE，但先建好也可以）
CREATE DATABASE IF NOT EXISTS `k8s-platform`
  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. 执行初始化
USE `k8s-platform`;
SOURCE docs/sql/k8s_platform_full_init.sql;

-- 3. 验证
SHOW TABLES;                       -- 应返回 50+ 张表
SELECT id, username FROM sys_admin WHERE username = 'admin';
SHOW COLUMNS FROM monitor_datasource LIKE 'cluster_id';   -- 应有结果
```

#### 脚本包含的内容

| 类别 | 内容 |
|------|------|
| **DDL** | 50+ 张业务表：用户/角色/RBAC/CICD/制品/监控/AI/应用商城/构建探针 等 |
| **字典数据** | 18 个资源模板 + 5 条环境规则 + 4 个流水线模板 + 系统角色 |
| **演示业务数据** | 4 流水线 + 6 流水线运行 + 11 阶段 + 4 制品 + 2 监控数据源 |
| **默认账号** | `admin / admin123`（首次登录后请立即在「平台管理 → 用户中心」修改密码） |
| **幂等保证** | 所有 INSERT 用 `INSERT IGNORE`，所有兜底 ALTER 走 `information_schema` 判断 |

> 📚 SQL 脚本细节请见 [`docs/sql/README.md`](./sql/README.md)

### 4.4 准备配置文件

```bash
# 1. 主配置（数据库 / Redis / 日志 / JWT 等）
cp configs/config.yaml.example configs/config.yaml

# 2. K8s 集群 KubeConfig（可选，不连 K8s 可跳过）
cp configs/k8s.yaml.example configs/k8s.yaml
```

#### 关键配置项（`configs/config.yaml`）

```yaml
Server:
  RunMode: debug              # 开发 debug，生产 release
  Port: 8080                  # 后端监听端口

Database:
  DBType: mysql
  Username: root
  Password: admin123          # ⚠️ 改成你的真实密码
  Host: 127.0.0.1
  Port: 3306
  DBName: k8s-platform        # ⚠️ 与脚本中的库名保持一致
  Charset: utf8mb4
  ParseTime: true
  MaxIdleConns: 5
  MaxOpenConns: 100

Cache:
  Type: redis
  Address: 127.0.0.1:6379
  Username: ""
  Password: "admin123"        # ⚠️ Redis 密码

App:
  LogLevel: debug             # 生产建议 info
  TIMEZONE: "Asia/Shanghai"
  LogFileName: storage/logs/app.log         # 本地相对路径；K8s 部署时改绝对路径
  BusinessLogFileName: storage/logs/biz.log
  JWTSigningKey: "请改成 32 位随机串"        # ⚠️ 生产环境必须改
  JWTExpireTime: 120000
  GlobalKubeConfigPath: configs/k8s.yaml    # 不接 K8s 时留空 ""
  DefaultClusterID: 1
  AutoInitK8s: true

# === 以下均为可选模块，不配置则功能优雅降级 ===

Jenkins:
  URL: ""                     # CI/CD：留空则禁用
  Username: ""
  APIToken: ""

Monitor:
  Prometheus:
    URL: ""                   # 监控：留空则使用 DB 中数据源；DB 也无则禁用

Loki:
  URL: ""                     # 日志：留空则禁用

AIAssistant:
  Enabled: false              # AI：true + APIKey 才启用
  APIKey: "sk-xxx"
  BaseURL: ""                 # 国内代理地址
  Model: "gpt-4o"
```

> 💡 **运行期权威配置在 DB**：监控数据源 / Loki / AI 模型等运行期可变配置应通过「平台管理」页面在数据库中配置，`config.yaml` 仅作首次启动引导。详见 [`docs/配置文件与部署说明.md`](./配置文件与部署说明.md)。

### 4.5 编译并启动后端

```bash
# 1. 下载依赖
go env -w GOPROXY=https://goproxy.cn,direct      # 国内推荐
go mod download

# 2. 安装 swag（生成 Swagger 文档；可选）
go install github.com/swaggo/swag/cmd/swag@latest
export PATH=$PATH:$HOME/go/bin

# 3. 创建日志目录（首次启动）
mkdir -p storage/logs

# 4. 编译
make build
# 等价于：
# swag init -g cmd/k8soperation/main.go -o docs
# go build -trimpath -ldflags="-s -w" -o bin/k8soperation ./cmd/k8soperation

# 5. 启动
./bin/k8soperation

# 或一步到位
make run
```

**Windows**：
```powershell
go mod download
New-Item -ItemType Directory -Path storage\logs -Force
go build -trimpath -ldflags="-s -w" -o bin\k8soperation.exe .\cmd\k8soperation
.\bin\k8soperation.exe
```

启动成功的标志：
```
[GIN-debug] Listening and serving HTTP on :8080
```

### 4.6 启动前端

```bash
cd k8s-web

# 安装依赖
npm install

# 开发模式（热更新，:5173；vite 自动代理 /api → :8080）
npm run dev

# 或生产构建
npm run build
# 产物在 k8s-web/dist/，交给 Nginx 托管
```

浏览器访问：
- 开发模式：<http://localhost:5173>
- 生产模式：<http://localhost>（Nginx）

> 📚 前端独立部署细节见 [`docs/前端管理系统部署文档.md`](./前端管理系统部署文档.md)。

---

## 五、Docker 单机部署

适合小规模/单机生产部署。

### 5.1 构建后端镜像

```bash
# 单架构
make docker-build

# 多架构（amd64 + arm64，需先 docker buildx create --use）
make docker-buildx IMAGE=registry.example.com/k8soperation:v2.6.0
```

### 5.2 docker-compose 一键拉起完整环境

在项目根目录创建 `docker-compose.yaml`：

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: k8s-mysql
    ports: ["3306:3306"]
    environment:
      MYSQL_ROOT_PASSWORD: "admin123"
      MYSQL_DATABASE: "k8s-platform"
    command: >
      --character-set-server=utf8mb4
      --collation-server=utf8mb4_unicode_ci
      --default-authentication-plugin=mysql_native_password
    volumes:
      - mysql-data:/var/lib/mysql
      # 容器首次启动时自动 source 该 SQL
      - ./docs/sql/k8s_platform_full_init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    restart: always

  redis:
    image: redis:7
    container_name: k8s-redis
    ports: ["6379:6379"]
    command: ["redis-server", "--requirepass", "admin123"]
    restart: always

  backend:
    image: k8soperation:latest      # make docker-build 产物
    container_name: k8s-backend
    ports: ["8080:8080"]
    environment:
      TZ: Asia/Shanghai
    volumes:
      - ./configs:/app/configs:ro
      - app-logs:/app/storage/logs
    depends_on:
      - mysql
      - redis
    restart: always

  web:
    image: nginx:alpine
    container_name: k8s-web
    ports: ["80:80"]
    volumes:
      - ./k8s-web/dist:/usr/share/nginx/html:ro
      - ./deploy/nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      - backend
    restart: always

volumes:
  mysql-data:
  app-logs:
```

启动：
```bash
docker compose up -d
docker compose logs -f backend
```

> ⚠️ 容器内 `config.yaml` 路径需要使用**绝对路径**（`/app/storage/logs/*`），与本地开发的相对路径不一致，详见 [`docs/配置文件与部署说明.md`](./配置文件与部署说明.md)。

---

## 六、Kubernetes 生产部署

> 完整 K8s 部署有专门文档：[`docs/K8s部署指南.md`](./K8s部署指南.md)（已含 deploy/ 目录的所有 yaml 详解、Secret 编码、PVC 规划、RBAC 配置、Ingress、健康探针、滚动升级、运维操作等）。

**速览**：

```bash
# 1. 准备 K8s 命名空间
kubectl apply -f deploy/namespace.yaml

# 2. 编辑敏感信息（base64 编码后填入）
vim deploy/secret.yaml

# 3. 调整业务配置（数据库地址、Jenkins 地址等）
vim deploy/configmap.yaml

# 4. 修改镜像 tag
vim deploy/deployment.yaml

# 5. 使用 Kustomize 一键部署（推荐）
kubectl apply -k deploy/

# 6. 或按顺序逐个 apply
kubectl apply -f deploy/secret.yaml
kubectl apply -f deploy/configmap.yaml
kubectl apply -f deploy/pvc.yaml
kubectl apply -f deploy/service.yaml
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/ingress.yaml      # 按需

# 7. 观察滚动升级
kubectl rollout status deployment/k8soperation -n k8soperation
kubectl get pods -n k8soperation -w
```

**关键检查点**：
- `App.GlobalKubeConfigPath` 留空字符串 → 自动使用 InCluster 模式（推荐）
- 日志路径必须使用 `/app/storage/logs/*` 绝对路径
- `ErrorCode.AllowOverride: false`（生产强制）
- PVC 多副本场景必须使用 RWX（NFS / CephFS / 云厂商共享存储）

---

## 七、可选组件接入

### 7.1 Jenkins（CI/CD）

```yaml
# configs/config.yaml
Jenkins:
  URL: "http://jenkins.local:8080/"
  Username: "admin"
  APIToken: "11xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  CallbackURL: "http://k8soperation.local:8080"   # Jenkins → 平台回调
  HMACSecret: "32 位随机字符串，与 Jenkins 共享"
```

模板与流水线接入指南：[`docs/模板化CICD快速接入指南.md`](./模板化CICD快速接入指南.md)、[`docs/Java项目CICD完整接入指南.md`](./Java项目CICD完整接入指南.md)。

### 7.2 Prometheus / Loki（监控与日志）

**推荐通过 DB 配置**：登录平台 → 「平台管理 → 监控数据源」→ 新增 Prometheus / Loki / VictoriaMetrics 数据源（支持按 `cluster_id` 隔离）。

如需 `config.yaml` 引导配置：
```yaml
Monitor:
  Prometheus:
    URL: "http://prometheus.monitoring.svc:9090"
Loki:
  URL: "http://loki.monitoring.svc:3100"
```

详见 [`docs/监控模块扩展设计.md`](./监控模块扩展设计.md)。

### 7.3 SonarQube（代码扫描）

`configs/config.yaml`：
```yaml
SonarQube:
  URL: "http://sonarqube.local:9000/"
  Token: "squ_xxxxxxxxxxxxxxxxxxxx"
  QualityGateBlock: true       # 质量门禁卡板
```

详见 [`docs/SonarQube_代码质量检测设计文档.md`](./SonarQube_代码质量检测设计文档.md)。

### 7.4 AI 大模型

```yaml
AIAssistant:
  Enabled: true
  Provider: openai             # openai / azure / kimi / qwen / deepseek 等
  APIKey: "sk-xxx"
  BaseURL: "https://api.openai.com/v1"   # 国内代理时填代理地址
  Model: "gpt-4o"
  Temperature: 0.3
```

详见 [`docs/AI助手大模型配置指南.md`](./AI助手大模型配置指南.md)。

---

## 八、部署后验证

### 8.1 健康检查

```bash
# 存活探针（仅检测进程）
curl -s http://127.0.0.1:8080/healthz/live
# {"status":"ok"}

# 就绪探针（检测 DB / Redis / K8s 等依赖）
curl -s http://127.0.0.1:8080/healthz/ready
```

### 8.2 登录测试

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

返回 `{"code":0,"data":{"token":"eyJ..."}}` 即说明 DB / Redis / JWT 全链路正常。

### 8.3 浏览器访问

| URL | 期望 |
|-----|------|
| <http://localhost:5173>（开发）/ <http://localhost>（生产） | 跳转登录页 |
| 用 `admin / admin123` 登录 | 进入工作台 |
| 进入「监控总览」 | 双下拉「集群 / 数据源」可切换 |
| 进入「平台管理 → 用户中心」 | **立即修改 admin 密码** |

### 8.4 Swagger 文档

```bash
make swagger-ui
# 浏览器打开 http://localhost:8081
```

或直接访问后端：<http://127.0.0.1:8080/swagger/index.html>

---

## 九、常见问题排错

### Q1：端口被占用

```bash
# Linux / macOS
lsof -i :8080 && kill -9 <PID>
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

### Q2：MySQL 连接失败 `connect: connection refused`

- 检查 MySQL 是否启动：`systemctl status mysql` / `docker ps`
- 检查 `configs/config.yaml` 中 Host/Port/Username/Password
- 数据库名是否一致：默认 `k8s-platform`（带短横线）
- 远程连接需放开：`bind-address = 0.0.0.0` + 防火墙

### Q3：Redis 连接失败 `redis client not initialized`

- Redis 是否启动、端口是否通
- `Cache.Password` 是否与 `redis-server --requirepass` 一致
- ACL 用户是否需要：见 [`docs/📄 K8sOperation 后台系统部署文档.md`](#) Redis ACL 章节（如启用 ACL，则同时填 `Cache.Username`）

### Q4：前端访问 404 / 接口跨域

- 开发模式：检查后端是否在 8080，vite 已自动代理 `/api`
- 生产模式：Nginx 必须配置 `try_files $uri $uri/ /index.html;` + `/api/` 反向代理到后端
- 详见 [`docs/前端管理系统部署文档.md`](./前端管理系统部署文档.md)

### Q5：K8s 集群初始化失败

```
K8s 集群初始化失败，集群管理功能暂不可用，其他功能正常
```

**这是正常降级**，K8s 集群是可选项。需要时：
1. 把集群 KubeConfig 内容放入 `configs/k8s.yaml`
2. 或登录平台 → 「集群管理」页面动态添加

详见 [`docs/登录认证流程详细说明.md`](./登录认证流程详细说明.md)。

### Q6：日志写入失败

- **本地开发**：`LogFileName` 用相对路径 `storage/logs/app.log`
- **Docker / K8s**：`LogFileName` 用绝对路径 `/app/storage/logs/app.log`
- 路径差异详见 [`docs/配置文件与部署说明.md`](./配置文件与部署说明.md)

### Q7：`make build` 报 `swag: command not found`

```bash
go install github.com/swaggo/swag/cmd/swag@latest
export PATH=$PATH:$HOME/go/bin     # Linux/Mac
$env:PATH += ";$HOME\go\bin"       # Windows PowerShell
```

### Q8：Windows PowerShell 命令兼容性

- 不能用 `&&`，改用 `;` 分隔
- 中文乱码：`chcp 65001` 切到 UTF-8
- 详见 [`scripts/check-env.ps1`](../scripts/check-env.ps1)

### Q9：监控数据源切换无效

- 确认 `monitor_datasource` 表已存在 `cluster_id` 字段（执行最新版 `k8s_platform_full_init.sql` 即可）
- 验证：`SHOW COLUMNS FROM monitor_datasource LIKE 'cluster_id';`
- 详见 [`docs/sql/README.md`](./sql/README.md)

### Q10：AI 助手不响应

- `AIAssistant.Enabled: true`
- `APIKey` 真实有效
- 国内服务器需配置 `BaseURL` 走代理
- 详见 [`docs/AI助手大模型配置指南.md`](./AI助手大模型配置指南.md)

---

## 十、运维操作清单

### 10.1 升级到新版本

```bash
git pull
make build              # 重新编译

# 数据库：直接重跑 full_init.sql 即可，幂等不会丢数据
mysql -u root -p k8s-platform < docs/sql/k8s_platform_full_init.sql

# 重启
systemctl restart k8soperation       # systemd
# 或
pkill k8soperation && ./bin/k8soperation &
# K8s
kubectl rollout restart deployment/k8soperation -n k8soperation
```

### 10.2 备份与恢复

```bash
# 备份
mysqldump -u root -padmin123 --single-transaction \
  --default-character-set=utf8mb4 \
  k8s-platform > backup-$(date +%F).sql

# 恢复
mysql -u root -padmin123 k8s-platform < backup-2026-05-18.sql
```

### 10.3 查看日志

```bash
# 本地
tail -f storage/logs/app.log
tail -f storage/logs/biz.log

# K8s
kubectl logs -f deployment/k8soperation -n k8soperation
```

### 10.4 修改 admin 密码

平台 → 「平台管理 → 用户中心」→ 编辑 admin → 修改密码。
或直接 SQL 重置（密码采用 bcrypt）：
```sql
UPDATE sys_admin SET password = '$2a$10$<bcrypt 哈希>' WHERE username='admin';
```

---

## 十一、目录与脚本对照表

### 11.1 核心目录速览

| 目录 / 文件 | 说明 |
|------------|------|
| `cmd/k8soperation/main.go` | 后端入口 |
| `internal/` | 后端业务实现（按模块拆分） |
| `pkg/` | 可复用基础库（k8s / jwt / logger / openai 等） |
| `configs/config.yaml` | 主配置（不入 Git） |
| `configs/k8s.yaml` | KubeConfig（不入 Git） |
| `docs/sql/k8s_platform_full_init.sql` | **唯一**初始化 SQL |
| `docs/` | 全部设计与运维文档 |
| `deploy/` | K8s 部署 yaml |
| `k8s-web/` | Vue3 前端 |
| `scripts/` | 一键脚本（环境检查 / 初始化 / 启动） |
| `storage/logs/` | 运行期日志（不入 Git） |
| `storage/artifacts/` | CI/CD 制品文件（不入 Git） |
| `Makefile` | 编译 / Docker / Swagger 入口 |

### 11.2 环境变量一览

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_HOST` | 127.0.0.1 | MySQL 地址 |
| `DB_PORT` | 3306 | MySQL 端口 |
| `DB_USER` | root | MySQL 用户名 |
| `DB_PASS` | admin123 | MySQL 密码 |
| `DB_NAME` | k8s-platform | 数据库名 |
| `REDIS_HOST` | 127.0.0.1 | Redis 地址 |
| `REDIS_PORT` | 6379 | Redis 端口 |
| `REDIS_PASS` | admin123 | Redis 密码 |
| `BACKEND_PORT` | 8080 | 后端端口 |
| `FRONTEND_PORT` | 5173 | 前端端口 |

### 11.3 Makefile 常用命令

```bash
make build          # 编译后端 + 生成 Swagger
make run            # 编译并启动
make run-local      # go run 开发模式
make test           # 运行单元测试
make docker-build   # 构建单架构镜像
make docker-buildx  # 构建多架构镜像（amd64+arm64）
make docker-run     # 启动容器
make swagger-ui     # 启动 Swagger UI（:8081）
make help           # 查看全部命令
```

### 11.4 文档导航

| 主题 | 文档 |
|------|------|
| **5 分钟跑起来** | [`QUICK_START.md`](../QUICK_START.md) |
| **K8s 集群部署** | [`docs/K8s部署指南.md`](./K8s部署指南.md) |
| **前端独立部署** | [`docs/前端管理系统部署文档.md`](./前端管理系统部署文档.md) |
| **配置项详解** | [`docs/配置文件与部署说明.md`](./配置文件与部署说明.md) |
| **数据库初始化** | [`docs/sql/README.md`](./sql/README.md) |
| **CI/CD 接入** | [`docs/模板化CICD快速接入指南.md`](./模板化CICD快速接入指南.md) |
| **Java 项目接入** | [`docs/Java项目CICD完整接入指南.md`](./Java项目CICD完整接入指南.md) |
| **AI 助手配置** | [`docs/AI助手大模型配置指南.md`](./AI助手大模型配置指南.md) |
| **告警配置** | [`docs/告警配置与触发测试指南.md`](./告警配置与触发测试指南.md) |
| **监控扩展** | [`docs/监控模块扩展设计.md`](./监控模块扩展设计.md) |
| **登录认证流程** | [`docs/登录认证流程详细说明.md`](./登录认证流程详细说明.md) |
| **架构总览** | [`docs/平台整体架构总览.md`](./平台整体架构总览.md) |
| **本地存储说明** | [`docs/本地存储目录与文件说明.md`](./本地存储目录与文件说明.md) |

---

## 🎉 部署完成自检卡

```
┌──────────────────────────────────────────────────┐
│           K8sOperation 部署完成清单               │
├──────────────────────────────────────────────────┤
│  ✅ MySQL & Redis 已启动并可连通                  │
│  ✅ docs/sql/k8s_platform_full_init.sql 执行成功  │
│  ✅ configs/config.yaml 已按真实环境填写          │
│  ✅ 后端 :8080 启动并 /healthz/live 返回 ok       │
│  ✅ 前端 :5173 / :80 可访问                       │
│  ✅ 用 admin/admin123 登录成功                    │
│  ✅ 已立即修改 admin 默认密码                     │
│  ✅ 已修改 JWTSigningKey 为随机串                 │
│  ✅ （可选）K8s 集群已添加并能拉到 namespace      │
│  ✅ （可选）Jenkins / Prometheus / AI 已对接      │
└──────────────────────────────────────────────────┘
```

**至此，K8sOperation 平台已就绪。** 进入「**平台管理**」做基础设置（数据库密码改成强密码、JWT 密钥换随机串、添加业务用户与角色），然后即可开始多集群、CI/CD、监控、AI 智能运维全功能体验。

> 部署期间遇到本文档未覆盖的问题，请优先在 [`docs/`](.) 目录中检索关键词，或提 Issue：
> - Gitee: <https://gitee.com/jay-kim/k8s_operation/issues>
> - GitHub: <https://github.com/jay-codemine/k8s_operation/issues>
