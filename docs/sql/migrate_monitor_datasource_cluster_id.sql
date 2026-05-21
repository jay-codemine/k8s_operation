-- =====================================================
-- 监控数据源关联集群迁移
-- 用途：让 monitor_datasource 关联到具体 K8s 集群（kube_cluster.id）
-- 这样前端"监控视野"可以按集群组装数据源 chip-tab，
-- 选生产集群 → 自动列出 prom-prd 数据源；选测试集群 → 列出 prom-test
--
-- 兼容性：
--   - cluster_id = 0 表示"全局共享"数据源（旧数据默认值），任意集群都能看到
--   - 已存在该列时不会报错（用 information_schema 判断）
--
-- 适用：MySQL 5.7+ / 8.x
-- =====================================================

-- 1) 加列（幂等：已存在则跳过）
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
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2) 加索引（幂等）
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
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 3) 校验
SELECT
    '✅ migration done' AS status,
    COUNT(*) AS total_datasources,
    SUM(CASE WHEN cluster_id = 0 THEN 1 ELSE 0 END) AS global_count,
    SUM(CASE WHEN cluster_id > 0 THEN 1 ELSE 0 END) AS bound_count
FROM monitor_datasource WHERE is_del = 0;
