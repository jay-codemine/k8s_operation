package namespace

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	nspkg "k8soperation/pkg/k8s/namespace"
	"k8soperation/pkg/valid"
)

type KubeNamespaceController struct {
}

func NewKubeNamespaceController() *KubeNamespaceController {
	return &KubeNamespaceController{}
}

// @Summary     创建 Namespace
// @Description 创建命名空间，并可选设置 labels/annotations 和资源配额（CPU/Memory/Pods）
// @Tags        K8s Namespace 管理
// @Accept      json
// @Produce     json
// @Param       body body requests.KubeNamespaceCreateRequest true "Namespace 创建参数"
// @Success     200 {object} response.Response
// @Failure     400 {object} errorcode.Error
// @Failure     500 {object} errorcode.Error
// @Router      /api/v1/k8s/namespace/create [post]
func (ctl *KubeNamespaceController) Create(ctx *gin.Context) {
	// 标准响应对象
	r := response.NewResponse(ctx)

	// DTO
	req := requests.NewKubeNamespaceCreateRequest()

	// 参数绑定+校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeNamespaceCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// Service 层
	svc := services.NewServices()

	// 调用 Service 创建 Namespace
	ns, err := svc.KubeCreateNamespace(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeCreateNamespace error", zap.Error(err))
		return
	}

	// 返回创建结果
	r.Success(gin.H{
		"name":        ns.Name,
		"labels":      ns.Labels,
		"annotations": ns.Annotations,
		"status":      ns.Status.Phase,
		"createdAt":   ns.CreationTimestamp,
	})
}

// @Summary 获取 Namespace 列表
// @Description 支持分页、名称模糊查询，返回每个命名空间的资源统计
// @Tags K8s Namespace 管理
// @Produce json
// @Param name  query string false "命名空间名称(模糊匹配)" maxlength(100)
// @Param page  query int    true  "页码(从1开始)"
// @Param limit query int    true  "每页数量(默认20)"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error   "请求参数错误"
// @Failure 500 {object} errorcode.Error   "内部错误"
// @Router /api/v1/k8s/namespace/list [get]
func (ctl *KubeNamespaceController) List(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeNamespaceListRequest()

	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceListRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	items, total, err := svc.KubeNamespaceList(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNamespaceList error", zap.Error(err))
		return
	}

	// 为每个 namespace 添加资源统计
	type NamespaceWithStats struct {
		Name              string            `json:"name"`
		Status            string            `json:"status"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		CreationTimestamp interface{}       `json:"creation_timestamp"`
		PodNum            int               `json:"pod_num"`
		ServiceNum        int               `json:"service_num"`
		DeploymentNum     int               `json:"deployment_num"`
	}

	result := make([]NamespaceWithStats, 0, len(items))
	for _, ns := range items {
		var podNum, serviceNum, deploymentNum int

		// 统计 Pods 数量
		if podList, err := cli.Kube.CoreV1().Pods(ns.Name).List(ctx.Request.Context(), metav1.ListOptions{}); err == nil {
			podNum = len(podList.Items)
		}

		// 统计 Services 数量
		if svcList, err := cli.Kube.CoreV1().Services(ns.Name).List(ctx.Request.Context(), metav1.ListOptions{}); err == nil {
			serviceNum = len(svcList.Items)
		}

		// 统计 Deployments 数量
		if deployList, err := cli.Kube.AppsV1().Deployments(ns.Name).List(ctx.Request.Context(), metav1.ListOptions{}); err == nil {
			deploymentNum = len(deployList.Items)
		}

		result = append(result, NamespaceWithStats{
			Name:              ns.Name,
			Status:            string(ns.Status.Phase),
			Labels:            ns.Labels,
			Annotations:       ns.Annotations,
			CreationTimestamp: ns.CreationTimestamp,
			PodNum:            podNum,
			ServiceNum:        serviceNum,
			DeploymentNum:     deploymentNum,
		})
	}

	r.SuccessList(result, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 Namespace 列表成功，共 %d 条", total),
	})
}

// Detail godoc
// @Summary 获取 Namespace 详情
// @Description 查询指定 Namespace 的详细信息，包含资源统计
// @Tags K8s Namespace 管理
// @Produce json
// @Param name query string true "Namespace 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "资源不存在"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/namespace/detail [get]
func (c *KubeNamespaceController) Detail(ctx *gin.Context) {
	param := requests.NewKubeNamespaceDetailRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	nsDetail, err := svc.KubeNamespaceDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNamespaceDetail error", zap.Error(err))
		return
	}

	// 统计命名空间下的资源数量
	var podNum, serviceNum, deploymentNum int

	// 统计 Pods 数量
	podList, err := cli.Kube.CoreV1().Pods(param.Name).List(ctx.Request.Context(), metav1.ListOptions{})
	if err == nil {
		podNum = len(podList.Items)
	}

	// 统计 Services 数量
	svcList, err := cli.Kube.CoreV1().Services(param.Name).List(ctx.Request.Context(), metav1.ListOptions{})
	if err == nil {
		serviceNum = len(svcList.Items)
	}

	// 统计 Deployments 数量
	deployList, err := cli.Kube.AppsV1().Deployments(param.Name).List(ctx.Request.Context(), metav1.ListOptions{})
	if err == nil {
		deploymentNum = len(deployList.Items)
	}

	r.Success(gin.H{
		"message":        fmt.Sprintf("获取 Namespace %s 详情成功", param.Name),
		"name":           nsDetail.Name,
		"status":         nsDetail.Status.Phase,
		"labels":         nsDetail.Labels,
		"annotations":    nsDetail.Annotations,
		"created_at":     nsDetail.CreationTimestamp,
		"pod_num":        podNum,
		"service_num":    serviceNum,
		"deployment_num": deploymentNum,
	})
}

// @Summary 删除 Namespace
// @Description 删除指定 Namespace（级联删除内部资源）
// @Tags K8s Namespace 管理
// @Produce json
// @Param name query string true "Namespace 名称"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "删除失败"
// @Router /api/v1/k8s/namespace/delete [delete]
func (ctl *KubeNamespaceController) Delete(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubeNamespaceDeleteRequest()

	// 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceDeleteRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用 Service
	svc := services.NewServices()
	if err := svc.KubeNamespaceDelete(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNamespaceDelete error", zap.Error(err))
		return
	}

	// 返回成功
	r.Success(gin.H{
		"message": fmt.Sprintf("Namespace %s 删除成功", param.Name),
	})
}

// Patch godoc
// @Summary 修改 Namespace（labels / annotations）
// @Description 支持新增、更新、删除 labels 与 annotations
// @Tags K8s Namespace 管理
// @Produce json
// @Param name query string true "Namespace 名称"
// @Param patch body requests.KubeNamespaceUpdateRequest true "Patch 内容"
// @Success 200 {object} response.Response "修改成功"
// @Failure 400 {object} errorcode.Error "参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/namespace/patch [patch]
func (c *KubeNamespaceController) Patch(ctx *gin.Context) {
	param := requests.NewKubeNamespaceUpdateRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceUpdateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	updated, err := svc.KubeNamespaceUpdate(ctx, cli, param)
	if err != nil {
		ctx.Error(err)
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("Namespace %s 修改成功", param.Name),
		"data":    updated,
	})
}

// PatchLabels godoc
// @Summary 修改 Namespace 标签
// @Description 添加或删除 Namespace 标签
// @Tags K8s Namespace 管理
// @Accept json
// @Produce json
// @Param data body requests.KubeNamespaceLabelPatchRequest true "标签修改参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/namespace/labels [patch]
func (c *KubeNamespaceController) PatchLabels(ctx *gin.Context) {
	param := requests.NewKubeNamespaceLabelPatchRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceLabelPatchRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()

	if err := svc.KubeNamespacePatchLabels(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeNamespacePatchLabels error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message": fmt.Sprintf("Namespace %s 标签修改成功", param.Name),
		"name":    param.Name,
	})
}

// Yaml godoc
// @Summary 获取 Namespace YAML 配置
// @Description 获取指定 Namespace 的 YAML 配置
// @Tags K8s Namespace 管理
// @Produce json
// @Param name query string true "Namespace 名称"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/namespace/yaml [get]
func (c *KubeNamespaceController) Yaml(ctx *gin.Context) {
	param := requests.NewKubeNamespaceDetailRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeNamespaceDetailRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	yamlStr, err := nspkg.GetYaml(ctx.Request.Context(), cli.Kube, param.Name)
	if err != nil {
		global.Logger.Error("获取 Namespace YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"name": param.Name,
		"yaml": yamlStr,
	})
}

// ApplyYaml godoc
// @Summary 应用 Namespace YAML 配置
// @Description 应用修改后的 YAML 配置到 Namespace
// @Tags K8s Namespace 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeApplyYamlClusterRequest true "YAML内容"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/namespace/apply_yaml [put]
func (c *KubeNamespaceController) ApplyYaml(ctx *gin.Context) {
	param := requests.NewKubeApplyYamlClusterRequest()
	r := response.NewResponse(ctx)
	if ok := valid.Validate(ctx, param, requests.ValidKubeApplyYamlClusterRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	_, err := nspkg.ApplyYaml(ctx.Request.Context(), cli.Kube, param.Name, param.Yaml)
	if err != nil {
		global.Logger.Error("应用 Namespace YAML 失败", zap.Error(err))
		r.ToErrorResponse(errorcode.ServerError.WithDetails(err.Error()))
		return
	}

	r.Success(gin.H{
		"name":    param.Name,
		"message": "YAML 应用成功",
	})
}
