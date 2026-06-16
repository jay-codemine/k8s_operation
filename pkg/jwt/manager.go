// jwt 包实现了JWT(JSON Web Token)的生成和验证功能
// 提供了Token的创建、刷新等核心功能
package jwt

import (
	jwtpkg "github.com/golang-jwt/jwt" // 导入JWT标准库
	"k8soperation/global"              // 导入全局配置
	"k8soperation/pkg/utils"
	"time" // 导入时间处理库
)

// Manager 结构体管理JWT的签名密钥和刷新时间
type Manager struct {
	SignKey    []byte        // JWT签名密钥
	MaxRefresh time.Duration // Token最大刷新时间
}

// New 创建并返回一个新的JWT Manager实例
// 使用全局配置中的JWT签名密钥和最大刷新时间进行初始化
func NewManager() *Manager {
	return &Manager{
		SignKey:    []byte(global.AppSetting.JWTSigningKey),                          // 从全局配置获取签名密钥
		MaxRefresh: time.Duration(global.AppSetting.JWTMaxRefreshTime) * time.Minute, // 从全局配置获取最大刷新时间并转换为time.Duration类型
	}
}

// IssueToken 根据用户ID和用户名生成JWT Token
// userID: 用户唯一标识
// userName: 用户名称
// 返回: 生成的Token字符串和可能的错误
func (m *Manager) IssueToken(userID, userName string) (string, error) {
	now := utils.TimenowInTimezone().Unix() // 获取当前时间戳
	exp := m.expireAtTime()                 // 计算Token过期时间

	// 创建Token的声明(Claims)信息
	claims := Claims{
		UserID:   userID,   // 用户ID
		UserName: userName, // 用户名
		StandardClaims: jwtpkg.StandardClaims{ // JWT标准声明
			NotBefore: now,                       // 生效时间
			IssuedAt:  now,                       // 签发时间
			ExpiresAt: exp,                       // 过期时间
			Issuer:    global.AppSetting.AppName, // 签发者
		},
	}
	return m.createToken(claims)
}

// expireAtTime 计算Token的过期时间戳
// 返回: 过期时间的时间戳
func (m *Manager) expireAtTime() int64 {
	expire := time.Duration(int64(global.AppSetting.JWTExpireTime)) * time.Minute // 从全局配置获取过期时间并转换为time.Duration
	return utils.TimenowInTimezone().Add(expire).Unix()                           // 计算并返回过期时间戳
}

// createToken 根据Claims创建并签名JWT Token
// claims: Token的声明信息
// 返回: 签名后的Token字符串和可能的错误
func (m *Manager) createToken(claims Claims) (string, error) {
	t := jwtpkg.NewWithClaims(jwtpkg.SigningMethodHS256, claims) // 创建新的Token对象，使用HS256签名算法
	return t.SignedString(m.SignKey)                             // 使用签名密钥对Token进行签名并返回
}
