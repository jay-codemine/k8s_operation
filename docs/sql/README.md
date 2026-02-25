# 数据库初始化说明（K8s Platform）

本文档用于说明 **K8s Platform 后端系统** 的数据库初始化流程。

当前数据库主要用于：
- 用户管理
- Kubernetes 集群管理（集群接入、状态维护等）

---

## 一、数据库信息

- 数据库类型：MySQL 8.0+
- 字符集：utf8mb4
- 排序规则：utf8mb4_0900_ai_ci
- 数据库名称：`k8s-platform`

> ⚠️ 请确保 MySQL 版本 >= 8.0，否则部分字符集可能不兼容。

---

## 二、初始化顺序（非常重要）

**首次部署时，必须按以下顺序执行 SQL：**

1. 创建并初始化 **Kubernetes 集群管理表**
2. 创建并初始化 **用户表**

执行顺序如下：

```sql
-- 1. 集群管理表
source kube_cluster.sql;

-- 2. 用户表
source user.sql;
```

---

## 三、执行方式

### 方式一：命令行执行（推荐）

```bash
# 进入 sql 目录
cd sql

# 登录 MySQL
mysql -h 127.0.0.1 -u root -p

# 选择数据库
USE `k8s-platform`;

# 执行建表 SQL
source kube_platform.sql;
```

---

### 方式二：一次性导入

```bash
mysql -h 127.0.0.1 -u root -p k8s-platform < kube_platform.sql
```

---

## 四、表说明

### 1. kube_cluster（K8s 集群管理表）

用于存储平台接入的 Kubernetes 集群信息，包括：

- 集群名称
- kubeconfig 配置（base64 编码）
- 集群版本
- 当前状态
- 最近一次健康检查时间
- 最近一次异常原因

该表是 **平台核心表之一**。

---

### 2. user（用户表）

用于存储平台用户信息，包括：

- 用户名
- 密码（加密存储）
- 创建 / 修改 / 删除时间
- 逻辑删除标识

---

## 五、注意事项

- 本项目采用 **逻辑删除**（`is_del` 字段），业务查询需注意过滤已删除数据
- `kube_config` 字段内容为 **base64 编码后的 kubeconfig**
- 初始化 SQL **可重复执行**（`IF NOT EXISTS`，不会影响已有数据）
- 数据库结构变更建议通过 **版本迭代统一管理**，不建议直接在线修改表结构

---

## 六、常见问题

### Q1：是否需要先创建数据库？

```sql
CREATE DATABASE IF NOT EXISTS `k8s-platform`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;
```

---

### Q2：用户初始数据是否在 SQL 中创建？

- 当前版本 **不包含默认用户**
- 用户由平台接口创建，或通过后续初始化脚本完成
