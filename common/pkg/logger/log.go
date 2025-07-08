package logger

import "go.uber.org/zap"

var _defaultLogger *zap.Logger

func GetLogger() *zap.Logger {
	return _defaultLogger
}
