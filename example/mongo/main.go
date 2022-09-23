package main

import (
	"context"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib"
	logInterface "github.com/fajarardiyanto/flt-go-logger/interfaces"
	log "github.com/fajarardiyanto/flt-go-logger/lib"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func main() {
	logger := log.NewLib()
	logger.Init("Modules Mongo Database")

	db := lib.NewLib()
	db.Init(logger)

	mongoDB := db.LoadMongoDatabase(interfaces.MongoProviderConfig{
		Enable:            true,
		Host:              "0.0.0.0",
		Port:              27017,
		Username:          "admin",
		Password:          "secret",
		AutoReconnect:     true,
		TimeoutConnection: 3000,
	})

	if err := mongoDB.Init(); err != nil {
		logger.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res := make(chan []bson.M)
	go mongoDB.LoadPostChannel(ctx, "afaik", "users", bson.M{}, res)
	result := <-res
	logger.Info(result)
}

func Manual(logger logInterface.Logger, mongoDB interfaces.Mongo) {
	mdb := mongoDB.SetDatabase("afaik")

	filter := make(bson.M)
	err := mdb.Collection("user").FindOne(context.Background(), bson.M{}).Decode(filter)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Info("result %v", filter)
}
