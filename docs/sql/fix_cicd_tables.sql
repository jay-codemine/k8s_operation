-- ============================================================
-- CI/CD 表结构修复脚本（一键修复所有缺失字段）
-- 执行命令: mysql -u root -p k8s-platform < docs/sql/fix_cicd_tables.sql
-- 兼容 MySQL 5.7 和 8.0
-- ============================================================

USE `k8s-platform`;

-- 关闭严格模式，忽略重复字段错误
SET sql_mode = '';

-- ==================== 1. 创建 cicd_release 表（如果不存在） ====================

CREATE TABLE IF NOT EXISTS `cicd_release` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '发布单ID',
  `app_name` varchar(191) NOT NULL DEFAULT '' COMMENT '应用名称',
  `namespace` varchar(191) NOT NULL DEFAULT 'default' COMMENT '命名空间',
  `workload_kind` varchar(50) NOT NULL DEFAULT 'Deployment' COMMENT '工作负载类型',
  `workload_name` varchar(191) NOT NULL DEFAULT '' COMMENT '工作负载名称',
  `container_name` varchar(191) NOT NULL DEFAULT '' COMMENT '容器名称',
  `strategy` varchar(50) NOT NULL DEFAULT 'rolling' COMMENT '发布策略',
  `timeout_sec` int UNSIGNED NOT NULL DEFAULT 300 COMMENT '超时时间(秒)',
  `concurrency` int UNSIGNED NOT NULL DEFAULT 3 COMMENT '并发数',
  `status` varchar(50) NOT NULL DEFAULT 'Pending' COMMENT '状态',
  `message` varchar(1024) NOT NULL DEFAULT '' COMMENT '状态消息',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建用户ID',
  `request_id` varchar(191) NOT NULL DEFAULT '' COMMENT '请求ID',
  `build_id` bigint NOT NULL DEFAULT 0 COMMENT 'Jenkins构建ID',
  `image_repo` varchar(500) NOT NULL DEFAULT '' COMMENT '镜像仓库地址',
  `image_tag` varchar(191) NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(191) DEFAULT NULL COMMENT '镜像摘要',
  `created_at` bigint UNSIGNED NOT NULL DEFAULT 0,
  `modified_at` bigint UNSIGNED NOT NULL DEFAULT 0,
  `deleted_at` bigint UNSIGNED NOT NULL DEFAULT 0,
  `is_del` tinyint UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `idx_app_name` (`app_name`),
  INDEX `idx_status` (`status`),
  INDEX `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CICD发布单表';

-- ==================== 2. 创建 cicd_release_task 表（如果不存在） ====================

CREATE TABLE IF NOT EXISTS `cicd_release_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `release_id` bigint NOT NULL DEFAULT 0,
  `cluster_id` bigint NOT NULL DEFAULT 0,
  `status` varchar(50) NOT NULL DEFAULT 'Pending',
  `message` varchar(1024) NOT NULL DEFAULT '',
  `prev_image` varchar(500) NOT NULL DEFAULT '',
  `target_image` varchar(500) NOT NULL DEFAULT '',
  `started_at` bigint UNSIGNED NOT NULL DEFAULT 0,
  `finished_at` bigint UNSIGNED NOT NULL DEFAULT 0,
  `created_at` bigint UNSIGNED NOT NULL DEFAULT 0,
  `modified_at` bigint UNSIGNED NOT NULL DEFAULT 0,
  `deleted_at` bigint UNSIGNED NOT NULL DEFAULT 0,
  `is_del` tinyint UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  INDEX `idx_release_id` (`release_id`),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CICD发布任务表';

-- ==================== 3. 创建 cicd_pipeline_stage 表 ====================

CREATE TABLE IF NOT EXISTS `cicd_pipeline_stage` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `run_id` bigint(20) NOT NULL COMMENT '运行记录ID',
    `pipeline_id` bigint(20) NOT NULL COMMENT '流水线ID',
    `stage_order` int(11) NOT NULL DEFAULT 0 COMMENT '阶段顺序',
    `stage_type` varchar(32) NOT NULL COMMENT '阶段类型',
    `stage_name` varchar(100) NOT NULL COMMENT '阶段名称',
    `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '状态',
    `started_at` bigint(20) DEFAULT NULL COMMENT '开始时间',
    `finished_at` bigint(20) DEFAULT NULL COMMENT '结束时间',
    `duration_sec` int(11) DEFAULT 0 COMMENT '执行耗时(秒)',
    `logs` longtext COMMENT '执行日志',
    `jenkins_stage_id` varchar(100) DEFAULT '' COMMENT 'Jenkins阶段ID',
    `approval_user_id` bigint(20) DEFAULT NULL COMMENT '审批人ID',
    `approval_comment` varchar(500) DEFAULT '' COMMENT '审批备注',
    `approval_decision` varchar(20) DEFAULT '' COMMENT '审批决定',
    `deploy_cluster_id` bigint(20) DEFAULT NULL COMMENT '部署目标集群ID',
    `deploy_namespace` varchar(100) DEFAULT '' COMMENT '部署命名空间',
    `deploy_workload_kind` varchar(50) DEFAULT '' COMMENT '工作负载类型',
    `deploy_workload_name` varchar(200) DEFAULT '' COMMENT '工作负载名称',
    `deploy_container` varchar(100) DEFAULT '' COMMENT '容器名称',
    `deploy_image` varchar(500) DEFAULT '' COMMENT '部署镜像',
    `deploy_replicas` int(11) DEFAULT 0 COMMENT '副本数',
    `error_message` varchar(1000) DEFAULT '' COMMENT '错误信息',
    `created_at` bigint(20) NOT NULL COMMENT '创建时间',
    `modified_at` bigint(20) NOT NULL COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `idx_run_id` (`run_id`),
    KEY `idx_pipeline_id` (`pipeline_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流水线阶段执行记录';

-- ==================== 4. cicd_pipeline 表添加字段（忽略已存在的字段） ====================

-- 添加 auto_deploy
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'auto_deploy');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN auto_deploy tinyint(1) NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 target_cluster_id
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'target_cluster_id');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN target_cluster_id bigint(20) DEFAULT NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 target_namespace
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'target_namespace');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN target_namespace varchar(100) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 target_workload_kind
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'target_workload_kind');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN target_workload_kind varchar(50) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 target_workload_name
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'target_workload_name');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN target_workload_name varchar(200) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 target_container
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'target_container');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN target_container varchar(100) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 deploy_env
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'deploy_env');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN deploy_env varchar(20) DEFAULT "dev"', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 require_approval
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'require_approval');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN require_approval tinyint(1) NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 last_deploy_image
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'last_deploy_image');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN last_deploy_image varchar(500) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 last_deploy_digest
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'last_deploy_digest');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN last_deploy_digest varchar(100) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 last_deploy_time
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'last_deploy_time');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN last_deploy_time bigint(20) DEFAULT NULL', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 last_deploy_status
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'last_deploy_status');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN last_deploy_status varchar(32) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 last_deploy_version
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline' AND COLUMN_NAME = 'last_deploy_version');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline ADD COLUMN last_deploy_version varchar(100) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ==================== 5. cicd_pipeline_run 表添加字段 ====================

-- 添加 image_url
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline_run' AND COLUMN_NAME = 'image_url');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline_run ADD COLUMN image_url varchar(500) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 image_digest
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline_run' AND COLUMN_NAME = 'image_digest');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline_run ADD COLUMN image_digest varchar(100) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 callback_received
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline_run' AND COLUMN_NAME = 'callback_received');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline_run ADD COLUMN callback_received tinyint(1) NOT NULL DEFAULT 0', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 添加 error_message
SET @exist := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline_run' AND COLUMN_NAME = 'error_message');
SET @sql := IF(@exist = 0, 'ALTER TABLE cicd_pipeline_run ADD COLUMN error_message varchar(1000) DEFAULT ""', 'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ==================== 完成 ====================
SELECT '修复完成！请重启后端服务。' AS message;
