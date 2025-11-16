package logger

import (
	"context"
	"fmt"
	//"os"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

type Logger struct {
	zap         *zap.Logger
	serviceName string
	env         string
}

func New(env string, serviceName string) (*Logger, error) {
	var cfg zap.Config

	switch env {
	case "development", "staging", "production":
		cfg = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:      env == "development",
			Encoding:         "json", // JSON logs for Loki/Promtail
			//OutputPaths:      []string{"stdout", fmt.Sprintf("./logs/%s.log", serviceName)},
			//ErrorOutputPaths: []string{"stderr", fmt.Sprintf("./logs/%s.log", serviceName)},
			EncoderConfig: zapcore.EncoderConfig{
				TimeKey:        "timestamp",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				MessageKey:     "message",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		}
	default:
		return nil, fmt.Errorf("unknown environment: %s", env)
	}

	// logDir := "./logs"
	// if _, err := os.Stat(logDir); os.IsNotExist(err) {
	// 	if err := os.MkdirAll(logDir, 0755); err != nil {
	// 		return nil, fmt.Errorf("failed to create log directory: %w", err)
	// 	}
	// }	

	// // Absolute path for log file per service
	// logFile := fmt.Sprintf("%s/%s.log", logDir, serviceName)
	// cfg.OutputPaths = []string{"stdout", logFile}
	// cfg.ErrorOutputPaths = []string{"stderr", logFile}
cfg.OutputPaths = []string{"stdout"}
cfg.ErrorOutputPaths = []string{"stderr"}		
	zapLogger, err := cfg.Build()
	if err != nil {
		fmt.Println("logger: failed in cfg build")
		return nil, err
	}

	// ✅ Redirect standard library log to Zap
	zap.RedirectStdLog(zapLogger)
	return &Logger{zap: zapLogger, serviceName: serviceName, env: env}, nil
}

// Info logs info-level messages with structured fields
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields,
		zap.String("service", l.serviceName),
		zap.String("env", l.env),
	)
	l.injectTraceIDs(ctx, &fields)
	l.zap.Info(msg, fields...)
}

// Error logs error-level messages with structured fields
func (l *Logger) Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	fields = append(fields,
		zap.String("service", l.serviceName),
		zap.String("env", l.env),
		zap.Error(err),
	)
	l.injectTraceIDs(ctx, &fields)
	l.zap.Error(msg, fields...)
}

// Trace logs a message and creates an OTEL span if context is provided
func (l *Logger) Trace(ctx context.Context, msg string, fields ...zap.Field) {
	tracer := otel.Tracer("logger")
	ctx, span := tracer.Start(ctx, msg)
	defer span.End()

	fields = append(fields,
		zap.String("service", l.serviceName),
		zap.String("env", l.env),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
	)
	l.injectTraceIDs(ctx, &fields)
	l.zap.Info(msg, fields...)

	span.SetAttributes(
		attribute.String("log.message", msg),
		attribute.String("service", l.serviceName),
		attribute.String("env", l.env),
	)
}

// Sync flushes buffered logs
func (l *Logger) Sync() {
	_ = l.zap.Sync()
}

// injectTraceIDs adds OTEL trace_id and span_id to log fields if available
func (l *Logger) injectTraceIDs(ctx context.Context, fields *[]zap.Field) {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return
	}
	sc := span.SpanContext()
	*fields = append(*fields,
		zap.String("trace_id", sc.TraceID().String()),
		zap.String("span_id", sc.SpanID().String()),
	)
}
