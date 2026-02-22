package requests

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/pkg/valid"
)

//
// ================= PersistentVolumeClaim 创建 =================
//

func NewKubePVCCreateRequest() *KubePVCCreateRequest { return &KubePVCCreateRequest{} }

type KubePVCCreateRequest struct {
	// 支持两种模式：表单模式和 YAML 模式
	// 1. 如果有 YamlContent，则使用 YAML 创建
	// 2. 否则使用表单字段创建
	YamlContent string `json:"yamlContent" valid:"-"` // YAML 模式
	
	// 表单模式字段
	Namespace        string                              `json:"namespace"        valid:"namespace"`
	Name             string                              `json:"name"             valid:"name"`
	Storage          string                              `json:"storage"          valid:"-"` // e.g. "10Gi"
	AccessMode       string                              `json:"accessMode"       valid:"-"` // 单个访问模式，前端发送
	AccessModes      []corev1.PersistentVolumeAccessMode `json:"accessModes"      valid:"-"` // 兼容旧版
	StorageClassName string                              `json:"storageClassName" valid:"-"`     // 可为空
	VolumeMode       *corev1.PersistentVolumeMode        `json:"volumeMode,omitempty" valid:"-"` // Filesystem/Block

	// 选择器：用于绑定满足标签/表达式的 PV（可选）
	SelectorMatchLabels map[string]string `json:"selectorMatchLabels,omitempty" swaggertype:"string" valid:"-"`
	// 后续如果需要 MatchExpressions 可再扩展

	Labels      map[string]string `json:"labels,omitempty"      swaggertype:"string" valid:"-"`
	Annotations map[string]string `json:"annotations,omitempty" swaggertype:"string" valid:"-"`

	// 可选：从已有 PVC/快照恢复（具备能力时再放开）
	DataSource *DataSourceRef    `json:"dataSource,omitempty" valid:"-"`
	Selector   map[string]string `json:"selector,omitempty" valid:"-"`
}

// （如需要克隆/快照请解开注释）
type DataSourceRef struct {
	Kind     string `json:"kind,omitempty"` // PersistentVolumeClaim / VolumeSnapshot
	Name     string `json:"name,omitempty"`
	APIGroup string `json:"apiGroup,omitempty"` // "snapshot.storage.k8s.io" 等
}

func ValidKubePVCCreateRequest(data interface{}, _ *gin.Context) map[string][]string {
	req := data.(*KubePVCCreateRequest)
	errs := map[string][]string{}

	// 如果是 YAML 模式，只需要验证 YamlContent
	if req.YamlContent != "" {
		// YAML 模式，不需要验证其他字段
		return nil
	}
	
	// 表单模式验证
	if req.Namespace == "" {
		errs["namespace"] = append(errs["namespace"], "namespace 不能为空")
	}
	if req.Name == "" {
		errs["name"] = append(errs["name"], "name 不能为空")
	}
	if req.Storage == "" {
		errs["storage"] = append(errs["storage"], "storage 不能为空")
	}
	// AccessMode 和 AccessModes 至少有一个
	if req.AccessMode == "" && len(req.AccessModes) == 0 {
		errs["accessMode"] = append(errs["accessMode"], "accessMode 或 accessModes 至少指定一个")
	}

	// 这里不强制校验 StorageClassName；为空时走默认或无 SC 绑定

	if len(errs) > 0 {
		return errs
	}
	return nil
}

//
// ================= PersistentVolumeClaim 列表 =================
//

func NewKubePVCListRequest() *KubePVCListRequest { return &KubePVCListRequest{} }

type KubePVCListRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"-"`
	Name      string `json:"name"      form:"name"      valid:"name"` // 可选模糊匹配
	Phase     string `json:"phase"     form:"phase"     valid:"-"`    // Pending / Bound（可选）
	Page      int    `json:"page"      form:"page"      valid:"page"`
	Limit     int    `json:"limit"     form:"limit"     valid:"limit"`
}

func ValidKubePVCListRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"page":  []string{"required", "numeric", "min:1"},
		"limit": []string{"required", "numeric", "min:1", "max:1000"},
	}
	msgs := govalidator.MapData{
		"page": {
			"required:页码必填",
			"numeric:页码必须为数字",
			"min:页码不能小于 1",
		},
		"limit": {
			"required:每页数量必填",
			"numeric:每页数量必须为数字",
			"min:每页数量不能小于 1",
			"max:每页数量不能超过 1000",
		},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

//
// ================= PVC 详情 / 删除 =================
//

func NewKubePVCDetailRequest() *KubePVCDetailRequest { return &KubePVCDetailRequest{} }

type KubePVCDetailRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	Name      string `json:"name"      form:"name"      valid:"name"`
}

func ValidKubePVCDetailRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
	})
}

func NewKubePVCDeleteRequest() *KubePVCDeleteRequest { return &KubePVCDeleteRequest{} }

type KubePVCDeleteRequest struct {
	Namespace         string `json:"namespace"         form:"namespace"         valid:"namespace"`
	Name              string `json:"name"              form:"name"              valid:"name"`
	PropagationPolicy string `json:"propagationPolicy" form:"propagationPolicy" valid:"-"` // Foreground/Background/Orphan
}

func ValidKubePVCDeleteRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
	})
}

//
// ================= PVC 更新（PATCH） =================
//

func NewKubePVCUpdateRequest() *KubePVCUpdateRequest { return &KubePVCUpdateRequest{} }

type KubePVCUpdateRequest struct {
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name"      valid:"name"`
	// PatchType 可选: application/strategic-merge-patch+json | application/merge-patch+json | application/json-patch+json
	PatchType string `json:"patchType" valid:"-"`

	// Content: 原样 JSON 字符串（由前端构造）
	Content string `json:"content"  valid:"content"`
}

func ValidKubePVCUpdateRequest(data interface{}, _ *gin.Context) map[string][]string {
	req := data.(*KubePVCUpdateRequest)
	errs := valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
		"content":   {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
		"content":   {"required: content 不能为空"},
	})

	// 可选：快速校验 content 是否为合法 JSON（仅在 merge/strategic 时有意义）
	if req.Content != "" {
		var js map[string]any
		if e := json.Unmarshal([]byte(req.Content), &js); e != nil {
			if errs == nil {
				errs = map[string][]string{}
			}
			errs["content"] = append(errs["content"], "content 必须为合法 JSON")
		}
	}
	return errs
}

//
// ================= PVC 扩容（只增不减） =================
//

func NewKubePVCResizeRequest() *KubePVCResizeRequest { return &KubePVCResizeRequest{} }

type KubePVCResizeRequest struct {
	Namespace string `json:"namespace" valid:"namespace"`
	Name      string `json:"name"      valid:"name"`
	Storage   string `json:"storage"   valid:"storage"` // e.g. "20Gi"
}

func ValidKubePVCResizeRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"namespace": {"required"},
		"name":      {"required"},
		"storage":   {"required"},
	}, govalidator.MapData{
		"namespace": {"required: namespace 不能为空"},
		"name":      {"required: name 不能为空"},
		"storage":   {"required: storage 不能为空（例如 20Gi）"},
	})
}
