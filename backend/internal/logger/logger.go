package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New constructs a zap logger based on the provided environment and level.
func New(env, level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if strings.EqualFold(env, "development") {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if level != "" {
		lc := strings.ToLower(level)
		switch lc {
		case "debug":
			cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		case "info":
			cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		case "warn":
			cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
		case "error":
			cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
		default:
			cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		}
	}

	return cfg.Build()
}
