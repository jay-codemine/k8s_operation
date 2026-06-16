package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/models"
)

// ACRClient 阿里云容器镜像服务客户端
// 使用阿里云 OpenAPI 获取命名空间、仓库和标签信息
type ACRClient struct {
	accessKeyID     string
	accessKeySecret string
	region          string
	instanceID      string // 企业版实例 ID（个人版为空）
	httpClient      *http.Client
}

// NewACRClient 创建阿里云 ACR 客户端
func NewACRClient(registry *models.ImageRegistry) (*ACRClient, error) {
	if registry.AccessKeyID == "" || registry.AccessKeySecret == "" {
		return nil, fmt.Errorf("阿里云 ACR 需要配置 AccessKey ID 和 AccessKey Secret")
	}

	region := registry.Region
	if region == "" {
		// 尝试从 URL 中解析区域
		// URL 格式: https://registry.cn-hangzhou.aliyuncs.com
		region = parseRegionFromURL(registry.URL)
	}
	if region == "" {
		return nil, fmt.Errorf("未配置区域，请配置 Region 或使用标准 ACR URL")
	}

	return &ACRClient{
		accessKeyID:     registry.AccessKeyID,
		accessKeySecret: registry.AccessKeySecret,
		region:          region,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// parseRegionFromURL 从 ACR URL 中解析区域
func parseRegionFromURL(acrURL string) string {
	// https://registry.cn-hangzhou.aliyuncs.com -> cn-hangzhou
	u, err := url.Parse(acrURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(u.Host, ".")
	if len(parts) >= 3 && parts[0] == "registry" {
		return parts[1]
	}
	return ""
}

// ListRepositories 列出所有仓库（先获取命名空间，再获取每个命名空间下的仓库）
func (c *ACRClient) ListRepositories(ctx context.Context) ([]Repository, error) {
	// 1. 获取所有命名空间
	namespaces, err := c.listNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取命名空间失败: %w", err)
	}

	global.Logger.Info("[ACR] 获取到命名空间", zap.Int("count", len(namespaces)), zap.Strings("namespaces", namespaces))

	// 2. 获取每个命名空间下的仓库
	var allRepos []Repository
	for _, ns := range namespaces {
		repos, err := c.listReposInNamespace(ctx, ns)
		if err != nil {
			global.Logger.Warn("[ACR] 获取命名空间仓库失败", zap.String("namespace", ns), zap.Error(err))
			continue
		}
		allRepos = append(allRepos, repos...)
	}

	return allRepos, nil
}

// listNamespaces 获取命名空间列表
func (c *ACRClient) listNamespaces(ctx context.Context) ([]string, error) {
	params := map[string]string{
		"Action":   "GetNamespaceList",
		"RegionId": c.region,
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code       string `json:"code"`
		Message    string `json:"message"`
		Namespaces struct {
			Namespace []struct {
				Namespace string `json:"Namespace"`
			} `json:"namespace"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, body: %s", err, string(body))
	}

	if resp.Code != "200" && resp.Code != "" {
		return nil, fmt.Errorf("API 错误: %s - %s", resp.Code, resp.Message)
	}

	names := make([]string, 0, len(resp.Namespaces.Namespace))
	for _, ns := range resp.Namespaces.Namespace {
		names = append(names, ns.Namespace)
	}
	return names, nil
}

// listReposInNamespace 获取命名空间下的仓库列表
func (c *ACRClient) listReposInNamespace(ctx context.Context, namespace string) ([]Repository, error) {
	params := map[string]string{
		"Action":        "GetRepoList",
		"RegionId":      c.region,
		"RepoNamespace": namespace,
		"Page":          "1",
		"PageSize":      "100",
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Repos struct {
				Repo []struct {
					RepoName        string `json:"RepoName"`
					RepoNamespace   string `json:"RepoNamespace"`
					Summary         string `json:"Summary"`
					RepoStatus      string `json:"RepoStatus"`
					Downloads       int64  `json:"Downloads"`
					GmtCreate       int64  `json:"GmtCreate"`
					GmtModified     int64  `json:"GmtModified"`
					RepoAuthorizeType string `json:"RepoAuthorizeType"`
				} `json:"repo"`
			} `json:"repos"`
			Total int `json:"total"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Code != "200" && resp.Code != "" {
		return nil, fmt.Errorf("API 错误: %s - %s", resp.Code, resp.Message)
	}

	repos := make([]Repository, 0, len(resp.Data.Repos.Repo))
	for _, r := range resp.Data.Repos.Repo {
		fullName := fmt.Sprintf("%s/%s", r.RepoNamespace, r.RepoName)
		repos = append(repos, Repository{
			Name:        r.RepoName,
			FullName:    fullName,
			Description: r.Summary,
			PullCount:   r.Downloads,
			CreatedAt:   r.GmtCreate,
			UpdatedAt:   r.GmtModified,
		})
	}

	return repos, nil
}

// ListTags 获取镜像标签列表
func (c *ACRClient) ListTags(ctx context.Context, repository string) ([]ImageTag, error) {
	// repository 格式: namespace/repo_name
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的仓库格式，需要 namespace/repo_name")
	}

	namespace := parts[0]
	repoName := parts[1]

	params := map[string]string{
		"Action":        "GetRepoTags",
		"RegionId":      c.region,
		"RepoNamespace": namespace,
		"RepoName":      repoName,
		"Page":          "1",
		"PageSize":      "100",
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Tags struct {
				Tag []struct {
					Tag         string `json:"tag"`
					ImageId     string `json:"imageId"`
					Digest      string `json:"digest"`
					ImageSize   int64  `json:"imageSize"`
					ImageCreate int64  `json:"imageCreate"`
					ImageUpdate int64  `json:"imageUpdate"`
				} `json:"tag"`
			} `json:"tags"`
			Total int `json:"total"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Code != "200" && resp.Code != "" {
		return nil, fmt.Errorf("API 错误: %s - %s", resp.Code, resp.Message)
	}

	tags := make([]ImageTag, 0, len(resp.Data.Tags.Tag))
	for _, t := range resp.Data.Tags.Tag {
		tags = append(tags, ImageTag{
			Name:      t.Tag,
			Digest:    t.Digest,
			Size:      t.ImageSize,
			CreatedAt: t.ImageCreate,
			PushedAt:  t.ImageUpdate,
		})
	}

	// 按更新时间排序（最新的在前）
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].PushedAt > tags[j].PushedAt
	})

	return tags, nil
}

// GetManifest 获取镜像详情
func (c *ACRClient) GetManifest(ctx context.Context, repository, tag string) (*ImageManifest, error) {
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的仓库格式")
	}

	namespace := parts[0]
	repoName := parts[1]

	params := map[string]string{
		"Action":        "GetRepoTagManifest",
		"RegionId":      c.region,
		"RepoNamespace": namespace,
		"RepoName":      repoName,
		"Tag":           tag,
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Digest        string `json:"digest"`
			MediaType     string `json:"mediaType"`
			SchemaVersion int    `json:"schemaVersion"`
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
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		// 如果 Manifest 接口不可用，使用 Tag 详情接口
		return c.getTagDetail(ctx, namespace, repoName, tag)
	}

	if resp.Code != "200" && resp.Code != "" {
		return c.getTagDetail(ctx, namespace, repoName, tag)
	}

	result := &ImageManifest{
		Digest:        resp.Data.Digest,
		MediaType:     resp.Data.MediaType,
		SchemaVersion: resp.Data.SchemaVersion,
	}

	if resp.Data.Config.Digest != "" {
		result.Config = &ManifestConfig{
			Digest:    resp.Data.Config.Digest,
			Size:      resp.Data.Config.Size,
			MediaType: resp.Data.Config.MediaType,
		}
	}

	for _, layer := range resp.Data.Layers {
		result.Layers = append(result.Layers, ManifestLayer{
			Digest:    layer.Digest,
			Size:      layer.Size,
			MediaType: layer.MediaType,
		})
		result.Size += layer.Size
	}

	return result, nil
}

// getTagDetail 获取标签详情（备用方法）
func (c *ACRClient) getTagDetail(ctx context.Context, namespace, repoName, tag string) (*ImageManifest, error) {
	params := map[string]string{
		"Action":        "GetRepoTag",
		"RegionId":      c.region,
		"RepoNamespace": namespace,
		"RepoName":      repoName,
		"Tag":           tag,
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Tag       string `json:"tag"`
			Digest    string `json:"digest"`
			ImageSize int64  `json:"imageSize"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &ImageManifest{
		Digest: resp.Data.Digest,
		Size:   resp.Data.ImageSize,
	}, nil
}

// DeleteTag 删除镜像标签
func (c *ACRClient) DeleteTag(ctx context.Context, repository, tag string) error {
	parts := strings.SplitN(repository, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("无效的仓库格式")
	}

	namespace := parts[0]
	repoName := parts[1]

	params := map[string]string{
		"Action":        "DeleteRepoTag",
		"RegionId":      c.region,
		"RepoNamespace": namespace,
		"RepoName":      repoName,
		"Tag":           tag,
	}

	body, err := c.doRequest(ctx, params)
	if err != nil {
		return err
	}

	var resp struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Code != "200" && resp.Code != "" {
		return fmt.Errorf("删除失败: %s - %s", resp.Code, resp.Message)
	}

	return nil
}

// doRequest 执行阿里云 API 请求
func (c *ACRClient) doRequest(ctx context.Context, params map[string]string) ([]byte, error) {
	// 阿里云 API 端点
	endpoint := fmt.Sprintf("https://cr.%s.aliyuncs.com", c.region)

	// 添加公共参数
	params["Format"] = "JSON"
	params["Version"] = "2016-06-07"
	params["AccessKeyId"] = c.accessKeyID
	params["SignatureMethod"] = "HMAC-SHA1"
	params["Timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	params["SignatureVersion"] = "1.0"
	params["SignatureNonce"] = fmt.Sprintf("%d", time.Now().UnixNano())

	// 计算签名
	signature := c.computeSignature(params)
	params["Signature"] = signature

	// 构建请求 URL
	queryStr := c.buildQueryString(params)
	reqURL := endpoint + "/?" + queryStr

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// computeSignature 计算阿里云 API 签名
func (c *ACRClient) computeSignature(params map[string]string) string {
	// 1. 按参数名排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 构建规范化请求字符串
	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, percentEncode(k)+"="+percentEncode(params[k]))
	}
	canonicalizedQueryString := strings.Join(pairs, "&")

	// 3. 构建待签名字符串
	stringToSign := "GET&" + percentEncode("/") + "&" + percentEncode(canonicalizedQueryString)

	// 4. 计算 HMAC-SHA1 签名
	h := hmac.New(sha1.New, []byte(c.accessKeySecret+"&"))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

// buildQueryString 构建查询字符串
func (c *ACRClient) buildQueryString(params map[string]string) string {
	var pairs []string
	for k, v := range params {
		pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	return strings.Join(pairs, "&")
}

// percentEncode URL 编码（阿里云特殊编码规则）
func percentEncode(s string) string {
	s = url.QueryEscape(s)
	s = strings.ReplaceAll(s, "+", "%20")
	s = strings.ReplaceAll(s, "*", "%2A")
	s = strings.ReplaceAll(s, "%7E", "~")
	return s
}
