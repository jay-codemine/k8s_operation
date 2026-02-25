package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/valid"
)

// =======================
// Create
// =======================

type K8sClusterCreateRequest struct {
	ClusterName    string `json:"cluster_name" form:"cluster_name" valid:"cluster_name"`
	ClusterVersion string `json:"cluster_version" form:"cluster_version" valid:"cluster_version"`
	KubeConfig     string `json:"kube_config" form:"kube_config" valid:"kube_config"`
}

func NewK8sClusterCreateRequest() *K8sClusterCreateRequest {
	return &K8sClusterCreateRequest{}
}

func ValidK8sClusterCreateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"cluster_name":    []string{"required"},
		"cluster_version": []string{"required"},
		"kube_config":     []string{"required"},
	}

	messages := govalidator.MapData{
		"cluster_name": []string{
			"required: 集群名称为必填字段(cluster_name)",
		},
		"cluster_version": []string{
			"required: 集群版本为必填字段(cluster_version)",
		},
		"kube_config": []string{
			"required: kubeconfig 为必填字段(kube_config)",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

// =======================
// Update
// =======================

type K8sClusterUpdateRequest struct {
	ID             uint32 `json:"id,omitempty" form:"id" valid:"id"`
	ClusterName    string `json:"cluster_name" form:"cluster_name" valid:"cluster_name"`
	ClusterVersion string `json:"cluster_version" form:"cluster_version" valid:"cluster_version"`
	KubeConfig     string `json:"kube_config,omitempty" form:"kube_config" valid:"kube_config"` // ✅ 可选（不填就不更新）
}

func NewK8sClusterUpdateRequest() *K8sClusterUpdateRequest {
	return &K8sClusterUpdateRequest{}
}

func ValidK8sClusterUpdateRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id":              []string{"required"},
		"cluster_name":    []string{"required"},
		"cluster_version": []string{"required"},
		"kube_config":     []string{"required"}, // 需要强制更新 kubeconfig 就打开这行
	}

	messages := govalidator.MapData{
		"id": []string{
			"required: 集群ID不能为空(id)",
		},
		"cluster_name": []string{
			"required: 集群名称为必填字段(cluster_name)",
		},
		"cluster_version": []string{
			"required: 集群版本为必填字段(cluster_version)",
		},
		"kube_config": []string{
			"required: kubeconfig 为必填字段(kube_config)",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

// =======================
// List
// =======================

type K8sClusterListRequest struct {
	ClusterName string `json:"cluster_name,omitempty" form:"cluster_name"`
	Page        int    `json:"page,omitempty" form:"page" valid:"page"`
	Limit       int    `json:"limit,omitempty" form:"limit" valid:"limit"`
}

func NewK8sClusterListRequest() *K8sClusterListRequest {
	return &K8sClusterListRequest{}
}

func ValidK8sClusterListRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"page":  []string{"required"},
		"limit": []string{"required"},
	}
	messages := govalidator.MapData{
		"page": []string{
			"required: 页码为必填项(page)",
		},
		"limit": []string{
			"required: 每页数量为必填项(limit)",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

// =======================
// Delete
// =======================

type K8sClusterDeleteRequest struct {
	ID    uint32 `json:"id,omitempty" form:"id" valid:"id"`
	Force bool   `json:"force" form:"force"` // 可选：是否强制删除
}

func NewK8sClusterDeleteRequest() *K8sClusterDeleteRequest {
	return &K8sClusterDeleteRequest{}
}

func ValidK8sClusterDeleteRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}

	messages := govalidator.MapData{
		"id": []string{
			"required: 集群ID不能为空(id)",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}

// =======================
// Init
// =======================

type K8sClusterInitRequest struct {
	ID uint32 `json:"id,omitempty" form:"id" valid:"id"`
}

func NewK8sClusterInitRequest() *K8sClusterInitRequest {
	return &K8sClusterInitRequest{}
}

func ValidK8sClusterInitRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"id": []string{"required"},
	}
	messages := govalidator.MapData{
		"id": []string{
			"required: 集群ID不能为空(id)",
		},
	}

	return valid.ValidateOptions(data, rules, messages)
}
