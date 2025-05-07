package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init initializes the logger
func Init(level string, json bool) {
	logLevel := getLogLevel(level)
	logConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(logLevel),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    getEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if json {
		logConfig.Encoding = "json"
	}

	var err error
	log, err = logConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}
}

// getEncoderConfig returns the encoder config
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// getLogLevel converts a string log level to zapcore.Level
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Sync flushes any buffered log entries
func Sync() {
	_ = log.Sync()
}

// With creates a child logger with the added fields
func With(fields ...zap.Field) *zap.Logger {
	return log.With(fields...)
}

// WithContext returns a logger with request context fields
func WithContext(ctx context.Context) *zap.Logger {
	// Extract fields from context like request ID, user ID, etc.
	// Here's a simple example, but you can add more fields based on your context
	fields := []zap.Field{}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		fields = append(fields, zap.String("user_id", userID))
	}

	return log.With(fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Fatal logs a fatal message and then calls os.Exit(1)
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}

// DPanic logs a message and then panics (in development mode)
func DPanic(msg string, fields ...zap.Field) {
	log.DPanic(msg, fields...)
}
