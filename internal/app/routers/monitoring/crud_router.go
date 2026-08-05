package monitoring

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
)

// MonitorCRUDRouter 监控 CRUD 路由
type MonitorCRUDRouter struct {
	svc *services.MonitorCRUDService
}

// NewMonitorCRUDRouter 创建监控 CRUD 路由
func NewMonitorCRUDRouter() *MonitorCRUDRouter {
	return &MonitorCRUDRouter{
		svc: services.NewMonitorCRUDService(),
	}
}

// Inject 注入路由
func (r *MonitorCRUDRouter) Inject(router *gin.RouterGroup) {
	// 数据源管理
	ds := router.Group("/datasource")
	{
		ds.GET("", r.ListDatasources)
		ds.GET("/:id", r.GetDatasource)
		ds.POST("", r.CreateDatasource)
		ds.PUT("/:id", r.UpdateDatasource)
		ds.DELETE("/:id", r.DeleteDatasource)
		ds.POST("/test", r.TestDatasourceConnection)
		ds.POST("/:id/test", r.TestDatasourceByID)
	}

	// 告警规则管理
	rule := router.Group("/alert-rule")
	{
		rule.GET("", r.ListAlertRules)
		rule.GET("/groups", r.GetAlertRuleGroups)
		rule.GET("/:id", r.GetAlertRule)
		rule.POST("", r.CreateAlertRule)
		rule.PUT("/:id", r.UpdateAlertRule)
		rule.DELETE("/:id", r.DeleteAlertRule)
		rule.PUT("/:id/toggle", r.ToggleAlertRule)
		rule.POST("/import-yaml", r.ImportAlertRulesYAML)
		rule.GET("/export-yaml", r.ExportAlertRulesYAML)
		rule.POST("/batch-bind-channels", r.BatchBindChannels)
		rule.POST("/batch-delete", r.BatchDeleteAlertRules)
		rule.POST("/batch-update", r.BatchUpdateAlertRules)
	}

	// 告警事件
	event := router.Group("/alert-event")
	{
		event.GET("", r.ListAlertEvents)
		event.GET("/stats", r.GetAlertStats)
		event.GET("/:id", r.GetAlertEvent)
		event.PUT("/:id/ack", r.AckAlertEvent)
		event.PUT("/:id/resolve", r.ResolveAlertEvent)
		event.POST("/batch-delete", r.BatchDeleteAlertEvents)
	}

	// 通知渠道管理
	notify := router.Group("/notify-channel")
	{
		notify.GET("", r.ListNotifyChannels)
		notify.GET("/:id", r.GetNotifyChannel)
		notify.POST("", r.CreateNotifyChannel)
		notify.PUT("/:id", r.UpdateNotifyChannel)
		notify.DELETE("/:id", r.DeleteNotifyChannel)
		notify.POST("/:id/test", r.TestNotifyChannel)
		notify.POST("/batch-delete", r.BatchDeleteNotifyChannels)
		notify.POST("/batch-update", r.BatchUpdateNotifyChannels)
	}

	// 静默规则管理
	silence := router.Group("/silence-rule")
	{
		silence.GET("", r.ListSilenceRules)
		silence.GET("/:id", r.GetSilenceRule)
		silence.POST("", r.CreateSilenceRule)
		silence.PUT("/:id", r.UpdateSilenceRule)
		silence.DELETE("/:id", r.DeleteSilenceRule)
		silence.POST("/batch-delete", r.BatchDeleteSilenceRules)
	}

	// 抑制规则管理
	inhibit := router.Group("/inhibit-rule")
	{
		inhibit.GET("", r.ListInhibitRules)
		inhibit.POST("", r.CreateInhibitRule)
		inhibit.PUT("/:id", r.UpdateInhibitRule)
		inhibit.DELETE("/:id", r.DeleteInhibitRule)
	}

	// 聚合规则管理
	aggregate := router.Group("/aggregate-rule")
	{
		aggregate.GET("", r.ListAggregateRules)
		aggregate.POST("", r.CreateAggregateRule)
		aggregate.PUT("/:id", r.UpdateAggregateRule)
		aggregate.DELETE("/:id", r.DeleteAggregateRule)
	}

	// 通知模板管理
	tpl := router.Group("/notify-template")
	{
		tpl.GET("", r.ListNotifyTemplates)
		tpl.GET("/:id", r.GetNotifyTemplate)
		tpl.POST("", r.CreateNotifyTemplate)
		tpl.PUT("/:id", r.UpdateNotifyTemplate)
		tpl.DELETE("/:id", r.DeleteNotifyTemplate)
		tpl.POST("/preview", r.PreviewNotifyTemplate)
		tpl.PUT("/:id/default", r.SetDefaultTemplate)
	}

	// 通知路由策略管理（大厂级多群自动路由）
	route := router.Group("/notify-route")
	{
		route.GET("", r.ListRoutePolicies)
		route.GET("/:id", r.GetRoutePolicy)
		route.POST("", r.CreateRoutePolicy)
		route.PUT("/:id", r.UpdateRoutePolicy)
		route.DELETE("/:id", r.DeleteRoutePolicy)
	}
}

// ==================== 数据源 ====================

func (r *MonitorCRUDRouter) ListDatasources(c *gin.Context) {
	var req services.DatasourceListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	result, err := r.svc.ListDatasources(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (r *MonitorCRUDRouter) GetDatasource(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ds, err := r.svc.GetDatasource(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "数据源不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": ds})
}

func (r *MonitorCRUDRouter) CreateDatasource(c *gin.Context) {
	var ds models.MonitorDatasource
	if err := c.ShouldBindJSON(&ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	if err := r.svc.CreateDatasource(c.Request.Context(), &ds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": ds, "msg": "创建成功"})
}

func (r *MonitorCRUDRouter) UpdateDatasource(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var ds models.MonitorDatasource
	if err := c.ShouldBindJSON(&ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	ds.ID = id
	if err := r.svc.UpdateDatasource(c.Request.Context(), &ds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

func (r *MonitorCRUDRouter) DeleteDatasource(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := r.svc.DeleteDatasource(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

func (r *MonitorCRUDRouter) TestDatasourceConnection(c *gin.Context) {
	var ds models.MonitorDatasource
	if err := c.ShouldBindJSON(&ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	ok, msg := r.svc.TestDatasourceConnection(c.Request.Context(), &ds)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"connected": ok, "message": msg}})
}

func (r *MonitorCRUDRouter) TestDatasourceByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ds, err := r.svc.GetDatasource(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "数据源不存在"})
		return
	}
	ok, msg := r.svc.TestDatasourceConnection(c.Request.Context(), ds)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"connected": ok, "message": msg}})
}

// ==================== 告警规则 ====================

func (r *MonitorCRUDRouter) ListAlertRules(c *gin.Context) {
	var req services.AlertRuleListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	result, err := r.svc.ListAlertRules(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (r *MonitorCRUDRouter) GetAlertRuleGroups(c *gin.Context) {
	groups, err := r.svc.GetAlertRuleGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": groups})
}

func (r *MonitorCRUDRouter) GetAlertRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rule, err := r.svc.GetAlertRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "规则不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rule})
}

func (r *MonitorCRUDRouter) CreateAlertRule(c *gin.Context) {
	var rule models.MonitorAlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	if err := r.svc.CreateAlertRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rule, "msg": "创建成功"})
}

func (r *MonitorCRUDRouter) UpdateAlertRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var rule models.MonitorAlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	rule.ID = id
	if err := r.svc.UpdateAlertRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

func (r *MonitorCRUDRouter) DeleteAlertRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := r.svc.DeleteAlertRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

func (r *MonitorCRUDRouter) ToggleAlertRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	if err := r.svc.ToggleAlertRule(c.Request.Context(), id, body.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

func (r *MonitorCRUDRouter) BatchDeleteAlertRules(c *gin.Context) {
	var req services.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	result, err := r.svc.BatchDeleteAlertRules(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result, "msg": fmt.Sprintf("批量删除完成: 成功%d 失败%d", result.Success, result.Failed)})
}

func (r *MonitorCRUDRouter) BatchUpdateAlertRules(c *gin.Context) {
	var req services.BatchUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	result, err := r.svc.BatchUpdateAlertRules(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result, "msg": fmt.Sprintf("批量更新完成: 成功%d 失败%d", result.Success, result.Failed)})
}

// ==================== 告警事件 ====================

func (r *MonitorCRUDRouter) ListAlertEvents(c *gin.Context) {
	var req services.AlertEventListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	result, err := r.svc.ListAlertEvents(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (r *MonitorCRUDRouter) GetAlertStats(c *gin.Context) {
	stats, err := r.svc.GetAlertStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": stats})
}

func (r *MonitorCRUDRouter) GetAlertEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	event, err := r.svc.GetAlertEvent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "事件不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": event})
}

func (r *MonitorCRUDRouter) AckAlertEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	// TODO: 从 JWT 获取 userID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "未登录"})
		return
	}
	userID := userIDVal.(int64)
	if err := r.svc.AckAlertEvent(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "确认成功"})
}

func (r *MonitorCRUDRouter) ResolveAlertEvent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := r.svc.ResolveAlertEvent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已解决"})
}

// ==================== YAML 批量导入/导出 ====================

func (r *MonitorCRUDRouter) ImportAlertRulesYAML(c *gin.Context) {
	var req services.AlertRuleImportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	result, err := r.svc.ImportAlertRulesFromYAML(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result, "msg": fmt.Sprintf("导入完成: 新建%d 更新%d 跳过%d 失败%d", result.Created, result.Updated, result.Skipped, result.Failed)})
}

func (r *MonitorCRUDRouter) ExportAlertRulesYAML(c *gin.Context) {
	group := c.Query("group")

	// 支持按 ID 导出：?ids=1,2,3
	var ids []int64
	if idsStr := c.Query("ids"); idsStr != "" {
		for _, idStr := range strings.Split(idsStr, ",") {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
				ids = append(ids, id)
			}
		}
	}

	yamlContent, err := r.svc.ExportAlertRulesToYAML(c.Request.Context(), group, ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	// 判断是下载还是预览
	if c.Query("download") == "true" {
		filename := "alert-rules.yaml"
		if group != "" {
			filename = fmt.Sprintf("alert-rules-%s.yaml", group)
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "application/x-yaml; charset=utf-8")
		c.String(http.StatusOK, yamlContent)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"yaml": yamlContent}})
}
