-- ============================================================
-- 迁移：为 cicd_environment 表新增「失败自动回滚」开关列
-- 对应模型字段：models.CicdEnvironment.AutoRollbackOnFail (bool)
-- 场景：部署失败后，若目标环境开启该开关，则自动回滚到部署前版本（建议生产环境开启）
--
-- 说明：运行配置 SkipAutoMigrate=true，GORM 不会自动建列，需手动执行本脚本。
-- 幂等：重复执行安全（列已存在时跳过）。
-- 使用：mysql -h<host> -P<port> -u<user> -p <dbname> < 20260719_add_auto_rollback_on_fail.sql
-- ============================================================

SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'cicd_environment'
    AND COLUMN_NAME = 'auto_rollback_on_fail'
);

SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `cicd_environment` ADD COLUMN `auto_rollback_on_fail` tinyint(1) NOT NULL DEFAULT ''0'' COMMENT ''部署失败时自动回滚到上一版本(建议生产环境开启)'' AFTER `require_approval`',
  'SELECT ''column auto_rollback_on_fail already exists, skip'' AS msg'
);

PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
