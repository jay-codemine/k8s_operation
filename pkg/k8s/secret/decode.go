package secret

import (
	"context"
	"encoding/base64"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8soperation/internal/app/requests"
	"strings"
)

func DecodeSecretData(ctx context.Context, _ kubernetes.Interface, req *requests.KubeSecretDecodeRequest) (map[string]string, error) {
	result := make(map[string]string)

	// 先批量解码
	if len(req.Data) > 0 {
		for key, encoded := range req.Data {
			if strings.TrimSpace(encoded) == "" {
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				result[key] = fmt.Sprintf("[decode error] %v", err)
				continue
			}
			result[key] = string(decoded)
		}
	}

	// 再单值解码（如果有）
	if req.Value != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.Value)
		if err != nil {
			return result, fmt.Errorf("decode single value failed: %v", err)
		}
		result["decoded"] = string(decoded)
	}

	// 最后检查是否真的有内容
	if len(result) == 0 {
		return nil, fmt.Errorf("必须提供 data 或 value 参数")
	}

	return result, nil
}
