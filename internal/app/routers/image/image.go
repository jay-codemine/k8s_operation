package image

import (
	"github.com/gin-gonic/gin"
	v1 "k8soperation/internal/app/controllers/api/v1/image"
)

type ImageRouter struct{}

func NewImageRouter() *ImageRouter {
	return &ImageRouter{}
}

func (r *ImageRouter) Inject(router *gin.RouterGroup) {
	registryCtrl := v1.NewImageRegistryController()
	browseCtrl := v1.NewImageBrowseController()
	cleanupCtrl := v1.NewCleanupPolicyController()

	// 镜像仓库管理
	registry := router.Group("/registry")
	{
		registry.GET("/list", registryCtrl.List)          // 分页列表
		registry.GET("/all", registryCtrl.ListAll)        // 所有仓库（下拉用）
		registry.GET("/detail", registryCtrl.Detail)      // 详情
		registry.GET("/stats", registryCtrl.Stats)        // 统计
		registry.POST("/create", registryCtrl.Create)     // 创建
		registry.POST("/update", registryCtrl.Update)     // 更新
		registry.POST("/delete", registryCtrl.Delete)     // 删除
		registry.POST("/check", registryCtrl.CheckConnection) // 检测连接
		registry.POST("/default", registryCtrl.SetDefault)    // 设置默认
	}

	// 镜像浏览
	browse := router.Group("/browse")
	{
		browse.GET("/repositories", browseCtrl.ListRepositories) // 镜像项目列表
		browse.GET("/tags", browseCtrl.ListTags)                 // 镜像标签列表
		browse.GET("/detail", browseCtrl.GetImageDetail)         // 镜像详情
		browse.POST("/delete", browseCtrl.DeleteTag)             // 删除标签
	}

	// 清理策略
	cleanup := router.Group("/cleanup")
	{
		cleanup.GET("/list", cleanupCtrl.List)       // 策略列表
		cleanup.GET("/logs", cleanupCtrl.Logs)       // 清理日志
		cleanup.POST("/create", cleanupCtrl.Create)  // 创建策略
		cleanup.POST("/update", cleanupCtrl.Update)  // 更新策略
		cleanup.POST("/delete", cleanupCtrl.Delete)  // 删除策略
		cleanup.POST("/toggle", cleanupCtrl.Toggle)  // 启用/禁用
		cleanup.POST("/run", cleanupCtrl.Run)        // 手动执行
	}
}
