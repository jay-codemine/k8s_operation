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

// copyBufferSize 文件拷贝缓冲区大小（1MB，默认 io.Copy 仅 32KB）
const copyBufferSize = 1 << 20

// ArtifactUploadDir 制品存储根目录
const ArtifactUploadDir = "storage/artifacts"

// ArtifactUpload 上传制品（Jenkins 构建完成后回调上传，或手动上传）
func (s *Services) ArtifactUpload(ctx context.Context, file multipart.File, header *multipart.FileHeader, artifact *models.CicdArtifact, userID int64) (*models.CicdArtifact, error) {
	// 1. 生成存储路径：storage/artifacts/{pipeline_id}/{YYYYMMDD}/{filename}
	dateDir := time.Now().Format("20060102")
	var subDir string
	if artifact.PipelineID > 0 {
		subDir = fmt.Sprintf("%d/%s", artifact.PipelineID, dateDir)
	} else {
		subDir = fmt.Sprintf("manual/%s", dateDir)
	}
	storeDir := filepath.Join(ArtifactUploadDir, subDir)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	// 2. 写入文件并计算 SHA256（使用大缓冲区加速 I/O）
	storePath := filepath.Join(storeDir, header.Filename)
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

	// 3. 填充制品信息
	artifact.FilePath = storePath
	artifact.FileSize = written
	artifact.Sha256 = fmt.Sprintf("%x", hasher.Sum(nil))
	artifact.StorageType = "local"
	artifact.Status = models.ArtifactStatusReady
	artifact.CreatedUserID = userID

	// 自动推导制品名称
	if artifact.Name == "" {
		artifact.Name = header.Filename
	}

	// 自动推导制品类型
	if artifact.ArtifactType == "" {
		artifact.ArtifactType = inferArtifactType(header.Filename, artifact.LanguageType)
	}

	// 4. 写入数据库
	if err := s.dao.ArtifactCreate(ctx, artifact); err != nil {
		os.Remove(storePath)
		return nil, fmt.Errorf("保存制品记录失败: %w", err)
	}

	global.Logger.Info("[制品库] 上传成功",
		zap.Int64("artifact_id", artifact.ID),
		zap.String("name", artifact.Name),
		zap.Int64("size", artifact.FileSize),
		zap.String("sha256", artifact.Sha256),
	)

	return artifact, nil
}

// ArtifactCreateRecord 仅创建制品记录（镜像类型，不需要上传文件）
func (s *Services) ArtifactCreateRecord(ctx context.Context, artifact *models.CicdArtifact, userID int64) (*models.CicdArtifact, error) {
	artifact.CreatedUserID = userID
	artifact.Status = models.ArtifactStatusReady
	if artifact.ArtifactType == "" {
		artifact.ArtifactType = models.ArtifactTypeImage
	}
	if err := s.dao.ArtifactCreate(ctx, artifact); err != nil {
		return nil, fmt.Errorf("创建制品记录失败: %w", err)
	}
	return artifact, nil
}

// ArtifactList 制品列表
func (s *Services) ArtifactList(ctx context.Context, pipelineID int64, artifactType, languageType, status string, page, pageSize int) ([]*models.CicdArtifact, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.dao.ArtifactList(ctx, pipelineID, artifactType, languageType, status, page, pageSize)
}

// ArtifactDetail 制品详情
func (s *Services) ArtifactDetail(ctx context.Context, id int64) (*models.CicdArtifact, error) {
	return s.dao.ArtifactGetByID(ctx, id)
}

// ArtifactDownload 制品下载（返回文件路径）
func (s *Services) ArtifactDownload(ctx context.Context, id int64) (*models.CicdArtifact, error) {
	artifact, err := s.dao.ArtifactGetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("制品不存在: %w", err)
	}
	if artifact.FilePath == "" {
		return nil, fmt.Errorf("该制品无文件（镜像类型仅记录引用）")
	}
	if _, err := os.Stat(artifact.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("制品文件已丢失: %s", artifact.FilePath)
	}
	// 增加下载计数
	_ = s.dao.ArtifactIncrDownload(ctx, id)
	return artifact, nil
}

// ArtifactDelete 删除制品
func (s *Services) ArtifactDelete(ctx context.Context, id int64) error {
	artifact, err := s.dao.ArtifactGetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("制品不存在: %w", err)
	}
	// 软删除数据库记录
	if err := s.dao.ArtifactDelete(ctx, id); err != nil {
		return fmt.Errorf("删除制品记录失败: %w", err)
	}
	// 删除文件（如果存在）
	if artifact.FilePath != "" {
		if err := os.Remove(artifact.FilePath); err != nil && !os.IsNotExist(err) {
			global.Logger.Warn("[制品库] 删除文件失败",
				zap.String("file_path", artifact.FilePath),
				zap.Error(err),
			)
		}
	}
	return nil
}

// ArtifactUpdate 更新制品信息
func (s *Services) ArtifactUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
	// 校验制品是否存在
	_, err := s.dao.ArtifactGetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("制品不存在: %w", err)
	}
	if err := s.dao.ArtifactUpdate(ctx, id, updates); err != nil {
		return fmt.Errorf("更新制品失败: %w", err)
	}
	global.Logger.Info("[制品库] 更新成功",
		zap.Int64("artifact_id", id),
	)
	return nil
}

// ArtifactBatchDelete 批量删除制品
func (s *Services) ArtifactBatchDelete(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("删除ID列表不能为空")
	}
	// 先查询文件路径，用于清理文件
	artifacts, err := s.dao.ArtifactGetByIDs(ctx, ids)
	if err != nil {
		return 0, fmt.Errorf("查询制品列表失败: %w", err)
	}
	// 执行软删除
	affected, err := s.dao.ArtifactBatchDelete(ctx, ids)
	if err != nil {
		return 0, fmt.Errorf("批量删除制品记录失败: %w", err)
	}
	// 清理文件（异步，不影响主流程）
	for _, a := range artifacts {
		if a.FilePath != "" {
			if err := os.Remove(a.FilePath); err != nil && !os.IsNotExist(err) {
				global.Logger.Warn("[制品库] 批量删除-清理文件失败",
					zap.String("file_path", a.FilePath),
					zap.Error(err),
				)
			}
		}
	}
	global.Logger.Info("[制品库] 批量删除成功",
		zap.Int64("affected", affected),
		zap.Int("requested", len(ids)),
	)
	return affected, nil
}

// ArtifactAttachFile 为已有制品记录补传/替换文件
func (s *Services) ArtifactAttachFile(ctx context.Context, id int64, file multipart.File, header *multipart.FileHeader) (*models.CicdArtifact, error) {
	artifact, err := s.dao.ArtifactGetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("制品不存在: %w", err)
	}

	// 删除旧文件（如果有）
	if artifact.FilePath != "" {
		_ = os.Remove(artifact.FilePath)
	}

	// 生成存储路径
	dateDir := time.Now().Format("20060102")
	var subDir string
	if artifact.PipelineID > 0 {
		subDir = fmt.Sprintf("%d/%s", artifact.PipelineID, dateDir)
	} else {
		subDir = fmt.Sprintf("manual/%s", dateDir)
	}
	storeDir := filepath.Join(ArtifactUploadDir, subDir)
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	storePath := filepath.Join(storeDir, header.Filename)
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

	// 更新数据库
	updates := map[string]interface{}{
		"file_path":    storePath,
		"file_size":    written,
		"sha256":       fmt.Sprintf("%x", hasher.Sum(nil)),
		"storage_type": "local",
		"status":       models.ArtifactStatusReady,
	}
	// 如果名称为空则用文件名
	if artifact.Name == "" {
		updates["name"] = header.Filename
	}
	if err := s.dao.ArtifactUpdate(ctx, id, updates); err != nil {
		os.Remove(storePath)
		return nil, fmt.Errorf("更新制品记录失败: %w", err)
	}

	// 返回更新后的记录
	return s.dao.ArtifactGetByID(ctx, id)
}

// ArtifactStats 制品统计
func (s *Services) ArtifactStats(ctx context.Context, pipelineID int64) ([]map[string]interface{}, error) {
	return s.dao.ArtifactStats(ctx, pipelineID)
}

// ArtifactListByRunID 获取某次运行的制品列表
func (s *Services) ArtifactListByRunID(ctx context.Context, runID int64) ([]*models.CicdArtifact, error) {
	return s.dao.ArtifactListByRunID(ctx, runID)
}

// inferArtifactType 根据文件名和语言类型推导制品类型
func inferArtifactType(filename string, languageType string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jar":
		return models.ArtifactTypeJar
	case ".war":
		return models.ArtifactTypeWar
	case ".whl":
		return models.ArtifactTypeWheel
	case ".tar", ".gz", ".tgz", ".zip":
		if strings.Contains(filename, "dist") {
			return models.ArtifactTypeDist
		}
		return models.ArtifactTypeArchive
	default:
		// 根据语言类型推导
		if t, ok := models.ArtifactTypeByLanguage[languageType]; ok {
			return t
		}
		return models.ArtifactTypeArchive
	}
}
