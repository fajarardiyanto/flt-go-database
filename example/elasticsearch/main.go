package main

import (
	"fmt"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib"
)

func main() {
	db := lib.NewLib()
	db.Init("Test Database ElasticSearch Modules")

	es := db.LoadElasticSearch("elasticsearch", interfaces.ElasticSearchProviderConfig{
		Enable:    true,
		Host:      "http://127.0.0.1",
		Port:      9200,
		IndexName: "user_idx",
	})

	if err := es.ElasticSearch(); err != nil {
		fmt.Println(err)
		return
	}

	// Search
	r, err := Search(es)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r)
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
