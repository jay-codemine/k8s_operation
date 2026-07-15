package initialize

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"
	"k8soperation/global"
	"k8soperation/internal/app/dao"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	"k8soperation/pkg/database"
	"log"
	"time"
)

// SetupDB 初始化DB
/*
SetupDB 函数用于初始化和配置数据库连接
根据全局配置中的数据库类型设置相应的数据库连接
目前支持 MySQL 数据库
返回值: error - 如果连接或配置过程中出现错误则返回错误信息
*/
func SetupDB() error {
	// 拼接 DSN，加上超时参数（防止连不通时卡很久）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&collation=utf8mb4_general_ci&parseTime=%t&loc=Local&timeout=1s&readTimeout=2s&writeTimeout=2s",
		global.DatabaseSetting.Username,
		global.DatabaseSetting.Password,
		global.DatabaseSetting.Host,
		global.DatabaseSetting.Port,
		global.DatabaseSetting.DBName,
		global.DatabaseSetting.Charset,
		global.DatabaseSetting.ParseTime,
	)

	// 创建 gorm dialector
	dbConfig := mysql.New(mysql.Config{DSN: dsn})

	// 连接数据库
	var err error
	global.DB, global.SQLDB, err = database.Connect(dbConfig, logger.Default.LogMode(logger.Silent))
	if err != nil {
		return fmt.Errorf("connect db failed: %w", err)
	}

	// 连接池设置
	global.SQLDB.SetMaxOpenConns(global.DatabaseSetting.MaxOpenConns)
	global.SQLDB.SetMaxIdleConns(global.DatabaseSetting.MaxIdleConns)
	global.SQLDB.SetConnMaxLifetime(time.Duration(global.DatabaseSetting.MaxLifeSeconds) * time.Second)

	// 快速 Ping 测试连接，最多等 1 秒
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := global.SQLDB.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping failed: %w", err)
	}

	// 自动迁移表结构
	if err := autoMigrateTables(); err != nil {
		return fmt.Errorf("auto migrate tables failed: %w", err)
	}

	// 初始化默认数据
	if err := initDefaultData(); err != nil {
		return fmt.Errorf("init default data failed: %w", err)
	}

	return nil
}

// autoMigrateTables 自动迁移表结构（仅当 SkipAutoMigrate=false 时执行）
// 生产环境已通过 init.sql 初始化完成，建议设为 true 跳过
func autoMigrateTables() error {
	if global.AppSetting != nil && global.AppSetting.SkipAutoMigrate {
		log.Println("[DB] SkipAutoMigrate=true, 跳过自动建表（依赖 init.sql 已完成初始化）")
		return nil
	}

	// 平台基础表
	if err := global.DB.AutoMigrate(
		&models.PlatformSettings{},
		&models.AppStoreApp{},
		&models.AppStoreInstall{},
		&models.AppStoreComponent{},
		&models.CicdApproval{},
		&models.CicdBuildAgent{},
		&models.CicdEnvironment{},
	); err != nil {
		return fmt.Errorf("migrate base tables: %w", err)
	}

	// RBAC 表（v2 三域六角色：自动补齐 scope_platform/scope_cluster/scope_cicd/access_level 等新字段）
	if err := global.DB.AutoMigrate(
		&models.SysRole{},
		&models.SysPermission{},
		&models.SysUserCluster{},
	); err != nil {
		log.Printf("[AutoMigrate] RBAC 表迁移失败: %v", err)
		// 不 return，允许降级运行
	}

	// cicd_pipeline 表字段补全（不用 AutoMigrate 避免 GORM 与已有 UNIQUE KEY 冲突）
	if err := ensurePipelineColumns(); err != nil {
		log.Printf("[AutoMigrate] cicd_pipeline 字段补全失败: %v", err)
	}

	// monitor_notify_channel 表字段补全（兼容旧版本初始化脚本缺少该列的情况）
	if err := ensureNotifyChannelColumns(); err != nil {
		log.Printf("[AutoMigrate] monitor_notify_channel 字段补全失败: %v", err)
	}

	// cicd_pipeline_run 表字段补全（兼容旧版本初始化脚本缺少该列的情况）
	if err := ensurePipelineRunColumns(); err != nil {
		log.Printf("[AutoMigrate] cicd_pipeline_run 字段补全失败: %v", err)
	}

	// AI 助手模块（逐表迁移，确保每张表都成功）
	aiModels := []struct {
		name  string
		model interface{}
	}{
		{"ai_conversations", &models.AIConversation{}},
		{"ai_messages", &models.AIMessage{}},
		{"ai_approval_requests", &models.AIApprovalRequest{}},
		{"ai_approval_logs", &models.AIApprovalLog{}},
	}
	for _, m := range aiModels {
		if err := global.DB.AutoMigrate(m.model); err != nil {
			log.Printf("[AutoMigrate] 创建表 %s 失败: %v", m.name, err)
			return fmt.Errorf("migrate %s: %w", m.name, err)
		}
	}

	// 监控中心模块
	monitorModels := []struct {
		name  string
		model interface{}
	}{
		{"monitor_datasource", &models.MonitorDatasource{}},
		{"monitor_alert_rule", &models.MonitorAlertRule{}},
		{"monitor_alert_event", &models.MonitorAlertEvent{}},
		{"monitor_notify_channel", &models.MonitorNotifyChannel{}},
		{"monitor_silence_rule", &models.MonitorSilenceRule{}},
		{"monitor_inhibit_rule", &models.MonitorInhibitRule{}},
		{"monitor_aggregate_rule", &models.MonitorAggregateRule{}},
		{"monitor_notify_template", &models.MonitorNotifyTemplate{}},
		{"monitor_notify_route_policy", &models.MonitorNotifyRoutePolicy{}},
	}
	for _, m := range monitorModels {
		if err := global.DB.AutoMigrate(m.model); err != nil {
			log.Printf("[AutoMigrate] 创建表 %s 失败: %v", m.name, err)
			return fmt.Errorf("migrate %s: %w", m.name, err)
		}
	}

	// 审计日志表
	if err := global.DB.AutoMigrate(&models.AuditLog{}); err != nil {
		log.Printf("[AutoMigrate] 创建表 audit_log 失败: %v", err)
		// 不返回错误，允许降级运行
	}

	// 汇总一行（只记总数，有错才刷详情）
	log.Printf("[AutoMigrate] OK (ai: %d, monitor: %d)", len(aiModels), len(monitorModels))

	// AIOps 智能运维模块
	aiopsModels := []struct {
		name  string
		model interface{}
	}{
		{"aiops_analysis_record", &models.AIOpsAnalysisRecord{}},
		{"aiops_inspection_report", &models.AIOpsInspectionReport{}},
	}
	for _, m := range aiopsModels {
		if err := global.DB.AutoMigrate(m.model); err != nil {
			log.Printf("[AutoMigrate] 创建表 %s 失败: %v", m.name, err)
			// 不返回错误，允许降级运行
		}
	}

	return nil
}

// initDefaultData 初始化默认数据
func initDefaultData() error {
	ctx := context.Background()
	d := dao.NewDao(global.DB)

	// 初始化平台设置默认值
	if err := d.PlatformSettingsInitDefaults(ctx); err != nil {
		return fmt.Errorf("init platform settings failed: %w", err)
	}

	// RBAC v2: 回填存量角色的 scope 值（仅当三域均为默认 none 时才触发）
	backfillRBACScopes()

	// ⚠️ 管理员角色自修复（防止admin自锁后无法恢复）
	repairAdminRole()

	// 初始化应用商城种子数据
	svc := services.NewServices()
	if err := svc.AppStoreSeed(ctx); err != nil {
		return fmt.Errorf("init appstore seed data failed: %w", err)
	}

	// 初始化构建探针种子数据（OTEL Java Agent）
	if err := svc.BuildAgentSeedOTEL(ctx); err != nil {
		log.Printf("[InitData] 构建探针种子数据初始化失败: %v", err)
	}

	// 修复因终端编码导致的中文乱码（? 字符）告警规则
	repairCorruptedAlertRules()

	// CICD 细粒度权限种子数据
	seedCICDPermissions()

	// 默认部署环境种子数据
	seedDefaultEnvironments()

	return nil
}

// repairCorruptedAlertRules 修复告警规则中因 PowerShell 编码问题产生的乱码
// 检测规则名中包含连续 `?` 的记录，根据残留 ASCII 字符和分组信息映射到正确的中文名
func repairCorruptedAlertRules() {
	type fixRule struct {
		pattern string // SQL LIKE 模式匹配 name
		group   string // 分组过滤（空字符串表示不限组）
		name    string // 正确的名称
		summary string // 正确的摘要
	}

	fixes := []fixRule{
		// infrastructure 组
		{"CPU%?%", "infrastructure", "节点CPU使用率过高", "节点 {{ $labels.instance }} CPU使用率超过80%"},
		{"%CPU%?%?%", "infrastructure", "节点CPU使用率过高", "节点 {{ $labels.instance }} CPU使用率超过80%"},
		{"?%NotReady%", "infrastructure", "节点NotReady", "节点 {{ $labels.node }} 状态为 NotReady"},
		{"?%IO%?%", "infrastructure", "网络IO流量异常", "节点 {{ $labels.instance }} 入站流量超过100MB/s"},
		{"%?%IO%?%", "infrastructure", "网络IO流量异常", "节点 {{ $labels.instance }} 入站流量超过100MB/s"},
		// kubernetes 组
		{"Pod%?%Ready%?%", "kubernetes", "Pod非Ready状态", "Pod {{ $labels.namespace }}/{{ $labels.pod }} 状态异常"},
		{"Pod%?%Ready%", "kubernetes", "Pod非Ready状态", "Pod {{ $labels.namespace }}/{{ $labels.pod }} 状态异常"},
		{"Deployment%?%", "kubernetes", "Deployment副本不足", "Deployment {{ $labels.namespace }}/{{ $labels.deployment }} 存在不可用副本"},
		{"K8s Job%?%", "kubernetes", "K8s Job失败", "K8s Job {{ $labels.namespace }}/{{ $labels.job_name }} 执行失败"},
		{"K8s%?%Job%?%", "kubernetes", "K8s Job失败", "K8s Job {{ $labels.namespace }}/{{ $labels.job_name }} 执行失败"},
		{"DaemonSet%?%", "kubernetes", "DaemonSet滚动更新异常", "DaemonSet {{ $labels.namespace }}/{{ $labels.daemonset }} 滚动更新卡住"},
		{"StatefulSet%?%", "kubernetes", "StatefulSet副本异常", "StatefulSet {{ $labels.namespace }}/{{ $labels.statefulset }} 存在不可用副本"},
		{"HPA%?%", "kubernetes", "HPA扩缩容异常", "HPA {{ $labels.namespace }}/{{ $labels.horizontalpodautoscaler }} 无法完成扩缩容"},
		{"CronJob%?%", "kubernetes", "CronJob执行异常", "CronJob {{ $labels.namespace }}/{{ $labels.cronjob }} 执行超时或失败"},
		{"PVC%?%", "kubernetes", "PVC存储异常", "PVC {{ $labels.namespace }}/{{ $labels.persistentvolumeclaim }} 状态异常"},
		// application 组
		{"API%?%", "application", "API延迟过高", "接口 {{ $labels.handler }} P95延迟超过2秒"},
		{"API%?%?%", "application", "API延迟过高", "接口 {{ $labels.handler }} P95延迟超过2秒"},
		{"HTTP%?%5%", "application", "HTTP 5xx错误率过高", "HTTP 5xx 错误率超过5%"},
		{"%?%?%?%?%?%?%?%", "", "", ""}, // 占位，通用兜底逻辑后面处理
	}

	// 特殊情况：全是 ? 的规则（无 ASCII 残留），需要按组+表达式匹配
	type exprFix struct {
		exprLike string
		group    string
		name     string
		summary  string
	}
	exprFixes := []exprFix{
		{"%node_memory%", "infrastructure", "节点内存使用率过高", "节点 {{ $labels.instance }} 内存使用率超过85%"},
		{"%node_filesystem%", "infrastructure", "磁盘使用率告警", "节点 {{ $labels.instance }} 磁盘使用率超过90%"},
		{"%node_filesystem_avail%", "infrastructure", "磁盘可用空间不足", "节点 {{ $labels.instance }} 磁盘可用空间不足10%"},
		{"%network_receive%", "infrastructure", "网络入站流量异常", "节点 {{ $labels.instance }} 入站流量超过100MB/s"},
		{"%network_transmit%", "infrastructure", "网络出站流量异常", "节点 {{ $labels.instance }} 出站流量超过100MB/s"},
		{"%node_load%", "infrastructure", "节点负载过高", "节点 {{ $labels.instance }} 系统负载超过阈值"},
		{"%kube_pod_container_status_waiting%", "kubernetes", "容器等待状态异常", "容器 {{ $labels.namespace }}/{{ $labels.pod }} 处于 Waiting 状态"},
		{"%kube_pod_container_status_terminated%", "kubernetes", "容器异常终止", "容器 {{ $labels.namespace }}/{{ $labels.pod }} 异常终止"},
		{"%kube_deployment_status_replicas_unavailable%", "kubernetes", "Deployment副本不足", "Deployment {{ $labels.namespace }}/{{ $labels.deployment }} 存在不可用副本"},
		{"%kube_daemonset_status%", "kubernetes", "DaemonSet状态异常", "DaemonSet {{ $labels.namespace }}/{{ $labels.daemonset }} 状态异常"},
		{"%kube_statefulset%", "kubernetes", "StatefulSet副本异常", "StatefulSet {{ $labels.namespace }}/{{ $labels.statefulset }} 存在不可用副本"},
		{"%kube_job_status_failed%", "kubernetes", "K8s Job失败", "K8s Job {{ $labels.namespace }}/{{ $labels.job_name }} 执行失败"},
		{"%http_request%5%", "application", "HTTP 5xx错误率过高", "HTTP 5xx 错误率超过5%"},
		{"%request_duration%", "application", "API延迟过高", "接口 {{ $labels.handler }} P95延迟超过2秒"},
		{"%container_memory%", "kubernetes", "容器内存使用率过高", "容器 {{ $labels.namespace }}/{{ $labels.pod }} 内存使用率超过90%"},
		{"%container_cpu%", "kubernetes", "容器CPU使用率过高", "容器 {{ $labels.namespace }}/{{ $labels.pod }} CPU使用率超过80%"},
	}

	var repaired int

	// 第一轮：按 name 的 ASCII 残留 + 分组修复（跳过占位项）
	for _, fix := range fixes {
		if fix.name == "" {
			continue
		}
		sql := "UPDATE `monitor_alert_rule` SET `name` = ?, `summary` = ? WHERE `name` LIKE ? AND `name` LIKE '%?%' AND `is_del` = 0"
		args := []interface{}{fix.name, fix.summary, fix.pattern}
		if fix.group != "" {
			sql += " AND `group` = ?"
			args = append(args, fix.group)
		}
		result := global.DB.Exec(sql, args...)
		if result.RowsAffected > 0 {
			repaired += int(result.RowsAffected)
		}
	}

	// 第二轮：纯 ? 开头的规则（name 中无法识别 ASCII），通过 expr 内容匹配
	for _, fix := range exprFixes {
		result := global.DB.Exec(
			"UPDATE `monitor_alert_rule` SET `name` = ?, `summary` = ? WHERE `name` LIKE '%?%' AND `group` = ? AND `expr` LIKE ? AND `is_del` = 0",
			fix.name, fix.summary, fix.group, fix.exprLike,
		)
		if result.RowsAffected > 0 {
			repaired += int(result.RowsAffected)
		}
	}

	// 第三轮：仍然有 ? 的 Pod 类规则（特征匹配）
	global.DB.Exec(
		"UPDATE `monitor_alert_rule` SET `name` = 'Pod频繁重启', `summary` = 'Pod {{ $labels.namespace }}/{{ $labels.pod }} 频繁重启' WHERE `name` LIKE '%Pod%?%' AND `group` = 'kubernetes' AND `expr` LIKE '%restarts_total%' AND `is_del` = 0",
	)
	global.DB.Exec(
		"UPDATE `monitor_alert_rule` SET `name` = 'HTTP 5xx错误率过高', `summary` = 'HTTP 5xx 错误率超过5%' WHERE `name` LIKE '%?%' AND `group` = 'application' AND `expr` LIKE '%5..%' AND `is_del` = 0",
	)

	// 第四轮：通用兜底 - 剩余含 ? 的规则，根据 group 赋予通用名称
	type groupFallback struct {
		group   string
		name    string
		summary string
	}
	fallbacks := []groupFallback{
		{"infrastructure", "基础设施告警", "基础设施资源异常，请排查相关指标"},
		{"kubernetes", "Kubernetes资源告警", "K8s 资源状态异常，请排查相关工作负载"},
		{"application", "应用层告警", "应用指标异常，请排查相关服务"},
	}
	for _, fb := range fallbacks {
		result := global.DB.Exec(
			"UPDATE `monitor_alert_rule` SET `name` = ?, `summary` = ? WHERE `name` LIKE '%?%' AND `group` = ? AND `is_del` = 0",
			fb.name, fb.summary, fb.group,
		)
		if result.RowsAffected > 0 {
			repaired += int(result.RowsAffected)
		}
	}

	// 修复告警事件中的乱码（同步规则名到已有事件）
	global.DB.Exec(
		"UPDATE `monitor_alert_event` e INNER JOIN `monitor_alert_rule` r ON e.rule_id = r.id SET e.rule_name = r.name, e.summary = r.summary WHERE e.rule_name LIKE '%?%'",
	)

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// 修复 告警降噪 模块中的乱码（静默规则/抑制规则/聚合规则）
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

	// 静默规则: name + comment 修复
	r1 := global.DB.Exec(
		"UPDATE `monitor_silence_rule` SET `name` = '夜间维护静默', `comment` = '凌晨0-8点静默warning/info告警' WHERE `name` LIKE '%?%' AND `is_del` = 0",
	)
	if r1.RowsAffected > 0 {
		repaired += int(r1.RowsAffected)
	}

	// 抑制规则: name + description 修复
	r2 := global.DB.Exec(
		"UPDATE `monitor_inhibit_rule` SET `name` = 'Critical抑制Warning', `description` = '当同一实例有critical告警时，抑制其warning告警' WHERE `name` LIKE '%?%' AND `is_del` = 0",
	)
	if r2.RowsAffected > 0 {
		repaired += int(r2.RowsAffected)
	}

	// 聚合规则: name 修复
	r3 := global.DB.Exec(
		"UPDATE `monitor_aggregate_rule` SET `name` = '按实例聚合告警' WHERE `name` LIKE '%?%' AND `is_del` = 0",
	)
	if r3.RowsAffected > 0 {
		repaired += int(r3.RowsAffected)
	}

	if repaired > 0 {
		log.Printf("[DataRepair] 修复 %d 条中文乱码记录（告警规则+降噪规则）", repaired)
	}
}

// ensurePipelineColumns 检查并补全 cicd_pipeline 表缺失的列
// 不使用 AutoMigrate 是因为 GORM 会尝试将 varchar 改为 longtext，与 UNIQUE KEY 冲突
func ensurePipelineColumns() error {

	// 检查 cicd_pipeline 表是否存在
	var count int64
	global.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'cicd_pipeline'").Scan(&count)
	if count == 0 {
		return nil // 表不存在，跳过（全新安装由 SQL 初始化脚本负责）
	}

	// 所有可能缺失的列（模型新增但旧表可能没有的）
	type colDef struct {
		name string
		sql  string
	}
	columns := []colDef{
		{"language_type", "ALTER TABLE `cicd_pipeline` ADD COLUMN `language_type` varchar(20) NOT NULL DEFAULT 'custom' COMMENT '语言类型' AFTER `jenkins_credential_id`"},
		{"enable_sonar", "ALTER TABLE `cicd_pipeline` ADD COLUMN `enable_sonar` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用SonarQube代码扫描' AFTER `require_approval`"},
		{"last_deploy_image", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_image` varchar(500) DEFAULT '' COMMENT '最新部署镜像' AFTER `enable_sonar`"},
		{"last_deploy_digest", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_digest` varchar(100) DEFAULT '' COMMENT '镜像摘要' AFTER `last_deploy_image`"},
		{"last_deploy_time", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_time` bigint DEFAULT NULL COMMENT '最新部署时间' AFTER `last_deploy_digest`"},
		{"last_deploy_status", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_status` varchar(32) DEFAULT '' COMMENT '最新部署状态' AFTER `last_deploy_time`"},
		{"last_deploy_version", "ALTER TABLE `cicd_pipeline` ADD COLUMN `last_deploy_version` varchar(100) DEFAULT '' COMMENT '最新部署版本' AFTER `last_deploy_status`"},
	}

	for _, col := range columns {
		var exists int64
		global.DB.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'cicd_pipeline' AND column_name = ?", col.name).Scan(&exists)
		if exists == 0 {
			if err := global.DB.Exec(col.sql).Error; err != nil {
				return fmt.Errorf("add column %s: %w", col.name, err)
			}
			log.Printf("[AutoMigrate] cicd_pipeline 补全列: %s", col.name)
		}
	}
	return nil
}

// ensurePipelineRunColumns 检查并补全 cicd_pipeline_run 表缺失的列
func ensurePipelineRunColumns() error {
	var count int64
	global.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'cicd_pipeline_run'").Scan(&count)
	if count == 0 {
		return nil
	}
	type colDef struct {
		name string
		sql  string
	}
	columns := []colDef{
		{"workflow_name", "ALTER TABLE `cicd_pipeline_run` ADD COLUMN `workflow_name` varchar(200) NOT NULL DEFAULT '' COMMENT 'workflow name' AFTER `jenkins_build_url`"},
		{"argo_app_name", "ALTER TABLE `cicd_pipeline_run` ADD COLUMN `argo_app_name` varchar(200) NOT NULL DEFAULT '' COMMENT 'argo app name' AFTER `workflow_name`"},
		{"sync_revision", "ALTER TABLE `cicd_pipeline_run` ADD COLUMN `sync_revision` varchar(100) NOT NULL DEFAULT '' COMMENT 'sync revision' AFTER `argo_app_name`"},
		{"sync_status", "ALTER TABLE `cicd_pipeline_run` ADD COLUMN `sync_status` varchar(50) NOT NULL DEFAULT '' COMMENT 'sync status' AFTER `sync_revision`"},
	}
	for _, col := range columns {
		var exists int64
		global.DB.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'cicd_pipeline_run' AND column_name = ?", col.name).Scan(&exists)
		if exists == 0 {
			if err := global.DB.Exec(col.sql).Error; err != nil {
				return fmt.Errorf("add column %s: %w", col.name, err)
			}
			log.Printf("[AutoMigrate] cicd_pipeline_run 补全列: %s", col.name)
		}
	}
	return nil
}

// ensureNotifyChannelColumns 检查并补全 monitor_notify_channel 表缺失的列
// 兼容旧版本初始化脚本中未包含 security_keyword 列的情况
func ensureNotifyChannelColumns() error {
	// 检查表是否存在
	var count int64
	global.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'monitor_notify_channel'").Scan(&count)
	if count == 0 {
		return nil // 表不存在，跳过（全新安装由 SQL 初始化脚本负责）
	}

	type colDef struct {
		name string
		sql  string
	}
	columns := []colDef{
		{
			"security_keyword",
			"ALTER TABLE `monitor_notify_channel` ADD COLUMN `security_keyword` varchar(100) NOT NULL DEFAULT '' COMMENT '钉钉安全关键字（多个用逗号分隔）' AFTER `secret`",
		},
	}

	for _, col := range columns {
		var exists int64
		global.DB.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'monitor_notify_channel' AND column_name = ?", col.name).Scan(&exists)
		if exists == 0 {
			if err := global.DB.Exec(col.sql).Error; err != nil {
				return fmt.Errorf("add column %s: %w", col.name, err)
			}
			log.Printf("[AutoMigrate] monitor_notify_channel 补全列: %s", col.name)
		}
	}
	return nil
}

// backfillRBACScopes 回填存量角色的 scope 值（仅当三域均为默认 none 时才触发，不会覆盖已配置的值）
func backfillRBACScopes() {
	db := global.DB

	// 检查 sys_role 表是否存在且有 scope_platform 列
	var colCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'sys_role' AND column_name = 'scope_platform'").Scan(&colCount)
	if colCount == 0 {
		return // 列不存在，跳过
	}

	// scope 回填规则（WHERE 条件保证只对未配置的角色生效）
	scoped := "scope_platform = 'none' AND scope_cluster = 'none' AND scope_cicd = 'none'"
	updates := []struct {
		roleType string
		set      string
	}{
		{"super_admin", "scope_platform='admin', scope_cluster='admin', scope_cicd='admin'"},
		{"platform_admin", "scope_platform='admin', scope_cluster='read', scope_cicd='read'"},
		{"devops", "scope_platform='read', scope_cluster='admin', scope_cicd='admin'"},
		{"developer", "scope_platform='none', scope_cluster='write', scope_cicd='write'"},
		{"tester", "scope_platform='none', scope_cluster='read', scope_cicd='write'"},
		{"viewer", "scope_platform='read', scope_cluster='read', scope_cicd='read'"},
	}

	total := int64(0)
	for _, u := range updates {
		result := db.Exec(
			fmt.Sprintf("UPDATE sys_role SET %s WHERE role_type = ? AND %s", u.set, scoped),
			u.roleType,
		)
		total += result.RowsAffected
	}

	// cluster_admin → devops 映射（存量旧角色类型）
	result := db.Exec(
		fmt.Sprintf("UPDATE sys_role SET role_type='devops', scope_platform='read', scope_cluster='admin', scope_cicd='admin' WHERE role_type = 'cluster_admin' AND %s", scoped),
	)
	total += result.RowsAffected

	if total > 0 {
		log.Printf("[InitData] RBAC scope 回填完成，影响 %d 个角色", total)
	}
}

// repairAdminRole 确保 admin 用户（user_id=1）始终拥有 super_admin 角色
// 防止管理员意外自降权后无法恢复的情况
func repairAdminRole() {
	db := global.DB

	// 检查 sys_user_role 表是否存在
	var tableCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'sys_user_role'").Scan(&tableCount)
	if tableCount == 0 {
		return
	}

	// 检查 sys_role 中 role_type='super_admin' 的角色 ID
	var superAdminRoleID int64
	db.Raw("SELECT id FROM sys_role WHERE role_type = 'super_admin' AND is_del = 0 LIMIT 1").Scan(&superAdminRoleID)
	if superAdminRoleID == 0 {
		return // 超级管理员角色不存在，跳过
	}

	// 检查 admin 用户（id=1）是否已拥有 super_admin 角色
	var hasAdmin int64
	db.Raw("SELECT COUNT(*) FROM sys_user_role WHERE user_id = 1 AND role_id = ?", superAdminRoleID).Scan(&hasAdmin)
	if hasAdmin > 0 {
		return // 已有管理员角色，无需修复
	}

	// 修复：为 admin 用户追加 super_admin 角色（不删除已有角色，只追加）
	now := time.Now().Unix()
	result := db.Exec(
		"INSERT IGNORE INTO sys_user_role (user_id, role_id, created_at, created_by) VALUES (1, ?, ?, 0)",
		superAdminRoleID, now,
	)
	if result.RowsAffected > 0 {
		log.Printf("[InitData] ⚠️ admin 用户角色已修复：补充 super_admin 角色 (role_id=%d)", superAdminRoleID)
	}
}

// seedDefaultEnvironments 初始化默认部署环境（dev/staging/prod）
// 幂等操作：已存在的环境不会重复插入
func seedDefaultEnvironments() {
	db := global.DB

	// 检查 cicd_environment 表是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'cicd_environment'").Scan(&count)
	if count == 0 {
		return
	}

	// 检查是否已有环境数据
	var existCount int64
	db.Raw("SELECT COUNT(*) FROM cicd_environment WHERE is_del = 0").Scan(&existCount)
	if existCount > 0 {
		return // 已有环境数据，跳过
	}

	now := time.Now().Unix()

	envs := []struct {
		Name            string
		DisplayName     string
		Color           string
		SortOrder       int
		RequireApproval int
		ApprovalLevels  string
	}{
		{"dev", "开发环境", "#52c41a", 1, 0, "[]"},
		{"test", "测试环境", "#1677ff", 2, 0, "[]"},
		{"staging", "预发环境", "#faad14", 3, 1, `[{"label":"测试负责人审批","approver_type":"role","approver_value":"tester"}]`},
		{"prod", "生产环境", "#f5222d", 4, 1, `[{"label":"运维经理审批","approver_type":"role","approver_value":"devops"},{"label":"技术总监审批","approver_type":"role","approver_value":"platform_admin"}]`},
	}

	inserted := 0
	for _, env := range envs {
		var exists int64
		db.Raw("SELECT COUNT(*) FROM cicd_environment WHERE name = ? AND is_del = 0", env.Name).Scan(&exists)
		if exists > 0 {
			continue
		}
		db.Exec(
			"INSERT INTO cicd_environment (name, display_name, cluster_id, namespace, color, sort_order, require_approval, approval_levels, is_del, created_at, modified_at) VALUES (?, ?, 0, ?, ?, ?, ?, ?, 0, ?, ?)",
			env.Name, env.DisplayName, env.Name, env.Color, env.SortOrder, env.RequireApproval, env.ApprovalLevels, now, now,
		)
		inserted++
	}

	if inserted > 0 {
		log.Printf("[InitData] 默认环境初始化完成，新增 %d 个环境（dev/test/staging/prod）", inserted)
	}
}

// seedCICDPermissions 初始化 CICD 细粒度权限种子数据
// 幂等操作：已存在的权限不会重复插入
func seedCICDPermissions() {
	db := global.DB

	// 检查 sys_permission 表是否存在
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'sys_permission'").Scan(&count)
	if count == 0 {
		return
	}

	// 细粒度 CICD 权限定义（4个标签分组）
	type permDef struct {
		Name         string
		DisplayName  string
		Description  string
		ResourceType string
		Action       string
		Tag          string
		Path         string
		SortOrder    int
	}

	perms := []permDef{
		// 标签: 流水线管理
		{"cicd:pipeline:view", "查看流水线", "查看流水线列表和详情", "pipeline", "view", "流水线管理", "/cicd/pipeline/view", 100},
		{"cicd:pipeline:create", "创建流水线", "创建新的流水线配置", "pipeline", "create", "流水线管理", "/cicd/pipeline/create", 101},
		{"cicd:pipeline:edit", "编辑流水线", "修改流水线配置参数", "pipeline", "update", "流水线管理", "/cicd/pipeline/edit", 102},
		{"cicd:pipeline:delete", "删除流水线", "删除流水线配置", "pipeline", "delete", "流水线管理", "/cicd/pipeline/delete", 103},
		{"cicd:pipeline:run", "运行流水线", "触发/运行流水线构建", "pipeline", "exec", "流水线管理", "/cicd/pipeline/run", 104},

		// 标签: 构建与部署
		{"cicd:build:view", "查看构建记录", "查看构建历史和日志", "build", "view", "构建与部署", "/cicd/build/view", 110},
		{"cicd:build:trigger", "触发构建", "手动触发代码构建", "build", "exec", "构建与部署", "/cicd/build/trigger", 111},
		{"cicd:build:cancel", "取消构建", "取消正在进行的构建", "build", "exec", "构建与部署", "/cicd/build/cancel", 112},
		{"cicd:deploy:dev", "部署开发环境", "部署到开发环境", "deploy", "exec", "构建与部署", "/cicd/deploy/dev", 113},
		{"cicd:deploy:test", "部署测试环境", "部署到测试环境", "deploy", "exec", "构建与部署", "/cicd/deploy/test", 114},
		{"cicd:deploy:prod", "部署生产环境", "部署到生产环境", "deploy", "exec", "构建与部署", "/cicd/deploy/prod", 115},
		{"cicd:deploy:rollback", "回滚部署", "回滚到上一版本", "deploy", "exec", "构建与部署", "/cicd/deploy/rollback", 116},

		// 标签: 制品与镜像
		{"cicd:artifact:view", "查看制品", "查看制品库和镜像列表", "artifact", "view", "制品与镜像", "/cicd/artifact/view", 120},
		{"cicd:artifact:upload", "上传制品", "上传制品到制品库", "artifact", "create", "制品与镜像", "/cicd/artifact/upload", 121},
		{"cicd:artifact:delete", "删除制品", "从制品库删除制品", "artifact", "delete", "制品与镜像", "/cicd/artifact/delete", 122},
		{"cicd:image:manage", "镜像管理", "管理镜像仓库（推送/删除/清理）", "image", "manage", "制品与镜像", "/cicd/image/manage", 123},

		// 标签: 审批与管理
		{"cicd:approval:view", "查看审批", "查看审批记录和详情", "approval", "view", "审批与管理", "/cicd/approval/view", 130},
		{"cicd:approval:action", "执行审批", "通过或拒绝审批申请", "approval", "exec", "审批与管理", "/cicd/approval/action", 131},
		{"cicd:approval:manage", "管理审批策略", "配置审批流程和审批人", "approval", "manage", "审批与管理", "/cicd/approval/manage", 132},
		{"cicd:environment:manage", "环境管理", "创建/编辑/删除环境配置", "environment", "manage", "审批与管理", "/cicd/environment/manage", 133},
		{"cicd:template:manage", "模板管理", "管理流水线模板", "template", "manage", "审批与管理", "/cicd/template/manage", 134},

		// 标签: 平台运维
		{"cicd:resource:manage", "资源模板管理", "管理资源模板和环境规则", "resource", "manage", "平台运维", "/cicd/resource/manage", 140},
		{"cicd:agent:manage", "构建探针管理", "上传/更新/删除构建探针", "agent", "manage", "平台运维", "/cicd/agent/manage", 141},
	}

	now := uint64(time.Now().Unix())
	inserted := 0
	for _, p := range perms {
		// 幂等：已存在则跳过
		var exists int64
		db.Raw("SELECT COUNT(*) FROM sys_permission WHERE name = ?", p.Name).Scan(&exists)
		if exists > 0 {
			// 更新 tag 字段（存量数据可能没有 tag）
			db.Exec("UPDATE sys_permission SET tag = ? WHERE name = ? AND (tag = '' OR tag IS NULL)", p.Tag, p.Name)
			continue
		}
		db.Exec(
			"INSERT INTO sys_permission (name, display_name, description, scope, resource_type, action, tag, parent_id, path, sort_order, created_at, modified_at) VALUES (?, ?, ?, 'cicd', ?, ?, ?, 0, ?, ?, ?, ?)",
			p.Name, p.DisplayName, p.Description, p.ResourceType, p.Action, p.Tag, p.Path, p.SortOrder, now, now,
		)
		inserted++
	}

	// 为旧的粗粒度权限补充 tag
	db.Exec("UPDATE sys_permission SET tag = '流水线管理' WHERE name = 'cicd:pipeline' AND (tag = '' OR tag IS NULL)")
	db.Exec("UPDATE sys_permission SET tag = '制品与镜像' WHERE name = 'cicd:artifact' AND (tag = '' OR tag IS NULL)")

	// 为超级管理员角色自动分配新权限
	if inserted > 0 {
		db.Exec(
			"INSERT IGNORE INTO sys_role_permission (role_id, permission_id, created_at) SELECT 1, id, ? FROM sys_permission WHERE scope = 'cicd' AND id NOT IN (SELECT permission_id FROM sys_role_permission WHERE role_id = 1)",
			now,
		)
		// 为 devops 角色分配所有 CICD 权限
		db.Exec(
			"INSERT IGNORE INTO sys_role_permission (role_id, permission_id, created_at) SELECT 3, id, ? FROM sys_permission WHERE scope = 'cicd' AND id NOT IN (SELECT permission_id FROM sys_role_permission WHERE role_id = 3)",
			now,
		)
		// 为 developer 角色分配基础 CICD 权限（查看+运行+构建+部署开发/测试+查看制品）
		devPerms := []string{
			"cicd:pipeline:view", "cicd:pipeline:create", "cicd:pipeline:edit", "cicd:pipeline:run",
			"cicd:build:view", "cicd:build:trigger", "cicd:build:cancel",
			"cicd:deploy:dev", "cicd:deploy:test", "cicd:deploy:rollback",
			"cicd:artifact:view", "cicd:approval:view",
		}
		for _, pName := range devPerms {
			db.Exec(
				"INSERT IGNORE INTO sys_role_permission (role_id, permission_id, created_at) SELECT 4, id, ? FROM sys_permission WHERE name = ? AND id NOT IN (SELECT permission_id FROM sys_role_permission WHERE role_id = 4)",
				now, pName,
			)
		}
		// 为 tester 角色分配查看+部署测试权限
		testerPerms := []string{
			"cicd:pipeline:view", "cicd:pipeline:run",
			"cicd:build:view", "cicd:build:trigger",
			"cicd:deploy:dev", "cicd:deploy:test",
			"cicd:artifact:view", "cicd:approval:view",
		}
		for _, pName := range testerPerms {
			db.Exec(
				"INSERT IGNORE INTO sys_role_permission (role_id, permission_id, created_at) SELECT 5, id, ? FROM sys_permission WHERE name = ? AND id NOT IN (SELECT permission_id FROM sys_role_permission WHERE role_id = 5)",
				now, pName,
			)
		}
		log.Printf("[InitData] CICD 细粒度权限初始化完成，新增 %d 条权限", inserted)
	}
}
