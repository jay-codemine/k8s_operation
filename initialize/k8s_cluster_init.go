package initialize

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
)

// SetupK8sBootstrap 初始化 K8s 集群连接
// 优化逻辑：
// 1. 优先从数据库加载 DefaultClusterID 的集群
// 2. 如果 DB 无数据，尝试读取本地 kubeconfig 并自动入库
// 3. 如果都没有，允许空启动（用户后续通过界面添加）
func SetupK8sBootstrap() error {
	svc := services.NewBackgroundServices()
	ctx := context.Background()

	// 1) 尝试从 DB 加载默认集群
	cli, err := svc.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{
		ID: global.AppSetting.DefaultClusterID,
	})

	if err == nil {
		// DB 有数据，初始化成功
		setGlobalClients(cli)
		printClusterInfo("数据库集群")
		return nil
	}

	// 2) DB 无数据或初始化失败，尝试从本地 kubeconfig 文件加载
	global.Logger.Warn("从数据库加载集群失败，尝试读取本地 kubeconfig...",
		zap.Uint32("cluster_id", global.AppSetting.DefaultClusterID),
		zap.Error(err))

	localKubeConfig := global.AppSetting.GlobalKubeConfigPath
	if localKubeConfig == "" {
		localKubeConfig = "configs/k8s.yaml"
	}

	kubeConfigContent, readErr := os.ReadFile(localKubeConfig)
	if readErr != nil {
		// 本地文件也不存在，允许空启动
		global.Logger.Warn("本地 kubeconfig 文件不存在，允许空启动（请通过界面添加集群）",
			zap.String("path", localKubeConfig),
			zap.Error(readErr))
		printEmptyStartWarning()
		return nil
	}

	kubeConfigStr := strings.TrimSpace(string(kubeConfigContent))
	if kubeConfigStr == "" {
		global.Logger.Warn("本地 kubeconfig 文件为空，允许空启动（请通过界面添加集群）",
			zap.String("path", localKubeConfig))
		printEmptyStartWarning()
		return nil
	}

	// 3) 本地文件存在，尝试自动入库
	global.Logger.Info("发现本地 kubeconfig，尝试自动入库...", zap.String("path", localKubeConfig))

	createErr := svc.K8sClusterCreate(ctx, &requests.K8sClusterCreateRequest{
		ClusterName:    "default-cluster",
		ClusterVersion: "auto-detected",
		KubeConfig:     kubeConfigStr,
	})
	if createErr != nil {
		// 入库失败（可能是重复名称等），尝试直接用本地配置初始化
		global.Logger.Warn("自动入库失败，尝试直接使用本地配置初始化",
			zap.Error(createErr))

		// 直接从本地配置构建 client
		cli, buildErr := services.BuildClientsFromKubeconfig(kubeConfigStr)
		if buildErr != nil {
			global.Logger.Warn("本地 kubeconfig 无法初始化，允许空启动",
				zap.Error(buildErr))
			printEmptyStartWarning()
			return nil
		}

		setGlobalClients(cli)
		printClusterInfo("本地 kubeconfig（未入库）")
		return nil
	}

	// 4) 入库成功，重新初始化
	global.Logger.Info("本地 kubeconfig 已自动入库，正在初始化...")
	cli, initErr := svc.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{
		ID: 1, // 新创建的记录 ID 通常为 1
	})
	if initErr != nil {
		global.Logger.Warn("入库后初始化失败，允许空启动", zap.Error(initErr))
		printEmptyStartWarning()
		return nil
	}

	setGlobalClients(cli)
	printClusterInfo("本地 kubeconfig（已入库）")
	return nil
}

// setGlobalClients 设置全局 K8s 客户端
func setGlobalClients(cli *services.K8sClients) {
	global.ManagementKubeConfig = cli.Config
	global.ManagementKubeClient = cli.Kube
	global.ManagementMetricsClient = cli.Metrics
	global.ManagementSupportsEventsV1 = cli.SupportsEvV1

	global.KubeConfig = cli.Config
	global.KubeClient = cli.Kube
	global.MetricsClient = cli.Metrics
	global.SupportsEventsV1 = cli.SupportsEvV1

	if cli.Metrics == nil {
		global.Logger.Warn("metrics client not initialized (metrics-server not installed?)")
	}

	global.Logger.Warn(
		"【多集群模式】global.KubeClient 仅代表管理集群，请勿在业务请求中直接使用",
	)
}

// printClusterInfo 打印集群初始化成功信息
func printClusterInfo(source string) {
	fmt.Printf("K8s 集群初始化成功（来源：%s）\n", source)
	if global.SupportsEventsV1 {
		fmt.Println("→ 支持新版事件 API：events.k8s.io/v1")
	} else {
		fmt.Println("→ 不支持新版事件 API，自动回退至 core/v1")
	}
}

// printEmptyStartWarning 打印空启动警告
func printEmptyStartWarning() {
	fmt.Println("K8s 集群未初始化（空启动模式）")
	fmt.Println("→ 请通过 Web 界面添加集群，或将 kubeconfig 放到 configs/k8s.yaml")
	fmt.Println("→ 集群管理功能暂不可用，其他功能正常")
}
