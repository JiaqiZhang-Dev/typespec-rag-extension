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
		Select: "title",
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

func GetCompleteContext(chunk Index) ([]Index, error) {
	req := &QueryIndexRequest{
		Count:   false,
		OrderBy: "ordinal_position",
		Select:  "chunk_id, chunk, title, header_1, header_2, header_3, ordinal_position",
		Filter:  fmt.Sprintf("title eq '%s'", chunk.Title),
	}
	resp, err := QueryIndex(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("QueryIndex() got an error: %v", err)
	}
	return resp.Value, nil
}
