package cicd

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8soperation/global"
	"k8soperation/internal/app/services"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
)

// AppDiscovery 以应用为中心的发现结果
// 将工作负载及其关联的 ConfigMap/Secret/Service/PVC 聚合为一个"应用"
type AppDiscovery struct {
	AppName       string           `json:"app_name"`       // 从标签推断的应用名
	Namespace     string           `json:"namespace"`
	Workloads     []WorkloadInfo   `json:"workloads"`      // 工作负载列表
	ConfigMaps    []KVResource     `json:"configmaps"`     // 关联的 ConfigMap
	Secrets       []KVResource     `json:"secrets"`        // 关联的 Secret
	Services      []SvcResource    `json:"services"`       // 关联的 Service
	PVCs          []PVCResource    `json:"pvcs"`           // 关联的 PVC
}

type WorkloadInfo struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	ContainerName string `json:"container_name"`
	Image         string `json:"image"`
	ImageRepo     string `json:"image_repo"`
	ImageTag      string `json:"image_tag"`
	Replicas      int32  `json:"replicas"`
	Schedule      string `json:"schedule,omitempty"`
}

type KVResource struct {
	Name string            `json:"name"`
	Data map[string]string `json:"data,omitempty"`
	Keys []string          `json:"keys,omitempty"` // 只返回 key 名，不返回内容
}

type SvcResource struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`       // ClusterIP/NodePort/LoadBalancer
	ClusterIP string   `json:"cluster_ip"`
	Ports     []string `json:"ports"`      // "80/TCP", "443/TCP"
}

type PVCResource struct {
	Name         string `json:"name"`
	Size         string `json:"size"`
	StorageClass string `json:"storage_class"`
	AccessMode   string `json:"access_mode"`
}

// DiscoverApplications 发现命名空间下所有应用（以标签聚合工作负载+关联资源）
// @Router /api/v1/k8s/cicd/pipeline/discover-apps [get]
func (c *PipelineController) DiscoverApplications(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)

	clusterIDStr := ctx.Query("cluster_id")
	clusterID, err := strconv.ParseInt(clusterIDStr, 10, 64)
	if err != nil || clusterID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的集群ID"))
		return
	}
	namespace := ctx.Query("namespace")
	if namespace == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("命名空间不能为空"))
		return
	}

	svc := services.NewServices()
	factory := services.NewClusterClientFactory(svc)
	cli, err := factory.GetClient(ctx.Request.Context(), clusterID)
	if err != nil {
		global.Logger.Error("DiscoverApps: 集群连接失败", zap.Error(err))
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails("集群连接失败: "+err.Error()))
		return
	}

	// 第一步：扫描所有工作负载，提取 app 名和引用的资源名
	type rawApp struct {
		appName    string
		namespace  string
		workloads  []WorkloadInfo
		cmRefs     map[string]bool // ConfigMap 名称集合
		secretRefs map[string]bool // Secret 名称集合
		pvcRefs    map[string]bool // PVC 名称集合
	}
	appMap := make(map[string]*rawApp)

	getOrCreate := func(appName string) *rawApp {
		if a, ok := appMap[appName]; ok { return a }
		a := &rawApp{appName: appName, namespace: namespace, cmRefs: map[string]bool{}, secretRefs: map[string]bool{}, pvcRefs: map[string]bool{}}
		appMap[appName] = a
		return a
	}

	extractAppName := func(labels map[string]string, fallback string) string {
		if v := labels["app.kubernetes.io/name"]; v != "" { return v }
		if v := labels["app.kubernetes.io/part-of"]; v != "" { return v }
		if v := labels["app"]; v != "" { return v }
		if v := labels["name"]; v != "" { return v }
		return fallback
	}

	// Deployments
	if dps, e := cli.Kube.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, d := range dps.Items {
			appName := extractAppName(d.Labels, d.Name)
			app := getOrCreate(appName)
			rep := int32(1); if d.Spec.Replicas != nil { rep = *d.Spec.Replicas }
			for _, c := range d.Spec.Template.Spec.Containers {
				rp, tg := splitImage(c.Image)
				app.workloads = append(app.workloads, WorkloadInfo{"Deployment", d.Name, c.Name, c.Image, rp, tg, rep, ""})
				// 从 envFrom 收集 ConfigMap/Secret 引用
				for _, ef := range c.EnvFrom {
					if ef.ConfigMapRef != nil { app.cmRefs[ef.ConfigMapRef.Name] = true }
					if ef.SecretRef != nil { app.secretRefs[ef.SecretRef.Name] = true }
				}
			}
			// 从 volumes 收集 ConfigMap/Secret/PVC 引用
			for _, v := range d.Spec.Template.Spec.Volumes {
				if v.ConfigMap != nil { app.cmRefs[v.ConfigMap.Name] = true }
				if v.Secret != nil { app.secretRefs[v.Secret.SecretName] = true }
				if v.PersistentVolumeClaim != nil { app.pvcRefs[v.PersistentVolumeClaim.ClaimName] = true }
			}
		}
	}

	// StatefulSets
	if sts, e := cli.Kube.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, s := range sts.Items {
			appName := extractAppName(s.Labels, s.Name)
			app := getOrCreate(appName)
			rep := int32(1); if s.Spec.Replicas != nil { rep = *s.Spec.Replicas }
			for _, c := range s.Spec.Template.Spec.Containers {
				rp, tg := splitImage(c.Image)
				app.workloads = append(app.workloads, WorkloadInfo{"StatefulSet", s.Name, c.Name, c.Image, rp, tg, rep, ""})
				for _, ef := range c.EnvFrom {
					if ef.ConfigMapRef != nil { app.cmRefs[ef.ConfigMapRef.Name] = true }
					if ef.SecretRef != nil { app.secretRefs[ef.SecretRef.Name] = true }
				}
			}
			for _, v := range s.Spec.Template.Spec.Volumes {
				if v.ConfigMap != nil { app.cmRefs[v.ConfigMap.Name] = true }
				if v.Secret != nil { app.secretRefs[v.Secret.SecretName] = true }
				if v.PersistentVolumeClaim != nil { app.pvcRefs[v.PersistentVolumeClaim.ClaimName] = true }
			}
		}
	}

	// DaemonSets
	if dss, e := cli.Kube.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, d := range dss.Items {
			appName := extractAppName(d.Labels, d.Name)
			app := getOrCreate(appName)
			for _, c := range d.Spec.Template.Spec.Containers {
				rp, tg := splitImage(c.Image)
				app.workloads = append(app.workloads, WorkloadInfo{"DaemonSet", d.Name, c.Name, c.Image, rp, tg, 0, ""})
			}
			for _, v := range d.Spec.Template.Spec.Volumes {
				if v.ConfigMap != nil { app.cmRefs[v.ConfigMap.Name] = true }
				if v.Secret != nil { app.secretRefs[v.Secret.SecretName] = true }
				if v.PersistentVolumeClaim != nil { app.pvcRefs[v.PersistentVolumeClaim.ClaimName] = true }
			}
		}
	}

	// CronJobs
	if cjs, e := cli.Kube.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, cj := range cjs.Items {
			appName := extractAppName(cj.Labels, cj.Name)
			app := getOrCreate(appName)
			for _, c := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers {
				rp, tg := splitImage(c.Image)
				app.workloads = append(app.workloads, WorkloadInfo{"CronJob", cj.Name, c.Name, c.Image, rp, tg, 0, cj.Spec.Schedule})
			}
		}
	}

	// Jobs
	if jobs, e := cli.Kube.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, j := range jobs.Items {
			appName := extractAppName(j.Labels, j.Name)
			app := getOrCreate(appName)
			for _, c := range j.Spec.Template.Spec.Containers {
				rp, tg := splitImage(c.Image)
				app.workloads = append(app.workloads, WorkloadInfo{"Job", j.Name, c.Name, c.Image, rp, tg, 0, ""})
			}
		}
	}

	// 第二步：根据引用名称查找实际资源
	results := make([]AppDiscovery, 0, len(appMap))
	for _, app := range appMap {
		discovery := AppDiscovery{AppName: app.appName, Namespace: app.namespace, Workloads: app.workloads}

		// ConfigMaps
		for cmName := range app.cmRefs {
			if cm, e := cli.Kube.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, metav1.GetOptions{}); e == nil {
				keys := make([]string, 0, len(cm.Data))
				for k := range cm.Data { keys = append(keys, k) }
				discovery.ConfigMaps = append(discovery.ConfigMaps, KVResource{Name: cmName, Keys: keys})
			}
		}

		// Secrets
		for secName := range app.secretRefs {
			if sec, e := cli.Kube.CoreV1().Secrets(namespace).Get(ctx, secName, metav1.GetOptions{}); e == nil {
				keys := make([]string, 0, len(sec.Data))
				for k := range sec.Data { keys = append(keys, k) }
				discovery.Secrets = append(discovery.Secrets, KVResource{Name: secName, Keys: keys})
			}
		}

		// PVCs
		for pvcName := range app.pvcRefs {
			if pvc, e := cli.Kube.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{}); e == nil {
				size := ""
				if qty, ok := pvc.Spec.Resources.Requests["storage"]; ok { size = qty.String() }
				sc := ""; if pvc.Spec.StorageClassName != nil { sc = *pvc.Spec.StorageClassName }
				am := string(pvc.Spec.AccessModes[0])
				discovery.PVCs = append(discovery.PVCs, PVCResource{Name: pvcName, Size: size, StorageClass: sc, AccessMode: am})
			}
		}

		// 第三步：查找匹配的 Service（通过 label selector 匹配 app）
		if svcs, e := cli.Kube.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{}); e == nil {
			for _, svc := range svcs.Items {
				if svc.Spec.Selector == nil { continue }
				// 检查 selector 是否匹配 app 标签
				if svc.Spec.Selector["app"] == app.appName ||
					svc.Spec.Selector["app.kubernetes.io/name"] == app.appName {
					ports := make([]string, 0, len(svc.Spec.Ports))
					for _, p := range svc.Spec.Ports {
						ports = append(ports, p.Name+":"+p.TargetPort.String()+"/"+string(p.Protocol))
					}
					discovery.Services = append(discovery.Services, SvcResource{
						Name: svc.Name, Type: string(svc.Spec.Type),
						ClusterIP: svc.Spec.ClusterIP, Ports: ports,
					})
				}
			}
		}

		if len(discovery.Workloads) > 0 {
			results = append(results, discovery)
		}
	}

	rsp.Success(gin.H{"apps": results, "total": len(results)})
}

// ==================== 保留旧接口兼容 ====================

type WorkloadDiscovery struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	ContainerName string `json:"container_name"`
	Image         string `json:"image"`
	ImageRepo     string `json:"image_repo"`
	ImageTag      string `json:"image_tag"`
	Replicas      int32  `json:"replicas"`
	Schedule      string `json:"schedule,omitempty"`
}

func (c *PipelineController) DiscoverWorkloads(ctx *gin.Context) {
	rsp := response.NewResponse(ctx)
	clusterIDStr := ctx.Query("cluster_id")
	clusterID, err := strconv.ParseInt(clusterIDStr, 10, 64)
	if err != nil || clusterID <= 0 {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("无效的集群ID"))
		return
	}
	namespace := ctx.Query("namespace")
	if namespace == "" {
		rsp.ToErrorResponse(errorcode.InvalidParams.WithDetails("命名空间不能为空"))
		return
	}
	svc := services.NewServices()
	factory := services.NewClusterClientFactory(svc)
	cli, err := factory.GetClient(ctx.Request.Context(), clusterID)
	if err != nil {
		rsp.ToErrorResponse(errorcode.ServerError.WithDetails("集群连接失败: "+err.Error()))
		return
	}
	var list []WorkloadDiscovery
	addWL := func(kind, name, ns, cName, image string, rep int32, schedule string) {
		rp, tg := splitImage(image)
		list = append(list, WorkloadDiscovery{kind, name, ns, cName, image, rp, tg, rep, schedule})
	}
	if dps, e := cli.Kube.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, d := range dps.Items {
			rep := int32(1); if d.Spec.Replicas != nil { rep = *d.Spec.Replicas }
			for _, c := range d.Spec.Template.Spec.Containers { addWL("Deployment", d.Name, d.Namespace, c.Name, c.Image, rep, "") }
		}
	}
	if sts, e := cli.Kube.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, s := range sts.Items {
			rep := int32(1); if s.Spec.Replicas != nil { rep = *s.Spec.Replicas }
			for _, c := range s.Spec.Template.Spec.Containers { addWL("StatefulSet", s.Name, s.Namespace, c.Name, c.Image, rep, "") }
		}
	}
	if dss, e := cli.Kube.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, d := range dss.Items {
			for _, c := range d.Spec.Template.Spec.Containers { addWL("DaemonSet", d.Name, d.Namespace, c.Name, c.Image, 0, "") }
		}
	}
	if cjs, e := cli.Kube.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, cj := range cjs.Items {
			for _, c := range cj.Spec.JobTemplate.Spec.Template.Spec.Containers { addWL("CronJob", cj.Name, cj.Namespace, c.Name, c.Image, 0, cj.Spec.Schedule) }
		}
	}
	if jobs, e := cli.Kube.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{}); e == nil {
		for _, j := range jobs.Items {
			for _, c := range j.Spec.Template.Spec.Containers { addWL("Job", j.Name, j.Namespace, c.Name, c.Image, 0, "") }
		}
	}
	rsp.Success(gin.H{"workloads": list, "total": len(list)})
}

func splitImage(image string) (string, string) {
	if idx := strings.LastIndex(image, ":"); idx > 0 {
		after := image[idx+1:]
		if !strings.Contains(after, "/") { return image[:idx], after }
	}
	return image, "latest"
}
