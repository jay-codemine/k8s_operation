package valid

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/thedevsaddam/govalidator"
	"go.uber.org/zap"
	"io"
	"k8soperation/global"
	"k8soperation/internal/errorcode"
	"k8soperation/pkg/app/response"
	"net/http"
	"strings"
)

type ValidationErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

// ---------------------------------------------
// 通用验证入口
// ---------------------------------------------
// ValidatorFunc 验证函数类型签名
// 任何函数只要符合 (数据, gin.Context) -> 错误 map 的形式，就可以作为校验处理函数
// 返回值 map[string][]string ：key 是字段名，value 是该字段的所有错误提示
type ValidatorFunc func(interface{}, *gin.Context) map[string][]string

func Validate(c *gin.Context, obj interface{}, handler ValidatorFunc) bool {
	devMode := global.ServerSetting.RunMode != "release"
	const maxLogBytes = 4096

	// ======== 1. 打印 raw body，仅开发模式 ========
	if devMode && c.Request.Method != http.MethodGet {
		raw, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
		if len(raw) > maxLogBytes {
			raw = raw[:maxLogBytes]
		}
		global.Logger.Info("REQ RAW BODY", zap.ByteString("body", raw))
	}

	// ======== 2. 参数绑定 ========
	var err error
	if c.Request.Method == http.MethodGet {
		err = c.ShouldBindQuery(obj)
	} else {
		ct := strings.ToLower(c.GetHeader("Content-Type"))
		switch {
		case strings.Contains(ct, "application/json"):
			err = c.ShouldBindBodyWith(obj, binding.JSON)
		case strings.Contains(ct, "application/x-www-form-urlencoded"),
			strings.Contains(ct, "multipart/form-data"):
			err = c.ShouldBind(obj)
		default:
			err = c.ShouldBind(obj)
		}
	}

	if err != nil {
		// 返回统一格式，兼容旧逻辑
		response.NewResponse(c).ToErrorResponse(
			errorcode.InvalidParams.WithDetails(err.Error()),
		)
		return false
	}

	// ======== 3. 字段校验规则 ========
	if handler != nil {
		errs := handler(obj, c)
		if len(errs) > 0 {
			var details []string
			for field, msgs := range errs {
				for _, msg := range msgs {
					details = append(details, fmt.Sprintf("%s: %s", field, msg))
				}
			}

			// 统一 errorcode 返回前端
			response.NewResponse(c).ToErrorResponse(
				errorcode.InvalidParams.WithDetails(details...),
			)
			return false
		}
	}

	return true // 通过校验
}

// ---------------------------------------------
// 直接调用 govalidator 库的封装
// ---------------------------------------------
// ValidateOptions 底层调用 govalidator 库，直接传入 struct、rules、messages
// 返回值：map[string][]string（字段对应的所有错误提示）
func ValidateOptions(data interface{}, rules govalidator.MapData, messages govalidator.MapData) map[string][]string {
	opts := govalidator.Options{
		Data:          data,     // 要校验的数据（struct 指针）
		Rules:         rules,    // 校验规则
		TagIdentifier: "valid",  // struct tag 使用的标识符，例如 `valid:"required"`
		Messages:      messages, // 自定义错误提示
	}
	return govalidator.New(opts).ValidateStruct()
}

// ---------------------------------------------
// 自定义扩展：校验两次输入的密码是否一致
// ---------------------------------------------
func ValidatePasswordConfirm(password, PasswordConfirm string, errs map[string][]string) map[string][]string {
	if password != PasswordConfirm {
		errs["PasswordConfirm"] = append(errs["PasswordConfirm"], "两次输入的密码不一致")
	}
	return errs
}
