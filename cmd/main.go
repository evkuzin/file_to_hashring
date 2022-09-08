package main

import (
	"file-to-hashring/src/config"
	"file-to-hashring/src/logger"
	"file-to-hashring/src/server"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfgPath := flag.String("c", "../config.yaml", "path to a config file")
	flag.Parse()
	cfg, err := config.NewConfig(*cfgPath)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	l, err := cfg.Logger.Build()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	logger.InitLogger(l.Sugar())
	s := server.NewServer(cfg)
	err = s.Init()
	if err != nil {
		logger.L.Fatal(err)
	}
	s.Start()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.L.Info("graceful shutdown")
	s.Stop()
	os.Exit(0)
}
