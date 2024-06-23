package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func InitLogger() error {
	// initialize the logger
}

func Sync() error {
	// sync the logger before application exit
}
