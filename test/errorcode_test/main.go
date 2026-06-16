package main

import (
	"fmt"
	"k8soperation/internal/errorcode"
	"net/http"
)

func main() {
	// 1. 基础错误信息
	fmt.Println(errorcode.Success.Error())     // 错误码: 0，错误信息: 成功
	fmt.Println(errorcode.ServerError.Error()) // 错误码: 100001，错误信息: 服务内部错误

	// 2. 使用 Msgf
	err := errorcode.InvalidParams.WithDetails("字段 username 缺失")
	// 输出: 参数 username 不合法
	fmt.Println("Msgf:", err.Error())

	// 3. withDetails
	err = errorcode.InvalidParams.WithDetails("字段 username 缺失", "请求ID=12345")
	fmt.Println("Details:", err.Error())

	// 4. StatusCode 映射
	fmt.Println("Success =>", errorcode.Success.StatusCode() == http.StatusOK)                     // true
	fmt.Println("InvalidParams =>", errorcode.InvalidParams.StatusCode() == http.StatusBadRequest) // true
}
