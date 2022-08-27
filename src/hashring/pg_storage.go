package hashring

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
)

type pgServer struct {
	*sql.DB
}

func NewPGServer(db *sql.DB) RingMember {
	return &pgServer{db}
}

func (p pgServer) Put(name string, raw []byte) error {
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

func (p pgServer) GetData(name string) ([]byte, error) {
	rows, err := p.DB.Query("SELECT data FROM public.files WHERE file = $1", name)
	if err != nil {
		return nil, err
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	var data []byte
	err = rows.Scan(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p pgServer) GetSize(name string) (int64, error) {
	rows, err := p.DB.Query("SELECT size FROM public.files WHERE file = $1", name)
	if err != nil {
		return 0, err
	}
	err = rows.Err()
	if err != nil {
		return 0, err
	}
	var size int64
	err = rows.Scan(&size)
	if err != nil {
		return 0, err
	}

	return size, nil
}
