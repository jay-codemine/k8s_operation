package models

// ErrorResponse Swagger 文档用的错误响应结构
// @Description 错误响应
type ErrorResponse struct {
	Code    int      `json:"code" example:"400001"`
	Msg     string   `json:"msg" example:"请求参数错误"`
	Details []string `json:"details,omitempty"`
}

// SuccessResponse Swagger 文档用的成功响应结构
// @Description 成功响应
type SuccessResponse struct {
	Code int         `json:"code" example:"0"`
	Msg  string      `json:"msg" example:"OK"`
	Data interface{} `json:"data,omitempty"`
}
