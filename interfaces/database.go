package interfaces

import "github.com/elastic/go-elasticsearch/v7"

type Database interface {
	Init(spacename string)
	LoadElasticSearch(string, ElasticSearchProviderConfig) ElasticSearch
}

type ElasticSearch interface {
	Elastic() *elasticsearch.Client
	ElasticSearch() (err error)
	Search(config ElasticSearchOptions) (*SearchResultsElasticSearch, error)
	CreateIndex(name string, mapping string) error
	Create(index string, id string, values interface{}) error
	Delete(id string) error
}
