package logger

import "go.uber.org/zap"

var log *zap.Logger

func Init() {
	log, _ = zap.NewProduction()
}

func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}
