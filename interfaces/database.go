package interfaces

import (
	"github.com/elastic/go-elasticsearch/v7"
	logger "gitlab.com/fajardiyanto/flt-go-logger/interfaces"
	"gorm.io/gorm"
)

type Database interface {
	Init(logger logger.Logger)
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

type SQL interface {
	Orm() *gorm.DB
	MySQL() error
	LoadSQL() error
}
