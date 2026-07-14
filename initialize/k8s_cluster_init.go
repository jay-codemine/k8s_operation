package initialize

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
)

var ErrNoClusterConfig = errors.New("生产环境禁止空启动：无法加载 K8s 集群配置，请检查数据库或 kubeconfig 文件")
// logK8sInitWarn K8s 启动 WARN（可由 LogControl.SuppressK8sClusterInitWarn 抑制）
func logK8sInitWarn(msg string, fields ...zap.Field) {
	if global.LogControlSetting != nil && global.LogControlSetting.SuppressK8sClusterInitWarn {
		return
	}
	global.Logger.Warn(msg, fields...)
}

// logK8sInitError K8s 启动 ERROR（可由 LogControl.SuppressK8sClusterInitError 抑制）
func logK8sInitError(msg string, fields ...zap.Field) {
	if global.LogControlSetting != nil && global.LogControlSetting.SuppressK8sClusterInitError {
		return
	}
	global.Logger.Error(msg, fields...)
}

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
	logK8sInitWarn("从数据库加载集群失败，尝试读取本地 kubeconfig...",
		zap.Uint32("cluster_id", global.AppSetting.DefaultClusterID),
		zap.Error(err))

	localKubeConfig := global.AppSetting.GlobalKubeConfigPath
	if localKubeConfig == "" {
		localKubeConfig = "configs/k8s.yaml"
	}

	kubeConfigContent, readErr := os.ReadFile(localKubeConfig)
	if readErr != nil {
		// 本地文件也不存在
		if !global.AppSetting.AllowEmptyStart {
			// 生产环境不允许空启动，返回错误
			logK8sInitError("生产环境禁止空启动",
				zap.String("path", localKubeConfig),
				zap.Error(readErr))
			return ErrNoClusterConfig
		}
		// 开发环境允许空启动
		logK8sInitWarn("本地 kubeconfig 文件不存在，允许空启动（请通过界面添加集群）",
			zap.String("path", localKubeConfig),
			zap.Error(readErr))
		printEmptyStartWarning()
		return nil
	}

	kubeConfigStr := strings.TrimSpace(string(kubeConfigContent))
	if kubeConfigStr == "" {
		if !global.AppSetting.AllowEmptyStart {
			logK8sInitError("生产环境禁止空启动：kubeconfig 文件为空",
				zap.String("path", localKubeConfig))
			return ErrNoClusterConfig
		}
		logK8sInitWarn("本地 kubeconfig 文件为空，允许空启动（请通过界面添加集群）",
			zap.String("path", localKubeConfig))
		printEmptyStartWarning()
		return nil
	}

	// 3) 本地文件存在，先检查是否已有 "default-cluster" 记录（避免重复创建）
	global.Logger.Info("发现本地 kubeconfig，检查是否已入库...", zap.String("path", localKubeConfig))

	existing, _ := svc.K8sClusterGetByName(ctx, "default-cluster")
	if existing != nil {
		// 已存在，直接用它初始化
		global.Logger.Info("已存在 default-cluster 记录，复用", zap.Uint32("id", existing.ID))
		cli, initErr := svc.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{ID: existing.ID})
		if initErr == nil {
			setGlobalClients(cli)
			printClusterInfo("数据库集群(default-cluster)")
			return nil
		}
		logK8sInitWarn("复用 default-cluster 初始化失败，尝试更新 kubeconfig", zap.Error(initErr))
		// 更新 kubeconfig 后重试
		_ = svc.K8sClusterUpdate(ctx, &requests.K8sClusterUpdateRequest{
			ID: existing.ID, ClusterName: existing.ClusterName,
			ClusterVersion: existing.ClusterVersion, KubeConfig: kubeConfigStr,
		})
		cli, initErr = svc.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{ID: existing.ID})
		if initErr == nil {
			setGlobalClients(cli)
			printClusterInfo("数据库集群(default-cluster, kubeconfig已更新)")
			return nil
		}
	}

	// 4) 不存在，创建新记录
	createErr := svc.K8sClusterCreate(ctx, &requests.K8sClusterCreateRequest{
		ClusterName:    "default-cluster",
		ClusterVersion: "auto-detected",
		KubeConfig:     kubeConfigStr,
	})
	if createErr != nil {
		// 入库失败，尝试直接用本地配置初始化
		logK8sInitWarn("自动入库失败，尝试直接使用本地配置初始化", zap.Error(createErr))
		cli, buildErr := services.BuildClientsFromKubeconfig(kubeConfigStr)
		if buildErr != nil {
			if !global.AppSetting.AllowEmptyStart {
				logK8sInitError("生产环境禁止空启动：本地 kubeconfig 无法初始化", zap.Error(buildErr))
				return ErrNoClusterConfig
			}
			logK8sInitWarn("本地 kubeconfig 无法初始化，允许空启动", zap.Error(buildErr))
			printEmptyStartWarning()
			return nil
		}
		setGlobalClients(cli)
		printClusterInfo("本地 kubeconfig（未入库）")
		return nil
	}

	// 5) 入库成功，查询刚创建的记录获取 ID
	global.Logger.Info("本地 kubeconfig 已自动入库，正在初始化...")
	newCluster, _ := svc.K8sClusterGetByName(ctx, "default-cluster")
	var initID uint32 = 1
	if newCluster != nil {
		initID = newCluster.ID
	}
	cli, initErr := svc.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{ID: initID})
	if initErr != nil {
		if !global.AppSetting.AllowEmptyStart {
			logK8sInitError("生产环境禁止空启动：入库后初始化失败", zap.Error(initErr))
			return ErrNoClusterConfig
		}
		logK8sInitWarn("入库后初始化失败，允许空启动", zap.Error(initErr))
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
		logK8sInitWarn("metrics client not initialized (metrics-server not installed?)")
	}
}

// printClusterInfo 打印集群初始化成功信息
func printClusterInfo(source string) {
	fmt.Printf("K8s 集群初始化成功（来源：%s）\n", source)
}

// printEmptyStartWarning 打印空启动警告
func printEmptyStartWarning() {
	fmt.Println("K8s 集群未初始化（空启动模式）")
	fmt.Println("→ 请通过 Web 界面添加集群，或将 kubeconfig 放到 configs/k8s.yaml")
	fmt.Println("→ 集群管理功能暂不可用，其他功能正常")
}
