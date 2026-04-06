package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kaido-atsuya/tool_logger_middleware/internal/lock"
	"github.com/kaido-atsuya/tool_logger_middleware/internal/server"
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

	// 4. UDPサーバー起動
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if err := server.Run(ctx, addr); err != nil {
		slog.Error("サーバーエラー", "error", err)
		os.Exit(1)
	}
}
