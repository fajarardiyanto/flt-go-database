package lib

import (
	"github.com/fajarardiyanto/flt-go-database/lib/mongo"
	"github.com/fajarardiyanto/flt-go-database/lib/rabbitmq"
	"sync"

	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib/elasticsearch"
	"github.com/fajarardiyanto/flt-go-database/lib/redis"
	"github.com/fajarardiyanto/flt-go-database/lib/sql"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
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
	return sql.NewSQL(m.logging, config)
}

func (m *Modules) LoadRedisDatabase(config database.RedisProviderConfig) database.Redis {
	return redis.NewRedis(m.logging, config)
}

func (m *Modules) LoadMongoDatabase(config database.MongoProviderConfig) database.Mongo {
	return mongo.NewMongo(m.logging, config)
}

func (c *Modules) LoadRabbitMQ(tag string, config database.RabbitMQProviderConfig) database.RabbitMQ {
	return rabbitmq.NewRabbitMQ(tag, c.logging, config)
}
