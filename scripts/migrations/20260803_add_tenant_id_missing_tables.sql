-- ============================================================
-- 迁移：为所有业务表补齐 tenant_id 列 + idx_tenant_id 索引（MySQL 8.x 可用）
--
-- 背景：
--   多租户隔离依赖 pkg/tenant.NewScopedDB 注入的 `WHERE tenant_id = ?`，
--   凡是被租户隔离 DB 访问的表都必须有 tenant_id 列。
--
--   docs/sql/k8s_platform_full_init.sql 的「多租户扩展」段用的是
--   `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`，这是 MariaDB 专有语法，
--   MySQL 8.x 直接报 ERROR 1064，那 48 条 ALTER 会全部失败（且容易被
--   --force 之类的选项掩盖成静默失效）。因此在 MySQL 上必须执行本脚本。
--   本脚本改用 information_schema 判定 + PREPARE 动态 DDL，MySQL 8.x 幂等可靠。
--
--   下列 6 张表既不在那段 SQL 的列表里，也没有对应 Go model
--   （现网多数表的 tenant_id 实际由 GORM AutoMigrate 从 models.Base 建出），
--   所以任何路径都建不出列，属于必须靠本脚本补齐的：
--     cicd_language_profile / iam_group / iam_group_user
--     iam_project / iam_project_member / iam_role_template
--
-- 幂等：重复执行安全（列/索引已存在时跳过，表不存在时跳过）。
-- 使用：mysql -h<host> -P<port> -u<user> -p <dbname> < 20260803_add_tenant_id_missing_tables.sql
-- ============================================================

DROP PROCEDURE IF EXISTS `__add_tenant_id`;

DELIMITER $$
CREATE PROCEDURE `__add_tenant_id`(IN tbl VARCHAR(64))
BEGIN
  -- 表不存在则直接跳过（不同版本部署可能没有这些遗留表）
  IF (SELECT COUNT(*) FROM information_schema.TABLES
      WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = tbl) = 0 THEN
    SELECT CONCAT('table ', tbl, ' not found, skip') AS msg;
  ELSE
    -- 1) 补列
    IF (SELECT COUNT(*) FROM information_schema.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = tbl
          AND COLUMN_NAME = 'tenant_id') = 0 THEN
      SET @ddl = CONCAT('ALTER TABLE `', tbl,
        '` ADD COLUMN `tenant_id` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT ''归属租户ID(tenant.id)''');
      PREPARE s FROM @ddl; EXECUTE s; DEALLOCATE PREPARE s;
      SELECT CONCAT('column added: ', tbl, '.tenant_id') AS msg;
    ELSE
      SELECT CONCAT('column exists: ', tbl, '.tenant_id, skip') AS msg;
    END IF;

    -- 2) 补索引
    IF (SELECT COUNT(*) FROM information_schema.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = tbl
          AND INDEX_NAME = 'idx_tenant_id') = 0 THEN
      SET @idx = CONCAT('ALTER TABLE `', tbl, '` ADD INDEX `idx_tenant_id` (`tenant_id`)');
      PREPARE s2 FROM @idx; EXECUTE s2; DEALLOCATE PREPARE s2;
      SELECT CONCAT('index added: ', tbl, '.idx_tenant_id') AS msg;
    ELSE
      SELECT CONCAT('index exists: ', tbl, '.idx_tenant_id, skip') AS msg;
    END IF;
  END IF;
END$$
DELIMITER ;

-- ---------- 无 Go model、任何路径都建不出列的 6 张表 ----------
CALL `__add_tenant_id`('cicd_language_profile');
CALL `__add_tenant_id`('iam_group');
CALL `__add_tenant_id`('iam_group_user');
CALL `__add_tenant_id`('iam_project');
CALL `__add_tenant_id`('iam_project_member');
CALL `__add_tenant_id`('iam_role_template');

-- ---------- 以下与 init SQL「多租户扩展」段一一对应 ----------
-- 在 MySQL 上那段 ALTER 全部失败，故此处重做一遍（已有列则跳过）
CALL `__add_tenant_id`('user');
CALL `__add_tenant_id`('sys_user_role');
CALL `__add_tenant_id`('sys_user_cluster');
CALL `__add_tenant_id`('kube_cluster');
CALL `__add_tenant_id`('cicd_pipeline');
CALL `__add_tenant_id`('cicd_pipeline_run');
CALL `__add_tenant_id`('cicd_pipeline_stage');
CALL `__add_tenant_id`('cicd_pipeline_target');
CALL `__add_tenant_id`('cicd_pipeline_template');
CALL `__add_tenant_id`('cicd_release');
CALL `__add_tenant_id`('cicd_release_stage');
CALL `__add_tenant_id`('cicd_release_task');
CALL `__add_tenant_id`('cicd_artifact');
CALL `__add_tenant_id`('cicd_build');
CALL `__add_tenant_id`('cicd_build_agent');
CALL `__add_tenant_id`('cicd_environment');
CALL `__add_tenant_id`('cicd_approval');
CALL `__add_tenant_id`('cicd_resource_template');
CALL `__add_tenant_id`('cicd_env_resource_rule');
CALL `__add_tenant_id`('cicd_deploy_approval');
CALL `__add_tenant_id`('cicd_resource_change_log');
CALL `__add_tenant_id`('monitor_datasource');
CALL `__add_tenant_id`('monitor_alert_rule');
CALL `__add_tenant_id`('monitor_alert_event');
CALL `__add_tenant_id`('monitor_notify_channel');
CALL `__add_tenant_id`('monitor_silence_rule');
CALL `__add_tenant_id`('monitor_inhibit_rule');
CALL `__add_tenant_id`('monitor_aggregate_rule');
CALL `__add_tenant_id`('monitor_notify_template');
CALL `__add_tenant_id`('monitor_notify_route_policy');
CALL `__add_tenant_id`('image_registry');
CALL `__add_tenant_id`('image_cleanup_policy');
CALL `__add_tenant_id`('image_cleanup_log');
CALL `__add_tenant_id`('audit_log');
CALL `__add_tenant_id`('platform_settings');
CALL `__add_tenant_id`('ai_conversations');
CALL `__add_tenant_id`('ai_messages');
CALL `__add_tenant_id`('ai_approval_requests');
CALL `__add_tenant_id`('ai_approval_logs');
CALL `__add_tenant_id`('aiops_analysis_record');
CALL `__add_tenant_id`('aiops_inspection_report');
CALL `__add_tenant_id`('app_store_apps');
CALL `__add_tenant_id`('app_store_components');
CALL `__add_tenant_id`('app_store_installs');

-- RBAC 表：models.SysRole/SysUserRole/SysPermission/SysRolePermission 均未定义
-- TenantID 字段，列只能由本脚本建出；而 models.IsSuperAdmin/HasUserPermission
-- 的 SQL 又强制过滤 sys_role.tenant_id 与 sys_user_role.tenant_id，缺列会直接报错
CALL `__add_tenant_id`('sys_role');
CALL `__add_tenant_id`('sys_permission');
CALL `__add_tenant_id`('sys_role_permission');

-- IAM 环境管理
CALL `__add_tenant_id`('iam_env_audit_log');
CALL `__add_tenant_id`('iam_env_binding');
CALL `__add_tenant_id`('iam_grant');

DROP PROCEDURE IF EXISTS `__add_tenant_id`;

-- ============================================================
-- 校验：执行后应返回 0 行（即全库业务表都有 tenant_id）
-- 排除 information_schema 视图类与无需隔离的表
-- ============================================================
SELECT t.TABLE_NAME AS table_missing_tenant_id
FROM information_schema.TABLES t
WHERE t.TABLE_SCHEMA = DATABASE()
  AND t.TABLE_TYPE = 'BASE TABLE'
  AND NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS c
    WHERE c.TABLE_SCHEMA = t.TABLE_SCHEMA
      AND c.TABLE_NAME = t.TABLE_NAME
      AND c.COLUMN_NAME = 'tenant_id'
  )
ORDER BY t.TABLE_NAME;
