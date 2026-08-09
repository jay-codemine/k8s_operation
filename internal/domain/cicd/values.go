package cicd

import (
	"fmt"
	"strings"
)

// DeployEnv 部署环境（dev/staging/prod）
type DeployEnv struct{ val string }

var validEnvs = map[string]bool{"dev": true, "staging": true, "prod": true, "test": true}

func NewDeployEnv(s string) (DeployEnv, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return DeployEnv{}, fmt.Errorf("部署环境不能为空")
	}
	if !validEnvs[s] {
		return DeployEnv{}, fmt.Errorf("无效的部署环境: %s (允许: dev/staging/prod/test)", s)
	}
	return DeployEnv{val: s}, nil
}

func (e DeployEnv) String() string { return e.val }
func (e DeployEnv) IsProd() bool   { return e.val == "prod" }

// BuildNumber 构建编号
type BuildNumber struct{ val int }

func NewBuildNumber(n int) (BuildNumber, error) {
	if n < 0 {
		return BuildNumber{}, fmt.Errorf("构建编号不能为负数")
	}
	return BuildNumber{val: n}, nil
}

func (b BuildNumber) Int() int    { return b.val }
func (b BuildNumber) Next() BuildNumber { return BuildNumber{val: b.val + 1} }
