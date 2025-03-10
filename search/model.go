package search

type QueryIndexRequest struct {
	Count         bool          `json:"count"`
	Search        string        `json:"search"`
	Select        string        `json:"select"`
	Top           int           `json:"top"`
	VectorQueries []VectorQuery `json:"vectorQueries"`
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
	Score    float64 `json:"@search.score"`
	ChunkID  string  `json:"chunk_id"`
	ParentID string  `json:"parent_id"`
	Chunk    string  `json:"chunk"`
	Title    string  `json:"title"`
	Header1  string  `json:"header_1"`
	Header2  string  `json:"header_2"`
	Header3  string  `json:"header_3"`
}
