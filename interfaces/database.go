package interfaces

import (
	"context"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	logger "gitlab.com/fajardiyanto/flt-go-logger/interfaces"
	"gorm.io/gorm"
	"time"
)

type Database interface {
	Init(logger logger.Logger)
	LoadElasticSearch(string, ElasticSearchProviderConfig) ElasticSearch
	LoadSQLDatabase(config SQLConfig) SQL
	LoadRedisDatabase(config RedisProviderConfig) Redis
}

type ElasticSearch interface {
	Elastic() *elasticsearch.Client
	ElasticSearch() (err error)
	Search(config ElasticSearchOptions) (*SearchResultsElasticSearch, error)
	CreateIndex(name string, mapping string) error
	Create(index string, id string, values interface{}) error
	Delete(id string) error
}

type Redis interface {
	Init() error
	GetPool() *redis.PoolStats
	Set(context.Context, string, interface{}, time.Duration) error
	Get(context.Context, string) (string, error)
}

type SQL interface {
	Orm() *gorm.DB
	MySQL() error
	LoadSQL() error
}
