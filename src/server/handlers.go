package server

import (
	"file-to-hashring/src/logger"
	"file-to-hashring/src/storages/postgres"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type File struct {
	name        string
	size        int64
	nodes       int
	contentType string
}

func (s *Server) UploadFile(w http.ResponseWriter, r *http.Request) {

	// Это полная лажа, по хорощему надо вытаскивать файл
	// и раскидывать его по частям по мере поступления.
	// Может горутинками даже, но мне некогда над этим сейчас работать.
	//TODO: make file streaming
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
		w.WriteHeader(http.StatusGatewayTimeout)
		logger.L.Warnf("rrror during retrieving the file body: %s", err)
		return
	}
	defer file.Close()
	logger.L.Infof("Uploaded File: %+v", handler.Filename)
	logger.L.Infof("File Size: %+v", handler.Size)
	logger.L.Infof("MIME Header: %+v", handler.Header)
	err = s.SaveFileMetadata(&File{
		name:        handler.Filename,
		size:        handler.Size,
		nodes:       0,
		contentType: handler.Header.Get("Content-Type"),
	})
	if err != nil {
		w.WriteHeader(http.StatusGatewayTimeout)
		logger.L.Warnf("error during saving file metadata: %s", err)
		return
	}
	if handler.Size < int64(s.ring.Chunks()) {
		//TODO: make it work
		logger.L.Warnf("number of chunks is too much for this file. can't do that")
		w.Write([]byte(`{"err": "number of chunks is too much for this file. can't do that"}`))
		return
	}
	filePartSize := handler.Size / int64(s.ring.Chunks())
	for i := 0; i < s.ring.Chunks(); i++ {
		filePart := make([]byte, filePartSize)
		bytesRead, err := file.Read(filePart)
		if err != nil {
			if err != io.EOF {
				w.WriteHeader(http.StatusGatewayTimeout)
				logger.L.Errorf("cant read file part. err: %s", err)
				return
			}
		}
		if filePartSize != int64(bytesRead) {
			logger.L.Errorf("supposed to read %d, but read only %d, reading more", filePartSize, bytesRead)

		}

		if i == s.ring.Chunks()-1 && err == nil {
			endOfTheFile, err := ioutil.ReadAll(file)
			if err != nil {
				w.WriteHeader(http.StatusGatewayTimeout)
				logger.L.Errorf("cant read last bits: %s", err)
				return
			}
			for _, b := range endOfTheFile {
				filePart = append(filePart, b)
			}
		}
		fileName := fmt.Sprintf("%s_%d", handler.Filename, i)
		err = s.ring.GetServer(fileName).Put(fileName, filePart)
		if err != nil {
			//TODO: retries/skip/evict bad http http?
			logger.L.Errorf("cant save part of the file on the storage. err: %s", err)
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		logger.L.Debugf("file part %s uploaded", fileName)
	}
}

func (s *Server) DownloadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.L.Errorf("err: %s", err)
		return
	}
	fileName := r.Form.Get("filename")
	logger.L.Infof("file %s requested", fileName)
	fileMeta, err := s.GetFileMetadata(fileName)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.L.Errorf("err: %s", err)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMeta.name))
	w.Header().Set("Content-Type", fileMeta.contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileMeta.size, 10))
	for i := 0; i < s.ring.Chunks(); i++ {
		fileNamePart := fmt.Sprintf("%s_%d", fileName, i)
		filePart, err := s.ring.GetServer(fileNamePart).GetData(fileNamePart)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			logger.L.Errorf("err: %s", err)
			return
		}
		written, err := w.Write(filePart)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			logger.L.Errorf("err: %s", err)
			return
		}
		if written != len(filePart) {
			w.WriteHeader(http.StatusServiceUnavailable)
			logger.L.Errorf("written %d, supposed to write %d", written, len(filePart))
		}
	}
	return
}

func (s *Server) AddServer(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		//TODO: retries/skip/evict bad http http?
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.L.Errorf("err: %s", err)
		return
	}
	server := r.Form.Get("server")
	logger.L.Infof("http %s will be added", server)
	err = s.ring.AddServer(postgres.NewPGServer(server))
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		logger.L.Error(err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
