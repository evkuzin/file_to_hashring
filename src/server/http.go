package server

import (
	"errors"
	"file-to-hashring/src/config"
	"file-to-hashring/src/hashring"
	"file-to-hashring/src/logger"
	"file-to-hashring/src/storages/postgres"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var ring *hashring.HashRing
var itemsPerServer int

func downloadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		logger.L.Fatalf("err: %s", err)
		return
	}
	fileName := r.Form.Get("filename")
	logger.L.Infof("file %s requested", fileName)
	var fileSize int64
	for i := 0; i < itemsPerServer; i++ {
		fileNamePart := fmt.Sprintf("%s_%d", fileName, i)
		filePartSize, err := ring.GetServer(fileNamePart).GetSize(fileNamePart)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			logger.L.Fatalf("err: %s", err)
			return
		}
		fileSize += filePartSize
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))

	for i := 0; i < itemsPerServer; i++ {
		fileNamePart := fmt.Sprintf("%s_%d", fileName, i)
		filePart, err := ring.GetServer(fileNamePart).GetData(fileNamePart)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			logger.L.Fatalf("err: %s", err)
			return
		}
		if i == 0 {
			w.Header().Set("Content-Type", http.DetectContentType(filePart[:512]))
			w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		}
		written, err := w.Write(filePart)
		if err != nil {
			logger.L.Fatalf("err: %s", err)
			return
		}
		if written != len(filePart) {
			logger.L.Fatalf("written %d, supposed to write %d", written, len(filePart))
		}
	}
	return
}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	// Это полная лажа, по хорощему надо вытаскивать файл
	// и раскидывать его по частям по мере поступления.
	// Может горутинками даже, но мне некогда над этим сейчас работать.
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		logger.L.Warnf("error while ParseMultipartForm: %s", err)
		return
	}
	//err := r.ParseForm()
	//if err != nil {
	//	logger.L.Warnf("error parsing form: %s", err)
	//	return
	//}
	//fileSize := r.ContentLength
	//logger.L.Infof("Headers: %v", r.Header)
	//if err != nil {
	//	logger.L.Warnf("error getting ontent length size: %s", err)
	//	return
	//}
	//
	//reader, err := r.MultipartReader()
	//if err != nil {
	//	logger.L.Warnf("error calling MultipartReader: %s", err)
	//	return
	//}
	//part, err := reader.NextPart()
	//if err != nil {
	//	logger.L.Warnf("error: %s", err)
	//	return
	//}

	//for {
	//
	//	for {
	//		partData := make([]byte, 1024)
	//		bytesRead, err := part.Read(partData)
	//		if err != nil {
	//			logger.L.Warnf("error: %s", err)
	//			return
	//		}
	//		logger.L.Infof("part %d, filename: %s, header: %v, bytes read: %d", iterator, part.FileName(), part.Header, bytesRead)
	//		logger.L.Infof("raw: %v", partData)
	//	}
	//
	//}
	file, handler, err := r.FormFile("upload")
	if err != nil {
		logger.L.Warnf("Error Retrieving the File: %s", err)
		return
	}
	defer file.Close()
	logger.L.Infof("Uploaded File: %+v", handler.Filename)
	logger.L.Infof("File Size: %+v", handler.Size)
	logger.L.Infof("MIME Header: %+v", handler.Header)
	if handler.Size < int64(itemsPerServer) {
		logger.L.Warnf("number of chunks is too much for this file. can't do that")
		w.Write([]byte(`{"err": "number of chunks is too much for this file. can't do that"}`))
		return
	}
	filePartSize := handler.Size / int64(itemsPerServer)
	for i := 0; i < itemsPerServer; i++ {
		filePart := make([]byte, filePartSize)
		bytesRead, err := file.Read(filePart)
		if err != nil {
			if err != errors.New("EOF") {
				logger.L.Fatalf("cant read file part. err: %s", err)
				return
			}
		}
		if filePartSize != int64(bytesRead) {
			logger.L.Errorf("supposed to read %d, but read only %d, reading more", filePartSize, bytesRead)

		}

		if i == itemsPerServer-1 && err == nil {
			endOfTheFile, err := ioutil.ReadAll(file)
			if err != nil {
				logger.L.Fatalf("cant read file part with ioutil.ReadAll. err: %s", err)
			}
			for _, b := range endOfTheFile {
				filePart = append(filePart, b)
			}
		}
		fileName := fmt.Sprintf("%s_%d", handler.Filename, i)
		err = ring.GetServer(fileName).Put(fileName, filePart)
		if err != nil {
			logger.L.Fatalf("cant save part of the file in the database. err: %s", err)
			return
		}
		logger.L.Infof("file part %s uploaded", fileName)
	}
}

func Start(cfg *config.Config) {
	hashRingMembers := postgres.NewHashRing(cfg.Servers)
	itemsPerServer = cfg.ChunksPerServer * len(cfg.Servers)
	ring = hashring.New(hashRingMembers)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/download", downloadFile)
	log.Println("starting web server...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.L.Fatal(err)
		return
	}
}
