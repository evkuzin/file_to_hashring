package postgres

import (
	"database/sql"
	"errors"
	"file-to-hashring/src/hashring"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"strings"
)

type PgServer struct {
	*sql.DB
}

func newPGServer(db *sql.DB) hashring.RingMember {
	return &PgServer{db}
}

func NewHashRing(servers []string) []hashring.RingMember {
	hashRingMembers := make([]hashring.RingMember, len(servers))
	for i, pgServer := range servers {
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
		hashRingMembers[i] = newPGServer(db)
	}
	return hashRingMembers
}

func (p PgServer) Put(name string, raw []byte) error {
	res, err := p.DB.Exec(
		"INSERT INTO public.files (name, size, data) VALUES ($1, $2, $3)",
		name,
		len(raw),
		raw,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("file part wasn't saved")
	}
	return nil
}

func (p PgServer) GetData(name string) ([]byte, error) {
	log.Println("GetData called")
	row := p.DB.QueryRow("SELECT data FROM public.files WHERE name = $1", name)

	var data []byte
	err := row.Scan(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p PgServer) GetSize(name string) (int64, error) {
	log.Println("GetSize called")
	row := p.DB.QueryRow("SELECT size FROM public.files WHERE name = $1", name)

	var size int64
	err := row.Scan(&size)
	if err != nil {
		return 0, err
	}

	return size, nil
}
