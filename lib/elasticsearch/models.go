package elasticsearch

type EnvelopeResponse struct {
	Took int
	Hits Hits
}

type Hits struct {
	Total struct {
		Value int
	}
	Hits []struct {
		ID     string      `json:"_id"`
		Source interface{} `json:"_source"`
	}
}
