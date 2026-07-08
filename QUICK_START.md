# K8sOperation 快速启动操作手册

> 本手册帮助你从零开始，在 **5 分钟内** 把 K8sOperation 平台跑起来。

---

## 目录

- [一、环境要求](#一环境要求)
- [二、一键启动（推荐）](#二一键启动推荐)
- [三、手动部署（分步操作）](#三手动部署分步操作)
  - [3.1 安装基础依赖](#31-安装基础依赖)
  - [3.2 启动 MySQL & Redis](#32-启动-mysql--redis)
  - [3.3 初始化数据库](#33-初始化数据库)
  - [3.4 配置文件](#34-配置文件)
  - [3.5 编译 & 启动后端](#35-编译--启动后端)
  - [3.6 编译 & 启动前端](#36-编译--启动前端)
- [四、Docker 部署](#四docker-部署)
- [五、验证部署](#五验证部署)
- [六、常见问题](#六常见问题)
- [七、脚本工具一览](#七脚本工具一览)

---

## 一、环境要求

### 必须

| 组件 | 最低版本 | 推荐版本 | 说明 |
|------|---------|---------|------|
| **Go** | 1.21+ | 1.24+ | 后端编译 |
| **MySQL** | 5.7+ | 8.0+ | 主数据库 |
| **Redis** | 5.0+ | 7.0+ | Session / 缓存 / 消息队列 |

### 可选

| 组件 | 版本要求 | 说明 |
|------|---------|------|
| Node.js | 20+ | 前端开发/编译（不开发前端可不装） |
| npm | 10+ | 前端包管理器（随 Node.js 安装） |
| Docker | 20+ | 容器化部署 |
| kubectl | 1.25+ | K8s 集群管理（连接集群时需要） |
| Git | 2.0+ | 代码版本管理 |

### 端口占用

| 端口 | 用途 |
|------|------|
| 8080 | Go 后端 API |
| 5173 | Vue3 前端开发服务器 |
| 3306 | MySQL |
| 6379 | Redis |

---

## 二、一键启动（推荐）

### Linux / macOS

```bash
# 1. 克隆项目
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation

# 2. 赋予执行权限
chmod +x scripts/*.sh

# 3. 检查环境（可选，先看看缺什么）
bash scripts/check-env.sh

# 4. 一键启动（自动完成：环境检查 → 数据库初始化 → 配置生成 → 编译 → 启动）
bash scripts/quick-start.sh
```

### Windows (PowerShell)

```powershell
# 1. 克隆项目
git clone https://gitee.com/jay-kim/k8s_operation.git
cd k8s_operation

# 2. 检查环境（可选）
powershell -ExecutionPolicy Bypass -File scripts\check-env.ps1

# 3. 一键启动
powershell -ExecutionPolicy Bypass -File scripts\quick-start.ps1
```

### 自定义配置

一键脚本支持通过环境变量自定义连接信息：

```bash
# Linux / macOS
DB_HOST=10.0.0.100 DB_PORT=3306 DB_USER=root DB_PASS=MyPassword \
REDIS_HOST=10.0.0.100 REDIS_PASS=RedisPass \
bash scripts/quick-start.sh
```

```powershell
# Windows
$env:DB_HOST="10.0.0.100"
$env:DB_PASS="MyPassword"
$env:REDIS_HOST="10.0.0.100"
.\scripts\quick-start.ps1
```

---

## 三、手动部署（分步操作）

### 3.1 安装基础依赖

#### Go 安装

```bash
# Linux
wget https://go.dev/dl/go1.24.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# macOS
brew install go

# Windows
# 下载安装包: https://go.dev/dl/go1.24.6.windows-amd64.msi
```

设置 Go 代理（国内推荐）：
```bash
go env -w GOPROXY=https://goproxy.cn,https://goproxy.io,direct
```

#### Node.js 安装（前端需要）

```bash
# Linux / macOS (推荐 nvm)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash
nvm install 20
nvm use 20

# Windows
# 下载安装包: https://nodejs.org/
```

### 3.2 启动 MySQL & Redis

#### 方式一：Docker（推荐，最快）

```bash
# 启动 MySQL 8.0
docker run -d --name k8s-mysql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=admin123 \
  -v k8s-mysql-data:/var/lib/mysql \
  mysql:8.0 --default-authentication-plugin=mysql_native_password \
  --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

# 启动 Redis 7
docker run -d --name k8s-redis \
  -p 6379:6379 \
  redis:7 redis-server --requirepass admin123
```

#### 方式二：本地安装

```bash
# Ubuntu / Debian
sudo apt install mysql-server redis-server
sudo systemctl start mysql redis-server

# CentOS / RHEL
sudo yum install mysql-server redis
sudo systemctl start mysqld redis

# macOS
brew install mysql redis
brew services start mysql
brew services start redis

# Windows
# MySQL: https://dev.mysql.com/downloads/installer/
# Redis: https://github.com/tporadowski/redis/releases
```

### 3.3 初始化数据库

#### 方式一：使用脚本（推荐）

```bash
# Linux / macOS
bash scripts/init-db.sh

# Windows
powershell -ExecutionPolicy Bypass -File scripts\init-db.ps1
```

#### 方式二：手动执行 SQL

```bash
# 一行命令完成初始化
mysql -h 127.0.0.1 -P 3306 -u root -padmin123 \
  --default-character-set=utf8mb4 < docs/sql/k8s_platform_full_init.sql
```

#### 方式三：进入 MySQL 命令行

```sql
-- 登录 MySQL
mysql -u root -p

-- 执行初始化脚本
source docs/sql/k8s_platform_full_init.sql

-- 验证
USE `k8s-platform`;
SHOW TABLES;
SELECT username, role FROM user;
```

> **初始化内容**：34 张表 + 2 个视图 + RBAC 权限数据 + CICD 资源模板
>
> **默认管理员**：`admin` / `admin123`

### 3.4 配置文件

```bash
# 从示例文件复制
cp configs/config.yaml.example configs/config.yaml
cp configs/k8s.yaml.example configs/k8s.yaml    # K8s 集群配置（可选）
```

编辑 `configs/config.yaml`，核心配置说明：

```yaml
Server:
  Port: 8080                # 后端端口

Database:
  Host: 127.0.0.1           # MySQL 地址
  Port: 3306
  Username: root
  Password: admin123           # ⚠️ 修改为你的密码
  DBName: k8s-platform

Cache:
  Address: 127.0.0.1:6379   # Redis 地址
  Password: "admin123"         # ⚠️ 修改为你的 Redis 密码

App:
  LogLevel: debug            # 日志级别: debug / info / warn / error
  JWTSigningKey: xxxx        # ⚠️ 生产环境请修改为随机字符串

# K8s 集群配置（可选，不配置则集群管理功能不可用，其他功能正常）
# GlobalKubeConfigPath: configs/k8s.yaml

# Jenkins CI/CD 配置（可选）
Jenkins:
  URL: ""                    # 留空则 CI/CD 功能不可用

# AI 助手配置（可选）
AIAssistant:
  Enabled: false             # 设为 true 并填入 APIKey 启用 AI
  APIKey: "sk-xxxx"
  BaseURL: ""                # 国内代理地址
  Model: "gpt-4o"
```

### 3.5 编译 & 启动后端

```bash
# 下载依赖
go mod download

# 创建日志目录
mkdir -p storage/logs

# 编译
go build -trimpath -ldflags="-s -w" -o bin/k8soperation ./cmd/k8soperation

# 启动
./bin/k8soperation

# 或使用 Makefile
make build    # 编译（含 Swagger 文档生成）
make run      # 编译 + 启动
```

Windows:
```powershell
go mod download
New-Item -ItemType Directory -Path storage\logs -Force
go build -trimpath -ldflags="-s -w" -o bin\k8soperation.exe ./cmd/k8soperation
.\bin\k8soperation.exe
```

看到以下日志说明启动成功：
```
[GIN-debug] Listening and serving HTTP on :8080
```

### 3.6 编译 & 启动前端

```bash
cd k8s-web

# 安装依赖
npm install

# 开发模式启动（热更新）
npm run dev
# 前端运行在 http://localhost:5173
# 自动代理 /api 请求到 http://localhost:8080

# 生产构建
npm run build
# 产物在 k8s-web/dist/
```

---

## 四、Docker 部署

### 单架构构建

```bash
# 标准构建
make docker-build

# BuildKit 加速构建
make bk-build

# 启动容器
make docker-run

# 查看日志
make docker-logs
```

### 多架构构建（amd64 + arm64）

```bash
# 需要 docker buildx
make docker-buildx
```

### Docker Compose（完整环境）

```yaml
# docker-compose.yaml
version: '3.8'
services:
  mysql:
    image: mysql:8.0
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: "admin123"
    volumes:
      - mysql-data:/var/lib/mysql
      - ./docs/sql/k8s_platform_full_init.sql:/docker-entrypoint-initdb.d/init.sql
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

  redis:
    image: redis:7
    ports:
      - "6379:6379"
    command: redis-server --requirepass admin123

  backend:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./configs:/app/configs:ro
    depends_on:
      - mysql
      - redis
    restart: always

volumes:
  mysql-data:
```

```bash
docker-compose up -d
```

---

## 五、验证部署

### 1. 健康检查

```bash
# 后端健康检查
curl http://localhost:8080/healthz/live

# 预期返回
# {"status":"ok"}
```

### 2. 登录测试

打开浏览器访问 http://localhost:5173（开发模式）

- 用户名: `admin`
- 密码: `admin123`

### 3. API 测试

```bash
# 登录获取 Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 4. Swagger 文档

```bash
# 生成并访问 API 文档
make swagger-ui
# 打开 http://localhost:8081
```

---

## 六、常见问题

### Q1: 端口被占用

```bash
# Linux / macOS - 查找占用端口的进程
lsof -i :8080
kill -9 <PID>

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

### Q2: MySQL 连接失败

```
Error: connect db failed: dial tcp 127.0.0.1:3306: connect: connection refused
```

- 确认 MySQL 已启动
- 确认 `configs/config.yaml` 中的数据库配置正确
- 确认数据库 `k8s-platform` 已创建

### Q3: Redis 连接失败

```
Error: redis client not initialized
```

- 确认 Redis 已启动
- 确认 `configs/config.yaml` 中的 Cache 配置正确

### Q4: 前端代理 404

确认后端已启动在 8080 端口，前端的 vite 代理配置会自动转发 `/api` 请求。

### Q5: K8s 集群连接失败

```
K8s 集群初始化失败，集群管理功能暂不可用，其他功能正常
```

这是正常现象！如果没有 K8s 集群，其他功能（用户管理、RBAC、CI/CD 等）仍可正常使用。

需要连接 K8s 集群时：
1. 将集群的 KubeConfig 内容放入 `configs/k8s.yaml`
2. 或通过管理后台「集群管理」页面添加集群

### Q6: AI 助手不可用

确认 `configs/config.yaml` 中：
```yaml
AIAssistant:
  Enabled: true
  APIKey: "sk-你的真实API Key"
  BaseURL: "https://你的代理地址"  # 国内用户需配置代理
```

### Q7: Windows 编译出错 `swag: command not found`

```powershell
go install github.com/swaggo/swag/cmd/swag@latest
```

---

## 七、脚本工具一览

所有脚本位于 `scripts/` 目录：

| 脚本 | 平台 | 用途 |
|------|------|------|
| `quick-start.sh` | Linux/Mac | 一键启动（环境检查→数据库→配置→编译→启动） |
| `quick-start.ps1` | Windows | 一键启动（同上） |
| `check-env.sh` | Linux/Mac | 仅检查环境依赖，不做修改 |
| `check-env.ps1` | Windows | 仅检查环境依赖，不做修改 |
| `init-db.sh` | Linux/Mac | 独立数据库初始化（支持安全追加/删除重建） |
| `init-db.ps1` | Windows | 独立数据库初始化（同上） |

### 环境变量参考

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

### Makefile 命令

```bash
make build          # 编译后端
make run            # 编译 + 启动
make run-local      # go run 开发模式
make test           # 运行测试
make docker-build   # Docker 构建
make docker-buildx  # 多架构构建 (amd64+arm64)
make docker-run     # 启动容器
make swagger-ui     # 启动 Swagger UI
make help           # 查看所有命令
```

---

## 快速参考卡片

```
┌─────────────────────────────────────────────────┐
│            K8sOperation 快速部署                 │
├─────────────────────────────────────────────────┤
│                                                  │
│  1. git clone → cd k8s_operation                 │
│  2. 启动 MySQL + Redis                          │
│  3. bash scripts/quick-start.sh  (一键搞定)      │
│                                                  │
│  后端: http://localhost:8080                      │
│  前端: http://localhost:5173                      │
│  账号: admin / admin123                          │
│                                                  │
│  Swagger: make swagger-ui → localhost:8081        │
│  Docker:  make docker-build && make docker-run   │
│                                                  │
└─────────────────────────────────────────────────┘
```
