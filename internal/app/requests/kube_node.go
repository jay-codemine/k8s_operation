package requests

import (
	"github.com/gin-gonic/gin"
	"github.com/thedevsaddam/govalidator"
	corev1 "k8s.io/api/core/v1"
	"k8soperation/pkg/valid"
)

//
// ================= Node 列表 =================
//

func NewKubeNodeListRequest() *KubeNodeListRequest { return &KubeNodeListRequest{} }

type KubeNodeListRequest struct {
	// 名称模糊匹配
	Name string `json:"name" form:"name" valid:"name"`

	// Label 选择器（与 kubectl 一致，如 "node-role.kubernetes.io/worker=,env=prod"）
	LabelSelector string `json:"labelSelector" form:"labelSelector" valid:"-"`

	// 仅看不可调度节点（true）/ 仅看可调度（false）/ 不筛（为空）
	Unschedulable *bool `json:"unschedulable" form:"unschedulable" swaggertype:"boolean" valid:"-"`

	Page  int `json:"page"  form:"page"  valid:"page"`
	Limit int `json:"limit" form:"limit" valid:"limit"`
}

func ValidKubeNodeListRequest(data interface{}, _ *gin.Context) map[string][]string {
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
// ================= Node 详情 =================
//

func NewKubeNodeDetailRequest() *KubeNodeDetailRequest { return &KubeNodeDetailRequest{} }

type KubeNodeDetailRequest struct {
	Name string `json:"name" form:"name" valid:"name"`
}

func ValidKubeNodeDetailRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

//
// ================= Node 删除 =================
//

func NewKubeNodeDeleteRequest() *KubeNodeDeleteRequest { return &KubeNodeDeleteRequest{} }

type KubeNodeDeleteRequest struct {
	// 直接删除 Node 资源（注意：不会自动驱逐 Pod；请结合 drain）
	Name string `json:"name" form:"name" valid:"name"`
}

func ValidKubeNodeDeleteRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

//
// ================= Node 通用 PATCH =================
//

func NewKubeNodePatchRequest() *KubeNodePatchRequest { return &KubeNodePatchRequest{} }

type KubeNodePatchRequest struct {
	Name string `json:"name" valid:"name"`
	// PatchType:
	//   application/strategic-merge-patch+json
	//   application/merge-patch+json
	//   application/json-patch+json
	PatchType string `json:"patchType" valid:"-"`

	// Content: 原样 JSON 字符串（由前端构造）
	Content string `json:"content" valid:"content"`
}

func ValidKubeNodePatchRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name":    {"required"},
		"content": {"required"},
	}, govalidator.MapData{
		"name":    {"required: name 不能为空"},
		"content": {"required: content 不能为空"},
	})
}

//
// ================= Node Cordon / Uncordon =================
//

func NewKubeNodeCordonRequest() *KubeNodeCordonRequest { return &KubeNodeCordonRequest{} }

type KubeNodeCordonRequest struct {
	NodeName      string `json:"nodeName" valid:"nodeName"`
	Unschedulable bool   `json:"unschedulable" valid:"unschedulable"`
}

func ValidKubeNodeCordonRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"nodeName": {"required"},
	}, govalidator.MapData{
		"nodeName": {"required: name 不能为空"},
	})
}

func NewKubeNodeUncordonRequest() *KubeNodeUncordonRequest { return &KubeNodeUncordonRequest{} }

type KubeNodeUncordonRequest struct {
	Name     string `json:"name" valid:"name"`
	NodeName string `json:"nodeName" valid:"nodeName"`
}

func ValidKubeNodeUncordonRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

// ================= Node Drain（驱逐工作负载） =================
func NewKubeNodeDrainRequest() *KubeNodeDrainRequest {
	return &KubeNodeDrainRequest{
		GracePeriodSeconds: -1, // -1 = 使用 Pod 自己的 terminationGracePeriodSeconds
		TimeoutSeconds:     0,  // 0  = 不做超时控制，由上层自己控制 ctx 超时
		DeleteEmptyDir:     false,
		IgnoreDaemonSets:   true, // 默认跳过 DaemonSet Pod（和 kubectl drain 一致）
		Force:              false,
		DryRun:             false,
	}
}

// Node Drain（驱逐工作负载）请求
type KubeNodeDrainRequest struct {
	// 要 drain 的节点名（必填）
	NodeName string `json:"nodeName" valid:"nodeName"`

	// Pod 的优雅退出时间，单位秒：
	//   -1 = 使用 Pod 自己的 terminationGracePeriodSeconds
	//    0 = 立刻删除（不推荐）
	//  > 0 = 指定宽限期
	GracePeriodSeconds int32 `json:"gracePeriodSeconds" swaggertype:"integer"`

	// 整个 drain 操作的超时时间（秒），0 表示不超时（可交给 HTTP ctx 控制）
	TimeoutSeconds int `json:"timeoutSeconds" swaggertype:"integer"`

	// 是否删除使用 emptyDir 的 Pod（等价于 kubectl drain 的 --delete-emptydir-data）
	DeleteEmptyDir bool `json:"deleteEmptyDir"`

	// 是否忽略 DaemonSet Pod（等价于 --ignore-daemonsets）
	IgnoreDaemonSets bool `json:"ignoreDaemonSets" default:"true"`

	// 是否强制驱逐（等价于 --force），比如没有 PDB / 本地存储等场景
	Force bool `json:"force"`

	// 仅匹配特定 Label 的 Pod，格式与 labelSelector 相同：
	//   例如： "app=myapp,env=prod"
	PodSelector string `json:"podSelector"`

	// 仅打印将要执行的动作，而不真正执行（DryRun 模式）
	DryRun bool `json:"dryRun"`
}

func ValidKubeNodeDrainRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"nodeName":           {"required"},
		"gracePeriodSeconds": {"min:-1"}, // 允许 -1，表示使用默认
		"timeoutSeconds":     {"min:0"},  // 允许 0，表示不超时
	}
	msgs := govalidator.MapData{
		"nodeName":           {"required: nodeName 不能为空"},
		"gracePeriodSeconds": {"min: gracePeriodSeconds 不能小于 -1"},
		"timeoutSeconds":     {"min: timeoutSeconds 不能小于 0"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

//
// ================= Node 标签 / 注解 / 污点 变更 =================
//

func NewKubeNodeLabelPatchRequest() *KubeNodeLabelPatchRequest { return &KubeNodeLabelPatchRequest{} }

type KubeNodeLabelPatchRequest struct {
	Name   string            `json:"name" valid:"name"`
	Add    map[string]string `json:"add"    swaggertype:"string"`
	Remove []string          `json:"remove"`
}

func ValidKubeNodeLabelPatchRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

func NewKubeNodeAnnotationPatchRequest() *KubeNodeAnnotationPatchRequest {
	return &KubeNodeAnnotationPatchRequest{}
}

type KubeNodeAnnotationPatchRequest struct {
	Name   string            `json:"name" valid:"name"`
	Add    map[string]string `json:"add"    swaggertype:"string"`
	Remove []string          `json:"remove"`
}

func ValidKubeNodeAnnotationPatchRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

func NewKubeNodeTaintPatchRequest() *KubeNodeTaintPatchRequest { return &KubeNodeTaintPatchRequest{} }

type KubeNodeTaintPatchRequest struct {
	Name       string         `json:"name" valid:"name"`
	Add        []corev1.Taint `json:"add"`        // 按 key+effect 覆盖/新增
	RemoveKeys []string       `json:"removeKeys"` // 按 key 删除（同 effect 一并移除）
}

func ValidKubeNodeTaintPatchRequest(data interface{}, _ *gin.Context) map[string][]string {
	return valid.ValidateOptions(data, govalidator.MapData{
		"name": {"required"},
	}, govalidator.MapData{
		"name": {"required: name 不能为空"},
	})
}

//
// ================= Node 附属查询：Pods / Events / Metrics =================
//

func NewKubeNodePodsRequest() *KubeNodePodsRequest { return &KubeNodePodsRequest{} }

type KubeNodePodsRequest struct {
	Name       string `json:"name"       form:"name"       valid:"name"`
	Namespace  string `json:"namespace"  form:"namespace"  valid:"-"` // 可选：仅查看某 ns
	Phase      string `json:"phase"      form:"phase"      valid:"-"` // Running/Pending/Failed/Succeeded…
	OwnerKinds string `json:"ownerKinds" form:"ownerKinds" valid:"-"` // 逗号分隔: Deployment,StatefulSet,DaemonSet,...
	Page       int    `json:"page"       form:"page"       valid:"page"`
	Limit      int    `json:"limit"      form:"limit"      valid:"limit"`
}

func ValidKubeNodePodsRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":  {"required"},
		"page":  {"required", "numeric", "min:1"},
		"limit": {"required", "numeric", "min:1", "max:1000"},
	}
	msgs := govalidator.MapData{
		"name":  {"required: name 不能为空"},
		"page":  {"required:页码必填", "numeric:页码必须为数字", "min:页码不能小于 1"},
		"limit": {"required:每页数量必填", "numeric:每页数量必须为数字", "min:每页数量不能小于 1", "max:每页数量不能超过 1000"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

func NewKubeNodeEventsRequest() *KubeNodeEventsRequest {
	return &KubeNodeEventsRequest{
		Limit: 50, // 默认50条
	}
}

type KubeNodeEventsRequest struct {
	Name  string `json:"name"  form:"name"  valid:"name"`
	Limit int    `json:"limit" form:"limit" valid:"limit"`
}

func ValidKubeNodeEventsRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":  {"required"},
		"limit": {"min:1", "max:1000"},
	}
	msgs := govalidator.MapData{
		"name":  {"required: name 不能为空"},
		"limit": {"min:每页数量不能小于 1", "max:每页数量不能超过 1000"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

type KubeNodeMetricsRequest struct {
	Name  string `form:"name" json:"name" valid:"name"` // 节点名，可选（空=全量）
	Page  int    `form:"page"     json:"page" valid:"page"`
	Limit int    `form:"limit"    json:"limit" valid:"limit"`
}

func NewKubeNodeMetricsRequest() *KubeNodeMetricsRequest {
	return &KubeNodeMetricsRequest{
		Page:  1,
		Limit: 20,
	}
}

func ValidKubeNodeMetricsRequest(data interface{}, ctx *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"name":  []string{}, // 可选，空=全量
		"page":  []string{"min:1"},
		"limit": []string{"min:1", "max:500"},
	}
	msgs := govalidator.MapData{
		"name":  []string{},
		"page":  []string{"min: page 必须 >= 1"},
		"limit": []string{"min: limit 必须 >= 1", "max: limit 不能超过 500"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}

// KubePodEvictRequest 指定 Pod 驱逐请求参数
type KubePodEvictRequest struct {
	Namespace string `json:"namespace" form:"namespace" valid:"namespace"`
	PodName   string `json:"podName"   form:"podName"   valid:"podName"`
}

func NewKubePodEvictRequest() *KubePodEvictRequest {
	return &KubePodEvictRequest{}
}

func ValidKubePodEvictRequest(data interface{}, _ *gin.Context) map[string][]string {
	rules := govalidator.MapData{
		"namespace": []string{"required"},
		"podName":   []string{"required"},
	}
	msgs := govalidator.MapData{
		"namespace": []string{"required: namespace 不能为空"},
		"podName":   []string{"required: podName 不能为空"},
	}
	return valid.ValidateOptions(data, rules, msgs)
}
