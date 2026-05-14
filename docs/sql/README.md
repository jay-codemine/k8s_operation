# K8sOperation 数据库初始化

## 快速开始

### 一键初始化
```bash
mysql -u root -padmin123 < k8s_platform_init.sql
```

### 指定主机初始化
```bash
mysql -h 127.0.0.1 -P 3306 -u root -padmin123 < k8s_platform_init.sql
```

## 文件说明

| 文件 | 说明 |
|------|------|
| `k8s_platform_init.sql` | 完整项目初始化SQL，包含所有表结构和初始数据 |
| `verify_tables.sql` | 数据库验证脚本，检查表结构完整性 |

## 数据库信息

- **数据库名**: `k8s-platform`
- **字符集**: `utf8mb4`
- **排序规则**: `utf8mb4_0900_ai_ci`

## 表结构清单 (共35+张表)

| 序号 | 表名 | 说明 |
|------|------|------|
| 1 | `user` | 用户表 |
| 2 | `kube_cluster` | K8s集群表 |
| 3 | `cicd_build` | CI/CD构建记录表 |
| 4 | `cicd_pipeline` | CI/CD流水线表 |
| 5 | `cicd_pipeline_run` | 流水线运行记录表 |
| 6 | `cicd_pipeline_stage` | 流水线阶段执行记录表 |
| 7 | `cicd_environment` | CI/CD环境管理表 |
| 8 | `cicd_approval` | CI/CD审批记录表 |
| 9 | `cicd_release` | CI/CD发布单表 |
| 10 | `cicd_release_stage` | CI/CD发布阶段表 |
| 11 | `cicd_release_task` | CI/CD发布任务表 |
| 12 | `ai_conversations` | AI助手会话表 |
| 13 | `ai_messages` | AI助手消息表 |
| 14 | `ai_approval_requests` | AI高危操作审批请求表 |
| 15 | `ai_approval_logs` | AI审批日志表 |
| 16 | `cicd_artifact` | CI/CD制品库表 |
| 17 | `cicd_resource_template` | CI/CD资源模板表 |
| 18 | `cicd_env_resource_rule` | CI/CD环境资源规则表 |
| 19 | `cicd_deploy_approval` | CI/CD发布审批表 |
| 20 | `cicd_resource_change_log` | CI/CD资源配置变更日志表 |
| ... | ... | ... |

## 默认账户

- **用户名**: `admin`
- **密码**: `admin123`

## 验证安装

执行验证脚本检查表结构：
```bash
mysql -u root -padmin123 < verify_tables.sql
```

或在 MySQL 命令行中执行：
```sql
source verify_tables.sql
```
