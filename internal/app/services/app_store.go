package services

import (
	"context"
	"encoding/json"
	"fmt"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"strings"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ============================================================
// 应用商城 Service
// ============================================================

// AppStoreList 获取应用列表（分页 + 筛选）
func (s *Services) AppStoreList(ctx context.Context, req *models.AppStoreListRequest) ([]*models.AppStoreApp, int64, error) {
	return s.dao.AppStoreList(ctx, req)
}

// AppStoreDetail 获取应用详情
func (s *Services) AppStoreDetail(ctx context.Context, id uint32) (*models.AppStoreApp, error) {
	return s.dao.AppStoreGetByID(ctx, id)
}

// AppStoreCreate 创建应用
func (s *Services) AppStoreCreate(ctx context.Context, req *models.AppStoreCreateRequest) (*models.AppStoreApp, error) {
	// 检查名称唯一性
	existing, _ := s.dao.AppStoreGetByName(ctx, req.Name)
	if existing != nil {
		return nil, fmt.Errorf("应用名称 '%s' 已存在", req.Name)
	}

	app := &models.AppStoreApp{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Category:    req.Category,
		Version:     req.Version,
		Icon:        req.Icon,
		Description: req.Description,
		Provider:    req.Provider,
		ChartURL:    req.ChartURL,
		DocURL:      req.DocURL,
		Status:      req.Status,
		Featured:    req.Featured,
		SortOrder:   req.SortOrder,
		Tags:        req.Tags,
		MinK8s:      req.MinK8s,
		Namespace:   req.Namespace,
		ValuesYAML:  req.ValuesYAML,
	}

	if app.Status == 0 {
		app.Status = 1 // 默认可用
	}
	if app.Provider == "" {
		app.Provider = "official"
	}

	now := uint32(time.Now().Unix())
	app.CreatedAt = now
	app.ModifiedAt = now

	if err := s.dao.AppStoreCreate(ctx, app); err != nil {
		global.Logger.Error("创建应用失败", zap.Error(err))
		return nil, err
	}

	return app, nil
}

// AppStoreUpdate 更新应用
func (s *Services) AppStoreUpdate(ctx context.Context, req *models.AppStoreUpdateRequest) error {
	existing, err := s.dao.AppStoreGetByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("应用不存在: %w", err)
	}

	// 更新字段
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.DisplayName != "" {
		existing.DisplayName = req.DisplayName
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Version != "" {
		existing.Version = req.Version
	}
	if req.Icon != "" {
		existing.Icon = req.Icon
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Provider != "" {
		existing.Provider = req.Provider
	}
	existing.ChartURL = req.ChartURL
	existing.DocURL = req.DocURL
	if req.Status > 0 {
		existing.Status = req.Status
	}
	existing.Featured = req.Featured
	existing.SortOrder = req.SortOrder
	existing.Tags = req.Tags
	existing.MinK8s = req.MinK8s
	existing.Namespace = req.Namespace
	existing.ValuesYAML = req.ValuesYAML
	existing.ModifiedAt = uint32(time.Now().Unix())

	return s.dao.AppStoreUpdate(ctx, existing)
}

// AppStoreDelete 删除应用
func (s *Services) AppStoreDelete(ctx context.Context, id uint32) error {
	_, err := s.dao.AppStoreGetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("应用不存在: %w", err)
	}
	return s.dao.AppStoreDelete(ctx, id)
}

// AppStoreCategories 获取分类列表
func (s *Services) AppStoreCategories(ctx context.Context) ([]models.AppStoreCategoryCount, error) {
	return s.dao.AppStoreCategories(ctx)
}

// ============================================================
// 安装相关 Service
// ============================================================

// AppStoreInstall 安装应用到目标集群
// 流程：校验应用 → 校验集群 → 查重 → 创建NS → 写安装记录 → 异步安装
func (s *Services) AppStoreInstall(ctx context.Context, factory *ClusterClientFactory, req *models.AppStoreInstallRequest, operator string) (*models.AppStoreInstall, error) {
	// 1. 校验应用存在
	app, err := s.dao.AppStoreGetByID(ctx, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("应用不存在: %w", err)
	}

	// 2. 校验集群存在并获取集群信息
	cluster, err := s.dao.KubeClusterGetByID(ctx, req.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("集群不存在: %w", err)
	}

	// 3. 检查是否已安装（防重复安装）
	existing, _ := s.dao.AppStoreInstallFindActive(ctx, req.AppID, req.ClusterID, req.Namespace)
	if existing != nil {
		return nil, fmt.Errorf("应用 %s 已在集群 %s 的 %s 命名空间安装(状态: %d)",
			app.Name, cluster.ClusterName, req.Namespace, existing.Status)
	}

	// 4. 获取集群客户端
	cli, err := factory.Get(ctx, req.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("连接集群失败: %w", err)
	}

	// 5. 确保目标命名空间存在
	if err := s.ensureNamespace(ctx, cli, req.Namespace); err != nil {
		global.Logger.Warn("创建命名空间失败（可能已存在）",
			zap.String("namespace", req.Namespace), zap.Error(err))
		// 不中断流程，命名空间可能已存在
	}

	// 6. 创建安装记录
	now := uint32(time.Now().Unix())
	install := &models.AppStoreInstall{
		AppID:       req.AppID,
		AppName:     app.DisplayName,
		ClusterID:   req.ClusterID,
		ClusterName: cluster.ClusterName,
		Namespace:   req.Namespace,
		ReleaseName: req.RelName,
		Version:     app.Version,
		Values:      req.Values,
		Status:      models.InstallStatusInstalling,
		Message:     "安装中...",
		Operator:    operator,
	}
	if install.AppName == "" {
		install.AppName = app.Name
	}
	install.CreatedAt = now
	install.ModifiedAt = now

	if err := s.dao.AppStoreInstallCreate(ctx, install); err != nil {
		return nil, fmt.Errorf("创建安装记录失败: %w", err)
	}

	// 7. 执行安装（同步，后续可改为异步）
	go s.doInstall(install, cli, app)

	return install, nil
}

// doInstall 异步执行安装逻辑
// 参考 Rancher App Catalog 设计：一个应用包含多个组件，每个组件创建独立的 Deployment + Service
func (s *Services) doInstall(install *models.AppStoreInstall, cli *K8sClients, app *models.AppStoreApp) {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	now := uint32(time.Now().Unix())
	ns := install.Namespace
	relName := install.ReleaseName
	managedLabels := map[string]string{
		"app.kubernetes.io/managed-by": "k8s-operation-appstore",
		"app.kubernetes.io/name":       app.Name,
		"app.kubernetes.io/version":    app.Version,
		"app.kubernetes.io/instance":   relName,
		"appstore.k8s-operation/id":    fmt.Sprintf("%d", install.ID),
	}

	// ---- Step 1: 创建安装标记 ConfigMap ----
	cmName := fmt.Sprintf("appstore-%s", relName)
	components := resolveAppComponents(app.Name)
	// 优先从数据库加载组件定义（动态配置），无则用硬编码 fallback
	dbComps, dbErr := s.dao.AppStoreComponentListByAppID(ctx, app.ID)
	if dbErr == nil && len(dbComps) > 0 {
		components = dbComponentsToSpecs(dbComps)
		global.Logger.Info("使用数据库组件定义",
			zap.String("app", app.Name), zap.Int("components", len(components)))
	} else {
		global.Logger.Info("使用内置组件定义(数据库无配置)",
			zap.String("app", app.Name), zap.Int("components", len(components)))
	}
	componentNames := make([]string, 0, len(components))
	for _, c := range components {
		componentNames = append(componentNames, c.Name)
	}
	cmData := map[string]string{
		"app_name":     app.Name,
		"display_name": app.DisplayName,
		"version":      app.Version,
		"category":     app.Category,
		"description":  app.Description,
		"chart_url":    app.ChartURL,
		"doc_url":      app.DocURL,
		"installed_at": time.Now().Format(time.RFC3339),
		"release_name": relName,
		"namespace":    ns,
		"components":   fmt.Sprintf("%v", componentNames),
	}
	if install.Values != "" {
		cmData["custom_values"] = install.Values
	}
	if err := s.createInstallMarkerConfigMap(ctx, cli, ns, cmName, managedLabels, cmData); err != nil {
		s.failInstall(ctx, install.ID, now, "创建安装标记失败", err, app.Name, ns)
		return
	}

	// ---- Step 2: 为每个组件创建 Deployment + Service ----
	totalComponents := len(components)
	createdCount := 0

	for _, comp := range components {
		deployName := fmt.Sprintf("%s-%s", relName, comp.Name)
		replicas := comp.Replicas

		// 组件标签（继承 managed 标签 + 组件标识）
		compLabels := map[string]string{}
		for k, v := range managedLabels {
			compLabels[k] = v
		}
		compLabels["app.kubernetes.io/component"] = comp.Name

		// 构建容器端口
		containerPorts := make([]corev1.ContainerPort, 0, len(comp.Ports))
		for _, p := range comp.Ports {
			containerPorts = append(containerPorts, corev1.ContainerPort{
				Name:          p.Name,
				ContainerPort: p.Port,
				Protocol:      corev1.ProtocolTCP,
			})
		}
		if len(containerPorts) == 0 {
			containerPorts = []corev1.ContainerPort{{Name: "http", ContainerPort: 80, Protocol: corev1.ProtocolTCP}}
		}

		// 构建容器 args
		var containerArgs []string
		if len(comp.Args) > 0 {
			containerArgs = comp.Args
		}

		deploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployName,
				Namespace: ns,
				Labels:    compLabels,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app.kubernetes.io/instance":  relName,
						"app.kubernetes.io/component": comp.Name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: compLabels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  comp.Name,
								Image: comp.Image,
								Ports: containerPorts,
								Args:  containerArgs,
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse(comp.CPUReq),
										corev1.ResourceMemory: resource.MustParse(comp.MemReq),
									},
									Limits: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse(comp.CPULim),
										corev1.ResourceMemory: resource.MustParse(comp.MemLim),
									},
								},
							},
						},
					},
				},
			},
		}

		// 创建或更新 Deployment
		existingDeploy, err := cli.Kube.AppsV1().Deployments(ns).Get(ctx, deployName, metav1.GetOptions{})
		if err == nil {
			existingDeploy.Labels = compLabels
			existingDeploy.Spec = deploy.Spec
			_, err = cli.Kube.AppsV1().Deployments(ns).Update(ctx, existingDeploy, metav1.UpdateOptions{})
		} else if apierrors.IsNotFound(err) {
			_, err = cli.Kube.AppsV1().Deployments(ns).Create(ctx, deploy, metav1.CreateOptions{})
		}
		if err != nil {
			global.Logger.Error("创建组件 Deployment 失败",
				zap.String("component", comp.Name), zap.String("deploy", deployName), zap.Error(err))
			continue // 非致命，继续创建其他组件
		}

		// 为有端口的组件创建 Service
		if len(comp.Ports) > 0 {
			svcPorts := make([]corev1.ServicePort, 0, len(comp.Ports))
			for _, p := range comp.Ports {
				svcPorts = append(svcPorts, corev1.ServicePort{
					Name:       p.Name,
					Port:       p.Port,
					TargetPort: intstr.FromString(p.Name),
					Protocol:   corev1.ProtocolTCP,
				})
			}
			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deployName,
					Namespace: ns,
					Labels:    compLabels,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app.kubernetes.io/instance":  relName,
						"app.kubernetes.io/component": comp.Name,
					},
					Ports: svcPorts,
					Type:  corev1.ServiceTypeClusterIP,
				},
			}

			existingSvc, svcErr := cli.Kube.CoreV1().Services(ns).Get(ctx, deployName, metav1.GetOptions{})
			if svcErr == nil {
				existingSvc.Labels = compLabels
				existingSvc.Spec.Selector = svc.Spec.Selector
				existingSvc.Spec.Ports = svc.Spec.Ports
				_, _ = cli.Kube.CoreV1().Services(ns).Update(ctx, existingSvc, metav1.UpdateOptions{})
			} else if apierrors.IsNotFound(svcErr) {
				_, _ = cli.Kube.CoreV1().Services(ns).Create(ctx, svc, metav1.CreateOptions{})
			}
		}

		createdCount++
		global.Logger.Info("组件已创建",
			zap.String("component", comp.Name), zap.String("image", comp.Image),
			zap.String("deploy", deployName))
	}

	if createdCount == 0 {
		s.failInstall(ctx, install.ID, now, "所有组件创建失败", fmt.Errorf("0/%d components created", totalComponents), app.Name, ns)
		return
	}

	// ---- Step 3: 等待所有组件 Pod 就绪 ----
	allReady := true
	notReadyComponents := make([]string, 0)
	for _, comp := range components {
		deployName := fmt.Sprintf("%s-%s", relName, comp.Name)
		podReady := s.waitForDeploymentReady(ctx, cli, ns, deployName, 60*time.Second)
		if !podReady {
			allReady = false
			notReadyComponents = append(notReadyComponents, comp.Name)
		}
	}

	// 根据所有组件的就绪状态决定最终状态
	var status uint8
	var msg string
	if allReady {
		// 所有组件全部 Ready
		status = models.InstallStatusInstalled
		msg = fmt.Sprintf("应用 %s v%s 已部署 (%d/%d 组件), 所有 Pod Ready",
			app.DisplayName, app.Version, createdCount, totalComponents)
	} else if len(notReadyComponents) < totalComponents {
		// 部分组件 Ready，降级运行
		status = models.InstallStatusPartialReady
		msg = fmt.Sprintf("应用 %s v%s 已部署 (%d/%d 组件), 部分组件未就绪: %s",
			app.DisplayName, app.Version, createdCount, totalComponents,
			strings.Join(notReadyComponents, ", "))
	} else {
		// 所有组件都未 Ready
		status = models.InstallStatusPartialReady
		msg = fmt.Sprintf("应用 %s v%s 已部署 (%d/%d 组件), 所有组件启动中(可能需要拉取镜像)",
			app.DisplayName, app.Version, createdCount, totalComponents)
	}

	_ = s.dao.AppStoreInstallUpdate(ctx, install.ID, map[string]interface{}{
		"status":      status,
		"message":     msg,
		"modified_at": uint32(time.Now().Unix()),
	})

	global.Logger.Info("应用安装完成",
		zap.String("app", app.Name), zap.String("namespace", ns),
		zap.Int("created", createdCount), zap.Int("total", totalComponents),
		zap.Bool("allReady", allReady))
}

// failInstall 标记安装失败
func (s *Services) failInstall(ctx context.Context, installID uint32, now uint32, reason string, err error, appName, ns string) {
	global.Logger.Error("安装应用失败",
		zap.String("reason", reason), zap.String("app", appName),
		zap.String("namespace", ns), zap.Error(err))
	_ = s.dao.AppStoreInstallUpdate(ctx, installID, map[string]interface{}{
		"status":      models.InstallStatusFailed,
		"message":     fmt.Sprintf("%s: %s", reason, err.Error()),
		"modified_at": now,
	})
}

// ============================================================
// 多组件定义系统（参考 Rancher App Catalog 设计）
// 每个应用包含多个组件，每个组件独立 Deployment + Service
// ============================================================

// AppComponentSpec 应用组件定义
type AppComponentSpec struct {
	Name     string   // 组件名称（如 prometheus, grafana）
	Image    string   // 容器镜像
	Replicas int32    // 副本数
	Ports    []PortSpec // 端口列表
	Args     []string // 容器启动参数
	CPUReq   string   // CPU Request
	CPULim   string   // CPU Limit
	MemReq   string   // Memory Request
	MemLim   string   // Memory Limit
}

// PortSpec 端口定义
type PortSpec struct {
	Name string
	Port int32
}

// resolveAppComponents 根据应用名称返回组件列表
// 参考真实 Helm Chart 的组件拆分方式
func resolveAppComponents(appName string) []AppComponentSpec {
	registry := map[string][]AppComponentSpec{
		// ====== Prometheus Stack: 5 组件 ======
		"prometheus-stack": {
			{Name: "prometheus", Image: "prom/prometheus:v2.51.2", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 9090}},
				Args:  []string{"--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/prometheus", "--web.enable-lifecycle"},
				CPUReq: "100m", CPULim: "500m", MemReq: "256Mi", MemLim: "512Mi"},
			{Name: "grafana", Image: "grafana/grafana:10.4.2", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 3000}},
				CPUReq: "50m", CPULim: "200m", MemReq: "128Mi", MemLim: "256Mi"},
			{Name: "alertmanager", Image: "prom/alertmanager:v0.27.0", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 9093}},
				CPUReq: "25m", CPULim: "100m", MemReq: "64Mi", MemLim: "128Mi"},
			{Name: "node-exporter", Image: "prom/node-exporter:v1.7.0", Replicas: 1,
				Ports: []PortSpec{{Name: "metrics", Port: 9100}},
				CPUReq: "25m", CPULim: "100m", MemReq: "32Mi", MemLim: "64Mi"},
			{Name: "kube-state-metrics", Image: "registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.12.0", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 8080}, {Name: "telemetry", Port: 8081}},
				CPUReq: "25m", CPULim: "100m", MemReq: "64Mi", MemLim: "128Mi"},
		},
		// ====== Ingress NGINX: 2 组件 ======
		"ingress-nginx": {
			{Name: "controller", Image: "registry.k8s.io/ingress-nginx/controller:v1.10.1", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 80}, {Name: "https", Port: 443}},
				CPUReq: "100m", CPULim: "500m", MemReq: "128Mi", MemLim: "256Mi"},
			{Name: "default-backend", Image: "registry.k8s.io/defaultbackend-amd64:1.5", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 8080}},
				CPUReq: "10m", CPULim: "50m", MemReq: "16Mi", MemLim: "32Mi"},
		},
		// ====== ArgoCD: 3 组件 ======
		"argocd": {
			{Name: "server", Image: "quay.io/argoproj/argocd:v2.10.7", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 8080}, {Name: "https", Port: 8443}},
				Args:  []string{"argocd-server"},
				CPUReq: "50m", CPULim: "300m", MemReq: "128Mi", MemLim: "256Mi"},
			{Name: "repo-server", Image: "quay.io/argoproj/argocd:v2.10.7", Replicas: 1,
				Ports: []PortSpec{{Name: "server", Port: 8081}},
				Args:  []string{"argocd-repo-server"},
				CPUReq: "50m", CPULim: "200m", MemReq: "128Mi", MemLim: "256Mi"},
			{Name: "application-controller", Image: "quay.io/argoproj/argocd:v2.10.7", Replicas: 1,
				Ports: []PortSpec{{Name: "metrics", Port: 8082}},
				Args:  []string{"argocd-application-controller"},
				CPUReq: "50m", CPULim: "500m", MemReq: "256Mi", MemLim: "512Mi"},
		},
		// ====== EFK Stack: 3 组件 ======
		"efk-stack": {
			{Name: "elasticsearch", Image: "docker.elastic.co/elasticsearch/elasticsearch:8.13.2", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 9200}, {Name: "transport", Port: 9300}},
				CPUReq: "200m", CPULim: "1000m", MemReq: "512Mi", MemLim: "1Gi"},
			{Name: "fluentbit", Image: "fluent/fluent-bit:3.0", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 2020}},
				CPUReq: "25m", CPULim: "100m", MemReq: "64Mi", MemLim: "128Mi"},
			{Name: "kibana", Image: "docker.elastic.co/kibana/kibana:8.13.2", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 5601}},
				CPUReq: "100m", CPULim: "500m", MemReq: "256Mi", MemLim: "512Mi"},
		},
		// ====== Loki Stack: 3 组件 ======
		"loki-stack": {
			{Name: "loki", Image: "grafana/loki:2.9.6", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 3100}},
				CPUReq: "50m", CPULim: "200m", MemReq: "128Mi", MemLim: "256Mi"},
			{Name: "promtail", Image: "grafana/promtail:2.9.6", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 3101}},
				CPUReq: "25m", CPULim: "100m", MemReq: "64Mi", MemLim: "128Mi"},
			{Name: "grafana", Image: "grafana/grafana:10.4.2", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 3000}},
				CPUReq: "50m", CPULim: "200m", MemReq: "128Mi", MemLim: "256Mi"},
		},
		// ====== Cert Manager: 3 组件 ======
		"cert-manager": {
			{Name: "controller", Image: "quay.io/jetstack/cert-manager-controller:v1.14.5", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 9402}},
				CPUReq: "25m", CPULim: "200m", MemReq: "64Mi", MemLim: "128Mi"},
			{Name: "webhook", Image: "quay.io/jetstack/cert-manager-webhook:v1.14.5", Replicas: 1,
				Ports: []PortSpec{{Name: "https", Port: 10250}},
				CPUReq: "25m", CPULim: "100m", MemReq: "32Mi", MemLim: "64Mi"},
			{Name: "cainjector", Image: "quay.io/jetstack/cert-manager-cainjector:v1.14.5", Replicas: 1,
				Ports: []PortSpec{},
				CPUReq: "25m", CPULim: "100m", MemReq: "64Mi", MemLim: "128Mi"},
		},
		// ====== MetalLB: 2 组件 ======
		"metallb": {
			{Name: "controller", Image: "quay.io/metallb/controller:v0.14.5", Replicas: 1,
				Ports: []PortSpec{{Name: "metrics", Port: 7472}},
				CPUReq: "25m", CPULim: "100m", MemReq: "64Mi", MemLim: "128Mi"},
			{Name: "speaker", Image: "quay.io/metallb/speaker:v0.14.5", Replicas: 1,
				Ports: []PortSpec{{Name: "metrics", Port: 7473}},
				CPUReq: "25m", CPULim: "100m", MemReq: "32Mi", MemLim: "64Mi"},
		},
		// ====== Harbor: 4 组件 ======
		"harbor": {
			{Name: "core", Image: "goharbor/harbor-core:v2.10.2", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 8080}},
				CPUReq: "50m", CPULim: "300m", MemReq: "128Mi", MemLim: "256Mi"},
			{Name: "registry", Image: "goharbor/registry-photon:v2.10.2", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 5000}},
				CPUReq: "50m", CPULim: "200m", MemReq: "128Mi", MemLim: "256Mi"},
			{Name: "portal", Image: "goharbor/harbor-portal:v2.10.2", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 8080}},
				CPUReq: "25m", CPULim: "100m", MemReq: "64Mi", MemLim: "128Mi"},
			{Name: "jobservice", Image: "goharbor/harbor-jobservice:v2.10.2", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 8080}},
				CPUReq: "25m", CPULim: "200m", MemReq: "128Mi", MemLim: "256Mi"},
		},
		// ====== Kafka (Strimzi): 2 组件 ======
		"kafka": {
			{Name: "strimzi-operator", Image: "quay.io/strimzi/operator:0.40.0", Replicas: 1,
				Ports: []PortSpec{{Name: "http", Port: 8080}},
				CPUReq: "50m", CPULim: "200m", MemReq: "256Mi", MemLim: "512Mi"},
			{Name: "kafka-broker", Image: "quay.io/strimzi/kafka:0.40.0-kafka-3.7.0", Replicas: 1,
				Ports: []PortSpec{{Name: "tcp", Port: 9092}},
				CPUReq: "100m", CPULim: "500m", MemReq: "256Mi", MemLim: "512Mi"},
		},
	}

	if comps, ok := registry[appName]; ok {
		return comps
	}

	// 默认：单组件
	return []AppComponentSpec{
		{Name: "app", Image: "nginx:1.25-alpine", Replicas: 1,
			Ports:  []PortSpec{{Name: "http", Port: 80}},
			CPUReq: "25m", CPULim: "100m", MemReq: "32Mi", MemLim: "64Mi"},
	}
}

// dbComponentsToSpecs 将数据库组件记录转换为 AppComponentSpec 列表
func dbComponentsToSpecs(dbComps []*models.AppStoreComponent) []AppComponentSpec {
	specs := make([]AppComponentSpec, 0, len(dbComps))
	for _, c := range dbComps {
		spec := AppComponentSpec{
			Name:     c.Name,
			Image:    c.Image,
			Replicas: c.Replicas,
			CPUReq:   c.CPUReq,
			CPULim:   c.CPULim,
			MemReq:   c.MemReq,
			MemLim:   c.MemLim,
		}
		if spec.Replicas <= 0 {
			spec.Replicas = 1
		}
		if spec.CPUReq == "" {
			spec.CPUReq = "50m"
		}
		if spec.CPULim == "" {
			spec.CPULim = "200m"
		}
		if spec.MemReq == "" {
			spec.MemReq = "64Mi"
		}
		if spec.MemLim == "" {
			spec.MemLim = "256Mi"
		}
		// 解析端口 JSON: [{"name":"http","port":9090}]
		if c.Ports != "" {
			var ports []PortSpec
			if err := json.Unmarshal([]byte(c.Ports), &ports); err == nil {
				spec.Ports = ports
			}
		}
		// 解析启动参数 JSON: ["--config.file=..."]
		if c.Args != "" {
			var args []string
			if err := json.Unmarshal([]byte(c.Args), &args); err == nil {
				spec.Args = args
			}
		}
		specs = append(specs, spec)
	}
	return specs
}

// waitForDeploymentReady 等待 Deployment 的 Pod 就绪
func (s *Services) waitForDeploymentReady(ctx context.Context, cli *K8sClients, ns, deployName string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		deploy, err := cli.Kube.AppsV1().Deployments(ns).Get(ctx, deployName, metav1.GetOptions{})
		if err == nil && deploy.Status.ReadyReplicas > 0 {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(3 * time.Second):
		}
	}
	return false
}

// AppStoreInstallStatus 实时查询安装的 K8s 资源状态（Pod/Deployment/Service）
func (s *Services) AppStoreInstallStatus(ctx context.Context, factory *ClusterClientFactory, installID uint32) (*models.AppInstallStatusResponse, error) {
	install, err := s.dao.AppStoreInstallGetByID(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("安装记录不存在: %w", err)
	}

	resp := &models.AppInstallStatusResponse{
		InstallID:   install.ID,
		AppName:     install.AppName,
		ClusterName: install.ClusterName,
		Namespace:   install.Namespace,
		ReleaseName: install.ReleaseName,
		Version:     install.Version,
		DbStatus:    install.Status,
		DbMessage:   install.Message,
	}

	// 尝试连接集群获取实时状态
	cli, err := factory.Get(ctx, install.ClusterID)
	if err != nil {
		resp.ClusterReachable = false
		resp.ClusterError = fmt.Sprintf("集群连接失败: %s", err.Error())
		return resp, nil
	}
	resp.ClusterReachable = true

	// 使用标签选择器查询该 release 的所有资源
	labelSel := fmt.Sprintf("app.kubernetes.io/instance=%s", install.ReleaseName)

	// 用于后续 Pod 详情查询的变量（在 Deployment 状态检查中可能已查询）
	var releasePodList *corev1.PodList

	// 查询该 release 的所有 Deployment（多组件）
	deployList, err := cli.Kube.AppsV1().Deployments(install.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSel,
	})
	if err != nil {
		resp.DeploymentStatus = "Error"
		resp.DeploymentMessage = err.Error()
	} else if len(deployList.Items) == 0 {
		resp.DeploymentStatus = "NotFound"
		resp.DeploymentMessage = "Deployment 未创建或已被删除"
	} else {
		// 汇总所有组件的 Deployment 状态
		totalDesired := int32(0)
		totalReady := int32(0)
		totalUpdated := int32(0)
		totalAvailable := int32(0)
		allDeploymentsReady := true
		anyFailed := false

		for _, deploy := range deployList.Items {
			desired := int32(1)
			if deploy.Spec.Replicas != nil {
				desired = *deploy.Spec.Replicas
			}
			totalDesired += desired
			totalReady += deploy.Status.ReadyReplicas
			totalUpdated += deploy.Status.UpdatedReplicas
			totalAvailable += deploy.Status.AvailableReplicas

			// 只有当 ReadyReplicas == 期望副本数 且 AvailableReplicas == 期望副本数 且 期望 > 0 时才算该 Deployment 就绪
			if desired == 0 || deploy.Status.ReadyReplicas != desired || deploy.Status.AvailableReplicas != desired {
				allDeploymentsReady = false
			}

			for _, cond := range deploy.Status.Conditions {
				if cond.Type == appsv1.DeploymentProgressing && cond.Status == corev1.ConditionFalse {
					anyFailed = true
				}
				if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionFalse {
					anyFailed = true
				}
			}
		}

		// 额外检查: 查询该 release 所有 Pod，确认每个 Pod 的每个容器都 Ready
		allPodsReady := true
		var podErr error
		releasePodList, podErr = cli.Kube.CoreV1().Pods(install.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSel,
		})
		if podErr != nil || releasePodList == nil || len(releasePodList.Items) == 0 {
			allPodsReady = false
		} else {
			for _, pod := range releasePodList.Items {
				// Pod Phase 必须是 Running
				if pod.Status.Phase != corev1.PodRunning {
					allPodsReady = false
					break
				}
				// 所有容器必须 Ready
				if len(pod.Status.ContainerStatuses) == 0 {
					allPodsReady = false
					break
				}
				for _, cs := range pod.Status.ContainerStatuses {
					if !cs.Ready {
						allPodsReady = false
						break
					}
				}
				if !allPodsReady {
					break
				}
			}
		}

		resp.DesiredReplicas = int(totalDesired)
		resp.ReadyReplicas = int(totalReady)
		resp.UpdatedReplicas = int(totalUpdated)
		resp.AvailableReplicas = int(totalAvailable)

		// 只有 Deployment 全部就绪 + 所有 Pod 所有容器都 Ready 才显示 Running
		if allDeploymentsReady && allPodsReady && totalDesired > 0 {
			resp.DeploymentStatus = "Running"
			resp.DeploymentMessage = fmt.Sprintf("所有组件就绪 (%d 个 Deployment, %d/%d 副本, 所有 Pod Ready)", len(deployList.Items), totalReady, totalDesired)
		} else if anyFailed {
			resp.DeploymentStatus = "Failed"
			resp.DeploymentMessage = fmt.Sprintf("部分组件失败 (%d/%d 副本就绪)", totalReady, totalDesired)
		} else if totalReady > 0 {
			resp.DeploymentStatus = "PartialReady"
			resp.DeploymentMessage = fmt.Sprintf("部分组件就绪 (%d/%d 副本就绪, 等待所有 Pod Ready)", totalReady, totalDesired)
		} else {
			resp.DeploymentStatus = "Pending"
			resp.DeploymentMessage = fmt.Sprintf("组件启动中... (0/%d 副本就绪)", totalDesired)
		}
	}

	// 查询该 release 的所有 Service
	svcList, err := cli.Kube.CoreV1().Services(install.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSel,
	})
	if err != nil || len(svcList.Items) == 0 {
		resp.ServiceStatus = "NotFound"
	} else {
		resp.ServiceStatus = "Active"
		resp.ServiceType = string(svcList.Items[0].Spec.Type)
		resp.ServicePorts = make([]string, 0)
		for _, svcObj := range svcList.Items {
			for _, p := range svcObj.Spec.Ports {
				resp.ServicePorts = append(resp.ServicePorts, fmt.Sprintf("%s/%s:%d/%s", svcObj.Name, p.Name, p.Port, p.Protocol))
			}
		}
		resp.ClusterIP = svcList.Items[0].Spec.ClusterIP
	}

	// 查询该 release 的所有 Pods（复用已查询的 podList，若未查询则重新查询）
	if releasePodList == nil {
		releasePodList, _ = cli.Kube.CoreV1().Pods(install.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSel,
		})
	}
	if releasePodList != nil {
		for _, pod := range releasePodList.Items {
			// 判断 Pod 真实状态：Phase=Running 且所有容器 Ready 才算 Running
			podPhase := string(pod.Status.Phase)
			allContainersReady := true
			if pod.Status.Phase == corev1.PodRunning {
				for _, cs := range pod.Status.ContainerStatuses {
					if !cs.Ready {
						allContainersReady = false
						break
					}
				}
				if len(pod.Status.ContainerStatuses) == 0 {
					allContainersReady = false
				}
				if !allContainersReady {
					podPhase = "NotReady" // Pod Phase 虽然是 Running，但有容器未就绪
				}
			}

			podInfo := models.PodStatusInfo{
				Name:      pod.Name,
				Phase:     podPhase,
				NodeName:  pod.Spec.NodeName,
				PodIP:     pod.Status.PodIP,
				StartTime: "",
				Restarts:  0,
			}
			if pod.Status.StartTime != nil {
				podInfo.StartTime = pod.Status.StartTime.Format(time.RFC3339)
			}
			// 容器状态
			for _, cs := range pod.Status.ContainerStatuses {
				podInfo.Restarts += int(cs.RestartCount)
				cInfo := models.ContainerStatusInfo{
					Name:         cs.Name,
					Image:        cs.Image,
					Ready:        cs.Ready,
					RestartCount: int(cs.RestartCount),
				}
				if cs.State.Running != nil {
					cInfo.State = "Running"
					cInfo.StartedAt = cs.State.Running.StartedAt.Format(time.RFC3339)
				} else if cs.State.Waiting != nil {
					cInfo.State = "Waiting"
					cInfo.Reason = cs.State.Waiting.Reason
					cInfo.Message = cs.State.Waiting.Message
				} else if cs.State.Terminated != nil {
					cInfo.State = "Terminated"
					cInfo.Reason = cs.State.Terminated.Reason
					cInfo.Message = cs.State.Terminated.Message
				}
				podInfo.Containers = append(podInfo.Containers, cInfo)
			}
			resp.Pods = append(resp.Pods, podInfo)
		}
	}

	// 查询 ConfigMaps（带 managed-by 标签的）
	cmLabelSel := "app.kubernetes.io/managed-by=k8s-operation-appstore,app.kubernetes.io/instance=" + install.ReleaseName
	cmList, err := cli.Kube.CoreV1().ConfigMaps(install.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: cmLabelSel,
	})
	if err == nil && cmList != nil {
		for _, cm := range cmList.Items {
			cmInfo := models.ConfigMapStatusInfo{
				Name:      cm.Name,
				Namespace: cm.Namespace,
				Data:      cm.Data,
			}
			if !cm.CreationTimestamp.IsZero() {
				cmInfo.CreatedAt = cm.CreationTimestamp.Format(time.RFC3339)
			}
			resp.ConfigMaps = append(resp.ConfigMaps, cmInfo)
		}
	}

	// ====== 命名空间级别资源扫描 ======
	ns := install.Namespace
	overview := &models.NamespaceOverview{}

	// 扫描命名空间内所有 Deployment
	allDeployList, err := cli.Kube.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	if err == nil && allDeployList != nil {
		overview.TotalDeployments = len(allDeployList.Items)
		for _, d := range allDeployList.Items {
			dInfo := models.DeploymentStatusInfo{
				Name:              d.Name,
				Replicas:          int(d.Status.Replicas),
				ReadyReplicas:     int(d.Status.ReadyReplicas),
				UpdatedReplicas:   int(d.Status.UpdatedReplicas),
				AvailableReplicas: int(d.Status.AvailableReplicas),
			}
			if d.Spec.Replicas != nil && d.Status.ReadyReplicas == *d.Spec.Replicas && *d.Spec.Replicas > 0 && d.Status.AvailableReplicas == *d.Spec.Replicas {
				dInfo.Status = "Running"
			} else if d.Status.ReadyReplicas > 0 {
				dInfo.Status = "PartialReady"
			} else if d.Status.Replicas == 0 {
				dInfo.Status = "Scaled0"
			} else {
				dInfo.Status = "Pending"
			}
			if len(d.Spec.Template.Spec.Containers) > 0 {
				dInfo.Image = d.Spec.Template.Spec.Containers[0].Image
			}
			if !d.CreationTimestamp.IsZero() {
				dInfo.CreatedAt = d.CreationTimestamp.Format(time.RFC3339)
			}
			resp.AllDeployments = append(resp.AllDeployments, dInfo)
		}
	}

	// 扫描命名空间内所有 Service
	allSvcList, err := cli.Kube.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	if err == nil && allSvcList != nil {
		overview.TotalServices = len(allSvcList.Items)
		for _, svc := range allSvcList.Items {
			sInfo := models.ServiceStatusInfo{
				Name:      svc.Name,
				Type:      string(svc.Spec.Type),
				ClusterIP: svc.Spec.ClusterIP,
			}
			for _, p := range svc.Spec.Ports {
				sInfo.Ports = append(sInfo.Ports, fmt.Sprintf("%s:%d/%s", p.Name, p.Port, p.Protocol))
			}
			if !svc.CreationTimestamp.IsZero() {
				sInfo.CreatedAt = svc.CreationTimestamp.Format(time.RFC3339)
			}
			resp.AllServices = append(resp.AllServices, sInfo)
		}
	}

	// 扫描命名空间内所有 Pod（统计概览 — 所有容器 Ready 才算 Running）
	allPodList, err := cli.Kube.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err == nil && allPodList != nil {
		overview.TotalPods = len(allPodList.Items)
		for _, p := range allPodList.Items {
			switch p.Status.Phase {
			case corev1.PodRunning:
				// Phase=Running 还需要检查所有容器是否都 Ready
				allReady := true
				for _, cs := range p.Status.ContainerStatuses {
					if !cs.Ready {
						allReady = false
						break
					}
				}
				if len(p.Status.ContainerStatuses) == 0 {
					allReady = false
				}
				if allReady {
					overview.RunningPods++
				} else {
					overview.PendingPods++ // 容器未全部就绪，归入 Pending
				}
			case corev1.PodPending:
				overview.PendingPods++
			case corev1.PodFailed:
				overview.FailedPods++
			}
		}
	}

	// 扫描命名空间内 ConfigMap 数量
	allCmList, err := cli.Kube.CoreV1().ConfigMaps(ns).List(ctx, metav1.ListOptions{})
	if err == nil && allCmList != nil {
		overview.TotalConfigMaps = len(allCmList.Items)
	}

	resp.NamespaceOverview = overview

	// 查询命名空间最近 Events（最多50条，按时间倒序）
	eventList, err := cli.Kube.CoreV1().Events(ns).List(ctx, metav1.ListOptions{
		Limit: 50,
	})
	if err == nil && eventList != nil {
		for _, ev := range eventList.Items {
			evInfo := models.K8sEventInfo{
				Type:    ev.Type,
				Reason:  ev.Reason,
				Message: ev.Message,
				Object:  fmt.Sprintf("%s/%s", ev.InvolvedObject.Kind, ev.InvolvedObject.Name),
				Count:   ev.Count,
			}
			if !ev.FirstTimestamp.IsZero() {
				evInfo.FirstTime = ev.FirstTimestamp.Format(time.RFC3339)
			}
			if !ev.LastTimestamp.IsZero() {
				evInfo.LastTime = ev.LastTimestamp.Format(time.RFC3339)
			}
			resp.Events = append(resp.Events, evInfo)
		}
	}

	return resp, nil
}

// getAppNameFromInstall 获取应用的 name 字段
func (s *Services) getAppNameFromInstall(ctx context.Context, appID uint32) string {
	app, err := s.dao.AppStoreGetByID(ctx, appID)
	if err != nil {
		return ""
	}
	return app.Name
}

// ensureNamespace 确保命名空间存在（幂等）
func (s *Services) ensureNamespace(ctx context.Context, cli *K8sClients, ns string) error {
	_, err := cli.Kube.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
	if err == nil {
		return nil // 已存在
	}

	if !apierrors.IsNotFound(err) {
		return err
	}

	// 创建命名空间
	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "k8s-operation-appstore",
			},
		},
	}
	_, err = cli.Kube.CoreV1().Namespaces().Create(ctx, nsObj, metav1.CreateOptions{})
	return err
}

// createInstallMarkerConfigMap 创建安装标记 ConfigMap
func (s *Services) createInstallMarkerConfigMap(ctx context.Context, cli *K8sClients, namespace, name string, labels, data map[string]string) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: data,
	}

	// 先尝试获取，存在则更新，不存在则创建
	existing, err := cli.Kube.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		// 已存在，更新
		existing.Labels = labels
		existing.Data = data
		_, err = cli.Kube.CoreV1().ConfigMaps(namespace).Update(ctx, existing, metav1.UpdateOptions{})
		return err
	}

	if !apierrors.IsNotFound(err) {
		return err
	}

	// 不存在，创建
	_, err = cli.Kube.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	return err
}

// AppStoreUninstall 卸载应用（删除所有组件的 Deployment + Service + ConfigMap）
func (s *Services) AppStoreUninstall(ctx context.Context, factory *ClusterClientFactory, installID uint32) error {
	install, err := s.dao.AppStoreInstallGetByID(ctx, installID)
	if err != nil {
		return fmt.Errorf("安装记录不存在: %w", err)
	}

	if install.Status != models.InstallStatusInstalled && install.Status != models.InstallStatusFailed {
		return fmt.Errorf("当前状态不允许卸载(status=%d)", install.Status)
	}

	// 更新状态为卸载中
	now := uint32(time.Now().Unix())
	_ = s.dao.AppStoreInstallUpdate(ctx, installID, map[string]interface{}{
		"status":      models.InstallStatusUninstalling,
		"message":     "卸载中...",
		"modified_at": now,
	})

	// 尝试删除集群中的资源
	cli, err := factory.Get(ctx, install.ClusterID)
	if err != nil {
		global.Logger.Warn("卸载时连接集群失败", zap.Error(err))
	} else {
		ns := install.Namespace
		relName := install.ReleaseName

		// 查找应用的组件列表，并通过标签选择器查找该 release 的所有资源
		labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", relName)

		// 删除所有关联的 Deployment
		deployList, listErr := cli.Kube.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
		if listErr == nil {
			for _, d := range deployList.Items {
				err := cli.Kube.AppsV1().Deployments(ns).Delete(ctx, d.Name, metav1.DeleteOptions{})
				if err != nil && !apierrors.IsNotFound(err) {
					global.Logger.Warn("删除 Deployment 失败", zap.String("name", d.Name), zap.Error(err))
				} else {
					global.Logger.Info("已删除 Deployment", zap.String("name", d.Name))
				}
			}
		}

		// 删除所有关联的 Service
		svcList, listErr := cli.Kube.CoreV1().Services(ns).List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
		if listErr == nil {
			for _, svc := range svcList.Items {
				err := cli.Kube.CoreV1().Services(ns).Delete(ctx, svc.Name, metav1.DeleteOptions{})
				if err != nil && !apierrors.IsNotFound(err) {
					global.Logger.Warn("删除 Service 失败", zap.String("name", svc.Name), zap.Error(err))
				} else {
					global.Logger.Info("已删除 Service", zap.String("name", svc.Name))
				}
			}
		}

		// 删除 ConfigMap 标记
		cmName := fmt.Sprintf("appstore-%s", relName)
		err = cli.Kube.CoreV1().ConfigMaps(ns).Delete(ctx, cmName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			global.Logger.Warn("删除 ConfigMap 失败", zap.String("name", cmName), zap.Error(err))
		}
	}

	// 更新为已卸载
	_ = s.dao.AppStoreInstallUpdate(ctx, installID, map[string]interface{}{
		"status":      models.InstallStatusUninstalled,
		"message":     "已卸载, 所有组件资源已清理",
		"modified_at": uint32(time.Now().Unix()),
	})

	return nil
}

// AppStoreInstallList 获取安装记录列表
func (s *Services) AppStoreInstallList(ctx context.Context, req *models.AppStoreInstallListRequest) ([]*models.AppStoreInstall, int64, error) {
	return s.dao.AppStoreInstallList(ctx, req)
}

// AppStoreInstallDetail 获取安装记录详情
func (s *Services) AppStoreInstallDetail(ctx context.Context, id uint32) (*models.AppStoreInstall, error) {
	return s.dao.AppStoreInstallGetByID(ctx, id)
}

// AppStoreInstallUpdate 编辑安装（更新 Deployment 副本数、镜像、资源限制等）
func (s *Services) AppStoreInstallUpdate(ctx context.Context, factory *ClusterClientFactory, installID uint32, req *models.AppStoreInstallUpdateRequest) error {
	install, err := s.dao.AppStoreInstallGetByID(ctx, installID)
	if err != nil {
		return fmt.Errorf("安装记录不存在: %w", err)
	}

	if install.Status != models.InstallStatusInstalled && install.Status != models.InstallStatusFailed {
		return fmt.Errorf("当前状态不允许编辑(status=%d)", install.Status)
	}

	cli, err := factory.Get(ctx, install.ClusterID)
	if err != nil {
		return fmt.Errorf("连接集群失败: %w", err)
	}

	deployName := install.ReleaseName
	deploy, err := cli.Kube.AppsV1().Deployments(install.Namespace).Get(ctx, deployName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("获取 Deployment 失败: %w", err)
	}

	updated := false

	// 更新副本数
	if req.Replicas != nil && *req.Replicas >= 0 {
		deploy.Spec.Replicas = req.Replicas
		updated = true
	}

	// 更新镜像
	if req.Image != "" && len(deploy.Spec.Template.Spec.Containers) > 0 {
		deploy.Spec.Template.Spec.Containers[0].Image = req.Image
		updated = true
	}

	// 更新资源限制
	if len(deploy.Spec.Template.Spec.Containers) > 0 {
		container := &deploy.Spec.Template.Spec.Containers[0]
		if req.CPUReq != "" {
			if container.Resources.Requests == nil {
				container.Resources.Requests = corev1.ResourceList{}
			}
			container.Resources.Requests[corev1.ResourceCPU] = resource.MustParse(req.CPUReq)
			updated = true
		}
		if req.MemReq != "" {
			if container.Resources.Requests == nil {
				container.Resources.Requests = corev1.ResourceList{}
			}
			container.Resources.Requests[corev1.ResourceMemory] = resource.MustParse(req.MemReq)
			updated = true
		}
		if req.CPULim != "" {
			if container.Resources.Limits == nil {
				container.Resources.Limits = corev1.ResourceList{}
			}
			container.Resources.Limits[corev1.ResourceCPU] = resource.MustParse(req.CPULim)
			updated = true
		}
		if req.MemLim != "" {
			if container.Resources.Limits == nil {
				container.Resources.Limits = corev1.ResourceList{}
			}
			container.Resources.Limits[corev1.ResourceMemory] = resource.MustParse(req.MemLim)
			updated = true
		}
	}

	if !updated {
		return fmt.Errorf("没有可更新的字段")
	}

	_, err = cli.Kube.AppsV1().Deployments(install.Namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("更新 Deployment 失败: %w", err)
	}

	global.Logger.Info("Deployment 更新成功",
		zap.String("name", deployName),
		zap.String("namespace", install.Namespace),
	)

	return nil
}

// AppStoreSeed 初始化种子数据（首次使用自动填充）
func (s *Services) AppStoreSeed(ctx context.Context) error {
	// 检查是否已有数据
	req := &models.AppStoreListRequest{Page: 1, PageSize: 1}
	_, total, err := s.dao.AppStoreList(ctx, req)
	if err != nil {
		return err
	}
	if total > 0 {
		// 应用已有数据，但仍需检查组件种子数据
		if err := s.seedAppComponents(ctx); err != nil {
			global.Logger.Warn("组件种子数据初始化失败", zap.Error(err))
		}
		return nil
	}

	seeds := []models.AppStoreCreateRequest{
		{
			Name: "ingress-nginx", DisplayName: "Ingress NGINX",
			Category: "网络", Version: "4.10.1",
			Icon: "ingress", Provider: "official",
			Description: "Kubernetes 官方 Ingress Controller，基于 NGINX 实现 HTTP/HTTPS 反向代理与负载均衡",
			Tags: "ingress,nginx,负载均衡,反向代理", MinK8s: "v1.22+",
			Namespace: "ingress-nginx", Featured: 1, SortOrder: 100,
			DocURL: "https://kubernetes.github.io/ingress-nginx/",
		},
		{
			Name: "prometheus-stack", DisplayName: "Prometheus Stack",
			Category: "监控", Version: "58.2.2",
			Icon: "prometheus", Provider: "official",
			Description: "一站式 Kubernetes 监控套件：Prometheus + Grafana + Alertmanager + Node Exporter",
			Tags: "prometheus,grafana,alertmanager,监控,告警", MinK8s: "v1.24+",
			Namespace: "monitoring", Featured: 1, SortOrder: 95,
			DocURL: "https://github.com/prometheus-community/helm-charts",
		},
		{
			Name: "argocd", DisplayName: "Argo CD",
			Category: "GitOps", Version: "6.7.3",
			Icon: "argocd", Provider: "official",
			Description: "声明式 GitOps 持续交付工具，自动同步 Git 仓库到 Kubernetes 集群",
			Tags: "gitops,cd,持续交付,同步", MinK8s: "v1.25+",
			Namespace: "argocd", Featured: 1, SortOrder: 90,
			DocURL: "https://argo-cd.readthedocs.io/",
		},
		{
			Name: "efk-stack", DisplayName: "EFK Stack",
			Category: "日志", Version: "0.3.0",
			Icon: "elasticsearch", Provider: "community",
			Description: "Elasticsearch + Fluentbit + Kibana 日志收集与分析平台",
			Tags: "elasticsearch,fluentbit,kibana,日志,EFK", MinK8s: "v1.22+",
			Namespace: "logging", Featured: 1, SortOrder: 85,
			DocURL: "https://www.elastic.co/guide/",
		},
		{
			Name: "cert-manager", DisplayName: "Cert Manager",
			Category: "安全", Version: "1.14.5",
			Icon: "cert-manager", Provider: "official",
			Description: "Kubernetes 证书自动管理器，支持 Let's Encrypt 自动签发与续期 TLS 证书",
			Tags: "证书,TLS,HTTPS,安全,Let's Encrypt", MinK8s: "v1.22+",
			Namespace: "cert-manager", Featured: 0, SortOrder: 80,
			DocURL: "https://cert-manager.io/docs/",
		},
		{
			Name: "metallb", DisplayName: "MetalLB",
			Category: "网络", Version: "0.14.5",
			Icon: "metallb", Provider: "community",
			Description: "裸金属 Kubernetes 集群的负载均衡器实现，支持 L2 和 BGP 模式",
			Tags: "负载均衡,网络,裸金属,MetalLB", MinK8s: "v1.22+",
			Namespace: "metallb-system", Featured: 0, SortOrder: 75,
			DocURL: "https://metallb.universe.tf/",
		},
		{
			Name: "redis-cluster", DisplayName: "Redis Cluster",
			Category: "数据库", Version: "19.1.5",
			Icon: "redis", Provider: "official",
			Description: "高可用 Redis 集群，支持主从复制、哨兵模式和 Cluster 模式",
			Tags: "redis,缓存,数据库,KV", MinK8s: "v1.22+",
			Namespace: "redis", Featured: 0, SortOrder: 70,
			DocURL: "https://redis.io/docs/",
		},
		{
			Name: "mysql-operator", DisplayName: "MySQL Operator",
			Category: "数据库", Version: "8.3.0",
			Icon: "mysql", Provider: "official",
			Description: "Oracle 官方 MySQL Operator，自动管理 MySQL InnoDB Cluster 生命周期",
			Tags: "mysql,数据库,operator", MinK8s: "v1.24+",
			Namespace: "mysql-operator", Featured: 0, SortOrder: 65,
			DocURL: "https://dev.mysql.com/doc/mysql-operator/en/",
		},
		{
			Name: "kafka", DisplayName: "Apache Kafka (Strimzi)",
			Category: "消息队列", Version: "0.40.0",
			Icon: "kafka", Provider: "community",
			Description: "基于 Strimzi Operator 部署 Apache Kafka 消息队列集群",
			Tags: "kafka,消息队列,事件流,strimzi", MinK8s: "v1.23+",
			Namespace: "kafka", Featured: 0, SortOrder: 60,
			DocURL: "https://strimzi.io/documentation/",
		},
		{
			Name: "harbor", DisplayName: "Harbor",
			Category: "安全", Version: "1.14.2",
			Icon: "harbor", Provider: "official",
			Description: "企业级容器镜像仓库，支持镜像扫描、签名、复制策略",
			Tags: "镜像仓库,harbor,安全,registry", MinK8s: "v1.22+",
			Namespace: "harbor", Featured: 0, SortOrder: 55,
			DocURL: "https://goharbor.io/docs/",
		},
		{
			Name: "loki-stack", DisplayName: "Loki Stack",
			Category: "日志", Version: "2.10.0",
			Icon: "loki", Provider: "official",
			Description: "轻量级日志聚合系统 Loki + Promtail + Grafana，成本远低于 EFK",
			Tags: "loki,日志,grafana,promtail", MinK8s: "v1.22+",
			Namespace: "loki", Featured: 0, SortOrder: 50,
			DocURL: "https://grafana.com/docs/loki/latest/",
		},
		{
			Name: "longhorn", DisplayName: "Longhorn",
			Category: "存储", Version: "1.6.1",
			Icon: "longhorn", Provider: "official",
			Description: "云原生分布式块存储系统，Rancher 出品，支持快照与灾备",
			Tags: "存储,分布式,longhorn,PV", MinK8s: "v1.21+",
			Namespace: "longhorn-system", Featured: 0, SortOrder: 45,
			DocURL: "https://longhorn.io/docs/",
		},
	}

	for _, seed := range seeds {
		if _, err := s.AppStoreCreate(ctx, &seed); err != nil {
			global.Logger.Warn("种子数据写入失败", zap.String("name", seed.Name), zap.Error(err))
		}
	}

	global.Logger.Info("应用商城种子数据初始化完成", zap.Int("count", len(seeds)))

	// 初始化组件种子数据（将硬编码组件写入数据库）
	if err := s.seedAppComponents(ctx); err != nil {
		global.Logger.Warn("组件种子数据初始化失败", zap.Error(err))
	}

	return nil
}

// seedAppComponents 将硬编码的组件定义写入数据库
func (s *Services) seedAppComponents(ctx context.Context) error {
	registry := resolveAllBuiltinComponents()
	seeded := 0

	for appName, comps := range registry {
		// 查找应用
		app, err := s.dao.AppStoreGetByName(ctx, appName)
		if err != nil {
			continue // 应用不存在，跳过
		}

		// 检查该应用是否已有组件
		count, _ := s.dao.AppStoreComponentCountByAppID(ctx, app.ID)
		if count > 0 {
			continue // 已有组件数据，跳过
		}

		// 写入组件
		now := uint32(time.Now().Unix())
		for i, comp := range comps {
			portsJSON, _ := json.Marshal(comp.Ports)
			argsJSON, _ := json.Marshal(comp.Args)

			dbComp := &models.AppStoreComponent{
				AppID:     app.ID,
				Name:      comp.Name,
				Image:     comp.Image,
				Replicas:  comp.Replicas,
				Ports:     string(portsJSON),
				Args:      string(argsJSON),
				CPUReq:    comp.CPUReq,
				CPULim:    comp.CPULim,
				MemReq:    comp.MemReq,
				MemLim:    comp.MemLim,
				SortOrder: 100 - i*10, // 主组件排序最高
			}
			dbComp.CreatedAt = now
			dbComp.ModifiedAt = now

			if err := s.dao.AppStoreComponentCreate(ctx, dbComp); err != nil {
				global.Logger.Warn("组件种子写入失败",
					zap.String("app", appName), zap.String("comp", comp.Name), zap.Error(err))
			}
		}
		seeded++
	}

	if seeded > 0 {
		global.Logger.Info("组件种子数据初始化完成", zap.Int("apps", seeded))
	}
	return nil
}

// resolveAllBuiltinComponents 返回所有内置应用的组件定义（用于种子数据）
func resolveAllBuiltinComponents() map[string][]AppComponentSpec {
	registry := map[string][]AppComponentSpec{}
	appNames := []string{
		"prometheus-stack", "ingress-nginx", "argocd", "efk-stack",
		"loki-stack", "cert-manager", "metallb", "harbor", "kafka",
	}
	for _, name := range appNames {
		comps := resolveAppComponents(name)
		if len(comps) > 0 && comps[0].Name != "app" { // 排除默认单组件
			registry[name] = comps
		}
	}
	return registry
}

// ============================================================
// 组件管理 Service
// ============================================================

// AppStoreComponentList 获取应用的组件列表
func (s *Services) AppStoreComponentList(ctx context.Context, appID uint32) ([]*models.AppStoreComponent, error) {
	return s.dao.AppStoreComponentListByAppID(ctx, appID)
}

// AppStoreComponentCreate 创建组件
func (s *Services) AppStoreComponentCreate(ctx context.Context, req *models.AppStoreComponentRequest) (*models.AppStoreComponent, error) {
	// 校验应用存在
	_, err := s.dao.AppStoreGetByID(ctx, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("应用不存在: %w", err)
	}

	now := uint32(time.Now().Unix())
	comp := &models.AppStoreComponent{
		AppID:     req.AppID,
		Name:      req.Name,
		Image:     req.Image,
		Replicas:  req.Replicas,
		Ports:     req.Ports,
		Args:      req.Args,
		CPUReq:    req.CPUReq,
		CPULim:    req.CPULim,
		MemReq:    req.MemReq,
		MemLim:    req.MemLim,
		SortOrder: req.SortOrder,
	}
	if comp.Replicas <= 0 {
		comp.Replicas = 1
	}
	if comp.CPUReq == "" {
		comp.CPUReq = "50m"
	}
	if comp.CPULim == "" {
		comp.CPULim = "200m"
	}
	if comp.MemReq == "" {
		comp.MemReq = "64Mi"
	}
	if comp.MemLim == "" {
		comp.MemLim = "256Mi"
	}
	comp.CreatedAt = now
	comp.ModifiedAt = now

	if err := s.dao.AppStoreComponentCreate(ctx, comp); err != nil {
		return nil, err
	}
	return comp, nil
}

// AppStoreComponentUpdate 更新组件
func (s *Services) AppStoreComponentUpdate(ctx context.Context, req *models.AppStoreComponentRequest) error {
	existing, err := s.dao.AppStoreComponentGetByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("组件不存在: %w", err)
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Image != "" {
		existing.Image = req.Image
	}
	if req.Replicas > 0 {
		existing.Replicas = req.Replicas
	}
	existing.Ports = req.Ports
	existing.Args = req.Args
	if req.CPUReq != "" {
		existing.CPUReq = req.CPUReq
	}
	if req.CPULim != "" {
		existing.CPULim = req.CPULim
	}
	if req.MemReq != "" {
		existing.MemReq = req.MemReq
	}
	if req.MemLim != "" {
		existing.MemLim = req.MemLim
	}
	existing.SortOrder = req.SortOrder
	existing.ModifiedAt = uint32(time.Now().Unix())

	return s.dao.AppStoreComponentUpdate(ctx, existing)
}

// AppStoreComponentDelete 删除组件
func (s *Services) AppStoreComponentDelete(ctx context.Context, id uint32) error {
	_, err := s.dao.AppStoreComponentGetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("组件不存在: %w", err)
	}
	return s.dao.AppStoreComponentDelete(ctx, id)
}

// AppStoreComponentBatchDelete 批量删除组件
func (s *Services) AppStoreComponentBatchDelete(ctx context.Context, ids []uint32) error {
	if len(ids) == 0 {
		return fmt.Errorf("请选择要删除的组件")
	}
	return s.dao.AppStoreComponentBatchDelete(ctx, ids)
}

// AppStoreComponentSort 批量更新组件排序
func (s *Services) AppStoreComponentSort(ctx context.Context, req *models.AppStoreComponentSortRequest) error {
	for _, item := range req.Items {
		if err := s.dao.AppStoreComponentUpdateSort(ctx, item.ID, item.SortOrder); err != nil {
			return fmt.Errorf("更新排序失败(id=%d): %w", item.ID, err)
		}
	}
	return nil
}
