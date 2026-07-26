package initialize

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
	"k8soperation/internal/app/services"
)

func logK8sInitWarn(msg string, fields ...zap.Field) {
	if global.LogControlSetting != nil && global.LogControlSetting.SuppressK8sClusterInitWarn {
		return
	}
	global.Logger.Warn(msg, fields...)
}

// SetupK8sBootstrap 初始化 K8s 集群连接
// 从数据库加载 DefaultClusterID 的集群；DB 无数据则允许空启动
func SetupK8sBootstrap() error {
	svc := services.NewBackgroundServices()
	ctx := context.Background()

	cli, err := svc.K8sClusterInit(ctx, &requests.K8sClusterInitRequest{
		ID: global.AppSetting.DefaultClusterID,
	})

	if err == nil {
		setGlobalClients(cli)
		printClusterInfo("数据库集群")
		return nil
	}

	logK8sInitWarn("数据库集群初始化失败，允许空启动（请通过界面添加集群）",
		zap.Uint32("cluster_id", global.AppSetting.DefaultClusterID),
		zap.Error(err))
	printEmptyStartWarning()
	return nil
}

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

func printClusterInfo(source string) {
	fmt.Printf("K8s 集群初始化成功（来源：%s）\n", source)
}

func printEmptyStartWarning() {
	fmt.Println("K8s 集群未初始化（空启动模式）")
	fmt.Println("→ 请通过 Web 界面添加集群")
	fmt.Println("→ 集群管理功能暂不可用，其他功能正常")
}
