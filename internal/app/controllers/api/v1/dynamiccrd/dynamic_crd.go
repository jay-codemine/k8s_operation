package dynamiccrd

import (
	"strings"

	"github.com/gin-gonic/gin"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
)

// DynamicCRDController CRD/CR 动态资源管理控制器
type DynamicCRDController struct{}

func NewDynamicCRDController() *DynamicCRDController {
	return &DynamicCRDController{}
}

// ==================== CRD 管理 ====================

// ListCRDs 列出所有 CRD
func (c *DynamicCRDController) ListCRDs(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	keyword := ctx.Query("keyword")
	group := ctx.Query("group")

	svc := services.NewServices()
	items, total, err := svc.KubeCRDList(ctx.Request.Context(), cli, keyword, group)
	if err != nil {
		r.ToErrorResponse(errorcode.ErrorCRDListFail.WithDetails(err.Error()))
		return
	}
	r.SuccessList(items, total)
}

// GetCRD 获取 CRD 详情
func (c *DynamicCRDController) GetCRD(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	name := ctx.Query("name")
	if name == "" {
		r.ToErrorResponse(errorcode.InvalidParams.WithDetails("name is required"))
		return
	}

	svc := services.NewServices()
	obj, err := svc.KubeCRDGet(ctx.Request.Context(), cli, name)
	if err != nil {
		r.ToErrorResponse(errorcode.ErrorCRDGetFail.WithDetails(err.Error()))
		return
	}
	r.Success(obj)
}

// DeleteCRD 删除 CRD
func (c *DynamicCRDController) DeleteCRD(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	name := ctx.Query("name")
	if name == "" {
		r.ToErrorResponse(errorcode.InvalidParams.WithDetails("name is required"))
		return
	}

	svc := services.NewServices()
	if err := svc.KubeCRDDelete(ctx.Request.Context(), cli, name); err != nil {
		if strings.Contains(err.Error(), "删除保护") {
			r.ToErrorResponse(errorcode.ErrorCRDDeleteProtected.WithDetails(err.Error()))
			return
		}
		r.ToErrorResponse(errorcode.ErrorCRDDeleteFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{"message": "CRD 删除成功", "name": name})
}

// ==================== CR 实例管理 ====================

// ListCRs 列出 CR 实例
func (c *DynamicCRDController) ListCRs(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	group := ctx.Query("group")
	version := ctx.Query("version")
	resource := ctx.Query("resource")
	namespace := ctx.Query("namespace")
	labelSelector := ctx.Query("label_selector")

	if version == "" || resource == "" {
		r.ToErrorResponse(errorcode.ErrorCRGVRInvalid.WithDetails("version and resource are required"))
		return
	}

	svc := services.NewServices()
	list, err := svc.KubeCRList(ctx.Request.Context(), cli, group, version, resource, namespace, labelSelector)
	if err != nil {
		r.ToErrorResponse(errorcode.ErrorCRListFail.WithDetails(err.Error()))
		return
	}

	// 转换为前端友好的扁平结构
	type CRItem struct {
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		CreatedAt         string            `json:"created_at"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string `json:"annotations"`
		UID               string            `json:"uid"`
		ResourceVersion   string            `json:"resource_version"`
	}

	items := make([]CRItem, 0, len(list.Items))
	for _, item := range list.Items {
		labels := item.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		annotations := item.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		items = append(items, CRItem{
			Name:            item.GetName(),
			Namespace:       item.GetNamespace(),
			CreatedAt:       item.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
			Labels:          labels,
			Annotations:     annotations,
			UID:             string(item.GetUID()),
			ResourceVersion: item.GetResourceVersion(),
		})
	}
	r.SuccessList(items, int64(len(items)))
}

// GetCR 获取单个 CR
func (c *DynamicCRDController) GetCR(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	group := ctx.Query("group")
	version := ctx.Query("version")
	resource := ctx.Query("resource")
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	if version == "" || resource == "" || name == "" {
		r.ToErrorResponse(errorcode.ErrorCRGVRInvalid.WithDetails("version, resource, name are required"))
		return
	}

	svc := services.NewServices()
	obj, err := svc.KubeCRGet(ctx.Request.Context(), cli, group, version, resource, namespace, name)
	if err != nil {
		r.ToErrorResponse(errorcode.ErrorCRGetFail.WithDetails(err.Error()))
		return
	}
	r.Success(obj)
}

// CreateCR 创建 CR（支持 DryRun）
func (c *DynamicCRDController) CreateCR(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	var req struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Resource  string `json:"resource"`
		Namespace string `json:"namespace"`
		Yaml      string `json:"yaml"`
		DryRun    bool   `json:"dry_run"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		r.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}
	if req.Version == "" || req.Resource == "" || req.Yaml == "" {
		r.ToErrorResponse(errorcode.ErrorCRGVRInvalid.WithDetails("version, resource, yaml are required"))
		return
	}

	svc := services.NewServices()
	created, dryRunResult, err := svc.KubeCRCreate(ctx.Request.Context(), cli, req.Group, req.Version, req.Resource, req.Namespace, req.Yaml, req.DryRun)
	if err != nil {
		if strings.Contains(err.Error(), "YAML") {
			r.ToErrorResponse(errorcode.ErrorCRYamlParseFail.WithDetails(err.Error()))
			return
		}
		r.ToErrorResponse(errorcode.ErrorCRCreateFail.WithDetails(err.Error()))
		return
	}

	if req.DryRun {
		r.Success(gin.H{"dry_run_result": dryRunResult})
		return
	}
	r.Success(gin.H{
		"message": "CR 创建成功",
		"name":    created.GetName(),
	})
}

// UpdateCR 更新 CR（支持 DryRun）
func (c *DynamicCRDController) UpdateCR(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	var req struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Resource  string `json:"resource"`
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Yaml      string `json:"yaml"`
		DryRun    bool   `json:"dry_run"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		r.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}
	if req.Version == "" || req.Resource == "" || req.Name == "" || req.Yaml == "" {
		r.ToErrorResponse(errorcode.ErrorCRGVRInvalid.WithDetails("version, resource, name, yaml are required"))
		return
	}

	svc := services.NewServices()
	updated, dryRunResult, err := svc.KubeCRUpdate(ctx.Request.Context(), cli, req.Group, req.Version, req.Resource, req.Namespace, req.Name, req.Yaml, req.DryRun)
	if err != nil {
		if strings.Contains(err.Error(), "YAML") {
			r.ToErrorResponse(errorcode.ErrorCRYamlParseFail.WithDetails(err.Error()))
			return
		}
		r.ToErrorResponse(errorcode.ErrorCRUpdateFail.WithDetails(err.Error()))
		return
	}

	if req.DryRun {
		r.Success(gin.H{"dry_run_result": dryRunResult})
		return
	}
	r.Success(gin.H{
		"message": "CR 更新成功",
		"name":    updated.GetName(),
	})
}

// DeleteCR 删除 CR
func (c *DynamicCRDController) DeleteCR(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	group := ctx.Query("group")
	version := ctx.Query("version")
	resource := ctx.Query("resource")
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	if version == "" || resource == "" || name == "" {
		r.ToErrorResponse(errorcode.ErrorCRGVRInvalid.WithDetails("version, resource, name are required"))
		return
	}

	svc := services.NewServices()
	if err := svc.KubeCRDelete(ctx.Request.Context(), cli, group, version, resource, namespace, name); err != nil {
		if strings.Contains(err.Error(), "删除保护") {
			r.ToErrorResponse(errorcode.ErrorCRDeleteProtected.WithDetails(err.Error()))
			return
		}
		r.ToErrorResponse(errorcode.ErrorCRDeleteFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{"message": "CR 删除成功", "name": name})
}

// GetCRYaml 获取 CR 的 YAML
func (c *DynamicCRDController) GetCRYaml(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	group := ctx.Query("group")
	version := ctx.Query("version")
	resource := ctx.Query("resource")
	namespace := ctx.Query("namespace")
	name := ctx.Query("name")

	if version == "" || resource == "" || name == "" {
		r.ToErrorResponse(errorcode.ErrorCRGVRInvalid.WithDetails("version, resource, name are required"))
		return
	}

	svc := services.NewServices()
	yamlStr, err := svc.KubeCRGetYaml(ctx.Request.Context(), cli, group, version, resource, namespace, name)
	if err != nil {
		r.ToErrorResponse(errorcode.ErrorCRGetFail.WithDetails(err.Error()))
		return
	}
	r.Success(gin.H{"yaml": yamlStr})
}

// DryRunCR 仅 DryRun 校验
func (c *DynamicCRDController) DryRunCR(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	cli := middlewares.MustGetK8sClients(ctx)

	var req struct {
		Group     string `json:"group"`
		Version   string `json:"version"`
		Resource  string `json:"resource"`
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		Yaml      string `json:"yaml"`
		IsUpdate  bool   `json:"is_update"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		r.ToErrorResponse(errorcode.InvalidParams.WithDetails(err.Error()))
		return
	}

	svc := services.NewServices()
	if req.IsUpdate {
		_, dryRunResult, err := svc.KubeCRUpdate(ctx.Request.Context(), cli, req.Group, req.Version, req.Resource, req.Namespace, req.Name, req.Yaml, true)
		if err != nil {
			r.ToErrorResponse(errorcode.ErrorCRDryRunFail.WithDetails(err.Error()))
			return
		}
		r.Success(dryRunResult)
	} else {
		_, dryRunResult, err := svc.KubeCRCreate(ctx.Request.Context(), cli, req.Group, req.Version, req.Resource, req.Namespace, req.Yaml, true)
		if err != nil {
			r.ToErrorResponse(errorcode.ErrorCRDryRunFail.WithDetails(err.Error()))
			return
		}
		r.Success(dryRunResult)
	}
}
