package services

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"k8soperation/global"
	"k8soperation/internal/app/requests"
)

// credInURLRe 用于在日志中隐藏 URL 内嵌的凭证：https://user:pass@host → https://***@host
var credInURLRe = regexp.MustCompile(`(https?://)[^/@\s]+@`)

// GitGetBranches 获取 Git 仓库的远程分支列表
//
// 同时支持公网仓库与内网自建仓库（是否可达取决于后端进程/Pod 的网络连通性）。
// 私有仓库需要认证，可通过以下任一方式提供凭证：
//   - 传入 username + password（password 也可以是访问令牌 Token）
//   - 在 repoURL 中内嵌凭证，如 https://user:token@host/org/repo.git
//
// 失败时返回带有明确原因的错误（认证/网络/仓库不存在/超时等），不再静默返回假分支。
func (s *Services) GitGetBranches(ctx context.Context, repoURL, credentialID, username, password string) ([]requests.GitBranch, error) {
	_ = credentialID // 预留：凭证 ID（平台侧暂不解析 Jenkins 凭证明文，认证以直传/URL 内嵌为准）

	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return nil, fmt.Errorf("仓库地址不能为空")
	}
	if _, err := url.Parse(repoURL); err != nil {
		return nil, fmt.Errorf("无效的仓库地址: %w", err)
	}

	// 未显式传凭证时，回退到平台配置的默认 Git 凭证（自动复用，无需每次手输）
	username, password = resolveGitCredentials(username, password)

	branches, err := s.gitLsRemoteBranches(ctx, repoURL, username, password)
	if err != nil {
		return nil, err
	}
	if len(branches) == 0 {
		return nil, fmt.Errorf("未获取到任何分支：请确认该仓库存在且当前账号有访问权限")
	}
	return branches, nil
}

// gitLsRemoteBranches 使用 git ls-remote 获取分支列表
func (s *Services) gitLsRemoteBranches(ctx context.Context, repoURL, username, password string) ([]requests.GitBranch, error) {
	// 设置超时
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 注入认证（仅 http/https 生效）
	authedURL, err := injectGitCredentials(repoURL, username, password)
	if err != nil {
		return nil, fmt.Errorf("无效的仓库地址: %w", err)
	}

	// 执行 git ls-remote --heads
	args := append(gitBaseArgs(), "ls-remote", "--heads", authedURL)
	cmd := exec.CommandContext(cmdCtx, "git", args...)
	cmd.Env = gitCommandEnv()

	// 捕获 stdout 和 stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// 注意：只记录原始 repoURL，绝不记录内嵌凭证的 authedURL
		global.Logger.Error("git ls-remote 执行失败",
			zap.String("repo", repoURL),
			zap.Error(err),
			zap.String("stderr", maskCredentials(stderr.String())),
		)
		return nil, classifyGitError(cmdCtx, err, stderr.String())
	}

	output := stdout.String()

	// 解析输出
	branches := s.parseGitLsRemoteOutput(output)
	global.Logger.Debug("git ls-remote 解析结果",
		zap.String("repo", repoURL),
		zap.Int("branch_count", len(branches)),
	)

	// 获取默认分支
	defaultBranch, _ := s.getDefaultBranch(ctx, repoURL, username, password)
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

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析格式: <sha>\t<ref> 或 <sha> <ref>(空格分隔)
		parts := strings.Fields(line)
		if len(parts) < 2 {
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
		}
	}

	// 排序：默认分支优先，然后按字母顺序
	s.sortBranches(branches)

	return branches
}

// getDefaultBranch 获取仓库的默认分支
func (s *Services) getDefaultBranch(ctx context.Context, repoURL, username, password string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	authedURL, err := injectGitCredentials(repoURL, username, password)
	if err != nil {
		return global.DefaultBranch(), nil
	}

	// 尝试获取 HEAD 指向的分支
	args := append(gitBaseArgs(), "ls-remote", "--symref", authedURL, "HEAD")
	cmd := exec.CommandContext(cmdCtx, "git", args...)
	cmd.Env = gitCommandEnv()
	output, err := cmd.Output()
	if err != nil {
		return global.DefaultBranch(), nil // 默认返回配置的分支
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

	return global.DefaultBranch(), nil
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

// GitValidateRepo 验证 Git 仓库连接（支持公网/内网，私有仓库需传入凭证）
func (s *Services) GitValidateRepo(ctx context.Context, repoURL, credentialID, username, password string) (bool, string, error) {
	_ = credentialID

	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return false, "仓库地址不能为空", nil
	}
	if _, err := url.Parse(repoURL); err != nil {
		return false, "仓库地址格式无效", nil
	}

	// 未显式传凭证时，回退到平台配置的默认 Git 凭证
	username, password = resolveGitCredentials(username, password)

	authedURL, err := injectGitCredentials(repoURL, username, password)
	if err != nil {
		return false, "仓库地址格式无效", nil
	}

	// 使用 git ls-remote 验证连接
	cmdCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	args := append(gitBaseArgs(), "ls-remote", "--exit-code", "--heads", authedURL)
	cmd := exec.CommandContext(cmdCtx, "git", args...)
	cmd.Env = gitCommandEnv()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		global.Logger.Warn("git ls-remote 验证失败",
			zap.String("repo", repoURL),
			zap.Error(err),
			zap.String("stderr", maskCredentials(stderr.String())),
		)
		return false, gitErrorMessage(cmdCtx, err, stderr.String()), nil
	}

	return true, "仓库连接正常", nil
}

// ==================== Git 命令辅助函数 ====================

// gitBaseArgs 返回 git 命令的公共前置参数：
// 关闭凭证助手与交互，确保访问私有仓库无凭证时快速失败，而不是挂起等待输入。
func gitBaseArgs() []string {
	return []string{
		"-c", "credential.helper=",         // 禁用任何已配置的凭证助手
		"-c", "credential.interactive=false", // 禁用交互式凭证获取
	}
}

// gitCommandEnv 返回执行 git 命令的环境变量：
// 禁用交互式凭证输入，避免访问私有仓库时命令挂起直到超时。
func gitCommandEnv() []string {
	return append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0", // 禁止终端交互输入用户名/密码
		"GCM_INTERACTIVE=never", // 禁用 Git Credential Manager 弹窗（Windows/macOS）
	)
}

// resolveGitCredentials 解析实际使用的 Git 凭证：
// 优先使用请求显式传入的 username/password（用于临时访问其它仓库）；
// 二者都为空时，回退到平台配置的默认 Git 凭证
// （global.JenkinsSetting.GitUsername/GitToken，与 Jenkins gitee-id 同源），
// 从而实现「配一次、检测仓库自动复用、无需每次手输账号密码」。
func resolveGitCredentials(username, password string) (string, string) {
	if username != "" || password != "" {
		return username, password
	}
	if global.JenkinsSetting != nil {
		return global.JenkinsSetting.GitUsername, global.JenkinsSetting.GitToken
	}
	return "", ""
}

// injectGitCredentials 将用户名/密码(或 Token) 注入到 http/https 仓库地址中。
//   - 未提供凭证：原样返回；
//   - 同时有 username 和 password：注入 user:password；
//   - 仅有 username：只注入用户名；
//   - 仅有 password/Token：以其作为用户名注入（兼容 GitHub/Gitee Personal Access Token）；
//   - 非 http/https 协议（如 ssh、git）：不处理，原样返回。
//
// url.UserPassword 会对特殊字符做百分号编码，Token 中包含 @ / : 等字符也能正确处理。
func injectGitCredentials(repoURL, username, password string) (string, error) {
	if username == "" && password == "" {
		return repoURL, nil
	}
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return repoURL, nil
	}
	switch {
	case username != "" && password != "":
		u.User = url.UserPassword(username, password)
	case username != "":
		u.User = url.User(username)
	default: // 仅有 password/Token
		u.User = url.User(password)
	}
	return u.String(), nil
}

// maskCredentials 隐藏字符串中可能出现的 URL 内嵌凭证，避免写入日志泄露。
func maskCredentials(s string) string {
	return credInURLRe.ReplaceAllString(s, "$1***@")
}

// gitErrorMessage 依据 git 命令的错误与 stderr 输出，归类成可读的中文提示。
func gitErrorMessage(ctx context.Context, runErr error, stderr string) string {
	if ctx != nil && ctx.Err() == context.DeadlineExceeded {
		return "连接仓库超时：请检查网络连通性（内网仓库需确认后端进程/Pod 能访问该地址）"
	}
	low := strings.ToLower(stderr)
	switch {
	case strings.Contains(low, "authentication failed"),
		strings.Contains(low, "could not read username"),
		strings.Contains(low, "could not read password"),
		strings.Contains(low, "terminal prompts disabled"),
		strings.Contains(low, "403"):
		return "认证失败或需要凭证：私有仓库请提供正确的用户名/访问令牌（Token）"
	case strings.Contains(low, "repository not found"),
		strings.Contains(low, "not found"),
		strings.Contains(low, "404"):
		return "仓库不存在或无访问权限：请检查地址是否正确、账号是否有权限"
	case strings.Contains(low, "could not resolve host"),
		strings.Contains(low, "name or service not known"):
		return "无法解析仓库域名：请检查地址拼写（内网仓库需后端 DNS 能解析该域名）"
	case strings.Contains(low, "connection refused"),
		strings.Contains(low, "failed to connect"),
		strings.Contains(low, "could not connect"),
		strings.Contains(low, "timed out"),
		strings.Contains(low, "network is unreachable"),
		strings.Contains(low, "unable to access"):
		return "无法连接到仓库：请检查网络（内网仓库需确认后端与该地址网络互通、防火墙已放行）"
	case strings.Contains(low, "ssl"), strings.Contains(low, "certificate"):
		return "TLS/证书校验失败：自建仓库如使用自签证书，请为后端配置受信任的 CA"
	}
	msg := strings.TrimSpace(stderr)
	if msg == "" && runErr != nil {
		msg = runErr.Error()
	}
	if len(msg) > 300 {
		msg = msg[:300] + "..."
	}
	return "获取仓库信息失败: " + maskCredentials(msg)
}

// classifyGitError 将 git 命令错误转换为带明确原因的 error。
func classifyGitError(ctx context.Context, runErr error, stderr string) error {
	return fmt.Errorf("%s", gitErrorMessage(ctx, runErr, stderr))
}
