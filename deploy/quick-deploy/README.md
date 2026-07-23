# K8sOperation 一键部署（Docker Compose）

本目录提供**一键部署脚本**，基于项目根目录的 `docker-compose.yaml`，一条命令拉起完整平台：

```
MySQL 8.0  +  Redis 7  +  后端 API (Go)  +  前端 Web (Vue3 + Nginx)
```

---

## 一、环境要求

| 组件 | 版本 | 说明 |
|------|------|------|
| Docker | 20.10+ | 需已安装并**正在运行** |
| Docker Compose | v2（`docker compose`） | Docker Desktop 已内置；Linux 需装 compose 插件 |

> 内存建议 ≥ 4GB。首次部署需要构建镜像 + 拉取基础镜像，请保证网络可用。

---

## 二、一键部署

### Windows（PowerShell）

在**项目根目录**下执行：

```powershell
.\deploy\quick-deploy\deploy.ps1
```

> 若提示脚本被禁止运行，先执行一次：
> ```powershell
> Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
> ```

### Linux / macOS

```bash
chmod +x deploy/quick-deploy/deploy.sh      # 首次赋予执行权限
./deploy/quick-deploy/deploy.sh
```

脚本会自动完成：**环境检查 → 构建镜像 → 启动服务 → 等待健康检查 → 打印访问信息**。

---

## 三、访问信息

部署完成后：

| 入口 | 地址 |
|------|------|
| 前端控制台 | http://localhost |
| 后端 API | http://localhost:8080 |
| Swagger 文档 | http://localhost:8080/swagger/index.html |

**默认管理员账号**（后端首次启动自动创建）：

```
用户名：admin
密码：  123456
```

> ⚠️ 生产环境请登录后**立即修改默认密码**。

中间件连接信息（本机调试用）：

| 组件 | 地址 | 账号 |
|------|------|------|
| MySQL | localhost:3306 | root / admin123，库名 `k8s-platform` |
| Redis | localhost:6379 | 密码 admin123 |

---

## 四、常用命令

| 命令（Windows） | 命令（Linux/Mac） | 作用 |
|------|------|------|
| `deploy.ps1` 或 `deploy.ps1 up` | `deploy.sh` 或 `deploy.sh up` | 一键部署 / 更新（构建 + 启动） |
| `deploy.ps1 status` | `deploy.sh status` | 查看服务状态 |
| `deploy.ps1 logs` | `deploy.sh logs` | 跟踪查看全部日志 |
| `deploy.ps1 restart` | `deploy.sh restart` | 重启全部服务 |
| `deploy.ps1 down` | `deploy.sh down` | 停止并删除容器（**保留**数据） |
| `deploy.ps1 clean` | `deploy.sh clean` | 停止并删除容器 **+ 数据卷**（彻底清空，慎用） |

---

## 五、常见问题

**Q1：`admin/123456` 登录不了？**
后端首次启动才会自动建号。用 `status` 确认 `backend` 服务是 `healthy`，或用 `logs` 看后端启动日志里是否有 `Admin ... assigned super_admin role`。

**Q2：改了初始化 SQL，但没生效？**
`docs/sql/k8s_platform_full_init.sql` 只在 **MySQL 数据卷为空（首次）** 时执行。已经初始化过要重新导入，需执行 `clean` 清空数据卷后再 `up`。

**Q3：端口被占用（80 / 8080 / 3306 / 6379）？**
修改根目录 `docker-compose.yaml` 里对应服务的 `ports` 映射（如 `"8081:8080"`）后重新 `up`。

**Q4：前端能打开但接口 502？**
多半是 `backend` 未就绪。前端已配置 `depends_on: backend healthy`，等后端健康后刷新即可；仍不行用 `logs` 看后端报错（通常是连不上 MySQL）。

**Q5：K8s 集群管理功能用不了？**
Compose 环境默认不挂载 kubeconfig（`AutoInitK8s: false`）。这是纯平台本体的运行环境，连接目标 K8s 集群需在平台内「集群管理」中添加 kubeconfig。

---

## 六、目录说明

```
deploy/quick-deploy/
├── deploy.ps1     # Windows 一键部署脚本
├── deploy.sh      # Linux / macOS 一键部署脚本
└── README.md      # 本说明

# 相关文件（项目根目录）
docker-compose.yaml               # 编排定义（脚本调用它）
configs/config-docker.yaml        # 后端在 Compose 环境下的配置
docker/backend/Dockerfile         # 后端镜像构建
docker/frontend/Dockerfile        # 前端镜像构建
docs/sql/k8s_platform_full_init.sql   # 数据库初始化脚本
```
