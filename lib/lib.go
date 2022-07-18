package lib

import (
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib/elasticsearch"
	"sync"
)

type Modules struct {
	spacename string
	sync.RWMutex
}

func NewLib() database.Database {
	return &Modules{}
}

func (m *Modules) Init(spacename string) {
	m.spacename = spacename
}

func (m *Modules) LoadElasticSearch(tag string, config database.ElasticSearchProviderConfig) database.ElasticSearch {
	return elasticsearch.NewElasticSearch(tag, config)
}
