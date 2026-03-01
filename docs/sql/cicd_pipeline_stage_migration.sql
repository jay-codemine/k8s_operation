-- =============================================================================
-- CI/CD 流水线阶段执行记录表迁移脚本
-- 功能：支持流水线每个阶段（构建、审批、部署）独立记录和展示
-- 参考：Rancher Pipeline / KubeSphere DevOps / Jenkins Blue Ocean
-- =============================================================================

-- 创建流水线阶段执行记录表
CREATE TABLE IF NOT EXISTS `cicd_pipeline_stage` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `run_id` bigint(20) NOT NULL COMMENT '关联流水线运行记录ID',
    `pipeline_id` bigint(20) NOT NULL COMMENT '关联流水线ID',
    `stage_order` int(11) NOT NULL DEFAULT 0 COMMENT '阶段顺序(1,2,3...)',
    `stage_type` varchar(32) NOT NULL COMMENT '阶段类型: checkout/build/test/push/approval/deploy',
    `stage_name` varchar(100) NOT NULL COMMENT '阶段名称',
    `status` varchar(32) NOT NULL DEFAULT 'pending' COMMENT '执行状态: pending/running/success/failed/skipped/waiting/aborted',
    
    -- 执行信息
    `started_at` bigint(20) DEFAULT NULL COMMENT '开始时间(Unix时间戳)',
    `finished_at` bigint(20) DEFAULT NULL COMMENT '结束时间(Unix时间戳)',
    `duration_sec` int(11) DEFAULT 0 COMMENT '执行时长(秒)',
    `logs` longtext COMMENT '阶段执行日志',
    
    -- Jenkins 构建信息（适用于 build/test/push 类型）
    `jenkins_stage_id` varchar(100) DEFAULT NULL COMMENT 'Jenkins 阶段ID',
    
    -- 审批信息（适用于 approval 类型）
    `approval_user_id` bigint(20) DEFAULT NULL COMMENT '审批人ID',
    `approval_comment` text COMMENT '审批评论',
    `approval_decision` varchar(32) DEFAULT NULL COMMENT '审批决定: approved/rejected',
    
    -- 部署信息（适用于 deploy 类型）
    `deploy_cluster_id` bigint(20) DEFAULT NULL COMMENT '目标集群ID',
    `deploy_namespace` varchar(100) DEFAULT NULL COMMENT '目标命名空间',
    `deploy_workload_kind` varchar(50) DEFAULT NULL COMMENT '工作负载类型(Deployment/StatefulSet/DaemonSet)',
    `deploy_workload_name` varchar(100) DEFAULT NULL COMMENT '工作负载名称',
    `deploy_container` varchar(100) DEFAULT NULL COMMENT '容器名称',
    `deploy_image` varchar(500) DEFAULT NULL COMMENT '部署镜像地址',
    `deploy_replicas` int(11) DEFAULT NULL COMMENT '副本数',
    
    -- 错误信息
    `error_message` text COMMENT '错误信息',
    
    -- 元数据
    `created_at` bigint(20) NOT NULL COMMENT '创建时间(Unix时间戳)',
    `modified_at` bigint(20) NOT NULL COMMENT '修改时间(Unix时间戳)',
    
    PRIMARY KEY (`id`),
    KEY `idx_run_id` (`run_id`),
    KEY `idx_pipeline_id` (`pipeline_id`),
    KEY `idx_status` (`status`),
    KEY `idx_stage_type` (`stage_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='流水线阶段执行记录表';

-- 为 cicd_pipeline 表添加阶段配置字段
ALTER TABLE `cicd_pipeline` 
ADD COLUMN IF NOT EXISTS `stage_config` JSON DEFAULT NULL COMMENT '阶段配置(定义流水线包含哪些阶段)';

-- 更新 cicd_pipeline 表的 stages 字段说明
-- stages 字段存储默认阶段模板，stage_config 存储用户自定义配置

-- =============================================================================
-- 阶段类型说明
-- =============================================================================
-- checkout  : 代码检出阶段 - 从 Git 仓库拉取代码
-- build     : 构建阶段 - Jenkins 执行编译/打包
-- test      : 测试阶段 - 执行单元测试/集成测试
-- push      : 推送阶段 - 推送镜像到镜像仓库
-- approval  : 审批阶段 - 人工审批(生产环境必选)
-- deploy    : 部署阶段 - 部署到 K8s 集群

-- =============================================================================
-- 阶段状态说明
-- =============================================================================
-- pending   : 等待执行
-- running   : 执行中
-- success   : 执行成功
-- failed    : 执行失败
-- skipped   : 跳过(条件不满足)
-- waiting   : 等待审批(仅 approval 类型)
-- aborted   : 已中止

-- =============================================================================
-- 默认阶段配置示例 (JSON)
-- =============================================================================
/*
{
  "stages": [
    {"order": 1, "type": "checkout", "name": "代码检出", "enabled": true},
    {"order": 2, "type": "build", "name": "构建", "enabled": true},
    {"order": 3, "type": "test", "name": "测试", "enabled": true},
    {"order": 4, "type": "push", "name": "推送镜像", "enabled": true},
    {"order": 5, "type": "approval", "name": "人工审批", "enabled": true, "config": {"required_approvers": 1}},
    {"order": 6, "type": "deploy", "name": "部署", "enabled": true}
  ]
}
*/
