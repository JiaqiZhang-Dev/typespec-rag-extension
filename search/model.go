package search

import "strings"

type QueryIndexRequest struct {
	Count         bool          `json:"count,omitempty";`
	Search        string        `json:"search,omitempty"`
	Select        string        `json:"select,omitempty"`
	Top           int           `json:"top,omitempty"`
	VectorQueries []VectorQuery `json:"vectorQueries,omitempty"`
	OrderBy       string        `json:"orderby,omitempty"`
	Filter        string        `json:"filter,omitempty"`
}

type VectorQuery struct {
	Text       string `json:"text"`
	K          int    `json:"k"`
	Fields     string `json:"fields"`
	Kind       string `json:"kind"`
	Exhaustive bool   `json:"exhaustive"`
}

type QueryIndexResponse struct {
	Context string  `json:"@odata.context"`
	Count   int     `json:"@odata.count"`
	Value   []Index `json:"value"`
}

type Index struct {
	Score           float64 `json:"@search.score"`
	ChunkID         string  `json:"chunk_id"`
	ParentID        string  `json:"parent_id"`
	Chunk           string  `json:"chunk"`
	Title           string  `json:"title"`
	Header1         string  `json:"header_1"`
	Header2         string  `json:"header_2"`
	Header3         string  `json:"header_3"`
	OrdinalPosition int     `json:"ordinal_position"`
}

func GetIndexLink(chunk Index) string {
	path := strings.Join(strings.Split(chunk.Title, "_"), "/")
	path = strings.TrimSuffix(path, ".md")
	path = strings.TrimSuffix(path, ".mdx")
	path = strings.TrimPrefix(path, "docs/")
	return "https://typespec.io/docs/" + path
}
