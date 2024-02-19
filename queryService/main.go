package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go/events-cqrs-go/database"
	"github.com/go/events-cqrs-go/events"
	"github.com/go/events-cqrs-go/repository"
	"github.com/go/events-cqrs-go/search"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	PostgresDB           string `envconfig:"POSTGRES_DB"`
	PostgresUser         string `envconfig:"POSTGRES_USER"`
	PostgresPassword     string `envconfig:"POSTGRES_PASSWORD"`
	NatsAddress          string `envconfig:"NATS_ADDRESS"`
	ElasticsearchAddress string `envconfig:"ELASTICSEARCH_ADDRESS"`
}

func newRouter() (router *mux.Router) {
	router = mux.NewRouter()
	router.HandleFunc("/feeds", listFeedHandler).Methods(http.MethodGet)
	router.HandleFunc("/search", searchHandler).Methods(http.MethodGet)
	return
}

func main() {
	var cfg Config

	err := envconfig.Process("", &cfg)

	if err != nil {
		log.Fatalf("%v", err)
	}

	// conexion a postgres
	addr := fmt.Sprintf("postgres://%s:%s@postgres/%s?sslmode=disable", cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)

	repo, err := database.NewPostgresRepository(addr)
	if err != nil {
		log.Fatal(err)
	}

	repository.SetRepository(repo)

	// conexion a elasticsearch
	elasticsearch, err := search.NewElastic(fmt.Sprintf("http://%s", cfg.ElasticsearchAddress))
	if err != nil {
		log.Fatal(err)
	}

	search.SetSearchRepository(elasticsearch)

	defer search.Close()

	// conexion a nat
	n, err := events.NewNats(fmt.Sprintf("nats://%s", cfg.NatsAddress))
	if err != nil {
		log.Fatal(err)
	}

	// Hacer que el query se suscriba a un evento
	err = n.OnCreateFeed(onCreateFeed) // TODO: crear la funcion onCreateFeed mas adelante
	if err != nil {
		log.Fatal(err)
	}
	events.SetEventStore(n)
	defer events.Close()

	// crear un router
	router := newRouter()
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
