-- ============================================================
-- K8sOperation 平台初始化 SQL 脚本
-- 版本: 1.0
-- 说明: 部署时一次性执行，创建数据库、表结构和初始数据
-- 执行命令: mysql -u root -p < init_database.sql
-- ============================================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `k8s-platform` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
USE `k8s-platform`;

-- ============================================================
-- 用户表
-- ============================================================
CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码',
  `created_at` int unsigned DEFAULT '0' COMMENT '创建时间',
  `modified_at` int unsigned DEFAULT '0' COMMENT '修改时间',
  `deleted_at` int unsigned DEFAULT '0' COMMENT '删除时间',
  `is_del` tinyint unsigned DEFAULT '0' COMMENT '是否删除，0表示未删除，1表示删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户表';

-- ============================================================
-- K8s集群表
-- ============================================================
CREATE TABLE IF NOT EXISTS `kube_cluster` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `cluster_name` varchar(191) NOT NULL DEFAULT '' COMMENT '集群名称',
  `kube_config` longtext NOT NULL COMMENT 'kubeconfig内容(base64编码)',
  `cluster_version` varchar(191) NOT NULL DEFAULT '' COMMENT '集群版本',
  `status` tinyint unsigned NOT NULL DEFAULT '2' COMMENT '集群状态:0正常,1异常,2未检测',
  `created_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建时间(Unix时间戳)',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '修改时间(Unix时间戳)',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间(Unix时间戳)',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否删除:0未删除,1删除',
  `last_check_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '最近一次连通性探测时间(Unix时间戳)',
  `last_error` varchar(1024) NOT NULL DEFAULT '' COMMENT '最近一次异常原因',
  PRIMARY KEY (`id`),
  KEY `idx_is_del` (`is_del`),
  KEY `idx_status` (`status`),
  KEY `idx_modified_at` (`modified_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='K8s集群表';

-- ============================================================
-- CI/CD 流水线表
-- ============================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '流水线ID',
  `name` varchar(191) NOT NULL COMMENT '流水线名称',
  `description` varchar(500) NOT NULL DEFAULT '' COMMENT '描述',
  `git_repo` varchar(500) NOT NULL COMMENT 'Git仓库URL',
  `git_branch` varchar(100) NOT NULL DEFAULT 'main' COMMENT 'Git分支',
  `jenkins_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Jenkins服务器地址',
  `jenkins_job` varchar(191) NOT NULL COMMENT 'Jenkins Job名称',
  `jenkins_credential_id` varchar(191) NOT NULL DEFAULT '' COMMENT 'Jenkins凭据ID',
  `auto_deploy` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否自动部署',
  `target_cluster_id` bigint DEFAULT NULL COMMENT '目标集群ID',
  `target_namespace` varchar(100) DEFAULT '' COMMENT '目标命名空间',
  `target_workload_kind` varchar(50) DEFAULT '' COMMENT '工作负载类型(Deployment/StatefulSet/DaemonSet)',
  `target_workload_name` varchar(200) DEFAULT '' COMMENT '工作负载名称',
  `target_container` varchar(100) DEFAULT '' COMMENT '容器名称',
  `deploy_env` varchar(20) DEFAULT 'dev' COMMENT '部署环境(dev/staging/prod)',
  `require_approval` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否需要审批',
  `last_deploy_image` varchar(500) DEFAULT '' COMMENT '最新部署镜像',
  `last_deploy_digest` varchar(100) DEFAULT '' COMMENT '最新部署镜像摘要',
  `last_deploy_time` bigint DEFAULT NULL COMMENT '最新部署时间',
  `last_deploy_status` varchar(32) DEFAULT '' COMMENT '最新部署状态',
  `last_deploy_version` varchar(100) DEFAULT '' COMMENT '最新部署版本',
  `status` varchar(50) NOT NULL DEFAULT 'idle' COMMENT '状态(idle/running/disabled)',
  `last_run_status` varchar(50) NOT NULL DEFAULT '' COMMENT '上次运行状态(success/failed/running/pending)',
  `last_run_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '上次运行时间(Unix时间戳)',
  `last_build_number` int NOT NULL DEFAULT '0' COMMENT '最近Jenkins构建号',
  `last_build_url` varchar(500) NOT NULL DEFAULT '' COMMENT '最近构建URL',
  `env_vars` json DEFAULT NULL COMMENT '环境变量JSON: [{"name":"KEY","value":"VALUE"}]',
  `deploy_config` json DEFAULT NULL COMMENT '部署配置JSON',
  `stages` json DEFAULT NULL COMMENT '流水线阶段配置JSON',
  `created_user_id` bigint NOT NULL DEFAULT '0' COMMENT '创建者ID',
  `created_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建时间(Unix时间戳)',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '修改时间(Unix时间戳)',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间(Unix时间戳)',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否删除:0未删除,1删除',
  PRIMARY KEY (`id`),
  KEY `idx_is_del` (`is_del`),
  KEY `idx_name` (`name`),
  KEY `idx_status` (`status`),
  KEY `idx_jenkins_job` (`jenkins_job`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='CI/CD流水线表';

-- ============================================================
-- CI/CD 流水线运行记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline_run` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '运行记录ID',
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `build_number` int NOT NULL DEFAULT '0' COMMENT 'Jenkins构建号',
  `jenkins_build_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Jenkins构建URL',
  `git_branch` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git Commit SHA',
  `git_commit_message` varchar(500) NOT NULL DEFAULT '' COMMENT 'Commit消息',
  `image_url` varchar(500) NOT NULL DEFAULT '' COMMENT '构建镜像地址',
  `image_digest` varchar(100) NOT NULL DEFAULT '' COMMENT '镜像摘要',
  `status` varchar(50) NOT NULL DEFAULT 'pending' COMMENT '运行状态(pending/running/success/failed/aborted)',
  `error_message` text COMMENT '错误信息',
  `callback_received` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否已收到回调',
  `trigger_user_id` bigint NOT NULL DEFAULT '0' COMMENT '触发者ID',
  `trigger_type` varchar(50) NOT NULL DEFAULT 'manual' COMMENT '触发类型(manual/webhook/schedule)',
  `duration_sec` int NOT NULL DEFAULT '0' COMMENT '运行时长(秒)',
  `started_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '开始时间(Unix时间戳)',
  `finished_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '结束时间(Unix时间戳)',
  `created_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建时间(Unix时间戳)',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '修改时间(Unix时间戳)',
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_status` (`status`),
  KEY `idx_build_number` (`build_number`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='流水线运行记录表';

-- ============================================================
-- CI/CD 流水线阶段执行记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline_stage` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '阶段ID',
  `run_id` bigint NOT NULL COMMENT '运行记录ID',
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `stage_name` varchar(100) NOT NULL COMMENT '阶段名称',
  `stage_type` varchar(50) NOT NULL COMMENT '阶段类型(checkout/build/test/push/approval/deploy)',
  `stage_order` int NOT NULL DEFAULT '0' COMMENT '执行顺序',
  `status` varchar(50) NOT NULL DEFAULT 'pending' COMMENT '执行状态(pending/running/success/failed/skipped/waiting)',
  `started_at` bigint DEFAULT NULL COMMENT '开始时间(Unix时间戳)',
  `finished_at` bigint DEFAULT NULL COMMENT '结束时间(Unix时间戳)',
  `duration_sec` int DEFAULT '0' COMMENT '执行时长(秒)',
  `logs` longtext COMMENT '阶段执行日志',
  `jenkins_stage_id` varchar(100) DEFAULT NULL COMMENT 'Jenkins阶段ID',
  `approval_user_id` bigint DEFAULT NULL COMMENT '审批人ID',
  `approval_comment` text COMMENT '审批评论',
  `approval_decision` varchar(32) DEFAULT NULL COMMENT '审批决定: approved/rejected',
  `deploy_cluster_id` bigint DEFAULT NULL COMMENT '目标集群ID',
  `deploy_namespace` varchar(100) DEFAULT NULL COMMENT '目标命名空间',
  `deploy_workload_kind` varchar(50) DEFAULT NULL COMMENT '工作负载类型(Deployment/StatefulSet/DaemonSet)',
  `deploy_workload_name` varchar(100) DEFAULT NULL COMMENT '工作负载名称',
  `deploy_container` varchar(100) DEFAULT NULL COMMENT '容器名称',
  `deploy_image` varchar(500) DEFAULT NULL COMMENT '部署镜像地址',
  `deploy_replicas` int DEFAULT NULL COMMENT '副本数',
  `error_message` text COMMENT '错误信息',
  `created_at` bigint NOT NULL COMMENT '创建时间(Unix时间戳)',
  `modified_at` bigint NOT NULL COMMENT '修改时间(Unix时间戳)',
  PRIMARY KEY (`id`),
  KEY `idx_run_id` (`run_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_status` (`status`),
  KEY `idx_stage_type` (`stage_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='流水线阶段执行记录表';

-- ============================================================
-- CI/CD 发布单表
-- ============================================================
CREATE TABLE IF NOT EXISTS `cicd_release` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `app_name` varchar(191) NOT NULL DEFAULT '' COMMENT '应用/服务名',
  `namespace` varchar(191) NOT NULL DEFAULT 'default' COMMENT '命名空间',
  `workload_kind` varchar(32) NOT NULL DEFAULT 'Deployment' COMMENT 'Deployment/StatefulSet',
  `workload_name` varchar(191) NOT NULL DEFAULT '' COMMENT '工作负载名称',
  `container_name` varchar(191) NOT NULL DEFAULT '' COMMENT '容器名',
  `strategy` varchar(32) NOT NULL DEFAULT 'rolling' COMMENT '发布策略 rolling/canary/batch',
  `timeout_sec` int unsigned NOT NULL DEFAULT '300' COMMENT '单集群超时（秒）',
  `concurrency` int unsigned NOT NULL DEFAULT '3' COMMENT '并发执行集群数',
  `status` varchar(32) NOT NULL DEFAULT 'Pending' COMMENT 'Pending/Running/Succeeded/Failed/Canceled',
  `message` varchar(1024) NOT NULL DEFAULT '' COMMENT '发布单摘要',
  `created_user_id` bigint NOT NULL DEFAULT '0' COMMENT '发起人 user.id',
  `request_id` varchar(64) NOT NULL DEFAULT '' COMMENT '幂等ID（可选）',
  `build_id` bigint NOT NULL DEFAULT '0' COMMENT '关联构建ID',
  `image_repo` varchar(512) NOT NULL DEFAULT '' COMMENT '镜像仓库',
  `image_tag` varchar(191) NOT NULL DEFAULT '' COMMENT '镜像tag',
  `image_digest` varchar(255) DEFAULT NULL COMMENT '镜像digest',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_request_id` (`request_id`),
  KEY `idx_is_del` (`is_del`),
  KEY `idx_status` (`status`),
  KEY `idx_app_name` (`app_name`),
  KEY `idx_build_id` (`build_id`),
  KEY `idx_modified_at` (`modified_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='CICD发布单表';

-- ============================================================
-- CI/CD 发布阶段表
-- ============================================================
CREATE TABLE IF NOT EXISTS `cicd_release_stage` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `release_id` bigint NOT NULL DEFAULT '0',
  `stage_name` varchar(64) NOT NULL DEFAULT '',
  `stage_order` int NOT NULL DEFAULT '0',
  `status` varchar(32) NOT NULL DEFAULT 'pending',
  `message` varchar(1024) NOT NULL DEFAULT '',
  `logs` text,
  `start_time` bigint unsigned NOT NULL DEFAULT '0',
  `end_time` bigint unsigned NOT NULL DEFAULT '0',
  `duration` bigint NOT NULL DEFAULT '0',
  `build_number` varchar(64) NOT NULL DEFAULT '',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_release_stage` (`release_id`,`stage_name`),
  KEY `idx_release_id` (`release_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='CI/CD发布阶段';

-- ============================================================
-- CI/CD 发布任务表（按集群）
-- ============================================================
CREATE TABLE IF NOT EXISTS `cicd_release_task` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `release_id` bigint NOT NULL COMMENT 'cicd_release.id',
  `cluster_id` bigint NOT NULL COMMENT 'kube_cluster.id',
  `status` varchar(32) NOT NULL DEFAULT 'Pending' COMMENT 'Pending/Running/Succeeded/Failed/Canceled',
  `message` varchar(2048) NOT NULL DEFAULT '' COMMENT '失败原因/事件摘要',
  `prev_image` varchar(512) NOT NULL DEFAULT '' COMMENT '发布前镜像（回滚）',
  `target_image` varchar(512) NOT NULL DEFAULT '' COMMENT '目标镜像',
  `started_at` bigint unsigned NOT NULL DEFAULT '0',
  `finished_at` bigint unsigned NOT NULL DEFAULT '0',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_release_cluster` (`release_id`,`cluster_id`),
  KEY `idx_release_id` (`release_id`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='CICD发布任务表（按集群）';

-- ============================================================
-- 初始数据: 创建管理员账户
-- 默认密码: admin123 (bcrypt加密)
-- ============================================================
INSERT INTO `user` (`id`, `username`, `password`, `created_at`, `modified_at`, `is_del`) VALUES
(1, 'admin', '$2a$10$qKmH4hOg4lmhpz1hG.x6V.UD1vJqxqz.gRnMlEVPn4X1K3YWF4Z0q', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0)
ON DUPLICATE KEY UPDATE `modified_at` = UNIX_TIMESTAMP();

-- ============================================================
-- 完成提示
-- ============================================================
SELECT '数据库初始化完成！' AS '状态', 
       '默认管理员账户: admin / admin123' AS '提示';

