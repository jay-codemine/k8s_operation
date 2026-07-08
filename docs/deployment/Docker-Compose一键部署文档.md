# K8sOperation Docker Compose 一键部署文档

## 一、概述

本文档提供 K8sOperation 统一运维平台的 Docker Compose 一键部署方案，包含以下服务：

| 服务 | 镜像 | 端口 | 说明 |
|------|------|------|------|
| MySQL 8.0 | `mysql:8.0` | 3306 | 数据库（自动建库建表） |
| Redis 7 | `redis:7-alpine` | 6379 | 缓存/Session |
| Backend (Go) | `devops-be:latest` | 8080 | 后端 API |
| Frontend (Vue3) | `devops-fe:latest` | 80 | 前端 Web (Nginx) |

---

## 二、前置条件

### 2.1 系统要求

| 项目 | 最低要求 | 推荐 |
|------|---------|------|
| 操作系统 | Linux / macOS / Windows | Ubuntu 22.04+ / CentOS 8+ |
| Docker | 20.10+ | 24.0+ |
| Docker Compose | v2.0+ | v2.20+ |
| 内存 | 4GB | 8GB+ |
| 磁盘 | 10GB 可用空间 | 20GB+ |

### 2.2 检查 Docker 环境

```bash
# 检查 Docker 版本
docker --version

# 检查 Docker Compose 版本
docker compose version

# 确保 Docker 正在运行
docker info
```

---

## 三、快速开始（一键部署）

### 3.1 克隆项目

```bash
# 从 Gitee 克隆（国内推荐）
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation

# 或从 GitLab 克隆
git clone https://gitlab.maitian-yun.com/DevOps/k8s_operation.git
cd k8s_operation
```

### 3.2 一键启动

```bash
# 构建并启动所有服务（首次会自动构建镜像，耗时约 3-5 分钟）
docker compose up -d --build
```

### 3.3 查看服务状态

```bash
docker compose ps
```

预期输出：
```
NAME              IMAGE             STATUS                    PORTS
k8sop-mysql       mysql:8.0         Up (healthy)              0.0.0.0:3306->3306/tcp
k8sop-redis       redis:7-alpine    Up (healthy)              0.0.0.0:6379->6379/tcp
k8sop-backend     devops-be:latest  Up (healthy)              0.0.0.0:8080->8080/tcp
k8sop-frontend    devops-fe:latest  Up (healthy)              0.0.0.0:80->80/tcp
```

### 3.4 访问平台

| 入口 | 地址 | 说明 |
|------|------|------|
| 前端界面 | http://localhost | 管理控制台 |
| 后端 API | http://localhost:8080 | REST API |
| Swagger 文档 | http://localhost:8080/swagger/index.html | API 文档 |

**默认登录账号：**
- 用户名：`admin`
- 密码：`admin123`

---

## 四、配置说明

### 4.1 目录结构

```
k8s_operation/
├── docker-compose.yaml              # 编排文件（核心）
├── configs/
│   ├── config-docker.yaml           # 后端配置（Docker 专用）
│   └── k8s.yaml                     # K8s 集群配置（可选）
├── docker/
│   ├── backend/
│   │   ├── Dockerfile               # 后端多阶段构建
│   │   └── Dockerfile.runtime       # 后端纯运行时构建
│   └── frontend/
│       └── Dockerfile               # 前端多阶段构建
└── docs/sql/
    └── k8s_platform_full_init.sql   # 数据库初始化脚本
```

### 4.2 环境变量一览

#### MySQL
| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MYSQL_ROOT_PASSWORD` | `admin123` | root 密码 |
| `MYSQL_DATABASE` | `k8s-platform` | 默认数据库名 |

#### Redis
| 配置 | 默认值 | 说明 |
|------|--------|------|
| 密码 | `admin123` | `--requirepass` |
| 最大内存 | `256mb` | `--maxmemory` |

#### Backend
| 变量 | 默认值 | 说明 |
|------|--------|------|
| `GIN_MODE` | `release` | Gin 运行模式 |
| `APP_CONFIG` | `/app/configs/config.yaml` | 配置文件路径 |

#### Frontend
| 变量 | 默认值 | 说明 |
|------|--------|------|
| `API_BACKEND_URL` | `http://backend:8080` | 后端 API 地址 |

### 4.3 自定义配置

如需修改后端配置，编辑 `configs/config-docker.yaml`：

```yaml
# 修改数据库密码
Database:
  Password: your_new_password

# 启用 AI 助手
AIAssistant:
  Enabled: true
  APIKey: "sk-your-openai-key"
  BaseURL: "https://api.openai.com/v1"  # 国内可用代理地址

# 启用 Jenkins CI/CD
Jenkins:
  URL: "http://your-jenkins:8080/"
  Username: "admin"
  APIToken: "your-api-token"

# 启用监控
Monitoring:
  Enabled: true
  PrometheusURL: "http://prometheus:9090"
```

修改后重启后端：
```bash
docker compose restart backend
```

### 4.4 接入 K8s 集群

1. 准备 KubeConfig 文件，复制到 `configs/k8s.yaml`
2. 修改 `configs/config-docker.yaml`：
```yaml
App:
  AutoInitK8s: true
  GlobalKubeConfigPath: configs/k8s.yaml
```
3. 重启后端：
```bash
docker compose restart backend
```

> **注意**：如果 K8s 集群 API Server 地址是内网 IP，需确保 Docker 容器可以访问该网络。

---

## 五、构建选项

### 5.1 仅构建不启动

```bash
docker compose build
```

### 5.2 指定平台构建（Apple Silicon → Linux amd64）

```bash
# 单独构建后端（交叉编译）
docker build -f docker/backend/Dockerfile --platform linux/amd64 -t devops-be:latest .

# 单独构建前端
docker build -f docker/frontend/Dockerfile --platform linux/amd64 -t devops-fe:latest ./k8s-web
```

### 5.3 使用预构建镜像（跳过本地构建）

如果已有 CI/CD 产出的镜像，修改 `docker-compose.yaml`：

```yaml
backend:
  image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v1.1.5
  # build:  # 注释掉 build 段
  #   context: .
  #   dockerfile: docker/backend/Dockerfile

frontend:
  image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:v1.1.5
  # build:
  #   context: ./k8s-web
  #   dockerfile: ../docker/frontend/Dockerfile
```

然后直接启动：
```bash
docker compose up -d
```

---

## 六、运维操作

### 6.1 日常命令

```bash
# 查看所有服务状态
docker compose ps

# 查看日志（实时跟踪）
docker compose logs -f

# 查看单个服务日志
docker compose logs -f backend
docker compose logs -f frontend

# 重启单个服务
docker compose restart backend

# 停止所有服务（保留数据）
docker compose stop

# 启动已停止的服务
docker compose start

# 停止并删除容器（保留数据卷）
docker compose down

# 停止并删除容器+数据卷（⚠️ 清除所有数据）
docker compose down -v
```

### 6.2 数据备份

```bash
# 备份 MySQL 数据
docker exec k8sop-mysql mysqldump -u root -padmin123 k8s-platform > backup_$(date +%Y%m%d).sql

# 恢复 MySQL 数据
docker exec -i k8sop-mysql mysql -u root -padmin123 k8s-platform < backup_20260625.sql

# 备份 Redis 数据
docker exec k8sop-redis redis-cli -a admin123 BGSAVE
docker cp k8sop-redis:/data/dump.rdb ./redis_backup.rdb
```

### 6.3 版本升级

```bash
# 拉取最新代码
git pull origin main

# 重新构建并启动（保留数据）
docker compose up -d --build

# 如果有数据库变更，需要手动执行迁移 SQL
docker exec -i k8sop-mysql mysql -u root -padmin123 k8s-platform < docs/sql/k8s_platform_full_init.sql
```

### 6.4 健康检查

```bash
# 检查后端健康
curl http://localhost:8080/healthz/live

# 检查前端健康
curl http://localhost/health

# 检查 MySQL
docker exec k8sop-mysql mysqladmin -u root -padmin123 ping

# 检查 Redis
docker exec k8sop-redis redis-cli -a admin123 ping
```

---

## 七、网络与端口

### 7.1 默认端口映射

| 宿主机端口 | 容器端口 | 服务 |
|-----------|---------|------|
| 80 | 80 | Frontend (Nginx) |
| 8080 | 8080 | Backend (Go API) |
| 3306 | 3306 | MySQL |
| 6379 | 6379 | Redis |

### 7.2 修改端口

如端口冲突，编辑 `docker-compose.yaml` 中的 `ports` 映射：

```yaml
services:
  frontend:
    ports:
      - "8088:80"      # 宿主机 8088 → 容器 80
  backend:
    ports:
      - "38180:8080"   # 宿主机 38180 → 容器 8080
  mysql:
    ports:
      - "13306:3306"   # 宿主机 13306 → 容器 3306
```

### 7.3 内部网络

所有服务通过 `k8sop-net` 桥接网络互通，容器间通过**服务名**访问：
- 后端连接 MySQL：`mysql:3306`
- 后端连接 Redis：`redis:6379`
- 前端代理后端：`backend:8080`

---

## 八、常见问题 (FAQ)

### Q1: 首次启动后端报数据库连接失败？

**原因**：MySQL 初始化需要 20-30 秒，后端可能在 MySQL 就绪前尝试连接。

**解决**：docker-compose 已配置 `depends_on` + `healthcheck`，正常情况会自动等待。如仍有问题：
```bash
# 手动重启后端
docker compose restart backend
```

### Q2: 前端页面打开白屏？

**排查步骤**：
```bash
# 1. 检查后端是否正常
curl http://localhost:8080/healthz/live

# 2. 检查前端 Nginx 日志
docker compose logs frontend

# 3. 确认前端能访问后端（进入容器测试）
docker exec k8sop-frontend wget -qO- http://backend:8080/healthz/live
```

### Q3: Apple Silicon (M1/M2/M3) 构建报错？

```bash
# 指定 amd64 平台构建
docker compose build --no-cache
# 或使用 buildx
DOCKER_DEFAULT_PLATFORM=linux/amd64 docker compose up -d --build
```

### Q4: 磁盘空间不足？

```bash
# 清理无用的 Docker 资源
docker system prune -f

# 清理构建缓存
docker builder prune -f
```

### Q5: 如何修改 MySQL/Redis 密码？

1. 修改 `docker-compose.yaml` 中的密码
2. 修改 `configs/config-docker.yaml` 中对应的连接密码
3. 删除旧数据卷并重建：
```bash
docker compose down -v
docker compose up -d --build
```

### Q6: 如何只启动 MySQL + Redis（本地开发用）？

```bash
# 仅启动基础设施
docker compose up -d mysql redis

# 确认服务就绪
docker compose ps
```

然后本地运行后端：
```bash
make run-local
```

---

## 九、生产部署建议

| 项目 | 开发/测试 | 生产环境 |
|------|----------|---------|
| MySQL | Docker 内置 | 外部 RDS / 独立主从 |
| Redis | Docker 内置 | 外部 Redis Cluster |
| 日志 | 容器 stdout | ELK / Loki 采集 |
| HTTPS | 不需要 | Nginx/Ingress + 证书 |
| 数据持久化 | Docker Volume | 云盘 / NFS |
| 镜像仓库 | 本地构建 | ACR / Harbor |
| 备份 | 手动 | 定时自动备份 |

> **生产环境建议使用 K8s 部署**，详见 `deploy/` 目录下的 Kustomize 配置。

---

## 十、完整部署流程（端到端）

```bash
# === Step 1: 克隆代码 ===
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation

# === Step 2: 一键启动 ===
docker compose up -d --build

# === Step 3: 等待服务就绪（约 60 秒） ===
echo "等待服务启动..."
sleep 60

# === Step 4: 验证 ===
echo "--- 服务状态 ---"
docker compose ps

echo "--- 后端健康 ---"
curl -s http://localhost:8080/healthz/live

echo "--- 前端健康 ---"
curl -s http://localhost/health

# === Step 5: 访问 ===
echo ""
echo "✅ 部署完成！"
echo "   前端: http://localhost"
echo "   账号: admin / admin123"
```

---

## 附录：docker-compose.yaml 架构图

```
┌─────────────────────────────────────────────────────┐
│                   k8sop-net (bridge)                 │
│                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────────┐  │
│  │  MySQL   │    │  Redis   │    │   Backend    │  │
│  │  :3306   │◄───┤  :6379   │◄───┤   :8080      │  │
│  └──────────┘    └──────────┘    └──────┬───────┘  │
│                                         │          │
│                                         ▼          │
│                                  ┌──────────────┐  │
│                                  │   Frontend   │  │
│                                  │   (Nginx)    │  │
│                                  │    :80       │  │
│                                  └──────────────┘  │
│                                                     │
└─────────────────────────────────────────────────────┘
         ↕ 3306    ↕ 6379    ↕ 8080      ↕ 80
      宿主机端口映射（可自定义）
```
