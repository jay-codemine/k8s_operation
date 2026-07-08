-- ============================================================
-- GitOps 模式数据库变更 (ArgoCD + Argo Workflows)
-- 兼容现有 Jenkins 模式，所有新增字段带默认值
-- ============================================================

-- 1. cicd_pipeline 表：新增 deploy_mode + gitops_config
ALTER TABLE cicd_pipeline
  ADD COLUMN deploy_mode VARCHAR(20) NOT NULL DEFAULT 'jenkins'
    COMMENT '部署模式: jenkins | gitops',
  ADD COLUMN gitops_config JSON DEFAULT NULL
    COMMENT 'GitOps 配置: ArgoCD 应用名/仓库/路径/Workflow 模板等';

-- 2. cicd_pipeline_run 表：新增 Argo Workflow + ArgoCD 跟踪字段
ALTER TABLE cicd_pipeline_run
  ADD COLUMN workflow_name VARCHAR(200) DEFAULT ''
    COMMENT 'Argo Workflow 资源名称',
  ADD COLUMN argo_app_name VARCHAR(200) DEFAULT ''
    COMMENT 'ArgoCD Application 名称',
  ADD COLUMN sync_revision VARCHAR(100) DEFAULT ''
    COMMENT 'ArgoCD 同步目标版本 (Git commit SHA)',
  ADD COLUMN sync_status VARCHAR(30) DEFAULT ''
    COMMENT 'ArgoCD 同步状态: Synced/OutOfSync/Unknown';

-- 3. 索引
CREATE INDEX idx_pipeline_deploy_mode ON cicd_pipeline(deploy_mode);
CREATE INDEX idx_run_sync_status ON cicd_pipeline_run(sync_status);
