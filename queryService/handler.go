package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go/events-cqrs-go/events"
	"github.com/go/events-cqrs-go/models"
	"github.com/go/events-cqrs-go/repository"
	"github.com/go/events-cqrs-go/search"
)

// Esta funcion avisa a elasticsearch que debe indexar un nuevo feed recien creado
func onCreateFeed(message events.CreatedFeedMessage) {
	feed := models.Feed{
		ID:          message.ID,
		Title:       message.Title,
		Description: message.Description,
		CreatedAt:   message.CreatedAt,
	}

	if err := search.IndexFeed(context.Background(), feed); err != nil {
		log.Printf("field to index feed: %v", err)
	}
}

// Funcion para listar los feed recien creados
func listFeedHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error
	feeds, err := repository.ListFeeds(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(feeds)
}

// Funcion de busqueda
func searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var err error

	query := r.URL.Query().Get("q")
	if len(query) == 0 {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	feeds, err := search.SearchFeed(ctx, query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(feeds)
}
