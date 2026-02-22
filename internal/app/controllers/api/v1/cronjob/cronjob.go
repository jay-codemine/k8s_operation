package cronjob

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/k8s/cronjob"
	"k8soperation/pkg/valid"
)

type KubeCronJobController struct{}

func NewKubeCronJobController() *KubeCronJobController {
	return &KubeCronJobController{}
}

// Create godoc
// @Summary 创建 CronJob
// @Tags K8s CronJob 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeCronJobCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/cronjob/create [post]
func (c *KubeCronJobController) Create(ctx *gin.Context) {
	req := requests.NewKubeCronJobCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeCronJobCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 CronJob 创建逻辑
	cronJobObj, err := svc.KubeCronJobCreate(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobCreate error", zap.Error(err))
		return
	}

	// 返回结果（BuildCronJobResponse 是你类似 job.BuildJobResponse 的构造函数）
	r.Success(cronjob.BuildCronJobResponse(cronJobObj, req))
}

// List godoc
// @Summary 获取 K8s CronJob 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s CronJob 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "CronJob 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cronjob/list [get]
func (c *KubeCronJobController) List(ctx *gin.Context) {
	param := requests.NewKubeCronJobListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeCronJobListRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	result, err := svc.KubeCronJobList(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("获取CronJob列表失败", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorCronJobQueryFail.WithDetails(err.Error()))
		return
	}

	// 使用格式化函数转换 CronJob 列表（包含 Job 统计）
	list := cronjob.BuildCronJobListResponse(result.CronJobs, result.Jobs)
	resp.SuccessList(list, result.Total)
}

// Detail godoc
// @Summary 获取 CronJob 详情（含历史 Job 列表）
// @Tags K8s CronJob 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "CronJob 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cronjob/detail [get]
func (c *KubeCronJobController) Detail(ctx *gin.Context) {
	param := requests.NewKubeCronJobDetailRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeCronJobDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	cj, jobs, err := svc.KubeCronJobDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobDetail error", zap.Error(err))
		return
	}

	// 构造统一返回格式
	jobSummaries := make([]gin.H, 0, len(jobs))
	for _, j := range jobs {
		phase := "Pending"
		switch {
		case j.Status.Succeeded > 0:
			phase = "Complete"
		case j.Status.Failed > 0:
			phase = "Failed"
		case j.Status.Active > 0:
			phase = "Running"
		}
		jobSummaries = append(jobSummaries, gin.H{
			"name":           j.Name,
			"namespace":      j.Namespace,
			"startTime":      j.Status.StartTime,
			"completionTime": j.Status.CompletionTime,
			"active":         j.Status.Active,
			"succeeded":      j.Status.Succeeded,
			"failed":         j.Status.Failed,
			"phase":          phase,
		})
	}

	r.Success(gin.H{
		"message": "获取 CronJob 详情成功",
		"cronjob": gin.H{
			"name":              cj.Name,
			"namespace":         cj.Namespace,
			"schedule":          cj.Spec.Schedule,
			"suspend":           cj.Spec.Suspend != nil && *cj.Spec.Suspend,
			"concurrencyPolicy": string(cj.Spec.ConcurrencyPolicy),
			"lastScheduleTime":  cj.Status.LastScheduleTime,
			"lastSuccessfulTime": func() interface{} {
				if cj.Status.LastSuccessfulTime == nil {
					return nil
				}
				return cj.Status.LastSuccessfulTime.Time
			}(),
		},
		"jobs": jobSummaries,
	})
}

// Delete godoc
// @Summary 删除 CronJob
// @Tags K8s CronJob 管理
// @Produce json
// @Param body body requests.KubeCronJobDeleteRequest true "删除参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/cronjob/delete [delete]
func (c *KubeCronJobController) Delete(ctx *gin.Context) {
	req := requests.NewKubeCronJobDeleteRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, req, requests.ValidKubeCronJobDeleteRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	if err := svc.KubeCronJobDelete(ctx.Request.Context(), cli, req); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobDelete error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("CronJob %s/%s 删除成功", req.Namespace, req.Name),
	})
}

// Suspend godoc
// @Summary 暂停或恢复 CronJob
// @Description 用于暂停或恢复指定命名空间下的 CronJob（通过设置 .spec.suspend 字段为 true/false）
// @Tags K8s CronJob 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeCronJobSuspendRequest true "CronJob 暂停/恢复请求体"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cronjob/suspend [patch]
func (c *KubeCronJobController) Suspend(ctx *gin.Context) {
	req := requests.NewKubeCronJobSuspendRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeCronJobSuspendRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Service 层逻辑
	if err := svc.KubeCronJobSuspend(ctx.Request.Context(), cli, req); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobSuspend error", zap.Error(err))
		return
	}

	// 成功响应
	action := "暂停"
	if !req.Suspend {
		action = "恢复"
	}

	r.Success(gin.H{
		"namespace": req.Namespace,
		"name":      req.Name,
		"message":   fmt.Sprintf("CronJob 已成功%s", action),
	})
}

// CreateFromYaml godoc
// @Summary 从 YAML 创建 CronJob
// @Tags K8s CronJob 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/cronjob/create-from-yaml [post]
func (c *KubeCronJobController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Service 层创建逻辑
	cronJobObj, err := svc.KubeCronJobCreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobCreateFromYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "CronJob 创建成功",
		"name":      cronJobObj.Name,
		"namespace": cronJobObj.Namespace,
	})
}

// UpdateFromYaml godoc
// @Summary 从 YAML 更新 CronJob
// @Tags K8s CronJob 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 更新参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/cronjob/update-from-yaml [put]
func (c *KubeCronJobController) UpdateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Service 层更新逻辑
	cronJobObj, err := svc.KubeCronJobUpdateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobUpdateFromYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "CronJob 更新成功",
		"name":      cronJobObj.Name,
		"namespace": cronJobObj.Namespace,
	})
}

// Trigger godoc
// @Summary 手动触发 CronJob
// @Description 立即创建一个 Job（不等待调度时间）
// @Tags K8s CronJob 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeCronJobTriggerRequest true "触发参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cronjob/trigger [post]
func (c *KubeCronJobController) Trigger(ctx *gin.Context) {
	req := requests.NewKubeCronJobTriggerRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeCronJobTriggerRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Service 层触发逻辑
	job, err := svc.KubeCronJobTrigger(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCronJobTrigger error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "CronJob 已手动触发",
		"job_name":  job.Name,
		"namespace": job.Namespace,
	})
}
