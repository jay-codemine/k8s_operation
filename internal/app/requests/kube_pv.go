package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/pkg/valid"
)

//
// ================= PersistentVolume 创建 =================
//

// PV 支持的常见字段：
// - Capacity: 存储大小（如 "10Gi"）
// - AccessModes: ["ReadWriteOnce", "ReadOnlyMany", "ReadWriteMany"]
// - ReclaimPolicy: "Delete" / "Retain"
// - StorageClassName: 关联 StorageClass
// - VolumeMode: "Filesystem" / "Block"
// - Source: HostPath / NFS / CSI / 其他存储后端
func NewKubePVCreateRequest() *KubePVCreateRequest {
	return &KubePVCreateRequest{}
}

type KubePVCreateRequest struct {
	Name             string                              `json:"name"               valid:"name"`
	Capacity         string                              `json:"capacity"           valid:"required"` // e.g. 10Gi
	AccessModes      []corev1.PersistentVolumeAccessMode `json:"accessModes"  valid:"required"`
	ReclaimPolicy    string                              `json:"reclaimPolicy"      valid:"-"` // Delete / Retain
	StorageClassName string                              `json:"storageClassName"   valid:"-"`
	VolumeMode       *corev1.PersistentVolumeMode        `json:"volumeMode,omitempty" valid:"-"`

	Labels      map[string]string `json:"labels,omitempty"      swaggertype:"string" valid:"-"`
	Annotations map[string]string `json:"annotations,omitempty" swaggertype:"string" valid:"-"`

	// 常见后端定义
	HostPath string           `json:"hostPath,omitempty"` // e.g. "/data/pv1"
	NFS      *NFSVolumeConfig `json:"nfs,omitempty"`      // 若使用 NFS，填写 Server / Path
}

// 用于 NFS 配置的结构体
type NFSVolumeConfig struct {
	Server   string `json:"server"`
	Path     string `json:"path"`
	ReadOnly bool   `json:"readOnly"`
}

func ValidKubePVCreateRequest(data interface{}, _ *gin.Context) map[string][]string {
	req := data.(*KubePVCreateRequest)
	errs := map[string][]string{}

	if req.Name == "" {
		errs["name"] = append(errs["name"], "name 不能为空")
	}
	if req.Capacity == "" {
		errs["capacity"] = append(errs["capacity"], "capacity 不能为空")
	}
	if len(req.AccessModes) == 0 {
		errs["accessModes"] = append(errs["accessModes"], "accessModes 至少指定一个")
	}
	if req.HostPath == "" && req.NFS == nil {
		errs["source"] = append(errs["source"], "必须指定 HostPath 或 NFS 存储源")
	}
	if req.NFS != nil {
		if req.NFS.Server == "" || req.NFS.Path == "" {
			errs["nfs"] = append(errs["nfs"], "NFS 模式下 server 与 path 均不能为空")
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

//
// ================= PersistentVolume 列表 =================
//

func NewKubePVListRequest() *KubePVListRequest { return &KubePVListRequest{} }

type KubePVListRequest struct {
	Page  int    `json:"page"  form:"page"  valid:"page"`
	Limit int    `json:"limit" form:"limit" valid:"limit"`
	Name  string `json:"name"  form:"name"  valid:"name"` // 可选模糊匹配
}

func ValidKubePVListRequest(data interface{}, _ *gin.Context) map[string][]string {
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
// ================= PersistentVolume 详情 / 删除 =================
//

func NewKubePVDetailRequest() *KubePVDetailRequest { return &KubePVDetailRequest{} }

type KubePVDetailRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
}

func ValidKubePVDetailRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

func NewKubePVDeleteRequest() *KubePVDeleteRequest { return &KubePVDeleteRequest{} }

type KubePVDeleteRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
}

func ValidKubePVDeleteRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

//
// ================= PersistentVolume 更新 =================
//

func NewKubePVUpdateRequest() *KubePVUpdateRequest { return &KubePVUpdateRequest{} }

type KubePVUpdateRequest struct {
	Name    string `json:"name"    valid:"name"`
	Content string `json:"content" valid:"content"` // JSON/YAML 字符串
}

func ValidKubePVUpdateRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name":    {"required"},
		"content": {"required"},
	}, govalidator.MapData{
		"name":    {"required: name 不能为空"},
		"content": {"required: content 不能为空"},
	})
}

//
// ================= 修改回收策略（ReclaimPolicy） =================
//

func NewKubePVReclaimRequest() *KubePVReclaimRequest { return &KubePVReclaimRequest{} }

type KubePVReclaimRequest struct {
	Name          string `json:"name"          valid:"name"`
	ReclaimPolicy string `json:"reclaimPolicy" valid:"reclaimPolicy"` // Delete / Retain
}

func ValidKubePVReclaimRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name":          {"required"},
		"reclaimPolicy": {"required"},
	}, govalidator.MapData{
		"name":          {"required: name 不能为空"},
		"reclaimPolicy": {"required: reclaimPolicy 不能为空（Delete 或 Retain）"},
	})
}

//
// ================= PV 扩容 =================
//

func NewKubePVExpandRequest() *KubePVExpandRequest { return &KubePVExpandRequest{} }

type KubePVExpandRequest struct {
	Name        string `json:"name"        form:"name"        valid:"name"`
	NewCapacity string `json:"newCapacity" valid:"newCapacity"` // 例如: 20Gi
}

func ValidKubePVExpandRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name":        {"required"},
		"newCapacity": {"required"},
	}, govalidator.MapData{
		"name":        {"required: name 不能为空"},
		"newCapacity": {"required: newCapacity 不能为空（例如 20Gi）"},
	})
}

//
// ================= YAML 操作 =================
//

// GetYaml 请求参数
func NewKubePVGetYamlRequest() *KubePVGetYamlRequest { return &KubePVGetYamlRequest{} }

type KubePVGetYamlRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
}

func ValidKubePVGetYamlRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

// ApplyYaml 请求参数
func NewKubePVApplyYamlRequest() *KubePVApplyYamlRequest { return &KubePVApplyYamlRequest{} }

type KubePVApplyYamlRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
	Yaml string `json:"yaml" valid:"yaml"`
}

func ValidKubePVApplyYamlRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
		"yaml": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
		"yaml": {"required: yaml 内容不能为空"},
	})
}

// CreateFromYaml 请求参数
func NewKubePVCreateFromYamlRequest() *KubePVCreateFromYamlRequest {
	return &KubePVCreateFromYamlRequest{}
}

type KubePVCreateFromYamlRequest struct {
	Yaml string `json:"yaml" valid:"yaml"`
}

func ValidKubePVCreateFromYamlRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"yaml": {"required"},
	}, govalidator.MapData{
		"yaml": {"required: yaml 内容不能为空"},
	})
}
