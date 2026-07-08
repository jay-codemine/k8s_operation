package jwt

import (
	jwtpkg "github.com/golang-jwt/jwt"
	"k8soperation/internal/errorcode"
)

// ParseToken 从 token 字符串解析并校验
// ParseToken 解析并验证JWT令牌，返回声明信息
// 参数:
//   - tokenString: 要解析的JWT令牌字符串

// 返回值:
//   - *Claims: 解析后的声明信息
//   - error: 解析过程中出现的错误
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	// 创建JWT解析器，指定使用HS256签名方法
	parser := &jwtpkg.Parser{ValidMethods: []string{jwtpkg.SigningMethodHS256.Alg()}}
	// 使用解析器解析令牌，传入空Claims结构体和签名密钥
	token, err := parser.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwtpkg.Token) (interface{}, error) { return m.SignKey, nil },
	)
	// 处理解析过程中可能出现的错误
	if err != nil {
		// 检查错误类型是否为JWT验证错误
		if ve, ok := err.(*jwtpkg.ValidationError); ok {
			// 判断令牌格式是否错误
			if ve.Errors&jwtpkg.ValidationErrorMalformed != 0 {
				return nil, errorcode.ErrTokenMalformed
			}
			// 判断令牌是否过期
			if ve.Errors&jwtpkg.ValidationErrorExpired != 0 {
				return nil, errorcode.ErrTokenExpired
			}
			// 其他类型的令牌错误
			return nil, errorcode.ErrTokenInvalid
		}
		// 返回其他类型的错误
		return nil, err
	}

	// 验证令牌是否有效，并获取声明信息
	cs, ok := token.Claims.(*Claims)
	// 检查声明类型转换是否成功，以及令牌是否有效
	if !ok || !token.Valid {
		return nil, errorcode.ErrTokenInvalid
		// 返回解析后的声明信息
	}
	return cs, nil
}
