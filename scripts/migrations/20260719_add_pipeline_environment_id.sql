-- ============================================================
-- 迁移：为 cicd_pipeline 表新增「关联环境ID」列，实现环境强隔离
-- 对应模型字段：models.CicdPipeline.EnvironmentID (int64)
-- 场景：流水线绑定 cicd_environment 后，命名空间/集群/审批策略以环境为单一事实来源
--
-- 说明：运行配置 SkipAutoMigrate=true 时，GORM 不会自动建列，需手动执行本脚本。
-- 幂等：重复执行安全（列已存在时跳过；回填不覆盖已绑定记录）。
-- 使用：mysql -h<host> -P<port> -u<user> -p <dbname> < 20260719_add_pipeline_environment_id.sql
-- ============================================================

-- 1) 新增 environment_id 列（0=未绑定）
SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'cicd_pipeline'
    AND COLUMN_NAME = 'environment_id'
);

SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE `cicd_pipeline` ADD COLUMN `environment_id` bigint NOT NULL DEFAULT 0 COMMENT ''关联环境ID(cicd_environment)，0=未绑定'' AFTER `auto_deploy`',
  'SELECT ''column environment_id already exists, skip'' AS msg'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2) 为 environment_id 建立普通索引（便于按环境过滤）
SET @idx_exists := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'cicd_pipeline'
    AND INDEX_NAME = 'idx_cicd_pipeline_environment_id'
);

SET @idx_ddl := IF(
  @idx_exists = 0,
  'ALTER TABLE `cicd_pipeline` ADD INDEX `idx_cicd_pipeline_environment_id` (`environment_id`)',
  'SELECT ''index idx_cicd_pipeline_environment_id already exists, skip'' AS msg'
);
PREPARE stmt2 FROM @idx_ddl;
EXECUTE stmt2;
DEALLOCATE PREPARE stmt2;

-- 3) 存量回填：把未绑定环境的旧流水线按 target_namespace 匹配 cicd_environment.namespace 回填
--    幂等：仅处理 environment_id=0 且命名空间能匹配到环境的记录，不覆盖已绑定的
UPDATE cicd_pipeline p
SET p.environment_id = (
  SELECT e.id FROM cicd_environment e
  WHERE e.namespace = p.target_namespace AND e.is_del = 0
  ORDER BY e.id LIMIT 1
)
WHERE p.environment_id = 0 AND p.is_del = 0 AND COALESCE(p.target_namespace, '') <> ''
  AND EXISTS (SELECT 1 FROM cicd_environment e2 WHERE e2.namespace = p.target_namespace AND e2.is_del = 0);
