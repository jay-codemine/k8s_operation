package requests

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// AppConfigSpec 定义期望状态
type AppConfigSpec struct {
	AppName  string            `json:"appName"`
	Env      string            `json:"env,omitempty"`
	Image    string            `json:"image"`
	Replicas *int32            `json:"replicas"`
	Port     int32             `json:"port,omitempty"`
	Config   map[string]string `json:"config,omitempty"`
}

// AppConfigStatus 定义实际状态
type AppConfigStatus struct {
	Phase          string      `json:"phase,omitempty"`
	Message        string      `json:"message,omitempty"`
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// AppConfig 是 CRD 根对象
type AppConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppConfigSpec   `json:"spec,omitempty"`
	Status AppConfigStatus `json:"status,omitempty"`
}

// AppConfigList 列表
type AppConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppConfig `json:"items"`
}
