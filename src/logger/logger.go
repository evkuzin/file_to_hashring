package logger

import "go.uber.org/zap"

var L *zap.SugaredLogger

func InitLogger() {
	logger, _ := zap.NewProduction()
	L = logger.Sugar()
}
