-- ============================================================
-- RBAC multi-tenant: make role names unique PER TENANT
--
-- sys_role.name carried a global UNIQUE index (idx_sys_role_name), so seeding
-- the three default roles for a second tenant fails with ER_DUP_ENTRY 1062 --
-- every tenant needs its own "super_admin" row. Replace it with a composite
-- unique key on (tenant_id, name).
--
-- sys_permission.name is deliberately left global: permissions are a shared
-- catalogue, tenant isolation happens through sys_user_role.tenant_id.
--
-- Idempotent, MySQL 8.x only. Note that "DROP INDEX IF EXISTS" is MariaDB
-- syntax and raises ER_PARSE_ERROR 1064 here, hence the information_schema
-- probe plus PREPARE/EXECUTE dance.
-- ============================================================

-- 1. drop the old single-column unique index if it is still there
SET @has_old := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_role'
    AND INDEX_NAME = 'idx_sys_role_name'
);
SET @sql := IF(@has_old > 0,
  'ALTER TABLE sys_role DROP INDEX idx_sys_role_name',
  'SELECT ''idx_sys_role_name already gone'' AS msg');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 2. add the composite unique key if missing
SET @has_new := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_role'
    AND INDEX_NAME = 'uk_sys_role_tenant_name'
);
SET @sql := IF(@has_new = 0,
  'ALTER TABLE sys_role ADD UNIQUE KEY uk_sys_role_tenant_name (tenant_id, name)',
  'SELECT ''uk_sys_role_tenant_name already present'' AS msg');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 3. report the resulting index layout
SELECT INDEX_NAME, NON_UNIQUE, SEQ_IN_INDEX, COLUMN_NAME
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_role'
ORDER BY INDEX_NAME, SEQ_IN_INDEX;
