package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
	"strings"
)

//
// ================= Secret 创建 =================
//

// Secret 类型/模式常量（前端下拉选用）
const (
	SecretTypeOpaque           = "Opaque"
	SecretTypeTLS              = "kubernetes.io/tls"
	SecretTypeDockerConfigJSON = "kubernetes.io/dockerconfigjson"
	SecretTypeBasicAuth        = "kubernetes.io/basic-auth"
	SecretTypeSSHAuth          = "kubernetes.io/ssh-auth"

	SecretModeOpaque    = "opaque"
	SecretModeTLS       = "tls"
	SecretModeDocker    = "docker"
	SecretModeBasicAuth = "basic-auth"
	SecretModeSSHAuth   = "ssh-auth"
)

// 各类型专用字段
type TLSFields struct {
	Cert string `json:"cert"` // -> tls.crt
	Key  string `json:"key"`  // -> tls.key
}

type DockerFields struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

type BasicAuthFields struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SSHAuthFields struct {
	PrivateKey string `json:"private_key"`
	KnownHosts string `json:"known_hosts,omitempty"`
}

func NewKubeSecretCreateRequest() *KubeSecretCreateRequest { return &KubeSecretCreateRequest{} }

type KubeSecretCreateRequest struct {
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name"      valid:"name"`

	// 推荐传 mode（更语义化），兼容 type（原生 k8s 类型字符串）
	Mode string `json:"mode" valid:"-"`
	Type string `json:"type" valid:"-"`

	// 通用键值（Opaque 或直接键值对）
	Data map[string]string `json:"data" swaggertype:"string" valid:"-"`

	// 专用结构体
	TLS    *TLSFields       `json:"tls,omitempty"`
	Docker *DockerFields    `json:"docker,omitempty"`
	Basic  *BasicAuthFields `json:"basic,omitempty"`
	SSH    *SSHAuthFields   `json:"ssh,omitempty"`

	Labels      map[string]string `json:"labels" swaggertype:"string" valid:"-"`
	Annotations map[string]string `json:"annotations" swaggertype:"string" valid:"-"`
	StringData  map[string]string `json:"string_data" swaggertype:"string" valid:"string_data"`
}

// 根据类型做差异化验证
func ValidKubeSecretCreateRequest(data interface{}, c *gin.Context) map[string][]string {
	req := data.(*KubeSecretCreateRequest)
	errs := map[string][]string{}

	// 1) 基础字段
	if req.Namespace == "" {
		errs["namespace"] = append(errs["namespace"], "namespace 不能为空")
	}
	if req.Name == "" {
		errs["name"] = append(errs["name"], "name 不能为空")
	}

	// 2) 规范化：允许前端传 mode 或 type，最终归一到 canonicalType/canonicalMode
	canonicalType, canonicalMode := normalizeTypeAndMode(req.Type, req.Mode)
	if canonicalType == "" && canonicalMode == "" {
		errs["type"] = append(errs["type"], "必须指定 mode 或 type")
		return errs
	}

	// 3) 各类型校验（允许 data 或 string_data 二选一）
	switch canonicalMode {
	case SecretModeOpaque:
		if !hasAnyKV(req) {
			// 以前这里只认 data，这里改成二选一
			errs["data"] = append(errs["data"], "Opaque 模式下需要提供 data 或 string_data（至少一个键）")
		}

	case SecretModeTLS:
		// 允许两种写法：
		// A) 结构体：req.TLS.Cert / req.TLS.Key
		// B) 直接键：tls.crt / tls.key 放到 data 或 string_data
		hasCrt := hasKey(req, "tls.crt") || (req.TLS != nil && req.TLS.Cert != "")
		hasKey := hasKey(req, "tls.key") || (req.TLS != nil && req.TLS.Key != "")
		if !hasCrt || !hasKey {
			errs["tls"] = append(errs["tls"], "TLS 模式下需提供 tls.crt 与 tls.key（支持 string_data/data 或 TLS 结构体）")
		}

	case SecretModeDocker:
		// 允许两种写法：
		// A) 结构体：server/username/password（email 可选）
		// B) 直接键：.dockerconfigjson 放到 data 或 string_data
		hasCfg := hasKey(req, ".dockerconfigjson")
		hasCred := req.Docker != nil && req.Docker.Server != "" && req.Docker.Username != "" && req.Docker.Password != ""
		if !hasCfg && !hasCred {
			errs["docker"] = append(errs["docker"], "Docker 模式下需提供 .dockerconfigjson 或 (server/username/password)")
		}

	case SecretModeBasicAuth:
		// 密码可选，更通用；如需强制可改为 && req.Basic.Password != ""
		hasUser := hasKey(req, "username") || (req.Basic != nil && req.Basic.Username != "")
		if !hasUser {
			errs["basic"] = append(errs["basic"], "BasicAuth 模式下 username 不能为空（支持 string_data/data 或 Basic 结构体）")
		}

	case SecretModeSSHAuth:
		// 允许两种写法：
		// A) 结构体：req.SSH.PrivateKey
		// B) 直接键：ssh-privatekey 放到 data 或 string_data
		hasPk := hasKey(req, "ssh-privatekey") || (req.SSH != nil && req.SSH.PrivateKey != "")
		if !hasPk {
			errs["ssh"] = append(errs["ssh"], "SSH 模式下需提供 ssh-privatekey（支持 string_data/data 或 SSH 结构体）")
		}

	default:
		errs["mode"] = append(errs["mode"], "不支持的 Secret 模式")
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// —— 辅助函数 ——

// 规范化 mode/type，允许两种来源，最终回到统一值
func normalizeTypeAndMode(typ, mode string) (canonicalType, canonicalMode string) {
	t := strings.TrimSpace(strings.ToLower(typ))
	m := strings.TrimSpace(strings.ToLower(mode))

	// 先看 type
	switch t {
	case strings.ToLower(SecretTypeOpaque):
		return SecretTypeOpaque, SecretModeOpaque
	case strings.ToLower(SecretTypeTLS):
		return SecretTypeTLS, SecretModeTLS
	case strings.ToLower(SecretTypeDockerConfigJSON):
		return SecretTypeDockerConfigJSON, SecretModeDocker
	case strings.ToLower(SecretTypeBasicAuth):
		return SecretTypeBasicAuth, SecretModeBasicAuth
	case strings.ToLower(SecretTypeSSHAuth):
		return SecretTypeSSHAuth, SecretModeSSHAuth
	}

	// 再看 mode
	switch m {
	case SecretModeOpaque:
		return SecretTypeOpaque, SecretModeOpaque
	case SecretModeTLS:
		return SecretTypeTLS, SecretModeTLS
	case SecretModeDocker:
		return SecretTypeDockerConfigJSON, SecretModeDocker
	case SecretModeBasicAuth:
		return SecretTypeBasicAuth, SecretModeBasicAuth
	case SecretModeSSHAuth:
		return SecretTypeSSHAuth, SecretModeSSHAuth
	}
	return "", ""
}

// 是否有任意 KV（data 或 string_data）
func hasAnyKV(req *KubeSecretCreateRequest) bool {
	if len(req.Data) > 0 {
		return true
	}
	if len(req.StringData) > 0 {
		return true
	}
	return false
}

// 统一判断键是否存在（data 或 string_data）
func hasKey(req *KubeSecretCreateRequest, key string) bool {
	if v, ok := req.Data[key]; ok && len(v) > 0 {
		return true
	}
	if v, ok := req.StringData[key]; ok && v != "" {
		return true
	}
	return false
}

//
// ================= Secret 列表 =================
//

func NewKubeSecretListRequest() *KubeSecretListRequest { return &KubeSecretListRequest{} }

type KubeSecretListRequest struct {
	KubeCommonRequest
	Page  int    `json:"page" form:"page" valid:"page"`
	Limit int    `json:"limit" form:"limit" valid:"limit"`
	Type  string `json:"type" form:"type" valid:"-"` // 可选过滤类型
}

func ValidKubeSecretListRequest(data interface{}, _ *gin.Context) map[string][]string {
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
// ================= Secret 详情 / 删除 =================
//

func NewKubeSecretDetailRequest() *KubeSecretDetailRequest { return &KubeSecretDetailRequest{} }

type KubeSecretDetailRequest struct {
	KubeCommonRequest
}

func ValidKubeSecretDetailRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{"namespace": {"required"}, "name": {"required"}}
	msgs := govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

func NewKubeSecretDeleteRequest() *KubeSecretDeleteRequest { return &KubeSecretDeleteRequest{} }

type KubeSecretDeleteRequest struct {
	KubeCommonRequest
}

func ValidKubeSecretDeleteRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{"namespace": {"required"}, "name": {"required"}}
	msgs := govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

//
// ================= Secret 通用 Patch 更新 =================
//

func NewKubeSecretUpdateRequest() *KubeSecretUpdateRequest { return &KubeSecretUpdateRequest{} }

type KubeSecretUpdateRequest struct {
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name" valid:"name"`
	Content   string `json:"content" valid:"required"` // JSON 字符串
}

func ValidKubeSecretUpdateRequest(data interface{}, _ *gin.Context) map[string][]string {
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
// ================= Secret 局部更新 =================
//

// 更新 Labels
func NewKubeSecretUpdateLabelsRequest() *KubeSecretUpdateLabelsRequest {
	return &KubeSecretUpdateLabelsRequest{}
}

type KubeSecretUpdateLabelsRequest struct {
	KubeCommonRequest
	Labels map[string]string `json:"labels" valid:"required"`
}

func ValidKubeSecretUpdateLabelsRequest(data interface{}, _ *gin.Context) map[string][]string {
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

// 更新 Annotations
func NewKubeSecretUpdateAnnotationsRequest() *KubeSecretUpdateAnnotationsRequest {
	return &KubeSecretUpdateAnnotationsRequest{}
}

type KubeSecretUpdateAnnotationsRequest struct {
	KubeCommonRequest
	Annotations map[string]string `json:"annotations" valid:"required"`
}

func ValidKubeSecretUpdateAnnotationsRequest(data interface{}, _ *gin.Context) map[string][]string {
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

// 更新 Data
func NewKubeSecretUpdateDataRequest() *KubeSecretUpdateDataRequest {
	return &KubeSecretUpdateDataRequest{}
}

type KubeSecretUpdateDataRequest struct {
	KubeCommonRequest
	Data map[string]string `json:"data" valid:"required"`
}

func ValidKubeSecretUpdateDataRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
		"data":      {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
		"data":      {"required: data 不能为空"},
	})
}

//
// ================= Secret 辅助查询 =================
//

// 查询命名空间内 Secret 名称（带类型过滤）
func NewKubeSecretNamesRequest() *KubeSecretNamesRequest { return &KubeSecretNamesRequest{} }

type KubeSecretNamesRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	Type      string `json:"type" form:"type" valid:"-"`
}

func ValidKubeSecretNamesRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
	})
}

//
// ================= Secret Base64 解码 =================
//

// 请求结构
func NewKubeSecretDecodeRequest() *KubeSecretDecodeRequest {
	return &KubeSecretDecodeRequest{}
}

// KubeSecretDecodeRequest 用于 Base64 解码 Secret 内容
type KubeSecretDecodeRequest struct {
	// 单个字符串解码，例如 "YWRtaW4="
	Value string `json:"value" form:"value"`

	// 批量解码，例如 {"username": "YWRtaW4=", "password": "MTIzNDU2"}
	Data map[string]string `json:"data" form:"data"`
}

// 校验函数
func ValidKubeSecretDecodeRequest(data interface{}, _ *gin.Context) map[string][]string {
	req := data.(*KubeSecretDecodeRequest)
	errs := map[string][]string{}

	// 必须至少提供一个参数
	if req.Value == "" && len(req.Data) == 0 {
		errs["value"] = append(errs["value"], "必须提供 value 或 data 参数")
	}

	return errs
}
