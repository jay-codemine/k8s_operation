-- ============================================================
-- 多租户迁移脚本
-- 所有业务表增加 tenant_id 列，默认归属租户 ID=1
-- 可重复执行（IF NOT EXISTS 保护）
-- ============================================================

-- 1. 创建租户表
CREATE TABLE IF NOT EXISTS `tenant` (
  `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '租户名称',
  `code`       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '租户编码',
  `status`     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态:1启用,0禁用',
  `tenant_id`  INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '租户ID',
  `created_at` INT UNSIGNED NOT NULL DEFAULT 0,
  `modified_at` INT UNSIGNED NOT NULL DEFAULT 0,
  `deleted_at` INT UNSIGNED NOT NULL DEFAULT 0,
  `is_del`     TINYINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_tenant_id` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 2. 默认租户
INSERT IGNORE INTO `tenant` (`id`, `name`, `code`, `status`, `tenant_id`, `created_at`)
VALUES (1, '默认租户', 'default', 1, 1, UNIX_TIMESTAMP());

-- ============================================================
-- 3. 所有业务表加 tenant_id 列
-- ============================================================

-- 用户与权限
ALTER TABLE `user`                    ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `sys_user_role`           ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `sys_user_cluster`        ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `sys_role`                ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `sys_permission`          ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `sys_role_permission`     ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- K8s 集群
ALTER TABLE `kube_cluster`            ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- CICD 流水线
ALTER TABLE `cicd_pipeline`           ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_pipeline_run`       ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_pipeline_stage`     ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_pipeline_target`    ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_pipeline_template`  ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- CICD 发布
ALTER TABLE `cicd_release`            ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_release_stage`      ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_release_task`       ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_artifact`           ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_build`              ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_build_agent`        ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_environment`        ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_approval`           ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- 资源模板
ALTER TABLE `cicd_resource_template`  ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_env_resource_rule`  ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_deploy_approval`    ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `cicd_resource_change_log` ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- 监控
ALTER TABLE `monitor_datasource`      ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `monitor_alert_rule`      ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `monitor_alert_event`     ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `monitor_notify_channel`  ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `monitor_silence_rule`    ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `monitor_inhibit_rule`    ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `monitor_aggregate_rule`  ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `monitor_notify_template`  ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `monitor_notify_route_policy` ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- 镜像管理
ALTER TABLE `image_registry`          ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `image_cleanup_policy`    ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `image_cleanup_log`       ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- 审计 & 平台
ALTER TABLE `audit_log`               ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `platform_settings`       ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- AI
ALTER TABLE `ai_conversations`        ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `ai_messages`             ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `ai_approval_requests`    ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `ai_approval_logs`        ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- AIOps
ALTER TABLE `aiops_analysis_record`   ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `aiops_inspection_report` ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);

-- 应用商城
ALTER TABLE `app_store_app`           ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `app_store_component`     ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
ALTER TABLE `app_store_install`       ADD COLUMN IF NOT EXISTS `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1, ADD INDEX IF NOT EXISTS `idx_tenant_id` (`tenant_id`);
