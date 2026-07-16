-- ============================================================
-- K8sOperation 平台全量初始化 SQL
-- 用法: mysql -u root -p < docs/sql/k8s_platform_full_init.sql
-- ============================================================
CREATE DATABASE IF NOT EXISTS `k8s-platform` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `k8s-platform`;

-- MySQL dump 10.13  Distrib 8.0.31, for Win64 (x86_64)
--
-- Host: localhost    Database: k8s-platform
-- ------------------------------------------------------
-- Server version	8.3.0

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `ai_approval_logs`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_approval_logs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `approval_id` int unsigned NOT NULL,
  `user_id` int unsigned NOT NULL,
  `action` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `comment` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_approval_id` (`approval_id`),
  KEY `idx_ai_approval_logs_approval_id` (`approval_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 审批操作日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ai_approval_requests`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_approval_requests` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `conversation_id` int unsigned DEFAULT NULL,
  `request_user_id` int unsigned NOT NULL,
  `approver_user_id` int unsigned DEFAULT '0',
  `intent` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `resource` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `resource_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `namespace` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cluster_id` int unsigned DEFAULT '0',
  `risk_level` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `operation_json` text COLLATE utf8mb4_unicode_ci,
  `tool_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `tool_args_json` text COLLATE utf8mb4_unicode_ci,
  `tool_call_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `execute_result` text COLLATE utf8mb4_unicode_ci,
  `executed` tinyint(1) DEFAULT '0',
  `summary` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned DEFAULT '1',
  `approve_comment` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `expire_at` int unsigned DEFAULT '0',
  `created_at` int unsigned DEFAULT NULL,
  `modified_at` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_request_user_id` (`request_user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_conversation_id` (`conversation_id`),
  KEY `idx_ai_approval_requests_conversation_id` (`conversation_id`),
  KEY `idx_ai_approval_requests_request_user_id` (`request_user_id`),
  KEY `idx_ai_approval_requests_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 高危操作审批请求表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ai_conversations`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_conversations` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int unsigned NOT NULL,
  `title` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '新对话',
  `status` tinyint unsigned DEFAULT '1',
  `created_at` int unsigned DEFAULT NULL,
  `modified_at` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_ai_conversations_user_id` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 助手会话表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `ai_messages`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ai_messages` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `conversation_id` int unsigned NOT NULL,
  `role` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text COLLATE utf8mb4_unicode_ci,
  `intent_json` text COLLATE utf8mb4_unicode_ci,
  `token_used` bigint DEFAULT '0',
  `created_at` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_conversation_id` (`conversation_id`),
  KEY `idx_ai_messages_conversation_id` (`conversation_id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 助手消息表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `aiops_analysis_record`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `aiops_analysis_record` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `type` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `ref_id` bigint DEFAULT NULL,
  `title` varchar(300) COLLATE utf8mb4_unicode_ci NOT NULL,
  `input` text COLLATE utf8mb4_unicode_ci,
  `result` longtext COLLATE utf8mb4_unicode_ci,
  `severity` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `suggestions` text COLLATE utf8mb4_unicode_ci,
  `model` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `tokens_used` bigint DEFAULT '0',
  `latency_ms` bigint DEFAULT '0',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'success',
  `error` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user_id` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_aiops_analysis_record_type` (`type`),
  KEY `idx_aiops_analysis_record_ref_id` (`ref_id`),
  KEY `idx_aiops_analysis_record_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `aiops_inspection_report`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `aiops_inspection_report` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `type` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `scope` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT 'full',
  `scope_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `health_score` bigint DEFAULT '0',
  `level` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `summary` text COLLATE utf8mb4_unicode_ci,
  `details` longtext COLLATE utf8mb4_unicode_ci,
  `ai_analysis` longtext COLLATE utf8mb4_unicode_ci,
  `findings` bigint DEFAULT '0',
  `suggestions` bigint DEFAULT '0',
  `duration` bigint DEFAULT '0',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'running',
  `error` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `triggered_by` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `completed_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_aiops_inspection_report_type` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=35 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `app_store_apps`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `app_store_apps` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `display_name` varchar(256) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `category` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `version` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `icon` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `provider` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `chart_url` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `doc_url` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` tinyint unsigned DEFAULT '1',
  `featured` tinyint unsigned DEFAULT '0',
  `sort_order` bigint DEFAULT '0',
  `tags` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `min_k8s` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `namespace` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `values_yaml` text COLLATE utf8mb4_unicode_ci,
  `created_at` int unsigned DEFAULT NULL,
  `modified_at` int unsigned DEFAULT NULL,
  `deleted_at` int unsigned DEFAULT NULL,
  `is_del` tinyint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_app_store_apps_name` (`name`),
  KEY `idx_category` (`category`),
  KEY `idx_status` (`status`),
  KEY `idx_app_store_apps_category` (`category`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用商城应用表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `app_store_components`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `app_store_components` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int unsigned NOT NULL,
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `image` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL,
  `replicas` int DEFAULT '1',
  `ports` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `args` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cpu_req` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '50m',
  `cpu_lim` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '200m',
  `mem_req` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '64Mi',
  `mem_lim` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '256Mi',
  `sort_order` bigint DEFAULT '0',
  `created_at` int unsigned DEFAULT NULL,
  `modified_at` int unsigned DEFAULT NULL,
  `deleted_at` int unsigned DEFAULT NULL,
  `is_del` tinyint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_app_store_components_app_id` (`app_id`)
) ENGINE=InnoDB AUTO_INCREMENT=28 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用商城组件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `app_store_installs`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `app_store_installs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `app_id` int unsigned NOT NULL,
  `app_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `cluster_id` int unsigned NOT NULL,
  `cluster_name` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `namespace` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `release_name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL,
  `version` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `values` text COLLATE utf8mb4_unicode_ci,
  `status` tinyint unsigned DEFAULT '1',
  `message` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `operator` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` int unsigned DEFAULT NULL,
  `modified_at` int unsigned DEFAULT NULL,
  `deleted_at` int unsigned DEFAULT NULL,
  `is_del` tinyint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_app_id` (`app_id`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_status` (`status`),
  KEY `idx_app_store_installs_app_id` (`app_id`),
  KEY `idx_app_store_installs_cluster_id` (`cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应用商城安装记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `audit_log`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `audit_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `user_id` bigint NOT NULL,
  `username` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_ip` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user_agent` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `action` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `action_display` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `module` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `target_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `target_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `target_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `request_uri` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `request_method` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `request_body` text COLLATE utf8mb4_unicode_ci,
  `response_code` bigint DEFAULT '0',
  `response_message` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `detail` json DEFAULT NULL,
  `extra` json DEFAULT NULL,
  `cluster_id` bigint DEFAULT NULL,
  `cluster_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `namespace` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `pipeline_id` bigint DEFAULT NULL,
  `pipeline_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `run_id` bigint DEFAULT NULL COMMENT '关联运行记录ID',
  `project_id` bigint DEFAULT NULL,
  `project_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'success',
  `error_message` varchar(1000) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `duration_ms` bigint DEFAULT '0',
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_module` (`module`),
  KEY `idx_target` (`target_type`,`target_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_audit_user` (`user_id`),
  KEY `idx_audit_action` (`action`),
  KEY `idx_audit_module` (`module`),
  KEY `idx_audit_target` (`target_type`,`target_id`),
  KEY `idx_audit_cluster` (`cluster_id`),
  KEY `idx_audit_pipeline` (`pipeline_id`),
  KEY `idx_audit_status` (`status`),
  KEY `idx_audit_created` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=822 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_approval`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_approval` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint DEFAULT NULL,
  `pipeline_run_id` bigint DEFAULT NULL,
  `release_id` bigint DEFAULT NULL,
  `env_name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `image_digest` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `request_user_id` bigint DEFAULT NULL,
  `request_reason` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `approve_user_id` bigint DEFAULT NULL,
  `approve_reason` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `approve_time` bigint unsigned DEFAULT NULL,
  `expire_time` bigint unsigned DEFAULT NULL,
  `feishu_token` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `modified_at` bigint unsigned DEFAULT NULL,
  `stage_id` bigint DEFAULT NULL,
  `approval_level` bigint DEFAULT '1',
  `total_levels` bigint DEFAULT '1',
  `level_label` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_pipeline_run_id` (`pipeline_run_id`),
  KEY `idx_status` (`status`),
  KEY `idx_request_user_id` (`request_user_id`),
  KEY `idx_stage_id` (`stage_id`),
  KEY `idx_feishu_token` (`feishu_token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD审批记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_artifact`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_artifact` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint DEFAULT NULL COMMENT '关联流水线ID',
  `run_id` bigint DEFAULT NULL COMMENT '关联运行记录ID',
  `build_number` int DEFAULT '0' COMMENT 'Jenkins构建号',
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '制品名称（如 order-service-1.0.0.jar）',
  `artifact_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '制品类型：jar/war/binary/dist/wheel/image/archive',
  `version` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '版本号',
  `language_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '语言类型：go/java/frontend/python',
  `file_path` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '文件存储路径',
  `file_size` bigint NOT NULL DEFAULT '0' COMMENT '文件大小（字节）',
  `sha256` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'SHA256校验和',
  `storage_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'local' COMMENT '存储类型：local/s3/oss',
  `git_repo` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Git仓库地址',
  `git_branch` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Git Commit SHA',
  `image_repo` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '镜像仓库地址',
  `image_tag` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '镜像摘要',
  `build_duration` int NOT NULL DEFAULT '0' COMMENT '构建耗时（秒）',
  `build_log` text COLLATE utf8mb4_unicode_ci COMMENT '构建摘要日志',
  `metadata` json DEFAULT NULL COMMENT '扩展元数据',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'ready' COMMENT '状态：uploading/ready/expired/deleted',
  `download_count` int NOT NULL DEFAULT '0' COMMENT '下载次数',
  `created_user_id` bigint NOT NULL DEFAULT '0' COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_run_id` (`run_id`),
  KEY `idx_artifact_type` (`artifact_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD 制品库表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_build`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_build` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '应用名称',
  `git_url` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Git URL',
  `git_branch` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Git Commit',
  `jenkins_job` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Jenkins Job',
  `jenkins_queue_id` bigint NOT NULL DEFAULT '0' COMMENT 'Jenkins队列ID',
  `jenkins_build_number` int NOT NULL DEFAULT '0' COMMENT 'Jenkins构建号',
  `jenkins_build_url` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Jenkins构建URL',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态',
  `message` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '消息',
  `image_repo` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '镜像仓库',
  `image_tag` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '镜像摘要',
  `sbom_ref` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'SBOM引用',
  `sign_ref` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '签名引用',
  `created_user_id` bigint NOT NULL DEFAULT '0' COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_app_name` (`app_name`),
  KEY `idx_jenkins_job` (`jenkins_job`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD构建记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_build_agent`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_build_agent` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `display_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(1000) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `category` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `scope` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `version` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `file_name` varchar(300) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `file_path` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `file_size` bigint DEFAULT NULL,
  `sha256` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `download_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `doc_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `icon` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `docker_copy_dest` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `env_key` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `env_value` varchar(2000) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'active',
  `download_count` bigint DEFAULT '0',
  `used_count` bigint DEFAULT '0',
  `created_user_id` bigint DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `modified_at` bigint unsigned DEFAULT NULL,
  `deleted_at` bigint unsigned DEFAULT NULL,
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_cicd_build_agent_name` (`name`),
  KEY `idx_category` (`category`),
  KEY `idx_scope` (`scope`),
  KEY `idx_cicd_build_agent_category` (`category`),
  KEY `idx_cicd_build_agent_scope` (`scope`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD 构建探针表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_deploy_approval`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_deploy_approval` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint unsigned NOT NULL COMMENT '流水线ID',
  `release_id` bigint unsigned DEFAULT '0' COMMENT '发布单ID',
  `env` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标环境',
  `requested_config` text COLLATE utf8mb4_unicode_ci COMMENT '申请的资源配置JSON',
  `current_config` text COLLATE utf8mb4_unicode_ci COMMENT '当前线上配置JSON',
  `risk_level` enum('low','medium','high') COLLATE utf8mb4_unicode_ci DEFAULT 'low',
  `risk_warnings` text COLLATE utf8mb4_unicode_ci COMMENT '风险提示列表JSON',
  `status` enum('pending','approved','rejected','expired','cancelled') COLLATE utf8mb4_unicode_ci DEFAULT 'pending',
  `applicant_id` bigint unsigned NOT NULL COMMENT '申请人ID',
  `applicant_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `approver_id` bigint unsigned DEFAULT '0' COMMENT '审批人ID',
  `approver_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `approve_comment` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '审批意见',
  `applied_at` bigint unsigned DEFAULT '0' COMMENT '申请时间',
  `approved_at` bigint unsigned DEFAULT '0' COMMENT '审批时间',
  `expired_at` bigint unsigned DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY (`id`),
  KEY `idx_pipeline` (`pipeline_id`),
  KEY `idx_status` (`status`,`applied_at`),
  KEY `idx_applicant` (`applicant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD发布审批记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_env_resource_rule`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_env_resource_rule` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `env` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'dev/test/staging/prod',
  `service_type` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '服务类型，空=通用',
  `cpu_limit_max` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '4' COMMENT 'CPU limit 最大值',
  `memory_limit_max` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '8Gi' COMMENT '内存 limit 最大值',
  `replicas_max` int NOT NULL DEFAULT '10' COMMENT '副本数上限',
  `cpu_request_min` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'CPU request 最小值',
  `memory_request_min` varchar(16) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '内存 request 最小值',
  `replicas_min` int DEFAULT '1' COMMENT '副本数下限',
  `require_approval` tinyint(1) DEFAULT '0' COMMENT '是否需要审批',
  `approval_role` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '审批角色：sre/admin',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '规则说明',
  `created_at` bigint unsigned DEFAULT '0',
  `modified_at` bigint unsigned DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_env_type` (`env`,`service_type`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD环境资源规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_environment`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_environment` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `display_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cluster_id` bigint DEFAULT NULL,
  `namespace` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `color` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort_order` bigint DEFAULT NULL,
  `require_approval` tinyint(1) DEFAULT NULL,
  `approval_users` json DEFAULT NULL,
  `created_user_id` bigint DEFAULT NULL,
  `created_at` bigint unsigned DEFAULT NULL,
  `modified_at` bigint unsigned DEFAULT NULL,
  `deleted_at` bigint unsigned DEFAULT NULL,
  `is_del` tinyint unsigned DEFAULT NULL,
  `approval_levels` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_is_del` (`is_del`),
  KEY `idx_cicd_environment_name` (`name`),
  KEY `idx_cicd_environment_cluster_id` (`cluster_id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD环境配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_language_profile`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_language_profile` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `language_key` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `display_name` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `icon` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `build_image` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `pre_build` varchar(1000) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `build_cmd` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `test_cmd` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `dockerfile` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT 'Dockerfile',
  `build_ctx` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '.',
  `dockerfile_template` text COLLATE utf8mb4_unicode_ci,
  `is_builtin` tinyint(1) DEFAULT '0',
  `sort_order` bigint DEFAULT '0',
  `created_at` bigint unsigned DEFAULT NULL,
  `modified_at` bigint unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_cicd_language_profile_language_key` (`language_key`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='璇?█閰嶇疆娉ㄥ唽琛';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_pipeline`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_pipeline` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `project_id` bigint DEFAULT NULL COMMENT '所属项目ID',
  `name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` longtext COLLATE utf8mb4_unicode_ci,
  `git_repo` longtext COLLATE utf8mb4_unicode_ci,
  `git_branch` longtext COLLATE utf8mb4_unicode_ci,
  `jenkins_url` longtext COLLATE utf8mb4_unicode_ci,
  `jenkins_job` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'Jenkins Job名称',
  `jenkins_credential_id` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Jenkins凭证ID',
  `language_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'custom' COMMENT '语言类型:go/java/frontend/python/custom',
  `auto_deploy` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否自动部署',
  `target_cluster_id` bigint DEFAULT NULL COMMENT '目标集群ID',
  `target_namespace` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '目标命名空间',
  `target_workload_kind` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '工作负载类型',
  `target_workload_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '工作负载名称',
  `target_container` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '目标容器名称',
  `deploy_env` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'dev' COMMENT '部署环境',
  `require_approval` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否需要审批',
  `enable_canary` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否启用金丝雀部署',
  `canary_replicas` int NOT NULL DEFAULT '1' COMMENT '金丝雀副本数',
  `canary_traffic_ratio` int NOT NULL DEFAULT '10' COMMENT '金丝雀流量比例(%)',
  `canary_duration_sec` int NOT NULL DEFAULT '300' COMMENT '金丝雀观察时长(秒)',
  `canary_auto_promote` tinyint(1) NOT NULL DEFAULT '0' COMMENT '观察通过自动晋升',
  `canary_analysis_rules` text COLLATE utf8mb4_unicode_ci COMMENT '金丝雀分析规则JSON',
  `enable_sonar` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否启用SonarQube代码扫描',
  `enable_artifact_upload` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否启用制品上传',
  `enable_deploy_silence` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'silence flag',
  `silence_buffer_minutes` int NOT NULL DEFAULT '10' COMMENT 'silence buffer',
  `silence_severities` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'warning,info' COMMENT 'silence severities',
  `last_deploy_image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '最新部署镜像',
  `last_deploy_digest` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '镜像摘要',
  `last_deploy_time` bigint DEFAULT NULL COMMENT '最新部署时间',
  `last_deploy_status` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '最新部署状态',
  `last_deploy_version` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '最新部署版本',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'idle' COMMENT '状态:idle,running,disabled',
  `last_run_status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '最后运行状态',
  `last_run_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '最后运行时间',
  `last_build_number` int NOT NULL DEFAULT '0' COMMENT '最后构建号',
  `last_build_url` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '最后构建URL',
  `env_vars` json DEFAULT NULL COMMENT '环境变量',
  `deploy_config` json DEFAULT NULL COMMENT '部署配置',
  `stages` json DEFAULT NULL COMMENT '阶段配置',
  `created_user_id` bigint NOT NULL DEFAULT '0' COMMENT '创建人',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  `deploy_mode` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'jenkins',
  `gitops_config` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_project_id` (`project_id`),
  KEY `idx_jenkins_job` (`jenkins_job`),
  KEY `idx_status` (`status`),
  KEY `idx_auto_deploy` (`auto_deploy`),
  KEY `idx_target_cluster` (`target_cluster_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD流水线表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_pipeline_run`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_pipeline_run` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint DEFAULT NULL,
  `build_number` bigint DEFAULT NULL,
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态:pending,running,success,failed,aborted',
  `trigger_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'manual' COMMENT '触发类型:manual,webhook,scheduled',
  `trigger_user_id` bigint NOT NULL DEFAULT '0' COMMENT '触发人ID',
  `git_commit` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Git Commit',
  `git_branch` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Git分支',
  `git_commit_message` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '提交消息',
  `jenkins_build_url` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'Jenkins构建URL',
  `duration_sec` int NOT NULL DEFAULT '0' COMMENT '执行时长(秒)',
  `console_log` longtext COLLATE utf8mb4_unicode_ci COMMENT '控制台日志',
  `stages_result` json DEFAULT NULL COMMENT '阶段结果',
  `started_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '开始时间',
  `finished_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '结束时间',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `image_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '构建镜像地址',
  `image_digest` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '镜像摘要',
  `workflow_name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'workflow name',
  `argo_app_name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'argo app name',
  `sync_revision` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'sync revision',
  `sync_status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'sync status',
  `callback_received` tinyint(1) DEFAULT '0' COMMENT '是否收到回调',
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_build_number` (`build_number`),
  KEY `idx_status` (`status`),
  KEY `idx_started_at` (`started_at`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线运行记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_pipeline_stage`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_pipeline_stage` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `run_id` bigint NOT NULL COMMENT '运行记录ID',
  `pipeline_id` bigint NOT NULL COMMENT '流水线ID',
  `stage_order` int NOT NULL DEFAULT '0' COMMENT '阶段顺序',
  `stage_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '阶段类型:checkout,dependencies,compile,test,lint,sonar,quality_gate,build_binary,upload_artifact,build,push,approval,deploy',
  `stage_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '阶段名称',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态',
  `started_at` bigint DEFAULT NULL COMMENT '开始时间',
  `finished_at` bigint DEFAULT NULL COMMENT '结束时间',
  `duration_sec` int DEFAULT '0' COMMENT '执行时长',
  `logs` longtext COLLATE utf8mb4_unicode_ci COMMENT '阶段日志',
  `jenkins_stage_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Jenkins阶段ID',
  `approval_user_id` bigint DEFAULT NULL COMMENT '审批人ID',
  `approval_comment` text COLLATE utf8mb4_unicode_ci COMMENT '审批评论',
  `approval_decision` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '审批决定:approved,rejected',
  `deploy_cluster_id` bigint DEFAULT NULL COMMENT '部署集群ID',
  `deploy_namespace` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '部署命名空间',
  `deploy_workload_kind` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '工作负载类型',
  `deploy_workload_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '工作负载名称',
  `deploy_container` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '容器名称',
  `deploy_image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '部署镜像',
  `deploy_old_image` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '旧镜像',
  `deploy_replicas` int DEFAULT NULL COMMENT '副本数',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `created_at` bigint NOT NULL,
  `modified_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_run_id` (`run_id`),
  KEY `idx_pipeline_id` (`pipeline_id`),
  KEY `idx_stage_type` (`stage_type`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线阶段执行记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_pipeline_template`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_pipeline_template` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模板名称',
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '模板描述',
  `type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'custom' COMMENT '模板类型: frontend/backend/microservice/database/custom',
  `stages` json DEFAULT NULL COMMENT '阶段配置',
  `default_env_vars` json DEFAULT NULL COMMENT '默认环境变量',
  `deploy_config` json DEFAULT NULL COMMENT '默认部署配置',
  `jenkins_template` text COLLATE utf8mb4_unicode_ci COMMENT 'Jenkinsfile模板',
  `usage_count` bigint NOT NULL DEFAULT '0' COMMENT '使用次数',
  `created_user_id` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='流水线模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_release`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_release` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `app_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '应用名称',
  `namespace` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'default' COMMENT '命名空间',
  `workload_kind` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Deployment' COMMENT '工作负载类型',
  `workload_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '工作负载名称',
  `container_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '容器名称',
  `strategy` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'rolling' COMMENT '发布策略',
  `timeout_sec` int unsigned NOT NULL DEFAULT '300' COMMENT '超时时间(秒)',
  `concurrency` int unsigned NOT NULL DEFAULT '3' COMMENT '并发数',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pending' COMMENT '状态',
  `message` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '消息',
  `created_user_id` bigint NOT NULL DEFAULT '0' COMMENT '创建人',
  `request_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '请求ID',
  `build_id` bigint NOT NULL DEFAULT '0' COMMENT '关联构建ID',
  `image_repo` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '镜像仓库',
  `image_tag` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '镜像标签',
  `image_digest` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '镜像摘要',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  `canary_config` json DEFAULT NULL COMMENT '閲戜笣闆?彂甯冮厤缃',
  `canary_status` json DEFAULT NULL COMMENT '閲戜笣闆?彂甯冪姸鎬',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_request_id` (`request_id`),
  KEY `idx_app_name` (`app_name`),
  KEY `idx_status` (`status`),
  KEY `idx_build_id` (`build_id`),
  KEY `idx_modified_at` (`modified_at`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD发布单表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_release_stage`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_release_stage` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `release_id` bigint NOT NULL DEFAULT '0' COMMENT '发布单ID',
  `stage_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '阶段名称',
  `stage_order` int NOT NULL DEFAULT '0' COMMENT '阶段顺序',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态',
  `message` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '消息',
  `logs` text COLLATE utf8mb4_unicode_ci COMMENT '日志',
  `start_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '开始时间',
  `end_time` bigint unsigned NOT NULL DEFAULT '0' COMMENT '结束时间',
  `duration` bigint NOT NULL DEFAULT '0' COMMENT '持续时间',
  `build_number` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '构建号',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_release_id` (`release_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD发布阶段表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_release_task`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_release_task` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `release_id` bigint NOT NULL COMMENT '发布单ID',
  `cluster_id` bigint NOT NULL COMMENT '集群ID',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Pending' COMMENT '状态',
  `message` varchar(2048) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '消息',
  `prev_image` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '原镜像',
  `target_image` varchar(512) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '目标镜像',
  `started_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '开始时间',
  `finished_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '结束时间',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_release_id` (`release_id`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CI/CD发布任务表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_resource_change_log`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_resource_change_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint unsigned NOT NULL,
  `env` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL,
  `change_type` enum('create','update','scale','rollback') COLLATE utf8mb4_unicode_ci NOT NULL,
  `before_config` text COLLATE utf8mb4_unicode_ci COMMENT '变更前配置JSON',
  `after_config` text COLLATE utf8mb4_unicode_ci COMMENT '变更后配置JSON',
  `operator_id` bigint unsigned NOT NULL,
  `operator_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
  `reason` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '变更原因',
  `created_at` bigint unsigned DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_env` (`pipeline_id`,`env`),
  KEY `idx_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD资源配置变更日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `cicd_resource_template`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `cicd_resource_template` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模板名称：small/medium/large/custom',
  `service_type` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '服务类型：java/go/node/python',
  `env` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '环境：dev/test/staging/prod',
  `replicas_default` int NOT NULL DEFAULT '1' COMMENT '默认副本数',
  `replicas_min` int NOT NULL DEFAULT '1' COMMENT '最小副本数',
  `replicas_max` int NOT NULL DEFAULT '10' COMMENT '最大副本数',
  `cpu_request` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '200m',
  `cpu_limit` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '500m',
  `memory_request` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '256Mi',
  `memory_limit` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '512Mi',
  `hpa_enabled` tinyint(1) DEFAULT '0' COMMENT '是否启用HPA',
  `hpa_min_replicas` int DEFAULT '2',
  `hpa_max_replicas` int DEFAULT '10',
  `hpa_cpu_target` int DEFAULT '70' COMMENT 'CPU目标利用率%',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '模板说明',
  `is_default` tinyint(1) DEFAULT '0' COMMENT '是否默认模板',
  `sort_order` int DEFAULT '0',
  `created_at` bigint unsigned DEFAULT '0',
  `modified_at` bigint unsigned DEFAULT '0',
  `deleted_at` bigint unsigned DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_type_env_name` (`service_type`,`env`,`name`,`deleted_at`)
) ENGINE=InnoDB AUTO_INCREMENT=39 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CICD资源档位模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `iam_env_audit_log`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iam_env_audit_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '操作用户ID',
  `username` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '用户名',
  `action` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作类型',
  `resource_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源类型',
  `resource_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '资源名称',
  `cluster_id` bigint DEFAULT NULL COMMENT '集群ID',
  `cluster_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  `env_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '环境类型',
  `namespace` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '命名空间',
  `success` tinyint(1) DEFAULT '1' COMMENT '是否成功',
  `error_message` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '错误信息',
  `client_ip` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '客户端IP',
  `user_agent` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'User-Agent',
  `request_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '请求ID',
  `detail` json DEFAULT NULL COMMENT '详情',
  `created_at` bigint unsigned NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_env_type` (`env_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='环境操作审计日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `iam_env_binding`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iam_env_binding` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `subject_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '主体类型: user/group',
  `subject_id` bigint NOT NULL COMMENT '主体ID',
  `subject_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '主体名称',
  `cluster_id` bigint NOT NULL COMMENT '集群ID',
  `cluster_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '集群名称',
  `env_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '环境类型',
  `namespaces` json DEFAULT NULL COMMENT '命名空间列表',
  `env_role` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '环境角色',
  `custom_actions` json DEFAULT NULL COMMENT '自定义操作权限',
  `max_env_level` int DEFAULT '1' COMMENT '最高环境级别',
  `bypass_approval` tinyint(1) DEFAULT '0' COMMENT '是否跳过审批',
  `k8s_synced` tinyint(1) DEFAULT '0' COMMENT 'K8s RBAC是否已同步',
  `k8s_role_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'K8s Role名称',
  `k8s_binding_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'K8s RoleBinding名称',
  `k8s_sync_error` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'K8s同步错误',
  `k8s_synced_at` bigint unsigned DEFAULT NULL COMMENT 'K8s同步时间',
  `expire_at` bigint unsigned DEFAULT NULL COMMENT '过期时间',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'active' COMMENT '状态: active/expired/revoked',
  `remark` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '备注',
  `granted_by` bigint DEFAULT '0' COMMENT '授权人ID',
  `revoked_by` bigint DEFAULT NULL COMMENT '撤销人ID',
  `revoked_at` bigint unsigned DEFAULT NULL COMMENT '撤销时间',
  `created_at` bigint unsigned NOT NULL COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_subject` (`subject_type`,`subject_id`),
  KEY `idx_cluster` (`cluster_id`),
  KEY `idx_env_type` (`env_type`),
  KEY `idx_env_role` (`env_role`),
  KEY `idx_status` (`status`),
  KEY `idx_k8s_synced` (`k8s_synced`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='环境权限绑定表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `iam_grant`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iam_grant` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '授权ID',
  `subject_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '主体类型: user/group/project',
  `subject_id` bigint NOT NULL COMMENT '主体ID',
  `subject_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '主体名称（冗余）',
  `scope_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '范围类型: cluster/namespace/cicd_project/cicd_pipeline',
  `scope_id` bigint DEFAULT NULL COMMENT '范围ID（集群ID/项目ID/流水线ID）',
  `scope_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '范围名称（冗余）',
  `namespaces` json DEFAULT NULL COMMENT '命名空间列表（支持通配符如 ["default","app-*"]）',
  `role_template_id` bigint NOT NULL COMMENT '权限模板ID',
  `role_template_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '模板名称（冗余）',
  `expire_at` bigint unsigned DEFAULT NULL COMMENT '过期时间（NULL 表示永不过期）',
  `k8s_synced` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'K8s RBAC 是否已同步',
  `k8s_role_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'K8s Role/ClusterRole 名称',
  `k8s_binding_name` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'K8s RoleBinding/ClusterRoleBinding 名称',
  `k8s_sync_error` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'K8s 同步错误信息',
  `k8s_synced_at` bigint unsigned DEFAULT NULL COMMENT 'K8s 同步时间',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active' COMMENT '状态: active/expired/revoked',
  `remark` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '备注',
  `granted_by` bigint NOT NULL DEFAULT '0' COMMENT '授权人ID',
  `revoked_by` bigint DEFAULT NULL COMMENT '撤销人ID',
  `revoked_at` bigint unsigned DEFAULT NULL COMMENT '撤销时间',
  `created_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `idx_subject` (`subject_type`,`subject_id`),
  KEY `idx_scope` (`scope_type`,`scope_id`),
  KEY `idx_role_template` (`role_template_id`),
  KEY `idx_status` (`status`),
  KEY `idx_expire_at` (`expire_at`),
  KEY `idx_granted_by` (`granted_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='授权记录表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `iam_group`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iam_group` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户组ID',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '组名称（唯一标识）',
  `display_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '显示名称',
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '描述',
  `type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'custom' COMMENT '类型: system/custom',
  `parent_id` bigint DEFAULT NULL COMMENT '父组ID（支持层级结构）',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序顺序',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_type` (`type`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `iam_group_user`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iam_group_user` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `group_id` bigint NOT NULL COMMENT '用户组ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'member' COMMENT '组内角色: owner/admin/member',
  `joined_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '加入时间',
  `created_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`,`user_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户组成员关系表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `iam_project`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iam_project` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '项目ID',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '项目名称（唯一标识）',
  `display_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '显示名称',
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '描述',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active' COMMENT '状态: active/archived/disabled',
  `owner_id` bigint NOT NULL DEFAULT '0' COMMENT '项目负责人ID',
  `default_cluster_id` bigint DEFAULT NULL COMMENT '默认集群ID',
  `default_namespace` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '默认命名空间',
  `allowed_clusters` json DEFAULT NULL COMMENT '允许的集群ID列表',
  `allowed_namespaces` json DEFAULT NULL COMMENT '允许的命名空间列表（支持通配符）',
  `labels` json DEFAULT NULL COMMENT '标签（键值对）',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_owner_id` (`owner_id`),
  KEY `idx_status` (`status`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `iam_project_member`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iam_project_member` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '记录ID',
  `project_id` bigint NOT NULL COMMENT '项目ID',
  `subject_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '主体类型: user/group',
  `subject_id` bigint NOT NULL COMMENT '主体ID（用户ID或组ID）',
  `role` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'viewer' COMMENT '项目角色: owner/admin/developer/viewer',
  `joined_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '加入时间',
  `created_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_subject` (`project_id`,`subject_type`,`subject_id`),
  KEY `idx_subject` (`subject_type`,`subject_id`),
  KEY `idx_project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='项目成员关系表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `iam_role_template`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `iam_role_template` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '模板ID',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模板名称（唯一标识）',
  `display_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '显示名称',
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '描述',
  `type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模板类型: k8s/cicd/platform',
  `builtin` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否内置模板',
  `k8s_rules` json DEFAULT NULL COMMENT 'K8s RBAC规则 [{apiGroups, resources, verbs}]',
  `k8s_cluster_scope` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否集群级别权限',
  `cicd_actions` json DEFAULT NULL COMMENT 'CICD操作权限 ["view","run","approve","deploy","rollback","delete"]',
  `platform_permissions` json DEFAULT NULL COMMENT '平台功能权限 ["cluster:manage","user:manage","audit:view"]',
  `sort_order` int NOT NULL DEFAULT '0' COMMENT '排序顺序',
  `created_by` bigint NOT NULL DEFAULT '0' COMMENT '创建人ID',
  `created_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '修改时间',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '是否删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_builtin` (`builtin`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `image_cleanup_log`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `image_cleanup_log` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `policy_id` bigint NOT NULL COMMENT '策略ID',
  `registry_id` bigint NOT NULL COMMENT '仓库ID',
  `start_time` bigint NOT NULL COMMENT '开始时间',
  `end_time` bigint DEFAULT '0' COMMENT '结束时间',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'running' COMMENT '状态',
  `scanned_count` int DEFAULT '0' COMMENT '扫描数',
  `deleted_count` int DEFAULT '0' COMMENT '删除数',
  `freed_size` bigint DEFAULT '0' COMMENT '释放空间(字节)',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '错误信息',
  `details` json DEFAULT NULL COMMENT '详情',
  PRIMARY KEY (`id`),
  KEY `idx_policy_id` (`policy_id`),
  KEY `idx_start_time` (`start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='镜像清理日志表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `image_cleanup_policy`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `image_cleanup_policy` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `registry_id` bigint NOT NULL COMMENT '仓库ID',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '策略名称',
  `enabled` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `repository_pattern` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '*' COMMENT '仓库匹配模式',
  `tag_pattern` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '*' COMMENT '标签匹配模式',
  `keep_last_count` int DEFAULT '5' COMMENT '保留最近N个',
  `keep_days` int DEFAULT '30' COMMENT '保留N天内',
  `cron_expression` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '0 2 * * *' COMMENT 'Cron表达式',
  `last_run_at` bigint DEFAULT '0' COMMENT '最后执行时间',
  `last_run_result` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '最后执行结果',
  `deleted_count` bigint DEFAULT '0' COMMENT '累计删除数',
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '描述',
  `created_by` bigint DEFAULT '0' COMMENT '创建人',
  `created_at` bigint DEFAULT '0',
  `modified_at` bigint DEFAULT '0',
  `is_del` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_registry_id` (`registry_id`),
  KEY `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='镜像清理策略表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `image_registry`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `image_registry` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '仓库名称',
  `type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'docker' COMMENT '类型:docker,harbor,acr,ecr,gcr,quay',
  `url` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '仓库地址',
  `username` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '用户名',
  `password` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '密码(加密)',
  `access_key_id` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'AccessKey ID(云厂商)',
  `access_key_secret` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'AccessKey Secret(加密)',
  `region` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '区域',
  `insecure` tinyint(1) DEFAULT '0' COMMENT '跳过TLS验证',
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '描述',
  `is_default` tinyint(1) DEFAULT '0' COMMENT '是否默认仓库',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'unknown' COMMENT '连接状态',
  `last_check_at` bigint DEFAULT '0' COMMENT '最后检测时间',
  `last_error` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '最后错误',
  `created_by` bigint DEFAULT '0' COMMENT '创建人',
  `created_at` bigint DEFAULT '0',
  `modified_at` bigint DEFAULT '0',
  `is_del` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_registry_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='镜像仓库配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `kube_cluster`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `kube_cluster` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `cluster_name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '集群名称',
  `kube_config` longtext COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'KubeConfig(加密存储)',
  `cluster_version` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '集群版本',
  `status` tinyint unsigned NOT NULL DEFAULT '2' COMMENT '状态:0=正常,1=异常,2=未检测',
  `env_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'development' COMMENT '环境类型:development,testing,staging,production',
  `env_display_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '环境显示名称',
  `env_level` int DEFAULT '1' COMMENT '环境级别:1-4(开发到生产)',
  `access_mode` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'restricted' COMMENT '访问模式:public,restricted,private',
  `require_approval` tinyint(1) DEFAULT '0' COMMENT '操作是否需要审批',
  `approval_users` json DEFAULT NULL COMMENT '审批人列表',
  `env_color` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '环境颜色标识',
  `env_description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '环境描述',
  `env_labels` json DEFAULT NULL COMMENT '环境标签',
  `project_ids` json DEFAULT NULL COMMENT '关联项目ID列表',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `modified_at` bigint unsigned NOT NULL DEFAULT '0',
  `deleted_at` bigint unsigned NOT NULL DEFAULT '0',
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  `last_check_at` bigint unsigned NOT NULL DEFAULT '0' COMMENT '最后检查时间',
  `last_error` varchar(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '最后错误信息',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_modified` (`modified_at`),
  KEY `idx_is_del` (`is_del`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='K8s集群配置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `monitor_aggregate_rule`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monitor_aggregate_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `group_by` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `group_wait` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '30s',
  `group_interval` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '5m',
  `repeat_interval` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '4h',
  `matchers` text COLLATE utf8mb4_unicode_ci,
  `channel_ids` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `enabled` tinyint(1) DEFAULT '1',
  `created_by` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `modified_at` bigint DEFAULT NULL,
  `is_del` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警聚合规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `monitor_alert_event`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monitor_alert_event` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `rule_id` bigint NOT NULL,
  `datasource_id` bigint NOT NULL,
  `rule_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `severity` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `labels` text COLLATE utf8mb4_unicode_ci,
  `annotations` text COLLATE utf8mb4_unicode_ci,
  `summary` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `fired_at` bigint NOT NULL,
  `resolved_at` bigint DEFAULT NULL,
  `acked_by` bigint DEFAULT NULL,
  `acked_at` bigint DEFAULT NULL,
  `silenced_until` bigint DEFAULT NULL,
  `notify_result` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_rule_id` (`rule_id`),
  KEY `idx_datasource_id` (`datasource_id`),
  KEY `idx_severity` (`severity`),
  KEY `idx_status` (`status`),
  KEY `idx_fired_at` (`fired_at`),
  KEY `idx_monitor_alert_event_rule_id` (`rule_id`),
  KEY `idx_monitor_alert_event_datasource_id` (`datasource_id`),
  KEY `idx_monitor_alert_event_severity` (`severity`),
  KEY `idx_monitor_alert_event_status` (`status`),
  KEY `idx_monitor_alert_event_fired_at` (`fired_at`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警事件表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `monitor_alert_rule`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monitor_alert_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `datasource_id` bigint NOT NULL,
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `group` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT 'default',
  `severity` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL,
  `expr` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `duration` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '5m',
  `summary` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` text COLLATE utf8mb4_unicode_ci,
  `labels` text COLLATE utf8mb4_unicode_ci,
  `annotations` text COLLATE utf8mb4_unicode_ci,
  `enabled` tinyint(1) DEFAULT '1',
  `notify_channels` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `notify_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `eval_interval` bigint DEFAULT '60',
  `last_eval_at` bigint DEFAULT NULL,
  `last_eval_result` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_by` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `modified_at` bigint DEFAULT NULL,
  `is_del` tinyint(1) DEFAULT '0',
  `pending_since` bigint DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_datasource_id` (`datasource_id`),
  KEY `idx_severity` (`severity`),
  KEY `idx_monitor_alert_rule_datasource_id` (`datasource_id`),
  KEY `idx_monitor_alert_rule_severity` (`severity`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `monitor_datasource`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monitor_datasource` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `url` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `cluster_id` bigint DEFAULT '0',
  `access_mode` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'proxy',
  `auth_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'none',
  `auth_user` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `auth_pass` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `tls_cert` text COLLATE utf8mb4_unicode_ci,
  `tls_key` text COLLATE utf8mb4_unicode_ci,
  `ca_cert` text COLLATE utf8mb4_unicode_ci,
  `is_default` tinyint(1) DEFAULT '0',
  `enabled` tinyint(1) DEFAULT '1',
  `timeout` int DEFAULT '30',
  `scrape_interval` int DEFAULT '15',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'unknown',
  `last_check_at` bigint DEFAULT NULL,
  `created_by` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `modified_at` bigint DEFAULT NULL,
  `is_del` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_monitor_datasource_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_monitor_datasource_type` (`type`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_monitor_datasource_cluster_id` (`cluster_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控数据源表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `monitor_inhibit_rule`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monitor_inhibit_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `source_matchers` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `target_matchers` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `equal_labels` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `enabled` tinyint(1) DEFAULT '1',
  `created_by` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `modified_at` bigint DEFAULT NULL,
  `is_del` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警抑制规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `monitor_notify_channel`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monitor_notify_channel` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `webhook_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `secret` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `at_mobiles` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `at_all` tinyint(1) DEFAULT '0',
  `smtp_host` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `smtp_port` int DEFAULT '465',
  `smtp_user` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `smtp_pass` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `smtp_to` text COLLATE utf8mb4_unicode_ci,
  `msg_template` text COLLATE utf8mb4_unicode_ci,
  `enabled` tinyint(1) DEFAULT '1',
  `send_resolved` tinyint(1) DEFAULT '1',
  `rate_limit` bigint DEFAULT '10',
  `created_by` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `modified_at` bigint DEFAULT NULL,
  `is_del` tinyint(1) DEFAULT '0',
  `security_keyword` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_monitor_notify_channel_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_monitor_notify_channel_type` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控通知渠道表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `monitor_notify_route_policy`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monitor_notify_route_policy` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `priority` bigint DEFAULT '100',
  `channel_ids` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL,
  `match_mode` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'any',
  `severities` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `groups` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `label_match` text COLLATE utf8mb4_unicode_ci,
  `is_default` tinyint(1) DEFAULT '0',
  `enabled` tinyint(1) DEFAULT '1',
  `created_by` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `modified_at` bigint DEFAULT NULL,
  `is_del` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_monitor_notify_route_policy_name` (`name`),
  KEY `idx_monitor_notify_route_policy_priority` (`priority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `monitor_notify_template`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monitor_notify_template` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `scene` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'alert',
  `title` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `content` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_default` tinyint(1) DEFAULT '0',
  `enabled` tinyint(1) DEFAULT '1',
  `created_by` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `modified_at` bigint DEFAULT NULL,
  `is_del` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_monitor_notify_template_name` (`name`),
  KEY `idx_type` (`type`),
  KEY `idx_monitor_notify_template_type` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控通知模板表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `monitor_silence_rule`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `monitor_silence_rule` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  `type` varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL,
  `matchers` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `starts_at` bigint DEFAULT NULL,
  `ends_at` bigint DEFAULT NULL,
  `duration` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `repeat_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'once',
  `repeat_cron` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `comment` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `enabled` tinyint(1) DEFAULT '1',
  `created_by` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `modified_at` bigint DEFAULT NULL,
  `is_del` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_monitor_silence_rule_type` (`type`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='监控告警静默规则表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `platform_settings`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `platform_settings` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `category` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `key` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` text COLLATE utf8mb4_unicode_ci,
  `value_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'string',
  `label` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `desc` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` int unsigned DEFAULT NULL,
  `modified_at` int unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_category_key` (`category`,`key`)
) ENGINE=InnoDB AUTO_INCREMENT=72 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台设置表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_permission`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `display_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `resource_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `action` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `parent_id` bigint DEFAULT '0',
  `path` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `sort_order` bigint DEFAULT '0',
  `created_at` bigint unsigned DEFAULT NULL,
  `modified_at` bigint unsigned DEFAULT NULL,
  `scope` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'cluster',
  `tag` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_sys_permission_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=71 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统权限定义表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_role`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_role` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `display_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `description` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `role_type` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `is_system` tinyint(1) DEFAULT '0',
  `color` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '#1890ff',
  `icon` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'user',
  `sort_order` bigint DEFAULT '0',
  `created_at` bigint unsigned DEFAULT NULL,
  `modified_at` bigint unsigned DEFAULT NULL,
  `deleted_at` bigint unsigned DEFAULT NULL,
  `is_del` tinyint unsigned NOT NULL DEFAULT '0',
  `scope_platform` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT 'none',
  `scope_cluster` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT 'none',
  `scope_cicd` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT 'none',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_sys_role_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统角色表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_role_permission`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_role_permission` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `permission_id` bigint NOT NULL COMMENT '权限ID',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB AUTO_INCREMENT=468 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_user_cluster`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user_cluster` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint DEFAULT NULL,
  `cluster_id` bigint DEFAULT NULL,
  `role_type` varchar(30) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `namespaces` text COLLATE utf8mb4_unicode_ci,
  `can_view` tinyint(1) DEFAULT '1',
  `can_create` tinyint(1) DEFAULT '0',
  `can_update` tinyint(1) DEFAULT '0',
  `can_delete` tinyint(1) DEFAULT '0',
  `can_exec` tinyint(1) DEFAULT '0',
  `expire_at` bigint unsigned DEFAULT '0',
  `created_at` bigint unsigned DEFAULT NULL,
  `modified_at` bigint unsigned DEFAULT NULL,
  `created_by` bigint DEFAULT NULL,
  `access_level` varchar(10) COLLATE utf8mb4_unicode_ci DEFAULT 'read',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_cluster_id` (`cluster_id`),
  KEY `idx_sys_user_cluster_user_id` (`user_id`),
  KEY `idx_sys_user_cluster_cluster_id` (`cluster_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户集群权限表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `sys_user_role`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `sys_user_role` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `created_at` bigint unsigned NOT NULL DEFAULT '0',
  `created_by` bigint DEFAULT '0' COMMENT '创建人ID',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_role_id` (`role_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user`
--

/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `username` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '密码(加密)',
  `role` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'user' COMMENT '基础角色(兼容旧版)',
  `email` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邮箱',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态:1=启用,0=禁用',
  `created_at` int unsigned DEFAULT '0',
  `modified_at` int unsigned DEFAULT '0',
  `deleted_at` int unsigned DEFAULT '0',
  `is_del` tinyint unsigned DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Temporary view structure for view `v_user_env_permissions`
--

SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `v_user_env_permissions` AS SELECT 
 1 AS `user_id`,
 1 AS `username`,
 1 AS `cluster_id`,
 1 AS `cluster_name`,
 1 AS `cluster_env_type`,
 1 AS `cluster_env_level`,
 1 AS `access_mode`,
 1 AS `env_role`,
 1 AS `max_env_level`,
 1 AS `bypass_approval`,
 1 AS `namespaces`,
 1 AS `status`,
 1 AS `expire_at`,
 1 AS `k8s_synced`*/;
SET character_set_client = @saved_cs_client;

--
-- Temporary view structure for view `v_user_permissions`
--

SET @saved_cs_client     = @@character_set_client;
/*!50503 SET character_set_client = utf8mb4 */;
/*!50001 CREATE VIEW `v_user_permissions` AS SELECT 
 1 AS `user_id`,
 1 AS `subject_type`,
 1 AS `subject_id`,
 1 AS `scope_type`,
 1 AS `scope_id`,
 1 AS `scope_name`,
 1 AS `namespaces`,
 1 AS `role_template_id`,
 1 AS `role_template_name`,
 1 AS `role_type`,
 1 AS `k8s_rules`,
 1 AS `cicd_actions`,
 1 AS `platform_permissions`,
 1 AS `expire_at`,
 1 AS `status`*/;
SET character_set_client = @saved_cs_client;

--
-- Final view structure for view `v_user_env_permissions`
--

/*!50001 DROP VIEW IF EXISTS `v_user_env_permissions`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = gbk */;
/*!50001 SET character_set_results     = gbk */;
/*!50001 SET collation_connection      = gbk_chinese_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=`root`@`localhost` SQL SECURITY DEFINER */
/*!50001 VIEW `v_user_env_permissions` AS select `eb`.`subject_id` AS `user_id`,`eb`.`subject_name` AS `username`,`eb`.`cluster_id` AS `cluster_id`,`eb`.`cluster_name` AS `cluster_name`,`kc`.`env_type` AS `cluster_env_type`,`kc`.`env_level` AS `cluster_env_level`,`kc`.`access_mode` AS `access_mode`,`eb`.`env_role` AS `env_role`,`eb`.`max_env_level` AS `max_env_level`,`eb`.`bypass_approval` AS `bypass_approval`,`eb`.`namespaces` AS `namespaces`,`eb`.`status` AS `status`,`eb`.`expire_at` AS `expire_at`,`eb`.`k8s_synced` AS `k8s_synced` from (`iam_env_binding` `eb` left join `kube_cluster` `kc` on((`eb`.`cluster_id` = `kc`.`id`))) where ((`eb`.`subject_type` = 'user') and (`eb`.`status` = 'active') and ((`eb`.`expire_at` is null) or (`eb`.`expire_at` > unix_timestamp())) and (`kc`.`is_del` = 0)) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;

--
-- Final view structure for view `v_user_permissions`
--

/*!50001 DROP VIEW IF EXISTS `v_user_permissions`*/;
/*!50001 SET @saved_cs_client          = @@character_set_client */;
/*!50001 SET @saved_cs_results         = @@character_set_results */;
/*!50001 SET @saved_col_connection     = @@collation_connection */;
/*!50001 SET character_set_client      = gbk */;
/*!50001 SET character_set_results     = gbk */;
/*!50001 SET collation_connection      = gbk_chinese_ci */;
/*!50001 CREATE ALGORITHM=UNDEFINED */
/*!50013 DEFINER=`root`@`localhost` SQL SECURITY DEFINER */
/*!50001 VIEW `v_user_permissions` AS select `g`.`user_id` AS `user_id`,`g`.`subject_type` AS `subject_type`,`g`.`subject_id` AS `subject_id`,`g`.`scope_type` AS `scope_type`,`g`.`scope_id` AS `scope_id`,`g`.`scope_name` AS `scope_name`,`g`.`namespaces` AS `namespaces`,`g`.`role_template_id` AS `role_template_id`,`rt`.`name` AS `role_template_name`,`rt`.`type` AS `role_type`,`rt`.`k8s_rules` AS `k8s_rules`,`rt`.`cicd_actions` AS `cicd_actions`,`rt`.`platform_permissions` AS `platform_permissions`,`g`.`expire_at` AS `expire_at`,`g`.`status` AS `status` from ((select `ig`.`subject_id` AS `user_id`,`ig`.`subject_type` AS `subject_type`,`ig`.`subject_id` AS `subject_id`,`ig`.`scope_type` AS `scope_type`,`ig`.`scope_id` AS `scope_id`,`ig`.`scope_name` AS `scope_name`,`ig`.`namespaces` AS `namespaces`,`ig`.`role_template_id` AS `role_template_id`,`ig`.`expire_at` AS `expire_at`,`ig`.`status` AS `status` from `iam_grant` `ig` where ((`ig`.`subject_type` = 'user') and (`ig`.`status` = 'active') and ((`ig`.`expire_at` is null) or (`ig`.`expire_at` > unix_timestamp()))) union all select `igu`.`user_id` AS `user_id`,`ig`.`subject_type` AS `subject_type`,`ig`.`subject_id` AS `subject_id`,`ig`.`scope_type` AS `scope_type`,`ig`.`scope_id` AS `scope_id`,`ig`.`scope_name` AS `scope_name`,`ig`.`namespaces` AS `namespaces`,`ig`.`role_template_id` AS `role_template_id`,`ig`.`expire_at` AS `expire_at`,`ig`.`status` AS `status` from (`iam_grant` `ig` join `iam_group_user` `igu` on((`ig`.`subject_id` = `igu`.`group_id`))) where ((`ig`.`subject_type` = 'group') and (`ig`.`status` = 'active') and ((`ig`.`expire_at` is null) or (`ig`.`expire_at` > unix_timestamp())))) `g` left join `iam_role_template` `rt` on(((`g`.`role_template_id` = `rt`.`id`) and (`rt`.`is_del` = 0)))) */;
/*!50001 SET character_set_client      = @saved_cs_client */;
/*!50001 SET character_set_results     = @saved_cs_results */;
/*!50001 SET collation_connection      = @saved_col_connection */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- ============================================================
-- 默认种子数据（首次启动必需）
-- ============================================================

-- 默认角色
INSERT IGNORE INTO `sys_role` (`name`, `role_type`, `description`, `created_at`, `modified_at`, `is_del`) VALUES
('超级管理员', 'super_admin', '平台最高权限，所有功能可见', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
('平台管理员', 'platform_admin', '用户/权限/审计等平台功能', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0),
('DevOps', 'devops', '集群操作 + CICD 发布', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);

-- 默认管理员 admin/123456（bcrypt）
INSERT IGNORE INTO `user` (`username`, `password`, `created_at`, `modified_at`, `is_del`) VALUES
('admin', '$2a$10$1ouK5KV43TVOcP6GZeGUJ.AuxKpY9hPmeM5zlz0cgBN7R8OTA685G', UNIX_TIMESTAMP(), UNIX_TIMESTAMP(), 0);

-- 将 admin 分配为超级管理员
INSERT IGNORE INTO `sys_user_role` (`user_id`, `role_id`, `created_at`)
SELECT u.id, r.id, UNIX_TIMESTAMP()
FROM `user` u, `sys_role` r
WHERE u.username = 'admin' AND r.role_type = 'super_admin';

-- Dump completed on 2026-07-14 16:38:24
