package utils

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CreateLogger constructs a logger with the given log level and formats the output for humans when developer mode is
// enabled.
func CreateLogger(logLevel int, enableDeveloperMode bool) (logr.Logger, *zap.Logger, error) {
	if enableDeveloperMode {
		return createDevelopmentLogger(logLevel)
	} else {
		return createProductionLogger(logLevel)
	}
}

func createDevelopmentLogger(logLevel int) (logr.Logger, *zap.Logger, error) {
	zapConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.Level(-logLevel)), //nolint:gosec // Log levels will always be small.
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      zapcore.OmitKey,
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "M",
			StacktraceKey:  "S",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	zapLogger, err := zapConfig.Build()
	if err != nil {
		return logr.Discard(), nil, fmt.Errorf("creating the zap logger: %w", err)
	}
	return zapr.NewLogger(zapLogger), zapLogger, nil
}

func createProductionLogger(logLevel int) (logr.Logger, *zap.Logger, error) {
	zapConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.Level(-logLevel)), //nolint:gosec // Log levels will always be small.
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	zapLogger, err := zapConfig.Build()
	if err != nil {
		return logr.Discard(), nil, fmt.Errorf("creating the zap logger: %w", err)
	}
	return zapr.NewLogger(zapLogger), zapLogger, nil
}
