package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go/events-cqrs-go/events"
	"github.com/go/events-cqrs-go/models"
	"github.com/go/events-cqrs-go/repository"
	"github.com/segmentio/ksuid"
)

type CreatedFeedRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func createdFeedHandler(w http.ResponseWriter, r *http.Request) {
	var req CreatedFeedRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdAt := time.Now().UTC()

	id, err := ksuid.NewRandom()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed := models.Feed{
		ID:          id.String(),
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   createdAt,
	}

	// guardar el feed y a la ves validamos si el error es diferente de nil
	if err := repository.InsertFeed(r.Context(), &feed); err != nil {
		http.Error(w, err.Error(), http.StatusInsufficientStorage)
	}

	// Vamos a trasmitir el feed creado a traves de NAT y a la ves validamos si el error es diferente de nil
	if err := events.PublishCreatedFeed(r.Context(), &feed); err != nil {
		log.Printf("failed to publish created feed event: %v", err)
	}

	// devolvemos la respuesta al cliente
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feed)

}
