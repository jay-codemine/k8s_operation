package license

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
)

// MachineID 生成本机机器码（确定性，格式 XXXX-XXXX-XXXX-XXXX）
//
// 取值优先级：
//  1. 环境变量 K8SOP_MACHINE_ID —— 容器/K8s 环境下 Pod 网卡 MAC 每次重启会变化，
//     部署时应在宿主机执行 `k8soperation.exe -machine-id`（或 license-gen 提供的方式）
//     获取机器码后，通过环境变量固定注入
//  2. /etc/machine-id —— Linux 物理机/虚拟机（systemd 生成，重装系统前不变）
//  3. 物理网卡 MAC 地址集合哈希 —— 兜底方案（Windows 开发机等）
func MachineID() string {
	// 1) 环境变量显式指定（容器场景）
	if v := strings.TrimSpace(os.Getenv("K8SOP_MACHINE_ID")); v != "" {
		return strings.ToUpper(v)
	}

	// 2) Linux machine-id
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		if id := strings.TrimSpace(string(data)); id != "" {
			return formatMachineID([]byte("machine-id:" + id))
		}
	}

	// 3) 物理网卡 MAC 哈希
	if macs := collectMACs(); len(macs) > 0 {
		return formatMachineID([]byte("macs:" + strings.Join(macs, ",")))
	}

	// 极端兜底：主机名（尽量避免走到这里）
	host, _ := os.Hostname()
	return formatMachineID([]byte("host:" + host))
}

// collectMACs 收集非回环、非虚拟网卡的 MAC 地址（排序去重，保证确定性）
func collectMACs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var macs []string
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagLoopback != 0 || len(ifc.HardwareAddr) == 0 {
			continue
		}
		name := strings.ToLower(ifc.Name)
		// 过滤常见虚拟网卡（docker/veth/vEthernet/tun 等），降低机器码抖动
		if strings.Contains(name, "docker") || strings.Contains(name, "veth") ||
			strings.Contains(name, "vethernet") || strings.Contains(name, "tun") ||
			strings.Contains(name, "br-") || strings.Contains(name, "flannel") ||
			strings.Contains(name, "cni") || strings.Contains(name, "wsl") {
			continue
		}
		mac := ifc.HardwareAddr.String()
		if mac != "" && !seen[mac] {
			seen[mac] = true
			macs = append(macs, mac)
		}
	}
	sort.Strings(macs)
	return macs
}

// formatMachineID 对源数据哈希后格式化为 XXXX-XXXX-XXXX-XXXX
func formatMachineID(src []byte) string {
	sum := sha256.Sum256(src)
	hexStr := strings.ToUpper(fmt.Sprintf("%x", sum[:8])) // 取前 8 字节 = 16 个 hex 字符
	return fmt.Sprintf("%s-%s-%s-%s", hexStr[0:4], hexStr[4:8], hexStr[8:12], hexStr[12:16])
}
