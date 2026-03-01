-- CI/CD 部署配置扩展
-- 执行时间: 2026-02-28
-- 功能: 为流水线添加部署配置、环境管理、审批流程支持

-- 1. 扩展 cicd_pipeline 表，添加部署配置字段
ALTER TABLE `cicd_pipeline` 
    ADD COLUMN `auto_deploy` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否自动部署' AFTER `jenkins_credential_id`,
    ADD COLUMN `target_cluster_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '目标集群ID' AFTER `auto_deploy`,
    ADD COLUMN `target_namespace` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '目标命名空间' AFTER `target_cluster_id`,
    ADD COLUMN `target_workload_kind` VARCHAR(50) NOT NULL DEFAULT 'Deployment' COMMENT '工作负载类型' AFTER `target_namespace`,
    ADD COLUMN `target_workload_name` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '工作负载名称' AFTER `target_workload_kind`,
    ADD COLUMN `target_container` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '容器名称' AFTER `target_workload_name`,
    ADD COLUMN `deploy_env` VARCHAR(20) NOT NULL DEFAULT 'dev' COMMENT '部署环境(dev/staging/prod)' AFTER `target_container`,
    ADD COLUMN `require_approval` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否需要审批' AFTER `deploy_env`,
    ADD COLUMN `last_deploy_image` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '最新部署镜像' AFTER `require_approval`,
    ADD COLUMN `last_deploy_digest` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '最新部署镜像摘要' AFTER `last_deploy_image`,
    ADD COLUMN `last_deploy_time` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '最新部署时间' AFTER `last_deploy_digest`,
    ADD COLUMN `last_deploy_status` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '最新部署状态' AFTER `last_deploy_time`,
    ADD COLUMN `last_deploy_version` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '最新部署版本' AFTER `last_deploy_status`;

-- 2. 创建环境管理表
CREATE TABLE IF NOT EXISTS `cicd_environment` (
    `id` BIGINT(20) NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL COMMENT '环境名称(dev/staging/prod)',
    `display_name` VARCHAR(100) NOT NULL COMMENT '显示名称',
    `description` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '描述',
    `cluster_id` BIGINT(20) NOT NULL COMMENT '关联集群ID',
    `namespace` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '默认命名空间',
    `color` VARCHAR(20) NOT NULL DEFAULT '#1890ff' COMMENT '环境颜色标识',
    `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
    `require_approval` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否需要审批',
    `approval_users` JSON COMMENT '审批人员列表',
    `created_user_id` BIGINT(20) NOT NULL DEFAULT 0,
    `created_at` BIGINT(20) NOT NULL DEFAULT 0,
    `modified_at` BIGINT(20) NOT NULL DEFAULT 0,
    `deleted_at` BIGINT(20) NOT NULL DEFAULT 0,
    `is_del` TINYINT(1) NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`, `is_del`),
    KEY `idx_cluster` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CI/CD 环境管理';

-- 3. 创建审批流程表
CREATE TABLE IF NOT EXISTS `cicd_approval` (
    `id` BIGINT(20) NOT NULL AUTO_INCREMENT,
    `pipeline_id` BIGINT(20) NOT NULL COMMENT '流水线ID',
    `pipeline_run_id` BIGINT(20) NOT NULL COMMENT '运行记录ID',
    `release_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '发布单ID',
    `env_name` VARCHAR(50) NOT NULL COMMENT '目标环境',
    `status` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '状态(pending/approved/rejected/expired)',
    `image` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '待部署镜像',
    `image_digest` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '镜像摘要',
    `request_user_id` BIGINT(20) NOT NULL COMMENT '申请人',
    `request_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '申请原因',
    `approve_user_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '审批人',
    `approve_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '审批意见',
    `approve_time` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '审批时间',
    `expire_time` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '过期时间',
    `created_at` BIGINT(20) NOT NULL DEFAULT 0,
    `modified_at` BIGINT(20) NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    KEY `idx_pipeline` (`pipeline_id`),
    KEY `idx_status` (`status`),
    KEY `idx_request_user` (`request_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CI/CD 审批记录';

-- 4. 为 cicd_pipeline_run 添加部署配置快照
ALTER TABLE `cicd_pipeline_run`
    ADD COLUMN `target_cluster_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '目标集群ID' AFTER `image_digest`,
    ADD COLUMN `target_namespace` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '目标命名空间' AFTER `target_cluster_id`,
    ADD COLUMN `target_workload` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '目标工作负载' AFTER `target_namespace`,
    ADD COLUMN `deploy_env` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '部署环境' AFTER `target_workload`,
    ADD COLUMN `release_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '关联发布单ID' AFTER `deploy_env`,
    ADD COLUMN `approval_id` BIGINT(20) NOT NULL DEFAULT 0 COMMENT '关联审批ID' AFTER `release_id`;

-- 5. 初始化默认环境数据
INSERT INTO `cicd_environment` (`name`, `display_name`, `description`, `cluster_id`, `color`, `sort_order`, `require_approval`, `created_at`, `modified_at`)
VALUES 
    ('dev', '开发环境', '用于开发测试', 1, '#52c41a', 1, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
    ('staging', '预发环境', '用于预发布测试', 1, '#faad14', 2, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
    ('prod', '生产环境', '生产线上环境', 1, '#f5222d', 3, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `modified_at` = UNIX_TIMESTAMP();

-- 6. 添加索引优化查询
ALTER TABLE `cicd_pipeline` ADD INDEX `idx_deploy_env` (`deploy_env`);
ALTER TABLE `cicd_pipeline` ADD INDEX `idx_target_cluster` (`target_cluster_id`);
