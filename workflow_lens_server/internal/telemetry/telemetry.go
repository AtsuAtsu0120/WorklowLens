// Package telemetry はOpenTelemetryの初期化・終了を提供する。
// OTEL_EXPORTER_OTLP_ENDPOINT が未設定の場合はno-opプロバイダーを使用し、
// 既存の動作に影響を与えない。
package telemetry

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

const defaultServiceName = "workflow_lens_server"

// Init はOpenTelemetryのTracerProviderとMeterProviderを初期化する。
// OTEL_EXPORTER_OTLP_ENDPOINT未設定またはOTEL_SDK_DISABLED=trueの場合、
// no-opプロバイダーを使用しshutdownは何もしない。
func Init(ctx context.Context) (shutdown func(context.Context), err error) {
	noop := func(context.Context) {}

	// OTEL_SDK_DISABLED=true の場合は無効
	if os.Getenv("OTEL_SDK_DISABLED") == "true" {
		slog.Info("OpenTelemetry無効（OTEL_SDK_DISABLED=true）")
		return noop, nil
	}

	// OTEL_EXPORTER_OTLP_ENDPOINT 未設定の場合はno-op
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		return noop, nil
	}

	// リソース定義
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	)
	if err != nil {
		return noop, err
	}

	// トレースエクスポーター（OTLP gRPC）
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return noop, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// W3C Trace Context プロパゲーター
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// メトリクスエクスポーター（OTLP gRPC）
	metricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		// トレースだけでも動作させる
		slog.Warn("メトリクスエクスポーター初期化失敗", "error", err)
		return func(ctx context.Context) {
			tp.Shutdown(ctx)
		}, nil
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	slog.Info("OpenTelemetry初期化完了", "service", serviceName)

	return func(ctx context.Context) {
		tp.Shutdown(ctx)
		mp.Shutdown(ctx)
	}, nil
}
