-- =====================================================
-- K8s Platform 完整数据库初始化脚本
-- 版本: 2.2.0
-- 日期: 2026-03-15
-- 说明: 一键创建数据库、所有表结构、视图和初始数据
-- 使用: mysql -u root -p123456 --default-character-set=utf8mb4 -e "source k8s_platform_full_init.sql"
-- =====================================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `k8s-platform` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `k8s-platform`;

-- =====================================================
-- 1. 用户表
-- =====================================================
CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码(加密)',
  `role` varchar(50) NOT NULL DEFAULT 'user' COMMENT '基础角色(兼容旧版)',
  `email` varchar(191) DEFAULT NULL COMMENT '邮箱',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号',
  `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用',
  `created_at` int unsigned DEFAULT 0,
  `modified_at` int unsigned DEFAULT 0,
  `deleted_at` int unsigned DEFAULT 0,
  `is_del` tinyint unsigned DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- =====================================================
-- 2. K8s集群表 (含环境字段)
-- =====================================================
CREATE TABLE IF NOT EXISTS `kube_cluster` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `cluster_name` varchar(191) NOT NULL DEFAULT '' COMMENT '集群名称',
  `kube_config` longtext NOT NULL COMMENT 'KubeConfig(加密存储)',
  `cluster_version` varchar(191) NOT NULL DEFAULT '' COMMENT '集群版本',
  `status` tinyint unsigned NOT NULL DEFAULT 2 COMMENT '状态:0=正常,1=异常,2=未检测',
  -- 环境相关字段
  `env_type` varchar(50) DEFAULT 'development' COMMENT '环境类型:development,testing,staging,production',
  `env_display_name` varchar(100) DEFAULT '' COMMENT '环境显示名称',
  `env_level` int DEFAULT 1 COMMENT '环境级别:1-4(开发到生产)',
  `access_mode` varchar(50) DEFAULT 'restricted' COMMENT '访问模式:public,restricted,private',
  `require_approval` tinyint(1) DEFAULT 0 COMMENT '操作是否需要审批',
  `approval_users` json DEFAULT NULL COMMENT '审批人列表',
  `env_color` varchar(20) DEFAULT '' COMMENT '环境颜色标识',
  `env_description` varchar(500) DEFAULT '' COMMENT '环境描述',
  `env_labels` json DEFAULT NULL COMMENT '环境标签',
  `project_ids` json DEFAULT NULL COMMENT '关联项目ID列表',
  -- 时间和状态字段
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  `last_check_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '最后检查时间',
  `last_error` varchar(1024) NOT NULL DEFAULT '' COMMENT '最后错误信息',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_modified` (`modified_at`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='K8s集群配置表';

-- =====================================================
-- 3. RBAC - 系统角色表
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_role` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '角色标识(唯一)',
  `display_name` varchar(100) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `role_type` varchar(30) NOT NULL DEFAULT 'custom' COMMENT '角色类型:super_admin,cluster_admin,developer,viewer,custom',
  `is_system` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否系统内置',
  `color` varchar(20) DEFAULT '#1890ff' COMMENT '角色颜色标识',
  `icon` varchar(50) DEFAULT 'user' COMMENT '图标',
  `sort_order` int DEFAULT 0 COMMENT '排序',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统角色表';

-- =====================================================
-- 4. RBAC - 系统权限表
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '权限标识(唯一)',
  `display_name` varchar(100) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `resource_type` varchar(50) NOT NULL COMMENT '资源类型:cluster,namespace,deployment,pod等',
  `action` varchar(30) NOT NULL COMMENT '操作类型:view,create,update,delete,exec,manage',
  `parent_id` bigint DEFAULT 0 COMMENT '父权限ID',
  `path` varchar(200) DEFAULT '' COMMENT '权限路径(树形展示用)',
  `sort_order` int DEFAULT 0 COMMENT '排序',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统权限定义表';

-- =====================================================
-- 5. RBAC - 角色权限关联表
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_role_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `permission_id` bigint NOT NULL COMMENT '权限ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- =====================================================
-- 6. RBAC - 用户角色关联表
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `created_by` bigint DEFAULT 0 COMMENT '创建人ID',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- =====================================================
-- 7. RBAC - 用户集群权限表 (细粒度权限控制)
-- =====================================================
CREATE TABLE IF NOT EXISTS `sys_user_cluster` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `cluster_id` bigint NOT NULL COMMENT '集群ID',
  `role_type` varchar(30) NOT NULL DEFAULT 'viewer' COMMENT '在该集群的角色类型',
  `namespaces` text DEFAULT NULL COMMENT '可访问的命名空间(JSON数组,空表示全部)',
  `can_view` tinyint(1) NOT NULL DEFAULT 1 COMMENT '查看权限',
  `can_create` tinyint(1) NOT NULL DEFAULT 0 COMMENT '创建权限',
  `can_update` tinyint(1) NOT NULL DEFAULT 0 COMMENT '更新权限',
  `can_delete` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除权限',
  `can_exec` tinyint(1) NOT NULL DEFAULT 0 COMMENT '执行权限(进入容器等)',
  `expire_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '过期时间(0=永不过期)',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `created_by` bigint DEFAULT 0 COMMENT '授权人ID',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_cluster_id` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户集群权限表';

-- =====================================================
-- 8. CI/CD - 流水线表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` bigint DEFAULT NULL COMMENT '所属项目ID',
  `name` varchar(191) NOT NULL COMMENT '流水线名称',
  `description` varchar(500) NOT NULL DEFAULT '' COMMENT '描述',
  `git_repo` varchar(500) NOT NULL COMMENT 'Git仓库地址',
  `git_branch` varchar(100) NOT NULL DEFAULT 'main' COMMENT 'Git分支',
  `jenkins_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Jenkins服务地址',
  `jenkins_job` varchar(191) NOT NULL COMMENT 'Jenkins Job名称',
  `jenkins_credential_id` varchar(191) NOT NULL DEFAULT '' COMMENT 'Jenkins凭证ID',
  `language_type` varchar(20) NOT NULL DEFAULT 'custom' COMMENT '语言类型:go/java/frontend/python/custom',
  -- 部署配置
  `auto_deploy` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否自动部署',
  `target_cluster_id` bigint DEFAULT NULL COMMENT '目标集群ID',
  `target_namespace` varchar(100) DEFAULT '' COMMENT '目标命名空间',
  `target_workload_kind` varchar(50) DEFAULT '' COMMENT '工作负载类型',
  `target_workload_name` varchar(200) DEFAULT '' COMMENT '工作负载名称',
  `target_container` varchar(100) DEFAULT '' COMMENT '目标容器名称',
  `deploy_env` varchar(20) DEFAULT 'dev' COMMENT '部署环境',
  `require_approval` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否需要审批',
  `enable_sonar` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用SonarQube代码扫描',
  `enable_artifact_upload` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用制品上传',
  -- 最新部署信息
  `last_deploy_image` varchar(500) DEFAULT '' COMMENT '最新部署镜像',
  `last_deploy_digest` varchar(100) DEFAULT '' COMMENT '镜像摘要',
  `last_deploy_time` bigint DEFAULT NULL COMMENT '最新部署时间',
  `last_deploy_status` varchar(32) DEFAULT '' COMMENT '最新部署状态',
  `last_deploy_version` varchar(100) DEFAULT '' COMMENT '最新部署版本',
  -- 运行状态
  `status` varchar(50) NOT NULL DEFAULT 'idle' COMMENT '状态:idle,running,disabled',
  `last_run_status` varchar(50) NOT NULL DEFAULT '' COMMENT '最后运行状态',
  `last_run_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '最后运行时间',
  `last_build_number` int NOT NULL DEFAULT 0 COMMENT '最后构建号',
  `last_build_url` varchar(500) NOT NULL DEFAULT '' COMMENT '最后构建URL',
  -- JSON配置
  `env_vars` json DEFAULT NULL COMMENT '环境变量',
  `deploy_config` json DEFAULT NULL COMMENT '部署配置',
  `stages` json DEFAULT NULL COMMENT '阶段配置',
  -- 元数据
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_jenkins_job` (`jenkins_job`),
  KEY `idx_status` (`status`),
  KEY `idx_auto_deploy` (`auto_deploy`),
  KEY `idx_target_cluster` (`target_cluster_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD流水线表';

-- =====================================================
-- 9. CI/CD - 流水线运行记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline_run` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `build_number` int NOT NULL DEFAULT 0 COMMENT '构建号',
  `status` varchar(50) NOT NULL DEFAULT 'pending' COMMENT '状态:pending,running,success,failed,aborted',
  `trigger_type` varchar(50) NOT NULL DEFAULT 'manual' COMMENT '触发类型:manual,webhook,scheduled',
  `trigger_user_id` bigint NOT NULL DEFAULT 0 COMMENT '触发人ID',
  `git_commit` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git Commit',
  `git_branch` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit_message` varchar(500) NOT NULL DEFAULT '' COMMENT '提交消息',
  `jenkins_build_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Jenkins构建URL',
  `duration_sec` int NOT NULL DEFAULT 0 COMMENT '执行时长(秒)',
  `console_log` longtext DEFAULT NULL COMMENT '控制台日志',
  `stages_result` json DEFAULT NULL COMMENT '阶段结果',
  `started_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '开始时间',
  `finished_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '结束时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `error_message` text DEFAULT NULL COMMENT '错误信息',
  `image_url` varchar(500) DEFAULT '' COMMENT '构建镜像地址',
  `image_digest` varchar(100) DEFAULT '' COMMENT '镜像摘要',
  `callback_received` tinyint(1) DEFAULT 0 COMMENT '是否收到回调',
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_build_number` (`build_number`),
  KEY `idx_status` (`status`),
  KEY `idx_started_at` (`started_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线运行记录表';

-- =====================================================
-- 10. CI/CD - 流水线阶段执行记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline_stage` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `run_id` bigint NOT NULL COMMENT '运行记录ID',
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `stage_order` int NOT NULL DEFAULT 0 COMMENT '阶段顺序',
  `stage_type` varchar(32) NOT NULL COMMENT '阶段类型:checkout,dependencies,compile,test,lint,sonar,quality_gate,build_binary,upload_artifact,build,push,approval,deploy',
  `stage_name` varchar(100) NOT NULL COMMENT '阶段名称',
  `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '状态',
  `started_at` bigint DEFAULT NULL COMMENT '开始时间',
  `finished_at` bigint DEFAULT NULL COMMENT '结束时间',
  `duration_sec` int DEFAULT 0 COMMENT '执行时长',
  `logs` longtext DEFAULT NULL COMMENT '阶段日志',
  `jenkins_stage_id` varchar(100) DEFAULT NULL COMMENT 'Jenkins阶段ID',
  -- 审批信息
  `approval_user_id` bigint DEFAULT NULL COMMENT '审批人ID',
  `approval_comment` text DEFAULT NULL COMMENT '审批评论',
  `approval_decision` varchar(32) DEFAULT NULL COMMENT '审批决定:approved,rejected',
  -- 部署信息
  `deploy_cluster_id` bigint DEFAULT NULL COMMENT '部署集群ID',
  `deploy_namespace` varchar(100) DEFAULT NULL COMMENT '部署命名空间',
  `deploy_workload_kind` varchar(50) DEFAULT NULL COMMENT '工作负载类型',
  `deploy_workload_name` varchar(100) DEFAULT NULL COMMENT '工作负载名称',
  `deploy_container` varchar(100) DEFAULT NULL COMMENT '容器名称',
  `deploy_image` varchar(500) DEFAULT NULL COMMENT '部署镜像',
  `deploy_old_image` varchar(500) DEFAULT NULL COMMENT '旧镜像',
  `deploy_replicas` int DEFAULT NULL COMMENT '副本数',
  `error_message` text DEFAULT NULL COMMENT '错误信息',
  `created_at` bigint NOT NULL,
  `modified_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_run_id` (`run_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_stage_type` (`stage_type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线阶段执行记录表';

-- =====================================================
-- 11. CI/CD - 环境配置表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_environment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '环境标识:dev,staging,prod',
  `display_name` varchar(100) NOT NULL COMMENT '显示名称',
  `description` varchar(500) NOT NULL DEFAULT '' COMMENT '描述',
  `cluster_id` bigint NOT NULL COMMENT '关联集群ID',
  `namespace` varchar(100) NOT NULL DEFAULT '' COMMENT '默认命名空间',
  `color` varchar(20) NOT NULL DEFAULT '#1890ff' COMMENT '环境颜色',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序',
  `require_approval` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否需要审批',
  `approval_users` json DEFAULT NULL COMMENT '审批人列表',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD环境配置表';

-- =====================================================
-- 12. CI/CD - 审批记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_approval` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `pipeline_run_id` bigint NOT NULL COMMENT '运行记录ID',
  `release_id` bigint NOT NULL DEFAULT 0 COMMENT '发布单ID',
  `env_name` varchar(50) NOT NULL COMMENT '目标环境',
  `status` varchar(20) NOT NULL DEFAULT 'pending' COMMENT '状态:pending,approved,rejected,expired',
  `image` varchar(500) NOT NULL DEFAULT '' COMMENT '待部署镜像',
  `image_digest` varchar(100) NOT NULL DEFAULT '' COMMENT '镜像摘要',
  `request_user_id` bigint NOT NULL COMMENT '申请人ID',
  `request_reason` varchar(500) NOT NULL DEFAULT '' COMMENT '申请原因',
  `approve_user_id` bigint NOT NULL DEFAULT 0 COMMENT '审批人ID',
  `approve_reason` varchar(500) NOT NULL DEFAULT '' COMMENT '审批意见',
  `approve_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '审批时间',
  `expire_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '过期时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_pipeline_run_id` (`pipeline_run_id`),
  KEY `idx_status` (`status`),
  KEY `idx_request_user_id` (`request_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD审批记录表';

-- =====================================================
-- 13. CI/CD - 发布单表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_release` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_name` varchar(191) NOT NULL DEFAULT '' COMMENT '应用名称',
  `namespace` varchar(191) NOT NULL DEFAULT 'default' COMMENT '命名空间',
  `workload_kind` varchar(32) NOT NULL DEFAULT 'Deployment' COMMENT '工作负载类型',
  `workload_name` varchar(191) NOT NULL DEFAULT '' COMMENT '工作负载名称',
  `container_name` varchar(191) NOT NULL DEFAULT '' COMMENT '容器名称',
  `strategy` varchar(32) NOT NULL DEFAULT 'rolling' COMMENT '发布策略',
  `timeout_sec` int unsigned NOT NULL DEFAULT 300 COMMENT '超时时间(秒)',
  `concurrency` int unsigned NOT NULL DEFAULT 3 COMMENT '并发数',
  `status` varchar(32) NOT NULL DEFAULT 'Pending' COMMENT '状态',
  `message` varchar(1024) NOT NULL DEFAULT '' COMMENT '消息',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `request_id` varchar(64) NOT NULL DEFAULT '' COMMENT '请求ID',
  `build_id` bigint NOT NULL DEFAULT 0 COMMENT '关联构建ID',
  `image_repo` varchar(512) NOT NULL DEFAULT '' COMMENT '镜像仓库',
  `image_tag` varchar(191) NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(255) DEFAULT NULL COMMENT '镜像摘要',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_request_id` (`request_id`),
  KEY `idx_app_name` (`app_name`),
  KEY `idx_status` (`status`),
  KEY `idx_build_id` (`build_id`),
  KEY `idx_modified_at` (`modified_at`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD发布单表';

-- =====================================================
-- 14. CI/CD - 发布阶段表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_release_stage` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `release_id` bigint NOT NULL DEFAULT 0 COMMENT '发布单ID',
  `stage_name` varchar(64) NOT NULL DEFAULT '' COMMENT '阶段名称',
  `stage_order` int NOT NULL DEFAULT 0 COMMENT '阶段顺序',
  `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '状态',
  `message` varchar(1024) NOT NULL DEFAULT '' COMMENT '消息',
  `logs` text DEFAULT NULL COMMENT '日志',
  `start_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '开始时间',
  `end_time` bigint unsigned NOT NULL DEFAULT 0 COMMENT '结束时间',
  `duration` bigint NOT NULL DEFAULT 0 COMMENT '持续时间',
  `build_number` varchar(64) NOT NULL DEFAULT '' COMMENT '构建号',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_release_id` (`release_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD发布阶段表';

-- =====================================================
-- 15. CI/CD - 发布任务表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_release_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `release_id` bigint NOT NULL COMMENT '发布单ID',
  `cluster_id` bigint NOT NULL COMMENT '集群ID',
  `status` varchar(32) NOT NULL DEFAULT 'Pending' COMMENT '状态',
  `message` varchar(2048) NOT NULL DEFAULT '' COMMENT '消息',
  `prev_image` varchar(512) NOT NULL DEFAULT '' COMMENT '原镜像',
  `target_image` varchar(512) NOT NULL DEFAULT '' COMMENT '目标镜像',
  `started_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '开始时间',
  `finished_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '结束时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_release_id` (`release_id`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD发布任务表';

-- =====================================================
-- 16. CI/CD - 构建记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_build` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_name` varchar(191) NOT NULL DEFAULT '' COMMENT '应用名称',
  `git_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Git URL',
  `git_branch` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git Commit',
  `jenkins_job` varchar(191) NOT NULL DEFAULT '' COMMENT 'Jenkins Job',
  `jenkins_queue_id` bigint NOT NULL DEFAULT 0 COMMENT 'Jenkins队列ID',
  `jenkins_build_number` int NOT NULL DEFAULT 0 COMMENT 'Jenkins构建号',
  `jenkins_build_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Jenkins构建URL',
  `status` varchar(50) NOT NULL DEFAULT 'pending' COMMENT '状态',
  `message` varchar(1024) NOT NULL DEFAULT '' COMMENT '消息',
  `image_repo` varchar(500) NOT NULL DEFAULT '' COMMENT '镜像仓库',
  `image_tag` varchar(191) NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(191) DEFAULT NULL COMMENT '镜像摘要',
  `sbom_ref` varchar(500) NOT NULL DEFAULT '' COMMENT 'SBOM引用',
  `sign_ref` varchar(500) NOT NULL DEFAULT '' COMMENT '签名引用',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_app_name` (`app_name`),
  KEY `idx_jenkins_job` (`jenkins_job`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD构建记录表';

-- =====================================================
-- 17. CI/CD - 流水线模板表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_pipeline_template` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(100) NOT NULL COMMENT '模板名称',
  `description` varchar(500) DEFAULT '' COMMENT '模板描述',
  `type` varchar(50) NOT NULL DEFAULT 'custom' COMMENT '模板类型: frontend/backend/microservice/database/custom',
  `stages` json DEFAULT NULL COMMENT '阶段配置',
  `default_env_vars` json DEFAULT NULL COMMENT '默认环境变量',
  `deploy_config` json DEFAULT NULL COMMENT '默认部署配置',
  `jenkins_template` text COMMENT 'Jenkinsfile模板',
  `usage_count` bigint NOT NULL DEFAULT 0 COMMENT '使用次数',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线模板表';

-- =====================================================
-- 18. 镜像仓库配置表
-- =====================================================
CREATE TABLE IF NOT EXISTS `image_registry` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '仓库名称',
  `type` varchar(50) NOT NULL DEFAULT 'docker' COMMENT '类型:docker,harbor,acr,ecr,gcr,quay',
  `url` varchar(500) NOT NULL COMMENT '仓库地址',
  `username` varchar(100) DEFAULT '' COMMENT '用户名',
  `password` varchar(500) DEFAULT '' COMMENT '密码(加密)',
  `access_key_id` varchar(100) DEFAULT '' COMMENT 'AccessKey ID(云厂商)',
  `access_key_secret` varchar(200) DEFAULT '' COMMENT 'AccessKey Secret(加密)',
  `region` varchar(50) DEFAULT '' COMMENT '区域',
  `insecure` tinyint(1) DEFAULT 0 COMMENT '跳过TLS验证',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `is_default` tinyint(1) DEFAULT 0 COMMENT '是否默认仓库',
  `status` varchar(50) DEFAULT 'unknown' COMMENT '连接状态',
  `last_check_at` bigint DEFAULT 0 COMMENT '最后检测时间',
  `last_error` varchar(500) DEFAULT '' COMMENT '最后错误',
  `created_by` bigint DEFAULT 0 COMMENT '创建人',
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_registry_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='镜像仓库配置表';

-- =====================================================
-- 19. 镜像清理策略表
-- =====================================================
CREATE TABLE IF NOT EXISTS `image_cleanup_policy` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `registry_id` bigint NOT NULL COMMENT '仓库ID',
  `name` varchar(100) NOT NULL COMMENT '策略名称',
  `enabled` tinyint(1) DEFAULT 1 COMMENT '是否启用',
  `repository_pattern` varchar(200) DEFAULT '*' COMMENT '仓库匹配模式',
  `tag_pattern` varchar(200) DEFAULT '*' COMMENT '标签匹配模式',
  `keep_last_count` int DEFAULT 5 COMMENT '保留最近N个',
  `keep_days` int DEFAULT 30 COMMENT '保留N天内',
  `cron_expression` varchar(50) DEFAULT '0 2 * * *' COMMENT 'Cron表达式',
  `last_run_at` bigint DEFAULT 0 COMMENT '最后执行时间',
  `last_run_result` varchar(500) DEFAULT '' COMMENT '最后执行结果',
  `deleted_count` bigint DEFAULT 0 COMMENT '累计删除数',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `created_by` bigint DEFAULT 0 COMMENT '创建人',
  `created_at` bigint DEFAULT 0,
  `modified_at` bigint DEFAULT 0,
  `is_del` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_registry_id` (`registry_id`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='镜像清理策略表';

-- =====================================================
-- 20. 镜像清理日志表
-- =====================================================
CREATE TABLE IF NOT EXISTS `image_cleanup_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `policy_id` bigint NOT NULL COMMENT '策略ID',
  `registry_id` bigint NOT NULL COMMENT '仓库ID',
  `start_time` bigint NOT NULL COMMENT '开始时间',
  `end_time` bigint DEFAULT 0 COMMENT '结束时间',
  `status` varchar(20) DEFAULT 'running' COMMENT '状态',
  `scanned_count` int DEFAULT 0 COMMENT '扫描数',
  `deleted_count` int DEFAULT 0 COMMENT '删除数',
  `freed_size` bigint DEFAULT 0 COMMENT '释放空间(字节)',
  `error_message` text DEFAULT NULL COMMENT '错误信息',
  `details` json DEFAULT NULL COMMENT '详情',
  PRIMARY KEY (`id`),
  KEY `idx_policy_id` (`policy_id`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='镜像清理日志表';

-- =====================================================
-- 21. 平台设置表
-- =====================================================
CREATE TABLE IF NOT EXISTS `platform_settings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `category` varchar(50) NOT NULL COMMENT '分类:basic,security,alert,notification',
  `key` varchar(100) NOT NULL COMMENT '设置键',
  `value` text DEFAULT NULL COMMENT '设置值',
  `value_type` varchar(20) DEFAULT 'string' COMMENT '值类型:string,int,bool,json',
  `label` varchar(100) DEFAULT NULL COMMENT '显示名称',
  `desc` varchar(500) DEFAULT NULL COMMENT '描述',
  `created_at` int unsigned DEFAULT NULL,
  `modified_at` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_category_key` (`category`, `key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台设置表';

-- =====================================================
-- 22. IAM - 项目表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_project` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '项目ID',
  `name` varchar(100) NOT NULL COMMENT '项目名称（唯一标识）',
  `display_name` varchar(191) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `status` varchar(50) NOT NULL DEFAULT 'active' COMMENT '状态: active/archived/disabled',
  `owner_id` bigint NOT NULL DEFAULT 0 COMMENT '项目负责人ID',
  `default_cluster_id` bigint DEFAULT NULL COMMENT '默认集群ID',
  `default_namespace` varchar(100) DEFAULT '' COMMENT '默认命名空间',
  `allowed_clusters` json DEFAULT NULL COMMENT '允许的集群ID列表',
  `allowed_namespaces` json DEFAULT NULL COMMENT '允许的命名空间列表（支持通配符）',
  `labels` json DEFAULT NULL COMMENT '标签（键值对）',
  `created_by` bigint NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '修改时间',
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  `is_del` tinyint unsigned NOT NULL DEFAULT 0 COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_owner_id` (`owner_id`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';

-- =====================================================
-- 23. IAM - 项目成员关系表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_project_member` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `subject_type` varchar(50) NOT NULL COMMENT '主体类型: user/group',
  `subject_id` bigint NOT NULL COMMENT '主体ID（用户ID或组ID）',
  `role` varchar(50) NOT NULL DEFAULT 'viewer' COMMENT '项目角色: owner/admin/developer/viewer',
  `joined_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '加入时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_subject` (`project_id`, `subject_type`, `subject_id`),
  KEY `idx_subject` (`subject_type`, `subject_id`),
  KEY `idx_project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目成员关系表';

-- =====================================================
-- 24. IAM - 用户组表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_group` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户组ID',
  `name` varchar(100) NOT NULL COMMENT '组名称（唯一标识）',
  `display_name` varchar(191) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `type` varchar(50) NOT NULL DEFAULT 'custom' COMMENT '类型: system/custom',
  `parent_id` bigint DEFAULT NULL COMMENT '父组ID（支持层级结构）',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序顺序',
  `created_by` bigint NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '修改时间',
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  `is_del` tinyint unsigned NOT NULL DEFAULT 0 COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_type` (`type`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组表';

-- =====================================================
-- 25. IAM - 用户组成员关系表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_group_user` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `group_id` bigint NOT NULL COMMENT '用户组ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role` varchar(50) NOT NULL DEFAULT 'member' COMMENT '组内角色: owner/admin/member',
  `joined_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '加入时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组成员关系表';

-- =====================================================
-- 26. IAM - 权限模板表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_role_template` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '模板ID',
  `name` varchar(100) NOT NULL COMMENT '模板名称（唯一标识）',
  `display_name` varchar(191) NOT NULL COMMENT '显示名称',
  `description` varchar(500) DEFAULT '' COMMENT '描述',
  `type` varchar(50) NOT NULL COMMENT '模板类型: k8s/cicd/platform',
  `builtin` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否内置模板',
  `k8s_rules` json DEFAULT NULL COMMENT 'K8s RBAC规则 [{apiGroups, resources, verbs}]',
  `k8s_cluster_scope` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否集群级别权限',
  `cicd_actions` json DEFAULT NULL COMMENT 'CICD操作权限 ["view","run","approve","deploy","rollback","delete"]',
  `platform_permissions` json DEFAULT NULL COMMENT '平台功能权限 ["cluster:manage","user:manage","audit:view"]',
  `sort_order` int NOT NULL DEFAULT 0 COMMENT '排序顺序',
  `created_by` bigint NOT NULL DEFAULT 0 COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '修改时间',
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '删除时间',
  `is_del` tinyint unsigned NOT NULL DEFAULT 0 COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_builtin` (`builtin`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限模板表';

-- =====================================================
-- 27. IAM - 授权记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_grant` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '授权ID',
  `subject_type` varchar(50) NOT NULL COMMENT '主体类型: user/group/project',
  `subject_id` bigint NOT NULL COMMENT '主体ID',
  `subject_name` varchar(191) DEFAULT '' COMMENT '主体名称（冗余）',
  `scope_type` varchar(50) NOT NULL COMMENT '范围类型: cluster/namespace/cicd_project/cicd_pipeline',
  `scope_id` bigint DEFAULT NULL COMMENT '范围ID（集群ID/项目ID/流水线ID）',
  `scope_name` varchar(191) DEFAULT '' COMMENT '范围名称（冗余）',
  `namespaces` json DEFAULT NULL COMMENT '命名空间列表（支持通配符如 ["default","app-*"]）',
  `role_template_id` bigint NOT NULL COMMENT '权限模板ID',
  `role_template_name` varchar(100) DEFAULT '' COMMENT '模板名称（冗余）',
  `expire_at` bigint unsigned DEFAULT NULL COMMENT '过期时间（NULL 表示永不过期）',
  `k8s_synced` tinyint(1) NOT NULL DEFAULT 0 COMMENT 'K8s RBAC 是否已同步',
  `k8s_role_name` varchar(191) DEFAULT '' COMMENT 'K8s Role/ClusterRole 名称',
  `k8s_binding_name` varchar(191) DEFAULT '' COMMENT 'K8s RoleBinding/ClusterRoleBinding 名称',
  `k8s_sync_error` varchar(500) DEFAULT '' COMMENT 'K8s 同步错误信息',
  `k8s_synced_at` bigint unsigned DEFAULT NULL COMMENT 'K8s 同步时间',
  `status` varchar(50) NOT NULL DEFAULT 'active' COMMENT '状态: active/expired/revoked',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  `granted_by` bigint NOT NULL DEFAULT 0 COMMENT '授权人ID',
  `revoked_by` bigint DEFAULT NULL COMMENT '撤销人ID',
  `revoked_at` bigint unsigned DEFAULT NULL COMMENT '撤销时间',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_subject` (`subject_type`, `subject_id`),
  KEY `idx_scope` (`scope_type`, `scope_id`),
  KEY `idx_role_template` (`role_template_id`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_granted_by` (`granted_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='授权记录表';

-- =====================================================
-- 28. IAM - 环境权限绑定表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_env_binding` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `subject_type` varchar(50) NOT NULL COMMENT '主体类型: user/group',
  `subject_id` bigint NOT NULL COMMENT '主体ID',
  `subject_name` varchar(191) DEFAULT '' COMMENT '主体名称',
  `cluster_id` bigint NOT NULL COMMENT '集群ID',
  `cluster_name` varchar(191) DEFAULT '' COMMENT '集群名称',
  `env_type` varchar(50) DEFAULT '' COMMENT '环境类型',
  `namespaces` json DEFAULT NULL COMMENT '命名空间列表',
  `env_role` varchar(50) NOT NULL COMMENT '环境角色',
  `custom_actions` json DEFAULT NULL COMMENT '自定义操作权限',
  `max_env_level` int DEFAULT 1 COMMENT '最高环境级别',
  `bypass_approval` tinyint(1) DEFAULT 0 COMMENT '是否跳过审批',
  `k8s_synced` tinyint(1) DEFAULT 0 COMMENT 'K8s RBAC是否已同步',
  `k8s_role_name` varchar(191) DEFAULT '' COMMENT 'K8s Role名称',
  `k8s_binding_name` varchar(191) DEFAULT '' COMMENT 'K8s RoleBinding名称',
  `k8s_sync_error` varchar(500) DEFAULT '' COMMENT 'K8s同步错误',
  `k8s_synced_at` bigint unsigned DEFAULT NULL COMMENT 'K8s同步时间',
  `expire_at` bigint unsigned DEFAULT NULL COMMENT '过期时间',
  `status` varchar(50) DEFAULT 'active' COMMENT '状态: active/expired/revoked',
  `remark` varchar(500) DEFAULT '' COMMENT '备注',
  `granted_by` bigint DEFAULT 0 COMMENT '授权人ID',
  `revoked_by` bigint DEFAULT NULL COMMENT '撤销人ID',
  `revoked_at` bigint unsigned DEFAULT NULL COMMENT '撤销时间',
  `created_at` bigint unsigned NOT NULL COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_subject` (`subject_type`, `subject_id`),
  KEY `idx_cluster` (`cluster_id`),
  KEY `idx_env_type` (`env_type`),
  KEY `idx_env_role` (`env_role`),
  KEY `idx_status` (`status`),
  KEY `idx_k8s_synced` (`k8s_synced`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='环境权限绑定表';

-- =====================================================
-- 29. IAM - 环境操作审计日志表
-- =====================================================
CREATE TABLE IF NOT EXISTS `iam_env_audit_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '操作用户ID',
  `username` varchar(191) DEFAULT '' COMMENT '用户名',
  `action` varchar(50) NOT NULL COMMENT '操作类型',
  `resource_type` varchar(50) NOT NULL COMMENT '资源类型',
  `resource_name` varchar(191) DEFAULT '' COMMENT '资源名称',
  `cluster_id` bigint DEFAULT NULL COMMENT '集群ID',
  `cluster_name` varchar(191) DEFAULT '' COMMENT '集群名称',
  `env_type` varchar(50) DEFAULT '' COMMENT '环境类型',
  `namespace` varchar(191) DEFAULT '' COMMENT '命名空间',
  `success` tinyint(1) DEFAULT 1 COMMENT '是否成功',
  `error_message` varchar(500) DEFAULT '' COMMENT '错误信息',
  `client_ip` varchar(50) DEFAULT '' COMMENT '客户端IP',
  `user_agent` varchar(500) DEFAULT '' COMMENT 'User-Agent',
  `request_id` varchar(64) DEFAULT '' COMMENT '请求ID',
  `detail` json DEFAULT NULL COMMENT '详情',
  `created_at` bigint unsigned NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_env_type` (`env_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='环境操作审计日志表';

-- =====================================================
-- 30. 审计日志表
-- =====================================================
CREATE TABLE IF NOT EXISTS `audit_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint NOT NULL COMMENT '操作用户ID',
  `username` varchar(191) NOT NULL COMMENT '操作用户名',
  `user_ip` varchar(50) DEFAULT '' COMMENT '用户IP',
  `user_agent` varchar(500) DEFAULT '' COMMENT 'User-Agent',
  `action` varchar(100) NOT NULL COMMENT '操作类型',
  `action_display` varchar(191) DEFAULT '' COMMENT '操作显示名称',
  `module` varchar(100) NOT NULL COMMENT '模块',
  `target_type` varchar(100) DEFAULT '' COMMENT '目标类型',
  `target_id` varchar(100) DEFAULT '' COMMENT '目标ID',
  `target_name` varchar(191) DEFAULT '' COMMENT '目标名称',
  `request_uri` varchar(500) DEFAULT '' COMMENT '请求URI',
  `request_method` varchar(10) DEFAULT '' COMMENT '请求方法',
  `request_body` text COMMENT '请求体',
  `response_code` int DEFAULT NULL COMMENT '响应状态码',
  `response_message` varchar(500) DEFAULT '' COMMENT '响应消息',
  `detail` json DEFAULT NULL COMMENT '操作详情',
  `extra` json DEFAULT NULL COMMENT '额外信息',
  `cluster_id` bigint DEFAULT NULL COMMENT '关联集群ID',
  `cluster_name` varchar(191) DEFAULT '' COMMENT '关联集群名称',
  `namespace` varchar(100) DEFAULT '' COMMENT '关联命名空间',
  `pipeline_id` bigint DEFAULT NULL COMMENT '关联流水线ID',
  `pipeline_name` varchar(191) DEFAULT '' COMMENT '关联流水线名称',
  `run_id` bigint DEFAULT NULL COMMENT '关联运行记录ID',
  `project_id` bigint DEFAULT NULL COMMENT '关联项目ID',
  `project_name` varchar(191) DEFAULT '' COMMENT '关联项目名称',
  `status` varchar(50) NOT NULL DEFAULT 'success' COMMENT '操作状态: success/failed',
  `error_message` varchar(1000) DEFAULT '' COMMENT '错误信息',
  `duration_ms` int DEFAULT 0 COMMENT '操作耗时(ms)',
  `created_at` bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_module` (`module`),
  KEY `idx_target` (`target_type`, `target_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';

-- =====================================================
-- 31. 视图 - 用户权限视图
-- =====================================================
CREATE OR REPLACE VIEW `v_user_permissions` AS
SELECT 
  g.user_id, g.subject_type, g.subject_id,
  g.scope_type, g.scope_id, g.scope_name, g.namespaces,
  g.role_template_id, rt.name AS role_template_name, rt.type AS role_type,
  rt.k8s_rules, rt.cicd_actions, rt.platform_permissions,
  g.expire_at, g.status
FROM (
  SELECT subject_id AS user_id, subject_type, subject_id, scope_type, scope_id, scope_name,
         namespaces, role_template_id, expire_at, status
  FROM iam_grant ig
  WHERE ig.subject_type = 'user' AND ig.status = 'active' 
    AND (ig.expire_at IS NULL OR ig.expire_at > UNIX_TIMESTAMP())
  UNION ALL
  SELECT igu.user_id, ig.subject_type, ig.subject_id, ig.scope_type, ig.scope_id, ig.scope_name,
         ig.namespaces, ig.role_template_id, ig.expire_at, ig.status
  FROM iam_grant ig
  JOIN iam_group_user igu ON ig.subject_id = igu.group_id
  WHERE ig.subject_type = 'group' AND ig.status = 'active'
    AND (ig.expire_at IS NULL OR ig.expire_at > UNIX_TIMESTAMP())
) g
LEFT JOIN iam_role_template rt ON g.role_template_id = rt.id AND rt.is_del = 0;

-- =====================================================
-- 32. 视图 - 用户环境权限视图
-- =====================================================
CREATE OR REPLACE VIEW `v_user_env_permissions` AS
SELECT 
  eb.subject_id AS user_id, eb.subject_name AS username,
  eb.cluster_id, eb.cluster_name, kc.env_type AS cluster_env_type, kc.env_level AS cluster_env_level,
  kc.access_mode, eb.env_role, eb.max_env_level, eb.bypass_approval,
  eb.namespaces, eb.status, eb.expire_at, eb.k8s_synced
FROM iam_env_binding eb
LEFT JOIN kube_cluster kc ON eb.cluster_id = kc.id
WHERE eb.subject_type = 'user' AND eb.status = 'active'
  AND (eb.expire_at IS NULL OR eb.expire_at > UNIX_TIMESTAMP())
  AND kc.is_del = 0;

-- =====================================================
-- 添加缺少的字段 (兼容已有数据库)
-- 注意: 如果字段已存在会报错，可忽略
-- =====================================================

-- image_registry 添加云厂商字段 (新安装无需执行，升级时按需执行)
-- 使用存储过程安全添加字段
DROP PROCEDURE IF EXISTS add_column_if_not_exists;
DELIMITER //
CREATE PROCEDURE add_column_if_not_exists()
BEGIN
    IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='k8s-platform' AND TABLE_NAME='image_registry' AND COLUMN_NAME='access_key_id') THEN
        ALTER TABLE `image_registry` ADD COLUMN `access_key_id` varchar(100) DEFAULT '' COMMENT 'AccessKey ID' AFTER `password`;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='k8s-platform' AND TABLE_NAME='image_registry' AND COLUMN_NAME='access_key_secret') THEN
        ALTER TABLE `image_registry` ADD COLUMN `access_key_secret` varchar(200) DEFAULT '' COMMENT 'AccessKey Secret' AFTER `access_key_id`;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='k8s-platform' AND TABLE_NAME='image_registry' AND COLUMN_NAME='region') THEN
        ALTER TABLE `image_registry` ADD COLUMN `region` varchar(50) DEFAULT '' COMMENT '区域' AFTER `access_key_secret`;
    END IF;
END //
DELIMITER ;
CALL add_column_if_not_exists();
DROP PROCEDURE IF EXISTS add_column_if_not_exists;

-- =====================================================
-- 初始化 RBAC 角色数据
-- =====================================================
INSERT IGNORE INTO `sys_role` (`id`, `name`, `display_name`, `description`, `role_type`, `is_system`, `color`, `icon`, `sort_order`, `created_at`, `modified_at`) VALUES
(1, 'super_admin', '超级管理员', '拥有系统所有权限，可管理所有集群和用户', 'super_admin', 1, '#f5222d', 'crown', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2, 'cluster_admin', '集群管理员', '可管理指定集群的所有资源', 'cluster_admin', 1, '#fa8c16', 'cluster', 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(3, 'developer', '开发者', '可查看和操作指定命名空间的资源', 'developer', 1, '#1890ff', 'code', 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(4, 'viewer', '只读用户', '只能查看指定资源，无法进行修改操作', 'viewer', 1, '#52c41a', 'eye', 4, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 初始化系统权限数据
-- =====================================================
INSERT IGNORE INTO `sys_permission` (`id`, `name`, `display_name`, `description`, `resource_type`, `action`, `parent_id`, `path`, `sort_order`, `created_at`, `modified_at`) VALUES
-- 集群权限
(1, 'cluster:view', '查看集群', '查看集群列表和详情', 'cluster', 'view', 0, '/cluster', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(2, 'cluster:create', '创建集群', '添加新的K8s集群', 'cluster', 'create', 0, '/cluster', 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(3, 'cluster:update', '更新集群', '修改集群配置', 'cluster', 'update', 0, '/cluster', 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(4, 'cluster:delete', '删除集群', '删除K8s集群', 'cluster', 'delete', 0, '/cluster', 4, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 命名空间权限
(5, 'namespace:view', '查看命名空间', '查看命名空间列表', 'namespace', 'view', 0, '/namespace', 5, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(6, 'namespace:create', '创建命名空间', '创建新命名空间', 'namespace', 'create', 0, '/namespace', 6, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(7, 'namespace:delete', '删除命名空间', '删除命名空间', 'namespace', 'delete', 0, '/namespace', 7, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 工作负载权限
(8, 'deployment:view', '查看Deployment', '查看Deployment列表和详情', 'deployment', 'view', 0, '/deployment', 10, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(9, 'deployment:create', '创建Deployment', '创建新的Deployment', 'deployment', 'create', 0, '/deployment', 11, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(10, 'deployment:update', '更新Deployment', '修改Deployment配置、扩缩容', 'deployment', 'update', 0, '/deployment', 12, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(11, 'deployment:delete', '删除Deployment', '删除Deployment', 'deployment', 'delete', 0, '/deployment', 13, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- Pod权限
(12, 'pod:view', '查看Pod', '查看Pod列表、日志', 'pod', 'view', 0, '/pod', 20, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(13, 'pod:delete', '删除Pod', '删除Pod', 'pod', 'delete', 0, '/pod', 21, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(14, 'pod:exec', '进入容器', '执行exec进入容器终端', 'pod', 'exec', 0, '/pod', 22, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- Service权限
(15, 'service:view', '查看Service', '查看Service列表', 'service', 'view', 0, '/service', 30, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(16, 'service:create', '创建Service', '创建新Service', 'service', 'create', 0, '/service', 31, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(17, 'service:update', '更新Service', '修改Service配置', 'service', 'update', 0, '/service', 32, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(18, 'service:delete', '删除Service', '删除Service', 'service', 'delete', 0, '/service', 33, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- ConfigMap/Secret权限
(19, 'configmap:view', '查看ConfigMap', '查看ConfigMap', 'configmap', 'view', 0, '/configmap', 40, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(20, 'configmap:manage', '管理ConfigMap', '创建、修改、删除ConfigMap', 'configmap', 'manage', 0, '/configmap', 41, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(21, 'secret:view', '查看Secret', '查看Secret', 'secret', 'view', 0, '/secret', 42, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(22, 'secret:manage', '管理Secret', '创建、修改、删除Secret', 'secret', 'manage', 0, '/secret', 43, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- CI/CD权限
(23, 'pipeline:view', '查看流水线', '查看CI/CD流水线', 'pipeline', 'view', 0, '/pipeline', 50, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(24, 'pipeline:manage', '管理流水线', '创建、修改、删除、执行流水线', 'pipeline', 'manage', 0, '/pipeline', 51, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 用户权限
(25, 'user:view', '查看用户', '查看用户列表', 'user', 'view', 0, '/user', 60, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
(26, 'user:manage', '管理用户', '创建、修改、删除用户，分配角色', 'user', 'manage', 0, '/user', 61, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 初始化角色权限关联 (超级管理员拥有所有权限)
-- =====================================================
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 1, id, UNIX_TIMESTAMP() FROM `sys_permission`;

-- 集群管理员权限
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 2, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE id BETWEEN 1 AND 24;

-- 开发者权限
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 3, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `action` IN ('view', 'create', 'update', 'exec') AND id NOT IN (25, 26);

-- 只读用户权限
INSERT IGNORE INTO `sys_role_permission` (`role_id`, `permission_id`, `created_at`)
SELECT 4, id, UNIX_TIMESTAMP() FROM `sys_permission` WHERE `action` = 'view';

-- =====================================================
-- 初始化管理员账户 (admin/admin123)
-- =====================================================
INSERT IGNORE INTO `user` (`id`, `username`, `password`, `role`, `status`, `created_at`, `modified_at`) VALUES
(1, 'admin', '$2a$10$jWcwxJ.3qLlHaXVZ1nL7MeCsSEXGosmaj1dIFoS74WXq5.gJrfChO', 'admin', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 关联admin用户为超级管理员
INSERT IGNORE INTO `sys_user_role` (`user_id`, `role_id`, `created_at`, `created_by`) VALUES
(1, 1, UNIX_TIMESTAMP(), 0);

-- =====================================================
-- 初始化平台设置
-- =====================================================
INSERT IGNORE INTO `platform_settings` (`category`, `key`, `value`, `value_type`, `label`, `desc`, `created_at`, `modified_at`) VALUES
-- 基本设置
('basic', 'default_page', '/clusters', 'string', '默认首页', '登录后默认跳转页面', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('basic', 'default_cluster', 'auto', 'string', '默认集群', '自动选择或指定集群ID', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('basic', 'language', 'zh-CN', 'string', '系统语言', '界面显示语言', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('basic', 'timezone', 'Asia/Shanghai', 'string', '时区', '系统时区设置', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 安全设置
('security', 'session_timeout', '120', 'int', '会话超时', '会话超时时间(分钟)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('security', 'enable_2fa', 'false', 'bool', '双因素认证', '是否启用双因素认证', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('security', 'password_policy', 'medium', 'string', '密码策略', '密码复杂度要求', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('security', 'audit_retention', '30', 'int', '审计保留', '审计日志保留天数', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 告警设置
('alert', 'cpu_threshold', '80', 'int', 'CPU告警阈值', 'CPU使用率告警阈值(%)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('alert', 'mem_threshold', '80', 'int', '内存告警阈值', '内存使用率告警阈值(%)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('alert', 'disk_threshold', '85', 'int', '磁盘告警阈值', '磁盘使用率告警阈值(%)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('alert', 'alert_silence', '15', 'int', '告警静默', '告警静默时间(分钟)', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
-- 通知设置
('notification', 'enable_email', 'false', 'bool', '邮件通知', '是否启用邮件通知', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'smtp_server', '', 'string', 'SMTP服务器', 'SMTP服务器地址', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'enable_dingtalk', 'false', 'bool', '钉钉通知', '是否启用钉钉通知', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'dingtalk_webhook', '', 'string', '钉钉Webhook', '钉钉机器人Webhook地址', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'enable_webhook', 'false', 'bool', 'Webhook通知', '是否启用自定义Webhook', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('notification', 'webhook_url', '', 'string', 'Webhook地址', '自定义Webhook地址', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 示例镜像仓库配置
-- =====================================================
INSERT IGNORE INTO `image_registry` (`name`, `type`, `url`, `description`, `is_default`, `status`, `created_at`, `modified_at`) VALUES
('Docker Hub', 'docker', 'https://registry-1.docker.io', 'Docker Hub 官方仓库', 1, 'unknown', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 初始化流水线模板数据
-- =====================================================
INSERT IGNORE INTO `cicd_pipeline_template` (`id`, `name`, `description`, `type`, `stages`, `default_env_vars`, `deploy_config`, `created_at`, `modified_at`) VALUES
(
  1,
  'Vue3 前端应用模板',
  '适用于 Vue3 + Vite 前端项目的标准流水线模板',
  'frontend',
  '[{"name": "checkout", "description": "拉取代码", "order": 1}, {"name": "install", "description": "安装依赖 (npm install)", "order": 2}, {"name": "build", "description": "构建应用 (npm run build)", "order": 3}, {"name": "test", "description": "运行测试 (npm run test)", "order": 4}, {"name": "build-image", "description": "构建 Docker 镜像", "order": 5}, {"name": "deploy", "description": "部署到 K8s", "order": 6}]',
  '[{"name": "NODE_ENV", "value": "production"}, {"name": "VITE_API_BASE", "value": "https://api.example.com"}]',
  '{"replicas": 3, "strategy": "rollingUpdate", "resources": {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "200m", "memory": "256Mi"}}}',
  UNIX_TIMESTAMP(),
  UNIX_TIMESTAMP()
),
(
  2,
  'Go 微服务模板',
  '适用于 Go 语言微服务的标准流水线模板',
  'backend',
  '[{"name": "checkout", "description": "拉取代码", "order": 1}, {"name": "test", "description": "运行单元测试 (go test)", "order": 2}, {"name": "build", "description": "编译 Go 二进制", "order": 3}, {"name": "build-image", "description": "构建 Docker 镜像", "order": 4}, {"name": "push", "description": "推送镜像到 Harbor", "order": 5}, {"name": "deploy", "description": "部署到 K8s", "order": 6}]',
  '[{"name": "GO111MODULE", "value": "on"}, {"name": "CGO_ENABLED", "value": "0"}]',
  '{"replicas": 2, "strategy": "rollingUpdate", "resources": {"limits": {"cpu": "1000m", "memory": "1024Mi"}, "requests": {"cpu": "500m", "memory": "512Mi"}}}',
  UNIX_TIMESTAMP(),
  UNIX_TIMESTAMP()
),
(
  3,
  'Java Spring Boot 模板',
  '适用于 Java Spring Boot 项目的标准流水线模板',
  'backend',
  '[{"name": "checkout", "description": "拉取代码", "order": 1}, {"name": "compile", "description": "Maven 编译 (mvn compile)", "order": 2}, {"name": "test", "description": "运行测试 (mvn test)", "order": 3}, {"name": "package", "description": "打包 (mvn package)", "order": 4}, {"name": "build-image", "description": "构建 Docker 镜像", "order": 5}, {"name": "deploy", "description": "部署到 K8s", "order": 6}]',
  '[{"name": "JAVA_HOME", "value": "/usr/lib/jvm/java-17"}, {"name": "MAVEN_OPTS", "value": "-Xmx1024m"}]',
  '{"replicas": 2, "strategy": "rollingUpdate", "resources": {"limits": {"cpu": "2000m", "memory": "2048Mi"}, "requests": {"cpu": "1000m", "memory": "1024Mi"}}}',
  UNIX_TIMESTAMP(),
  UNIX_TIMESTAMP()
),
(
  4,
  'Python Flask 模板',
  '适用于 Python Flask 项目的标准流水线模板',
  'backend',
  '[{"name": "checkout", "description": "拉取代码", "order": 1}, {"name": "install", "description": "安装依赖 (pip install)", "order": 2}, {"name": "test", "description": "运行测试 (pytest)", "order": 3}, {"name": "build-image", "description": "构建 Docker 镜像", "order": 4}, {"name": "deploy", "description": "部署到 K8s", "order": 5}]',
  '[{"name": "PYTHON_VERSION", "value": "3.11"}, {"name": "PIP_INDEX_URL", "value": "https://pypi.tuna.tsinghua.edu.cn/simple"}]',
  '{"replicas": 2, "strategy": "rollingUpdate", "resources": {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "200m", "memory": "256Mi"}}}',
  UNIX_TIMESTAMP(),
  UNIX_TIMESTAMP()
);

-- =====================================================
-- 30. CICD - 资源档位模板表
-- =====================================================
CREATE TABLE IF NOT EXISTS cicd_resource_template (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name            VARCHAR(64) NOT NULL COMMENT '模板名称：small/medium/large/custom',
    service_type    VARCHAR(32) NOT NULL COMMENT '服务类型：java/go/node/python',
    env             VARCHAR(16) NOT NULL COMMENT '环境：dev/test/staging/prod',
    
    -- 资源配置
    replicas_default INT NOT NULL DEFAULT 1 COMMENT '默认副本数',
    replicas_min     INT NOT NULL DEFAULT 1 COMMENT '最小副本数',
    replicas_max     INT NOT NULL DEFAULT 10 COMMENT '最大副本数',
    
    cpu_request      VARCHAR(16) NOT NULL DEFAULT '200m',
    cpu_limit        VARCHAR(16) NOT NULL DEFAULT '500m',
    memory_request   VARCHAR(16) NOT NULL DEFAULT '256Mi',
    memory_limit     VARCHAR(16) NOT NULL DEFAULT '512Mi',
    
    -- HPA 配置
    hpa_enabled      TINYINT(1) DEFAULT 0 COMMENT '是否启用HPA',
    hpa_min_replicas INT DEFAULT 2,
    hpa_max_replicas INT DEFAULT 10,
    hpa_cpu_target   INT DEFAULT 70 COMMENT 'CPU目标利用率%',
    
    description      VARCHAR(255) DEFAULT '' COMMENT '模板说明',
    is_default       TINYINT(1) DEFAULT 0 COMMENT '是否默认模板',
    sort_order       INT DEFAULT 0,
    created_at       BIGINT UNSIGNED DEFAULT 0,
    modified_at      BIGINT UNSIGNED DEFAULT 0,
    deleted_at       BIGINT UNSIGNED DEFAULT 0,
    
    UNIQUE KEY uk_type_env_name (service_type, env, name, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD资源档位模板表';

-- =====================================================
-- 31. CICD - 环境资源规则表
-- =====================================================
CREATE TABLE IF NOT EXISTS cicd_env_resource_rule (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    env             VARCHAR(16) NOT NULL COMMENT 'dev/test/staging/prod',
    service_type    VARCHAR(32) DEFAULT '' COMMENT '服务类型，空=通用',
    
    -- 资源上限
    cpu_limit_max       VARCHAR(16) NOT NULL DEFAULT '4' COMMENT 'CPU limit 最大值',
    memory_limit_max    VARCHAR(16) NOT NULL DEFAULT '8Gi' COMMENT '内存 limit 最大值',
    replicas_max        INT NOT NULL DEFAULT 10 COMMENT '副本数上限',
    
    -- 资源下限（生产环境）
    cpu_request_min     VARCHAR(16) DEFAULT '' COMMENT 'CPU request 最小值',
    memory_request_min  VARCHAR(16) DEFAULT '' COMMENT '内存 request 最小值',
    replicas_min        INT DEFAULT 1 COMMENT '副本数下限',
    
    -- 审批规则
    require_approval    TINYINT(1) DEFAULT 0 COMMENT '是否需要审批',
    approval_role       VARCHAR(64) DEFAULT '' COMMENT '审批角色：sre/admin',
    
    description         VARCHAR(255) DEFAULT '' COMMENT '规则说明',
    created_at          BIGINT UNSIGNED DEFAULT 0,
    modified_at         BIGINT UNSIGNED DEFAULT 0,
    
    UNIQUE KEY uk_env_type (env, service_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD环境资源规则表';

-- =====================================================
-- 32. CICD - 发布审批记录表
-- =====================================================
CREATE TABLE IF NOT EXISTS cicd_deploy_approval (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    pipeline_id     BIGINT UNSIGNED NOT NULL COMMENT '流水线ID',
    release_id      BIGINT UNSIGNED DEFAULT 0 COMMENT '发布单ID',
    env             VARCHAR(16) NOT NULL COMMENT '目标环境',
    
    -- 申请配置（JSON）
    requested_config TEXT COMMENT '申请的资源配置JSON',
    current_config   TEXT COMMENT '当前线上配置JSON',
    
    -- 风险提示
    risk_level       ENUM('low','medium','high') DEFAULT 'low',
    risk_warnings    TEXT COMMENT '风险提示列表JSON',
    
    -- 审批流程
    status           ENUM('pending','approved','rejected','expired','cancelled') DEFAULT 'pending',
    applicant_id     BIGINT UNSIGNED NOT NULL COMMENT '申请人ID',
    applicant_name   VARCHAR(64) NOT NULL,
    approver_id      BIGINT UNSIGNED DEFAULT 0 COMMENT '审批人ID',
    approver_name    VARCHAR(64) DEFAULT '',
    approve_comment  VARCHAR(500) DEFAULT '' COMMENT '审批意见',
    
    applied_at       BIGINT UNSIGNED DEFAULT 0 COMMENT '申请时间',
    approved_at      BIGINT UNSIGNED DEFAULT 0 COMMENT '审批时间',
    expired_at       BIGINT UNSIGNED DEFAULT 0 COMMENT '过期时间',
    
    INDEX idx_pipeline (pipeline_id),
    INDEX idx_status (status, applied_at),
    INDEX idx_applicant (applicant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD发布审批记录表';

-- =====================================================
-- 33. CICD - 资源配置变更日志表
-- =====================================================
CREATE TABLE IF NOT EXISTS cicd_resource_change_log (
    id              BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    pipeline_id     BIGINT UNSIGNED NOT NULL,
    env             VARCHAR(16) NOT NULL,
    
    change_type     ENUM('create','update','scale','rollback') NOT NULL,
    before_config   TEXT COMMENT '变更前配置JSON',
    after_config    TEXT COMMENT '变更后配置JSON',
    
    operator_id     BIGINT UNSIGNED NOT NULL,
    operator_name   VARCHAR(64) NOT NULL,
    reason          VARCHAR(500) DEFAULT '' COMMENT '变更原因',
    
    created_at      BIGINT UNSIGNED DEFAULT 0,
    
    INDEX idx_pipeline_env (pipeline_id, env),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD资源配置变更日志表';

-- =====================================================
-- 34. CI/CD - 制品库表
-- =====================================================
CREATE TABLE IF NOT EXISTS `cicd_artifact` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint DEFAULT NULL COMMENT '关联流水线ID',
  `run_id` bigint DEFAULT NULL COMMENT '关联运行记录ID',
  `build_number` int DEFAULT 0 COMMENT 'Jenkins构建号',

  -- 制品基本信息
  `name` varchar(200) NOT NULL DEFAULT '' COMMENT '制品名称（如 order-service-1.0.0.jar）',
  `artifact_type` varchar(20) NOT NULL DEFAULT '' COMMENT '制品类型：jar/war/binary/dist/wheel/image/archive',
  `version` varchar(100) NOT NULL DEFAULT '' COMMENT '版本号',
  `language_type` varchar(20) NOT NULL DEFAULT '' COMMENT '语言类型：go/java/frontend/python',

  -- 存储信息
  `file_path` varchar(500) NOT NULL DEFAULT '' COMMENT '文件存储路径',
  `file_size` bigint NOT NULL DEFAULT 0 COMMENT '文件大小（字节）',
  `sha256` varchar(64) NOT NULL DEFAULT '' COMMENT 'SHA256校验和',
  `storage_type` varchar(20) NOT NULL DEFAULT 'local' COMMENT '存储类型：local/s3/oss',

  -- Git 信息（构建来源追溯）
  `git_repo` varchar(500) NOT NULL DEFAULT '' COMMENT 'Git仓库地址',
  `git_branch` varchar(100) NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit` varchar(40) NOT NULL DEFAULT '' COMMENT 'Git Commit SHA',

  -- 镜像信息（如果制品已打包为镜像）
  `image_repo` varchar(500) NOT NULL DEFAULT '' COMMENT '镜像仓库地址',
  `image_tag` varchar(200) NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(100) NOT NULL DEFAULT '' COMMENT '镜像摘要',

  -- 构建元数据
  `build_duration` int NOT NULL DEFAULT 0 COMMENT '构建耗时（秒）',
  `build_log` text COMMENT '构建摘要日志',
  `metadata` json DEFAULT NULL COMMENT '扩展元数据',

  -- 状态
  `status` varchar(20) NOT NULL DEFAULT 'ready' COMMENT '状态：uploading/ready/expired/deleted',
  `download_count` int NOT NULL DEFAULT 0 COMMENT '下载次数',

  -- 元数据
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT 0,
  `modified_at` bigint unsigned NOT NULL DEFAULT 0,
  `deleted_at` bigint unsigned NOT NULL DEFAULT 0,
  `is_del` tinyint unsigned NOT NULL DEFAULT 0,

  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_run_id` (`run_id`),
  KEY `idx_artifact_type` (`artifact_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD 制品库表';

-- =====================================================
-- CICD 资源模板初始数据
-- =====================================================

-- Java 服务模板
INSERT INTO cicd_resource_template 
(name, service_type, env, replicas_default, replicas_min, replicas_max, cpu_request, cpu_limit, memory_request, memory_limit, hpa_enabled, description, is_default, sort_order, created_at, modified_at)
VALUES
('small',  'java', 'dev',  1, 1, 2,  '200m', '500m', '512Mi', '1Gi',   0, 'Java开发环境-小型，适合本地调试', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'java', 'test', 1, 1, 3,  '500m', '1',    '1Gi',   '2Gi',   0, 'Java测试环境-小型，适合功能测试', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'java', 'prod', 2, 2, 5,  '500m', '1',    '1Gi',   '2Gi',   0, 'Java生产环境-小型，适合低流量服务', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'java', 'prod', 2, 2, 10, '1',    '2',    '2Gi',   '4Gi',   1, 'Java生产环境-中型，适合中等流量服务', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('large',  'java', 'prod', 3, 2, 20, '2',    '4',    '4Gi',   '8Gi',   1, 'Java生产环境-大型，适合高流量核心服务', 0, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- Go 服务模板
INSERT INTO cicd_resource_template 
(name, service_type, env, replicas_default, replicas_min, replicas_max, cpu_request, cpu_limit, memory_request, memory_limit, hpa_enabled, description, is_default, sort_order, created_at, modified_at)
VALUES
('small',  'go', 'dev',  1, 1, 2,  '100m', '200m', '128Mi', '256Mi', 0, 'Go开发环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'go', 'test', 1, 1, 3,  '200m', '500m', '256Mi', '512Mi', 0, 'Go测试环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'go', 'prod', 2, 2, 5,  '200m', '500m', '256Mi', '512Mi', 0, 'Go生产环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'go', 'prod', 2, 2, 10, '500m', '1',    '512Mi', '1Gi',   1, 'Go生产环境-中型', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('large',  'go', 'prod', 3, 2, 20, '1',    '2',    '1Gi',   '2Gi',   1, 'Go生产环境-大型', 0, 3, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- Node 服务模板
INSERT INTO cicd_resource_template 
(name, service_type, env, replicas_default, replicas_min, replicas_max, cpu_request, cpu_limit, memory_request, memory_limit, hpa_enabled, description, is_default, sort_order, created_at, modified_at)
VALUES
('small',  'node', 'dev',  1, 1, 2,  '100m', '300m', '256Mi', '512Mi', 0, 'Node开发环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'node', 'test', 1, 1, 3,  '200m', '500m', '512Mi', '1Gi',   0, 'Node测试环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'node', 'prod', 2, 2, 5,  '200m', '500m', '512Mi', '1Gi',   0, 'Node生产环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'node', 'prod', 2, 2, 10, '500m', '1',    '1Gi',   '2Gi',   1, 'Node生产环境-中型', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- Python 服务模板
INSERT INTO cicd_resource_template 
(name, service_type, env, replicas_default, replicas_min, replicas_max, cpu_request, cpu_limit, memory_request, memory_limit, hpa_enabled, description, is_default, sort_order, created_at, modified_at)
VALUES
('small',  'python', 'dev',  1, 1, 2,  '100m', '300m', '256Mi', '512Mi', 0, 'Python开发环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'python', 'test', 1, 1, 3,  '200m', '500m', '512Mi', '1Gi',   0, 'Python测试环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('small',  'python', 'prod', 2, 2, 5,  '200m', '500m', '512Mi', '1Gi',   0, 'Python生产环境-小型', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('medium', 'python', 'prod', 2, 2, 10, '500m', '1',    '1Gi',   '2Gi',   1, 'Python生产环境-中型', 0, 2, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 环境资源规则
INSERT INTO cicd_env_resource_rule 
(env, service_type, cpu_limit_max, memory_limit_max, replicas_max, cpu_request_min, memory_request_min, replicas_min, require_approval, approval_role, description, created_at, modified_at)
VALUES
('dev', '', '1', '2Gi', 3, '', '', 1, 0, '', '开发环境通用规则，资源受限，无需审批', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('test', '', '2', '4Gi', 5, '', '', 1, 0, '', '测试环境通用规则，资源适中，无需审批', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('staging', '', '4', '8Gi', 10, '200m', '256Mi', 2, 0, '', '预发环境通用规则，接近生产配置', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('prod', '', '4', '8Gi', 20, '200m', '256Mi', 2, 1, 'sre', '生产环境通用规则，需要SRE审批', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('prod', 'java', '4', '8Gi', 20, '500m', '1Gi', 2, 1, 'sre', '生产环境Java服务规则，内存最低1Gi', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- =====================================================
-- 完成
-- =====================================================
SELECT '=====================================================';
SELECT 'K8s Platform 数据库初始化完成!';
SELECT '=====================================================';
SELECT CONCAT('表数量: ', COUNT(*)) as info FROM information_schema.tables WHERE table_schema = 'k8s-platform';
SELECT '默认管理员账户: admin / admin123';
SELECT '包含 35 张表 + 2 个视图: 用户表、集群表、RBAC表(4张)、CI/CD表(14张+制品库)、镜像表(3张)、IAM表(9张)、审计表';
SELECT '包含 4 条流水线模板: Vue3/Go/Java/Python';
SELECT '包含 CICD资源模板(18条) + 环境规则(5条)';
SELECT '=====================================================';
