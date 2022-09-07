package logger

import (
	"file-to-hashring/src/config"
	"fmt"
	"go.uber.org/zap"
	"os"
)

var L *zap.SugaredLogger

func InitLogger(cfg *config.Config) {
	logger, err := cfg.Logger.Build()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	L = logger.Sugar()
}
