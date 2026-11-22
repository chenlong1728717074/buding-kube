package api

import (
	"buding-kube/internal/kube"
	"buding-kube/internal/service"
	"buding-kube/internal/web/dto"
	"buding-kube/pkg/logs"
	"buding-kube/pkg/utils/jwt"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	// terminalSubprotocol 握手时服务端回应的子协议
	terminalSubprotocol = "buding-terminal"
	// terminalTokenPrefix 浏览器的 WebSocket 不能设置请求头, 令牌放在子协议 "bearer.<token>" 中, 不出现在地址和访问日志里
	terminalTokenPrefix  = "bearer."
	terminalPingInterval = 30 * time.Second
	terminalReadTimeout  = 90 * time.Second
	terminalWriteTimeout = 10 * time.Second
	// terminalIdleTimeout 长时间没有输入时自动断开, 避免忘记关闭的终端一直占用连接
	terminalIdleTimeout = 30 * time.Minute
	terminalReadLimit   = 64 * 1024

	closeCodeUnauthorized = 4401
	closeCodeForbidden    = 4403
	closeCodeError        = 4500
)

var terminalUpgrader = websocket.Upgrader{
	ReadBufferSize:  4 * 1024,
	WriteBufferSize: 32 * 1024,
	Subprotocols:    []string{terminalSubprotocol},
	// 令牌放在子协议中, 浏览器不会自动携带, 其他网站的页面拿不到令牌, 不存在跨站劫持的问题;
	// 前端与后端通常不同源(开发环境 5173 -> 8888), 因此不校验 Origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TerminalApi 容器终端(WebSocket). 认证在握手后进行, 这样失败原因可以告诉前端, 因此路由组上不挂 JWTAuth
type TerminalApi struct {
	router *gin.RouterGroup
	srv    *service.PodService
	BaseApi
}

func NewTerminalApi(router *gin.RouterGroup) *TerminalApi {
	api := TerminalApi{
		router: router,
		srv:    service.GetSingletonPodService(),
	}
	api.Router()
	return &api
}

func (a *TerminalApi) Router() {
	a.router.GET("/pod", a.PodTerminal)
	a.router.GET("/kubectl", a.KubectlTerminal)
}

// terminalMessage 文本帧中的控制消息. 容器输出使用二进制帧原样发送(可能截断 UTF-8 字符, 由前端终端拼接)
type terminalMessage struct {
	// 前端发送: stdin / resize; 后端发送: info / ready / error / exit
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
}

// accept 完成握手并校验令牌; 返回 false 时连接已关闭
func (a *TerminalApi) accept(ctx *gin.Context) (*terminalSession, *jwt.Claims, bool) {
	conn, err := terminalUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		// Upgrade 失败时已经写出了错误响应
		logs.Warn("终端握手失败: %v", err)
		return nil, nil, false
	}
	session := newTerminalSession(conn)
	claims, err := jwt.Token.ParseToken(terminalToken(ctx.Request))
	if err != nil {
		session.fail(closeCodeUnauthorized, "登录已过期，请重新登录")
		_ = conn.Close()
		return nil, nil, false
	}
	if os.Getenv("APP_ENV") == "demo" {
		session.fail(closeCodeForbidden, "演示环境不允许使用终端")
		_ = conn.Close()
		return nil, nil, false
	}
	return session, claims, true
}

// PodTerminal 进入容器终端
func (a *TerminalApi) PodTerminal(ctx *gin.Context) {
	var query dto.PodTerminalDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	session, claims, ok := a.accept(ctx)
	if !ok {
		return
	}
	defer session.conn.Close()
	a.serve(ctx, session, claims, query, "容器终端")
}

// KubectlTerminal 集群级 kubectl 终端, 只允许管理员使用(拥有 cluster-admin 权限)
func (a *TerminalApi) KubectlTerminal(ctx *gin.Context) {
	var query dto.KubectlTerminalDTO
	if err := a.BindQuery(ctx, &query); err != nil {
		a.ParamBindError(ctx, err)
		return
	}
	session, claims, ok := a.accept(ctx)
	if !ok {
		return
	}
	defer session.conn.Close()
	if level := kube.RoleLevel(claims.Role); level == 0 || level > kube.RoleLevel(kube.RoleAdmin) {
		session.fail(closeCodeForbidden, "Web kubectl 拥有集群管理员权限，只允许管理员使用")
		return
	}
	_ = session.send(terminalMessage{Type: "info", Data: "正在准备 kubectl 环境…"})
	prepareCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	namespace, name, err := service.EnsureKubectlPod(prepareCtx, query.ClusterId, claims.Username, func(msg string) {
		_ = session.send(terminalMessage{Type: "info", Data: msg})
	})
	if err != nil {
		logs.Warn("准备 kubectl 环境失败: %s %v", query.ClusterId, err)
		session.fail(closeCodeError, err.Error())
		return
	}
	a.serve(ctx, session, claims, dto.PodTerminalDTO{
		ClusterId: query.ClusterId,
		Namespace: namespace,
		Name:      name,
		Container: service.KubectlContainer,
		Shell:     "auto",
	}, "kubectl 终端")
}

// serve 连接容器并转发输入输出, 直到 Shell 退出或浏览器断开
func (a *TerminalApi) serve(ctx *gin.Context, session *terminalSession, claims *jwt.Claims, query dto.PodTerminalDTO, label string) {
	execCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	terminal, err := a.srv.OpenTerminal(execCtx, query)
	if err != nil {
		session.fail(closeCodeError, err.Error())
		return
	}
	_ = session.send(terminalMessage{Type: "ready", Data: terminal.Container})

	go session.readLoop(cancel)
	go session.keepAlive(execCtx, cancel)

	target := fmt.Sprintf("%s/%s/%s/%s", query.ClusterId, query.Namespace, query.Name, terminal.Container)
	start := time.Now()
	// 终端可以执行任意命令, 记录审计日志
	logs.Info("用户 %s 进入%s %s (shell=%s, ip=%s)", claims.Username, label, target, terminal.Shell, ctx.ClientIP())
	err = terminal.Stream(execCtx, session, session, session)
	logs.Info("用户 %s 退出%s %s, 时长 %s", claims.Username, label, target, time.Since(start).Round(time.Second))

	if execCtx.Err() != nil {
		// 浏览器已断开或空闲超时, 不需要再通知
		return
	}
	session.finish(err, terminal.Shell)
}

// terminalToken 从子协议中取出令牌
func terminalToken(r *http.Request) string {
	for _, p := range websocket.Subprotocols(r) {
		if strings.HasPrefix(p, terminalTokenPrefix) {
			return strings.TrimPrefix(p, terminalTokenPrefix)
		}
	}
	return ""
}

// terminalSession 把 WebSocket 适配为执行器需要的 stdin(io.Reader)、stdout(io.Writer) 和终端尺寸队列
type terminalSession struct {
	conn      *websocket.Conn
	writeMu   sync.Mutex
	stdinR    *io.PipeReader
	stdinW    *io.PipeWriter
	sizeCh    chan remotecommand.TerminalSize
	done      chan struct{}
	closeOnce sync.Once
	lastInput atomic.Int64
}

func newTerminalSession(conn *websocket.Conn) *terminalSession {
	r, w := io.Pipe()
	s := &terminalSession{
		conn:   conn,
		stdinR: r,
		stdinW: w,
		sizeCh: make(chan remotecommand.TerminalSize, 1),
		done:   make(chan struct{}),
	}
	s.lastInput.Store(time.Now().UnixNano())
	return s
}

// Read 容器的标准输入
func (s *terminalSession) Read(p []byte) (int, error) {
	return s.stdinR.Read(p)
}

// Write 容器的输出, 以二进制帧发送
func (s *terminalSession) Write(p []byte) (int, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = s.conn.SetWriteDeadline(time.Now().Add(terminalWriteTimeout))
	if err := s.conn.WriteMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Next 终端尺寸变化; 返回 nil 表示会话结束
func (s *terminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-s.sizeCh:
		return &size
	case <-s.done:
		return nil
	}
}

func (s *terminalSession) send(msg terminalMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = s.conn.SetWriteDeadline(time.Now().Add(terminalWriteTimeout))
	return s.conn.WriteMessage(websocket.TextMessage, data)
}

// pushSize 只保留最新的尺寸, 拖动窗口时不会堆积
func (s *terminalSession) pushSize(size remotecommand.TerminalSize) {
	for {
		select {
		case s.sizeCh <- size:
			return
		case <-s.done:
			return
		default:
			select {
			case <-s.sizeCh:
			default:
			}
		}
	}
}

func (s *terminalSession) close() {
	s.closeOnce.Do(func() {
		close(s.done)
		_ = s.stdinW.Close()
	})
}

// readLoop 读取浏览器发来的输入; 连接断开时结束会话
func (s *terminalSession) readLoop(cancel context.CancelFunc) {
	defer func() {
		s.close()
		cancel()
	}()
	s.conn.SetReadLimit(terminalReadLimit)
	_ = s.conn.SetReadDeadline(time.Now().Add(terminalReadTimeout))
	s.conn.SetPongHandler(func(string) error {
		return s.conn.SetReadDeadline(time.Now().Add(terminalReadTimeout))
	})
	for {
		msgType, data, err := s.conn.ReadMessage()
		if err != nil {
			return
		}
		_ = s.conn.SetReadDeadline(time.Now().Add(terminalReadTimeout))
		if msgType != websocket.TextMessage {
			continue
		}
		var msg terminalMessage
		if json.Unmarshal(data, &msg) != nil {
			continue
		}
		switch msg.Type {
		case "stdin":
			s.lastInput.Store(time.Now().UnixNano())
			if _, err = s.stdinW.Write([]byte(msg.Data)); err != nil {
				return
			}
		case "resize":
			if msg.Cols > 0 && msg.Rows > 0 {
				s.pushSize(remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows})
			}
		}
	}
}

// keepAlive 定时发送 ping 保持连接(浏览器会自动回复 pong), 并在长时间没有输入时断开
func (s *terminalSession) keepAlive(ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(terminalPingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if time.Since(time.Unix(0, s.lastInput.Load())) > terminalIdleTimeout {
				_ = s.send(terminalMessage{Type: "error", Data: fmt.Sprintf("超过 %d 分钟没有操作，终端已断开", int(terminalIdleTimeout.Minutes()))})
				s.closeWith(websocket.CloseNormalClosure, "idle timeout")
				cancel()
				return
			}
			if err := s.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(terminalWriteTimeout)); err != nil {
				cancel()
				return
			}
		}
	}
}

func (s *terminalSession) closeWith(code int, reason string) {
	_ = s.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(terminalWriteTimeout))
}

// fail 发送错误后关闭连接
func (s *terminalSession) fail(code int, message string) {
	_ = s.send(terminalMessage{Type: "error", Data: message})
	s.closeWith(code, "error")
}

// finish Shell 结束: 正常退出时告诉前端退出码, 连接失败时给出原因
func (s *terminalSession) finish(err error, shell string) {
	if code, ok := service.TerminalExitStatus(err); ok {
		// exec bash 找不到命令时 sh 以 127 退出
		if code == 127 && shell == "bash" {
			s.fail(closeCodeError, "容器中没有 bash，请切换为 sh")
			return
		}
		_ = s.send(terminalMessage{Type: "exit", Data: fmt.Sprint(code)})
		s.closeWith(websocket.CloseNormalClosure, "exit")
		return
	}
	logs.Warn("终端连接失败: %v", err)
	s.fail(closeCodeError, service.TerminalErrorMessage(err))
}
