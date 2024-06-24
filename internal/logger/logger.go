package logger

import (
	"chat_app/config"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() error {
	if config.AppConfig == nil {
		return fmt.Errorf("configuration not loaded")
	}

	zapConfig := zap.NewProductionConfig()

	zapConfig.OutputPaths = config.AppConfig.Logger.OutputPaths
	zapConfig.ErrorOutputPaths = config.AppConfig.Logger.ErrorOutputPaths

	level, err := zapcore.ParseLevel(config.AppConfig.Logger.Level)
	if err != nil {
		return err
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	Log, err = zapConfig.Build()
	if err != nil {
		return err
	}

	return nil
}
