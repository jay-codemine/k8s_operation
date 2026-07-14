-- ============================================================
-- 金丝雀部署功能 - 数据库迁移
-- 在 cicd_pipeline 表新增金丝雀相关字段
-- ============================================================

ALTER TABLE `cicd_pipeline`
  ADD COLUMN IF NOT EXISTS `enable_canary` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用金丝雀部署' AFTER `require_approval`,
  ADD COLUMN IF NOT EXISTS `canary_replicas` int NOT NULL DEFAULT 1 COMMENT '金丝雀副本数' AFTER `enable_canary`,
  ADD COLUMN IF NOT EXISTS `canary_traffic_ratio` int NOT NULL DEFAULT 10 COMMENT '金丝雀流量比例(%)' AFTER `canary_replicas`,
  ADD COLUMN IF NOT EXISTS `canary_duration_sec` int NOT NULL DEFAULT 300 COMMENT '金丝雀观察时长(秒)' AFTER `canary_traffic_ratio`,
  ADD COLUMN IF NOT EXISTS `canary_auto_promote` tinyint(1) NOT NULL DEFAULT 0 COMMENT '观察通过自动晋升' AFTER `canary_duration_sec`,
  ADD COLUMN IF NOT EXISTS `canary_analysis_rules` text COMMENT '金丝雀分析规则JSON' AFTER `canary_auto_promote`;
