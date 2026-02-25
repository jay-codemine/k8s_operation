package k8s_cluster

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type K8sClusterController struct{}

func NewK8sClusterController() *K8sClusterController { return &K8sClusterController{} }

// @Summary 创建K8s集群
// @Description 创建K8s集群
// @Tags K8s集群管理
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.K8sClusterCreateRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cluster/create [post]
func (c *K8sClusterController) Create(ctx *gin.Context) {
	param := requests.NewK8sClusterCreateRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidK8sClusterCreateRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.K8sClusterCreate(ctx.Request.Context(), param); err != nil {
		global.Logger.Error("创建K8s集群失败", zap.Error(err))
		// 如果项目里没有 ErrorClusterCreateFail，可替换为 ServerError 或自定义
		rsp.ToErrorResponse(errorcode.ErrorClusterCreateFail)
		return
	}
	rsp.Success(gin.H{"data": "创建K8s集群成功"})
}

// @Summary 列出K8s集群
// @Description 列出K8s集群
// @Tags K8s集群管理
// @Produce json
// @Security ApiKeyAuth
// @Param cluster_name query string false "K8s集群名" maxlength(100)
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cluster/list [get]
func (c *K8sClusterController) List(ctx *gin.Context) {
	param := requests.NewK8sClusterListRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidK8sClusterListRequest); !ok {
		return
	}

	svc := services.NewServices()
	list, total, err := svc.K8sClusterList(ctx.Request.Context(), param)
	if err != nil {
		global.Logger.Error("获取K8s集群列表失败", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorClusterQueryFail)
		return
	}
	rsp.SuccessList(list, total)
}

// @Summary 修改K8s集群
// @Description 修改K8s集群
// @Tags K8s集群管理
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.K8sClusterUpdateRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cluster/update [post]
func (c *K8sClusterController) Update(ctx *gin.Context) {
	param := requests.NewK8sClusterUpdateRequest()
	rsp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidK8sClusterUpdateRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.K8sClusterUpdate(ctx.Request.Context(), param); err != nil {
		global.Logger.Error("修改K8s集群失败", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorClusterUpdateFail)
		return
	}
	rsp.Success(gin.H{"data": "修改K8s集群成功"})
}

// @Summary 删除K8s集群
// @Description 删除K8s集群（软删除）
// @Tags K8s集群管理
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.K8sClusterDeleteRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cluster/delete [post]
func (c *K8sClusterController) Delete(ctx *gin.Context) {
	param := requests.NewK8sClusterDeleteRequest()
	rsp := response.NewResponse(ctx)

	// 打印原始 body（确认客户端到底传了啥）
	body, _ := io.ReadAll(ctx.Request.Body)
	global.Logger.Info("请求原始 Body", zap.String("body", string(body)))

	// 重新把 body 放回去（因为 ReadAll 会消耗掉 body）
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// 打印请求头 Content-Type
	global.Logger.Info("Content-Type", zap.String("ct", ctx.GetHeader("Content-Type")))

	if ok := valid.Validate(ctx, param, requests.ValidK8sClusterDeleteRequest); !ok {
		return
	}

	svc := services.NewServices()
	if err := svc.K8sClusterDelete(ctx.Request.Context(), param); err != nil {
		global.Logger.Error("删除K8s集群失败", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorClusterDeleteFail)
		return
	}
	rsp.Success(gin.H{"data": "删除K8s集群成功"})
}

// @Summary 初始化K8s集群
// @Description 初始化K8s集群（连通性检测/预热）
// @Tags K8s集群管理
// @Produce json
// @Security ApiKeyAuth
// @Param body body requests.K8sClusterInitRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/cluster/init [post]
func (c *K8sClusterController) Init(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	param := requests.NewK8sClusterInitRequest()

	if ok := valid.Validate(ctx, param, requests.ValidK8sClusterInitRequest); !ok {
		return
	}

	svc := services.NewServices()

	// 现在的返回值是 *services.K8sClients（里面封装了 RestConfig/Clientset/MetricsClient）
	cli, err := svc.K8sClusterInit(ctx.Request.Context(), param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("初始化K8s集群失败", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ErrorClusterInitFailed)
		return
	}

	// 将CLI配置和客户端对象赋值给全局变量，以便在程序其他部分使用
	// global.KubeConfig 用于存储Kubernetes的配置信息
	global.KubeConfig = cli.Config
	// global.KubeClient 用于存储Kubernetes的客户端实例，用于与Kubernetes API交互
	global.KubeClient = cli.Kube
	// global.MetricsClient 用于存储Metrics客户端实例，用于获取Kubernetes集群的指标数据
	global.MetricsClient = cli.Metrics

	rsp.Success(gin.H{
		"message":  "初始化K8s集群成功",
		"eventsV1": global.SupportsEventsV1,
	})
}
