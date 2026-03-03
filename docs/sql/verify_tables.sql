-- ============================================================
-- K8sOperation 数据库验证脚本
-- 用途: 检查数据库表结构是否完整
-- 执行: mysql -u root -p123456 < verify_tables.sql
-- ============================================================

USE `k8s-platform`;

-- 定义预期的表数量
SET @expected_tables = 11;

-- ============================================================
-- 1. 检查数据库是否存在
-- ============================================================
SELECT '========== 数据库验证开始 ==========' AS '';

SELECT IF(
    (SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = 'k8s-platform') = 1,
    '✅ 数据库 k8s-platform 存在',
    '❌ 数据库 k8s-platform 不存在'
) AS '数据库检查';

-- ============================================================
-- 2. 检查所有表是否存在
-- ============================================================
SELECT '' AS '';
SELECT '---------- 表结构检查 ----------' AS '';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'user')
        THEN '✅ user' ELSE '❌ user (缺失)' END AS '1. 用户表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'kube_cluster')
        THEN '✅ kube_cluster' ELSE '❌ kube_cluster (缺失)' END AS '2. K8s集群表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'cicd_build')
        THEN '✅ cicd_build' ELSE '❌ cicd_build (缺失)' END AS '3. 构建记录表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'cicd_pipeline')
        THEN '✅ cicd_pipeline' ELSE '❌ cicd_pipeline (缺失)' END AS '4. 流水线表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'cicd_pipeline_run')
        THEN '✅ cicd_pipeline_run' ELSE '❌ cicd_pipeline_run (缺失)' END AS '5. 运行记录表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'cicd_pipeline_stage')
        THEN '✅ cicd_pipeline_stage' ELSE '❌ cicd_pipeline_stage (缺失)' END AS '6. 阶段执行表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'cicd_environment')
        THEN '✅ cicd_environment' ELSE '❌ cicd_environment (缺失)' END AS '7. 环境管理表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'cicd_approval')
        THEN '✅ cicd_approval' ELSE '❌ cicd_approval (缺失)' END AS '8. 审批记录表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'cicd_release')
        THEN '✅ cicd_release' ELSE '❌ cicd_release (缺失)' END AS '9. 发布单表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'cicd_release_stage')
        THEN '✅ cicd_release_stage' ELSE '❌ cicd_release_stage (缺失)' END AS '10. 发布阶段表';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.TABLES WHERE TABLE_SCHEMA = 'k8s-platform' AND TABLE_NAME = 'cicd_release_task')
        THEN '✅ cicd_release_task' ELSE '❌ cicd_release_task (缺失)' END AS '11. 发布任务表';

-- ============================================================
-- 3. 统计表数量
-- ============================================================
SELECT '' AS '';
SELECT '---------- 统计汇总 ----------' AS '';

SELECT 
    COUNT(*) AS '实际表数量',
    @expected_tables AS '预期表数量',
    CASE WHEN COUNT(*) = @expected_tables 
        THEN '✅ 表数量正确' 
        ELSE CONCAT('❌ 缺少 ', @expected_tables - COUNT(*), ' 张表') 
    END AS '检查结果'
FROM information_schema.TABLES 
WHERE TABLE_SCHEMA = 'k8s-platform';

-- ============================================================
-- 4. 检查默认管理员账户
-- ============================================================
SELECT '' AS '';
SELECT '---------- 初始数据检查 ----------' AS '';

SELECT 
    CASE WHEN EXISTS (SELECT 1 FROM `user` WHERE username = 'admin' AND is_del = 0)
        THEN '✅ admin 账户已创建'
        ELSE '❌ admin 账户缺失'
    END AS '管理员账户';

-- ============================================================
-- 5. 显示各表记录数
-- ============================================================
SELECT '' AS '';
SELECT '---------- 各表记录数 ----------' AS '';

SELECT 'user' AS '表名', COUNT(*) AS '记录数' FROM `user`
UNION ALL SELECT 'kube_cluster', COUNT(*) FROM `kube_cluster`
UNION ALL SELECT 'cicd_build', COUNT(*) FROM `cicd_build`
UNION ALL SELECT 'cicd_pipeline', COUNT(*) FROM `cicd_pipeline`
UNION ALL SELECT 'cicd_pipeline_run', COUNT(*) FROM `cicd_pipeline_run`
UNION ALL SELECT 'cicd_pipeline_stage', COUNT(*) FROM `cicd_pipeline_stage`
UNION ALL SELECT 'cicd_environment', COUNT(*) FROM `cicd_environment`
UNION ALL SELECT 'cicd_approval', COUNT(*) FROM `cicd_approval`
UNION ALL SELECT 'cicd_release', COUNT(*) FROM `cicd_release`
UNION ALL SELECT 'cicd_release_stage', COUNT(*) FROM `cicd_release_stage`
UNION ALL SELECT 'cicd_release_task', COUNT(*) FROM `cicd_release_task`;

-- ============================================================
-- 6. 验证完成
-- ============================================================
SELECT '' AS '';
SELECT '========== 数据库验证完成 ==========' AS '';
