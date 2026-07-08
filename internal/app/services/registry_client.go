package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// RegistryClient 镜像仓库客户端接口
type RegistryClient interface {
	// ListRepositories 列出所有镜像仓库/项目
	ListRepositories(ctx context.Context) ([]Repository, error)
	// ListTags 列出镜像的所有标签
	ListTags(ctx context.Context, repository string) ([]ImageTag, error)
	// GetManifest 获取镜像 Manifest 信息
	GetManifest(ctx context.Context, repository, tag string) (*ImageManifest, error)
	// DeleteTag 删除指定标签
	DeleteTag(ctx context.Context, repository, tag string) error
}

// Repository 镜像仓库/项目
type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	TagCount    int    `json:"tag_count"`
	PullCount   int64  `json:"pull_count"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// ImageTag 镜像标签
type ImageTag struct {
	Name      string `json:"name"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	CreatedAt int64  `json:"created_at"`
	PushedAt  int64  `json:"pushed_at"`
	Author    string `json:"author"`
}

// ImageManifest 镜像 Manifest
type ImageManifest struct {
	Digest        string            `json:"digest"`
	MediaType     string            `json:"media_type"`
	SchemaVersion int               `json:"schema_version"`
	Size          int64             `json:"size"`
	Layers        []ManifestLayer   `json:"layers"`
	Config        *ManifestConfig   `json:"config"`
	Labels        map[string]string `json:"labels"`
	CreatedAt     int64             `json:"created_at"`
}

// ManifestLayer Manifest 层信息
type ManifestLayer struct {
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	MediaType string `json:"media_type"`
}

// ManifestConfig Manifest 配置
type ManifestConfig struct {
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	MediaType string `json:"media_type"`
}

// NewRegistryClient 创建仓库客户端
func NewRegistryClient(registry *models.ImageRegistry) (RegistryClient, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: registry.Insecure,
			},
		},
	}

	switch registry.Type {
	case "acr":
		// 阿里云容器镜像服务，使用 OpenAPI
		return NewACRClient(registry)
	case "harbor":
		return &HarborClient{
			baseURL:    strings.TrimSuffix(registry.URL, "/"),
			username:   registry.Username,
			password:   registry.Password,
			httpClient: httpClient,
		}, nil
	default:
		// 默认使用 Docker Registry V2 API
		return &DockerRegistryClient{
			baseURL:    strings.TrimSuffix(registry.URL, "/"),
			username:   registry.Username,
			password:   registry.Password,
			httpClient: httpClient,
		}, nil
	}
}

// ========================================
// Docker Registry V2 客户端
// ========================================

type DockerRegistryClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

func (c *DockerRegistryClient) doRequest(ctx context.Context, method, path string, headers map[string]string) (*http.Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *DockerRegistryClient) ListRepositories(ctx context.Context) ([]Repository, error) {
	resp, err := c.doRequest(ctx, "GET", "/v2/_catalog", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取仓库列表失败: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Repositories []string `json:"repositories"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	repos := make([]Repository, 0, len(result.Repositories))
	for _, name := range result.Repositories {
		repos = append(repos, Repository{
			Name:     name,
			FullName: name,
		})
	}

	return repos, nil
}

func (c *DockerRegistryClient) ListTags(ctx context.Context, repository string) ([]ImageTag, error) {
	path := fmt.Sprintf("/v2/%s/tags/list", repository)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取标签列表失败: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	tags := make([]ImageTag, 0, len(result.Tags))
	for _, tag := range result.Tags {
		tags = append(tags, ImageTag{Name: tag})
	}

	return tags, nil
}

func (c *DockerRegistryClient) GetManifest(ctx context.Context, repository, tag string) (*ImageManifest, error) {
	path := fmt.Sprintf("/v2/%s/manifests/%s", repository, tag)
	headers := map[string]string{
		"Accept": "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json",
	}

	resp, err := c.doRequest(ctx, "GET", path, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取 Manifest 失败: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manifest struct {
		SchemaVersion int    `json:"schemaVersion"`
		MediaType     string `json:"mediaType"`
		Config        struct {
			Digest    string `json:"digest"`
			Size      int64  `json:"size"`
			MediaType string `json:"mediaType"`
		} `json:"config"`
		Layers []struct {
			Digest    string `json:"digest"`
			Size      int64  `json:"size"`
			MediaType string `json:"mediaType"`
		} `json:"layers"`
	}

	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, err
	}

	result := &ImageManifest{
		Digest:        resp.Header.Get("Docker-Content-Digest"),
		MediaType:     manifest.MediaType,
		SchemaVersion: manifest.SchemaVersion,
	}

	if manifest.Config.Digest != "" {
		result.Config = &ManifestConfig{
			Digest:    manifest.Config.Digest,
			Size:      manifest.Config.Size,
			MediaType: manifest.Config.MediaType,
		}
	}

	for _, layer := range manifest.Layers {
		result.Layers = append(result.Layers, ManifestLayer{
			Digest:    layer.Digest,
			Size:      layer.Size,
			MediaType: layer.MediaType,
		})
		result.Size += layer.Size
	}

	return result, nil
}

func (c *DockerRegistryClient) DeleteTag(ctx context.Context, repository, tag string) error {
	// 先获取 digest
	manifest, err := c.GetManifest(ctx, repository, tag)
	if err != nil {
		return err
	}

	if manifest.Digest == "" {
		return fmt.Errorf("无法获取镜像 digest")
	}

	path := fmt.Sprintf("/v2/%s/manifests/%s", repository, manifest.Digest)
	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除失败: %s - %s", resp.Status, string(body))
	}

	return nil
}

// ========================================
// Harbor 客户端
// ========================================

type HarborClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

func (c *HarborClient) doRequest(ctx context.Context, method, path string, headers map[string]string) (*http.Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	req.Header.Set("Accept", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *HarborClient) ListRepositories(ctx context.Context) ([]Repository, error) {
	// Harbor API: GET /api/v2.0/projects/{project_name}/repositories
	// 先获取所有项目，再获取每个项目的仓库
	projects, err := c.listProjects(ctx)
	if err != nil {
		return nil, err
	}

	var allRepos []Repository
	for _, project := range projects {
		repos, err := c.listProjectRepositories(ctx, project)
		if err != nil {
			global.Logger.Warn("获取项目仓库失败", zap.String("project", project), zap.Error(err))
			continue
		}
		allRepos = append(allRepos, repos...)
	}

	return allRepos, nil
}

func (c *HarborClient) listProjects(ctx context.Context) ([]string, error) {
	resp, err := c.doRequest(ctx, "GET", "/api/v2.0/projects?page=1&page_size=100", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取项目列表失败: %s - %s", resp.Status, string(body))
	}

	var projects []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(projects))
	for _, p := range projects {
		names = append(names, p.Name)
	}
	return names, nil
}

func (c *HarborClient) listProjectRepositories(ctx context.Context, projectName string) ([]Repository, error) {
	path := fmt.Sprintf("/api/v2.0/projects/%s/repositories?page=1&page_size=100", projectName)
	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取仓库列表失败: %s - %s", resp.Status, string(body))
	}

	var harborRepos []struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		ArtifactCount int   `json:"artifact_count"`
		PullCount    int64  `json:"pull_count"`
		CreationTime string `json:"creation_time"`
		UpdateTime   string `json:"update_time"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&harborRepos); err != nil {
		return nil, err
	}

	repos := make([]Repository, 0, len(harborRepos))
	for _, r := range harborRepos {
		repos = append(repos, Repository{
			Name:        strings.TrimPrefix(r.Name, projectName+"/"),
			FullName:    r.Name,
			Description: r.Description,
			TagCount:    r.ArtifactCount,
			PullCount:   r.PullCount,
		})
	}

	return repos, nil
}

func (c *HarborClient) ListTags(ctx context.Context, repository string) ([]ImageTag, error) {
	// 解析 project/repo 格式
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的仓库名称格式，需要 project/repository 格式")
	}

	projectName := parts[0]
	repoName := parts[1]

	// Harbor API: GET /api/v2.0/projects/{project_name}/repositories/{repository_name}/artifacts
	path := fmt.Sprintf("/api/v2.0/projects/%s/repositories/%s/artifacts?page=1&page_size=100&with_tag=true",
		projectName, strings.ReplaceAll(repoName, "/", "%2F"))

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取标签列表失败: %s - %s", resp.Status, string(body))
	}

	var artifacts []struct {
		Digest   string `json:"digest"`
		Size     int64  `json:"size"`
		PushTime string `json:"push_time"`
		Tags     []struct {
			Name       string `json:"name"`
			PushTime   string `json:"push_time"`
			PullTime   string `json:"pull_time"`
		} `json:"tags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&artifacts); err != nil {
		return nil, err
	}

	var tags []ImageTag
	for _, artifact := range artifacts {
		for _, tag := range artifact.Tags {
			pushedAt, _ := time.Parse(time.RFC3339, tag.PushTime)
			tags = append(tags, ImageTag{
				Name:     tag.Name,
				Digest:   artifact.Digest,
				Size:     artifact.Size,
				PushedAt: pushedAt.Unix(),
			})
		}
	}

	// 按推送时间排序（最新的在前）
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].PushedAt > tags[j].PushedAt
	})

	return tags, nil
}

func (c *HarborClient) GetManifest(ctx context.Context, repository, tag string) (*ImageManifest, error) {
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的仓库名称格式")
	}

	projectName := parts[0]
	repoName := parts[1]

	path := fmt.Sprintf("/api/v2.0/projects/%s/repositories/%s/artifacts/%s",
		projectName, strings.ReplaceAll(repoName, "/", "%2F"), tag)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取 Manifest 失败: %s - %s", resp.Status, string(body))
	}

	var artifact struct {
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
		MediaType string `json:"media_type"`
		PushTime  string `json:"push_time"`
		Labels    map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&artifact); err != nil {
		return nil, err
	}

	pushedAt, _ := time.Parse(time.RFC3339, artifact.PushTime)
	return &ImageManifest{
		Digest:    artifact.Digest,
		MediaType: artifact.MediaType,
		Size:      artifact.Size,
		Labels:    artifact.Labels,
		CreatedAt: pushedAt.Unix(),
	}, nil
}

func (c *HarborClient) DeleteTag(ctx context.Context, repository, tag string) error {
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("无效的仓库名称格式")
	}

	projectName := parts[0]
	repoName := parts[1]

	path := fmt.Sprintf("/api/v2.0/projects/%s/repositories/%s/artifacts/%s",
		projectName, strings.ReplaceAll(repoName, "/", "%2F"), tag)

	resp, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除失败: %s - %s", resp.Status, string(body))
	}

	return nil
}
