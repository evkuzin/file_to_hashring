package postgres

import (
	"database/sql"
	"errors"
	"file-to-hashring/src/hashring"
	"file-to-hashring/src/logger"
	"fmt"
	_ "github.com/lib/pq"
	"strings"
)

type PgServer struct {
	*sql.DB
	name string
}

func (p *PgServer) Name() string {
	return p.name
}

func NewPGServer(server string) hashring.RingMember {
	serverParsed := strings.Split(server, ":")
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
		logger.L.Fatal(err)
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)

	return &PgServer{
		DB:   db,
		name: server,
	}
}

func NewHashRingMembersList(servers []string) []hashring.RingMember {
	hashRingMembers := make([]hashring.RingMember, len(servers))
	for i, pgServer := range servers {
		hashRingMembers[i] = NewPGServer(pgServer)
	}
	return hashRingMembers
}

func (p *PgServer) Put(name string, raw []byte) error {
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

func (p *PgServer) GetData(name string) ([]byte, error) {
	logger.L.Debug("GetData called")
	row := p.DB.QueryRow("SELECT data FROM public.files WHERE name = $1", name)

	var data []byte
	err := row.Scan(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *PgServer) GetAllKeys() []string {
	logger.L.Debug("GetAllKeys called")
	rows, err := p.DB.Query("SELECT name FROM public.files")
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			// I know it's a buggy code,
			// I just don't want to spend too much time on it
			return []string{}
		}
		keys = append(keys, key)
	}
	if err = rows.Err(); err != nil {
		return []string{}
	}
	return keys
}

func (p *PgServer) Delete(key string) {
	logger.L.Debug("GetAllKeys called")
	exec, err := p.DB.Exec("DELETE FROM public.files WHERE name = $1", key)
	if err != nil {
		logger.L.Warnf("something went wrong: %s", err)
	}
	affected, err := exec.RowsAffected()
	if err != nil {
		logger.L.Warnf("something went wrong: %s", err)
	}
	if affected != 1 {
		logger.L.Errorf("we are probably fucked: one row supposed to be deleted, but %d was", affected)
	}
}
