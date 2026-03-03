// Package initialize 用于系统初始化相关逻辑（配置、路由、Swagger 等）
package initialize

import (
	"github.com/gin-gonic/gin"

	"k8soperation/middlewares"

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

	// 平台集群管理路由（DB CRUD，不需要 ClusterMiddleware）
	"k8soperation/internal/app/routers/kube_cluster"

	// CICD 路由
	"k8soperation/internal/app/routers/kube_cicd"

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
	// API 根路由
	// ======================================================
	api := root.Group("/api")
	v1 := api.Group("/v1")

	// ======================================================
	// Public 分组（跳过 JWT）
	// 只放 login/register/logout/refresh（必须公开）
	// ======================================================
	public := v1.Group("")
	public.Use(middlewares.AuthJWTSkip())
	for _, r := range []injector{
		helloworldrouter.NewHelloWorldRouter(),
		authrouter.NewAuthRouter(),
	} {
		r.Inject(public)
	}

	// ======================================================
	// Protected 分组（需要 JWT）
	// ======================================================
	protected := v1.Group("")
	protected.Use(middlewares.AuthJWT())
	for _, r := range []injector{
		userrouter.NewUserRouter(),
	} {
		r.Inject(protected)
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
	// C 类：CICD 路由（JWT: yes, ClusterMiddleware: no）
	// /api/v1/k8s/cicd/...
	// ======================================================
	cicdGroup := k8sRoot.Group("/cicd")
	for _, r := range []injector{
		kube_cicd.NewCicdRouter(),
	} {
		r.Inject(cicdGroup)
	}

	// ======================================================
	// D 类：CICD 回调路由（公开，跳过 JWT）
	// Jenkins 回调接口不需要认证，使用 HMAC 签名验证
	// /api/v1/k8s/cicd/pipeline/callback
	// /api/v1/k8s/cicd/stage/callback
	// /api/v1/k8s/cicd/callback/build
	// ======================================================
	cicdPublic := v1.Group("/k8s/cicd")
	cicdPublic.Use(middlewares.AuthJWTSkip())
	kube_cicd.NewCicdCallbackRouter().Inject(cicdPublic)

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
}
