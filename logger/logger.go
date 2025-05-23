package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger создает новый экземпляр логгера.
func NewLogger(level string, isDevelopment bool) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	var zapLevel zapcore.Level
	var config zap.Config
	switch level {
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "debug":
		zapLevel = zapcore.DebugLevel
	default:
		return nil, fmt.Errorf("unsupported log level: %s", level)
	}

	if isDevelopment {
		// Development Logger: человекочитаемый формат
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Цветной вывод уровней логов
	} else {
		// Production Logger: структурированный JSON
		config = zap.NewProductionConfig()
	}
	config.Level.SetLevel(zapLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	logger, err = config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
