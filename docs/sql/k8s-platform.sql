/*
 DDL only (safe)
 - Create database if not exists
 - Create tables if not exists
 - Do NOT drop tables
 - No INSERT data
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 1) database (可选：你如果不想在脚本里建库，可以删掉这一段)
CREATE DATABASE IF NOT EXISTS `k8s-platform`
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;

USE `k8s-platform`;

-- ----------------------------
-- Table structure for kube_cluster
-- ----------------------------
CREATE TABLE IF NOT EXISTS `kube_cluster` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `cluster_name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '集群名称',
  `kube_config` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'kubeconfig内容(base64编码)',
  `cluster_version` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '集群版本',
  `status` tinyint UNSIGNED NOT NULL DEFAULT 2 COMMENT '集群状态:0正常,1异常,2未检测',
  `created_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(Unix时间戳)',
  `modified_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '修改时间(Unix时间戳)',
  `deleted_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(Unix时间戳)',
  `is_del` tinyint UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否删除:0未删除,1删除',
  `last_check_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '最近一次连通性探测时间(Unix时间戳)',
  `last_error` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '最近一次异常原因',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_is_del` (`is_del` ASC) USING BTREE,
  INDEX `idx_status` (`status` ASC) USING BTREE,
  INDEX `idx_modified_at` (`modified_at` ASC) USING BTREE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
  COMMENT='K8s集群表'
  ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for user
-- ----------------------------
CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `username` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '密码',
  `created_at` int UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间',
  `modified_at` int UNSIGNED NOT NULL DEFAULT 0 COMMENT '修改时间',
  `deleted_at` int UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间',
  `is_del` tinyint UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否删除，0表示未删除，1表示删除',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_username` (`username` ASC) USING BTREE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
  COMMENT='用户表'
  ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for cicd_release
-- ----------------------------
CREATE TABLE IF NOT EXISTS `cicd_release` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '发布单ID',
  `app_name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '应用名称',
  `namespace` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'default' COMMENT '命名空间',
  `workload_kind` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Deployment' COMMENT '工作负载类型(Deployment/StatefulSet/DaemonSet)',
  `workload_name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '工作负载名称',
  `container_name` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '容器名称',
  `strategy` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'rolling' COMMENT '发布策略(rolling/recreate/canary)',
  `timeout_sec` int UNSIGNED NOT NULL DEFAULT 300 COMMENT '超时时间(秒)',
  `concurrency` int UNSIGNED NOT NULL DEFAULT 3 COMMENT '并发数',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Pending' COMMENT '状态(Pending/Queued/Running/Succeeded/Failed/Canceled/Rollback)',
  `message` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '状态消息',
  `created_user_id` bigint NOT NULL DEFAULT 0 COMMENT '创建用户ID',
  `request_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '请求ID(幂等校验)',
  `build_id` bigint NOT NULL DEFAULT 0 COMMENT 'Jenkins构建ID',
  `image_repo` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '镜像仓库地址',
  `image_tag` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL COMMENT '镜像摘要(可选)',
  `created_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(Unix时间戳)',
  `modified_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '修改时间(Unix时间戳)',
  `deleted_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(Unix时间戳)',
  `is_del` tinyint UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否删除:0未删除,1删除',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_request_id` (`request_id` ASC) USING BTREE,
  INDEX `idx_app_name` (`app_name` ASC) USING BTREE,
  INDEX `idx_status` (`status` ASC) USING BTREE,
  INDEX `idx_build_id` (`build_id` ASC) USING BTREE,
  INDEX `idx_created_at` (`created_at` ASC) USING BTREE,
  INDEX `idx_is_del` (`is_del` ASC) USING BTREE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
  COMMENT='CICD发布单表'
  ROW_FORMAT=DYNAMIC;

-- ----------------------------
-- Table structure for cicd_release_task
-- ----------------------------
CREATE TABLE IF NOT EXISTS `cicd_release_task` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `release_id` bigint NOT NULL DEFAULT 0 COMMENT '发布单ID',
  `cluster_id` bigint NOT NULL DEFAULT 0 COMMENT '目标集群ID',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'Pending' COMMENT '状态(Pending/Running/Succeeded/Failed/Canceled)',
  `message` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '状态消息',
  `prev_image` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '更新前镜像(用于回滚)',
  `target_image` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '目标镜像',
  `started_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '开始执行时间(Unix时间戳)',
  `finished_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '完成时间(Unix时间戳)',
  `created_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '创建时间(Unix时间戳)',
  `modified_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '修改时间(Unix时间戳)',
  `deleted_at` bigint UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除时间(Unix时间戳)',
  `is_del` tinyint UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否删除:0未删除,1删除',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_release_id` (`release_id` ASC) USING BTREE,
  INDEX `idx_cluster_id` (`cluster_id` ASC) USING BTREE,
  INDEX `idx_status` (`status` ASC) USING BTREE,
  INDEX `idx_is_del` (`is_del` ASC) USING BTREE
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci
  COMMENT='CICD发布任务表'
  ROW_FORMAT=DYNAMIC;

SET FOREIGN_KEY_CHECKS = 1;