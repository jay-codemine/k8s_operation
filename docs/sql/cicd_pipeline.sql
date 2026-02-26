/*
 CICD Pipeline 表结构
 用于存储流水线配置信息，支持触发 Jenkins 构建
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

USE `k8s-platform`;

-- ----------------------------
-- Table structure for cicd_pipeline
-- ----------------------------
CREATE TABLE IF NOT EXISTS `cicd_pipeline` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '流水线ID',
  `name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '流水线名称',
  `description` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '描述',
  `git_repo` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Git仓库URL',
  `git_branch` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'main' COMMENT 'Git分支',
  `jenkins_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'Jenkins服务器地址',
  `jenkins_job` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Jenkins Job名称',
  `jenkins_credential_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'Jenkins凭据ID',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'idle' COMMENT '状态(idle/running/disabled)',
  `last_run_status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '上次运行状态(success/failed/running/pending)',
  `last_run_time` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '上次运行时间(Unix时间戳)',
  `last_build_number` int NOT NULL DEFAULT 0 COMMENT '最近Jenkins构建号',
  `last_build_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '最近构建URL',
  `env_vars` json DEFAULT NULL COMMENT '环境变量JSON: [{"name":"KEY","value":"VALUE"}]',
  `deploy_config` json DEFAULT NULL COMMENT '部署配置JSON',
  `stages` json DEFAULT NULL COMMENT '流水线阶段配置JSON',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建者ID',
  `created_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(Unix时间戳)',
  `modified_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '修改时间(Unix时间戳)',
  `deleted_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(Unix时间戳)',
  `is_del` tinyint UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否删除:0未删除,1删除',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_name` (`name` ASC) USING BTREE,
  INDEX `idx_status` (`status` ASC) USING BTREE,
  INDEX `idx_jenkins_job` (`jenkins_job` ASC) USING BTREE,
  INDEX `idx_is_del` (`is_del` ASC) USING BTREE,
  INDEX `idx_created_at` (`created_at` ASC) USING BTREE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
  COMMENT='CICD流水线表'
  ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for cicd_pipeline_run
-- 记录每次流水线运行的历史
-- ----------------------------
CREATE TABLE IF NOT EXISTS `cicd_pipeline_run` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '运行记录ID',
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `build_number` int NOT NULL DEFAULT 0 COMMENT 'Jenkins构建号',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'pending' COMMENT '状态(pending/running/success/failed/aborted)',
  `trigger_type` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'manual' COMMENT '触发类型(manual/webhook/scheduled)',
  `trigger_user_id` bigint NOT NULL DEFAULT 0 COMMENT '触发用户ID',
  `git_commit` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'Git提交SHA',
  `git_branch` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'Git分支',
  `duration_sec` int NOT NULL DEFAULT 0 COMMENT '运行时长(秒)',
  `console_log` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci COMMENT '控制台日志(可选缓存)',
  `stages_result` json DEFAULT NULL COMMENT '各阶段执行结果JSON',
  `started_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '开始时间(Unix时间戳)',
  `finished_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '结束时间(Unix时间戳)',
  `created_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(Unix时间戳)',
  `modified_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '修改时间(Unix时间戳)',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_pipeline_id` (`pipeline_id` ASC) USING BTREE,
  INDEX `idx_status` (`status` ASC) USING BTREE,
  INDEX `idx_build_number` (`build_number` ASC) USING BTREE,
  INDEX `idx_started_at` (`started_at` ASC) USING BTREE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
  COMMENT='CICD流水线运行记录表'
  ROW_FORMAT=DYNAMIC;

SET FOREIGN_KEY_CHECKS = 1;

