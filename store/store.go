package store

type Item struct {
	URL    string `json:"url" db:"url"`
	Count  int    `json:"count" db:"count"`
	Qrcode string `json:"qrcode" db:"qrcode"`
}

type Service interface {
	Load(string) (string, error)
	Save(string) (string, error)
	Info(string) (*Item, error)
	Close() error
}
