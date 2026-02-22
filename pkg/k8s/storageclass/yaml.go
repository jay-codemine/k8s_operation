package storageclass

import (
	"context"
	"fmt"
	"strings"

	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	"k8soperation/global"
)

// GetStorageClassYaml 获取 StorageClass 的 YAML 表示
func GetStorageClassYaml(ctx context.Context, Kube kubernetes.Interface, name string) (string, error) {
	sc, err := Kube.StorageV1().
		StorageClasses().
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// 清理不必要的元数据
	sc.ManagedFields = nil
	sc.ResourceVersion = ""
	sc.UID = ""
	sc.SelfLink = ""
	sc.Generation = 0

	yamlData, err := yaml.Marshal(sc)
	if err != nil {
		return "", fmt.Errorf("序列化 YAML 失败: %w", err)
	}

	return string(yamlData), nil
}

// CreateStorageClassFromYaml 从 YAML 创建 StorageClass（支持多资源）
func CreateStorageClassFromYaml(ctx context.Context, Kube kubernetes.Interface, yamlContent string) (*storagev1.StorageClass, error) {
	// 分割多文档 YAML
	docs := strings.Split(yamlContent, "---")

	var lastStorageClass *storagev1.StorageClass
	var createdCount int

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		// 移除开头的注释行
		lines := strings.Split(doc, "\n")
		var cleanedLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			cleanedLines = append(cleanedLines, line)
		}

		if len(cleanedLines) == 0 {
			continue
		}

		cleanedDoc := strings.Join(cleanedLines, "\n")

		// 检查是否是 StorageClass 类型
		if !strings.Contains(cleanedDoc, "kind:") || !strings.Contains(cleanedDoc, "StorageClass") {
			global.Logger.Warnf("跳过非 StorageClass 类型资源 (文档 %d)", i+1)
			continue
		}

		var sc storagev1.StorageClass
		if err := yaml.Unmarshal([]byte(cleanedDoc), &sc); err != nil {
			return nil, fmt.Errorf("解析文档 %d 失败: %w", i+1, err)
		}

		// 检查必要字段
		if sc.Name == "" {
			return nil, fmt.Errorf("文档 %d 中缺少 metadata.name", i+1)
		}

		// 创建 StorageClass
		created, err := Kube.StorageV1().
			StorageClasses().
			Create(ctx, &sc, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("创建 StorageClass [%s] 失败: %w", sc.Name, err)
		}

		global.Logger.Infof("StorageClass [%s] 创建成功", created.Name)
		lastStorageClass = created
		createdCount++
	}

	if lastStorageClass == nil {
		return nil, fmt.Errorf("YAML 中没有找到有效的 StorageClass 资源")
	}

	global.Logger.Infof("共创建 %d 个 StorageClass", createdCount)
	return lastStorageClass, nil
}

// ApplyStorageClassYaml 应用 YAML（创建或更新）
func ApplyStorageClassYaml(ctx context.Context, Kube kubernetes.Interface, yamlContent string) (*storagev1.StorageClass, error) {
	// 解析 YAML
	var sc storagev1.StorageClass
	if err := yaml.Unmarshal([]byte(yamlContent), &sc); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}

	// 检查必要字段
	if sc.Name == "" {
		return nil, fmt.Errorf("YAML 中缺少 metadata.name")
	}

	// 尝试获取现有资源
	existing, err := Kube.StorageV1().
		StorageClasses().
		Get(ctx, sc.Name, metav1.GetOptions{})

	if err == nil {
		// 资源存在，更新
		sc.ResourceVersion = existing.ResourceVersion
		updated, err := Kube.StorageV1().
			StorageClasses().
			Update(ctx, &sc, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("更新 StorageClass 失败: %w", err)
		}
		global.Logger.Infof("StorageClass [%s] 更新成功", updated.Name)
		return updated, nil
	}

	// 资源不存在，创建
	created, err := Kube.StorageV1().
		StorageClasses().
		Create(ctx, &sc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("创建 StorageClass 失败: %w", err)
	}

	global.Logger.Infof("StorageClass [%s] 创建成功", created.Name)
	return created, nil
}
