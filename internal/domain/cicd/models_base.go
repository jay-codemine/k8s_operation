package cicd

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// EnvVar 环境变量结构
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EnvVars JSON数组类型
type EnvVars []EnvVar

func (e EnvVars) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	return json.Marshal(e)
}

func (e *EnvVars) Scan(value interface{}) error {
	if value == nil {
		*e = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, e)
}

// DeployConfig 部署配置结构
type DeployConfig struct {
	Namespace      string `json:"namespace"`
	DeploymentName string `json:"deployment_name"`
	Image          string `json:"image"`
	Replicas       int    `json:"replicas"`
	Strategy       string `json:"strategy"`
}

// JSONMap 通用JSON Map类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// JSONArray 通用JSON数组类型
type JSONArray []interface{}

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}
