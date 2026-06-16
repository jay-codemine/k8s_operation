package global

import (
	"k8soperation/pkg/logger"
)

var (
	Logger         *logger.Logger // 系统日志
	BizLogger      *logger.Logger // 业务日志
	AILogger       *logger.Logger // AI 助手专属日志（独立文件，方便排查大模型问题）
	MirrorBizToSys bool           // 业务日志是否镜像到系统日志
)
