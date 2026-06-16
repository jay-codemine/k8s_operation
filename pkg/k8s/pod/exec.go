package pod

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// TerminalMessage 前后端通信消息协议
type TerminalMessage struct {
	Op   string `json:"op"`   // stdin, stdout, stderr, resize, ping, pong
	Data string `json:"data"` // 消息体（文本内容）
	Rows uint16 `json:"rows"` // resize 时的行数
	Cols uint16 `json:"cols"` // resize 时的列数
}

// WebSocketTerminal 实现 remotecommand.TerminalSizeQueue 和 io.Reader/Writer
// 桥接 WebSocket 和 K8s exec SPDY 流
type WebSocketTerminal struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	sizeChan  chan remotecommand.TerminalSize
	doneChan  chan struct{}
	closeOnce sync.Once
}

// NewWebSocketTerminal 创建终端实例
func NewWebSocketTerminal(conn *websocket.Conn) *WebSocketTerminal {
	return &WebSocketTerminal{
		conn:     conn,
		sizeChan: make(chan remotecommand.TerminalSize, 1),
		doneChan: make(chan struct{}),
	}
}

// Read 从 WebSocket 读取用户输入，实现 io.Reader
func (t *WebSocketTerminal) Read(p []byte) (int, error) {
	for {
		select {
		case <-t.doneChan:
			return 0, io.EOF
		default:
		}

		_, message, err := t.conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			// 非 JSON 数据直接当 stdin
			n := copy(p, message)
			return n, nil
		}

		switch msg.Op {
		case "stdin":
			n := copy(p, msg.Data)
			return n, nil
		case "resize":
			t.sizeChan <- remotecommand.TerminalSize{
				Width:  msg.Cols,
				Height: msg.Rows,
			}
		case "ping":
			t.mu.Lock()
			_ = t.conn.WriteJSON(TerminalMessage{Op: "pong"})
			t.mu.Unlock()
		}
		// 其他 op 继续循环
	}
}

// Write 将输出写入 WebSocket，实现 io.Writer
func (t *WebSocketTerminal) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	msg := TerminalMessage{
		Op:   "stdout",
		Data: string(p),
	}
	if err := t.conn.WriteJSON(msg); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Next 实现 remotecommand.TerminalSizeQueue
func (t *WebSocketTerminal) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	case <-t.doneChan:
		return nil
	}
}

// Close 停止终端（关闭 doneChan 停止心跳和 Read 循环），但不关闭 WebSocket 连接
// WebSocket 连接的生命周期由调用方（Terminal handler）管理
func (t *WebSocketTerminal) Close() {
	t.closeOnce.Do(func() {
		close(t.doneChan)
	})
}

// ShellCandidate 候选 shell 信息
type ShellCandidate struct {
	Path    string // shell 路径
	Display string // 显示名称
}

// AllShellCandidates 返回所有候选 shell 列表（按优先级排序）
func AllShellCandidates() []ShellCandidate {
	return []ShellCandidate{
		{"/bin/bash", "bash"},
		{"/bin/sh", "sh"},
		{"/bin/ash", "ash (alpine)"},
		{"bash", "bash"},
		{"sh", "sh"},
		{"ash", "ash"},
		{"/bin/zsh", "zsh"},
		{"zsh", "zsh"},
	}
}

// ExecInPod 通过 WebSocket 在 Pod 中执行命令（交互式 shell）
// 不再内部发送 Connected 消息，由调用方控制
func ExecInPod(
	ctx context.Context,
	config *rest.Config,
	kube kubernetes.Interface,
	namespace, podName, containerName string,
	shell string,
	wsConn *websocket.Conn,
) error {
	if shell == "" {
		shell = "sh"
	}

	// 构建 exec 请求
	req := kube.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{shell},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	// 创建 SPDY executor
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("create SPDY executor: %w", err)
	}

	// 桥接 WebSocket ↔ K8s exec
	terminal := NewWebSocketTerminal(wsConn)
	defer terminal.Close()

	// 启动心跳检测
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				terminal.mu.Lock()
				err := wsConn.WriteJSON(TerminalMessage{Op: "ping"})
				terminal.mu.Unlock()
				if err != nil {
					return
				}
			case <-terminal.doneChan:
				return
			}
		}
	}()

	// 执行 exec stream
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             terminal,
		Stdout:            terminal,
		Stderr:            terminal,
		Tty:               true,
		TerminalSizeQueue: terminal,
	})

	return err
}

// IsShellNotFoundErr 判断是否是 shell 不存在的错误
func IsShellNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "executable file not found") ||
		strings.Contains(s, "no such file or directory") ||
		strings.Contains(s, "oci runtime exec failed") ||
		strings.Contains(s, "not found in")
}

// DetectShell 检测容器中可用的 shell（带超时控制），返回 (shell路径, 是否找到)
func DetectShell(
	ctx context.Context,
	config *rest.Config,
	kube kubernetes.Interface,
	namespace, podName, containerName string,
) (string, bool) {
	for _, c := range AllShellCandidates() {
		// 每个 shell 检测最多 3 秒
		detectCtx, cancel := context.WithTimeout(ctx, 3*time.Second)

		req := kube.CoreV1().RESTClient().Post().
			Resource("pods").
			Name(podName).
			Namespace(namespace).
			SubResource("exec").
			VersionedParams(&corev1.PodExecOptions{
				Container: containerName,
				Command:   []string{c.Path, "-c", "echo ok"},
				Stdout:    true,
				Stderr:    true,
			}, scheme.ParameterCodec)

		exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
		if err != nil {
			cancel()
			continue
		}

		err = exec.StreamWithContext(detectCtx, remotecommand.StreamOptions{
			Stdout: io.Discard,
			Stderr: io.Discard,
		})
		cancel()
		if err == nil {
			return c.Path, true
		}
	}
	return "", false
}
