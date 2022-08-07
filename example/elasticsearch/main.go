package main

import (
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib"
	log "github.com/fajarardiyanto/flt-go-logger/lib"
)

func main() {
	logger := log.NewLib()
	logger.Init("Modules ElasticSearch")

	db := lib.NewLib()
	db.Init(logger)

	es := db.LoadElasticSearch("elasticsearch", interfaces.ElasticSearchProviderConfig{
		Enable:        true,
		Host:          "http://127.0.0.1",
		Port:          9200,
		IndexName:     "user_idx",
		AutoReconnect: true,
		StartInterval: 2,
	})

	if err := es.ElasticSearch(); err != nil {
		logger.Error(err)
		return
	}

	// Search
	r, err := Search(es)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(r)
}

func Search(es interfaces.ElasticSearch) (*interfaces.SearchResultsElasticSearch, error) {
	// SEARCH
	opts := interfaces.ElasticSearchOptions{
		Query: "backend",
		Sort:  "asc",
		Size:  1,
	}
	return es.Search(opts)
}
