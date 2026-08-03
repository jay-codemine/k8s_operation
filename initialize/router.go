// Package initialize 用于系统初始化相关逻辑（配置、路由、Swagger 等）
package initialize

import (
	"github.com/gin-gonic/gin"

	"k8soperation/global"
	"k8soperation/middlewares"
	"k8soperation/pkg/license"

	_ "k8soperation/docs"

	// Swagger
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// 业务路由
	authrouter "k8soperation/internal/app/routers/auth"
	helloworldrouter "k8soperation/internal/app/routers/helloword"
	userrouter "k8soperation/internal/app/routers/user"

	// k8s 资源路由（目标集群资源操作，需要 ClusterMiddleware）
	"k8soperation/internal/app/routers/kube_configmap"
	"k8soperation/internal/app/routers/kube_crd"
	"k8soperation/internal/app/routers/kube_cronjob"
	"k8soperation/internal/app/routers/kube_daemonset"
	"k8soperation/internal/app/routers/kube_deployment"
	"k8soperation/internal/app/routers/kube_ingress"
	"k8soperation/internal/app/routers/kube_job"
	"k8soperation/internal/app/routers/kube_namespace"
	"k8soperation/internal/app/routers/kube_node"
	"k8soperation/internal/app/routers/kube_pod"
	"k8soperation/internal/app/routers/kube_pv"
	"k8soperation/internal/app/routers/kube_pvc"
	"k8soperation/internal/app/routers/kube_secret"
	"k8soperation/internal/app/routers/kube_service"
	"k8soperation/internal/app/routers/kube_statefulset"
	"k8soperation/internal/app/routers/kube_storageclass"

	// HPA/VPA 弹性扩缩容路由
	"k8soperation/internal/app/routers/kube_autoscaler"

	// 平台集群管理路由（DB CRUD，不需要 ClusterMiddleware）
	"k8soperation/internal/app/routers/kube_cluster"

	// CICD 路由
	"k8soperation/internal/app/routers/kube_cicd"

	// RBAC 权限管理路由
	"k8soperation/internal/app/routers/rbac"

	// K8s RBAC 资源路由
	"k8soperation/internal/app/routers/k8srbac"

	// 镜像管理路由
	"k8soperation/internal/app/routers/image"

	// 平台路由
	"k8soperation/internal/app/routers/platform"

	// AI 助手路由
	"k8soperation/internal/app/routers/ai_assistant"

	// LDAP 认证管理路由
	"k8soperation/internal/app/routers/ldap"

	// 监控路由
	"k8soperation/internal/app/routers/monitoring"

	// 你需要的 factory 类型
	"k8soperation/internal/app/services"
)

type injector interface {
	Inject(rg *gin.RouterGroup)
}

func (s *Engine) injectRouterGroup(root *gin.RouterGroup, factory *services.ClusterClientFactory) {
	// ======================================================
	// Swagger（完全公共）
	// ======================================================
	root.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ======================================================
	// API 根路由（子系统前缀隔离，防止与统一平台路由冲突）
	// 统一平台网关路由规则：/api/k8s/* → k8s_operation:8080
	// ======================================================
	api := root.Group("/api")
	v1 := api.Group("/v1")

	// ======================================================
	// License 授权闸门（挂在 /api/v1 最外层，先于所有业务中间件）
	// 未激活/过期时除登录、License 激活、健康检查外全部拒绝
	// 通过 config.yaml License.Enabled 控制开关
	// ======================================================
	if global.LicenseSetting != nil && global.LicenseSetting.Enabled {
		license.Init() // 计算机器码并加载本地 License 文件
		v1.Use(middlewares.LicenseGate())
	}

	// ======================================================
	// License 状态/激活接口（公开，不受登录限流影响）
	// /api/v1/platform/license/...
	// ======================================================
	licensePublic := v1.Group("")
	licensePublic.Use(middlewares.AuthJWTSkip())
	platform.NewLicenseRouter().Inject(licensePublic)

	// ======================================================
	// Public 分组（跳过 JWT）
	// 只放 login/register/logout/refresh（必须公开）
	// ======================================================
	public := v1.Group("")
	public.Use(middlewares.AuthJWTSkip())
	public.Use(middlewares.LoginRateLimit()) // 登录接口限流（每 IP 10次/分钟）
	for _, r := range []injector{
		helloworldrouter.NewHelloWorldRouter(),
		authrouter.NewAuthRouter(),
		ldap.NewLDAPPublicRouter(),
	} {
		r.Inject(public)
	}

	// ======================================================
	// Protected 分组（需要 JWT + 多租户隔离）
	// ======================================================
	protected := v1.Group("")
	protected.Use(middlewares.AuthJWT())
	protected.Use(middlewares.TenantScope())
	for _, r := range []injector{
		userrouter.NewUserRouter(),
	} {
		r.Inject(protected)
	}

	// ======================================================
	// RBAC 权限管理分组（需要 JWT）
	// /api/v1/rbac/...
	// ======================================================
	rbacGroup := protected.Group("/rbac")
	for _, r := range []injector{
		rbac.NewRBACRouter(),
	} {
		r.Inject(rbacGroup)
	}

	// ======================================================
	// 平台健康检查分组（需要 JWT）
	// /api/v1/platform/health/...
	// 注入 factory 以支持多集群汇总
	// ======================================================
	platform.NewPlatformHealthRouterWithFactory(factory).Inject(protected)

	// ======================================================
	// 租户管理（需要 JWT）
	// /api/v1/platform/tenants
	// ======================================================
	platform.NewTenantRouter().Inject(protected)

	// ======================================================
	// 平台系统设置分组（需要 JWT）
	// /api/v1/platform/settings/...
	// ======================================================
	platform.NewPlatformSettingsRouter().Inject(protected)

	// ======================================================
	// 审计日志分组（需要 JWT）
	// /api/v1/platform/audit/...
	// ======================================================
	platform.NewAuditLogRouter().Inject(protected)

	// ======================================================
	// 应用商城分组（需要 JWT）
	// /api/v1/platform/appstore/...
	// 注入 factory 以支持安装应用到集群
	// ======================================================
	platform.NewAppStoreRouterWithFactory(factory).Inject(protected)

	// ======================================================
	// AI 助手分组（需要 JWT）
	// /api/v1/ai/...
	// 包含 AI 对话、高危操作审批等
	// ======================================================
	ai_assistant.NewAIAssistantRouterWithFactory(factory).Inject(protected)

	// ======================================================
	// LDAP 认证管理分组（需要 JWT）
	// /api/v1/ldap/...
	// ======================================================
	ldap.NewLDAPRouter().Inject(protected)

	// ======================================================
	// 监控分组（需要 JWT）
	// /api/v1/monitoring/...
	// ======================================================
	prometheusURL := ""
	if global.MonitoringSetting != nil && global.MonitoringSetting.Enabled {
		prometheusURL = global.MonitoringSetting.PrometheusURL
	}
	monitoring.NewMonitoringRouter(prometheusURL).Inject(protected)

	// ======================================================
	// 镜像管理分组（需要 JWT）
	// /api/v1/image/registry/...
	// ======================================================
	imageGroup := protected.Group("/image")
	for _, r := range []injector{
		image.NewImageRouter(),
	} {
		r.Inject(imageGroup)
	}

	// ======================================================
	// K8s 根分组（同样需要 JWT）
	// 这里直接挂在 protected 下，避免重复 AuthJWT
	// ======================================================
	k8sRoot := protected.Group("/k8s")

	// ======================================================
	// A 类：平台集群管理（DB CRUD）
	// /api/v1/k8s/cluster/...
	// JWT: yes
	// ClusterMiddleware: no
	// ======================================================
	clusterMgmt := k8sRoot.Group("/cluster")
	for _, r := range []injector{
		kube_cluster.NewKubeRouter(),
	} {
		r.Inject(clusterMgmt)
	}

	// ======================================================
	// C-1 类：CICD 回调路由（公开，跳过 JWT）
	// 重要：必须在 cicdGroup 之前注册，避免 Gin 路由树中继承中间件冲突
	// Jenkins 回调接口不需要 JWT 认证，使用 HMAC 签名验证
	// /api/v1/k8s/cicd/pipeline/callback
	// /api/v1/k8s/cicd/stage/callback
	// /api/v1/k8s/cicd/callback/build
	// ======================================================
	cicdPublic := v1.Group("/k8s/cicd")
	cicdPublic.Use(middlewares.AuthJWTSkip())
	kube_cicd.NewCicdCallbackRouter().Inject(cicdPublic)

	// ======================================================
	// C-2 类：CICD 路由（JWT: yes, ClusterMiddleware: no）
	// /api/v1/k8s/cicd/...
	// ======================================================
	cicdGroup := k8sRoot.Group("/cicd")
	for _, r := range []injector{
		kube_cicd.NewCicdRouterWithFactory(factory),
	} {
		r.Inject(cicdGroup)
	}

	// ======================================================
	// B 类：目标集群资源操作（必须传 clusterID）
	// /api/v1/k8s/pod/... /deployment/... /appconfig/...
	// JWT: yes
	// ClusterMiddleware: yes
	// ======================================================
	k8sTarget := k8sRoot.Group("")
	k8sTarget.Use(middlewares.ClusterMiddleware(factory))

	// pod
	pod := k8sTarget.Group("/pod")
	for _, r := range []injector{
		kube_pod.NewkubePodRouter(),
	} {
		r.Inject(pod)
	}

	// deployment
	deployment := k8sTarget.Group("/deployment")
	for _, r := range []injector{
		kube_deployment.NewKubeDeploymentRouter(),
	} {
		r.Inject(deployment)
	}

	// statefulset
	statefulset := k8sTarget.Group("/statefulset")
	for _, r := range []injector{
		kube_statefulset.NewKubeStatefulSetmentRouter(),
	} {
		r.Inject(statefulset)
	}

	// daemonset
	daemonset := k8sTarget.Group("/daemonset")
	for _, r := range []injector{
		kube_daemonset.NewKubeDaemonSetRouter(),
	} {
		r.Inject(daemonset)
	}

	// job
	job := k8sTarget.Group("/job")
	for _, r := range []injector{
		kube_job.NewKubeJobRouter(),
	} {
		r.Inject(job)
	}

	// cronjob
	cronjob := k8sTarget.Group("/cronjob")
	for _, r := range []injector{
		kube_cronjob.NewKubeCronJobRouter(),
	} {
		r.Inject(cronjob)
	}

	// service
	svc := k8sTarget.Group("/service")
	for _, r := range []injector{
		kube_service.NewKubeServiceRouter(),
	} {
		r.Inject(svc)
	}

	// ingress
	ingress := k8sTarget.Group("/ingress")
	for _, r := range []injector{
		kube_ingress.NewKubeIngressRouter(),
	} {
		r.Inject(ingress)
	}

	// secret
	secret := k8sTarget.Group("/secret")
	for _, r := range []injector{
		kube_secret.NewKubeSecretRouter(),
	} {
		r.Inject(secret)
	}

	// configmap
	configmap := k8sTarget.Group("/configmap")
	for _, r := range []injector{
		kube_configmap.NewKubeConfigMapRouter(),
	} {
		r.Inject(configmap)
	}

	// multi-resource (多资源 YAML 解析/应用)
	multiResource := k8sTarget.Group("/multi-resource")
	for _, r := range []injector{
		kube_configmap.NewKubeMultiResourceRouter(),
	} {
		r.Inject(multiResource)
	}

	// storageclass
	storageclass := k8sTarget.Group("/storageclass")
	for _, r := range []injector{
		kube_storageclass.NewKubeStorageClassRouter(),
	} {
		r.Inject(storageclass)
	}

	// pv
	pv := k8sTarget.Group("/pv")
	for _, r := range []injector{
		kube_pv.NewKubePersistentVolumeRouter(),
	} {
		r.Inject(pv)
	}

	// pvc
	pvc := k8sTarget.Group("/pvc")
	for _, r := range []injector{
		kube_pvc.NewKubePersistentVolumeClaimRouter(),
	} {
		r.Inject(pvc)
	}

	// node
	node := k8sTarget.Group("/node")
	for _, r := range []injector{
		kube_node.NewKubeNodeRouter(),
	} {
		r.Inject(node)
	}

	// namespace
	namespace := k8sTarget.Group("/namespace")
	for _, r := range []injector{
		kube_namespace.NewKubeNamespaceRouter(),
	} {
		r.Inject(namespace)
	}

	// appconfig (CRD)
	appconfig := k8sTarget.Group("/appconfig")
	for _, r := range []injector{
		kube_crd.NewKubeAppConfigRouter(),
	} {
		r.Inject(appconfig)
	}

	// CRD/CR 动态资源管理（DynamicClient）
	// /api/v1/k8s/crd/... 和 /api/v1/k8s/cr/...
	kube_crd.NewKubeDynamicCRDRouter().Inject(k8sTarget)

	// HPA 水平弹性扩缩容
	hpa := k8sTarget.Group("/hpa")
	for _, r := range []injector{
		kube_autoscaler.NewKubeHPARouter(),
	} {
		r.Inject(hpa)
	}

	// VPA 垂直弹性扩缩容
	vpa := k8sTarget.Group("/vpa")
	for _, r := range []injector{
		kube_autoscaler.NewKubeVPARouter(),
	} {
		r.Inject(vpa)
	}

	// K8s RBAC (ServiceAccount, Role, RoleBinding)
	for _, r := range []injector{
		k8srbac.NewK8sRBACRouter(),
	} {
		r.Inject(k8sTarget)
	}
}
