package postgre

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/kmollee/url-short/store"
	"github.com/lytics/base62"

	_ "github.com/lib/pq" // load postgresql driver
)

type storage struct{ db *sqlx.DB }

var schema = `
CREATE TABLE IF NOT EXISTS urlshorter (
	uid serial NOT NULL,
	url VARCHAR not NULL UNIQUE,
	count INTEGER DEFAULT 0
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
	id, err := base62.StdEncoding.DecodeString(hashCode)
	if err != nil {
		return nil, err
	}

	var item store.Item
	err = s.db.QueryRowx("SELECT * FROM urlshorter WHERE id=$1", id).Scan(&item)
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
		err := s.db.QueryRow("INSERT INTO urlshorter (url) VALUES ($1) RETURNING uid;", url).Scan(&id)
		if err != nil {
			return "", err
		}
	case nil:
	default:
		return "", err
	}

	return base62.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", id))), nil
}

func (s *storage) Load(hashCode string) (string, error) {
	id, err := base62.StdEncoding.DecodeString(hashCode)
	if err != nil {
		return "", err
	}
	var item store.Item
	err = s.db.QueryRowx("UPDATE urlshorter set count=count+1 WHERE uid=$1 RETURNING url, count;", id).StructScan(&item)
	if err != nil {
		return "", err
	}

	return item.URL, nil
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
