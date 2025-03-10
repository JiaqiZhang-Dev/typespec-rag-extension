package search

import (
	"context"
	"fmt"
)

func SearchTopKRelatedDocuments(query string, k int) ([]Index, error) {
	req := &QueryIndexRequest{
		Search: query,
		Count:  false,
		Top:    k,
		VectorQueries: []VectorQuery{
			{
				Text:       query,
				K:          k,
				Fields:     "text_vector",
				Kind:       "text",
				Exhaustive: true,
			},
		},
	}
	resp, err := QueryIndex(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("QueryIndex() got an error: %v", err)
	}
	return resp.Value, nil
}
