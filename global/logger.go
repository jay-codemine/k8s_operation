package global

import (
	"k8soperation/pkg/logger"
)

var (
	Logger         *logger.Logger // 系统日志
	BizLogger      *logger.Logger // 业务日志
	MirrorBizToSys bool           // 业务日志是否镜像到系统日志
)
