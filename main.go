package main

import (
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/kmollee/url-short/controller"
	"github.com/kmollee/url-short/store/postgre"
)

type config struct {
	Port string `env:"PORT" envDefault:"8000"`
	DB   struct {
		Host     string `env:"DB_HOST"`
		Port     string `env:"DB_PORT"`
		User     string `env:"DB_USER"`
		Password string `env:"DB_PASSWORD"`
		Name     string `env:"DB_NAME"`
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v\n", err)
	}

	svc, err := postgre.New(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
	if err != nil {
		time.Sleep(3 * time.Second)
		svc, err = postgre.New(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
		if err != nil {
			log.Fatal(err)
		}
	}

	r := controller.New(svc)

	log.Printf("start listen on %v", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}

}
