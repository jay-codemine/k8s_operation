package initialize

import (
	"fmt"
	"go.uber.org/zap"
	"k8soperation/global"
)

// LogDocsReady 打印可点击的 Swagger 文档链接
func LogDocsReady() {
	// 这里用 127.0.0.1；如果你有 Host/Scheme 配置，可自行替换
	base := fmt.Sprintf("http://127.0.0.1:%s", global.ServerSetting.Port)

	global.Logger.Info("docs ready",
		zap.String("swagger_ui", base+"/swagger/index.html"),
		zap.String("doc_json", base+"/swagger/doc.json"),
	)
}
