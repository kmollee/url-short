package store

type Item struct {
	URL   string `json:"url" db:"url"`
	Count int    `json:"count" db:"count"`
}

type Service interface {
	Load(string) (string, error)
	Save(string) (string, error)
	Info(string) (*Item, error)
	Close() error
}
