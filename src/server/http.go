package server

import (
	"file-to-hashring/src/config"
	"file-to-hashring/src/hashring/imp"
	"file-to-hashring/src/logger"
	"file-to-hashring/src/storages/postgres"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func Start(cfg *config.Config) {
	hashRingMembers := postgres.NewHashRing(cfg.Servers)
	ring := imp.NewHashRing(hashRingMembers)
	http.HandleFunc("/upload", ring.UploadFile)
	http.HandleFunc("/download", ring.DownloadFile)
	http.HandleFunc("/addserver", ring.AddServer)
	log.Println("starting web server...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.L.Fatal(err)
		return
	}
}
