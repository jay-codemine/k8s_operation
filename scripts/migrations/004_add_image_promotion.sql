-- ============================================================
-- 镜像晋级 (Image Promotion) 数据库变更
-- 目标：build once, promote everywhere —— 一次构建、跨环境晋级
-- 兼容现有 Jenkins / GitOps 模式，所有新增字段带默认值
-- ============================================================

-- 1. 新增 cicd_pipeline_target 表：流水线 -> 多环境部署目标映射
--    一条流水线可预先绑定 dev/test/staging/prod 各自的部署目标，
--    晋级时直接复用已构建镜像部署到目标环境，无需重新编译。
CREATE TABLE IF NOT EXISTS cicd_pipeline_target (
  id              BIGINT       NOT NULL AUTO_INCREMENT,
  pipeline_id     BIGINT       NOT NULL DEFAULT 0 COMMENT '关联流水线ID',
  env             VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '环境标识(dev/test/staging/prod)',
  cluster_id      BIGINT       NOT NULL DEFAULT 0 COMMENT '目标集群ID',
  namespace       VARCHAR(100) NOT NULL DEFAULT '' COMMENT '目标命名空间',
  workload_kind   VARCHAR(32)  NOT NULL DEFAULT 'Deployment' COMMENT '工作负载类型',
  workload_name   VARCHAR(200) NOT NULL DEFAULT '' COMMENT '工作负载名称',
  container       VARCHAR(200) NOT NULL DEFAULT '' COMMENT '容器名称',
  auto_deploy     TINYINT(1)   NOT NULL DEFAULT 0 COMMENT 'CI构建成功后是否自动部署到本环境',
  require_approval TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '晋级到本环境是否需要审批',
  promote_from    VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '本环境镜像上游来源环境',
  sort_order      INT          NOT NULL DEFAULT 0 COMMENT '晋级/展示顺序',
  created_user_id BIGINT       NOT NULL DEFAULT 0,
  created_at      BIGINT       NOT NULL DEFAULT 0,
  modified_at     BIGINT       NOT NULL DEFAULT 0,
  deleted_at      BIGINT       NOT NULL DEFAULT 0,
  is_del          TINYINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  KEY idx_pt_pipeline_env (pipeline_id, env)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流水线多环境部署目标映射';

-- 2. cicd_release 表：新增镜像晋级链追踪字段
ALTER TABLE cicd_release
  ADD COLUMN pipeline_id   BIGINT      NOT NULL DEFAULT 0 COMMENT '关联流水线ID',
  ADD COLUMN env           VARCHAR(32) NOT NULL DEFAULT '' COMMENT '目标环境',
  ADD COLUMN source_env    VARCHAR(32) NOT NULL DEFAULT '' COMMENT '晋级来源环境(首次部署为空)',
  ADD COLUMN source_run_id BIGINT      NOT NULL DEFAULT 0 COMMENT '构建该镜像的流水线运行记录ID';

-- 3. 索引
CREATE INDEX idx_release_pipeline_env ON cicd_release(pipeline_id, env);
