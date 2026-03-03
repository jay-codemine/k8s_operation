package initialize

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/services"
	"k8soperation/internal/health"
	"k8soperation/middlewares"
	"time"
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

	// ====== CORS 跨域配置，必须在路由前统一 Use ======
	s.Engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://james521.gnway.cc:30851", // 外网前端地址
			"http://james521.gnway.cc:10537", // 外网后端地址
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
	s.Use(middlewares.Logger())
	s.Use(middlewares.Recovery())
	s.Use(middlewares.K8sError())

	// 注册 session 中间件
	RegisterSession(s.Engine)
}

func (s *Engine) injectRouters() {
	// 1. 取出已经在 injectMiddlewares() 初始化好的 gin.Engine
	r := s.Engine

	// 先注册健康检查（不走 /api 前缀）
	health.Register(r, health.Checks{DB: global.SQLDB})

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

	global.Logger.Info("install session middleware",
		zap.String("cookie_name", name),
	)

	r.Use(sessions.Sessions(name, global.SessionStore))
}
