package secret

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"

	"k8soperation/internal/app/requests"
)

// --- 供前端下拉或后端统一常量 ---
const (
	ModeOpaque    = "opaque"
	ModeTLS       = "tls"
	ModeDocker    = "docker"
	ModeBasicAuth = "basic-auth"
	ModeSSHAuth   = "ssh-auth"
)

// BuildSecretFromReq 根据请求构建 Secret 对象（首选 StringData）
func BuildSecretFromReq(req *requests.KubeSecretCreateRequest) (*corev1.Secret, error) {
	// 1) 规范化类型（优先用 mode，兼容旧版 type）
	typ, err := normalizeSecretType(req.Mode, req.Type)
	if err != nil {
		return nil, err
	}
	// 冲突检测（可选）：如果 mode+type 同时提供但不匹配，直接报错
	if req.Mode != "" && req.Type != "" {
		stdTypByMode, _ := normalizeSecretType(req.Mode, "")
		stdTypByType, _ := normalizeSecretType("", req.Type)
		if stdTypByMode != stdTypByType {
			return nil, fmt.Errorf("mode %q conflicts with type %q", req.Mode, req.Type)
		}
	}

	// 2) 基本对象
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        req.Name,
			Namespace:   req.Namespace,
			Labels:      req.Labels,
			Annotations: req.Annotations,
		},
		Type:       typ,
		StringData: map[string]string{}, // 明文 → 由 APIServer 写入 data(base64)
	}

	// 3) 分类型填充/校验
	switch typ {

	case corev1.SecretTypeOpaque:
		// 优先使用 StringData，其次 Data
		if len(req.StringData) == 0 && len(req.Data) == 0 {
			return nil, fmt.Errorf("opaque secret requires non-empty data or string_data")
		}

		// 合并 string_data + data
		for k, v := range req.StringData {
			sec.StringData[k] = v
		}
		for k, v := range req.Data {
			sec.StringData[k] = v
		}

	case corev1.SecretTypeTLS:
		var crt, key string

		// 优先结构体
		if req.TLS != nil {
			crt = req.TLS.Cert
			key = req.TLS.Key
		}

		// 如果结构体为空，再看 string_data / data
		if crt == "" {
			if v, ok := req.StringData["tls.crt"]; ok {
				crt = v
			}
		}
		if key == "" {
			if v, ok := req.StringData["tls.key"]; ok {
				key = v
			}
		}

		// 最后再判断
		if crt == "" || key == "" {
			return nil, fmt.Errorf("tls secret requires tls.crt and tls.key")
		}

		sec.StringData["tls.crt"] = crt
		sec.StringData["tls.key"] = key

	case corev1.SecretTypeDockerConfigJson:
		var cfg string

		// 方式 A：string_data[".dockerconfigjson"]
		if v, ok := req.StringData[".dockerconfigjson"]; ok && strings.TrimSpace(v) != "" {
			cfg = v
		}

		// 方式 B：data[".dockerconfigjson"]（如果上面没取到）
		if cfg == "" {
			if v := strings.TrimSpace(req.Data[".dockerconfigjson"]); v != "" {
				cfg = v
			}
		}

		// 方式 C：由 docker{server,username,password} 组装（如果前两种都没有）
		if cfg == "" {
			if req.Docker != nil && req.Docker.Server != "" && req.Docker.Username != "" && req.Docker.Password != "" {
				j, err := buildDockerConfigJSON(req.Docker)
				if err != nil {
					return nil, err
				}
				cfg = j
			} else {
				return nil, fmt.Errorf("docker secret requires either data[.dockerconfigjson] or docker{server,username,password}")
			}
		}

		// 可选：校验 JSON 合法性，提前发现格式问题
		if !json.Valid([]byte(cfg)) {
			return nil, fmt.Errorf(".dockerconfigjson is not valid JSON")
		}

		// 用 StringData 交给 APIServer 编码一次即可
		sec.StringData[".dockerconfigjson"] = cfg

	case corev1.SecretTypeBasicAuth:
		var user, pass string

		// 方式 A：结构体字段（后端内部调用）
		if req.Basic != nil {
			user = strings.TrimSpace(req.Basic.Username)
			pass = req.Basic.Password // 密码不 trim，避免误伤空白字符
		}

		// 方式 B：string_data（前端/JSON 调用）
		if user == "" {
			if v, ok := req.StringData["username"]; ok && strings.TrimSpace(v) != "" {
				user = strings.TrimSpace(v)
			}
		}
		if pass == "" {
			if v, ok := req.StringData["password"]; ok && v != "" {
				pass = v
			}
		}

		// （可选）方式 C：data（如果你支持 data 为字符串；若是 []byte 请改成 string(b)）
		if user == "" {
			if v := strings.TrimSpace(req.Data["username"]); v != "" {
				user = v
			}
		}
		if pass == "" {
			if v := req.Data["password"]; v != "" {
				pass = v
			}
		}

		if user == "" || pass == "" {
			return nil, fmt.Errorf("basic-auth secret requires username and password")
		}

		// 最终写入 StringData，由 APIServer 编码到 .data
		sec.StringData["username"] = user
		sec.StringData["password"] = pass

	case corev1.SecretTypeSSHAuth:
		var pk, known string

		// 方式 A：结构体（后端内部调用）
		if req.SSH != nil {
			pk = req.SSH.PrivateKey    // 不要 TrimSpace，避免破坏换行/空格
			known = req.SSH.KnownHosts // 可选
		}

		// 方式 B：string_data（前端/JSON）
		if pk == "" {
			if v, ok := req.StringData["ssh-privatekey"]; ok && v != "" {
				pk = v
			}
		}
		if known == "" {
			if v, ok := req.StringData["known_hosts"]; ok && v != "" {
				known = v
			}
		}

		// 方式 C：data（如果你的 req.Data 是 map[string][]byte，则把 string(...) 换一下）
		if pk == "" {
			if v, ok := req.Data["ssh-privatekey"]; ok && v != "" { // 若是 []byte: len(v)>0 然后 pk = string(v)
				pk = v
			}
		}
		if known == "" {
			if v, ok := req.Data["known_hosts"]; ok && v != "" { // 若是 []byte: len(v)>0 然后 known = string(v)
				known = v
			}
		}

		if pk == "" {
			return nil, fmt.Errorf("ssh-auth secret requires private_key")
		}

		// 写入 StringData，让 APIServer 统一编码到 .data
		sec.StringData["ssh-privatekey"] = pk
		if known != "" {
			sec.StringData["known_hosts"] = known
		}

	}

	return sec, nil
}

// normalizeSecretType 将 mode / type 规范化为 corev1.SecretType
func normalizeSecretType(mode, typ string) (corev1.SecretType, error) {
	if mode != "" {
		switch strings.ToLower(strings.TrimSpace(mode)) {
		case ModeOpaque:
			return corev1.SecretTypeOpaque, nil
		case ModeTLS:
			return corev1.SecretTypeTLS, nil
		case ModeDocker:
			return corev1.SecretTypeDockerConfigJson, nil
		case ModeBasicAuth:
			return corev1.SecretTypeBasicAuth, nil
		case ModeSSHAuth:
			return corev1.SecretTypeSSHAuth, nil
		default:
			return "", fmt.Errorf("unsupported mode: %q", mode)
		}
	}
	// 兼容仅传 type
	switch strings.ToLower(strings.TrimSpace(typ)) {
	case "", "opaque":
		return corev1.SecretTypeOpaque, nil
	case "kubernetes.io/tls":
		return corev1.SecretTypeTLS, nil
	case "kubernetes.io/dockerconfigjson":
		return corev1.SecretTypeDockerConfigJson, nil
	case "kubernetes.io/basic-auth":
		return corev1.SecretTypeBasicAuth, nil
	case "kubernetes.io/ssh-auth":
		return corev1.SecretTypeSSHAuth, nil
	default:
		// 若你希望支持自定义类型，可放开：return corev1.SecretType(typ), nil
		return "", fmt.Errorf("unsupported secret type: %q", typ)
	}
}

// buildDockerConfigJSON 生成 .dockerconfigjson 内容
// 结构参考：docker login 认证文件
func buildDockerConfigJSON(d *requests.DockerFields) (string, error) {
	auth := d.Auth
	if auth == "" {
		auth = base64.StdEncoding.EncodeToString([]byte(d.Username + ":" + d.Password))
	}
	cfg := map[string]any{
		"auths": map[string]any{
			strings.TrimSpace(d.Server): map[string]string{
				"username": d.Username,
				"password": d.Password,
				"email":    d.Email,
				"auth":     auth,
			},
		},
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal docker config json failed: %w", err)
	}
	return string(b), nil
}

// MapKeys 返回 m 中所有键名的切片（顺序不固定）。
func MapKeys(m map[string][]byte) []string {
	// 预分配容量，减少 append 扩容
	keys := make([]string, 0, len(m))
	// 只取 key，不取 value
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
