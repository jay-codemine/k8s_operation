package services

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"k8soperation/global"
	"k8soperation/internal/app/models"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AgentUploadDir 探针文件存储根目录
const AgentUploadDir = "storage/agents"

// BuildAgentUpload 上传探针文件（创建或更新）
func (s *Services) BuildAgentUpload(ctx context.Context, file multipart.File, header *multipart.FileHeader, agent *models.CicdBuildAgent, userID int64) (*models.CicdBuildAgent, error) {
	// 1. 生成存储路径：storage/agents/{category}/{filename}
	storeDir := filepath.Join(AgentUploadDir, agent.Category)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	// 使用探针名称作为目录，版本号拼接文件名，确保唯一
	agentDir := filepath.Join(storeDir, agent.Name)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		return nil, fmt.Errorf("创建探针目录失败: %w", err)
	}

	storePath := filepath.Join(agentDir, header.Filename)

	// 2. 写入文件并计算 SHA256
	dst, err := os.Create(storePath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	bufDst := bufio.NewWriterSize(dst, copyBufferSize)
	hasher := sha256.New()
	writer := io.MultiWriter(bufDst, hasher)
	buf := make([]byte, copyBufferSize)
	written, err := io.CopyBuffer(writer, file, buf)
	if err != nil {
		os.Remove(storePath)
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}
	if err := bufDst.Flush(); err != nil {
		os.Remove(storePath)
		return nil, fmt.Errorf("刷新缓冲区失败: %w", err)
	}

	// 3. 填充探针信息
	agent.FilePath = storePath
	agent.FileSize = written
	agent.FileName = header.Filename
	agent.Sha256 = fmt.Sprintf("%x", hasher.Sum(nil))
	agent.Status = models.AgentStatusActive
	agent.CreatedUserID = userID

	// 自动推导 DockerCopyDest
	if agent.DockerCopyDest == "" {
		agent.DockerCopyDest = "/app/" + header.Filename
	}

	// 4. 写入数据库
	if err := s.dao.BuildAgentCreate(ctx, agent); err != nil {
		// 如果是重名，尝试更新
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			return s.buildAgentUpdateFile(ctx, agent, storePath, written, fmt.Sprintf("%x", hasher.Sum(nil)))
		}
		os.Remove(storePath)
		return nil, fmt.Errorf("保存探针记录失败: %w", err)
	}

	global.Logger.Info("[探针管理] 上传成功",
		zap.Int64("agent_id", agent.ID),
		zap.String("name", agent.Name),
		zap.Int64("size", agent.FileSize),
	)

	return agent, nil
}

// buildAgentUpdateFile 更新已有探针的文件（版本升级场景）
func (s *Services) buildAgentUpdateFile(ctx context.Context, agent *models.CicdBuildAgent, filePath string, fileSize int64, sha256Hash string) (*models.CicdBuildAgent, error) {
	existing, err := s.dao.BuildAgentGetByName(ctx, agent.Name)
	if err != nil {
		return nil, fmt.Errorf("查询已有探针失败: %w", err)
	}

	// 删除旧文件
	if existing.FilePath != "" && existing.FilePath != filePath {
		_ = os.Remove(existing.FilePath)
	}

	updates := map[string]interface{}{
		"file_path":       filePath,
		"file_size":       fileSize,
		"file_name":       agent.FileName,
		"sha256":          sha256Hash,
		"version":         agent.Version,
		"display_name":    agent.DisplayName,
		"description":     agent.Description,
		"docker_copy_dest": agent.DockerCopyDest,
		"env_key":         agent.EnvKey,
		"env_value":       agent.EnvValue,
		"status":          models.AgentStatusActive,
	}
	if agent.DownloadURL != "" {
		updates["download_url"] = agent.DownloadURL
	}
	if agent.DocURL != "" {
		updates["doc_url"] = agent.DocURL
	}
	if agent.Icon != "" {
		updates["icon"] = agent.Icon
	}

	if err := s.dao.BuildAgentUpdate(ctx, existing.ID, updates); err != nil {
		return nil, fmt.Errorf("更新探针失败: %w", err)
	}

	global.Logger.Info("[探针管理] 版本更新",
		zap.Int64("agent_id", existing.ID),
		zap.String("name", agent.Name),
		zap.String("old_version", existing.Version),
		zap.String("new_version", agent.Version),
	)

	return s.dao.BuildAgentGetByID(ctx, existing.ID)
}

// BuildAgentList 探针列表
func (s *Services) BuildAgentList(ctx context.Context, category, scope, status, keyword string, page, pageSize int) ([]*models.CicdBuildAgent, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.dao.BuildAgentList(ctx, category, scope, status, keyword, page, pageSize)
}

// BuildAgentDetail 探针详情
func (s *Services) BuildAgentDetail(ctx context.Context, id int64) (*models.CicdBuildAgent, error) {
	return s.dao.BuildAgentGetByID(ctx, id)
}

// BuildAgentDownload 下载探针文件（返回文件路径）
func (s *Services) BuildAgentDownload(ctx context.Context, id int64) (*models.CicdBuildAgent, error) {
	agent, err := s.dao.BuildAgentGetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("探针不存在: %w", err)
	}
	if agent.FilePath == "" {
		return nil, fmt.Errorf("该探针无文件")
	}
	if _, err := os.Stat(agent.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("探针文件已丢失: %s", agent.FilePath)
	}
	_ = s.dao.BuildAgentIncrDownload(ctx, id)
	return agent, nil
}

// BuildAgentDownloadByName 按名称下载探针（流水线构建使用）
func (s *Services) BuildAgentDownloadByName(ctx context.Context, name string) (*models.CicdBuildAgent, error) {
	agent, err := s.dao.BuildAgentGetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("探针不存在: %s", name)
	}
	if agent.Status != models.AgentStatusActive {
		return nil, fmt.Errorf("探针已停用: %s", name)
	}
	if agent.FilePath == "" {
		return nil, fmt.Errorf("该探针无文件")
	}
	if _, err := os.Stat(agent.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("探针文件已丢失: %s", agent.FilePath)
	}
	_ = s.dao.BuildAgentIncrDownload(ctx, agent.ID)
	return agent, nil
}

// BuildAgentListByScope 获取指定语言的已启用探针
func (s *Services) BuildAgentListByScope(ctx context.Context, scope string) ([]*models.CicdBuildAgent, error) {
	return s.dao.BuildAgentListByScope(ctx, scope)
}

// BuildAgentUpdate 更新探针信息
func (s *Services) BuildAgentUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	_, err := s.dao.BuildAgentGetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("探针不存在: %w", err)
	}
	return s.dao.BuildAgentUpdate(ctx, id, updates)
}

// BuildAgentToggleStatus 切换探针启用/停用状态
func (s *Services) BuildAgentToggleStatus(ctx context.Context, id int64) (*models.CicdBuildAgent, error) {
	agent, err := s.dao.BuildAgentGetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("探针不存在: %w", err)
	}
	newStatus := models.AgentStatusActive
	if agent.Status == models.AgentStatusActive {
		newStatus = models.AgentStatusInactive
	}
	if err := s.dao.BuildAgentUpdate(ctx, id, map[string]interface{}{"status": newStatus}); err != nil {
		return nil, fmt.Errorf("切换状态失败: %w", err)
	}
	return s.dao.BuildAgentGetByID(ctx, id)
}

// BuildAgentDelete 删除探针
func (s *Services) BuildAgentDelete(ctx context.Context, id int64) error {
	agent, err := s.dao.BuildAgentGetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("探针不存在: %w", err)
	}
	if err := s.dao.BuildAgentDelete(ctx, id); err != nil {
		return fmt.Errorf("删除探针记录失败: %w", err)
	}
	// 清理文件
	if agent.FilePath != "" {
		if err := os.Remove(agent.FilePath); err != nil && !os.IsNotExist(err) {
			global.Logger.Warn("[探针管理] 删除文件失败", zap.String("file_path", agent.FilePath), zap.Error(err))
		}
	}
	return nil
}

// BuildAgentSeedOTEL 初始化 OTEL Java Agent 种子数据（如果不存在）
func (s *Services) BuildAgentSeedOTEL(ctx context.Context) error {
	_, err := s.dao.BuildAgentGetByName(ctx, "opentelemetry-javaagent")
	if err == nil {
		return nil // 已存在，跳过
	}

	seed := &models.CicdBuildAgent{
		Name:           "opentelemetry-javaagent",
		DisplayName:    "OpenTelemetry Java Agent",
		Description:    "OpenTelemetry 自动检测代理，为 Java 应用提供分布式链路追踪、指标采集能力。支持 Spring Boot、gRPC、JDBC 等 100+ 框架自动埋点。",
		Category:       models.AgentCategoryObservability,
		Scope:          models.AgentScopeJava,
		Version:        "1.33.0",
		FileName:       "opentelemetry-javaagent.jar",
		DownloadURL:    "https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases",
		DocURL:         "https://opentelemetry.io/docs/languages/java/automatic/",
		Icon:           "🔭",
		DockerCopyDest: "/app/opentelemetry-javaagent.jar",
		EnvKey:         "OTEL_OPTS",
		EnvValue:       "-javaagent:/app/opentelemetry-javaagent.jar -Dotel.service.name=${SERVICE_NAME} -Dotel.traces.exporter=otlp -Dotel.metrics.exporter=none -Dotel.logs.exporter=none -Dotel.exporter.otlp.endpoint=http://otel-collector-monitoring.svc.cluster.local:4318",
		Status:         models.AgentStatusActive,
		CreatedAt:      uint64(time.Now().Unix()),
		ModifiedAt:     uint64(time.Now().Unix()),
	}

	if err := s.dao.BuildAgentCreate(ctx, seed); err != nil {
		global.Logger.Warn("[探针管理] 种子数据初始化失败（可能已存在）", zap.Error(err))
	} else {
		global.Logger.Info("[探针管理] 已初始化 OTEL Java Agent 种子数据（需手动上传 JAR 文件）")
	}
	return nil
}
