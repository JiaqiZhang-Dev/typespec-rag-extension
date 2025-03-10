package search

import (
	"context"
	"testing"
)

func TestQueryIndex(t *testing.T) {
	req := &QueryIndexRequest{
		Search: "how can i install typespec?",
		Count:  true,
		Top:    5,
		VectorQueries: []VectorQuery{
			{
				Text:       "how can i install typespec",
				K:          5,
				Fields:     "text_vector",
				Kind:       "text",
				Exhaustive: true,
			},
		},
	}
	resp, err := QueryIndex(context.Background(), req)
	if err != nil {
		t.Errorf("QueryIndex() got an error: %v", err)
	}
	print(resp)
}
