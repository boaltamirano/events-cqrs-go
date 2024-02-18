package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go/events-cqrs-go/database"
	"github.com/go/events-cqrs-go/events"
	"github.com/go/events-cqrs-go/repository"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	PostgresDB       string `envconfig:"POSTGRES_DB"`
	PostgresUser     string `envconfig:"POSTGRES_USER"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD"`
	NatsAddress      string `envconfig:"NATS_ADDRESS"`
}

func main() {
	var cfg Config

	err := envconfig.Process("", &cfg)

	if err != nil {
		log.Fatalf("%v", err)
	}

	addr := fmt.Sprintf("postgres://%s:%s@postgres/%s?sslmode=disable", cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)

	repo, err := database.NewPostgresRepository(addr)
	if err != nil {
		log.Fatal(err)
	}

	repository.SetRepository(repo)

	n, err := events.NewNats(fmt.Sprintf("nats://%s", cfg.NatsAddress))
	if err != nil {
		log.Fatal(err)
	}

	events.SetEventStore(n)

	defer events.Close()

	router := newRouter()
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
