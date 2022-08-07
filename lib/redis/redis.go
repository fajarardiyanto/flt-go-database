package redis

import (
	"context"
	"fmt"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/go-redis/redis/v8"
	logger "gitlab.com/fajardiyanto/flt-go-logger/interfaces"
	"time"
)

type service struct {
	reconnectAt int
	db          *redis.Client
	log         logger.Logger
	lastTimeout time.Duration
	pool        *redis.PoolStats
	config      interfaces.RedisProviderConfig
}

func NewRedis(log logger.Logger, config interfaces.RedisProviderConfig) interfaces.Redis {
	log.Debugf("ElasticSearch Client %s:%d has been registered", config.Host, config.Port)
	return &service{
		config: config,
		log:    log,
	}
}

func (s *service) Init() (err error) {
	if !s.config.Enable {
		msg := "aborted, redis database not enable in config, double check configuration again"
		s.log.Error(msg)
		return nil
	}

	s.log.Debugf("Connecting to redis database server %s:%d", s.config.Host, s.config.Port)

	redisAddress := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.db = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: s.config.Password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = s.db.Info(ctx).Err(); err != nil {
		return s.OnInitError(err)
	}

	s.log.Infof("Success to connect redis %s:%d", s.config.Host, s.config.Port)

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
	s.log.Warn("Reconnecting in %s", s.lastTimeout)
	time.Sleep(s.lastTimeout)

	return s.Init()
}

func (s *service) GetPool() *redis.PoolStats {
	return s.pool
}

func (s *service) Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	if s.pool = s.db.PoolStats(); s.pool == nil {
		return fmt.Errorf("redis instance not available or not connected")
	}

	return s.db.Set(ctx, key, val, ttl).Err()
}

func (s *service) Get(ctx context.Context, key string) (string, error) {
	if s.pool = s.db.PoolStats(); s.pool == nil {
		return "", fmt.Errorf("redis instance not available or not connected")
	}

	val, err := s.db.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("redis keys not found")
	} else if err != nil {
		return "", err
	}

	return val, err
}
