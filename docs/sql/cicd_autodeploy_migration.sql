-- ============================================================
-- CI/CD 自动部署功能数据库迁移脚本
-- 执行命令: mysql -u root -p k8s-platform < cicd_autodeploy_migration.sql
-- ============================================================

-- ==================== 1. cicd_pipeline 表添加自动部署字段 ====================

-- 添加自动部署配置字段
ALTER TABLE `cicd_pipeline` 
ADD COLUMN `auto_deploy` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否自动部署' AFTER `jenkins_credential_id`,
ADD COLUMN `target_cluster_id` bigint(20) DEFAULT NULL COMMENT '目标集群ID' AFTER `auto_deploy`,
ADD COLUMN `target_namespace` varchar(100) DEFAULT '' COMMENT '目标命名空间' AFTER `target_cluster_id`,
ADD COLUMN `target_workload_kind` varchar(50) DEFAULT '' COMMENT '工作负载类型(Deployment/StatefulSet/DaemonSet)' AFTER `target_namespace`,
ADD COLUMN `target_workload_name` varchar(200) DEFAULT '' COMMENT '工作负载名称' AFTER `target_workload_kind`,
ADD COLUMN `target_container` varchar(100) DEFAULT '' COMMENT '容器名称' AFTER `target_workload_name`,
ADD COLUMN `deploy_env` varchar(20) DEFAULT 'dev' COMMENT '部署环境(dev/staging/prod)' AFTER `target_container`,
ADD COLUMN `require_approval` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否需要审批' AFTER `deploy_env`;

-- 添加最新部署信息字段
ALTER TABLE `cicd_pipeline`
ADD COLUMN `last_deploy_image` varchar(500) DEFAULT '' COMMENT '最新部署镜像' AFTER `require_approval`,
ADD COLUMN `last_deploy_digest` varchar(100) DEFAULT '' COMMENT '最新部署镜像摘要' AFTER `last_deploy_image`,
ADD COLUMN `last_deploy_time` bigint(20) DEFAULT NULL COMMENT '最新部署时间' AFTER `last_deploy_digest`,
ADD COLUMN `last_deploy_status` varchar(32) DEFAULT '' COMMENT '最新部署状态' AFTER `last_deploy_time`,
ADD COLUMN `last_deploy_version` varchar(100) DEFAULT '' COMMENT '最新部署版本' AFTER `last_deploy_status`;

-- ==================== 2. cicd_pipeline_run 表添加回调字段 ====================

-- 添加 callback_received 字段（如果不存在）
SET @exist_callback := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline_run' AND COLUMN_NAME = 'callback_received');
SET @sql_callback := IF(@exist_callback = 0, 
    'ALTER TABLE `cicd_pipeline_run` ADD COLUMN `callback_received` tinyint(1) NOT NULL DEFAULT 0 COMMENT \'\u662f\u5426\u5df2\u6536\u5230\u56de\u8c03\' AFTER `image_digest`',
    'SELECT 1');
PREPARE stmt FROM @sql_callback;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 添加 error_message 字段（如果不存在）
SET @exist_error := (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline_run' AND COLUMN_NAME = 'error_message');
SET @sql_error := IF(@exist_error = 0, 
    'ALTER TABLE `cicd_pipeline_run` ADD COLUMN `error_message` varchar(1000) DEFAULT \'\' COMMENT \'\u9519\u8bef\u4fe1\u606f\' AFTER `stages_result`',
    'SELECT 1');
PREPARE stmt FROM @sql_error;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ==================== 3. 创建流水线阶段执行记录表 ====================

CREATE TABLE IF NOT EXISTS `cicd_pipeline_stage` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `run_id` bigint(20) NOT NULL COMMENT '运行记录ID',
    `pipeline_id` bigint(20) NOT NULL COMMENT '流水线ID',
    `stage_order` int(11) NOT NULL DEFAULT 0 COMMENT '阶段顺序',
    `stage_type` varchar(32) NOT NULL COMMENT '阶段类型: checkout/build/test/push/approval/deploy',
    `stage_name` varchar(100) NOT NULL COMMENT '阶段名称',
    `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '状态: pending/running/success/failed/skipped/waiting/aborted',
    
    -- 执行信息
    `started_at` bigint(20) DEFAULT NULL COMMENT '开始时间',
    `finished_at` bigint(20) DEFAULT NULL COMMENT '结束时间',
    `duration_sec` int(11) DEFAULT 0 COMMENT '执行耗时(秒)',
    `logs` longtext COMMENT '执行日志',
    
    -- Jenkins 构建信息
    `jenkins_stage_id` varchar(100) DEFAULT '' COMMENT 'Jenkins阶段ID',
    
    -- 审批信息（适用于 approval 类型）
    `approval_user_id` bigint(20) DEFAULT NULL COMMENT '审批人ID',
    `approval_comment` varchar(500) DEFAULT '' COMMENT '审批备注',
    `approval_decision` varchar(20) DEFAULT '' COMMENT '审批决定: approved/rejected',
    
    -- 部署信息（适用于 deploy 类型）
    `deploy_cluster_id` bigint(20) DEFAULT NULL COMMENT '部署目标集群ID',
    `deploy_namespace` varchar(100) DEFAULT '' COMMENT '部署命名空间',
    `deploy_workload_kind` varchar(50) DEFAULT '' COMMENT '工作负载类型',
    `deploy_workload_name` varchar(200) DEFAULT '' COMMENT '工作负载名称',
    `deploy_container` varchar(100) DEFAULT '' COMMENT '容器名称',
    `deploy_image` varchar(500) DEFAULT '' COMMENT '部署镜像',
    `deploy_replicas` int(11) DEFAULT 0 COMMENT '副本数',
    
    -- 错误信息
    `error_message` varchar(1000) DEFAULT '' COMMENT '错误信息',
    
    -- 元数据
    `created_at` bigint(20) NOT NULL COMMENT '创建时间',
    `modified_at` bigint(20) NOT NULL COMMENT '修改时间',
    
    PRIMARY KEY (`id`),
    KEY `idx_run_id` (`run_id`),
    KEY `idx_pipeline_id` (`pipeline_id`),
    KEY `idx_status` (`status`),
    KEY `idx_stage_type` (`stage_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流水线阶段执行记录';

-- ==================== 4. 添加索引优化查询 ====================

-- cicd_pipeline 表添加索引（忽略已存在的索引错误）
CREATE INDEX `idx_auto_deploy` ON `cicd_pipeline` (`auto_deploy`);
CREATE INDEX `idx_target_cluster` ON `cicd_pipeline` (`target_cluster_id`);

-- ==================== 完成提示 ====================
SELECT '迁移完成！请重启后端服务。' AS message;
