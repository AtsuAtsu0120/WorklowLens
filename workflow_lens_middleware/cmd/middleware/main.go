package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kaido-atsuya/workflow_lens_middleware/internal/forwarder"
	"github.com/kaido-atsuya/workflow_lens_middleware/internal/lock"
	"github.com/kaido-atsuya/workflow_lens_middleware/internal/server"
	"github.com/kaido-atsuya/workflow_lens_middleware/internal/telemetry"
)

const defaultPort = 59100

func main() {
	// 1. 多重起動防止ロックを取得
	instanceLock, err := lock.Acquire()
	if err != nil {
		slog.Error("起動エラー", "error", err)
		os.Exit(1)
	}
	defer instanceLock.Close()

	// 2. コマンドライン引数からポートを取得
	port := defaultPort
	if len(os.Args) > 1 {
		p, err := strconv.Atoi(os.Args[1])
		if err != nil {
			slog.Error("ポート番号が不正です", "value", os.Args[1])
			os.Exit(1)
		}
		port = p
	}

	// 3. シグナルハンドリング（SIGTERM, SIGINT）
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// 4. OpenTelemetry初期化
	otelShutdown, err := telemetry.Init(ctx)
	if err != nil {
		slog.Warn("OpenTelemetry初期化失敗（テレメトリ無効）", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	// 5. HTTP転送設定
	var fwd *forwarder.Forwarder
	if serverURL := os.Getenv("WORKFLOW_LENS_SERVER_URL"); serverURL != "" {
		fwd = forwarder.New(serverURL)
		go fwd.Run(ctx)
		defer fwd.Flush(context.Background())
		slog.Info("HTTP転送有効", "server_url", serverURL)
	}

	// 6. UDPサーバー起動
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if err := server.Run(ctx, addr, fwd); err != nil {
		slog.Error("サーバーエラー", "error", err)
		os.Exit(1)
	}
}
