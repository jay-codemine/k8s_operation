package ai

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/models"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

// AIApprovalController 高危操作审批控制器
type AIApprovalController struct {
	factory *services.ClusterClientFactory
}

func NewAIApprovalController() *AIApprovalController {
	svc := services.NewServices()
	return &AIApprovalController{
		factory: services.NewClusterClientFactory(svc),
	}
}

// isApprovalAdmin 检查当前用户是否有审批管理权限
// 只有 super_admin / platform_admin / cluster_admin 才能审批
func isApprovalAdmin(ctx *gin.Context) bool {
	userID := getUserID(ctx)
	if userID == 0 {
		return false
	}
	svc := services.NewServices()
	return svc.IsApprovalAdmin(int64(userID))
}

// getUserName 从上下文获取当前用户名
func getUserName(ctx *gin.Context) string {
	if u, ok := ctx.Get("current_user"); ok {
		if user, ok := u.(*models.User); ok {
			return user.Username
		}
		if user, ok := u.(models.User); ok {
			return user.Username
		}
	}
	if name, ok := ctx.Get("current_user_name"); ok {
		if s, ok := name.(string); ok {
			return s
		}
	}
	return ""
}

// List 获取审批列表
// @Summary 获取审批列表
// @Description 管理员获取所有审批，普通用户仅看自己的申请
// @Tags AI 审批管理
// @Produce json
// @Param status query int false "状态筛选: 1=待审批 2=已通过 3=已拒绝 4=已过期 5=已取消" default(0)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals [get]
func (c *AIApprovalController) List(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	status, _ := strconv.ParseUint(ctx.DefaultQuery("status", "0"), 10, 8)
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	viewAll := ctx.DefaultQuery("view", "") == "all" // 独立管理页面传 view=all 查看全部

	svc := services.NewServices()
	admin := isApprovalAdmin(ctx)

	var list []*models.AIApprovalRequest
	var total int64
	var err error

	if admin || viewAll {
		// 管理员或独立管理页面可以看到所有审批
		list, total, err = svc.AIApprovalList(ctx.Request.Context(), uint8(status), page, pageSize)
	} else {
		// AI 助手侧边栏：普通用户只能看到自己提交的审批
		list, total, err = svc.AIApprovalMyList(ctx.Request.Context(), userID, page, pageSize)
	}
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	// 扩展返回信息：添加申请人用户名、审批人用户名
	type ApprovalWithUser struct {
		*models.AIApprovalRequest
		RequestUserName  string `json:"request_user_name"`
		ApproverUserName string `json:"approver_user_name,omitempty"`
		CanApprove       bool   `json:"can_approve"`
	}

	var enrichedList []ApprovalWithUser
	for _, item := range list {
		rich := ApprovalWithUser{
			AIApprovalRequest: item,
			CanApprove:         admin && item.RequestUserID != userID && item.Status == models.AIApprovalPending,
		}
		// 查询申请人名称
		if u := models.NewUser().GetUserByID(strconv.FormatUint(uint64(item.RequestUserID), 10)); u.ID > 0 {
			rich.RequestUserName = u.Username
		}
		// 查询审批人名称
		if item.ApproverUserID > 0 {
			if u := models.NewUser().GetUserByID(strconv.FormatUint(uint64(item.ApproverUserID), 10)); u.ID > 0 {
				rich.ApproverUserName = u.Username
			}
		}
		enrichedList = append(enrichedList, rich)
	}

	resp.Success(gin.H{
		"list":     enrichedList,
		"total":    total,
		"is_admin": admin,
	})
}

// MyList 获取我的审批申请
// @Summary 获取我的审批申请
// @Tags AI 审批管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals/mine [get]
func (c *AIApprovalController) MyList(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	svc := services.NewServices()
	list, total, err := svc.AIApprovalMyList(ctx.Request.Context(), userID, page, pageSize)
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.SuccessList(list, total)
}

// Detail 获取审批详情
// @Summary 获取审批详情
// @Tags AI 审批管理
// @Param id path int true "审批ID"
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals/{id} [get]
func (c *AIApprovalController) Detail(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	approvalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if approvalID == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的审批ID"))
		return
	}

	svc := services.NewServices()
	req, logs, err := svc.AIApprovalDetail(ctx.Request.Context(), uint32(approvalID))
	if err != nil {
		resp.ToErrorResponse(errorcode.AIApprovalNotFound)
		return
	}

	resp.Success(gin.H{
		"approval": req,
		"logs":     logs,
	})
}

// Approve 通过审批
// @Summary 通过审批
// @Description 管理员通过高危操作审批（不可自审）
// @Tags AI 审批管理
// @Accept json
// @Produce json
// @Param id path int true "审批ID"
// @Param body body object true "审批备注" example({"comment":"确认执行"})
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals/{id}/approve [post]
func (c *AIApprovalController) Approve(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	// 权限校验：只有管理员角色才能审批
	if !isApprovalAdmin(ctx) {
		resp.ToErrorResponse(errorcode.AIApprovalForbidden.WithDetails("仅管理员（super_admin/platform_admin/cluster_admin）可审批"))
		return
	}

	approvalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if approvalID == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的审批ID"))
		return
	}

	var body struct {
		Comment       string `json:"comment"`
		AdminOverride bool   `json:"admin_override"` // 管理页面审批时可跳过自审限制
	}
	_ = ctx.ShouldBindJSON(&body)

	svc := services.NewServices()
	if err := svc.AIApprovalApprove(ctx.Request.Context(), uint32(approvalID), userID, body.Comment, c.factory, body.AdminOverride); err != nil {
		global.Logger.Error("审批通过失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AIApprovalProcessed.WithDetails(err.Error()))
		return
	}

	global.Logger.Info("审批已通过",
		zap.Uint64("approval_id", approvalID),
		zap.Uint32("approver_id", userID),
		zap.String("approver_name", getUserName(ctx)),
	)
	resp.Success(gin.H{"message": "审批已通过", "approver": getUserName(ctx)})
}

// Reject 拒绝审批
// @Summary 拒绝审批
// @Description 管理员拒绝高危操作审批（不可自审）
// @Tags AI 审批管理
// @Accept json
// @Produce json
// @Param id path int true "审批ID"
// @Param body body object true "拒绝原因" example({"comment":"操作风险过高"})
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals/{id}/reject [post]
func (c *AIApprovalController) Reject(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	// 权限校验：只有管理员角色才能审批
	if !isApprovalAdmin(ctx) {
		resp.ToErrorResponse(errorcode.AIApprovalForbidden.WithDetails("仅管理员（super_admin/platform_admin/cluster_admin）可审批"))
		return
	}

	approvalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if approvalID == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的审批ID"))
		return
	}

	var body struct {
		Comment       string `json:"comment"`
		AdminOverride bool   `json:"admin_override"` // 管理页面审批时可跳过自审限制
	}
	_ = ctx.ShouldBindJSON(&body)

	svc := services.NewServices()
	if err := svc.AIApprovalReject(ctx.Request.Context(), uint32(approvalID), userID, body.Comment, body.AdminOverride); err != nil {
		global.Logger.Error("审批拒绝失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AIApprovalProcessed.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "审批已拒绝", "approver": getUserName(ctx)})
}

// Cancel 取消审批（申请人自己取消）
// @Summary 取消审批
// @Tags AI 审批管理
// @Param id path int true "审批ID"
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals/{id}/cancel [post]
func (c *AIApprovalController) Cancel(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	approvalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if approvalID == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的审批ID"))
		return
	}

	svc := services.NewServices()
	if err := svc.AIApprovalCancel(ctx.Request.Context(), uint32(approvalID), userID); err != nil {
		global.Logger.Error("审批取消失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AIApprovalForbidden.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "审批已取消"})
}

// PendingCount 获取待审批数量
// @Summary 获取待审批数量
// @Description 管理员看全局待审批数，普通用户看自己的
// @Tags AI 审批管理
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals/pending-count [get]
func (c *AIApprovalController) PendingCount(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	admin := isApprovalAdmin(ctx)

	svc := services.NewServices()
	var count int64
	var err error

	if admin {
		// 管理员看全局待审批数量
		count, err = svc.AIApprovalPendingCount(ctx.Request.Context())
	} else {
		// 普通用户看自己提交的待审批数量
		count, err = svc.AIApprovalMyPendingCount(ctx.Request.Context(), userID)
	}
	if err != nil {
		global.Logger.Error("获取待审批数量失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"count": count, "is_admin": admin})
}

// Delete 删除审批记录
// @Summary 删除审批记录
// @Description 管理员可删除任何记录，普通用户只能删除自己的非已通过记录
// @Tags AI 审批管理
// @Param id path int true "审批ID"
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals/{id} [delete]
func (c *AIApprovalController) Delete(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	approvalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if approvalID == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的审批ID"))
		return
	}

	admin := isApprovalAdmin(ctx)
	svc := services.NewServices()
	if err := svc.AIApprovalDelete(ctx.Request.Context(), uint32(approvalID), userID, admin); err != nil {
		global.Logger.Error("删除审批失败", zap.Error(err))
		resp.ToErrorResponse(errorcode.AIApprovalForbidden.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "审批记录已删除"})
}

// Update 更新审批备注
// @Summary 更新审批备注
// @Description 仅待审批状态的记录可编辑备注
// @Tags AI 审批管理
// @Accept json
// @Produce json
// @Param id path int true "审批ID"
// @Param body body object true "更新内容" example({"comment":"备注内容"})
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals/{id} [put]
func (c *AIApprovalController) Update(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	userID := getUserID(ctx)
	if userID == 0 {
		resp.ToErrorResponse(errorcode.UserNotLogin)
		return
	}

	approvalID, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if approvalID == 0 {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的审批ID"))
		return
	}

	var body struct {
		Comment string `json:"comment"`
	}
	_ = ctx.ShouldBindJSON(&body)

	admin := isApprovalAdmin(ctx)
	svc := services.NewServices()
	if err := svc.AIApprovalUpdate(ctx.Request.Context(), uint32(approvalID), userID, body.Comment, admin); err != nil {
		resp.ToErrorResponse(errorcode.AIApprovalForbidden.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"message": "审批备注已更新"})
}

// Stats 获取审批统计数据
// @Summary 获取审批统计数据
// @Tags AI 审批管理
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Router /api/v1/ai/approvals/stats [get]
func (c *AIApprovalController) Stats(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	svc := services.NewServices()
	stats, err := svc.AIApprovalStats(ctx.Request.Context())
	if err != nil {
		resp.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{"stats": stats})
}
