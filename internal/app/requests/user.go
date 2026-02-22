package requests

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// UserCreateRequest 用户注册请求参数结构体
// - `json:"username"`         → 绑定 JSON 请求体中的 "username"
// - `form:"username"`         → 绑定表单参数中的 "username"
// - `valid:"username"`        → 在 rules/messages 里对应的字段名（govalidator 识别用）
type UserCreateRequest struct {
	Username        string `json:"username" form:"username" valid:"username"`
	Password        string `json:"password" form:"password" valid:"password"`
	PasswordConfirm string `json:"password_confirm" form:"password_confirm" valid:"password_confirm"`
}

func NewUserUserCreateRequest() *UserCreateRequest {
	return &UserCreateRequest{
		Password: "123456",
	}
}

// VaildUserCreateRequest 用户创建请求的验证函数
// 参数：
//   - data → 要验证的数据，一般传 &UserCreateRequest{}
//   - ctx  → gin.Context，上下文对象（可选，虽然这里没用上）
//
// 返回值：
//   - map[string][]string，key 是字段名，value 是该字段的错误信息数组
func VaildUserCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	// ctx 预留：未来可能需要用来做多语言错误提示/日志 trace

	// 定义规则（rules）
	rules := govalidator.MapData{
		// 用户名：必填 & 在 user 表中必须唯一
		"username": []string{"required", "not_exists:user,username"},
		// 密码：必填 & 至少 6 位
		"password": []string{"required", "min:6"},
		// 确认密码：必填
		"password_confirm": []string{"required"},
	}

	// 定义错误提示（messages）
	messages := govalidator.MapData{
		"username": []string{
			"required: 用户名为必填字段,字段为 username",
		},
		"password": []string{
			"required: 密码为必填字段,字段为 password",
			"min:密码长度需大于 6",
		},
		"password_confirm": []string{
			"required:确认密码框为必填项",
		},
	}

	// 调用封装好的通用验证方法（底层用 govalidator.New(opts).ValidateStruct()）
	errs := valid.ValidateOptions(data, rules, messages)

	// ------------------------------
	// 自定义校验：确认密码是否一致
	// ------------------------------

	// 第一种写法（手动）：
	// if data.(*UserCreateRequest).Password != data.(*UserCreateRequest).PasswordConfirm {
	//     // append() 往原有的错误数组里追加新的错误提示
	//     errs["password_confirm"] = append(errs["password_confirm"], "两次输入的密码不一致")
	// }

	// 第二种写法（推荐，调用封装函数）：
	// 使用 valid 包里定义的 ValidatePasswordConfirm
	errs = valid.ValidatePasswordConfirm(
		data.(*UserCreateRequest).Password,
		data.(*UserCreateRequest).PasswordConfirm,
		errs,
	)

	// 返回所有错误结果
	return errs
}

type UserUpdateRequest struct {
	ID       uint32 `json:"id" form:"id" valid:"id"`
	Username string `json:"username" form:"username" valid:"username"`
	Password string `json:"password" form:"password" valid:"password"`
}

func NewUserUpdateRequest() *UserUpdateRequest {
	return &UserUpdateRequest{}
}

func ValidUserUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	req := data.(*UserUpdateRequest)

	rules := govalidator.MapData{
		"id": []string{"required"},
		"username": []string{
			"required",
			fmt.Sprintf("not_exists:user,username,except_id=%d", req.ID),
		},
		"password": []string{"required", "min:6"},
	}

	messages := govalidator.MapData{
		"username": []string{
			"required: 用户名为必填字段,字段为 username",
			"not_exists: 用户名已存在",
		},
		"password": []string{
			"required: 密码为必填字段,字段为 password",
			"min: 密码长度需大于 6",
		},
		"id": []string{
			"required: 用户ID是必填项, 字段为 id",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

type UserListRequest struct {
	Username string `json:"username,omitempty" form:"username"`
	Page     int    `json:"page,omitempty" form:"page" valid:"page" description:"页码"`
	Limit    int    `json:"limit,omitempty" form:"limit" valid:"limit" description:"每页数量"`
}

func NewUserListRequest() *UserListRequest {
	return &UserListRequest{}
}

func ValidUserListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"page":  []string{"required"},
		"limit": []string{"required"},
	}

	message := govalidator.MapData{
		"page":  []string{"required:页码为必填项"},
		"limit": []string{"required:每页数量为必填项"},
	}

	// 校验入参
	return valid.ValidateOptions(data, rules, message)
}
