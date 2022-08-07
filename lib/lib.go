package lib

import (
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib/elasticsearch"
	"github.com/fajarardiyanto/flt-go-database/lib/redis"
	"github.com/fajarardiyanto/flt-go-database/lib/sql"
	logger "gitlab.com/fajardiyanto/flt-go-logger/interfaces"
	"sync"
)

type Modules struct {
	logging logger.Logger
	sync.RWMutex
}

func NewLib() database.Database {
	return &Modules{}
}

func (m *Modules) Init(lo logger.Logger) {
	m.logging = lo
}

func (m *Modules) LoadElasticSearch(tag string, config database.ElasticSearchProviderConfig) database.ElasticSearch {
	return elasticsearch.NewElasticSearch(tag, m.logging, config)
}

func (m *Modules) LoadSQLDatabase(config database.SQLConfig) database.SQL {
	return sql.NewSQL(config)
}

func (m *Modules) LoadRedisDatabase(config database.RedisProviderConfig) database.Redis {
	return redis.NewRedis(m.logging, config)
}
