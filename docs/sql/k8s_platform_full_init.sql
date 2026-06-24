-- =====================================================
-- K8s Platform 完整数据库初始化脚本（FULL）
-- 版本: 2.4.1
-- 日期: 2026-06-24
-- 说明: 一键创建数据库 + 全部 54 张表 + 默认种子数据
--      已整合: 用户/RBAC/集群/CI-CD/制品/资源模板/镜像/IAM/应用商城/AI助手/监控/AIOps
--      新增: CICD细粒度权限(21条) + 多级审批 + 审批策略 + test测试环境配置
-- 数据库账号: root / 123456
-- 使用方式（任选其一）:
--   ⚠️ PowerShell 不支持 `<` 重定向，请用以下方式之一：
--   1) PowerShell（推荐）:
--      mysql -u root -p123456 --default-character-set=utf8mb4 -e "source D:/k8s-go/k8s_operation/docs/sql/k8s_platform_full_init.sql"
--   2) PowerShell + cmd 包装（支持 < 重定向）:
--      cmd /c "mysql -u root -p123456 --default-character-set=utf8mb4 < docs\sql\k8s_platform_full_init.sql"
--   3) MySQL 客户端内:
--      source D:/k8s-go/k8s_operation/docs/sql/k8s_platform_full_init.sql
--   4) Linux / Git Bash / CMD:
--      mysql -u root -p123456 --default-character-set=utf8mb4 < k8s_platform_full_init.sql
-- =====================================================


-- 创建数据库
CREATE DATABASE IF NOT EXISTS `k8s-platform` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `k8s-platform`;

-- =====================================================
-- 1. 用户表
-- =====================================================
CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码(加密)',
  `role` varchar(50) NOT NULL DEFAULT 'user' COMMENT '基础角色(兼容旧版)',
  `email` varchar(191) DEFAULT NULL COMMENT '邮箱',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用',
  `created_at` int unsigned DEFAULT 0,
  `modified_at` int unsigned DEFAULT 0,
  `deleted_at` int unsigned DEFAULT 0,
  `is_del` tinyint unsigned DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- =====================================================
-- 2. K8s集群表 (含环境字段)
-- =====================================================
CREATE TABLE IF NOT EXISTS `kube_cluster` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `cluster_name` varchar(191) NOT NULL DEFAULT '' COMMENT '集群名称',
  `kube_config` longtext NOT NULL COMMENT 'KubeConfig(加密存储)',
  `cluster_version` varchar(191) NOT NULL DEFAULT '' COMMENT '集群版本',
  `status` tinyint unsigned NOT NULL DEFAULT 2 COMMENT '状态:0=正常,1=异常,2=未检测',
  -- 环境相关字段
  `env_type` varchar(50) DEFAULT 'development' COMMENT '环境类型:development,testing,staging,production',
  `env_display_name` varchar(100) DEFAULT '' COMMENT '环境显示名称',
  `env_level` int DEFAULT 1 COMMENT '环境级别:1-4(开发到生产)',
  `access_mode` varchar(50) DEFAULT 'restricted' COMMENT '访问模式:public,restricted,private',
  `require_approval` tinyint(1) DEFAULT 0 COMMENT '操作是否需要审批',
  `approval_users` json DEFAULT NULL COMMENT '审批人列表',
  `env_color` varchar(20) DEFAULT '' COMMENT '环境颜色标识',
  `env_description` varchar(500) DEFAULT '' COMMENT '环境描述',
  `env_labels` json DEFAULT NULL COMMENT '环境标签',
  `project_ids` json DEFAULT NULL COMMENT '关联项目ID列表',
  -- 时间和状态字段
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  `last_check_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '最后检查时间',
  `last_error` varchar(1024) NOT NULL DEFAULT '' COMMENT '最后错误信息',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_modified` (`modified_at`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='K8s集群配置表';

-- =====================================================
-- 3. RBAC - 系统角色表
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_role` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '角色标识(唯一)',
  `display_name` varchar(100) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `role_type` varchar(30) NOT NULL DEFAULT 'custom' COMMENT '角色类型:super_admin,platform_admin,devops,developer,tester,viewer,custom',
  `scope_platform` varchar(10) NOT NULL DEFAULT 'none' COMMENT '平台域权限级别:none/read/write/admin',
  `scope_cluster` varchar(10) NOT NULL DEFAULT 'none' COMMENT '集群域权限级别:none/read/write/admin',
  `scope_cicd` varchar(10) NOT NULL DEFAULT 'none' COMMENT '发布域权限级别:none/read/write/admin',
  `is_system` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否系统内置',
  `color` varchar(20) DEFAULT '#1890ff' COMMENT '角色颜色标识',
  `icon` varchar(50) DEFAULT 'user' COMMENT '图标',
  `sort_order` int DEFAULT 0 COMMENT '排序',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统角色表';

-- 【兼容存量集群】sys_role 幂等补丁：若老库缺 scope_platform/scope_cluster/scope_cicd 字段，自动补上
SET @col_exists := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sys_role' AND column_name = 'scope_platform');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE `sys_role` ADD COLUMN `scope_platform` VARCHAR(10) NOT NULL DEFAULT ''none'' COMMENT ''平台域权限级别:none/read/write/admin'' AFTER `role_type`', 'SELECT ''scope_platform exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sys_role' AND column_name = 'scope_cluster');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE `sys_role` ADD COLUMN `scope_cluster` VARCHAR(10) NOT NULL DEFAULT ''none'' COMMENT ''集群域权限级别:none/read/write/admin'' AFTER `scope_platform`', 'SELECT ''scope_cluster exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sys_role' AND column_name = 'scope_cicd');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE `sys_role` ADD COLUMN `scope_cicd` VARCHAR(10) NOT NULL DEFAULT ''none'' COMMENT ''发布域权限级别:none/read/write/admin'' AFTER `scope_cluster`', 'SELECT ''scope_cicd exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- =====================================================
-- 4. RBAC - 系统权限表
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '权限标识(唯一)',
  `display_name` varchar(100) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `scope` varchar(20) NOT NULL DEFAULT 'cluster' COMMENT '所属功能域:platform/cluster/cicd',
  `resource_type` varchar(50) NOT NULL COMMENT '模块标识',
  `action` varchar(30) NOT NULL COMMENT '操作类型:view,manage',
  `tag` varchar(50) DEFAULT '' COMMENT '标签分组(前端分类展示)',
  `parent_id` bigint DEFAULT 0 COMMENT '父权限ID',
  `path` varchar(200) DEFAULT '' COMMENT '权限路径(树形展示用)',
  `sort_order` int DEFAULT 0 COMMENT '排序',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统权限定义表';

-- 【兼容存量集群】sys_permission 幂等补丁：若老库缺 scope 字段，自动补上
SET @col_exists := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sys_permission' AND column_name = 'scope');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE `sys_permission` ADD COLUMN `scope` VARCHAR(20) NOT NULL DEFAULT ''cluster'' COMMENT ''所属功能域:platform/cluster/cicd'' AFTER `description`', 'SELECT ''scope exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 【兼容存量集群】sys_permission 幂等补丁：若老库缺 tag 字段，自动补上
SET @col_exists := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sys_permission' AND column_name = 'tag');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE `sys_permission` ADD COLUMN `tag` VARCHAR(50) DEFAULT '''' COMMENT ''标签分组(前端分类展示)'' AFTER `action`', 'SELECT ''tag exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- =====================================================
-- 5. RBAC - 角色权限关联表
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_role_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `permission_id` bigint NOT NULL COMMENT '权限ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- =====================================================
-- 6. RBAC - 用户角色关联表
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `created_by` bigint DEFAULT 0 COMMENT '创建人ID',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- =====================================================
-- 7. RBAC - 用户集群权限表 (细粒度权限控制)
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_user_cluster` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `cluster_id` bigint NOT NULL COMMENT '集群ID',
  `role_type` varchar(30) NOT NULL DEFAULT 'viewer' COMMENT '在该集群的角色类型',
  `access_level` varchar(10) NOT NULL DEFAULT 'read' COMMENT '权限级别:none/read/write/admin',
  `namespaces` text DEFAULT NULL COMMENT '可访问的命名空间(JSON数组,空表示全部)',
  `can_view` tinyint(1) NOT NULL DEFAULT 1 COMMENT 'deprecated:查看权限',
  `can_create` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'deprecated:创建权限',
  `can_update` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'deprecated:更新权限',
  `can_delete` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'deprecated:删除权限',
  `can_exec` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'deprecated:执行权限',
  `expire_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '过期时间(0=永不过期)',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `created_by` bigint DEFAULT 0 COMMENT '授权人ID',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_cluster_id` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户集群权限表';

-- 【兼容存量集群】sys_user_cluster 幂等补丁：若老库缺 access_level 字段，自动补上
SET @col_exists := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sys_user_cluster' AND column_name = 'access_level');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE `sys_user_cluster` ADD COLUMN `access_level` VARCHAR(10) NOT NULL DEFAULT ''read'' COMMENT ''权限级别:none/read/write/admin'' AFTER `role_type`', 'SELECT ''access_level exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- =====================================================
-- 8. CI/CD - 流水线表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` bigint DEFAULT NULL COMMENT '所属项目ID',
  `name` varchar(191) NOT NULL COMMENT '流水线名称',
  `description` varchar(500) NOT NULL DEFAULT '' COMMENT '描述',
  `git_repo` varchar(500) NOT NULL COMMENT 'Git仓库地址',
  `git_branch` varchar(100) NOT NULL DEFAULT 'main' COMMENT 'Git分支',
  `jenkins_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Jenkins服务地址',
  `jenkins_job` varchar(191) NOT NULL COMMENT 'Jenkins Job名称',
  `jenkins_credential_id` varchar(191) NOT NULL DEFAULT '' COMMENT 'Jenkins凭证ID',
  `language_type` varchar(20) NOT NULL DEFAULT 'custom' COMMENT '语言类型:go/java/frontend/python/custom',
  -- 部署配置
  `auto_deploy` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否自动部署',
  `target_cluster_id` bigint DEFAULT NULL COMMENT '目标集群ID',
  `target_namespace` varchar(100) DEFAULT '' COMMENT '目标命名空间',
  `target_workload_kind` varchar(50) DEFAULT '' COMMENT '工作负载类型',
  `target_workload_name` varchar(200) DEFAULT '' COMMENT '工作负载名称',
  `target_container` varchar(100) DEFAULT '' COMMENT '目标容器名称',
  `deploy_env` varchar(20) DEFAULT 'dev' COMMENT '部署环境',
  `require_approval` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否需要审批',
  `enable_sonar` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用SonarQube代码扫描',
  `enable_artifact_upload` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用制品上传',
  -- 发布联动告警静默
  `enable_deploy_silence` tinyint(1) NOT NULL DEFAULT 0 COMMENT '部署时自动创建告警静默规则',
  `silence_buffer_minutes` int NOT NULL DEFAULT 10 COMMENT '静默缓冲时间(分钟)',
  `silence_severities` varchar(100) NOT NULL DEFAULT 'warning,info' COMMENT '静默的告警级别(逗号分隔)',
  -- 最新部署信息
  `last_deploy_image` varchar(500) DEFAULT '' COMMENT '最新部署镜像',
  `last_deploy_digest` varchar(100) DEFAULT '' COMMENT '镜像摘要',
  `last_deploy_time` bigint DEFAULT NULL COMMENT '最新部署时间',
  `last_deploy_status` varchar(32) DEFAULT '' COMMENT '最新部署状态',
  `last_deploy_version` varchar(100) DEFAULT '' COMMENT '最新部署版本',
  -- 运行状态
  `status` varchar(50) NOT NULL DEFAULT 'idle' COMMENT '状态:idle,running,disabled',
  `last_run_status` varchar(50) NOT NULL DEFAULT '' COMMENT '最后运行状态',
  `last_run_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '最后运行时间',
  `last_build_number` int NOT NULL DEFAULT 0 COMMENT '最后构建号',
  `last_build_url` varchar(500) NOT NULL DEFAULT '' COMMENT '最后构建URL',
  -- JSON配置
  `env_vars` json DEFAULT NULL COMMENT '环境变量',
  `deploy_config` json DEFAULT NULL COMMENT '部署配置',
  `stages` json DEFAULT NULL COMMENT '阶段配置',
  -- 元数据
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_pipeline_name_del` (`name`, `deleted_at`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_jenkins_job` (`jenkins_job`),
  KEY `idx_status` (`status`),
  KEY `idx_auto_deploy` (`auto_deploy`),
  KEY `idx_target_cluster` (`target_cluster_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD流水线表';

-- =====================================================
-- 9. CI/CD - 流水线运行记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline_run` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `build_number` int NOT NULL DEFAULT 0 COMMENT '构建号',
  `status` varchar(50) NOT NULL DEFAULT 'pending' COMMENT '状态:pending,running,success,failed,aborted',
  `trigger_type` varchar(50) NOT NULL DEFAULT 'manual' COMMENT '触发类型:manual,webhook,scheduled',
  `trigger_user_id` bigint NOT NULL DEFAULT 0 COMMENT '触发人ID',
  `git_commit` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git Commit',
  `git_branch` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit_message` varchar(500) NOT NULL DEFAULT '' COMMENT '提交消息',
  `jenkins_build_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Jenkins构建URL',
  `duration_sec` int NOT NULL DEFAULT 0 COMMENT '执行时长(秒)',
  `console_log` longtext DEFAULT NULL COMMENT '控制台日志',
  `stages_result` json DEFAULT NULL COMMENT '阶段结果',
  `started_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '开始时间',
  `finished_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '结束时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `error_message` text DEFAULT NULL COMMENT '错误信息',
  `image_url` varchar(500) DEFAULT '' COMMENT '构建镜像地址',
  `image_digest` varchar(100) DEFAULT '' COMMENT '镜像摘要',
  `callback_received` tinyint(1) DEFAULT 0 COMMENT '是否收到回调',
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_build_number` (`build_number`),
  KEY `idx_status` (`status`),
  KEY `idx_started_at` (`started_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线运行记录表';

-- =====================================================
-- 10. CI/CD - 流水线阶段执行记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline_stage` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `run_id` bigint NOT NULL COMMENT '运行记录ID',
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `stage_order` int NOT NULL DEFAULT 0 COMMENT '阶段顺序',
  `stage_type` varchar(32) NOT NULL COMMENT '阶段类型:checkout,dependencies,compile,test,lint,sonar,quality_gate,build_binary,upload_artifact,build,push,approval,deploy',
  `stage_name` varchar(100) NOT NULL COMMENT '阶段名称',
  `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '状态',
  `started_at` bigint DEFAULT NULL COMMENT '开始时间',
  `finished_at` bigint DEFAULT NULL COMMENT '结束时间',
  `duration_sec` int DEFAULT 0 COMMENT '执行时长',
  `logs` longtext DEFAULT NULL COMMENT '阶段日志',
  `jenkins_stage_id` varchar(100) DEFAULT NULL COMMENT 'Jenkins阶段ID',
  -- 审批信息
  `approval_user_id` bigint DEFAULT NULL COMMENT '审批人ID',
  `approval_comment` text DEFAULT NULL COMMENT '审批评论',
  `approval_decision` varchar(32) DEFAULT NULL COMMENT '审批决定:approved,rejected',
  -- 部署信息
  `deploy_cluster_id` bigint DEFAULT NULL COMMENT '部署集群ID',
  `deploy_namespace` varchar(100) DEFAULT NULL COMMENT '部署命名空间',
  `deploy_workload_kind` varchar(50) DEFAULT NULL COMMENT '工作负载类型',
  `deploy_workload_name` varchar(100) DEFAULT NULL COMMENT '工作负载名称',
  `deploy_container` varchar(100) DEFAULT NULL COMMENT '容器名称',
  `deploy_image` varchar(500) DEFAULT NULL COMMENT '部署镜像',
  `deploy_old_image` varchar(500) DEFAULT NULL COMMENT '旧镜像',
  `deploy_replicas` int DEFAULT NULL COMMENT '副本数',
  `error_message` text DEFAULT NULL COMMENT '错误信息',
  `created_at` bigint NOT NULL,
  `modified_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_run_id` (`run_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_stage_type` (`stage_type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线阶段执行记录表';

-- =====================================================
-- 11. CI/CD - 环境配置表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_environment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '环境标识:dev,staging,prod',
  `display_name` varchar(100) NOT NULL COMMENT '显示名称',
  `description` varchar(500) NOT NULL DEFAULT '' COMMENT '描述',
  `cluster_id` bigint NOT NULL COMMENT '关联集群ID',
  `namespace` varchar(100) NOT NULL DEFAULT '' COMMENT '默认命名空间',
  `color` varchar(20) NOT NULL DEFAULT '#1890ff' COMMENT '环境颜色',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序',
  `require_approval` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否需要审批',
  `approval_users` json DEFAULT NULL COMMENT '审批人列表',
  `approval_levels` json DEFAULT NULL COMMENT '多级审批级别配置(JSON数组)',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD环境配置表';

-- =====================================================
-- 12. CI/CD - 审批记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_approval` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `pipeline_run_id` bigint NOT NULL COMMENT '运行记录ID',
  `stage_id` bigint NOT NULL DEFAULT 0 COMMENT '关联流水线阶段ID',
  `release_id` bigint NOT NULL DEFAULT 0 COMMENT '发布单ID',
  `env_name` varchar(50) NOT NULL COMMENT '目标环境',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '状态:pending,approved,rejected,expired',
  `image` varchar(500) NOT NULL DEFAULT '' COMMENT '待部署镜像',
  `image_digest` varchar(100) NOT NULL DEFAULT '' COMMENT '镜像摘要',
  `request_user_id` bigint NOT NULL COMMENT '申请人ID',
  `request_reason` varchar(500) NOT NULL DEFAULT '' COMMENT '申请原因',
  `approve_user_id` bigint NOT NULL DEFAULT 0 COMMENT '审批人ID',
  `approve_reason` varchar(500) NOT NULL DEFAULT '' COMMENT '审批意见',
  `approve_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '审批时间',
  `expire_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '过期时间',
  `feishu_token` varchar(64) NOT NULL DEFAULT '' COMMENT '飞书审批回调Token',
  `approval_level` int NOT NULL DEFAULT 1 COMMENT '当前审批级别(1=一级,2=二级...)',
  `total_levels` int NOT NULL DEFAULT 1 COMMENT '总审批级数',
  `level_label` varchar(64) NOT NULL DEFAULT '' COMMENT '当前级别显示名称',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_pipeline_run_id` (`pipeline_run_id`),
  KEY `idx_stage_id` (`stage_id`),
  KEY `idx_status` (`status`),
  KEY `idx_request_user_id` (`request_user_id`),
  KEY `idx_feishu_token` (`feishu_token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD审批记录表';

-- 【兼容存量集群】cicd_approval 幂等补丁：若老库已存在该表但缺 stage_id / feishu_token 字段/索引，自动补上
SET @col_exists := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_approval'
      AND column_name = 'stage_id'
);
SET @sql := IF(@col_exists = 0,
    'ALTER TABLE `cicd_approval` ADD COLUMN `stage_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''关联流水线阶段ID'' AFTER `pipeline_run_id`',
    'SELECT ''column stage_id already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_approval'
      AND column_name = 'feishu_token'
);
SET @sql := IF(@col_exists = 0,
    'ALTER TABLE `cicd_approval` ADD COLUMN `feishu_token` VARCHAR(64) NOT NULL DEFAULT '''' COMMENT ''飞书审批回调Token'' AFTER `expire_time`',
    'SELECT ''column feishu_token already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx_exists := (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_approval'
      AND index_name = 'idx_stage_id'
);
SET @sql := IF(@idx_exists = 0,
    'ALTER TABLE `cicd_approval` ADD INDEX `idx_stage_id` (`stage_id`)',
    'SELECT ''index idx_stage_id already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx_exists := (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_approval'
      AND index_name = 'idx_feishu_token'
);
SET @sql := IF(@idx_exists = 0,
    'ALTER TABLE `cicd_approval` ADD INDEX `idx_feishu_token` (`feishu_token`)',
    'SELECT ''index idx_feishu_token already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 【兼容存量集群】cicd_approval 多级审批字段补丁 (v2.4)
SET @col_exists := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_approval'
      AND column_name = 'approval_level'
);
SET @sql := IF(@col_exists = 0,
    'ALTER TABLE `cicd_approval` ADD COLUMN `approval_level` INT NOT NULL DEFAULT 1 COMMENT ''当前审批级别'' AFTER `feishu_token`',
    'SELECT ''column approval_level already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_approval'
      AND column_name = 'total_levels'
);
SET @sql := IF(@col_exists = 0,
    'ALTER TABLE `cicd_approval` ADD COLUMN `total_levels` INT NOT NULL DEFAULT 1 COMMENT ''总审批级数'' AFTER `approval_level`',
    'SELECT ''column total_levels already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @col_exists := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_approval'
      AND column_name = 'level_label'
);
SET @sql := IF(@col_exists = 0,
    'ALTER TABLE `cicd_approval` ADD COLUMN `level_label` VARCHAR(64) NOT NULL DEFAULT '''' COMMENT ''当前级别显示名称'' AFTER `total_levels`',
    'SELECT ''column level_label already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 【兼容存量集群】cicd_environment 多级审批策略字段补丁 (v2.4)
SET @col_exists := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_environment'
      AND column_name = 'approval_levels'
);
SET @sql := IF(@col_exists = 0,
    'ALTER TABLE `cicd_environment` ADD COLUMN `approval_levels` JSON DEFAULT NULL COMMENT ''多级审批级别配置'' AFTER `approval_users`',
    'SELECT ''column approval_levels already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- =====================================================
-- 13. CI/CD - 发布单表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_release` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_name` varchar(191) NOT NULL DEFAULT '' COMMENT '应用名称',
  `namespace` varchar(191) NOT NULL DEFAULT 'default' COMMENT '命名空间',
  `workload_kind` varchar(32) NOT NULL DEFAULT 'Deployment' COMMENT '工作负载类型',
  `workload_name` varchar(191) NOT NULL DEFAULT '' COMMENT '工作负载名称',
  `container_name` varchar(191) NOT NULL DEFAULT '' COMMENT '容器名称',
  `strategy` varchar(32) NOT NULL DEFAULT 'rolling' COMMENT '发布策略',
  `timeout_sec` int unsigned NOT NULL DEFAULT 300 COMMENT '超时时间(秒)',
  `concurrency` int unsigned NOT NULL DEFAULT 3 COMMENT '并发数',
  `status` varchar(32) NOT NULL DEFAULT 'Pending' COMMENT '状态',
  `message` varchar(1024) NOT NULL DEFAULT '' COMMENT '消息',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `request_id` varchar(64) NOT NULL DEFAULT '' COMMENT '请求ID',
  `build_id` bigint NOT NULL DEFAULT 0 COMMENT '关联构建ID',
  `image_repo` varchar(512) NOT NULL DEFAULT '' COMMENT '镜像仓库',
  `image_tag` varchar(191) NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(255) DEFAULT NULL COMMENT '镜像摘要',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_request_id` (`request_id`),
  KEY `idx_app_name` (`app_name`),
  KEY `idx_status` (`status`),
  KEY `idx_build_id` (`build_id`),
  KEY `idx_modified_at` (`modified_at`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD发布单表';

-- =====================================================
-- 14. CI/CD - 发布阶段表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_release_stage` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `release_id` bigint NOT NULL DEFAULT 0 COMMENT '发布单ID',
  `stage_name` varchar(64) NOT NULL DEFAULT '' COMMENT '阶段名称',
  `stage_order` int NOT NULL DEFAULT 0 COMMENT '阶段顺序',
  `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '状态',
  `message` varchar(1024) NOT NULL DEFAULT '' COMMENT '消息',
  `logs` text DEFAULT NULL COMMENT '日志',
  `start_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '开始时间',
  `end_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '结束时间',
  `duration` bigint NOT NULL DEFAULT 0 COMMENT '持续时间',
  `build_number` varchar(64) NOT NULL DEFAULT '' COMMENT '构建号',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_release_id` (`release_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD发布阶段表';

-- =====================================================
-- 15. CI/CD - 发布任务表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_release_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `release_id` bigint NOT NULL COMMENT '发布单ID',
  `cluster_id` bigint NOT NULL COMMENT '集群ID',
  `status` varchar(32) NOT NULL DEFAULT 'Pending' COMMENT '状态',
  `message` varchar(2048) NOT NULL DEFAULT '' COMMENT '消息',
  `prev_image` varchar(512) NOT NULL DEFAULT '' COMMENT '原镜像',
  `target_image` varchar(512) NOT NULL DEFAULT '' COMMENT '目标镜像',
  `started_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '开始时间',
  `finished_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '结束时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_release_id` (`release_id`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD发布任务表';

-- =====================================================
-- 16. CI/CD - 构建记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_build` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_name` varchar(191) NOT NULL DEFAULT '' COMMENT '应用名称',
  `git_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Git URL',
  `git_branch` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git Commit',
  `jenkins_job` varchar(191) NOT NULL DEFAULT '' COMMENT 'Jenkins Job',
  `jenkins_queue_id` bigint NOT NULL DEFAULT 0 COMMENT 'Jenkins队列ID',
  `jenkins_build_number` int NOT NULL DEFAULT 0 COMMENT 'Jenkins构建号',
  `jenkins_build_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Jenkins构建URL',
  `status` varchar(50) NOT NULL DEFAULT 'pending' COMMENT '状态',
  `message` varchar(1024) NOT NULL DEFAULT '' COMMENT '消息',
  `image_repo` varchar(500) NOT NULL DEFAULT '' COMMENT '镜像仓库',
  `image_tag` varchar(191) NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(191) DEFAULT NULL COMMENT '镜像摘要',
  `sbom_ref` varchar(500) NOT NULL DEFAULT '' COMMENT 'SBOM引用',
  `sign_ref` varchar(500) NOT NULL DEFAULT '' COMMENT '签名引用',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_app_name` (`app_name`),
  KEY `idx_jenkins_job` (`jenkins_job`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD构建记录表';

-- =====================================================
-- 17. CI/CD - 流水线模板表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline_template` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(100) NOT NULL COMMENT '模板名称',
  `description` varchar(500) DEFAULT '' COMMENT '模板描述',
  `type` varchar(50) NOT NULL DEFAULT 'custom' COMMENT '模板类型: frontend/backend/microservice/database/custom',
  `stages` json DEFAULT NULL COMMENT '阶段配置',
  `default_env_vars` json DEFAULT NULL COMMENT '默认环境变量',
  `deploy_config` json DEFAULT NULL COMMENT '默认部署配置',
  `jenkins_template` text COMMENT 'Jenkinsfile模板',
  `usage_count` bigint NOT NULL DEFAULT 0 COMMENT '使用次数',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线模板表';

-- =====================================================
-- 18. 镜像仓库配置表
-- =====================================================
CREATE TABLE IF NOT EXISTS `image_registry` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '仓库名称',
  `type` varchar(50) NOT NULL DEFAULT 'docker' COMMENT '类型:docker,harbor,acr,ecr,gcr,quay',
  `url` varchar(500) NOT NULL COMMENT '仓库地址',
  `username` varchar(100) DEFAULT '' COMMENT '用户名',
  `password` varchar(500) DEFAULT '' COMMENT '密码(加密)',
  `access_key_id` varchar(100) DEFAULT '' COMMENT 'AccessKey ID(云厂商)',
  `access_key_secret` varchar(200) DEFAULT '' COMMENT 'AccessKey Secret(加密)',
  `region` varchar(50) DEFAULT '' COMMENT '区域',
  `insecure` tinyint(1) DEFAULT 0 COMMENT '跳过TLS验证',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `is_default` tinyint(1) DEFAULT 0 COMMENT '是否默认仓库',
  `status` varchar(50) DEFAULT 'unknown' COMMENT '连接状态',
  `last_check_at` bigint DEFAULT 0 COMMENT '最后检测时间',
  `last_error` varchar(500) DEFAULT '' COMMENT '最后错误',
  `created_by` bigint DEFAULT 0 COMMENT '创建人',
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_registry_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='镜像仓库配置表';

-- =====================================================
-- 19. 镜像清理策略表
-- =====================================================
CREATE TABLE IF NOT EXISTS `image_cleanup_policy` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `registry_id` bigint NOT NULL COMMENT '仓库ID',
  `name` varchar(100) NOT NULL COMMENT '策略名称',
  `enabled` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  `repository_pattern` varchar(200) DEFAULT '*' COMMENT '仓库匹配模式',
  `tag_pattern` varchar(200) DEFAULT '*' COMMENT '标签匹配模式',
  `keep_last_count` int DEFAULT 5 COMMENT '保留最近N个',
  `keep_days` int DEFAULT 30 COMMENT '保留N天内',
  `cron_expression` varchar(50) DEFAULT '0 2 * * *' COMMENT 'Cron表达式',
  `last_run_at` bigint DEFAULT 0 COMMENT '最后执行时间',
  `last_run_result` varchar(500) DEFAULT '' COMMENT '最后执行结果',
  `deleted_count` bigint DEFAULT 0 COMMENT '累计删除数',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `created_by` bigint DEFAULT 0 COMMENT '创建人',
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_registry_id` (`registry_id`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='镜像清理策略表';

-- =====================================================
-- 20. 镜像清理日志表
-- =====================================================
CREATE TABLE IF NOT EXISTS `image_cleanup_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `policy_id` bigint NOT NULL COMMENT '策略ID',
  `registry_id` bigint NOT NULL COMMENT '仓库ID',
  `start_time` bigint NOT NULL COMMENT '开始时间',
  `end_time` bigint DEFAULT 0 COMMENT '结束时间',
  `status` varchar(20) DEFAULT 'running' COMMENT '状态',
  `scanned_count` int DEFAULT 0 COMMENT '扫描数',
  `deleted_count` int DEFAULT 0 COMMENT '删除数',
  `freed_size` bigint DEFAULT 0 COMMENT '释放空间(字节)',
  `error_message` text DEFAULT NULL COMMENT '错误信息',
  `details` json DEFAULT NULL COMMENT '详情',
  PRIMARY KEY (`id`),
  KEY `idx_policy_id` (`policy_id`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='镜像清理日志表';

-- =====================================================
-- 21. 平台设置表
-- =====================================================
CREATE TABLE IF NOT EXISTS `platform_settings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `category` varchar(50) NOT NULL COMMENT '分类:basic,security,alert,notification',
  `key` varchar(100) NOT NULL COMMENT '设置键',
  `value` text DEFAULT NULL COMMENT '设置值',
  `value_type` varchar(20) DEFAULT 'string' COMMENT '值类型:string,int,bool,json',
  `label` varchar(100) DEFAULT NULL COMMENT '显示名称',
  `desc` varchar(500) DEFAULT NULL COMMENT '描述',
  `created_at` int unsigned DEFAULT NULL,
  `modified_at` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_category_key` (`category`, `key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台设置表';

-- =====================================================
-- 22. IAM - 项目表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_project` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '项目ID',
  `name` varchar(100) NOT NULL COMMENT '项目名称（唯一标识）',
  `display_name` varchar(191) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `status` varchar(50) NOT NULL DEFAULT 'active' COMMENT '状态: active/archived/disabled',
  `owner_id` bigint NOT NULL DEFAULT 0 COMMENT '项目负责人ID',
  `default_cluster_id` bigint DEFAULT NULL COMMENT '默认集群ID',
  `default_namespace` varchar(100) DEFAULT '' COMMENT '默认命名空间',
  `allowed_clusters` json DEFAULT NULL COMMENT '允许的集群ID列表',
  `allowed_namespaces` json DEFAULT NULL COMMENT '允许的命名空间列表（支持通配符）',
  `labels` json DEFAULT NULL COMMENT '标签（键值对）',
  `created_by` bigint NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '修改时间',
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  `is_del` tinyint unsigned NOT NULL DEFAULT 0 COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_owner_id` (`owner_id`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- =====================================================
-- 23. IAM - 项目成员关系表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_project_member` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `subject_type` varchar(50) NOT NULL COMMENT '主体类型: user/group',
  `subject_id` bigint NOT NULL COMMENT '主体ID（用户ID或组ID）',
  `role` varchar(50) NOT NULL DEFAULT 'viewer' COMMENT '项目角色: owner/admin/developer/viewer',
  `joined_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '加入时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_subject` (`project_id`, `subject_type`, `subject_id`),
  KEY `idx_subject` (`subject_type`, `subject_id`),
  KEY `idx_project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目成员关系表';

-- =====================================================
-- 24. IAM - 用户组表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_group` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户组ID',
  `name` varchar(100) NOT NULL COMMENT '组名称（唯一标识）',
  `display_name` varchar(191) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `type` varchar(50) NOT NULL DEFAULT 'custom' COMMENT '类型: system/custom',
  `parent_id` bigint DEFAULT NULL COMMENT '父组ID（支持层级结构）',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序顺序',
  `created_by` bigint NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '修改时间',
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  `is_del` tinyint unsigned NOT NULL DEFAULT 0 COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_type` (`type`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组表';

-- =====================================================
-- 25. IAM - 用户组成员关系表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_group_user` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `group_id` bigint NOT NULL COMMENT '用户组ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role` varchar(50) NOT NULL DEFAULT 'member' COMMENT '组内角色: owner/admin/member',
  `joined_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '加入时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组成员关系表';

-- =====================================================
-- 26. IAM - 权限模板表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_role_template` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '模板ID',
  `name` varchar(100) NOT NULL COMMENT '模板名称（唯一标识）',
  `display_name` varchar(191) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `type` varchar(50) NOT NULL COMMENT '模板类型: k8s/cicd/platform',
  `builtin` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否内置模板',
  `k8s_rules` json DEFAULT NULL COMMENT 'K8s RBAC规则 [{apiGroups, resources, verbs}]',
  `k8s_cluster_scope` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否集群级别权限',
  `cicd_actions` json DEFAULT NULL COMMENT 'CICD操作权限 ["view","run","approve","deploy","rollback","delete"]',
  `platform_permissions` json DEFAULT NULL COMMENT '平台功能权限 ["cluster:manage","user:manage","audit:view"]',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序顺序',
  `created_by` bigint NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '修改时间',
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  `is_del` tinyint unsigned NOT NULL DEFAULT 0 COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_builtin` (`builtin`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限模板表';

-- =====================================================
-- 27. IAM - 授权记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_grant` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '授权ID',
  `subject_type` varchar(50) NOT NULL COMMENT '主体类型: user/group/project',
  `subject_id` bigint NOT NULL COMMENT '主体ID',
  `subject_name` varchar(191) DEFAULT '' COMMENT '主体名称（冗余）',
  `scope_type` varchar(50) NOT NULL COMMENT '范围类型: cluster/namespace/cicd_project/cicd_pipeline',
  `scope_id` bigint DEFAULT NULL COMMENT '范围ID（集群ID/项目ID/流水线ID）',
  `scope_name` varchar(191) DEFAULT '' COMMENT '范围名称（冗余）',
  `namespaces` json DEFAULT NULL COMMENT '命名空间列表（支持通配符如 ["default","app-*"]）',
  `role_template_id` bigint NOT NULL COMMENT '权限模板ID',
  `role_template_name` varchar(100) DEFAULT '' COMMENT '模板名称（冗余）',
  `expire_at` bigint unsigned DEFAULT NULL COMMENT '过期时间（NULL 表示永不过期）',
  `k8s_synced` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'K8s RBAC 是否已同步',
  `k8s_role_name` varchar(191) DEFAULT '' COMMENT 'K8s Role/ClusterRole 名称',
  `k8s_binding_name` varchar(191) DEFAULT '' COMMENT 'K8s RoleBinding/ClusterRoleBinding 名称',
  `k8s_sync_error` varchar(500) DEFAULT '' COMMENT 'K8s 同步错误信息',
  `k8s_synced_at` bigint unsigned DEFAULT NULL COMMENT 'K8s 同步时间',
  `status` varchar(50) NOT NULL DEFAULT 'active' COMMENT '状态: active/expired/revoked',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  `granted_by` bigint NOT NULL DEFAULT 0 COMMENT '授权人ID',
  `revoked_by` bigint DEFAULT NULL COMMENT '撤销人ID',
  `revoked_at` bigint unsigned DEFAULT NULL COMMENT '撤销时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_subject` (`subject_type`, `subject_id`),
  KEY `idx_scope` (`scope_type`, `scope_id`),
  KEY `idx_role_template` (`role_template_id`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_granted_by` (`granted_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='授权记录表';

-- =====================================================
-- 28. IAM - 环境权限绑定表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_env_binding` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `subject_type` varchar(50) NOT NULL COMMENT '主体类型: user/group',
  `subject_id` bigint NOT NULL COMMENT '主体ID',
  `subject_name` varchar(191) DEFAULT '' COMMENT '主体名称',
  `cluster_id` bigint NOT NULL COMMENT '集群ID',
  `cluster_name` varchar(191) DEFAULT '' COMMENT '集群名称',
  `env_type` varchar(50) DEFAULT '' COMMENT '环境类型',
  `namespaces` json DEFAULT NULL COMMENT '命名空间列表',
  `env_role` varchar(50) NOT NULL COMMENT '环境角色',
  `custom_actions` json DEFAULT NULL COMMENT '自定义操作权限',
  `max_env_level` int DEFAULT 1 COMMENT '最高环境级别',
  `bypass_approval` tinyint(1) DEFAULT 0 COMMENT '是否跳过审批',
  `k8s_synced` tinyint(1) DEFAULT 0 COMMENT 'K8s RBAC是否已同步',
  `k8s_role_name` varchar(191) DEFAULT '' COMMENT 'K8s Role名称',
  `k8s_binding_name` varchar(191) DEFAULT '' COMMENT 'K8s RoleBinding名称',
  `k8s_sync_error` varchar(500) DEFAULT '' COMMENT 'K8s同步错误',
  `k8s_synced_at` bigint unsigned DEFAULT NULL COMMENT 'K8s同步时间',
  `expire_at` bigint unsigned DEFAULT NULL COMMENT '过期时间',
  `status` varchar(50) DEFAULT 'active' COMMENT '状态: active/expired/revoked',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  `granted_by` bigint DEFAULT 0 COMMENT '授权人ID',
  `revoked_by` bigint DEFAULT NULL COMMENT '撤销人ID',
  `revoked_at` bigint unsigned DEFAULT NULL COMMENT '撤销时间',
  `created_at` bigint unsigned NOT NULL COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_subject` (`subject_type`, `subject_id`),
  KEY `idx_cluster` (`cluster_id`),
  KEY `idx_env_type` (`env_type`),
  KEY `idx_env_role` (`env_role`),
  KEY `idx_status` (`status`),
  KEY `idx_k8s_synced` (`k8s_synced`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='环境权限绑定表';

-- =====================================================
-- 29. IAM - 环境操作审计日志表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_env_audit_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '操作用户ID',
  `username` varchar(191) DEFAULT '' COMMENT '用户名',
  `action` varchar(50) NOT NULL COMMENT '操作类型',
  `resource_type` varchar(50) NOT NULL COMMENT '资源类型',
  `resource_name` varchar(191) DEFAULT '' COMMENT '资源名称',
  `cluster_id` bigint DEFAULT NULL COMMENT '集群ID',
  `cluster_name` varchar(191) DEFAULT '' COMMENT '集群名称',
  `env_type` varchar(50) DEFAULT '' COMMENT '环境类型',
  `namespace` varchar(191) DEFAULT '' COMMENT '命名空间',
  `success` tinyint(1) DEFAULT 1 COMMENT '是否成功',
  `error_message` varchar(500) DEFAULT '' COMMENT '错误信息',
  `client_ip` varchar(50) DEFAULT '' COMMENT '客户端IP',
  `user_agent` varchar(500) DEFAULT '' COMMENT 'User-Agent',
  `request_id` varchar(64) DEFAULT '' COMMENT '请求ID',
  `detail` json DEFAULT NULL COMMENT '详情',
  `created_at` bigint unsigned NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_env_type` (`env_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='环境操作审计日志表';

-- =====================================================
-- 30. 审计日志表
-- =====================================================
CREATE TABLE IF NOT EXISTS `audit_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint NOT NULL COMMENT '操作用户ID',
  `username` varchar(191) NOT NULL COMMENT '操作用户名',
  `user_ip` varchar(50) DEFAULT '' COMMENT '用户IP',
  `user_agent` varchar(500) DEFAULT '' COMMENT 'User-Agent',
  `action` varchar(100) NOT NULL COMMENT '操作类型',
  `action_display` varchar(191) DEFAULT '' COMMENT '操作显示名称',
  `module` varchar(100) NOT NULL COMMENT '模块',
  `target_type` varchar(100) DEFAULT '' COMMENT '目标类型',
  `target_id` varchar(100) DEFAULT '' COMMENT '目标ID',
  `target_name` varchar(191) DEFAULT '' COMMENT '目标名称',
  `request_uri` varchar(500) DEFAULT '' COMMENT '请求URI',
  `request_method` varchar(10) DEFAULT '' COMMENT '请求方法',
  `request_body` text COMMENT '请求体',
  `response_code` int DEFAULT NULL COMMENT '响应状态码',
  `response_message` varchar(500) DEFAULT '' COMMENT '响应消息',
  `detail` json DEFAULT NULL COMMENT '操作详情',
  `extra` json DEFAULT NULL COMMENT '额外信息',
  `cluster_id` bigint DEFAULT NULL COMMENT '关联集群ID',
  `cluster_name` varchar(191) DEFAULT '' COMMENT '关联集群名称',
  `namespace` varchar(100) DEFAULT '' COMMENT '关联命名空间',
  `pipeline_id` bigint DEFAULT NULL COMMENT '关联流水线ID',
  `pipeline_name` varchar(191) DEFAULT '' COMMENT '关联流水线名称',
  `run_id` bigint DEFAULT NULL COMMENT '关联运行记录ID',
  `project_id` bigint DEFAULT NULL COMMENT '关联项目ID',
  `project_name` varchar(191) DEFAULT '' COMMENT '关联项目名称',
  `status` varchar(50) NOT NULL DEFAULT 'success' COMMENT '操作状态: success/failed',
  `error_message` varchar(1000) DEFAULT '' COMMENT '错误信息',
  `duration_ms` int DEFAULT 0 COMMENT '操作耗时(ms)',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_module` (`module`),
  KEY `idx_target` (`target_type`, `target_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';

-- =====================================================
-- 31. 视图 - 用户权限视图
-- =====================================================
CREATE OR REPLACE VIEW `v_user_permissions` AS
SELECT 
  g.user_id, g.subject_type, g.subject_id,
  g.scope_type, g.scope_id, g.scope_name, g.namespaces,
  g.role_template_id, rt.name AS role_template_name, rt.type AS role_type,
  rt.k8s_rules, rt.cicd_actions, rt.platform_permissions,
  g.expire_at, g.status
FROM (
  SELECT subject_id AS user_id, subject_type, subject_id, scope_type, scope_id, scope_name,
         namespaces, role_template_id, expire_at, status
  FROM iam_grant ig
  WHERE ig.subject_type = 'user' AND ig.status = 'active' 
    AND (ig.expire_at IS NULL OR ig.expire_at > UNIX_TIMESTAMP())
  UNION ALL
  SELECT igu.user_id, ig.subject_type, ig.subject_id, ig.scope_type, ig.scope_id, ig.scope_name,
         ig.namespaces, ig.role_template_id, ig.expire_at, ig.status
  FROM iam_grant ig
  JOIN iam_group_user igu ON ig.subject_id = igu.group_id
  WHERE ig.subject_type = 'group' AND ig.status = 'active'
    AND (ig.expire_at IS NULL OR ig.expire_at > UNIX_TIMESTAMP())
) g
LEFT JOIN iam_role_template rt ON g.role_template_id = rt.id AND rt.is_del = 0;

-- =====================================================
-- 32. 视图 - 用户环境权限视图
-- =====================================================
CREATE OR REPLACE VIEW `v_user_env_permissions` AS
SELECT 
  eb.subject_id AS user_id, eb.subject_name AS username,
  eb.cluster_id, eb.cluster_name, kc.env_type AS cluster_env_type, kc.env_level AS cluster_env_level,
  kc.access_mode, eb.env_role, eb.max_env_level, eb.bypass_approval,
  eb.namespaces, eb.status, eb.expire_at, eb.k8s_synced
FROM iam_env_binding eb
LEFT JOIN kube_cluster kc ON eb.cluster_id = kc.id
WHERE eb.subject_type = 'user' AND eb.status = 'active'
  AND (eb.expire_at IS NULL OR eb.expire_at > UNIX_TIMESTAMP())
  AND kc.is_del = 0;

-- =====================================================
-- 添加缺少的字段 (兼容已有数据库)
-- 注意: 如果字段已存在会报错，可忽略
-- =====================================================

-- image_registry 添加云厂商字段 (新安装无需执行，升级时按需执行)
-- 使用存储过程安全添加字段
DROP PROCEDURE IF EXISTS add_column_if_not_exists;
DELIMITER //
CREATE PROCEDURE add_column_if_not_exists()
BEGIN
    IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='k8s-platform' AND TABLE_NAME='image_registry' AND COLUMN_NAME='access_key_id') THEN
        ALTER TABLE `image_registry` ADD COLUMN `access_key_id` varchar(100) DEFAULT '' COMMENT 'AccessKey ID' AFTER `password`;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='k8s-platform' AND TABLE_NAME='image_registry' AND COLUMN_NAME='access_key_secret') THEN
        ALTER TABLE `image_registry` ADD COLUMN `access_key_secret` varchar(200) DEFAULT '' COMMENT 'AccessKey Secret' AFTER `access_key_id`;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='k8s-platform' AND TABLE_NAME='image_registry' AND COLUMN_NAME='region') THEN
        ALTER TABLE `image_registry` ADD COLUMN `region` varchar(50) DEFAULT '' COMMENT '区域' AFTER `access_key_secret`;
    END IF;
END //
DELIMITER ;
CALL add_column_if_not_exists();
DROP PROCEDURE IF EXISTS add_column_if_not_exists;

-- =====================================================
-- 初始化 RBAC 角色数据
-- =====================================================
INSERT IGNORE INTO `sys_role` (`id`, `name`, `display_name`, `description`, `role_type`, `scope_platform`, `scope_cluster`, `scope_cicd`, `is_system`, `color`, `icon`, `sort_order`, `created_at`, `modified_at`) VALUES
(1, 'super_admin',    '超级管理员', '拥有系统所有权限，可管理所有集群、用户和发布', 'super_admin',    'admin', 'admin', 'admin', 1, '#f5222d', 'crown',    1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2, 'platform_admin', '平台管理员', '管理用户/角色/系统设置，集群和CI/CD只读',     'platform_admin', 'admin', 'read',  'read',  1, '#fa8c16', 'setting',  2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(3, 'devops',         '运维工程师', '集群全权+CI/CD全权，平台域只读',                'devops',         'read',  'admin', 'admin', 1, '#722ed1', 'tool',     3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(4, 'developer',      '开发工程师', '集群域读写(指定NS)+CI/CD读写(自己的流水线)',    'developer',      'none',  'write', 'write', 1, '#1890ff', 'code',     4, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(5, 'tester',         '测试工程师', '集群域只读(指定NS)+CI/CD读写(测试环境流水线)',    'tester',         'none',  'read',  'write', 1, '#52c41a', 'bug',      5, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(6, 'viewer',         '观察者',     '全域只读，无任何修改权限',                        'viewer',         'read',  'read',  'read',  1, '#8c8c8c', 'eye',      6, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 初始化系统权限数据（v2: 按功能模块聚合，12条模块级权限）
-- =====================================================
INSERT IGNORE INTO `sys_permission` (`id`, `name`, `display_name`, `description`, `scope`, `resource_type`, `action`, `tag`, `parent_id`, `path`, `sort_order`, `created_at`, `modified_at`) VALUES
-- 🏛 平台域
(1,  'platform:user',     '用户管理',   '创建/编辑/禁用用户，分配角色',           'platform', 'user',     'manage', '', 0, '/platform/user',     1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2,  'platform:role',     '角色权限',   '管理角色、分配权限、集群授权',         'platform', 'role',     'manage', '', 0, '/platform/role',     2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(3,  'platform:settings', '系统设置',   '平台参数/告警配置/数据源管理',           'platform', 'settings', 'manage', '', 0, '/platform/settings', 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(4,  'platform:audit',    '审计日志',   '查看操作审计日志',                       'platform', 'audit',    'view',   '', 0, '/platform/audit',    4, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- ☸ 集群域
(5,  'cluster:manage',    '集群管理',   '添加/编辑/删除K8s集群',                 'cluster',  'cluster',  'manage', '', 0, '/cluster/manage',    10, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(6,  'cluster:workload',  '工作负载',   'Deployment/Pod/DaemonSet/Job等生命周期管理', 'cluster',  'workload', 'manage', '', 0, '/cluster/workload',  11, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(7,  'cluster:network',   '服务与路由', 'Service/Ingress/NetworkPolicy',            'cluster',  'network',  'manage', '', 0, '/cluster/network',   12, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(8,  'cluster:config',    '配置管理',   'ConfigMap/Secret管理',                   'cluster',  'config',   'manage', '', 0, '/cluster/config',    13, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(9,  'cluster:storage',   '存储管理',   'PV/PVC/StorageClass管理',               'cluster',  'storage',  'manage', '', 0, '/cluster/storage',   14, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(10, 'cluster:node',      '节点管理',   '节点查看/隔离/驱逐/标签管理',               'cluster',  'node',     'manage', '', 0, '/cluster/node',      15, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(11, 'cluster:monitor',   '监控与日志', '监控总览/日志探索/告警查看/Web终端',       'cluster',  'monitor',  'manage', '', 0, '/cluster/monitor',   16, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 🚀 发布域
(12, 'cicd:pipeline',     '流水线管理', '创建/编辑/触发/停止流水线，审批管理',     'cicd',     'pipeline', 'manage', '流水线管理', 0, '/cicd/pipeline',     20, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(13, 'cicd:artifact',     '制品与镜像', '制品库/镜像仓库/代码扫描/发布记录',       'cicd',     'artifact', 'manage', '制品与镜像', 0, '/cicd/artifact',     21, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 初始化 CICD 细粒度权限数据（v2.4: 21条细粒度权限，按标签分组）
-- =====================================================
INSERT IGNORE INTO `sys_permission` (`name`, `display_name`, `description`, `scope`, `resource_type`, `action`, `tag`, `parent_id`, `path`, `sort_order`, `created_at`, `modified_at`) VALUES
-- 📦 流水线管理
('cicd:pipeline:view',     '查看流水线',     '查看流水线列表和详情',     'cicd', 'pipeline', 'view',   '流水线管理', 0, '/cicd/pipeline/view',     100, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:pipeline:create',   '创建流水线',     '创建新的流水线配置',     'cicd', 'pipeline', 'create', '流水线管理', 0, '/cicd/pipeline/create',   101, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:pipeline:edit',     '编辑流水线',     '修改流水线配置参数',     'cicd', 'pipeline', 'update', '流水线管理', 0, '/cicd/pipeline/edit',     102, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:pipeline:delete',   '删除流水线',     '删除流水线配置',         'cicd', 'pipeline', 'delete', '流水线管理', 0, '/cicd/pipeline/delete',   103, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:pipeline:run',      '运行流水线',     '触发/运行流水线构建',     'cicd', 'pipeline', 'exec',   '流水线管理', 0, '/cicd/pipeline/run',      104, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 🚀 构建与部署
('cicd:build:view',        '查看构建记录',   '查看构建历史和日志',     'cicd', 'build',    'view',   '构建与部署', 0, '/cicd/build/view',        110, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:build:trigger',     '触发构建',       '手动触发代码构建',       'cicd', 'build',    'exec',   '构建与部署', 0, '/cicd/build/trigger',     111, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:build:cancel',      '取消构建',       '取消正在进行的构建',     'cicd', 'build',    'exec',   '构建与部署', 0, '/cicd/build/cancel',      112, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:deploy:dev',        '部署开发环境',     '部署到开发环境',       'cicd', 'deploy',   'exec',   '构建与部署', 0, '/cicd/deploy/dev',        113, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:deploy:test',       '部署测试环境',     '部署到测试环境',       'cicd', 'deploy',   'exec',   '构建与部署', 0, '/cicd/deploy/test',       114, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:deploy:prod',       '部署生产环境',     '部署到生产环境',       'cicd', 'deploy',   'exec',   '构建与部署', 0, '/cicd/deploy/prod',       115, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:deploy:rollback',   '回滚部署',       '回滚到上一版本',       'cicd', 'deploy',   'exec',   '构建与部署', 0, '/cicd/deploy/rollback',   116, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 📦 制品与镜像
('cicd:artifact:view',     '查看制品',       '查看制品库和镜像列表',   'cicd', 'artifact', 'view',   '制品与镜像', 0, '/cicd/artifact/view',     120, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:artifact:upload',   '上传制品',       '上传制品到制品库',     'cicd', 'artifact', 'create', '制品与镜像', 0, '/cicd/artifact/upload',   121, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:artifact:delete',   '删除制品',       '从制品库删除制品',     'cicd', 'artifact', 'delete', '制品与镜像', 0, '/cicd/artifact/delete',   122, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:image:manage',      '镜像管理',       '管理镜像仓库(推送/删除/清理)', 'cicd', 'image',    'manage', '制品与镜像', 0, '/cicd/image/manage',      123, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- ✅ 审批与管理
('cicd:approval:view',     '查看审批',       '查看审批记录和详情',   'cicd', 'approval', 'view',   '审批与管理', 0, '/cicd/approval/view',     130, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:approval:action',   '执行审批',       '通过或拒绝审批申请',   'cicd', 'approval', 'exec',   '审批与管理', 0, '/cicd/approval/action',   131, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:approval:manage',   '管理审批策略',   '配置审批流程和审批人',   'cicd', 'approval', 'manage', '审批与管理', 0, '/cicd/approval/manage',   132, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:environment:manage','环境管理',       '创建/编辑/删除环境配置',   'cicd', 'environment','manage','审批与管理', 0, '/cicd/environment/manage', 133, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('cicd:template:manage',   '模板管理',       '管理流水线模板',       'cicd', 'template', 'manage', '审批与管理', 0, '/cicd/template/manage',   134, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 初始化角色权限关联 (v2: 基于三域 scope，模块级权限作为补充细控)
-- 超级管理员: 全部权限
-- =====================================================
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 1, id, UNIX_TIMESTAMP() FROM `sys_permission`;

-- 平台管理员: 平台域全部 + 集群域只读 + CI/CD只读
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 2, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `scope` = 'platform';
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 2, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `scope` IN ('cluster', 'cicd') AND `action` = 'view';

-- 运维工程师: 集群域全部 + CI/CD全部 + 平台域只读
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 3, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `scope` IN ('cluster', 'cicd');
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 3, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `scope` = 'platform' AND `action` = 'view';

-- 开发工程师: 集群域(工作负载/网络/配置/监控) + CI/CD全部
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 4, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `name` IN ('cluster:workload','cluster:network','cluster:config','cluster:monitor','cicd:pipeline','cicd:artifact');

-- 测试工程师: 集群域只读(工作负载/监控) + CI/CD流水线
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 5, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `name` IN ('cluster:workload','cluster:monitor','cicd:pipeline');

-- 观察者: 所有域的 view 类权限
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 6, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `action` = 'view';

-- =====================================================
-- CICD 细粒度权限角色分配（v2.4）
-- 超级管理员(1) + 运维(3): 全部 CICD 权限（已通过上方 scope=cicd 全量分配覆盖）
-- =====================================================

-- 开发工程师(4): 查看+创建+编辑+运行流水线 + 构建查看+触发+取消 + 部署开发/测试 + 查看制品 + 查看审批
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 4, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `name` IN (
  'cicd:pipeline:view', 'cicd:pipeline:create', 'cicd:pipeline:edit', 'cicd:pipeline:run',
  'cicd:build:view', 'cicd:build:trigger', 'cicd:build:cancel',
  'cicd:deploy:dev', 'cicd:deploy:test',
  'cicd:artifact:view', 'cicd:approval:view'
);

-- 测试工程师(5): 查看+运行流水线 + 构建查看+触发 + 部署开发/测试 + 查看制品 + 查看审批
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 5, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `name` IN (
  'cicd:pipeline:view', 'cicd:pipeline:run',
  'cicd:build:view', 'cicd:build:trigger',
  'cicd:deploy:dev', 'cicd:deploy:test',
  'cicd:artifact:view', 'cicd:approval:view'
);

-- 观察者(6): 仅查看类权限
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 6, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `scope` = 'cicd' AND `action` = 'view';

-- =====================================================
-- 【兼容存量集群】RBAC v2 存量数据回填
-- 说明: 全新库 INSERT IGNORE 已覆盖; 存量库仅当 scope 全为 none 时才回填
-- =====================================================

-- 旧角色 scope 回填（仅当三域均为默认 none 时才触发，不会覆盖已配置的值）
UPDATE `sys_role` SET `scope_platform`='admin', `scope_cluster`='admin', `scope_cicd`='admin'
WHERE `role_type`='super_admin' AND `scope_platform`='none' AND `scope_cluster`='none' AND `scope_cicd`='none';

UPDATE `sys_role` SET `scope_platform`='admin', `scope_cluster`='read', `scope_cicd`='read'
WHERE `role_type`='platform_admin' AND `scope_platform`='none' AND `scope_cluster`='none' AND `scope_cicd`='none';

UPDATE `sys_role` SET `role_type`='devops', `scope_platform`='read', `scope_cluster`='admin', `scope_cicd`='admin'
WHERE `role_type`='cluster_admin' AND `scope_platform`='none' AND `scope_cluster`='none' AND `scope_cicd`='none';

UPDATE `sys_role` SET `scope_platform`='none', `scope_cluster`='write', `scope_cicd`='write'
WHERE `role_type`='developer' AND `scope_platform`='none' AND `scope_cluster`='none' AND `scope_cicd`='none';

UPDATE `sys_role` SET `scope_platform`='read', `scope_cluster`='read', `scope_cicd`='read'
WHERE `role_type`='viewer' AND `scope_platform`='none' AND `scope_cluster`='none' AND `scope_cicd`='none';

-- sys_user_cluster.access_level 回填（从旧 bool 字段推导）
UPDATE `sys_user_cluster` SET `access_level`='admin'  WHERE `access_level`='read' AND `can_delete`=1;
UPDATE `sys_user_cluster` SET `access_level`='write'  WHERE `access_level`='read' AND `can_delete`=0 AND (`can_create`=1 OR `can_update`=1 OR `can_exec`=1);
UPDATE `sys_user_cluster` SET `access_level`='none'   WHERE `access_level`='read' AND `can_view`=0 AND `can_create`=0 AND `can_update`=0 AND `can_delete`=0 AND `can_exec`=0;

-- sys_permission.scope 回填（旧权限的 scope 全是默认值 cluster，按 resource_type 修正）
UPDATE `sys_permission` SET `scope`='platform' WHERE `scope`='cluster' AND `resource_type` IN ('user','role','settings','audit','permission');
UPDATE `sys_permission` SET `scope`='cicd' WHERE `scope`='cluster' AND `resource_type` IN ('pipeline','artifact','release','approval','build','image');

-- =====================================================
-- 初始化管理员账户 (admin/admin123)
-- =====================================================
INSERT IGNORE INTO `user` (`id`, `username`, `password`, `role`, `status`, `created_at`, `modified_at`) VALUES
(1, 'admin', '$2a$10$jWcwxJ.3qLlHaXVZ1nL7MeCsSEXGosmaj1dIFoS74WXq5.gJrfChO', 'admin', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 关联admin用户为超级管理员
INSERT IGNORE INTO `sys_user_role` (`user_id`, `role_id`, `created_at`, `created_by`) VALUES
(1, 1, UNIX_TIMESTAMP(), 0);

-- ⚠️ 安全保护：确保 admin 用户始终拥有 super_admin 角色（防止意外自降权后无法恢复）
-- 如果 sys_user_role 中 admin 的 super_admin 关联被删除，此语句会重新插入
INSERT IGNORE INTO `sys_user_role` (`user_id`, `role_id`, `created_at`, `created_by`)
SELECT 1, id, UNIX_TIMESTAMP(), 0 FROM `sys_role` WHERE `role_type` = 'super_admin' AND `is_del` = 0 LIMIT 1;

-- =====================================================
-- 初始化平台设置
-- =====================================================
INSERT IGNORE INTO `platform_settings` (`category`, `key`, `value`, `value_type`, `label`, `desc`, `created_at`, `modified_at`) VALUES
-- 基本设置
('basic', 'default_page', '/clusters', 'string', '默认首页', '登录后默认跳转页面', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('basic', 'default_cluster', 'auto', 'string', '默认集群', '自动选择或指定集群ID', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('basic', 'language', 'zh-CN', 'string', '系统语言', '界面显示语言', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('basic', 'timezone', 'Asia/Shanghai', 'string', '时区', '系统时区设置', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 安全设置
('security', 'session_timeout', '120', 'int', '会话超时', '会话超时时间(分钟)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('security', 'enable_2fa', 'false', 'bool', '双因素认证', '是否启用双因素认证', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('security', 'password_policy', 'medium', 'string', '密码策略', '密码复杂度要求', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('security', 'audit_retention', '30', 'int', '审计保留', '审计日志保留天数', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 告警设置
('alert', 'cpu_threshold', '80', 'int', 'CPU告警阈值', 'CPU使用率告警阈值(%)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('alert', 'mem_threshold', '80', 'int', '内存告警阈值', '内存使用率告警阈值(%)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('alert', 'disk_threshold', '85', 'int', '磁盘告警阈值', '磁盘使用率告警阈值(%)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('alert', 'alert_silence', '15', 'int', '告警静默', '告警静默时间(分钟)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 通知设置
('notification', 'enable_email', 'false', 'bool', '邮件通知', '是否启用邮件通知', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'smtp_server', '', 'string', 'SMTP服务器', 'SMTP服务器地址', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'enable_dingtalk', 'false', 'bool', '钉钉通知', '是否启用钉钉通知', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'dingtalk_webhook', '', 'string', '钉钉Webhook', '钉钉机器人Webhook地址', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'enable_webhook', 'false', 'bool', 'Webhook通知', '是否启用自定义Webhook', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'webhook_url', '', 'string', 'Webhook地址', '自定义Webhook地址', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 示例镜像仓库配置
-- =====================================================
INSERT IGNORE INTO `image_registry` (`name`, `type`, `url`, `description`, `is_default`, `status`, `created_at`, `modified_at`) VALUES
('Docker Hub', 'docker', 'https://registry-1.docker.io', 'Docker Hub 官方仓库', 1, 'unknown', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 初始化审批策略环境配置 (dev/test/staging/prod)
-- =====================================================
INSERT IGNORE INTO `cicd_environment` (`name`, `display_name`, `description`, `cluster_id`, `namespace`, `color`, `sort_order`, `require_approval`, `approval_levels`, `created_at`, `modified_at`, `is_del`) VALUES
('dev',     '开发环境', '开发编译调试用，无需审批',                                            0, 'dev',     '#52c41a', 1, 0, '[]', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
('test',    '测试环境', '功能测试与集成测试环境，默认无需审批，管理员可按需开启',               0, 'test',    '#1677ff', 2, 0, '[]', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
('staging', '预发环境', '上线前的最后验证环境，需测试负责人审批',                              0, 'staging', '#faad14', 3, 1, '[{"label":"测试负责人审批","approver_type":"role","approver_value":"tester"}]', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
('prod',    '生产环境', '正式生产环境，需多级审批（运维经理+技术总监）',                     0, 'prod',    '#f5222d', 4, 1, '[{"label":"运维经理审批","approver_type":"role","approver_value":"devops"},{"label":"技术总监审批","approver_type":"role","approver_value":"platform_admin"}]', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);

-- =====================================================
-- 初始化流水线模板数据
-- =====================================================
INSERT IGNORE INTO `cicd_pipeline_template` (`id`, `name`, `description`, `type`, `stages`, `default_env_vars`, `deploy_config`, `created_at`, `modified_at`) VALUES
(
  1,
  'Vue3 前端应用模板',
  '适用于 Vue3 + Vite 前端项目的标准流水线模板',
  'frontend',
  '[{"name": "checkout", "description": "拉取代码", "order": 1}, {"name": "install", "description": "安装依赖 (npm install)", "order": 2}, {"name": "build", "description": "构建应用 (npm run build)", "order": 3}, {"name": "test", "description": "运行测试 (npm run test)", "order": 4}, {"name": "build-image", "description": "构建 Docker 镜像", "order": 5}, {"name": "deploy", "description": "部署到 K8s", "order": 6}]',
  '[{"name": "NODE_ENV", "value": "production"}, {"name": "VITE_API_BASE", "value": "https://api.example.com"}]',
  '{"replicas": 3, "strategy": "rollingUpdate", "resources": {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "200m", "memory": "256Mi"}}}',
  UNIX_TIMESTAMP(),
  UNIX_TIMESTAMP()
),
(
  2,
  'Go 微服务模板',
  '适用于 Go 语言微服务的标准流水线模板',
  'backend',
  '[{"name": "checkout", "description": "拉取代码", "order": 1}, {"name": "test", "description": "运行单元测试 (go test)", "order": 2}, {"name": "build", "description": "编译 Go 二进制", "order": 3}, {"name": "build-image", "description": "构建 Docker 镜像", "order": 4}, {"name": "push", "description": "推送镜像到 Harbor", "order": 5}, {"name": "deploy", "description": "部署到 K8s", "order": 6}]',
  '[{"name": "GO111MODULE", "value": "on"}, {"name": "CGO_ENABLED", "value": "0"}]',
  '{"replicas": 2, "strategy": "rollingUpdate", "resources": {"limits": {"cpu": "1000m", "memory": "1024Mi"}, "requests": {"cpu": "500m", "memory": "512Mi"}}}',
  UNIX_TIMESTAMP(),
  UNIX_TIMESTAMP()
),
(
  3,
  'Java Spring Boot 模板',
  '适用于 Java Spring Boot 项目的标准流水线模板',
  'backend',
  '[{"name": "checkout", "description": "拉取代码", "order": 1}, {"name": "compile", "description": "Maven 编译 (mvn compile)", "order": 2}, {"name": "test", "description": "运行测试 (mvn test)", "order": 3}, {"name": "package", "description": "打包 (mvn package)", "order": 4}, {"name": "build-image", "description": "构建 Docker 镜像", "order": 5}, {"name": "deploy", "description": "部署到 K8s", "order": 6}]',
  '[{"name": "JAVA_HOME", "value": "/usr/lib/jvm/java-17"}, {"name": "MAVEN_OPTS", "value": "-Xmx1024m"}]',
  '{"replicas": 2, "strategy": "rollingUpdate", "resources": {"limits": {"cpu": "2000m", "memory": "2048Mi"}, "requests": {"cpu": "1000m", "memory": "1024Mi"}}}',
  UNIX_TIMESTAMP(),
  UNIX_TIMESTAMP()
),
(
  4,
  'Python Flask 模板',
  '适用于 Python Flask 项目的标准流水线模板',
  'backend',
  '[{"name": "checkout", "description": "拉取代码", "order": 1}, {"name": "install", "description": "安装依赖 (pip install)", "order": 2}, {"name": "test", "description": "运行测试 (pytest)", "order": 3}, {"name": "build-image", "description": "构建 Docker 镜像", "order": 4}, {"name": "deploy", "description": "部署到 K8s", "order": 5}]',
  '[{"name": "PYTHON_VERSION", "value": "3.11"}, {"name": "PIP_INDEX_URL", "value": "https://pypi.tuna.tsinghua.edu.cn/simple"}]',
  '{"replicas": 2, "strategy": "rollingUpdate", "resources": {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "200m", "memory": "256Mi"}}}',
  UNIX_TIMESTAMP(),
  UNIX_TIMESTAMP()
);

-- =====================================================
-- 30. CICD - 资源档位模板表
-- =====================================================
CREATE TABLE IF NOT EXISTS cicd_resource_template (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(64) NOT NULL COMMENT '模板名称：small/medium/large/custom',
    service_type    VARCHAR(32) NOT NULL COMMENT '服务类型：java/go/node/python',
    env             VARCHAR(16) NOT NULL COMMENT '环境：dev/test/staging/prod',
    
    -- 资源配置
    replicas_default INT NOT NULL DEFAULT 1 COMMENT '默认副本数',
    replicas_min     INT NOT NULL DEFAULT 1 COMMENT '最小副本数',
    replicas_max     INT NOT NULL DEFAULT 10 COMMENT '最大副本数',
    
    cpu_request      VARCHAR(16) NOT NULL DEFAULT '200m',
    cpu_limit        VARCHAR(16) NOT NULL DEFAULT '500m',
    memory_request   VARCHAR(16) NOT NULL DEFAULT '256Mi',
    memory_limit     VARCHAR(16) NOT NULL DEFAULT '512Mi',
    
    -- HPA 配置
    hpa_enabled      TINYINT(1) DEFAULT 0 COMMENT '是否启用HPA',
    hpa_min_replicas INT DEFAULT 2,
    hpa_max_replicas INT DEFAULT 10,
    hpa_cpu_target   INT DEFAULT 70 COMMENT 'CPU目标利用率%',
    
    description      VARCHAR(255) DEFAULT '' COMMENT '模板说明',
    is_default       TINYINT(1) DEFAULT 0 COMMENT '是否默认模板',
    sort_order       INT DEFAULT 0,
    created_at       BIGINT UNSIGNED DEFAULT 0,
    modified_at      BIGINT UNSIGNED DEFAULT 0,
    deleted_at       BIGINT UNSIGNED DEFAULT 0,
    
    UNIQUE KEY uk_type_env_name (service_type, env, name, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD资源档位模板表';

-- =====================================================
-- 31. CICD - 环境资源规则表
-- =====================================================
CREATE TABLE IF NOT EXISTS cicd_env_resource_rule (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    env             VARCHAR(16) NOT NULL COMMENT 'dev/test/staging/prod',
    service_type    VARCHAR(32) DEFAULT '' COMMENT '服务类型，空=通用',
    
    -- 资源上限
    cpu_limit_max       VARCHAR(16) NOT NULL DEFAULT '4' COMMENT 'CPU limit 最大值',
    memory_limit_max    VARCHAR(16) NOT NULL DEFAULT '8Gi' COMMENT '内存 limit 最大值',
    replicas_max        INT NOT NULL DEFAULT 10 COMMENT '副本数上限',
    
    -- 资源下限（生产环境）
    cpu_request_min     VARCHAR(16) DEFAULT '' COMMENT 'CPU request 最小值',
    memory_request_min  VARCHAR(16) DEFAULT '' COMMENT '内存 request 最小值',
    replicas_min        INT DEFAULT 1 COMMENT '副本数下限',
    
    -- 审批规则
    require_approval    TINYINT(1) DEFAULT 0 COMMENT '是否需要审批',
    approval_role       VARCHAR(64) DEFAULT '' COMMENT '审批角色：sre/admin',
    
    description         VARCHAR(255) DEFAULT '' COMMENT '规则说明',
    created_at          BIGINT UNSIGNED DEFAULT 0,
    modified_at         BIGINT UNSIGNED DEFAULT 0,
    
    UNIQUE KEY uk_env_type (env, service_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD环境资源规则表';

-- =====================================================
-- 32. CICD - 发布审批记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS cicd_deploy_approval (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    pipeline_id     BIGINT UNSIGNED NOT NULL COMMENT '流水线ID',
    release_id      BIGINT UNSIGNED DEFAULT 0 COMMENT '发布单ID',
    env             VARCHAR(16) NOT NULL COMMENT '目标环境',
    
    -- 申请配置（JSON）
    requested_config TEXT COMMENT '申请的资源配置JSON',
    current_config   TEXT COMMENT '当前线上配置JSON',
    
    -- 风险提示
    risk_level       ENUM('low','medium','high') DEFAULT 'low',
    risk_warnings    TEXT COMMENT '风险提示列表JSON',
    
    -- 审批流程
    status           ENUM('pending','approved','rejected','expired','cancelled') DEFAULT 'pending',
    applicant_id     BIGINT UNSIGNED NOT NULL COMMENT '申请人ID',
    applicant_name   VARCHAR(64) NOT NULL,
    approver_id      BIGINT UNSIGNED DEFAULT 0 COMMENT '审批人ID',
    approver_name    VARCHAR(64) DEFAULT '',
    approve_comment  VARCHAR(500) DEFAULT '' COMMENT '审批意见',
    
    applied_at       BIGINT UNSIGNED DEFAULT 0 COMMENT '申请时间',
    approved_at      BIGINT UNSIGNED DEFAULT 0 COMMENT '审批时间',
    expired_at       BIGINT UNSIGNED DEFAULT 0 COMMENT '过期时间',
    
    INDEX idx_pipeline (pipeline_id),
    INDEX idx_status (status, applied_at),
    INDEX idx_applicant (applicant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD发布审批记录表';

-- =====================================================
-- 33. CICD - 资源配置变更日志表
-- =====================================================
CREATE TABLE IF NOT EXISTS cicd_resource_change_log (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    pipeline_id     BIGINT UNSIGNED NOT NULL,
    env             VARCHAR(16) NOT NULL,
    
    change_type     ENUM('create','update','scale','rollback') NOT NULL,
    before_config   TEXT COMMENT '变更前配置JSON',
    after_config    TEXT COMMENT '变更后配置JSON',
    
    operator_id     BIGINT UNSIGNED NOT NULL,
    operator_name   VARCHAR(64) NOT NULL,
    reason          VARCHAR(500) DEFAULT '' COMMENT '变更原因',
    
    created_at      BIGINT UNSIGNED DEFAULT 0,
    
    INDEX idx_pipeline_env (pipeline_id, env),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD资源配置变更日志表';

-- =====================================================
-- 34. CI/CD - 制品库表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_artifact` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint DEFAULT NULL COMMENT '关联流水线ID',
  `run_id` bigint DEFAULT NULL COMMENT '关联运行记录ID',
  `build_number` int DEFAULT 0 COMMENT 'Jenkins构建号',

  -- 制品基本信息
  `name` varchar(200) NOT NULL DEFAULT '' COMMENT '制品名称（如 order-service-1.0.0.jar）',
  `artifact_type` varchar(20) NOT NULL DEFAULT '' COMMENT '制品类型：jar/war/binary/dist/wheel/image/archive',
  `version` varchar(100) NOT NULL DEFAULT '' COMMENT '版本号',
  `language_type` varchar(20) NOT NULL DEFAULT '' COMMENT '语言类型：go/java/frontend/python',

  -- 存储信息
  `file_path` varchar(500) NOT NULL DEFAULT '' COMMENT '文件存储路径',
  `file_size` bigint NOT NULL DEFAULT 0 COMMENT '文件大小（字节）',
  `sha256` varchar(64) NOT NULL DEFAULT '' COMMENT 'SHA256校验和',
  `storage_type` varchar(20) NOT NULL DEFAULT 'local' COMMENT '存储类型：local/s3/oss',

  -- Git 信息（构建来源追溯）
  `git_repo` varchar(500) NOT NULL DEFAULT '' COMMENT 'Git仓库地址',
  `git_branch` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit` varchar(40) NOT NULL DEFAULT '' COMMENT 'Git Commit SHA',

  -- 镜像信息（如果制品已打包为镜像）
  `image_repo` varchar(500) NOT NULL DEFAULT '' COMMENT '镜像仓库地址',
  `image_tag` varchar(200) NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(100) NOT NULL DEFAULT '' COMMENT '镜像摘要',

  -- 构建元数据
  `build_duration` int NOT NULL DEFAULT 0 COMMENT '构建耗时（秒）',
  `build_log` text COMMENT '构建摘要日志',
  `metadata` json DEFAULT NULL COMMENT '扩展元数据',

  -- 状态
  `status` varchar(20) NOT NULL DEFAULT 'ready' COMMENT '状态：uploading/ready/expired/deleted',
  `download_count` int NOT NULL DEFAULT 0 COMMENT '下载次数',

  -- 元数据
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,

  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_run_id` (`run_id`),
  KEY `idx_artifact_type` (`artifact_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD 制品库表';

-- =====================================================
-- CICD 资源模板初始数据
-- =====================================================

-- Java 服务模板
INSERT IGNORE INTO cicd_resource_template 
(name, service_type, env, replicas_default, replicas_min, replicas_max, cpu_request, cpu_limit, memory_request, memory_limit, hpa_enabled, description, is_default, sort_order, created_at, modified_at)
VALUES
('small',  'java', 'dev',  1, 1, 2,  '200m', '500m', '512Mi', '1Gi',   0, 'Java开发环境-小型，适合本地调试', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'java', 'test', 1, 1, 3,  '500m', '1',    '1Gi',   '2Gi',   0, 'Java测试环境-小型，适合功能测试', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'java', 'prod', 2, 2, 5,  '500m', '1',    '1Gi',   '2Gi',   0, 'Java生产环境-小型，适合低流量服务', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'java', 'prod', 2, 2, 10, '1',    '2',    '2Gi',   '4Gi',   1, 'Java生产环境-中型，适合中等流量服务', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('large',  'java', 'prod', 3, 2, 20, '2',    '4',    '4Gi',   '8Gi',   1, 'Java生产环境-大型，适合高流量核心服务', 0, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- Go 服务模板
INSERT IGNORE INTO cicd_resource_template 
(name, service_type, env, replicas_default, replicas_min, replicas_max, cpu_request, cpu_limit, memory_request, memory_limit, hpa_enabled, description, is_default, sort_order, created_at, modified_at)
VALUES
('small',  'go', 'dev',  1, 1, 2,  '100m', '200m', '128Mi', '256Mi', 0, 'Go开发环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'go', 'test', 1, 1, 3,  '200m', '500m', '256Mi', '512Mi', 0, 'Go测试环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'go', 'prod', 2, 2, 5,  '200m', '500m', '256Mi', '512Mi', 0, 'Go生产环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'go', 'prod', 2, 2, 10, '500m', '1',    '512Mi', '1Gi',   1, 'Go生产环境-中型', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('large',  'go', 'prod', 3, 2, 20, '1',    '2',    '1Gi',   '2Gi',   1, 'Go生产环境-大型', 0, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- Node 服务模板
INSERT IGNORE INTO cicd_resource_template 
(name, service_type, env, replicas_default, replicas_min, replicas_max, cpu_request, cpu_limit, memory_request, memory_limit, hpa_enabled, description, is_default, sort_order, created_at, modified_at)
VALUES
('small',  'node', 'dev',  1, 1, 2,  '100m', '300m', '256Mi', '512Mi', 0, 'Node开发环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'node', 'test', 1, 1, 3,  '200m', '500m', '512Mi', '1Gi',   0, 'Node测试环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'node', 'prod', 2, 2, 5,  '200m', '500m', '512Mi', '1Gi',   0, 'Node生产环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'node', 'prod', 2, 2, 10, '500m', '1',    '1Gi',   '2Gi',   1, 'Node生产环境-中型', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- Python 服务模板
INSERT IGNORE INTO cicd_resource_template 
(name, service_type, env, replicas_default, replicas_min, replicas_max, cpu_request, cpu_limit, memory_request, memory_limit, hpa_enabled, description, is_default, sort_order, created_at, modified_at)
VALUES
('small',  'python', 'dev',  1, 1, 2,  '100m', '300m', '256Mi', '512Mi', 0, 'Python开发环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'python', 'test', 1, 1, 3,  '200m', '500m', '512Mi', '1Gi',   0, 'Python测试环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'python', 'prod', 2, 2, 5,  '200m', '500m', '512Mi', '1Gi',   0, 'Python生产环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'python', 'prod', 2, 2, 10, '500m', '1',    '1Gi',   '2Gi',   1, 'Python生产环境-中型', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 环境资源规则
INSERT IGNORE INTO cicd_env_resource_rule 
(env, service_type, cpu_limit_max, memory_limit_max, replicas_max, cpu_request_min, memory_request_min, replicas_min, require_approval, approval_role, description, created_at, modified_at)
VALUES
('dev', '', '1', '2Gi', 3, '', '', 1, 0, '', '开发环境通用规则，资源受限，无需审批', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('test', '', '2', '4Gi', 5, '', '', 1, 0, '', '测试环境通用规则，资源适中，无需审批', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('staging', '', '4', '8Gi', 10, '200m', '256Mi', 2, 0, '', '预发环境通用规则，接近生产配置', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('prod', '', '4', '8Gi', 20, '200m', '256Mi', 2, 1, 'sre', '生产环境通用规则，需要SRE审批', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('prod', 'java', '4', '8Gi', 20, '500m', '1Gi', 2, 1, 'sre', '生产环境Java服务规则，内存最低1Gi', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 35. CI/CD - 构建探针表（Build Agent）
-- 管理 OTEL Java Agent、SkyWalking、Arthas 等构建探针
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_build_agent` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '探针名称',
  `display_name` varchar(200) DEFAULT '' COMMENT '显示名称',
  `description` varchar(1000) DEFAULT '' COMMENT '描述说明',
  `category` varchar(30) DEFAULT '' COMMENT '分类: observability/diagnostics/security/custom',
  `scope` varchar(20) DEFAULT '' COMMENT '适用语言: java/go/python/all',
  `version` varchar(50) DEFAULT '' COMMENT '版本号',
  `file_name` varchar(300) DEFAULT '' COMMENT '文件名',
  `file_path` varchar(500) DEFAULT '' COMMENT '存储路径',
  `file_size` bigint DEFAULT 0 COMMENT '文件大小(字节)',
  `sha256` varchar(64) DEFAULT '' COMMENT 'SHA256校验和',
  `download_url` varchar(500) DEFAULT '' COMMENT '官方下载地址',
  `doc_url` varchar(500) DEFAULT '' COMMENT '文档地址',
  `icon` varchar(50) DEFAULT '' COMMENT '图标',
  `docker_copy_dest` varchar(200) DEFAULT '' COMMENT '镜像内目标路径',
  `env_key` varchar(100) DEFAULT '' COMMENT '注入的环境变量名',
  `env_value` varchar(2000) DEFAULT '' COMMENT '环境变量默认值模板',
  `status` varchar(20) NOT NULL DEFAULT 'active' COMMENT '状态: active/inactive',
  `download_count` int NOT NULL DEFAULT 0 COMMENT '下载次数',
  `used_count` int NOT NULL DEFAULT 0 COMMENT '被引用次数',
  `created_user_id` bigint DEFAULT 0,
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  KEY `idx_category` (`category`),
  KEY `idx_scope` (`scope`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD 构建探针表';

-- =====================================================
-- 36. 应用商城 - 应用表
-- =====================================================
CREATE TABLE IF NOT EXISTS `app_store_apps` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '应用名称',
  `display_name` varchar(256) DEFAULT '' COMMENT '显示名称',
  `category` varchar(64) NOT NULL COMMENT '分类',
  `version` varchar(64) NOT NULL COMMENT '版本号',
  `icon` varchar(512) DEFAULT '' COMMENT '图标URL或icon名',
  `description` varchar(1024) DEFAULT '' COMMENT '应用描述',
  `provider` varchar(128) DEFAULT '' COMMENT '提供方',
  `chart_url` varchar(512) DEFAULT '' COMMENT 'Helm Chart地址',
  `doc_url` varchar(512) DEFAULT '' COMMENT '文档地址',
  `status` tinyint unsigned DEFAULT 1 COMMENT '1可用 2维护 3下架',
  `featured` tinyint unsigned DEFAULT 0 COMMENT '是否推荐',
  `sort_order` int DEFAULT 0 COMMENT '排序权重',
  `tags` varchar(512) DEFAULT '' COMMENT '标签,逗号分隔',
  `min_k8s` varchar(32) DEFAULT '' COMMENT '最低K8s版本',
  `namespace` varchar(128) DEFAULT '' COMMENT '默认安装命名空间',
  `values_yaml` text COMMENT '默认values.yaml',
  `created_at` int unsigned DEFAULT 0,
  `modified_at` int unsigned DEFAULT 0,
  `deleted_at` int unsigned DEFAULT 0,
  `is_del` tinyint unsigned DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  KEY `idx_category` (`category`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用商城应用表';

-- =====================================================
-- 37. 应用商城 - 组件表
-- =====================================================
CREATE TABLE IF NOT EXISTS `app_store_components` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int unsigned NOT NULL COMMENT '关联应用ID',
  `name` varchar(128) NOT NULL COMMENT '组件名称',
  `image` varchar(512) NOT NULL COMMENT '容器镜像',
  `replicas` int DEFAULT 1 COMMENT '副本数',
  `ports` varchar(512) DEFAULT '' COMMENT '端口定义JSON',
  `args` varchar(1024) DEFAULT '' COMMENT '启动参数JSON',
  `cpu_req` varchar(32) DEFAULT '50m' COMMENT 'CPU Request',
  `cpu_lim` varchar(32) DEFAULT '200m' COMMENT 'CPU Limit',
  `mem_req` varchar(32) DEFAULT '64Mi' COMMENT 'Memory Request',
  `mem_lim` varchar(32) DEFAULT '256Mi' COMMENT 'Memory Limit',
  `sort_order` int DEFAULT 0 COMMENT '排序(越大越靠前)',
  `created_at` int unsigned DEFAULT 0,
  `modified_at` int unsigned DEFAULT 0,
  `deleted_at` int unsigned DEFAULT 0,
  `is_del` tinyint unsigned DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_app_id` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用商城组件表';

-- =====================================================
-- 38. 应用商城 - 安装记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `app_store_installs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int unsigned NOT NULL COMMENT '应用ID',
  `app_name` varchar(128) NOT NULL COMMENT '应用名称(冗余)',
  `cluster_id` int unsigned NOT NULL COMMENT '集群ID',
  `cluster_name` varchar(128) DEFAULT '' COMMENT '集群名称(冗余)',
  `namespace` varchar(128) NOT NULL COMMENT '安装命名空间',
  `release_name` varchar(128) NOT NULL COMMENT 'Release名称',
  `version` varchar(64) DEFAULT '' COMMENT '安装版本',
  `values` text COMMENT '自定义values',
  `status` tinyint unsigned DEFAULT 1 COMMENT '1安装中 2已安装 3失败 4卸载中 5已卸载 6部分就绪',
  `message` varchar(1024) DEFAULT '' COMMENT '状态消息',
  `operator` varchar(64) DEFAULT '' COMMENT '操作人',
  `created_at` int unsigned DEFAULT 0,
  `modified_at` int unsigned DEFAULT 0,
  `deleted_at` int unsigned DEFAULT 0,
  `is_del` tinyint unsigned DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用商城安装记录表';

-- =====================================================
-- 39. AI 助手 - 会话表
-- =====================================================
CREATE TABLE IF NOT EXISTS `ai_conversations` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned NOT NULL COMMENT '关联用户ID',
  `title` varchar(200) NOT NULL DEFAULT '新对话' COMMENT '会话标题',
  `status` tinyint unsigned NOT NULL DEFAULT 1 COMMENT '1=活跃 2=归档',
  `created_at` int unsigned NOT NULL DEFAULT 0,
  `modified_at` int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 助手会话表';

-- =====================================================
-- 40. AI 助手 - 聊天消息表
-- =====================================================
CREATE TABLE IF NOT EXISTS `ai_messages` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `conversation_id` int unsigned NOT NULL COMMENT '关联会话ID',
  `role` varchar(20) NOT NULL COMMENT 'system/user/assistant/tool',
  `content` text COMMENT '消息内容',
  `intent_json` text COMMENT '意图识别结果JSON',
  `token_used` int NOT NULL DEFAULT 0 COMMENT 'Token消耗',
  `created_at` int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_conversation_id` (`conversation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 助手消息表';

-- =====================================================
-- 41. AI 助手 - 高危操作审批请求表
-- =====================================================
CREATE TABLE IF NOT EXISTS `ai_approval_requests` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `conversation_id` int unsigned NOT NULL DEFAULT 0 COMMENT '关联AI会话ID',
  `request_user_id` int unsigned NOT NULL COMMENT '发起人ID',
  `approver_user_id` int unsigned NOT NULL DEFAULT 0 COMMENT '审批人ID',
  `intent` varchar(50) NOT NULL COMMENT '操作意图: delete/drain/scale',
  `resource` varchar(100) NOT NULL DEFAULT '' COMMENT '资源类型',
  `resource_name` varchar(200) NOT NULL DEFAULT '' COMMENT '资源名称',
  `namespace` varchar(100) NOT NULL DEFAULT '' COMMENT '命名空间',
  `cluster_id` int unsigned NOT NULL DEFAULT 0 COMMENT '目标集群ID',
  `risk_level` varchar(20) NOT NULL DEFAULT 'medium' COMMENT '风险等级',
  `operation_json` text COMMENT '完整操作参数JSON',
  `tool_name` varchar(100) NOT NULL DEFAULT '' COMMENT 'Function Calling工具名',
  `tool_args_json` text COMMENT '工具调用参数JSON',
  `tool_call_id` varchar(100) NOT NULL DEFAULT '' COMMENT 'OpenAI tool_call_id',
  `execute_result` text COMMENT '执行结果',
  `executed` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已执行',
  `summary` varchar(500) NOT NULL DEFAULT '' COMMENT '操作摘要(AI生成)',
  `status` tinyint unsigned NOT NULL DEFAULT 1 COMMENT '1=待审批 2=已通过 3=已拒绝 4=已过期 5=已取消',
  `approve_comment` varchar(500) NOT NULL DEFAULT '' COMMENT '审批备注',
  `expire_at` int unsigned NOT NULL DEFAULT 0 COMMENT '过期时间戳',
  `created_at` int unsigned NOT NULL DEFAULT 0,
  `modified_at` int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_request_user_id` (`request_user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_conversation_id` (`conversation_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 高危操作审批请求表';

-- =====================================================
-- 42. AI 助手 - 审批操作日志表
-- =====================================================
CREATE TABLE IF NOT EXISTS `ai_approval_logs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `approval_id` int unsigned NOT NULL COMMENT '关联审批请求ID',
  `user_id` int unsigned NOT NULL COMMENT '操作人ID',
  `action` varchar(50) NOT NULL COMMENT 'create/approve/reject/cancel/expire',
  `comment` varchar(500) NOT NULL DEFAULT '' COMMENT '操作说明',
  `created_at` int unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_approval_id` (`approval_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 审批操作日志表';

-- =====================================================
-- 43. 监控 - 数据源表
-- =====================================================
CREATE TABLE IF NOT EXISTS `monitor_datasource` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '数据源名称',
  `type` varchar(30) NOT NULL COMMENT 'prometheus/loki/alertmanager/grafana/victoriametrics',
  `url` varchar(500) NOT NULL COMMENT '连接地址',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `cluster_id` bigint NOT NULL DEFAULT 0 COMMENT '关联 K8s 集群 ID（0=全局/未关联）',
  `access_mode` varchar(20) DEFAULT 'proxy' COMMENT 'proxy/direct',
  `auth_type` varchar(20) DEFAULT 'none' COMMENT 'none/basic/bearer/tls',
  `auth_user` varchar(100) DEFAULT '',
  `auth_pass` varchar(500) DEFAULT '',
  `tls_cert` text,
  `tls_key` text,
  `ca_cert` text,
  `is_default` tinyint(1) DEFAULT 0,
  `enabled` tinyint(1) DEFAULT 1,
  `timeout` int DEFAULT 30,
  `scrape_interval` int DEFAULT 15,
  `status` varchar(20) DEFAULT 'unknown' COMMENT 'connected/disconnected/unknown',
  `last_check_at` bigint DEFAULT 0,
  `created_by` bigint DEFAULT 0,
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_cluster_id` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控数据源表';

-- 【兼容存量集群】monitor_datasource 幂等补丁：若老库已存在该表但缺 `cluster_id`字段/索引，自动补上
SET @col_exists := (
    SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'monitor_datasource'
      AND column_name = 'cluster_id'
);
SET @sql := IF(@col_exists = 0,
    'ALTER TABLE `monitor_datasource` ADD COLUMN `cluster_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''关联 K8s 集群 ID（0=全局/未关联）'' AFTER `description`',
    'SELECT ''column cluster_id already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx_exists := (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'monitor_datasource'
      AND index_name = 'idx_cluster_id'
);
SET @sql := IF(@idx_exists = 0,
    'ALTER TABLE `monitor_datasource` ADD INDEX `idx_cluster_id` (`cluster_id`)',
    'SELECT ''index idx_cluster_id already exists, skip'' AS msg'
);
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- =====================================================
-- 44. 监控 - 告警规则表
-- =====================================================
CREATE TABLE IF NOT EXISTS `monitor_alert_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `datasource_id` bigint NOT NULL COMMENT '关联数据源',
  `name` varchar(200) NOT NULL COMMENT '规则名称',
  `group` varchar(100) DEFAULT 'default' COMMENT '规则分组',
  `severity` varchar(20) NOT NULL COMMENT 'critical/warning/info',
  `expr` text NOT NULL COMMENT 'PromQL 表达式',
  `duration` varchar(20) DEFAULT '5m' COMMENT '持续时间(for)',
  `summary` varchar(500) DEFAULT '' COMMENT '告警摘要模板',
  `description` text COMMENT '告警描述模板',
  `labels` text COMMENT '额外标签JSON',
  `annotations` text COMMENT '额外注解JSON',
  `enabled` tinyint(1) DEFAULT 1,
  `notify_channels` varchar(500) DEFAULT '' COMMENT '通知渠道',
  `notify_url` varchar(500) DEFAULT '' COMMENT 'webhook URL',
  `eval_interval` int DEFAULT 60 COMMENT '评估间隔(秒)',
  `last_eval_at` bigint DEFAULT 0,
  `last_eval_result` varchar(20) DEFAULT '' COMMENT 'normal/firing/pending/error',
  `pending_since` bigint DEFAULT 0 COMMENT '告警条件首次满足时间',
  `created_by` bigint DEFAULT 0,
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_datasource_id` (`datasource_id`),
  KEY `idx_severity` (`severity`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警规则表';

-- =====================================================
-- 45. 监控 - 告警事件表
-- =====================================================
CREATE TABLE IF NOT EXISTS `monitor_alert_event` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `rule_id` bigint NOT NULL COMMENT '关联告警规则',
  `datasource_id` bigint NOT NULL COMMENT '关联数据源',
  `rule_name` varchar(200) DEFAULT '' COMMENT '冗余规则名',
  `severity` varchar(20) NOT NULL COMMENT '严重级别',
  `status` varchar(20) NOT NULL COMMENT 'firing/resolved/silenced',
  `value` varchar(100) DEFAULT '' COMMENT '触发时的值',
  `labels` text COMMENT '标签JSON',
  `annotations` text COMMENT '注解JSON',
  `summary` varchar(500) DEFAULT '',
  `description` text,
  `fired_at` bigint NOT NULL COMMENT '触发时间',
  `resolved_at` bigint DEFAULT 0 COMMENT '恢复时间',
  `acked_by` bigint DEFAULT 0 COMMENT '确认人',
  `acked_at` bigint DEFAULT 0 COMMENT '确认时间',
  `silenced_until` bigint DEFAULT 0 COMMENT '静默截止',
  `notify_result` varchar(200) DEFAULT '' COMMENT '通知结果',
  `created_at` bigint DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_datasource_id` (`datasource_id`),
  KEY `idx_severity` (`severity`),
  KEY `idx_status` (`status`),
  KEY `idx_fired_at` (`fired_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警事件表';

-- =====================================================
-- 46. 监控 - 通知渠道表
-- =====================================================
CREATE TABLE IF NOT EXISTS `monitor_notify_channel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '渠道名称',
  `type` varchar(30) NOT NULL COMMENT 'dingtalk/feishu/webhook/email/wechat',
  `description` varchar(500) DEFAULT '',
  `webhook_url` varchar(500) DEFAULT '',
  `secret` varchar(500) DEFAULT '' COMMENT '签名密钥',
  `security_keyword` varchar(100) DEFAULT '' COMMENT '钉钉安全关键字（多个用逗号分隔）',
  `at_mobiles` varchar(500) DEFAULT '' COMMENT '@手机号列表',
  `at_all` tinyint(1) DEFAULT 0,
  `smtp_host` varchar(200) DEFAULT '',
  `smtp_port` int DEFAULT 465,
  `smtp_user` varchar(200) DEFAULT '',
  `smtp_pass` varchar(200) DEFAULT '',
  `smtp_to` text COMMENT '收件人列表',
  `msg_template` text COMMENT '消息模板',
  `enabled` tinyint(1) DEFAULT 1,
  `send_resolved` tinyint(1) DEFAULT 1,
  `rate_limit` int DEFAULT 10 COMMENT '限流',
  `created_by` bigint DEFAULT 0,
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控通知渠道表';

-- =====================================================
-- 47. 监控 - 告警静默规则表
-- =====================================================
CREATE TABLE IF NOT EXISTS `monitor_silence_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(200) NOT NULL,
  `type` varchar(30) NOT NULL COMMENT 'silence/inhibit/aggregate',
  `matchers` text NOT NULL COMMENT '匹配条件JSON',
  `starts_at` bigint DEFAULT 0,
  `ends_at` bigint DEFAULT 0,
  `duration` varchar(30) DEFAULT '',
  `repeat_type` varchar(20) DEFAULT 'once' COMMENT 'once/daily/weekly/cron',
  `repeat_cron` varchar(100) DEFAULT '',
  `comment` varchar(500) DEFAULT '',
  `enabled` tinyint(1) DEFAULT 1,
  `created_by` bigint DEFAULT 0,
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警静默规则表';

-- =====================================================
-- 48. 监控 - 告警抑制规则表
-- =====================================================
CREATE TABLE IF NOT EXISTS `monitor_inhibit_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(200) NOT NULL,
  `source_matchers` text NOT NULL COMMENT '源告警匹配JSON',
  `target_matchers` text NOT NULL COMMENT '目标告警匹配JSON',
  `equal_labels` varchar(500) DEFAULT '' COMMENT '关联标签',
  `description` varchar(500) DEFAULT '',
  `enabled` tinyint(1) DEFAULT 1,
  `created_by` bigint DEFAULT 0,
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警抑制规则表';

-- =====================================================
-- 49. 监控 - 告警聚合规则表
-- =====================================================
CREATE TABLE IF NOT EXISTS `monitor_aggregate_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(200) NOT NULL,
  `group_by` varchar(500) NOT NULL COMMENT '聚合维度',
  `group_wait` varchar(20) DEFAULT '30s',
  `group_interval` varchar(20) DEFAULT '5m',
  `repeat_interval` varchar(20) DEFAULT '4h',
  `matchers` text COMMENT '匹配条件JSON',
  `channel_ids` varchar(200) DEFAULT '' COMMENT '通知渠道ID列表',
  `enabled` tinyint(1) DEFAULT 1,
  `created_by` bigint DEFAULT 0,
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警聚合规则表';

-- =====================================================
-- 50. 监控 - 通知模板表
-- =====================================================
CREATE TABLE IF NOT EXISTS `monitor_notify_template` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL,
  `type` varchar(30) NOT NULL COMMENT 'dingtalk/feishu/wechat/email/webhook',
  `scene` varchar(30) NOT NULL DEFAULT 'alert' COMMENT 'alert/resolved/test',
  `title` varchar(200) DEFAULT '',
  `content` text NOT NULL COMMENT '模板内容',
  `description` varchar(500) DEFAULT '',
  `is_default` tinyint(1) DEFAULT 0,
  `enabled` tinyint(1) DEFAULT 1,
  `created_by` bigint DEFAULT 0,
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控通知模板表';

-- =====================================================
-- 51. 监控 - 通知路由策略表
-- =====================================================
CREATE TABLE IF NOT EXISTS `monitor_notify_route_policy` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(200) NOT NULL COMMENT '策略名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `priority` int DEFAULT 100 COMMENT '优先级(越小越优先,0为最高)',
  `channel_ids` varchar(500) NOT NULL DEFAULT '' COMMENT '目标通知渠道ID(逗号分隔)',
  `match_mode` varchar(20) DEFAULT 'any' COMMENT '匹配模式: all(全部满足)/any(任一满足)',
  `severities` varchar(100) DEFAULT '' COMMENT '匹配级别(逗号分隔): critical,warning,info',
  `groups` varchar(500) DEFAULT '' COMMENT '匹配规则分组(逗号分隔)',
  `label_match` text COMMENT '标签匹配条件JSON [{key,op,value}]',
  `is_default` tinyint(1) DEFAULT 0 COMMENT '是否为兜底默认策略',
  `enabled` tinyint(1) DEFAULT 1,
  `created_by` bigint DEFAULT 0,
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  KEY `idx_priority` (`priority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控通知路由策略表';

-- =====================================================
-- 52. AIOps - AI 分析记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `aiops_analysis_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `type` varchar(30) NOT NULL COMMENT '类型: alert_analysis/log_diagnosis/inspection',
  `ref_id` bigint DEFAULT 0 COMMENT '关联ID(告警事件ID/巡检报告ID)',
  `title` varchar(300) NOT NULL COMMENT '分析标题',
  `input` text COMMENT '输入数据(告警详情/日志片段/巡检摘要)',
  `result` longtext COMMENT 'AI分析结果(Markdown)',
  `severity` varchar(20) DEFAULT '' COMMENT 'AI判定严重级别',
  `suggestions` text COMMENT 'AI建议JSON',
  `model` varchar(100) DEFAULT '' COMMENT '使用的AI模型',
  `tokens_used` int DEFAULT 0 COMMENT '消耗Token数',
  `latency_ms` bigint DEFAULT 0 COMMENT '分析耗时(ms)',
  `status` varchar(20) DEFAULT 'success' COMMENT '状态: success/failed/timeout',
  `error` varchar(500) DEFAULT '' COMMENT '错误信息',
  `user_id` bigint DEFAULT 0 COMMENT '发起人',
  `created_at` bigint DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_ref_id` (`ref_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AIOps AI分析记录表';

-- =====================================================
-- 53. AIOps - 巡检报告表
-- =====================================================
CREATE TABLE IF NOT EXISTS `aiops_inspection_report` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `type` varchar(30) NOT NULL COMMENT '巡检类型: scheduled/manual',
  `scope` varchar(30) DEFAULT 'full' COMMENT '巡检范围: full/cluster/namespace',
  `scope_id` varchar(100) DEFAULT '' COMMENT '范围ID(集群ID等)',
  `health_score` int DEFAULT 0 COMMENT '健康评分 0-100',
  `level` varchar(20) NOT NULL COMMENT '健康等级: healthy/warning/critical',
  `summary` text COMMENT '巡检摘要(AI生成)',
  `details` longtext COMMENT '巡检详情JSON',
  `ai_analysis` longtext COMMENT 'AI综合分析(Markdown)',
  `findings` int DEFAULT 0 COMMENT '发现问题数',
  `suggestions_count` int DEFAULT 0 COMMENT '建议数',
  `duration` bigint DEFAULT 0 COMMENT '巡检耗时(ms)',
  `status` varchar(20) DEFAULT 'running' COMMENT '状态: running/completed/failed',
  `error` varchar(500) DEFAULT '' COMMENT '错误信息',
  `triggered_by` bigint DEFAULT 0 COMMENT '触发人(0=系统定时)',
  `created_at` bigint DEFAULT 0,
  `completed_at` bigint DEFAULT 0 COMMENT '完成时间',
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AIOps巡检报告表';

-- 【兼容存量集群】monitor_alert_rule 幂等补丁：若老库缺 pending_since 字段，自动补上
SET @col_exists := (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'monitor_alert_rule' AND column_name = 'pending_since');
SET @sql := IF(@col_exists = 0, 'ALTER TABLE `monitor_alert_rule` ADD COLUMN `pending_since` BIGINT DEFAULT 0 COMMENT ''告警条件首次满足时间'' AFTER `last_eval_result`', 'SELECT ''pending_since exists''');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- =====================================================
-- 构建探针种子数据（OpenTelemetry Java Agent）
-- =====================================================
INSERT IGNORE INTO `cicd_build_agent`
(name, display_name, description, category, scope, version, file_name, file_path, file_size, download_url, doc_url, icon, docker_copy_dest, env_key, env_value, status, created_at, modified_at)
VALUES
('opentelemetry-javaagent', 'OpenTelemetry Java Agent',
 'OpenTelemetry 官方 Java 自动埋点 Agent，通过 -javaagent 注入，自动采集 Trace/Metrics 并通过 OTLP 协议上报到 Collector。',
 'observability', 'java', '1.33.0',
 'opentelemetry-javaagent.jar', './storage/agents/observability/opentelemetry-javaagent/opentelemetry-javaagent.jar',
 0,
 'https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases',
 'https://opentelemetry.io/docs/instrumentation/java/automatic/',
 'OTEL',
 '/app/opentelemetry-javaagent.jar',
 'OTEL_OPTS',
 '-javaagent:/app/opentelemetry-javaagent.jar -Dotel.service.name=${SERVICE_NAME} -Dotel.traces.exporter=otlp -Dotel.metrics.exporter=none -Dotel.logs.exporter=none -Dotel.exporter.otlp.endpoint=http://otel-collector-monitoring.svc.cluster.local:4318',
 'active', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- CICD 资源档位模板（必需字典数据）
-- 唯一键 (service_type, env, name, deleted_at) 保证幂等
-- =====================================================
INSERT IGNORE INTO cicd_resource_template
(name, service_type, env, replicas_default, replicas_min, replicas_max, cpu_request, cpu_limit, memory_request, memory_limit, hpa_enabled, description, is_default, sort_order, created_at, modified_at)
VALUES
-- Java 服务模板
('small',  'java', 'dev',  1, 1, 2,  '200m', '500m', '512Mi', '1Gi',   0, 'Java开发环境-小型，适合本地调试',     1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'java', 'test', 1, 1, 3,  '500m', '1',    '1Gi',   '2Gi',   0, 'Java测试环境-小型，适合功能测试',     1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'java', 'prod', 2, 2, 5,  '500m', '1',    '1Gi',   '2Gi',   0, 'Java生产环境-小型，适合低流量服务',   1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'java', 'prod', 2, 2, 10, '1',    '2',    '2Gi',   '4Gi',   1, 'Java生产环境-中型，适合中等流量服务', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('large',  'java', 'prod', 3, 2, 20, '2',    '4',    '4Gi',   '8Gi',   1, 'Java生产环境-大型，适合高流量核心',   0, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- Go 服务模板
('small',  'go', 'dev',  1, 1, 2,  '100m', '200m', '128Mi', '256Mi', 0, 'Go开发环境-小型',  1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'go', 'test', 1, 1, 3,  '200m', '500m', '256Mi', '512Mi', 0, 'Go测试环境-小型',  1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'go', 'prod', 2, 2, 5,  '200m', '500m', '256Mi', '512Mi', 0, 'Go生产环境-小型',  1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'go', 'prod', 2, 2, 10, '500m', '1',    '512Mi', '1Gi',   1, 'Go生产环境-中型',  0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('large',  'go', 'prod', 3, 2, 20, '1',    '2',    '1Gi',   '2Gi',   1, 'Go生产环境-大型',  0, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- Node 服务模板
('small',  'node', 'dev',  1, 1, 2,  '100m', '300m', '256Mi', '512Mi', 0, 'Node开发环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'node', 'test', 1, 1, 3,  '200m', '500m', '512Mi', '1Gi',   0, 'Node测试环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'node', 'prod', 2, 2, 5,  '200m', '500m', '512Mi', '1Gi',   0, 'Node生产环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'node', 'prod', 2, 2, 10, '500m', '1',    '1Gi',   '2Gi',   1, 'Node生产环境-中型', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- Python 服务模板
('small',  'python', 'dev',  1, 1, 2,  '100m', '300m', '256Mi', '512Mi', 0, 'Python开发环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'python', 'test', 1, 1, 3,  '200m', '500m', '512Mi', '1Gi',   0, 'Python测试环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'python', 'prod', 2, 2, 5,  '200m', '500m', '512Mi', '1Gi',   0, 'Python生产环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'python', 'prod', 2, 2, 10, '500m', '1',    '1Gi',   '2Gi',   1, 'Python生产环境-中型', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 环境资源规则
INSERT IGNORE INTO cicd_env_resource_rule
(env, service_type, cpu_limit_max, memory_limit_max, replicas_max, cpu_request_min, memory_request_min, replicas_min, require_approval, approval_role, description, created_at, modified_at)
VALUES
('dev',     '',     '1', '2Gi',  3,  '',     '',     1, 0, '',    '开发环境通用规则，资源受限，无需审批',   UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('test',    '',     '2', '4Gi',  5,  '',     '',     1, 0, '',    '测试环境通用规则，资源适中，无需审批',   UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('staging', '',     '4', '8Gi',  10, '200m', '256Mi', 2, 0, '',    '预发环境通用规则，接近生产配置',         UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('prod',    '',     '4', '8Gi',  20, '200m', '256Mi', 2, 1, 'sre', '生产环境通用规则，需要SRE审批',          UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('prod',    'java', '4', '8Gi',  20, '500m', '1Gi',   2, 1, 'sre', '生产环境Java服务规则，内存最低1Gi',      UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 演示种子数据（可选）
-- 包含：4 流水线 + 6 运行 + 11 阶段 + 4 制品 + 2 监控数据源
-- 全部使用 INSERT IGNORE + 显式主键，重复执行幂等
-- 依赖：kube_cluster.id=1, image_registry.id=1, user.id=1
-- =====================================================
SET @now := UNIX_TIMESTAMP();

-- 流水线
INSERT IGNORE INTO cicd_pipeline
(id, name, description, git_repo, git_branch, jenkins_url, jenkins_job, jenkins_credential_id,
 language_type, auto_deploy, target_cluster_id, target_namespace, target_workload_kind, target_workload_name,
 target_container, deploy_env, require_approval, enable_sonar, enable_artifact_upload,
 last_deploy_image, last_deploy_time, last_deploy_status, last_deploy_version,
 status, last_run_status, last_run_time, last_build_number, last_build_url,
 created_user_id, created_at, modified_at)
VALUES
(1, 'order-service-go', '订单服务 Go 后端流水线', 'https://github.com/demo/order-service.git', 'main',
 'http://jenkins.local:8080', 'order-service-go', 'jenkins-cred-1',
 'go', 1, 1, 'demo', 'Deployment', 'order-service', 'order-service', 'dev', 0, 1, 1,
 'registry.local/demo/order-service:v1.2.3', @now-3600, 'success', 'v1.2.3',
 'idle', 'success', @now-3600, 12, 'http://jenkins.local:8080/job/order-service-go/12/',
 1, @now-86400*7, @now-3600),
(2, 'user-center-java', '用户中心 Java Spring Boot 流水线', 'https://github.com/demo/user-center.git', 'release/v2',
 'http://jenkins.local:8080', 'user-center-java', 'jenkins-cred-1',
 'java', 1, 1, 'demo', 'Deployment', 'user-center', 'user-center', 'staging', 1, 1, 1,
 'registry.local/demo/user-center:2.0.5', @now-7200, 'success', '2.0.5',
 'idle', 'success', @now-7200, 8, 'http://jenkins.local:8080/job/user-center-java/8/',
 1, @now-86400*5, @now-7200),
(3, 'admin-portal-frontend', '管理后台前端 Vue3 流水线', 'https://github.com/demo/admin-portal.git', 'develop',
 'http://jenkins.local:8080', 'admin-portal-frontend', 'jenkins-cred-1',
 'frontend', 0, 1, 'demo', 'Deployment', 'admin-portal', 'web', 'dev', 0, 0, 1,
 'registry.local/demo/admin-portal:dev-202605181030', @now-1800, 'failed', 'dev-202605181030',
 'idle', 'failed', @now-1800, 5, 'http://jenkins.local:8080/job/admin-portal-frontend/5/',
 1, @now-86400*3, @now-1800),
(4, 'data-pipeline-python', '数据处理 Python 流水线', 'https://github.com/demo/data-pipeline.git', 'main',
 'http://jenkins.local:8080', 'data-pipeline-python', 'jenkins-cred-1',
 'python', 0, 1, 'demo', 'CronJob', 'data-pipeline', 'worker', 'prod', 1, 1, 1,
 'registry.local/demo/data-pipeline:1.0.0', @now-86400, 'success', '1.0.0',
 'idle', 'success', @now-86400, 3, 'http://jenkins.local:8080/job/data-pipeline-python/3/',
 1, @now-86400*10, @now-86400);

-- 流水线运行记录
INSERT IGNORE INTO cicd_pipeline_run
(id, pipeline_id, build_number, status, trigger_type, trigger_user_id,
 git_commit, git_branch, git_commit_message, jenkins_build_url,
 duration_sec, started_at, finished_at, created_at, modified_at,
 image_url, image_digest, callback_received)
VALUES
(1, 1, 12, 'success',  'manual',    1, 'a1b2c3d4e5f6', 'main',       'feat: support batch order create', 'http://jenkins.local:8080/job/order-service-go/12/',     185, @now-3785,        @now-3600,        @now-3785,        @now-3600,        'registry.local/demo/order-service:v1.2.3', 'sha256:abc123def456', 1),
(2, 1, 11, 'success',  'webhook',   1, 'b2c3d4e5f6a1', 'main',       'fix: order amount precision',      'http://jenkins.local:8080/job/order-service-go/11/',     172, @now-86400-200,   @now-86400-28,    @now-86400-200,   @now-86400-28,    'registry.local/demo/order-service:v1.2.2', 'sha256:bbb222ccc333', 1),
(3, 2,  8, 'success',  'manual',    1, 'c3d4e5f6a1b2', 'release/v2', 'release: 2.0.5',                    'http://jenkins.local:8080/job/user-center-java/8/',     412, @now-7612,        @now-7200,        @now-7612,        @now-7200,        'registry.local/demo/user-center:2.0.5',    'sha256:ccc333ddd444', 1),
(4, 3,  5, 'failed',   'manual',    1, 'd4e5f6a1b2c3', 'develop',    'wip: refactor menu component',     'http://jenkins.local:8080/job/admin-portal-frontend/5/',  65, @now-1865,        @now-1800,        @now-1865,        @now-1800,        '',                                          '',                     1),
(5, 3,  4, 'success',  'webhook',   1, 'e5f6a1b2c3d4', 'develop',    'feat: add dark mode',              'http://jenkins.local:8080/job/admin-portal-frontend/4/', 128, @now-86400*2-200, @now-86400*2-72,  @now-86400*2-200, @now-86400*2-72,  'registry.local/demo/admin-portal:dev-202605161830', 'sha256:eee555fff666', 1),
(6, 4,  3, 'success',  'scheduled', 1, 'f6a1b2c3d4e5', 'main',       'release: 1.0.0',                    'http://jenkins.local:8080/job/data-pipeline-python/3/', 256, @now-86400-256,   @now-86400,       @now-86400-256,   @now-86400,       'registry.local/demo/data-pipeline:1.0.0',  'sha256:fff666aaa777', 1);

-- 流水线阶段（显式 id 保证幂等）
INSERT IGNORE INTO cicd_pipeline_stage
(id, run_id, pipeline_id, stage_order, stage_type, stage_name, status,
 started_at, finished_at, duration_sec, jenkins_stage_id, created_at, modified_at)
VALUES
(1,  1, 1, 1, 'checkout',     '代码拉取',      'success', @now-3785, @now-3780,  5, 'stage-1', @now-3785, @now-3780),
(2,  1, 1, 2, 'dependencies', '依赖安装',      'success', @now-3780, @now-3760, 20, 'stage-2', @now-3780, @now-3760),
(3,  1, 1, 3, 'compile',      '编译构建',      'success', @now-3760, @now-3720, 40, 'stage-3', @now-3760, @now-3720),
(4,  1, 1, 4, 'test',         '单元测试',      'success', @now-3720, @now-3690, 30, 'stage-4', @now-3720, @now-3690),
(5,  1, 1, 5, 'sonar',        'SonarQube扫描', 'success', @now-3690, @now-3660, 30, 'stage-5', @now-3690, @now-3660),
(6,  1, 1, 6, 'build',        '镜像构建',      'success', @now-3660, @now-3625, 35, 'stage-6', @now-3660, @now-3625),
(7,  1, 1, 7, 'push',         '镜像推送',      'success', @now-3625, @now-3610, 15, 'stage-7', @now-3625, @now-3610),
(8,  1, 1, 8, 'deploy',       'K8s部署',       'success', @now-3610, @now-3600, 10, 'stage-8', @now-3610, @now-3600),
(9,  4, 3, 1, 'checkout',     '代码拉取',      'success', @now-1865, @now-1860,  5, 'stage-1', @now-1865, @now-1860),
(10, 4, 3, 2, 'dependencies', 'npm install',   'success', @now-1860, @now-1820, 40, 'stage-2', @now-1860, @now-1820),
(11, 4, 3, 3, 'compile',      'vite build',    'failed',  @now-1820, @now-1800, 20, 'stage-3', @now-1820, @now-1800);

-- 制品库（显式 id 保证幂等）
INSERT IGNORE INTO cicd_artifact
(id, pipeline_id, run_id, build_number, name, artifact_type, version, language_type,
 file_path, file_size, sha256, storage_type,
 git_repo, git_branch, git_commit,
 image_repo, image_tag, image_digest,
 build_duration, status, download_count, created_user_id, created_at, modified_at)
VALUES
(1, 1, 1, 12, 'order-service-v1.2.3',           'binary', 'v1.2.3',           'go',       'storage/artifacts/1/12/order-service',                          18874368, 'a1b2c3d4e5f60011223344556677889900aabbccddeeff0011223344', 'local', 'https://github.com/demo/order-service.git',  'main',       'a1b2c3d4e5f6', 'registry.local/demo/order-service',  'v1.2.3',           'sha256:abc123def456', 185, 'ready', 3, 1, @now-3600,         @now-3600),
(2, 2, 3,  8, 'user-center-2.0.5.jar',          'jar',    '2.0.5',            'java',     'storage/artifacts/2/8/user-center-2.0.5.jar',                  41943040, 'c3d4e5f6a1b20011223344556677889900aabbccddeeff0011223344', 'local', 'https://github.com/demo/user-center.git',    'release/v2', 'c3d4e5f6a1b2', 'registry.local/demo/user-center',    '2.0.5',            'sha256:ccc333ddd444', 412, 'ready', 1, 1, @now-7200,         @now-7200),
(3, 3, 5,  4, 'admin-portal-dist-202605161830', 'dist',   'dev-202605161830', 'frontend', 'storage/artifacts/3/4/dist.tar.gz',                              8388608, 'e5f6a1b2c3d40011223344556677889900aabbccddeeff0011223344', 'local', 'https://github.com/demo/admin-portal.git',   'develop',    'e5f6a1b2c3d4', 'registry.local/demo/admin-portal',   'dev-202605161830', 'sha256:eee555fff666', 128, 'ready', 0, 1, @now-86400*2-72,   @now-86400*2-72),
(4, 4, 6,  3, 'data-pipeline-1.0.0.whl',        'wheel',  '1.0.0',            'python',   'storage/artifacts/4/3/data_pipeline-1.0.0-py3-none-any.whl',     2097152, 'f6a1b2c3d4e50011223344556677889900aabbccddeeff0011223344', 'local', 'https://github.com/demo/data-pipeline.git',  'main',       'f6a1b2c3d4e5', 'registry.local/demo/data-pipeline',  '1.0.0',            'sha256:fff666aaa777', 256, 'ready', 2, 1, @now-86400,        @now-86400);

-- 监控数据源（name 唯一键保证幂等）
INSERT IGNORE INTO monitor_datasource
(name, type, url, description, access_mode, auth_type,
 is_default, enabled, timeout, scrape_interval, status,
 created_by, created_at, modified_at)
VALUES
('Prometheus-默认', 'prometheus', 'http://prometheus.monitoring.svc:9090', '集群默认 Prometheus 数据源',  'proxy', 'none', 1, 1, 30, 15, 'unknown', 1, @now-86400*30, @now-86400*30),
('Loki-默认',       'loki',       'http://loki.monitoring.svc:3100',      '集群默认 Loki 日志数据源',    'proxy', 'none', 0, 1, 30, 15, 'unknown', 1, @now-86400*30, @now-86400*30);

-- =====================================================
-- 【兼容存量】cicd_pipeline 唯一索引迁移：单列 idx_name → 复合 idx_pipeline_name_del
-- 目的：软删除后允许重建同名流水线（活跃记录 deleted_at=0 保证唯一）
-- =====================================================
SET @idx_old_exists := (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_pipeline'
      AND index_name = 'idx_name'
);
SET @sql := IF(@idx_old_exists > 0,
    'ALTER TABLE `cicd_pipeline` DROP INDEX `idx_name`, ADD UNIQUE KEY `idx_pipeline_name_del` (`name`, `deleted_at`)',
    'SELECT ''idx_name not found, skip migration'' AS msg'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 如果旧表已有 idx_pipeline_name_del 则跳过（GORM AutoMigrate 可能已创建）
SET @idx_new_exists := (
    SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'cicd_pipeline'
      AND index_name = 'idx_pipeline_name_del'
);
SET @sql2 := IF(@idx_new_exists = 0,
    'ALTER TABLE `cicd_pipeline` ADD UNIQUE KEY `idx_pipeline_name_del` (`name`, `deleted_at`)',
    'SELECT ''idx_pipeline_name_del already exists, skip'' AS msg'
);
PREPARE stmt2 FROM @sql2;
EXECUTE stmt2;
DEALLOCATE PREPARE stmt2;

-- =====================================================
-- 完成（仅输出关键信息）
-- =====================================================
SELECT CONCAT('✅ 初始化完成 | 表数: ', COUNT(*), ' | 账号: admin / admin123 | 演示流水线: ',
              (SELECT COUNT(*) FROM cicd_pipeline)) AS result
FROM information_schema.tables WHERE table_schema = 'k8s-platform';

