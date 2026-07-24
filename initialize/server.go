package initialize

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8soperation/global"
	"k8soperation/internal/app/services"
	"k8soperation/internal/health"
	"k8soperation/middlewares"
	"time"

	// 引入 metrics 包，触发 promauto 注册
	_ "k8soperation/pkg/metrics"
)

type Engine struct {
	*gin.Engine
	Mode    string
	Factory *services.ClusterClientFactory
}

// NewEngine 创建一个新的引擎实例
// 返回一个初始化完成的 Engine 指针
func NewEngine(factory *services.ClusterClientFactory) *Engine {
	g := &Engine{
		Mode:    gin.ReleaseMode,
		Factory: factory,
	}

	gin.SetMode(g.Mode)

	g.injectMiddlewares()
	g.injectRouters()

	return g
}

func (s *Engine) injectMiddlewares() {
	// 初始化并赋值 gin.Engine
	s.Engine = gin.New()

	// 设置 multipart 内存阈值（128MB），避免大 JAR/制品文件先落盘临时文件再读取，加速上传
	s.Engine.MaxMultipartMemory = 128 << 20 // 128MB

	// ====== CORS 跨域配置，必须在路由前统一 Use ======
	// 使用 AllowOriginFunc 动态判断，支持 Nginx 反代同源部署 + 内网穿透 + 本地开发
	s.Engine.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// 允许所有来源（前后端通过 Nginx 合一部署时，Origin 与服务地址一致）
			// 生产环境如需收紧，可改为白名单判断
			return true
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
			"X-Cluster-ID",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 判断是否为测试模式
	if s.Mode == gin.TestMode {
		return
	}

	// 注册中间件
	s.Use(middlewares.PrometheusMiddleware()) // Prometheus 指标采集（必须在 Logger 之前）
	s.Use(middlewares.RateLimit())            // IP 级别限流（100 req/s，防止滥用）
	s.Use(middlewares.Logger())
	s.Use(middlewares.Recovery())
	s.Use(middlewares.K8sError())
	s.Use(middlewares.Audit(nil)) // 审计日志中间件（默认配置）

	// 注册 session 中间件
	RegisterSession(s.Engine)
}

func (s *Engine) injectRouters() {
	// 1. 取出已经在 injectMiddlewares() 初始化好的 gin.Engine
	r := s.Engine

	// 先注册健康检查（不走 /api 前缀）
	// 就绪探针 DB Ping 超时从配置 Server.ReadinessTimeout(秒) 注入，<=0 时 health 内部兜底为 2s
	health.Register(r, health.Checks{
		DB:           global.SQLDB,
		ReadyTimeout: global.ServerSetting.ReadinessTimeout * time.Second,
	})

	// 注册 Prometheus /metrics 端点（不走 JWT，公开暴露供 Prometheus 抓取）
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 2. 创建一个根分组
	//apiRouter 挂的路由，路径就是从根开始，比如 /login、/user。
	//这个是“根级分组”，用来直接注册全局路由。
	apiRouter := r.Group("")

	// 3. 把分组传给 injectRouterGroup，让它批量注册模块路由
	// apiRouter 挂的路由，路径就是从根开始，比如 /login、/user。
	//这个是“根级分组”，用来直接注册全局路由，或者后面再细分子分组。
	s.injectRouterGroup(apiRouter, s.Factory)

}

func RegisterSession(r *gin.Engine) {
	if global.SessionStore == nil {
		global.Logger.Warn("session store is nil, session middleware not installed")
		return
	}
	name := global.CacheSetting.Name
	if name == "" {
		name = "k8soperation_sid" // 兜底
	}

	r.Use(sessions.Sessions(name, global.SessionStore))
}
