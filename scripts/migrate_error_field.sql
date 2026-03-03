-- 添加错误信息字段到 cicd_pipeline_run 表
ALTER TABLE `cicd_pipeline_run` ADD COLUMN `error_message` TEXT COMMENT '错误信息';