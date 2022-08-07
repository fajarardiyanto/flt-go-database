package main

import (
	"context"
	"time"

	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib"
	log "github.com/fajarardiyanto/flt-go-logger/lib"
)

func main() {
	logger := log.NewLib()
	logger.Init("Modules Redis Database")

	db := lib.NewLib()
	db.Init(logger)

	rdb := db.LoadRedisDatabase(interfaces.RedisProviderConfig{
		Enable:        true,
		Host:          "127.0.0.1",
		Port:          6379,
		Password:      "",
		AutoReconnect: true,
		StartInterval: 2,
		MaxError:      5,
	})

	if err := rdb.Init(); err != nil {
		logger.Error(err)
		return
	}

	err := rdb.Set(context.Background(), "test", "test", 10*time.Minute)
	if err != nil {
		logger.Error(err)
		return
	}

	val, err := rdb.Get(context.Background(), "test")
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Info(val)
}
