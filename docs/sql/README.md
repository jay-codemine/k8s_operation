# K8sOperation 数据库初始化

> 🎯 **真正的一键初始化**：单个脚本搞定全部 50 张表 DDL + 字典数据 + 演示业务数据。
> 用户 / RBAC / 集群 / CI-CD / 制品库 / 资源模板 / 镜像 / IAM / 应用商城 / AI 助手 / 监控告警，全部内置。

## 🚀 快速开始

**数据库账号**：`root` / `123456`（本地默认）

> ⚠️ **PowerShell 用户注意**：PowerShell 不支持 `<` 重定向运算符，必须使用 `-e "source ..."` 方式或 `cmd /c` 包装。

### ✅ 方式一：Windows PowerShell（推荐）
```powershell
mysql -u root -p123456 --default-character-set=utf8mb4 -e "source D:/k8s-go/k8s_operation/docs/sql/k8s_platform_full_init.sql"
```

### ✅ 方式二：PowerShell 用 cmd 包装（支持 `<` 重定向）
```powershell
cmd /c "mysql -u root -p123456 --default-character-set=utf8mb4 < docs\sql\k8s_platform_full_init.sql"
```

### ✅ 方式三：MySQL 客户端内执行
```bash
mysql -u root -p123456
```
然后在 mysql 提示符下：
```sql
source D:/k8s-go/k8s_operation/docs/sql/k8s_platform_full_init.sql
```

### ✅ 方式四：Linux / Git Bash / CMD
```bash
mysql -u root -p123456 --default-character-set=utf8mb4 < k8s_platform_full_init.sql
```

### ❌ 错误示范（PowerShell 直接用 `<` 会报错）
```powershell
# 报错："<"运算符是为将来使用而保留的
mysql -u root -p123456 < k8s_platform_full_init.sql
```

执行成功后会输出：
```
✅ 初始化完成 | 表数: 53 | 账号: admin / admin123 | 演示流水线: 4
```

## 📦 脚本文件

| 文件 | 说明 |
|------|------|
| `k8s_platform_full_init.sql` | **【唯一脚本】** 数据库 + 50 张表 DDL + 字典数据（资源模板 18 / 环境规则 5 / 流水线模板 4 / 默认 admin / 系统角色等） + 演示业务数据（4 流水线 + 6 运行 + 11 阶段 + 4 制品 + 2 监控数据源）。全新部署与存量升级均适用：Source 一次即可，幂等可重复执行；包含针对存量集群的兼容补丁（例如自动为老库 `monitor_datasource` 补 `cluster_id` 字段及索引） |

> 此前分散的 `ai_assistant.sql` / `cicd_artifact.sql` / `cicd_resource_config.sql` / `demo_seed.sql` / `migrate_monitor_datasource_cluster_id.sql` 已全部合并到主脚本，不再单独维护。
> 后续新增字段一律在 `k8s_platform_full_init.sql` 中同步提供 CREATE TABLE + 幂等 ALTER 补丁，保证全量/存量场景都只需跳一个脚本。

## 🗂️ 完整表清单（50 张）

### 👤 用户与权限（7）
1. `user` — 用户表
2. `sys_role` — 系统角色
3. `sys_permission` — 系统权限
4. `sys_role_permission` — 角色权限关联
5. `sys_user_role` — 用户角色关联
6. `sys_user_cluster` — 用户集群权限（细粒度）
7. `iam_project`, `iam_project_member`, `iam_group`, `iam_group_user`（IAM 体系）

### ☸️ K8s 集群（1）
8. `kube_cluster` — K8s 集群配置（含环境字段）

### 🔄 CI/CD 体系（14）
9. `cicd_pipeline` — 流水线
10. `cicd_pipeline_run` — 运行记录
11. `cicd_pipeline_stage` — 阶段记录
12. `cicd_environment` — 环境配置
13. `cicd_approval` — 审批记录
14. `cicd_release` — 发布单
15. `cicd_release_stage` — 发布阶段
16. `cicd_release_task` — 发布任务
17. `cicd_build` — 构建记录
18. `cicd_pipeline_template` — 流水线模板（Vue3/Go/Java/Python 4 条预置）
19. `cicd_artifact` — 制品库
20. `cicd_resource_template` — 资源模板（18 条预置）
21. `cicd_env_resource_rule` — 环境资源规则（5 条预置）
22. `cicd_deploy_approval` — 发布审批
23. `cicd_resource_change_log` — 资源变更日志
24. `cicd_build_agent` — **【新增】** 构建探针（OTEL Java Agent 已预置）

### 🐳 镜像管理（3）
25. `image_registry` — 镜像仓库
26. `image_cleanup_policy` — 清理策略
27. `image_cleanup_log` — 清理日志

### 🛍️ 应用商城（3）— **【新增】**
28. `app_store_apps` — 应用市场
29. `app_store_components` — 应用组件
30. `app_store_installs` — 安装记录

### 🤖 AI 助手（4）— **【新增】**
31. `ai_conversations` — 对话会话
32. `ai_messages` — 聊天消息（支持多轮上下文）
33. `ai_approval_requests` — 高危操作审批
34. `ai_approval_logs` — 审批操作日志

### 📊 监控告警（8）— **【新增】**
35. `monitor_datasource` — 数据源（Prometheus/Loki/Alertmanager）
36. `monitor_alert_rule` — 告警规则
37. `monitor_alert_event` — 告警事件
38. `monitor_notify_channel` — 通知渠道（钉钉/飞书/邮件/Webhook）
39. `monitor_silence_rule` — 静默规则
40. `monitor_inhibit_rule` — 抑制规则
41. `monitor_aggregate_rule` — 聚合规则
42. `monitor_notify_template` — 通知模板

### ⚙️ 平台配置（1）
43. `platform_settings` — 平台全局配置

## 🔑 默认账户

| 账户 | 密码 | 角色 |
|------|------|------|
| `admin` | `admin123` | 超级管理员 |

## 🎯 与代码自动迁移的关系

后端启动时会通过 GORM `AutoMigrate` 自动创建未存在的表，但**生产环境推荐先执行本 SQL 脚本**：

| 方式 | 优势 | 劣势 |
|------|------|------|
| **SQL 脚本（本文件）** | 字段精准、含索引、含种子数据、可重复执行 | 需要手动执行一次 |
| **代码 AutoMigrate** | 启动自动建表 | GORM 字段长度可能与预期不符、无种子数据 |

> 二者可叠加使用：SQL 先初始化结构和种子数据 → AutoMigrate 兜底补全后续新增字段。

## ✅ 验证安装

执行后在 MySQL 中验证：
```sql
USE `k8s-platform`;
SHOW TABLES;
-- 应当输出 50 条记录

SELECT COUNT(*) FROM cicd_resource_template;
-- 应当 = 18

SELECT username FROM user;
-- 应当包含 admin
```

## 🛠️ 重新初始化（危险操作）

主脚本全部使用 `INSERT IGNORE` + 显式主键/唯一键，**支持反复 source 不报错**。如需彻底清空再来一次：

```powershell
# 第一步：删库
mysql -u root -p123456 -e "DROP DATABASE IF EXISTS ``k8s-platform``;"

# 第二步：单脚本一键重建（数据库 + 50 表 + 字典 + 演示数据）
mysql -u root -p123456 --default-character-set=utf8mb4 -e "source D:/k8s-go/k8s_operation/docs/sql/k8s_platform_full_init.sql"
```

> ⚠️ `source` 是 mysql 客户端内部命令，不能与其它 SQL 通过 `;` 拼接给 `-e`，必须独立调用。

## 🎬 内置演示数据

主脚本已经把以下演示业务数据合并进来，初始化完成后前端各页面**开箱即有数据**，无需任何额外操作：

| 表 | 数量 | 说明 |
|----|------|------|
| `cicd_pipeline` | 4 | Go / Java / Vue3 / Python 各一条流水线 |
| `cicd_pipeline_run` | 6 | 含 5 次成功 + 1 次失败 |
| `cicd_pipeline_stage` | 11 | 完整 8 阶段成功 + 1 个 compile 失败案例 |
| `cicd_artifact` | 4 | binary / jar / dist / wheel 四种制品类型 |
| `monitor_datasource` | 2 | Prometheus（默认）+ Loki |
| `cicd_resource_template` | 18 | Java / Go / Node / Python × dev / test / prod |
| `cicd_env_resource_rule` | 5 | dev / test / staging / prod / prod-java |

依赖资源（主脚本同样自动创建）：`user.id=1` admin、`image_registry.id=1` 默认仓库、系统角色等。`kube_cluster` 在后端启动后自动注册。

### 验证内置数据

```powershell
mysql --default-character-set=utf8mb4 -u root -p123456 -e "USE ``k8s-platform``; SELECT COUNT(*) AS pipeline FROM cicd_pipeline; SELECT id,name,type,url FROM monitor_datasource;"
```

刷新前端 `http://127.0.0.1:5173`：
- **CI/CD → 流水线管理**：可见 4 条流水线（总数 4 / 上次成功 3 / 上次失败 1）
- **CI/CD → 制品管理**：可见 4 条制品
- **监控中心 → 数据源**：可见 Prometheus 和 Loki

### 🐛 常见问题

| 现象 | 原因 | 解决 |
|------|------|------|
| `ERROR 1366: Incorrect string value: '\xAB\xAF...'` | 默认 client 字符集不是 utf8mb4，中文字段写入失败 | mysql 命令一定要加 `--default-character-set=utf8mb4` |
| `ERROR 1146: Table 'k8s-platform.k8s_cluster' doesn't exist` | 表名误写，实际是 `kube_cluster`、`user`、`sys_role` | 参考本文「表清单」里的真实表名 |
| `ERROR 1064 syntax error near 'source ...'` | 把 `source` 与其它 SQL 通过 `-e` 一起传 | `source` 必须独立调用一次 mysql |
| PowerShell 控制台显示 `Prometheus-榛樿` 等乱码 | 终端 GBK 解码 utf8 字节流（仅显示问题） | 数据库内实际为正确 utf8mb4，前端展示正常，可忽略 |

## 📝 修改记录

| 版本 | 日期 | 变更 |
|------|------|------|
| 2.6.0 | 2026-05-18 | **多集群 × 多数据源**：`monitor_datasource` 新增 `cluster_id BIGINT NOT NULL DEFAULT 0` 字段及 `idx_cluster_id` 索引，支撑「监控视野按集群隔离」能力（0=全局共享、>0=集群专属）。所有逻辑合并进主脚本 `k8s_platform_full_init.sql`，同时依靠 `information_schema` 提供存量集群的幂等补丁（老库补字段/索引不报错）。**单脚本同时覆盖全新部署 + 存量升级两种场景**，不再需要单独的 `migrate_*.sql` |
| 2.5.0 | 2026-05-18 | **真正一键初始化**：合并 `ai_assistant.sql` / `cicd_artifact.sql` / `cicd_resource_config.sql` / `demo_seed.sql` 到主脚本（53 表 + 18 资源模板 + 5 环境规则 + 4 演示流水线 + 6 运行 + 11 阶段 + 4 制品 + 2 监控数据源）。全部 INSERT 改用 `INSERT IGNORE` + 显式主键/唯一键，支持反复 source 不报错 |
| 2.4.0 | 2026-05-18 | 新增 `demo_seed.sql` 演示种子数据脚本（4 流水线 + 6 运行 + 11 阶段 + 4 制品 + 2 监控数据源），新增「演示种子数据导入」章节及 PowerShell 字符集/source 顺序常见问题 |
| 2.3.0 | 2026-05-18 | 新增 AI 助手 4 表、监控 8 表、应用商城 3 表、构建探针 1 表（共 16 张），表总数 35 → 50 |
| 2.2.0 | 2026-03-15 | 加入制品库、资源模板、镜像清理 |
| 2.1.0 | — | RBAC 4 表 + IAM 项目 |
| 2.0.0 | — | CI/CD 体系完整化 |
