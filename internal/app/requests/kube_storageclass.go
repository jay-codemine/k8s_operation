package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

//
// ================ StorageClass 创建 ================
//

func NewKubeStorageClassCreateRequest() *KubeStorageClassCreateRequest {
	return &KubeStorageClassCreateRequest{}
}

// 说明：StorageClass 是集群级资源【无 namespace】。
// 常用字段：
// - Provisioner: 必填，例如 "kubernetes.io/no-provisioner" 或 csi 驱动名 "ebs.csi.aws.com"
// - Parameters: 由各 provisioner 决定的键值对
// - ReclaimPolicy: "Delete" / "Retain"（可选）
// - VolumeBindingMode: "Immediate" / "WaitForFirstConsumer"（可选）
// - AllowVolumeExpansion: 是否允许扩容（可选）
// - MountOptions: 额外挂载选项（可选）
type KubeStorageClassCreateRequest struct {
	Name                 string            `json:"name"                  valid:"name"`
	Provisioner          string            `json:"provisioner"           valid:"provisioner"`
	Parameters           map[string]string `json:"parameters,omitempty"  swaggertype:"string" valid:"-"`
	ReclaimPolicy        string            `json:"reclaimPolicy"         valid:"-"` // Delete / Retain
	VolumeBindingMode    string            `json:"volumeBindingMode"     valid:"-"` // Immediate / WaitForFirstConsumer
	AllowVolumeExpansion *bool             `json:"allowVolumeExpansion"  valid:"-"`
	MountOptions         []string          `json:"mountOptions,omitempty" valid:"-"`
	// 备注：AllowedTopologies 较复杂，通常在控制台先不暴露；如需支持，可加 DTO 转换。
}

func ValidKubeStorageClassCreateRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name":        {"required"},
		"provisioner": {"required"},
	}, govalidator.MapData{
		"name":        {"required: name 不能为空"},
		"provisioner": {"required: provisioner 不能为空"},
	})
}

//
// ================ StorageClass 列表 ================
//

func NewKubeStorageClassListRequest() *KubeStorageClassListRequest {
	return &KubeStorageClassListRequest{}
}

type KubeStorageClassListRequest struct {
	// 集群级资源无需 namespace，这里仅做分页/名称过滤
	Page  int    `json:"page"  form:"page"  valid:"page"`
	Limit int    `json:"limit" form:"limit" valid:"limit"`
	Name  string `json:"name"  form:"name"  valid:"name"` // 可选：名称模糊过滤
}

func ValidKubeStorageClassListRequest(data interface{}, _ *gin.Context) map[string][]string {
	// 提供至少一条规则（不能为空，否则会 panic）
	rules := govalidator.MapData{
		"page":  []string{"required", "numeric", "min:1"},
		"limit": []string{"required", "numeric", "min:1", "max:1000"},
		"name":  []string{"max:100"},
	}

	messages := govalidator.MapData{
		"page": []string{
			"required:页码必填",
			"numeric:页码必须为数字",
			"min:页码不能小于 1",
		},
		"limit": []string{
			"required:每页数量必填",
			"numeric:每页数量必须为数字",
			"min:每页数量不能小于 1",
			"max:每页数量不能超过 1000",
		},
		"name": []string{
			"max:名称长度不能超过 100",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

//
// ================ StorageClass 详情 / 删除 ================
//

func NewKubeStorageClassDetailRequest() *KubeStorageClassDetailRequest {
	return &KubeStorageClassDetailRequest{}
}

type KubeStorageClassDetailRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
}

func ValidKubeStorageClassDetailRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

func NewKubeStorageClassDeleteRequest() *KubeStorageClassDeleteRequest {
	return &KubeStorageClassDeleteRequest{}
}

type KubeStorageClassDeleteRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
}

func ValidKubeStorageClassDeleteRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

//
// ================ StorageClass 名称列表（下拉用） ================
//

func NewKubeStorageClassNamesRequest() *KubeStorageClassNamesRequest {
	return &KubeStorageClassNamesRequest{}
}

// 集群级资源，无需任何参数；保留结构以便后续扩展（例如按 provisioner 过滤）
type KubeStorageClassNamesRequest struct {
	// 可选过滤：provisioner
	Provisioner string `json:"provisioner" form:"provisioner" valid:"-"`
}

func ValidKubeStorageClassNamesRequest(data interface{}, _ *gin.Context) map[string][]string {
	return nil
}
