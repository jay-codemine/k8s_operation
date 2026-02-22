package utils

func StripKeys(m map[string]string, keys []string) map[string]string {
	if m == nil {
		return map[string]string{}
	}

	for _, key := range keys {
		delete(m, key) // 删除键及对于值
	}

	return m
}
