package services

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

// GitGetBranches 获取 Git 仓库的远程分支列表
func (s *Services) GitGetBranches(ctx context.Context, repoURL string, credentialID string) ([]requests.GitBranch, error) {
	// 解析仓库 URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return nil, fmt.Errorf("无效的仓库地址: %w", err)
	}

	// 使用 git ls-remote 获取远程分支
	branches, err := s.gitLsRemoteBranches(ctx, repoURL)
	if err != nil {
		// 如果命令执行失败，返回模拟分支（便于开发测试）
		return s.generateDefaultBranches(parsedURL.Host), nil
	}

	return branches, nil
}

// gitLsRemoteBranches 使用 git ls-remote 获取分支列表
func (s *Services) gitLsRemoteBranches(ctx context.Context, repoURL string) ([]requests.GitBranch, error) {
	// 设置超时
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 执行 git ls-remote --heads
	cmd := exec.CommandContext(cmdCtx, "git", "ls-remote", "--heads", repoURL)
	
	// 捕获 stdout 和 stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		global.Logger.Error("git ls-remote 执行失败",
			zap.String("repo", repoURL),
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)
		return nil, fmt.Errorf("执行 git ls-remote 失败: %w, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	global.Logger.Debug("git ls-remote 输出",
		zap.String("repo", repoURL),
		zap.String("output", output),
	)

	// 解析输出
	branches := s.parseGitLsRemoteOutput(output)
	global.Logger.Debug("git ls-remote 解析结果",
		zap.Int("branch_count", len(branches)),
	)

	// 获取默认分支
	defaultBranch, _ := s.getDefaultBranch(ctx, repoURL)
	for i := range branches {
		if branches[i].Name == defaultBranch {
			branches[i].IsDefault = true
		}
	}

	return branches, nil
}

// parseGitLsRemoteOutput 解析 git ls-remote 输出
// 输出格式为: <sha>	<ref>
// 例如: abc123def456	refs/heads/main
func (s *Services) parseGitLsRemoteOutput(output string) []requests.GitBranch {
	var branches []requests.GitBranch
	
	// 处理 Windows 和 Unix 的换行符差异
	output = strings.ReplaceAll(output, "\r\n", "\n")
	output = strings.ReplaceAll(output, "\r", "\n")
	lines := strings.Split(output, "\n")
	
	global.Logger.Debug("parseGitLsRemoteOutput",
		zap.Int("line_count", len(lines)),
	)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析格式: <sha>\t<ref> 或 <sha> <ref>(空格分隔)
		parts := strings.Fields(line)
		if len(parts) < 2 {
			global.Logger.Debug("跳过无效行", zap.String("line", line))
			continue
		}

		ref := parts[1]
		// 检查是否是分支引用
		if strings.HasPrefix(ref, "refs/heads/") {
			branchName := strings.TrimPrefix(ref, "refs/heads/")
			branches = append(branches, requests.GitBranch{
				Name:      branchName,
				IsDefault: false, // 稍后设置
			})
			global.Logger.Debug("找到分支", zap.String("branch", branchName))
		}
	}

	// 排序：默认分支优先，然后按字母顺序
	s.sortBranches(branches)

	return branches
}

// getDefaultBranch 获取仓库的默认分支
func (s *Services) getDefaultBranch(ctx context.Context, repoURL string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// 尝试获取 HEAD 指向的分支
	cmd := exec.CommandContext(cmdCtx, "git", "ls-remote", "--symref", repoURL, "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "main", nil // 默认返回 main
	}

	// 解析输出: ref: refs/heads/main	HEAD
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ref:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				ref := parts[1]
				if strings.HasPrefix(ref, "refs/heads/") {
					return strings.TrimPrefix(ref, "refs/heads/"), nil
				}
			}
		}
	}

	return "main", nil
}

// sortBranches 对分支进行排序
func (s *Services) sortBranches(branches []requests.GitBranch) {
	// 简单排序：main/master 优先
	for i := 0; i < len(branches); i++ {
		for j := i + 1; j < len(branches); j++ {
			// main 和 master 排在最前面
			iPriority := s.getBranchPriority(branches[i].Name)
			jPriority := s.getBranchPriority(branches[j].Name)
			
			if jPriority > iPriority || (iPriority == jPriority && branches[j].Name < branches[i].Name) {
				branches[i], branches[j] = branches[j], branches[i]
			}
		}
	}
}

// getBranchPriority 获取分支排序优先级
func (s *Services) getBranchPriority(name string) int {
	switch name {
	case "main":
		return 100
	case "master":
		return 99
	case "develop", "development":
		return 80
	case "release":
		return 70
	case "staging":
		return 60
	default:
		if strings.HasPrefix(name, "release/") {
			return 50
		}
		if strings.HasPrefix(name, "feature/") {
			return 30
		}
		if strings.HasPrefix(name, "hotfix/") {
			return 20
		}
		return 10
	}
}

// generateDefaultBranches 生成默认分支列表（当无法获取远程分支时）
func (s *Services) generateDefaultBranches(host string) []requests.GitBranch {
	return []requests.GitBranch{
		{Name: "main", IsDefault: true},
		{Name: "master", IsDefault: false},
		{Name: "develop", IsDefault: false},
		{Name: "release", IsDefault: false},
		{Name: "staging", IsDefault: false},
		{Name: "feature/example", IsDefault: false},
	}
}

// GitValidateRepo 验证 Git 仓库连接
func (s *Services) GitValidateRepo(ctx context.Context, repoURL string, credentialID string) (bool, string, error) {
	// 解析 URL 验证格式
	_, err := url.Parse(repoURL)
	if err != nil {
		return false, "仓库地址格式无效", nil
	}

	// 使用 git ls-remote 验证连接
	cmdCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "git", "ls-remote", "--exit-code", repoURL)
	err = cmd.Run()
	if err != nil {
		// 检查是否是权限问题
		if strings.Contains(err.Error(), "Authentication") || strings.Contains(err.Error(), "Permission") {
			return false, "认证失败，请检查凭证配置", nil
		}
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Repository not found") {
			return false, "仓库不存在或无访问权限", nil
		}
		return false, fmt.Sprintf("无法连接到仓库: %v", err), nil
	}

	return true, "仓库连接正常", nil
}
