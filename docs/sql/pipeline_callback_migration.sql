-- ============================================================
-- 流水线回调机制数据库迁移脚本
-- 功能：添加回调状态和构建产物字段
-- ============================================================

-- 为 cicd_pipeline_run 表添加新字段
ALTER TABLE cicd_pipeline_run 
    ADD COLUMN image_url VARCHAR(500) DEFAULT '' COMMENT '构建产出的镜像地址',
    ADD COLUMN image_digest VARCHAR(100) DEFAULT '' COMMENT '镜像 digest',
    ADD COLUMN callback_received TINYINT(1) DEFAULT 0 COMMENT '是否已收到回调: 0-未收到, 1-已收到';

-- 添加索引：用于轮询查询
CREATE INDEX idx_pipeline_run_poll ON cicd_pipeline_run (status, callback_received, created_at);

-- 添加索引：幂等键（用于回调去重）
CREATE UNIQUE INDEX idx_pipeline_run_idempotent ON cicd_pipeline_run (pipeline_id, build_number);
