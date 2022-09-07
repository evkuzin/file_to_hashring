package server

import (
	"context"
	"database/sql"
	"errors"
	"file-to-hashring/src/config"
	"file-to-hashring/src/hashring"
	"file-to-hashring/src/hashring/imp"
	"file-to-hashring/src/logger"
	"file-to-hashring/src/storages/postgres"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	cfg   *config.Config
	ring  hashring.HashRing
	state *sql.DB
	http  *http.Server
}

func NewServer(cfg *config.Config) *Server {
	server := &Server{cfg: cfg}
	hashRingMembers := postgres.NewHashRingMembersList(cfg.Servers)
	server.ring = imp.NewHashRing(hashRingMembers)
	server.http = &http.Server{Addr: ":8080", Handler: nil}
	http.HandleFunc("/upload", server.UploadFile)
	http.HandleFunc("/download", server.DownloadFile)
	http.HandleFunc("/addserver", server.AddServer)
	return server
}

func (s *Server) Init() error {
	serverParsed := strings.Split(s.cfg.State, ":")
	// here supposed to be a nice validator
	if len(serverParsed) != 2 {
		logger.L.Fatalf("oops: something is wrong with this connection string")
	}
	connStr := fmt.Sprintf(
		"user=postgres dbname=postgres host=%s port=%s sslmode=disable",
		serverParsed[0],
		serverParsed[1],
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	s.state = db
	return nil
}

func (s *Server) GetFileMetadata(name string) (*File, error) {
	logger.L.Debug("GetFileMetadata called")
	row := s.state.QueryRow("SELECT name, size, nodes, content_type FROM public.metadata WHERE name = $1", name)

	file := &File{}
	err := row.Scan(&file.name, &file.size, &file.nodes, &file.contentType)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *Server) SaveFileMetadata(f *File) error {
	//TODO: check for duplicates
	res, err := s.state.Exec(
		"INSERT INTO public.metadata (name, size, nodes, content_type) VALUES ($1, $2, $3, $4)",
		f.name,
		f.size,
		f.nodes,
		f.contentType,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("metadata wasn't saved")
	}
	return nil
}

func (s *Server) Start() {
	logger.L.Info("starting web http...")
	err := s.http.ListenAndServe()
	if err != nil {
		logger.L.Fatal(err)
	}
}

func (s *Server) Stop() {
	logger.L.Info("stopping web http...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.http.Shutdown(ctx); err != nil {
		logger.L.Error(err)
	}

}
