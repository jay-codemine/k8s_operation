package monitoring

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	dm "k8soperation/internal/domain/monitor"
	"k8soperation/middlewares"
)

// ==================== 通知渠道 ====================

func (r *MonitorCRUDRouter) ListNotifyChannels(c *gin.Context) {
	var req services.NotifyChannelListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().ListNotifyChannels(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (r *MonitorCRUDRouter) GetNotifyChannel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ch, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().GetNotifyChannel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "通知渠道不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": ch})
}

func (r *MonitorCRUDRouter) CreateNotifyChannel(c *gin.Context) {
	var ch models.MonitorNotifyChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().CreateNotifyChannel(c.Request.Context(), &ch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": ch, "msg": "创建成功"})
}

func (r *MonitorCRUDRouter) UpdateNotifyChannel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var ch models.MonitorNotifyChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	ch.ID = id
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().UpdateNotifyChannel(c.Request.Context(), &ch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

func (r *MonitorCRUDRouter) DeleteNotifyChannel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().DeleteNotifyChannel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

func (r *MonitorCRUDRouter) TestNotifyChannel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ch, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().GetNotifyChannel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "通知渠道不存在"})
		return
	}
	ok, msg := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().TestNotifyChannel(c.Request.Context(), ch)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"success": ok, "message": msg}})
}

// ==================== 静默规则 ====================

func (r *MonitorCRUDRouter) ListSilenceRules(c *gin.Context) {
	var req services.SilenceRuleListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().ListSilenceRules(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (r *MonitorCRUDRouter) GetSilenceRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rule, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().GetSilenceRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "规则不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rule})
}

func (r *MonitorCRUDRouter) CreateSilenceRule(c *gin.Context) {
	var rule models.MonitorSilenceRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().CreateSilenceRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rule, "msg": "创建成功"})
}

func (r *MonitorCRUDRouter) UpdateSilenceRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var rule models.MonitorSilenceRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	rule.ID = id
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().UpdateSilenceRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

func (r *MonitorCRUDRouter) DeleteSilenceRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().DeleteSilenceRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

// ==================== 抑制规则 ====================

func (r *MonitorCRUDRouter) ListInhibitRules(c *gin.Context) {
	var req services.InhibitRuleListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().ListInhibitRules(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (r *MonitorCRUDRouter) CreateInhibitRule(c *gin.Context) {
	var rule models.MonitorInhibitRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().CreateInhibitRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rule, "msg": "创建成功"})
}

func (r *MonitorCRUDRouter) UpdateInhibitRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var rule models.MonitorInhibitRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	rule.ID = id
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().UpdateInhibitRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

func (r *MonitorCRUDRouter) DeleteInhibitRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().DeleteInhibitRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

// ==================== 聚合规则 ====================

func (r *MonitorCRUDRouter) ListAggregateRules(c *gin.Context) {
	var req services.AggregateRuleListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().ListAggregateRules(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (r *MonitorCRUDRouter) CreateAggregateRule(c *gin.Context) {
	var rule models.MonitorAggregateRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().CreateAggregateRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rule, "msg": "创建成功"})
}

func (r *MonitorCRUDRouter) UpdateAggregateRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var rule models.MonitorAggregateRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	rule.ID = id
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().UpdateAggregateRule(c.Request.Context(), &rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

func (r *MonitorCRUDRouter) DeleteAggregateRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().DeleteAggregateRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

// ==================== 通知模板 ====================

func (r *MonitorCRUDRouter) ListNotifyTemplates(c *gin.Context) {
	var req services.NotifyTemplateListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().ListNotifyTemplates(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (r *MonitorCRUDRouter) GetNotifyTemplate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	tpl, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().GetNotifyTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "模板不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": tpl})
}

func (r *MonitorCRUDRouter) CreateNotifyTemplate(c *gin.Context) {
	var tpl models.MonitorNotifyTemplate
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().CreateNotifyTemplate(c.Request.Context(), &tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": tpl, "msg": "创建成功"})
}

func (r *MonitorCRUDRouter) UpdateNotifyTemplate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var tpl models.MonitorNotifyTemplate
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	tpl.ID = id
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().UpdateNotifyTemplate(c.Request.Context(), &tpl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

func (r *MonitorCRUDRouter) DeleteNotifyTemplate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().DeleteNotifyTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

func (r *MonitorCRUDRouter) PreviewNotifyTemplate(c *gin.Context) {
	var tpl models.MonitorNotifyTemplate
	if err := c.ShouldBindJSON(&tpl); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	preview := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().PreviewNotifyTemplate(c.Request.Context(), &tpl)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"rendered": preview}})
}

func (r *MonitorCRUDRouter) SetDefaultTemplate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().SetDefaultTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "设置成功"})
}

// ==================== 批量绑定渠道 ====================

func (r *MonitorCRUDRouter) BatchBindChannels(c *gin.Context) {
	var req dm.BatchBindChannelsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().BatchBindChannels(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result, "msg": fmt.Sprintf("批量操作完成: 成功%d 失败%d", result.Success, result.Failed)})
}

// ==================== 通知路由策略 ====================

func (r *MonitorCRUDRouter) ListRoutePolicies(c *gin.Context) {
	var req services.RoutePolicyListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().ListRoutePolicies(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (r *MonitorCRUDRouter) GetRoutePolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	policy, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().GetRoutePolicy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "路由策略不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": policy})
}

func (r *MonitorCRUDRouter) CreateRoutePolicy(c *gin.Context) {
	var policy models.MonitorNotifyRoutePolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().CreateRoutePolicy(c.Request.Context(), &policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": policy, "msg": "创建成功"})
}

func (r *MonitorCRUDRouter) UpdateRoutePolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var policy models.MonitorNotifyRoutePolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	policy.ID = id
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().UpdateRoutePolicy(c.Request.Context(), &policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

func (r *MonitorCRUDRouter) DeleteRoutePolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().DeleteRoutePolicy(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

// ==================== 批量删除: 告警事件 ====================

func (r *MonitorCRUDRouter) BatchDeleteAlertEvents(c *gin.Context) {
	var req dm.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().BatchDeleteAlertEvents(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result, "msg": fmt.Sprintf("成功删除 %d 条", result.Success)})
}

// ==================== 批量删除: 通知渠道 ====================

func (r *MonitorCRUDRouter) BatchDeleteNotifyChannels(c *gin.Context) {
	var req dm.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().BatchDeleteNotifyChannels(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result, "msg": fmt.Sprintf("成功删除 %d 个渠道", result.Success)})
}

// ==================== 批量更新: 通知渠道 ====================

func (r *MonitorCRUDRouter) BatchUpdateNotifyChannels(c *gin.Context) {
	var req dm.BatchUpdateNotifyChannelsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().BatchUpdateNotifyChannels(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result, "msg": fmt.Sprintf("批量更新完成: 成功%d 失败%d", result.Success, result.Failed)})
}

// ==================== 批量删除: 静默规则 ====================

func (r *MonitorCRUDRouter) BatchDeleteSilenceRules(c *gin.Context) {
	var req dm.BatchDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误: " + err.Error()})
		return
	}
	result, err := middlewares.NewServicesFromContext(c).MonitorCRUDSvc().BatchDeleteSilenceRules(c.Request.Context(), req.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result, "msg": fmt.Sprintf("成功删除 %d 条规则", result.Success)})
}
