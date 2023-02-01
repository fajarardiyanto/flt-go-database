package interfaces

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Database interface {
	Init(logger logger.Logger)
	LoadElasticSearch(string, ElasticSearchProviderConfig) ElasticSearch
	LoadSQLDatabase(config SQLConfig) SQL
	LoadRedisDatabase(config RedisProviderConfig) Redis
	LoadMongoDatabase(config MongoProviderConfig) Mongo
	LoadRabbitMQ(tag string, config RabbitMQProviderConfig) RabbitMQ
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

type Mongo interface {
	Init() error
	SetDatabase(db string) *mongo.Database
	LoadPostChannel(ctx context.Context, db, table string, filter bson.M, res chan<- []bson.M, opt ...*options.FindOptions)
}

type Kafka interface {
	Consumer(KafkaOptions, ConsumerCallback)
	Producer(ProducerIsReady)
	Push(ctx context.Context, id string, options KafkaOptions, body interface{}, cb ConsumerCallback) error
}

type RabbitMQ interface {
	Consumer(RabbitMQOptions, ConsumerCallback)
	Producer(RabbitMQOptions)
	Push(ctx context.Context,
		id, key string,
		body interface{},
		cb ConsumerCallback) error
}

type ConsumerCallbackIsDone struct {
	Done       context.CancelFunc
	EndRequest func()
}

type ConsumerCallback func(Messages, ConsumerCallbackIsDone)

type ProducerIsReady func()
