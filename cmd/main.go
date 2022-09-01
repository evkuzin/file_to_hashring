package main

import (
	"file-to-hashring/src/config"
	"file-to-hashring/src/logger"
	"file-to-hashring/src/server"
	"flag"
)

func init() {
	logger.InitLogger()
}

func main() {

	cfgPath := flag.String("c", "../config.yaml", "path to a config file")
	flag.Parse()
	cfg, err := config.NewConfig(*cfgPath)
	if err != nil {
		logger.L.Fatal(err)
	}
	server.Start(cfg)
}
