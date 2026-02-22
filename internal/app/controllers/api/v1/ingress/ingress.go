package v1

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
)

type KubeIngressController struct{}

func NewKubeIngressController() *KubeIngressController {
	return &KubeIngressController{}
}

// Create godoc
// @Summary 创建 Ingress
// @Tags K8s Ingress 管理
// @Accept json
// @Produce json
// @Param body body requests.KubeIngressCreateRequest true "创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/create [post]
func (c *KubeIngressController) Create(ctx *gin.Context) {
	req := requests.NewKubeIngressCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidKubeIngressCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Ingress 创建逻辑
	ing, err := svc.KubeIngressCreate(ctx.Request.Context(), cli, req)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeIngressCreate error", zap.Error(err))
		return
	}

	// 和 Job 一样封装一下返回；如果你有统一的 Build*Response，也可替换为 ingress.BuildIngressResponse(…)
	r.Success(gin.H{
		"message":   "创建 Ingress 成功",
		"name":      ing.Name,
		"namespace": ing.Namespace,
	})
}

// List godoc
// @Summary 获取 K8s Ingress 列表
// @Description 支持分页、命名空间过滤、名称模糊查询
// @Tags K8s Ingress 管理
// @Produce json
// @Param namespace query string false "命名空间" maxlength(100)
// @Param name query string false "Ingress 名称(模糊匹配)" maxlength(100)
// @Param page query int true "页码 (从1开始)"
// @Param limit query int true "每页数量 (默认20)"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/ingress/list [get]
func (c *KubeIngressController) List(ctx *gin.Context) {
	param := requests.NewKubeIngressListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeIngressListRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	ingresses, total, err := svc.KubeIngressList(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("获取Ingress列表失败", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorIngressQueryFail.WithDetails(err.Error()))
		return
	}

	// 转换为前端期望的格式
	var list []gin.H
	for _, ing := range ingresses {
		// 构建 Host 列表
		var hosts []string
		for _, rule := range ing.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
		}
		hostStr := "-"
		if len(hosts) > 0 {
			hostStr = strings.Join(hosts, ", ")
		}

		// 构建 Path 列表
		var paths []string
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					paths = append(paths, path.Path)
				}
			}
		}
		pathStr := "-"
		if len(paths) > 0 {
			pathStr = strings.Join(paths, ", ")
		}

		// 构建 Service 列表
		var services []string
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil {
						svcName := path.Backend.Service.Name
						if path.Backend.Service.Port.Number > 0 {
							svcName += fmt.Sprintf(":%d", path.Backend.Service.Port.Number)
						} else if path.Backend.Service.Port.Name != "" {
							svcName += ":" + path.Backend.Service.Port.Name
						}
						services = append(services, svcName)
					}
				}
			}
		}
		serviceStr := "-"
		if len(services) > 0 {
			serviceStr = strings.Join(services, ", ")
		}

		// TLS 状态
		tlsEnabled := len(ing.Spec.TLS) > 0

		// IngressClass
		ingressClass := "-"
		if ing.Spec.IngressClassName != nil {
			ingressClass = *ing.Spec.IngressClassName
		}

		// Address （LoadBalancer IP）
		var addresses []string
		for _, lb := range ing.Status.LoadBalancer.Ingress {
			if lb.IP != "" {
				addresses = append(addresses, lb.IP)
			} else if lb.Hostname != "" {
				addresses = append(addresses, lb.Hostname)
			}
		}
		addressStr := "-"
		if len(addresses) > 0 {
			addressStr = strings.Join(addresses, ", ")
		}

		list = append(list, gin.H{
			"name":          ing.Name,
			"namespace":     ing.Namespace,
			"hosts":         hostStr,
			"paths":         pathStr,
			"services":      serviceStr,
			"ingress_class": ingressClass,
			"tls":           tlsEnabled,
			"address":       addressStr,
			"labels":        ing.Labels,
			"created_at":    ing.CreationTimestamp.Format("2006-01-02 15:04:05"),
		})
	}

	resp.SuccessList(list, gin.H{
		"total":   total,
		"message": fmt.Sprintf("获取 Ingress 列表成功，共 %d 条数据", total),
	})
}

// Detail godoc
// @Summary 获取 Ingress 详情
// @Tags K8s Ingress 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/detail [get]
func (c *KubeIngressController) Detail(ctx *gin.Context) {
	// 1. 构造请求参数
	param := requests.NewKubeIngressDetailRequest()
	r := response.NewResponse(ctx)

	// 2. 参数校验
	if ok := valid.Validate(ctx, param, requests.ValidKubeIngressDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 3. 调用 Service 层
	svc := services.NewServices()
	ing, err := svc.KubeIngressDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubeIngressDetail error", zap.Error(err))
		return
	}

	// 4. 返回成功响应
	r.Success(gin.H{
		"message": "获取 Ingress 详情成功",
		"data":    ing,
	})
}

// @Summary Patch Ingress（StrategicMergePatch）
// @Tags K8s Ingress 管理
// @Accept application/strategic-merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/patch [patch]
func (c *KubeIngressController) Patch(ctx *gin.Context) {
	param := requests.NewKubeIngressUpdateRequest() // 如果你有 NewKubeIngressUpdateRequest() 也可用它
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	out, err := svc.KubeIngressPatch(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeIngressPatch error", zap.Error(err))
		return
	}
	r.Success(gin.H{"message": "Ingress StrategicMergePatch 成功", "data": out})
}

// @Summary Patch Ingress（JSON Merge Patch – 覆盖式）
// @Tags K8s Ingress 管理
// @Accept application/merge-patch+json
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress 名称"
// @Param content body string true "Patch Body(JSON字符串)"
// @Success 200 {object} string
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/patch-json [post]
func (c *KubeIngressController) PatchJSON(ctx *gin.Context) {
	param := requests.NewKubeIngressUpdateRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, &param, nil); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	out, err := svc.KubeIngressPatchJSON(ctx.Request.Context(), cli, param)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeIngressPatchJSON error", zap.Error(err))
		return
	}
	r.Success(gin.H{"message": "Ingress JSON Merge Patch 成功", "data": out})
}

// Delete godoc
// @Summary 删除 Ingress
// @Tags K8s Ingress 管理
// @Produce json
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress 名称"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/delete [delete]
func (c *KubeIngressController) Delete(ctx *gin.Context) {
	param := requests.NewKubeIngressDeleteRequest()
	r := response.NewResponse(ctx)

	// 参数校验（通用 Valid）
	if ok := valid.Validate(ctx, param, requests.ValidKubeIngressDeleteRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 调用服务层
	svc := services.NewServices()
	if err := svc.KubeIngressDelete(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Error("service.KubeIngressDelete error", zap.Error(err))
		ctx.Error(err)
		return
	}

	// 成功响应
	r.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "Ingress 删除成功",
	})
}

// Yaml godoc
// @Summary 获取 Ingress 的 YAML
// @Tags K8s Ingress 管理
// @Produce plain
// @Param namespace query string true "命名空间"
// @Param name query string true "Ingress 名称"
// @Success 200 {string} string "YAML 内容"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/yaml [get]
func (c *KubeIngressController) Yaml(ctx *gin.Context) {
	param := requests.NewKubeIngressDetailRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubeIngressDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()

	yamlContent, err := svc.KubeIngressYaml(ctx.Request.Context(), cli, param.Namespace, param.Name)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeIngressYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"yaml": yamlContent,
	})
}

// ApplyYaml godoc
// @Summary 应用 Ingress YAML
// @Tags K8s Ingress 管理
// @Accept json
// @Produce json
// @Param body body object{yaml=string} true "YAML 内容"
// @Success 200 {object} string
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/apply-yaml [put]
func (c *KubeIngressController) ApplyYaml(ctx *gin.Context) {
	var req struct {
		Yaml string `json:"yaml" valid:"yaml"`
	}
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, &req, nil); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()

	ing, err := svc.KubeIngressApplyYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeIngressApplyYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "Ingress YAML 应用成功",
		"name":      ing.Name,
		"namespace": ing.Namespace,
	})
}

// CreateFromYaml godoc
// @Summary 从 YAML 创建 Ingress
// @Tags K8s Ingress 管理
// @Accept json
// @Produce json
// @Param body body object{yaml=string} true "YAML 内容"
// @Success 200 {object} string
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/ingress/create-from-yaml [post]
func (c *KubeIngressController) CreateFromYaml(ctx *gin.Context) {
	var req struct {
		Yaml string `json:"yaml" valid:"yaml"`
	}
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, &req, nil); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()

	ing, err := svc.KubeIngressCreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("KubeIngressCreateFromYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "Ingress 创建成功",
		"name":      ing.Name,
		"namespace": ing.Namespace,
	})
}
