package mongo

import (
	"context"
	"fmt"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type service struct {
	reconnectAt int
	db          *mongo.Client
	log         logger.Logger
	lastTimeout time.Duration
	pool        *redis.PoolStats
	config      interfaces.MongoProviderConfig
	sync.Mutex
}

func NewMongo(log logger.Logger, config interfaces.MongoProviderConfig) interfaces.Mongo {
	log.Debug("Mongo Client %s:%d has been registered", config.Host, config.Port)
	return &service{
		config: config,
		log:    log,
	}
}

func (s *service) Init() (err error) {
	if !s.config.Enable {
		msg := "aborted, mongo database not enable in config, double check configuration again"
		s.log.Error(msg)
		return nil
	}

	s.log.Debug("Connecting to mongo database server %s:%d", s.config.Host, s.config.Port)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.TimeoutConnection)*time.Millisecond)
	defer cancel()

	var addr string
	addr = fmt.Sprintf("mongodb://%s:%d", s.config.Host, s.config.Port)
	if s.config.Username != "" {
		addr = fmt.Sprintf("mongodb://%s:%s@%s:%d", s.config.Username, s.config.Password, s.config.Host, s.config.Port)
	}

	s.db, err = mongo.Connect(ctx, options.Client().ApplyURI(addr))
	if err != nil {
		return s.OnInitError(err)
	}

	s.log.Info("Success to connect mongo %s", addr)

	return nil
}

func (s *service) OnInitError(e error) (err error) {
	if s.reconnectAt >= s.config.MaxError {
		return e
	}

	s.reconnectAt++
	if !s.config.AutoReconnect {
		return e
	}

	var ttm = 2
	if s.lastTimeout.Seconds() == 0 {
		ttm = 1
		s.lastTimeout = time.Duration(s.config.StartInterval) * time.Second
	}

	s.lastTimeout = time.Duration(int(s.lastTimeout.Seconds())*ttm) * time.Second
	s.log.Error(e)
	s.log.Warning("Reconnecting in %s", s.lastTimeout)
	time.Sleep(s.lastTimeout)

	return s.Init()
}

func (s *service) SetDatabase(db string) *mongo.Database {
	return s.db.Database(db)
}

func (s *service) LoadPostChannel(ctx context.Context, db, table string, filter bson.M, res chan<- []bson.M, opt ...*options.FindOptions) {
	now := time.Now()

	cur, err := s.db.Database(db).Collection(table).Find(ctx, filter, opt...)
	if err != nil {
		s.log.Error(err)
		return
	}
	defer cur.Close(ctx)

	var wg sync.WaitGroup
	results := make([]bson.M, 0)

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for cur.Next(ctx) {
			var result bson.M
			if err = cur.Decode(&result); err != nil {
				s.Lock()
				s.log.Error(err)
				s.Unlock()

				continue
			}

			s.Lock()
			results = append(results, result)
			s.Unlock()
		}
	}(&wg)

	wg.Wait()

	te := time.Now().Sub(now).Seconds()
	s.log.Info("End Load Post %v/s", te)

	res <- results
}
