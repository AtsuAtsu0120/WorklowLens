package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/kaido-atsuya/workflow_lens_server/internal/handler"
	"github.com/kaido-atsuya/workflow_lens_server/internal/store"
	"github.com/kaido-atsuya/workflow_lens_server/internal/telemetry"
)

func main() {
	// 環境変数の読み取り
	driver := os.Getenv("STORE_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		switch driver {
		case "sqlite", "sqlite3":
			dsn = "workflowlens.db"
		default:
			slog.Error("DATABASE_URL is required for driver " + driver)
			os.Exit(1)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// DB接続
	ctx := context.Background()
	s, err := store.NewSQLStore(ctx, driver, dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer s.Close()

	slog.Info("database connected", "driver", driver)

	// OpenTelemetry初期化
	otelShutdown, err := telemetry.Init(ctx)
	if err != nil {
		slog.Warn("OpenTelemetry初期化失敗（テレメトリ無効）", "error", err)
	} else {
		defer otelShutdown(context.Background())
	}

	// ルーティング
	mux := http.NewServeMux()
	mux.HandleFunc("POST /logs", handler.HandlePostLogs(s))
	mux.HandleFunc("GET /health", handler.HandleHealth())

	// otelhttp でHTTPリクエストを自動計装
	otelHandler := otelhttp.NewMiddleware("workflow_lens_server")(mux)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      otelHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigCh
		slog.Info("shutting down", "signal", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	slog.Info("server starting", "port", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}
