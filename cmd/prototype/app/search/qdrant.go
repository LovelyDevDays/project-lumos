package search

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"

	"github.com/qdrant/go-client/qdrant"
)

type SearchResult struct {
	Score float32
	Key   string
	Value string
}

func Search(
	ctx context.Context,
	address string,
	name string,
	vectors []float32,
) ([]SearchResult, error) {
	client, err := createClient(address)
	if err != nil {
		return nil, err
	}

	resp, err := client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: name,
		Query:          qdrant.NewQuery(vectors...),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(resp))
	for _, point := range resp {
		results = append(results, SearchResult{
			Score: point.Score,
			Key:   point.Payload["key"].GetStringValue(),
			Value: point.Payload["value"].GetStringValue(),
		})
	}
	return results, nil
}

func createClient(address string) (*qdrant.Client, error) {
	host, portString, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, err
	}
	return qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
}
