package license

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DefaultLicenseFile License 文件默认存放路径（相对工作目录）
// 可通过环境变量 K8SOP_LICENSE_FILE 覆盖
const DefaultLicenseFile = "storage/license.lic"

// Status 当前授权状态（供接口返回与中间件判断）
type Status struct {
	Licensed  bool   `json:"licensed"`   // 是否已授权且有效
	Reason    string `json:"reason"`     // 未授权原因（未激活/已过期/机器码不匹配等）
	Licensee  string `json:"licensee"`   // 被授权方
	Edition   string `json:"edition"`    // 版本
	ExpireAt  int64  `json:"expire_at"`  // 到期时间（unix 秒）
	IssuedAt  int64  `json:"issued_at"`  // 签发时间（unix 秒）
	MachineID string `json:"machine_id"` // 本机机器码（用于向供应商申请 License）
}

type manager struct {
	mu        sync.RWMutex
	payload   *Payload // 最近一次校验通过的载荷（nil = 未授权）
	reason    string   // 未授权原因
	machineID string   // 本机机器码（进程内缓存）
}

var mgr = &manager{}

// Init 启动时调用：计算机器码并尝试加载本地 License 文件
func Init() {
	mgr.mu.Lock()
	mgr.machineID = MachineID()
	mgr.mu.Unlock()
	reload()
}

// licenseFilePath License 文件路径
func licenseFilePath() string {
	if v := strings.TrimSpace(os.Getenv("K8SOP_LICENSE_FILE")); v != "" {
		return v
	}
	return DefaultLicenseFile
}

// reload 从磁盘读取并校验 License，刷新内存状态
func reload() {
	data, err := os.ReadFile(licenseFilePath())
	if err != nil {
		mgr.setInvalid("未激活授权")
		return
	}
	p, err := Verify(string(data), CurrentMachineID())
	if err != nil {
		mgr.setInvalid(err.Error())
		return
	}
	mgr.mu.Lock()
	mgr.payload = p
	mgr.reason = ""
	mgr.mu.Unlock()
}

func (m *manager) setInvalid(reason string) {
	m.mu.Lock()
	m.payload = nil
	m.reason = reason
	m.mu.Unlock()
}

// CurrentMachineID 返回本机机器码
func CurrentMachineID() string {
	mgr.mu.RLock()
	id := mgr.machineID
	mgr.mu.RUnlock()
	if id == "" {
		id = MachineID()
		mgr.mu.Lock()
		mgr.machineID = id
		mgr.mu.Unlock()
	}
	return id
}

// Valid 当前授权是否有效（中间件每个请求调用，需高性能）
// 内存校验 + 实时过期判断，不读磁盘
func Valid() (bool, string) {
	mgr.mu.RLock()
	p, reason := mgr.payload, mgr.reason
	mgr.mu.RUnlock()

	if p == nil {
		if reason == "" {
			reason = "未激活授权"
		}
		return false, reason
	}
	// 实时过期判断（授权到期瞬间生效，无需重启）
	if p.ExpireAt > 0 && time.Now().Unix() > p.ExpireAt {
		return false, ErrExpired.Error()
	}
	return true, ""
}

// GetStatus 返回完整授权状态（激活页/状态接口使用）
func GetStatus() Status {
	st := Status{MachineID: CurrentMachineID()}

	mgr.mu.RLock()
	p, reason := mgr.payload, mgr.reason
	mgr.mu.RUnlock()

	if p == nil {
		st.Reason = reason
		if st.Reason == "" {
			st.Reason = "未激活授权"
		}
		return st
	}

	st.Licensee = p.Licensee
	st.Edition = p.Edition
	st.ExpireAt = p.ExpireAt
	st.IssuedAt = p.IssuedAt

	if p.ExpireAt > 0 && time.Now().Unix() > p.ExpireAt {
		st.Licensed = false
		st.Reason = ErrExpired.Error()
	} else {
		st.Licensed = true
	}
	return st
}

// Activate 校验并持久化新 License（激活接口调用）
// 校验通过后写入磁盘并刷新内存状态
func Activate(text string) (*Payload, error) {
	text = strings.TrimSpace(text)
	p, err := Verify(text, CurrentMachineID())
	if err != nil {
		return nil, err
	}

	path := licenseFilePath()
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	if err := os.WriteFile(path, []byte(text+"\n"), 0o644); err != nil {
		return nil, err
	}

	mgr.mu.Lock()
	mgr.payload = p
	mgr.reason = ""
	mgr.mu.Unlock()
	return p, nil
}
