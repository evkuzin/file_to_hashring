package http_server

import (
	"database/sql"
	"errors"
	"file-to-hashring/src/hashring"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var ring *hashring.HashRing
var pgServers = []string{
	"127.0.0.1:5432",
	"127.0.0.1:5433",
	"127.0.0.1:5434",
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Download Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	filePartSize := handler.Size / int64(len(pgServers))
	for i := 0; i < len(pgServers); i++ {
		filePart := make([]byte, filePartSize)
		_, err := file.Read(filePart)
		if err != nil {
			if err != errors.New("EOF") {
				log.Fatalf("cant read file part. err: %s", err)
			}
		}
		if i == len(pgServers)-1 && err == nil {
			endOfTheFile, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cant read file part with ioutil.ReadAll. err: %s", err)
			}
			for _, b := range endOfTheFile {
				filePart = append(filePart, b)
			}
		}
		fileName := fmt.Sprintf("%s_%d", handler.Filename, i)
		err = ring.GetServer(fileName).Put(fileName, filePart)
		if err != nil {
			log.Fatalf("cant save part of the file in the database. err: %s", err)
		}
		log.Printf("file part %s uploaded", fileName)
	}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	filePartSize := handler.Size / int64(len(pgServers))
	for i := 0; i < len(pgServers); i++ {
		filePart := make([]byte, filePartSize)
		_, err := file.Read(filePart)
		if err != nil {
			if err != errors.New("EOF") {
				log.Fatalf("cant read file part. err: %s", err)
			}
		}
		if i == len(pgServers)-1 && err == nil {
			endOfTheFile, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cant read file part with ioutil.ReadAll. err: %s", err)
			}
			for _, b := range endOfTheFile {
				filePart = append(filePart, b)
			}
		}
		fileName := fmt.Sprintf("%s_%d", handler.Filename, i)
		err = ring.GetServer(fileName).Put(fileName, filePart)
		if err != nil {
			log.Fatalf("cant save part of the file in the database. err: %s", err)
		}
		log.Printf("file part %s uploaded", fileName)
	}
}

func Start() {
	hashRingMembers := make([]hashring.RingMember, len(pgServers))
	for i, pgServer := range pgServers {
		serverParsed := strings.Split(pgServer, ":")
		connStr := fmt.Sprintf(
			"user=postgres dbname=postgres host=%s port=%s sslmode=disable",
			serverParsed[0],
			serverParsed[1],
		)
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatal(err)
		}
		hashRingMembers[i] = hashring.NewPGServer(db)
	}

	ring = hashring.New(hashRingMembers)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/download", downloadFile)
	http.ListenAndServe(":8080", nil)
}
