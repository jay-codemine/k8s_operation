package initialize

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"k8soperation/global"
)

// SetupSession 初始化 Gin 的 Redis Session Store，并设置 Cookie 选项
func SetupSession() error {
	if global.CacheSetting.Username == "" {
		global.Logger.Error("redis 用户名不能为空")
		return fmt.Errorf("redis username is empty")
	}
	if global.CacheSetting.Password == "" {
		global.Logger.Error("redis 密码不能为空")
		return fmt.Errorf("redis password is empty")
	}

	store, err := redis.NewStore(
		global.CacheSetting.MaxConnect,
		global.CacheSetting.Network,
		global.CacheSetting.Address,
		global.CacheSetting.Username,
		global.CacheSetting.Password,
		[]byte(global.CacheSetting.Secret),
	)
	if err != nil {
		return fmt.Errorf("new redis session store failed: %w", err)
	}

	secure := global.ServerSetting.RunMode == "release"
	sameSite := http.SameSiteLaxMode

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})

	global.SessionStore = store
	return nil
}
