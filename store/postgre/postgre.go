package postgre

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/kmollee/url-short/base62"
	"github.com/kmollee/url-short/store"
	"github.com/skip2/go-qrcode"

	_ "github.com/lib/pq" // load postgresql driver
)

const hashShift = 10000

type storage struct{ db *sqlx.DB }

var schema = `
CREATE TABLE IF NOT EXISTS urlshorter (
	uid serial NOT NULL,
	url VARCHAR not NULL UNIQUE,
	count INTEGER DEFAULT 0,
	qrcode TEXT NOT NULL
);
`

func (s *storage) createTalbe() error {
	_, err := s.db.Exec(schema)
	return err
}

func (s *storage) Close() error {
	return s.db.Close()
}

func (s *storage) Info(hashCode string) (*store.Item, error) {
	id := base62.Decode(hashCode) - hashShift

	var item store.Item
	err := s.db.QueryRowx("SELECT url, count, qrcode FROM urlshorter WHERE uid=$1", id).StructScan(&item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *storage) Save(url string) (string, error) {

	var id int64

	err := s.db.QueryRow("SELECT uid FROM urlshorter WHERE url=$1", url).Scan(&id)

	switch err {
	case sql.ErrNoRows:
		log.Println("create new record")
		var png []byte
		png, err := qrcode.Encode(url, qrcode.Medium, 256)

		encodeImage := base64.StdEncoding.EncodeToString(png)

		err = s.db.QueryRow("INSERT INTO urlshorter (url, qrcode) VALUES ($1, $2) RETURNING uid;", url, encodeImage).Scan(&id)
		if err != nil {
			return "", err
		}
	case nil:
	default:
		return "", err
	}

	return base62.Encode(int(id) + hashShift), nil
}

func (s *storage) Load(hashCode string) (string, error) {
	// id, err := base62.StdEncoding.DecodeString(hashCode)
	id := base62.Decode(hashCode) - hashShift
	var url string
	err := s.db.QueryRowx("UPDATE urlshorter set count=count+1 WHERE uid=$1 RETURNING url", id).Scan(&url)

	if err != nil {
		return "", err
	}

	return url, nil
}

// New returns a postgres backed storage service.
func New(host, port, user, password, dbName string) (store.Service, error) {
	// Connect postgres
	connect := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)
	log.Printf("CONNECT: %s", connect)

	db, err := sqlx.Connect("postgres", connect)
	if err != nil {
		return nil, err
	}

	s := &storage{db}

	err = s.createTalbe()
	if err != nil {
		return nil, err
	}
	return s, nil
}
