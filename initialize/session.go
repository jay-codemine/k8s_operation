package initialize

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"k8soperation/global"
)

// SetupSession 初始化 Gin 的 Redis Session Store，并设置 Cookie 选项
// 支持单节点模式和 Cluster 模式（Cluster 模式下使用第一个节点地址）
func SetupSession() error {
	isCluster := len(global.CacheSetting.Addresses) > 0

	// 单节点和集群模式均允许 Username/Password 为空（无密码 Redis）
	// 当 Username/Password 为空时，底层库会自动跳过 AUTH 认证

	// 确定 Session Store 使用的 Redis 地址
	// Cluster 模式下若 Address 为空，自动使用第一个集群节点
	address := global.CacheSetting.Address
	if address == "" && isCluster {
		address = global.CacheSetting.Addresses[0]
		log.Printf("[Session] Cluster 模式: 使用节点 %s 作为 Session Store", address)
	}

	if address == "" {
		return fmt.Errorf("redis address is empty, cannot init session store")
	}

	// 构建 NewStore 参数
	// gin-contrib/sessions/redis.NewStore(size, network, address, username, password, keyPairs...)
	// 当 username/password 为空时，底层库会跳过 AUTH 认证
	store, err := redis.NewStore(
		global.CacheSetting.MaxConnect,
		global.CacheSetting.Network,
		address,
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
	log.Printf("[Session] Redis Session Store 初始化成功 (address=%s, cluster=%v)", address, isCluster)
	return nil
}
