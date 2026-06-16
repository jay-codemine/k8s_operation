package pod

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"go.uber.org/zap"

	"k8soperation/global"
	"k8soperation/internal/app/services"
	"k8soperation/middlewares"
	"k8soperation/pkg/k8s/pod"
)

var upgrader = websocket.Upgrader{
	// 允许跨域 WebSocket 连接
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// 发送终端消息辅助函数
func wsSend(ws *websocket.Conn, text string) {
	_ = ws.WriteJSON(pod.TerminalMessage{Op: "stdout", Data: text})
}

// Terminal godoc
// @Summary     容器终端（WebSocket）
// @Description 通过 WebSocket 连接到 Pod 容器的交互式终端（类似 kubectl exec -it）
// @Tags        K8s Pod管理
// @Param       namespace query string true  "命名空间"
// @Param       name      query string true  "Pod 名称"
// @Param       container query string false "容器名称（多容器 Pod 建议指定）"
// @Param       shell     query string false "Shell 类型（bash/sh/zsh，默认自动检测）"
// @Success     101       "Switching Protocols - WebSocket 握手成功"
// @Failure     400       {object} map[string]interface{} "参数错误"
// @Failure     500       {object} map[string]interface{} "连接失败"
// @Router      /api/v1/k8s/pod/terminal [get]
func (c *PodController) Terminal(ctx *gin.Context) {
	namespace := ctx.Query("namespace")
	podName := ctx.Query("name")
	container := ctx.Query("container")
	shell := ctx.Query("shell")

	if namespace == "" || podName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": 1,
			"msg":  "namespace 和 name 参数必填",
		})
		return
	}

	// 获取 K8s 客户端
	cli := middlewares.MustGetK8sClients(ctx)

	// 如果没指定容器名，获取第一个容器
	if container == "" {
		podObj, err := cli.Kube.CoreV1().Pods(namespace).Get(ctx.Request.Context(), podName, metav1.GetOptions{})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": 1,
				"msg":  "获取 Pod 信息失败: " + err.Error(),
			})
			return
		}
		if len(podObj.Spec.Containers) > 0 {
			container = podObj.Spec.Containers[0].Name
		}
	}

	// WebSocket 协议升级
	wsConn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		global.Logger.Error("WebSocket 升级失败",
			zap.String("pod", podName),
			zap.Error(err),
		)
		return
	}
	defer wsConn.Close()

	global.Logger.Info("容器终端已连接",
		zap.String("namespace", namespace),
		zap.String("pod", podName),
		zap.String("container", container),
	)

	// 发送初始状态
	wsSend(wsConn, fmt.Sprintf("\x1b[1;36m↻ 正在连接 %s/%s [%s]...\x1b[0m\r\n", namespace, podName, container))

	// ========================
	// 策略一：用户指定了 shell，直接尝试
	// ========================
	if shell != "" {
		wsSend(wsConn, fmt.Sprintf("\x1b[90m  → 使用指定 shell: %s\x1b[0m\r\n", shell))
		execErr := pod.ExecInPod(
			ctx.Request.Context(), cli.Config, cli.Kube,
			namespace, podName, container, shell, wsConn,
		)
		if execErr != nil {
			global.Logger.Warn("指定 shell 执行失败",
				zap.String("pod", podName),
				zap.String("shell", shell),
				zap.Error(execErr),
			)
			wsSend(wsConn, fmt.Sprintf("\r\n\x1b[1;31m✗ %s 执行失败: %s\x1b[0m\r\n", shell, execErr.Error()))
		}
		return
	}

	// ========================
	// 策略二：自动检测 + 多轮回退
	// ========================
	wsSend(wsConn, "\x1b[90m  → 正在自动检测可用 Shell...\x1b[0m\r\n")

	detectedShell, found := pod.DetectShell(
		ctx.Request.Context(), cli.Config, cli.Kube,
		namespace, podName, container,
	)

	if found {
		global.Logger.Info("自动检测 shell 成功",
			zap.String("shell", detectedShell),
			zap.String("pod", podName),
		)
		wsSend(wsConn, fmt.Sprintf("\x1b[1;32m✓ 检测到 %s\x1b[0m\r\n", detectedShell))
		wsSend(wsConn, fmt.Sprintf("\x1b[1;32m✓ Connected to %s/%s [%s]\x1b[0m\r\n\r\n", namespace, podName, container))

		// 执行 exec stream
		execErr := pod.ExecInPod(
			ctx.Request.Context(), cli.Config, cli.Kube,
			namespace, podName, container, detectedShell, wsConn,
		)
		if execErr != nil {
			global.Logger.Warn("容器终端断开",
				zap.String("pod", podName),
				zap.String("shell", detectedShell),
				zap.Error(execErr),
			)
			// 如果连接后又失败，可能是容器重启了
			if pod.IsShellNotFoundErr(execErr) {
				wsSend(wsConn, fmt.Sprintf("\r\n\x1b[1;31m✗ %s 执行失败\x1b[0m\r\n", detectedShell))
				// 尝试回退到其他 shell
				c.tryFallbackShells(ctx, cli, wsConn, namespace, podName, container, detectedShell)
			}
		}
		global.Logger.Info("容器终端已断开",
			zap.String("namespace", namespace),
			zap.String("pod", podName),
		)
		return
	}

	// ========================
	// 检测失败，尝试直接 exec 每个 shell
	// ========================
	global.Logger.Warn("shell 检测未找到，尝试直接 exec",
		zap.String("pod", podName),
	)
	wsSend(wsConn, "\x1b[1;33m⚠ 快速检测未找到 Shell，尝试逐个 exec...\x1b[0m\r\n")

	c.tryFallbackShells(ctx, cli, wsConn, namespace, podName, container, "")
}

// tryFallbackShells 用非交互式快速检测逐个尝试候选 shell，找到可用的再进入交互终端
// 全部失败则发送友好提示并优雅退出
func (c *PodController) tryFallbackShells(
	ctx *gin.Context,
	cli *services.K8sClients,
	wsConn *websocket.Conn,
	namespace, podName, container, skip string,
) {
	candidates := pod.AllShellCandidates()
	seen := map[string]bool{}
	if skip != "" {
		seen[skip] = true
	}

	for _, sc := range candidates {
		if seen[sc.Path] {
			continue
		}
		seen[sc.Path] = true

		wsSend(wsConn, fmt.Sprintf("\x1b[90m  → 尝试 %s (%s)...\x1b[0m\r\n", sc.Path, sc.Display))

		// 用非交互式 exec 快速检测（5秒超时），不会阻塞 WebSocket
		available := c.quickTestShell(ctx.Request.Context(), cli, namespace, podName, container, sc.Path)

		if available {
			// 检测通过，进入交互式终端
			wsSend(wsConn, fmt.Sprintf("\x1b[1;32m  ✓ %s 可用，正在连接...\x1b[0m\r\n", sc.Path))
			wsSend(wsConn, fmt.Sprintf("\x1b[1;32m✓ Connected to %s/%s [%s]\x1b[0m\r\n\r\n", namespace, podName, container))

			execErr := pod.ExecInPod(
				ctx.Request.Context(), cli.Config, cli.Kube,
				namespace, podName, container, sc.Path, wsConn,
			)
			if execErr != nil {
				global.Logger.Warn("容器终端断开",
					zap.String("pod", podName),
					zap.String("shell", sc.Path),
					zap.Error(execErr),
				)
			}
			return
		}

		wsSend(wsConn, fmt.Sprintf("\x1b[90m    ✗ %s 不可用\x1b[0m\r\n", sc.Path))
	}

	// 所有 shell 都失败 —— 发送友好提示
	wsSend(wsConn, "\r\n\x1b[1;31m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\x1b[0m\r\n")
	wsSend(wsConn, "\x1b[1;31m  ✗ 所有 Shell 均不可用\x1b[0m\r\n")
	wsSend(wsConn, "\x1b[1;33m  ℹ 该容器可能是 distroless / scratch / static 镜像\x1b[0m\r\n")
	wsSend(wsConn, "\x1b[1;33m    这类镜像不包含 Shell，无法进入交互式终端\x1b[0m\r\n")
	wsSend(wsConn, "\x1b[90m  Tip: 可尝试 kubectl debug 创建临时调试容器\x1b[0m\r\n")
	wsSend(wsConn, "\x1b[1;31m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\x1b[0m\r\n")

	global.Logger.Warn("容器无可用 shell，终端无法进入",
		zap.String("namespace", namespace),
		zap.String("pod", podName),
		zap.String("container", container),
	)
}

// quickTestShell 非交互式快速检测指定 shell 是否可用（5秒超时）
func (c *PodController) quickTestShell(
	parentCtx context.Context,
	cli *services.K8sClients,
	namespace, podName, container, shell string,
) bool {
	testCtx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	req := cli.Kube.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{shell, "-c", "echo ok"},
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cli.Config, "POST", req.URL())
	if err != nil {
		return false
	}

	err = exec.StreamWithContext(testCtx, remotecommand.StreamOptions{
		Stdout: io.Discard,
		Stderr: io.Discard,
	})
	if err != nil {
		global.Logger.Debug("quickTestShell 失败",
			zap.String("shell", shell),
			zap.Error(err),
		)
		return false
	}
	return true
}
