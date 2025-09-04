package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/pires/go-proxyproto"
	"github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
	"github.com/reiver/go-telnet/telsh"
	"github.com/soheilhy/cmux"

	"github.com/busyster996/go-hotfix"
	"github.com/busyster996/go-hotfix/example/handler"
	"github.com/busyster996/go-hotfix/example/symbols"
	"github.com/busyster996/go-hotfix/example/utils"
)

// 配置常量
const (
	DefaultListenPort = ":3333"
	ShutdownTimeout   = 30 * time.Second
	ReadHeaderTimeout = 10 * time.Second
	WriteTimeout      = 30 * time.Second
	IdleTimeout       = 120 * time.Second
)

// 命令描述
var commands = map[string]string{
	"help":   "Display a list of available commands.",
	"hotfix": "Load a hotfix script and apply the specified function.",
	"exit":   "Exit the shell.",
}

// Server 封装服务器配置和状态
type Server struct {
	listener net.Listener
	mux      cmux.CMux
	wg       sync.WaitGroup
	shutdown chan struct{}
}

// NewServer 创建新的服务器实例
func NewServer(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// 添加 keep-alive 和 proxy protocol 支持
	ln = &proxyproto.Listener{
		Listener: &utils.Keepalive{Listener: ln},
	}

	mux := cmux.New(ln)

	return &Server{
		listener: ln,
		mux:      mux,
		shutdown: make(chan struct{}),
	}, nil
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	// 配置HTTP服务器
	httpL := s.mux.Match(cmux.HTTP1Fast())
	httpServer := &http.Server{
		Handler:           handler.New(),
		ReadHeaderTimeout: ReadHeaderTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
	}

	// 启动HTTP服务
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		slog.Info("Starting HTTP server", "address", httpL.Addr())

		if err := httpServer.Serve(httpL); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	// 启动TCP/Telnet服务
	tcpL := s.mux.Match(cmux.Any())
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		slog.Info("Starting TCP server", "address", tcpL.Addr())

		if err := telnet.Serve(tcpL, s.newTelnetHandler()); err != nil {
			slog.Error("TCP server error", "error", err)
		}
	}()

	// 启动连接复用器
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.mux.Serve(); err != nil {
			slog.Error("Connection multiplexer error", "error", err)
		}
	}()

	// 等待关闭信号
	go s.handleShutdown(ctx, httpServer)

	slog.Info("Server started successfully", "address", s.listener.Addr())
	return nil
}

// handleShutdown 处理优雅关闭
func (s *Server) handleShutdown(ctx context.Context, httpServer *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		slog.Info("Received shutdown signal", "signal", sig)
	case <-ctx.Done():
		slog.Info("Context cancelled, shutting down")
	}

	close(s.shutdown)

	// 创建关闭超时上下文
	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	// 优雅关闭HTTP服务器
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Failed to shutdown HTTP server gracefully", "error", err)
	}

	// 关闭监听器
	if err := s.listener.Close(); err != nil {
		slog.Error("Failed to close listener", "error", err)
	}

	// 等待所有goroutine完成
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("Server shutdown completed")
	case <-shutdownCtx.Done():
		slog.Warn("Forced shutdown due to timeout")
	}
}

// newTelnetHandler 创建telnet处理器
func (s *Server) newTelnetHandler() telnet.Handler {
	shellHandler := telsh.NewShellHandler()

	// 注册help命令
	if err := shellHandler.RegisterHandlerFunc("help", s.handleHelpCommand); err != nil {
		slog.Error("Failed to register help command", "error", err)
	}

	// 注册hotfix命令
	if err := shellHandler.RegisterHandlerFunc("hotfix", s.handleHotfixCommand); err != nil {
		slog.Error("Failed to register hotfix command", "error", err)
	}

	return shellHandler
}

// handleHelpCommand 处理help命令
func (s *Server) handleHelpCommand(stdin io.ReadCloser, stdout io.WriteCloser, stderr io.WriteCloser, args ...string) error {
	if _, err := stdout.Write([]byte("Available commands:\n")); err != nil {
		return err
	}

	for command, desc := range commands {
		line := fmt.Sprintf("\t%s - %s\n", command, desc)
		if _, err := stdout.Write([]byte(line)); err != nil {
			return err
		}
	}

	_, err := oi.LongWrite(stdout, []byte("\r\n"))
	return err
}

// handleHotfixCommand 处理hotfix命令
func (s *Server) handleHotfixCommand(stdin io.ReadCloser, stdout io.WriteCloser, stderr io.WriteCloser, args ...string) error {
	if len(args) < 2 {
		usage := "用法: hotfix 脚本路径 脚本函数名\n例如: hotfix patch/test_yaegi_01.patch patch.PatchTest()\n"
		if _, err := stdout.Write([]byte(usage)); err != nil {
			return err
		}
		_, err := oi.LongWrite(stdout, []byte("\r\n"))
		return err
	}

	scriptPath, funcName := args[0], args[1]
	slog.Info("Applying hotfix", "script", scriptPath, "function", funcName)

	_, err := hotfix.ApplyFunc(scriptPath, funcName, symbols.Symbols)
	if err != nil {
		slog.Error("Hotfix failed", "error", err, "script", scriptPath, "function", funcName)
		errMsg := fmt.Sprintf("Hotfix错误: %v\r\n", err)
		_, writeErr := oi.LongWriteString(stdout, errMsg)
		if writeErr != nil {
			return writeErr
		}
		return nil
	}

	slog.Info("Hotfix applied successfully", "script", scriptPath, "function", funcName)
	_, err = oi.LongWriteString(stdout, "Hotfix应用成功\r\n")
	return err
}

// Wait 等待服务器关闭
func (s *Server) Wait() {
	s.wg.Wait()
}

func main() {
	// 设置恢复处理
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Panic recovered", "error", err, "stack", string(debug.Stack()))
			os.Exit(1)
		}
	}()

	// 获取监听地址（可以通过环境变量配置）
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = DefaultListenPort
	}

	// 创建服务器
	server, err := NewServer(addr)
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务器
	if err = server.Start(ctx); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	// 等待服务器关闭
	server.Wait()
	slog.Info("Application exited")
}
