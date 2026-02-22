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

SET FOREIGN_KEY_CHECKS = 1;