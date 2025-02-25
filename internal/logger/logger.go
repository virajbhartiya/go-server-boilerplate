package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

var (
	globalLogger *zap.Logger
	levelMap     = map[string]zapcore.Level{
		"debug":  zapcore.DebugLevel,
		"info":   zapcore.InfoLevel,
		"warn":   zapcore.WarnLevel,
		"error":  zapcore.ErrorLevel,
		"dpanic": zapcore.DPanicLevel,
		"panic":  zapcore.PanicLevel,
		"fatal":  zapcore.FatalLevel,
	}
)

// Init initializes the global logger
func Init(level string, jsonFormat bool) {
	var cfg zap.Config

	if jsonFormat {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg.Level = zap.NewAtomicLevelAt(getLogLevel(level))
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	var err error
	globalLogger, err = cfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		panic(err)
	}
}

// WithContext creates a new context with the logger
func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, globalLogger)
}

// FromContext retrieves the logger from context
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return globalLogger
	}
	if logger, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return logger
	}
	return globalLogger
}

// With creates a child logger with additional fields
func With(fields ...zapcore.Field) *zap.Logger {
	return globalLogger.With(fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zapcore.Field) {
	globalLogger.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zapcore.Field) {
	globalLogger.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zapcore.Field) {
	globalLogger.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zapcore.Field) {
	globalLogger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zapcore.Field) {
	globalLogger.Fatal(msg, fields...)
}

// WithRequestID adds request ID to the logger
func WithRequestID(requestID string) *zap.Logger {
	return globalLogger.With(zap.String("request_id", requestID))
}

// WithError adds error to the logger
func WithError(err error) *zap.Logger {
	return globalLogger.With(zap.Error(err))
}

func getLogLevel(level string) zapcore.Level {
	if l, ok := levelMap[level]; ok {
		return l
	}
	return zapcore.InfoLevel
}

// Sync flushes any buffered log entries
func Sync() error {
	return globalLogger.Sync()
}
