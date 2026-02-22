package job

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
	"k8soperation/pkg/k8s/job"
	"k8soperation/pkg/valid"
)

type KubeJobController struct {
}

func NewKubeJobController() *KubeJobController {
	return &KubeJobController{}
}

// Create godoc
// @Summary 创建 Job
// @Tags K8s Job 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeJobCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/job/create [post]
func (c *KubeJobController) Create(ctx *gin.Context) {
	req := requests.NewKubeJobCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeJobCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Job 创建逻辑（只有 Job，没有 Service）
	jobObj, err := svc.KubeJobCreate(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeJobCreate error", zap.Error(err))
		return
	}

	r.Success(job.BuildJobResponse(jobObj, req))
}

// CreateFromYaml godoc
// @Summary 从 YAML 创建 Job
// @Tags K8s Job 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/job/create-from-yaml [post]
func (c *KubeJobController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Service 层创建逻辑
	jobObj, err := svc.KubeJobCreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeJobCreateFromYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "Job 创建成功",
		"name":      jobObj.Name,
		"namespace": jobObj.Namespace,
	})
}

// List godoc
// @Summary 获取 K8s Job 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s Job 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "Job 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/list [get]
func (c *KubeJobController) List(ctx *gin.Context) {
	param := requests.NewKubeJobListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeJobListRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	jobs, total, err := svc.KubeJobList(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("获取Job列表失败", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorJobQueryFail.WithDetails(err.Error()))
		return
	}

	// 使用格式化函数转换 Job 列表
	list := job.BuildJobListResponse(jobs)
	resp.SuccessList(list, total)
}

// Detail godoc
// @Summary 获取 Job 详情
// @Tags K8s Job 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Job 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/detail [get]
func (c *KubeJobController) Detail(ctx *gin.Context) {
	// 构造请求参数对象
	param := requests.NewKubeJobDetailRequest()
	r := response.NewResponse(ctx)

	// 参数校验（valid.Validate 内部负责 Bind + 校验 + 错误响应）
	if ok := valid.Validate(ctx, param, requests.ValidKubeJobDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service 层逻辑
	svc := services.NewServices()
	jobObj, err := svc.KubeJobDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeJobDetail error", zap.Error(err))
		return
	}

	// 格式化输出（使用 BuildJobListResponse 中的同一逻辑）
	detailData := job.BuildJobDetailResponse(jobObj)

	// 成功返回
	r.Success(detailData)
}

// Delete godoc
// @Summary 删除 Job
// @Tags K8s Job 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Job 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/delete [delete]
func (c *KubeJobController) Delete(ctx *gin.Context) {
	// 参数解析
	param := requests.NewKubeJobDeleteRequest()
	r := response.NewResponse(ctx)

	// 参数校验（valid 内部自动 Bind + 校验 + 返回错误响应）
	if ok := valid.Validate(ctx, param, requests.ValidKubeJobDeleteRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service 层删除逻辑
	svc := services.NewServices()
	if err := svc.KubeJobDelete(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Error("service.KubeJobDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	// 删除成功响应
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "删除 Job 成功",
	})
}

// Suspend godoc
// @Summary 暂停或恢复 Job
// @Description 用于暂停或恢复指定命名空间下的 Job（通过设置 .spec.suspend 字段为 true/false）
// @Tags K8s Job 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeJobSuspendRequest true "Job 暂停/恢复请求体"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/suspend [patch]
func (c *KubeJobController) Suspend(ctx *gin.Context) {
	req := requests.NewKubeJobSuspendRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeJobSuspendRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Service 层逻辑
	if err := svc.KubeJobSuspend(ctx.Request.Context(), cli, req); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeJobSuspend error", zap.Error(err))
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
		"message":   fmt.Sprintf("Job 已成功%s", action),
	})
}

// Restart godoc
// @Summary 重启 Job
// @Description 基于已有 Job 模板重新创建一个新的 Job（清除状态、生成新名称）
// @Tags K8s Job 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeJobRestartRequest true "Job 重启请求体"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/restart [post]
func (c *KubeJobController) Restart(ctx *gin.Context) {
	req := requests.NewKubeJobRestartRequest()
	r := response.NewResponse(ctx)

	// 参数校验（绑定 + 验证）
	if ok := valid.Validate(ctx, req, requests.ValidKubeJobRestartRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Service 层逻辑
	jobObj, err := svc.KubeJobRestart(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeJobRestart error", zap.Error(err))
		return
	}

	// 成功响应
	r.Success(gin.H{
		"namespace": req.Namespace,
		"name":      req.Name,
		"newJob":    jobObj.Name,
		"message":   fmt.Sprintf("Job %s 已成功重启为 %s", req.Name, jobObj.Name),
	})
}

// UpdateImage godoc
// @Summary 更新 Job 镜像
// @Description 更新指定容器的镜像地址
// @Tags K8s Job 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeJobUpdateImageRequest true "更新镜像参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/update-image [put]
func (c *KubeJobController) UpdateImage(ctx *gin.Context) {
	req := requests.NewKubeJobUpdateImageRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeJobUpdateImageRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Service 层逻辑
	jobObj, err := svc.KubeJobUpdateImage(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeJobUpdateImage error", zap.Error(err))
		return
	}

	// 成功响应
	r.Success(gin.H{
		"namespace": jobObj.Namespace,
		"name":      jobObj.Name,
		"message":   "更新镜像成功",
	})
}

// GetYaml godoc
// @Summary 获取 Job YAML
// @Tags K8s Job 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Job 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/yaml [get]
func (c *KubeJobController) GetYaml(ctx *gin.Context) {
	param := requests.NewKubeJobDetailRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeJobDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	yamlStr, err := svc.KubeJobGetYaml(ctx.Request.Context(), cli, param.Namespace, param.Name)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeJobGetYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"yaml": yamlStr,
	})
}

// ApplyYaml godoc
// @Summary 应用 YAML 修改 Job
// @Tags K8s Job 管理
// @Accept json
// @Produce json
// @Param body body requests.YamlApplyRequest true "YAML 应用参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/job/apply-yaml [put]
func (c *KubeJobController) ApplyYaml(ctx *gin.Context) {
	req := requests.NewYamlApplyRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, req, requests.ValidYamlApplyRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	jobObj, err := svc.KubeJobApplyYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeJobApplyYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "YAML 应用成功",
		"name":      jobObj.Name,
		"namespace": jobObj.Namespace,
	})
}
