# 📄 **K8sOperation 后台系统部署文档**

## 📘 目录

- 🚀 环境要求
- 0️⃣ 初始化配置文件
- 1️⃣ 安装 Go 12311
- 2️⃣ 安装 make
- 3️⃣ 配置 Go 代理
- 4️⃣ 安装 swag
- 5️⃣ 编译项目
- 6️⃣ 安装 Redis
- 7️⃣ 配置 Redis 服务端
- 8️⃣ 修改项目 Redis 配置（configs/configyaml）
- 9️⃣ 配置 MySQL
- 🔟 创建数据库表
- 1️⃣1️⃣ 配置 K8s 连接（configs/k8syaml）
- 1️⃣2️⃣ 启动项目
- 1️⃣3️⃣ 访问 Swagger

------

# 🚀 环境要求

| 组件     | 版本要求     |
| -------- | ------------ |
| 操作系统 | CentOS 7.9   |
| Go       | **1.23.11+** |
| make     | **4.4.1+**   |
| swag     | **1.10.1+**  |
| Redis    | **6.4.0+**   |
| MySQL    | **8.0.33+**  |

------



# 0️⃣ 初始化配置文件

项目提供了示例配置文件，部署前需要拷贝并改名，然后在文件中填入你自己的真实环境信息。

```bash
cd /path/to/k8s_operation

# 1. 应用配置：从 example 复制为实际使用的配置文件
cp configs/config.yaml.example configs/config.yaml

# 2. K8s 集群配置：从 example 复制为实际使用的配置文件
cp configs/k8s.yaml.example configs/k8s.yaml
```

- `configs/config.yaml` —— 项目主配置（Redis / MySQL / 其他参数）
- `configs/k8s.yaml` —— Kubernetes 集群 kubeconfig（实际文件，不提交）

> ⚠️ 注意：`config.yaml` 和 `k8s.yaml` 都包含敏感信息，必须加入 `.gitignore`，不要提交到 Git。

用于实际部署的配置文件包括：

| 文件路径              | 用途                      |
| --------------------- | ------------------------- |
| `configs/config.yaml` | 项目主配置（Redis/MySQL） |
| `configs/k8s.yaml`    | Kubernetes 集群连接配置   |

后续的 Redis、MySQL、kubeconfig 等都修改这两个文件即可。

------

# 1️⃣ 安装 Go 1.23.11

```bash
wget https://golang.google.cn/dl/go1.23.11.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.23.11.linux-amd64.tar.gz
/usr/local/go/bin/go version
```

### 加入 PATH

```bash
export PATH=/usr/local/go/bin:$PATH
echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
```

------

# 2️⃣ 安装 make

```bash
yum install -y make
```

------

# 3️⃣ 配置 Go 代理

```bash
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

------

# 4️⃣ 安装 swag

```bash
go install github.com/swaggo/swag/cmd/swag@latest
export PATH=$PATH:$HOME/go/bin
echo 'export PATH=$PATH:/root/go/bin' >> ~/.bashrc
source ~/.bashrc
swag --version
```

------

# 5️⃣ 编译项目

```bash
make build
```

------

# 6️⃣ 安装 Redis

```bash
yum install -y redis
systemctl enable redis --now
```

------

# 7️⃣ 配置 Redis 服务端

编辑配置：

```bash
vim /etc/redis.conf
```

修改：

```bash
bind 0.0.0.0
aclfile /etc/redis/user.acl
```

初始化 ACL：

```bash
mkdir /etc/redis
touch /etc/redis/user.acl
chown -R redis:redis /etc/redis
systemctl restart redis
```

创建 Redis 用户：

```bash
redis-cli
ACL SETUSER root on >123456 ~* &* +@all
ACL SAVE
```

------

# 8️⃣ 修改项目 Redis 配置（configs/config.yaml）

```yaml
# 文件：configs/config.yaml
Cache:
  Type: redis
  Name: sk_sid
  Address: localhost:6379     # 修改为你的 Redis 地址
  Username: "root"            # Redis 用户名
  Password: "123456"          # Redis 密码
  MaxConnect: 10
  Network: tcp
  Secret: "k8smana"           # 自定义密钥
```

------

# 9️⃣ 配置 MySQL

```bash
mysql -u root -p
```

创建数据库和用户：

```bash
CREATE DATABASE kubemana CHARACTER SET utf8 COLLATE utf8_general_ci;
CREATE USER 'kubemana'@'%' IDENTIFIED BY 'kubemana@123';
GRANT ALL PRIVILEGES ON kubemana.* TO 'kubemana'@'%';
```

选择数据库：

```bash
USE kubemana;
```

------

# 🔟 创建数据库表

## k8s_cluster 表

```sql
CREATE TABLE `k8s_cluster` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `cluster_name` varchar(191) NOT NULL DEFAULT '' COMMENT '集群名称',
  `kube_config` mediumtext NOT NULL COMMENT 'kubeconfig内容',
  `cluster_version` varchar(191) NOT NULL DEFAULT '' COMMENT '集群版本',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '集群状态,0=正常,1=不正常',
  `created_at` int unsigned NOT NULL DEFAULT '0',
  `modified_at` int unsigned NOT NULL DEFAULT '0',
  `deleted_at` int unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_cluster_name` (`cluster_name`),
  KEY `idx_status_isdel` (`status`,`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='K8s集群表';
```

## user 表

```sql
CREATE TABLE `user` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码',
  `created_at` int unsigned DEFAULT '0',
  `modified_at` int unsigned DEFAULT '0',
  `deleted_at` int unsigned DEFAULT '0',
  `is_del` tinyint unsigned DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COMMENT='用户表';
```

------

# 1️⃣1️⃣ 配置 K8s 连接（configs/k8s.yaml）

`configs/k8s.yaml` 用于配置要连接的 Kubernetes 集群。
 你可以直接拷贝 kubeconfig 内容到该文件。

## 方法一：直接拷贝 kubeconfig

```bash
scp /root/.kube/config <目标主机IP>:/项目路径/configs/k8s.yaml
```

## 方法二：手工复制内容

```bash
cat ~/.kube/config        # 查看 kubeconfig
vim configs/k8s.yaml      # 粘贴内容
```

> 📌 仓库中的 `k8s-test.yaml`、`test.yaml` 仅为模板文件，请不要直接使用。

------

# 1️⃣2️⃣ 启动项目

```bash
./bin/k8soperation
```

------

# 1️⃣3️⃣ 访问 Swagger

```bash
http://<服务器IP>:8080/swagger/index.html
```

------

# 🎉 完整部署文档整理完毕！

你现在可以直接把这份内容保存为：

```bash
docs/K8sOperation 后台系统部署文档.md
```

并提交到 `k8s_operation` 仓库。