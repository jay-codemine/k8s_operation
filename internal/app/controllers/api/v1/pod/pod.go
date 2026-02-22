package pod

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/middlewares"
	"k8soperation/pkg/app/response"
	"k8soperation/pkg/valid"
	podpkg "k8soperation/pkg/k8s/pod"
	"net/http"
	"strings"
)

type PodController struct{}

func NewPodController() *PodController {
	return &PodController{}
}

// Evict godoc
// @Summary 驱逐指定 Pod
// @Description 通过 Eviction API 驱逐某个 Pod（受 PDB 约束）
// @Tags K8s Pod管理
// @Accept json
// @Produce json
// @Param data body requests.KubePodEvictRequest true "Pod 驱逐参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 404 {object} errorcode.Error "Pod 不存在"
// @Failure 429 {object} errorcode.Error "被 PDB 限制"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/evict [post]
func (c *PodController) Evict(ctx *gin.Context) {
	r := response.NewResponse(ctx)
	param := requests.NewKubePodEvictRequest()

	// 2) 绑定 + 校验
	if ok := valid.Validate(ctx, param, requests.ValidKubePodEvictRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 3) 调 Service
	svc := services.NewServices()
	if err := svc.KubePodEvict(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubePodEvict error", zap.Error(err))
		return
	}

	// 4) 返回
	r.Success(gin.H{
		"message":   "Pod 驱逐成功",
		"namespace": param.Namespace,
		"podName":   param.PodName,
	})
}

// Create godoc
// @Summary 创建 K8s Pod
// @Description 创建一个 Kubernetes Pod（用于调试或临时用途）
// @Tags K8s Pod管理
// @Accept json
// @Produce json
// @Param body body requests.KubePodCreateRequest true "创建 Pod 请求参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "创建 Pod 失败"
// @Router /api/v1/k8s/pod [post]
func (c *PodController) Create(ctx *gin.Context) {
	param := requests.NewKubePodCreateRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	podObj, err := svc.KubePodCreate(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Error("创建Pod失败", zap.String("error", err.Error()))
		resp.ToErrorResponse(errorcode.ErrorPodCreateFail.WithDetails(err.Error()))
		return
	}

	resp.Success(podObj)
}

// CreateFromYaml godoc
// @Summary 从 YAML 创建 Pod
// @Tags K8s Pod管理
// @Accept json
// @Produce json
// @Param body body requests.YamlCreateRequest true "YAML 创建参数"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error
// @Failure 500 {object} errorcode.Error
// @Router /api/v1/k8s/pod/create-from-yaml [post]
func (c *PodController) CreateFromYaml(ctx *gin.Context) {
	req := requests.NewYamlCreateRequest()
	r := response.NewResponse(ctx)

	// 参数校验
	if ok := valid.Validate(ctx, req, requests.ValidYamlCreateRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()

	// 调用 Service 层创建逻辑
	podObj, err := svc.KubePodCreateFromYaml(ctx.Request.Context(), cli, req.Yaml)
	if err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubePodCreateFromYaml error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "Pod 创建成功",
		"name":      podObj.Name,
		"namespace": podObj.Namespace,
	})
}

// List godoc
// @Summary 列出K8s Pod
// @Description 列出K8s Pod
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Param page query int true "页码"
// @Param limit query int true "每页数量"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/list [get]
func (c *PodController) List(ctx *gin.Context) {
	param := requests.NewKubePodListRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodListRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	// 创建服务实例，传入上下文ctx
	svc := services.NewServices()
	// 调用服务实例的KubePodList方法获取Pod列表，传入参数param
	pods, total, err := svc.KubePodList(ctx.Request.Context(), cli, param)
	// 检查获取Pod列表时是否发生错误
	if err != nil {
		// 记录错误日志，包含错误信息
		global.Logger.Error("获取Pod列表失败", zap.String("error", err.Error()))
		// 返回错误响应，包含错误代码和详细信息
		resp.ToErrorResponse(errorcode.ErrorK8sPodListFail.WithDetails(err.Error()))
		// 终止当前函数执行
		return
	}

	// 使用新的转换函数，包含智能状态判断
	list := podpkg.BuildPodListResponse(pods)
	resp.SuccessList(list, total)
}

// Update godoc
// @Summary 更新Pod
// @Description 更新Pod
// @Tags K8s Pod管理
// @Produce json
// @Param body body requests.KubePodUpdateRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/update [post]
func (c *PodController) Update(ctx *gin.Context) {
	param := requests.NewKubePodUpdateRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodUpdateRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	if err := svc.KubePodUpdate(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Errorf("更新Pod失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sPodUpdateFail.WithDetails(err.Error()))
		return
	}
	resp.Success(gin.H{"msg": "Pod更新成功"})
}

// PatchImage godoc
// @Summary 更新 Pod 容器镜像
// @Description 基于 mergeKey=name 的 StrategicMergePatch 方式更新指定容器的镜像
// @Tags K8s Pod管理
// @Accept json
// @Produce json
// @Param body body requests.PatchPodImageRequest true "body"
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/patch_image [put]
func (c *PodController) PatchImage(ctx *gin.Context) {
	param := requests.NewPatchPodImageRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodPatchContainerImageRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	if err := svc.PatchPodImage(ctx.Request.Context(), cli, param); err != nil {
		global.Logger.Errorf("PatchPodImage 失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sPodPatchFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"msg":       "Pod 镜像更新成功",
		"namespace": param.Namespace,
		"name":      param.Name,
		"container": param.Container,
		"new_image": param.NewImage,
	})
}

// PatchLabels godoc
// @Summary 修改 Pod 标签
// @Description 添加或删除 Pod 标签
// @Tags K8s Pod管理
// @Accept json
// @Produce json
// @Param data body requests.KubePodLabelPatchRequest true "标签修改参数"
// @Success 200 {object} response.Response "成功"
// @Failure 400 {object} errorcode.Error "请求参数错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/labels [patch]
func (c *PodController) PatchLabels(ctx *gin.Context) {
	param := requests.NewKubePodLabelPatchRequest()
	r := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodLabelPatchRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()

	if err := svc.KubePodPatchLabels(ctx.Request.Context(), cli, param); err != nil {
		ctx.Error(err)
		global.Logger.Error("service.KubePodPatchLabels error", zap.Error(err))
		return
	}

	r.Success(gin.H{
		"message":   "Pod 标签修改成功",
		"namespace": param.Namespace,
		"name":      param.Name,
	})
}

// @Summary 删除 Pod
// @Description 删除指定命名空间下的 Pod（支持优雅终止/强制删除）
// @Tags K8s Pod管理
// @Produce json
// @Param namespace query string true  "命名空间"
// @Param name      query string true  "Pod 名称"
// @Param grace_seconds query int false "优雅终止秒数（默认30）"
// @Param force     query bool  false "是否强制删除（默认false）"
// @Success 200 {object} map[string]interface{} "删除请求已提交"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/grace_delete_pod [delete]
func (c *PodController) DeletePod(ctx *gin.Context) {
	resp := response.NewResponse(ctx)

	param := requests.NewKubePodDeleteRequest()
	if ok := valid.Validate(ctx, param, requests.ValidKubePodDeleteRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	if err := svc.KubePodDelete(ctx.Request.Context(), cli, param); err != nil {
		resp.ToErrorResponse(errorcode.ErrorK8sPodDeleteFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"namespace":     param.Namespace,
		"name":          param.Name,
		"force":         param.Force,
		"grace_seconds": param.GraceSeconds,
		"message":       "删除成功",
	})
}

// Detail godoc
// @Summary 获取Pod的详情
// @Description 获取Pod的详情
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/detail [get]
func (c *PodController) Detail(ctx *gin.Context) {
	param := requests.NewKubePodDetailRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodDetailRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	pod, err := svc.KubePodDetail(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Errorf("获取Pod详情失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sPodDetailFail.WithDetails(err.Error()))
		return
	}

	resp.Success(pod)
}

// GetContainerName godoc
// @Summary 获取Pod的容器名
// @Description 获取Pod的容器名
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/container_name [get]
func (c *PodController) GetContainerName(ctx *gin.Context) {
	param := requests.NewKubePodDetailRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodDetailRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	containerName, err := svc.GetContainerNames(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Errorf("获取Pod容器名失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetContainerName.WithDetails(err.Error()))
		return
	}

	resp.Success(containerName)
}

// GetInitContainerName godoc
// @Summary 获取Pod的Init容器名
// @Description 获取Pod的Init容器名
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/init_container_name [get]
func (c *PodController) GetInitContainerName(ctx *gin.Context) {
	param := requests.NewKubeCommonRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.VaildKubeCommonRequest); !ok {
		global.Logger.Errorf("获取Pod的Init容器名失败: %v", ok)
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	initContainerName, err := svc.GetInitContainerNames(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Errorf("获取Pod的Init容器名失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetInitContainerName.WithDetails(err.Error()))
		return
	}

	resp.Success(initContainerName)
}

// GetContainerImages godoc
// @Summary 获取Pod的容器镜像
// @Description 获取Pod的容器镜像
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/container_image [get]
func (c *PodController) GetContainerImages(ctx *gin.Context) {
	param := requests.NewKubePodDetailRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodDetailRequest); !ok {
		global.Logger.Errorf("校验失败：%v", ok)
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	containerImages, err := svc.GetContainerImages(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Errorf("获取Pod容器镜像失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetContainerImage.WithDetails(err.Error()))
		return
	}

	global.Logger.Infof("获取Pod容器镜像成功，总数: %d", len(containerImages))
	resp.SuccessList(containerImages, len(containerImages))
}

// GetInitContainerImages godoc
// @Summary 获取Pod的Init容器镜像
// @Description 获取Pod的Init容器镜像
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/init_container_image [get]
func (c *PodController) GetInitContainerImages(ctx *gin.Context) {
	param := requests.NewKubeCommonRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.VaildKubeCommonRequest); !ok {
		global.Logger.Errorf("校验失败：%v", ok)
		resp.ToErrorResponse(errorcode.InvalidParams)
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	initContainerImages, err := svc.GetInitContainerImages(ctx.Request.Context(), cli, param)
	if err != nil {
		global.Logger.Errorf("获取Pod的Init容器镜像失败: %v", err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetInitContainerImage.WithDetails(err.Error()))
		return
	}

	global.Logger.Infof("获取Pod的Init容器镜像成功: %d", len(initContainerImages))
	resp.SuccessList(initContainerImages, len(initContainerImages))
}

// GetContainerLog godoc
// @Summary 实时跟随 Pod 容器日志
// @Description 由全局开关 PodLog.EnableStreaming 控制：true=实时流式(text/plain)，false=一次性(JSON)
// @Tags K8s Pod管理
// @Produce json
// @Produce text/plain
// @Param name query string true "Pod名" maxlength(100)
// @Param namespace query string true "命名空间" maxlength(100)
// @Param container query string false "容器名(可选; 多容器建议指定)" maxlength(100)
// @Param tail query int false "仅返回最后N行(默认见配置)"
// @Success 200 {object} map[string]interface{} "成功(EnableStreaming=false)"
// @Success 200 {string} string "流式文本(EnableStreaming=true)"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/container_log [get]
func (c *PodController) GetContainerLogs(ctx *gin.Context) {
	resp := response.NewResponse(ctx)
	param := requests.NewKubePodLogRequest()

	if ok := valid.Validate(ctx, param, requests.ValidKubePodLogRequest); !ok {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("参数校验失败"))
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)
	svc := services.NewServices()
	rctx := ctx.Request.Context()

	if param.Follow {
		rc, err := svc.KubePodLogStream(rctx, cli,
			param.Name, param.Namespace, param.Container, param.Tail,
		)
		if err != nil {
			resp.ToErrorResponse(errorcode.ErrorK8sGetContainerLog.WithDetails(err.Error()))
			return
		}
		defer rc.Close()

		w := ctx.Writer
		h := w.Header()
		h.Set("Content-Type", "text/plain; charset=utf-8")
		h.Set("Cache-Control", "no-cache, no-store")
		h.Set("Pragma", "no-cache")
		h.Set("X-Accel-Buffering", "no")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			resp.ToErrorResponse(errorcode.ErrorK8sGetContainerLog.WithDetails("streaming unsupported"))
			return
		}
		flusher.Flush()

		buf := make([]byte, 4096)
		for {
			select {
			case <-rctx.Done():
				global.Logger.Debugf("stream closed by client ns=%s pod=%s container=%s",
					param.Namespace, param.Name, param.Container)
				return
			default:
			}

			n, rerr := rc.Read(buf)
			if n > 0 {
				if _, werr := w.Write(buf[:n]); werr != nil {
					global.Logger.Debugf("stream write closed ns=%s pod=%s container=%s err=%v",
						param.Namespace, param.Name, param.Container, werr)
					return
				}
				flusher.Flush()
			}

			if rerr != nil {
				if errors.Is(rerr, io.EOF) {
					return
				}
				global.Logger.Errorf("stream read err ns=%s pod=%s container=%s : %v",
					param.Namespace, param.Name, param.Container, rerr)
				return
			}
		}
		
		// 关键：follow 分支必须 return（虽然上面都 return 了，但保持结构安全）
		return
	}

	// —— 一次性 —— //
	logStr, err := svc.KubePodLog(rctx, cli,
		param.Name, param.Namespace, param.Container, param.Tail,
	)
	if err != nil {
		tailVal := int64(-1)
		if param.Tail != nil {
			tailVal = *param.Tail
		}
		global.Logger.Errorf("get pod log failed ns=%s pod=%s container=%s tail=%d : %v",
			param.Namespace, param.Name, param.Container, tailVal, err)
		resp.ToErrorResponse(errorcode.ErrorK8sGetContainerLog.WithDetails(err.Error()))
		return
	}

	tailVal := int64(-1)
	if param.Tail != nil {
		tailVal = *param.Tail
	}

	resp.Success(gin.H{
		"namespace": param.Namespace,
		"pod":       param.Name,
		"container": param.Container,
		"tail":      tailVal,
		"follow":    param.Follow,
		"log":       logStr,
	})
}

// GetContainerLog godoc
// @Summary 获取Pod的容器日志
// @Description 获取Pod的容器日志
// @Tags K8s Pod管理
// @Produce json
// @Param name query string false "Pod名" maxlength(100)
// @Param namespace query string false "命名空间" maxlength(100)
// @Param container query string false "容器" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/container_logs [get]
func (k *PodController) GetContainerLog(ctx *gin.Context) {
	param := requests.NewKubePodLogRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodLogRequest); !ok {
		return
	}
	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	logs, err := svc.KubePodLog(ctx.Request.Context(), cli, param.Name, param.Namespace, param.Container, param.Tail)
	if err != nil {
		global.Logger.Error("获取 Pod 日志失败",
			zap.String("namespace", param.Namespace),
			zap.String("pod", param.Name),
			zap.String("container", param.Container),
			zap.Int64p("tail", param.Tail),
			zap.Error(err),
		)
		// 检查是否是容器还未就绪
		if strings.Contains(err.Error(), "容器还未就绪") {
			resp.ToErrorResponse(errorcode.ErrorPodContainerNotReady.WithDetails(err.Error()))
			return
		}
		resp.ToErrorResponse(errorcode.ErrorK8sGetContainerLog.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"namespace": param.Namespace,
		"pod":       param.Name,
		"container": param.Container,
		"log":       logs,
	})
}

// Metrics godoc
// @Summary 获取单个 Pod 的资源使用情况
// @Description 获取 Pod 的 CPU 和内存使用情况（需要 metrics-server）
// @Tags K8s Pod管理
// @Produce json
// @Param namespace query string true "命名空间" maxlength(100)
// @Param name query string true "Pod名称" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 503 {object} errorcode.Error "metrics-server 不可用"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/metrics [get]
func (c *PodController) Metrics(ctx *gin.Context) {
	param := requests.NewKubePodDetailRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	metrics, err := svc.KubePodMetrics(ctx.Request.Context(), cli, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Error("获取 Pod metrics 失败",
			zap.String("namespace", param.Namespace),
			zap.String("name", param.Name),
			zap.Error(err),
		)
		
		// 判断是否是 metrics-server 不可用
		if err.Error() == "metrics-server 未安装或不可用" {
			resp.ToErrorResponse(errorcode.ErrorMetricsServerUnavailable.WithDetails(err.Error()))
			return
		}
		
		resp.ToErrorResponse(errorcode.ErrorK8sGetPodMetrics.WithDetails(err.Error()))
		return
	}

	resp.Success(metrics)
}

// MetricsList godoc
// @Summary 批量获取命名空间下所有 Pod 的资源使用情况
// @Description 获取指定命名空间下所有 Pod 的 CPU 和内存使用情况（需要 metrics-server）
// @Tags K8s Pod管理
// @Produce json
// @Param namespace query string true "命名空间" maxlength(100)
// @Success 200 {object} string "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 503 {object} errorcode.Error "metrics-server 不可用"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/metrics/list [get]
func (c *PodController) MetricsList(ctx *gin.Context) {
	// 使用 namespace query 参数
	namespace := ctx.Query("namespace")
	resp := response.NewResponse(ctx)

	if namespace == "" {
		resp.ToErrorResponse(errorcode.InvalidParams.WithDetails("namespace 参数必填"))
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	svc := services.NewServices()
	metricsMap, err := svc.KubePodsMetrics(ctx.Request.Context(), cli, namespace)
	if err != nil {
		global.Logger.Error("批量获取 Pod metrics 失败",
			zap.String("namespace", namespace),
			zap.Error(err),
		)
		
		// 判断是否是 metrics-server 不可用
		if err.Error() == "metrics-server 未安装或不可用" {
			resp.ToErrorResponse(errorcode.ErrorMetricsServerUnavailable.WithDetails(err.Error()))
			return
		}
		
		resp.ToErrorResponse(errorcode.ErrorK8sGetPodMetrics.WithDetails(err.Error()))
		return
	}

	resp.Success(metricsMap)
}

// Yaml godoc
// @Summary 获取 Pod 的 YAML 配置
// @Description 获取指定 Pod 的 YAML 格式配置
// @Tags K8s Pod管理
// @Produce json
// @Param namespace query string true "命名空间" maxlength(100)
// @Param name query string true "Pod名称" maxlength(100)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/yaml [get]
func (c *PodController) Yaml(ctx *gin.Context) {
	param := requests.NewKubePodDetailRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodDetailRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	yamlStr, err := podpkg.GetYaml(ctx.Request.Context(), cli.Kube, param.Namespace, param.Name)
	if err != nil {
		global.Logger.Error("获取 Pod YAML 失败",
			zap.String("namespace", param.Namespace),
			zap.String("name", param.Name),
			zap.Error(err),
		)
		resp.ToErrorResponse(errorcode.ErrorK8sPodDetailFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"yaml":      yamlStr,
	})
}

// ApplyYaml godoc
// @Summary 应用 Pod YAML 配置
// @Description 应用修改后的 YAML 配置到 Pod
// @Tags K8s Pod管理
// @Accept json
// @Produce json
// @Param body body requests.KubePodApplyYamlRequest true "YAML内容"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} errorcode.Error "请求错误"
// @Failure 500 {object} errorcode.Error "内部错误"
// @Router /api/v1/k8s/pod/apply_yaml [put]
func (c *PodController) ApplyYaml(ctx *gin.Context) {
	param := requests.NewKubePodApplyYamlRequest()
	resp := response.NewResponse(ctx)

	if ok := valid.Validate(ctx, param, requests.ValidKubePodApplyYamlRequest); !ok {
		return
	}

	cli := middlewares.MustGetK8sClients(ctx)

	_, err := podpkg.ApplyYaml(ctx.Request.Context(), cli.Kube, param.Namespace, param.Name, param.Yaml)
	if err != nil {
		global.Logger.Error("应用 Pod YAML 失败",
			zap.String("namespace", param.Namespace),
			zap.String("name", param.Name),
			zap.Error(err),
		)
		resp.ToErrorResponse(errorcode.ErrorK8sPodUpdateFail.WithDetails(err.Error()))
		return
	}

	resp.Success(gin.H{
		"namespace": param.Namespace,
		"name":      param.Name,
		"message":   "YAML 应用成功",
	})
}
