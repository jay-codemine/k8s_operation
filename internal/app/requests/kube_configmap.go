package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

//
// ================= ConfigMap 创建 =================
//

func NewKubeConfigMapCreateRequest() *KubeConfigMapCreateRequest {
	return &KubeConfigMapCreateRequest{}
}

type KubeConfigMapCreateRequest struct {
	Namespace   string            `json:"namespace" valid:"namespace"`
	Name        string            `json:"name"      valid:"name"`
	Labels      map[string]string `json:"labels"       swaggertype:"string" valid:"-"`
	Annotations map[string]string `json:"annotations"  swaggertype:"string" valid:"-"`

	// ConfigMap 的数据部分：文本型与二进制型
	Data       map[string]string `json:"data,omitempty"       swaggertype:"string" valid:"-"`
	BinaryData map[string][]byte `json:"binaryData,omitempty"                      valid:"-"`
}

// 校验：namespace/name 必填；Data 或 BinaryData 至少提供一个键
func ValidKubeConfigMapCreateRequest(data interface{}, _ *gin.Context) map[string][]string {
	req := data.(*KubeConfigMapCreateRequest)
	errs := map[string][]string{}

	if req.Namespace == "" {
		errs["namespace"] = append(errs["namespace"], "namespace 不能为空")
	}
	if req.Name == "" {
		errs["name"] = append(errs["name"], "name 不能为空")
	}

	if !cmHasAnyKV(req) {
		errs["data"] = append(errs["data"], "需要提供 data 或 binaryData（至少一个键）")
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// 是否包含任意键（data 或 binaryData）
func cmHasAnyKV(req *KubeConfigMapCreateRequest) bool {
	if len(req.Data) > 0 {
		return true
	}
	if len(req.BinaryData) > 0 {
		return true
	}
	return false
}

//
// ================= ConfigMap 列表 =================
//

func NewKubeConfigMapListRequest() *KubeConfigMapListRequest { return &KubeConfigMapListRequest{} }

type KubeConfigMapListRequest struct {
	KubeCommonRequest
	Page  int `json:"page"  form:"page"  valid:"page"`
	Limit int `json:"limit" form:"limit" valid:"limit"`
	// 可选过滤：按名称模糊（如果你的 DataSelector 支持 name 过滤）
	Name string `json:"name" form:"name" valid:"-"`
}

func ValidKubeConfigMapListRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		// name：非必填；如果传了，长度 1~64
		"name": []string{"min:1", "max:64"},

		// namespace：非必填；如果传了，长度 1~64
		"namespace": []string{"min:1", "max:64"},

		// page / limit：非必填；如果传了必须合法
		"page":  []string{"min:1"},
		"limit": []string{"min:1", "max:1000"},
	}

	messages := govalidator.MapData{
		"name": {
			"min: name长度不能小于1",
			"max: name长度不能超过64",
		},
		"namespace": {
			"min: namespace长度不能小于1",
			"max: namespace长度不能超过64",
		},
		"page": {
			"min: page必须>=1",
		},
		"limit": {
			"min: limit必须>=1",
			"max: limit不能超过200",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

//
// ================= ConfigMap 详情 / 删除 =================
//

func NewKubeConfigMapDetailRequest() *KubeConfigMapDetailRequest {
	return &KubeConfigMapDetailRequest{}
}

type KubeConfigMapDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeConfigMapDetailRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{"namespace": {"required"}, "name": {"required"}}
	msgs := govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

func NewKubeConfigMapDeleteRequest() *KubeConfigMapDeleteRequest {
	return &KubeConfigMapDeleteRequest{}
}

type KubeConfigMapDeleteRequest struct {
	KubeCommonRequest
}

func ValidKubeConfigMapDeleteRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{"namespace": {"required"}, "name": {"required"}}
	msgs := govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

//
// ================= ConfigMap 通用 Update（传 JSON 字符串，服务端做 Replace/Patch/SSA 均可复用） =================
//

func NewKubeConfigMapUpdateRequest() *KubeConfigMapUpdateRequest {
	return &KubeConfigMapUpdateRequest{}
}

type KubeConfigMapUpdateRequest struct {
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name"      valid:"name"`
	Content   string `json:"content"   valid:"required"` // JSON 字符串（或 YAML 字符串，服务端自行解析）
}

func ValidKubeConfigMapUpdateRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
		"content":   {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
		"content":   {"required: content 不能为空（JSON/YAML 字符串）"},
	})
}

//
// ================= ConfigMap YAML 更新（仅传 yaml 字段） =================
//

func NewKubeConfigMapApplyYamlRequest() *KubeConfigMapApplyYamlRequest {
	return &KubeConfigMapApplyYamlRequest{}
}

type KubeConfigMapApplyYamlRequest struct {
	Yaml string `json:"yaml" valid:"yaml"` // YAML 字符串，包含 namespace 和 name
}

func ValidKubeConfigMapApplyYamlRequest(data interface{}, _ *gin.Context) map[string][]string {
	rulues := govalidator.MapData{"yaml": {"required"}}
	message := govalidator.MapData{
		"yaml": {"required: yaml 不能为空"},
	}
	return valid.ValidateOptions(data, rulues, message)
}

//
// ================= ConfigMap 局部更新（Labels / Annotations / Data） =================
//

func NewKubeConfigMapUpdateLabelsRequest() *KubeConfigMapUpdateLabelsRequest {
	return &KubeConfigMapUpdateLabelsRequest{}
}

type KubeConfigMapUpdateLabelsRequest struct {
	KubeCommonRequest
	Labels map[string]string `json:"labels" valid:"required"`
}

func ValidKubeConfigMapUpdateLabelsRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
		"labels":    {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
		"labels":    {"required: labels 不能为空"},
	})
}

func NewKubeConfigMapUpdateAnnotationsRequest() *KubeConfigMapUpdateAnnotationsRequest {
	return &KubeConfigMapUpdateAnnotationsRequest{}
}

type KubeConfigMapUpdateAnnotationsRequest struct {
	KubeCommonRequest
	Annotations map[string]string `json:"annotations" valid:"required"`
}

func ValidKubeConfigMapUpdateAnnotationsRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace":   {"required"},
		"name":        {"required"},
		"annotations": {"required"},
	}, govalidator.MapData{
		"namespace":   {"required: namespace 不能为空"},
		"name":        {"required: name 不能为空"},
		"annotations": {"required: annotations 不能为空"},
	})
}

func NewKubeConfigMapUpdateDataRequest() *KubeConfigMapUpdateDataRequest {
	return &KubeConfigMapUpdateDataRequest{}
}

type KubeConfigMapUpdateDataRequest struct {
	KubeCommonRequest
	Data       map[string]string `json:"data"       valid:"-"` // 二选一
	BinaryData map[string][]byte `json:"binaryData" valid:"-"` // 二选一
}

func ValidKubeConfigMapUpdateDataRequest(data interface{}, _ *gin.Context) map[string][]string {
	req := data.(*KubeConfigMapUpdateDataRequest)
	errs := map[string][]string{}

	if req.Namespace == "" {
		errs["namespace"] = append(errs["namespace"], "namespace 不能为空")
	}
	if req.Name == "" {
		errs["name"] = append(errs["name"], "name 不能为空")
	}
	if len(req.Data) == 0 && len(req.BinaryData) == 0 {
		errs["data"] = append(errs["data"], "需要提供 data 或 binaryData（至少一个）")
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

//
// ================= ConfigMap 辅助查询（名称列表） =================
//

func NewKubeConfigMapNamesRequest() *KubeConfigMapNamesRequest { return &KubeConfigMapNamesRequest{} }

type KubeConfigMapNamesRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	// 可选：按名称模糊或按 key 是否存在做过滤（看你 service 层是否支持）
	NameLike string `json:"name_like" form:"name_like" valid:"-"`
	Key      string `json:"key"       form:"key"       valid:"-"`
}

func ValidKubeConfigMapNamesRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
	})
}
