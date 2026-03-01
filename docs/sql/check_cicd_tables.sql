-- ============================================================
-- CI/CD 表字段验证脚本
-- 执行命令: mysql -u root -p k8s-platform < check_cicd_tables.sql
-- ============================================================

USE `k8s-platform`;

-- 检查 cicd_pipeline 表的字段
SELECT '=== cicd_pipeline 表字段 ===' AS info;
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline'
ORDER BY ORDINAL_POSITION;

-- 检查 cicd_pipeline_run 表的字段
SELECT '=== cicd_pipeline_run 表字段 ===' AS info;
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_pipeline_run'
ORDER BY ORDINAL_POSITION;

-- 检查 cicd_release 表是否存在
SELECT '=== cicd_release 表字段 ===' AS info;
SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'cicd_release'
ORDER BY ORDINAL_POSITION;

-- 检查缺失的关键字段
SELECT '=== 缺失字段检查 ===' AS info;

SELECT 'cicd_pipeline.auto_deploy' AS field, 
       IF(COUNT(*) > 0, '存在', '缺失') AS status
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'cicd_pipeline' 
  AND COLUMN_NAME = 'auto_deploy'

UNION ALL

SELECT 'cicd_pipeline.require_approval', 
       IF(COUNT(*) > 0, '存在', '缺失')
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'cicd_pipeline' 
  AND COLUMN_NAME = 'require_approval'

UNION ALL

SELECT 'cicd_pipeline_run.image_url', 
       IF(COUNT(*) > 0, '存在', '缺失')
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'cicd_pipeline_run' 
  AND COLUMN_NAME = 'image_url'

UNION ALL

SELECT 'cicd_release 表', 
       IF(COUNT(*) > 0, '存在', '缺失')
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'cicd_release';
