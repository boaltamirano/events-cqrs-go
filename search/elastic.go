package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/go/events-cqrs-go/models"
)

type ElasticSearchRepository struct {
	client *elastic.Client
}

// constructor del struct
func NewElastic(url string) (*ElasticSearchRepository, error) {
	client, err := elastic.NewClient(elastic.Config{ // creacion de una conexion nueva a elasticsearch
		Addresses: []string{url},
	})

	if err != nil {
		return nil, err
	}

	return &ElasticSearchRepository{client: client}, nil
}

func (r *ElasticSearchRepository) Close() {
	//
}

func (r *ElasticSearchRepository) IndexFeed(ctx context.Context, feed models.Feed) error {
	body, _ := json.Marshal(feed)
	_, err := r.client.Index(
		"feeds",
		bytes.NewReader(body),
		r.client.Index.WithDocumentID(feed.ID),
		r.client.Index.WithContext(ctx),
		r.client.Index.WithRefresh("wait_for"),
	)
	return err
}

func (r *ElasticSearchRepository) SearchFeed(ctx context.Context, query string) (results []models.Feed, err error) {
	var buffer bytes.Buffer

	// fuzziness: para que acepte tipos que se escribieron mal,
	// cutoff_frequency: indica cuantas veces se debe repetir el termino a buscar
	searchQuery := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":            query,
				"fields":           []string{"title", "description"},
				"fuzziness":        3,
				"cutoff_frequency": 0.0001,
			},
		},
	}

	if err = json.NewEncoder(&buffer).Encode(searchQuery); err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("feeds"),
		r.client.Search.WithBody(&buffer),
		r.client.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			results = nil
		}
	}()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var eRes map[string]any
	if err := json.NewDecoder(res.Body).Decode(&eRes); err != nil {
		return nil, err
	}

	var feeds []models.Feed
	// hit: es un valor de elasticsearch
	for _, hit := range eRes["hits"].(map[string]any)["hits"].([]any) {
		feed := models.Feed{}
		source := hit.(map[string]any)["_source"]
		marshal, err := json.Marshal(source)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(marshal, &feed); err == nil {
			feeds = append(feeds, feed)
		}
	}
	return feeds, nil
}
