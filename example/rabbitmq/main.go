package main

import (
	"context"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-database/lib"
	loginterfaces "github.com/fajarardiyanto/flt-go-logger/interfaces"
	log "github.com/fajarardiyanto/flt-go-logger/lib"
	"github.com/fajarardiyanto/flt-go-utils/hash"
	"time"
)

type TestingCode struct {
	ID string
}

var logger loginterfaces.Logger

func main() {
	logger = log.NewLib()
	logger.Init("Testing database module")

	dbs := lib.NewLib()
	dbs.Init(logger)
	rabbit := dbs.LoadRabbitMQ("localtest", interfaces.RabbitMQProviderConfig{
		Enable:   true,
		Host:     "localhost",
		Port:     5672,
		Username: "guest",
		Password: "guest",
	})

	c := make(chan bool)
	//go func() {
	//	rabbit.Consumer(interfaces.RabbitMQOptions{
	//		Exchange: "MESSAGE_ENCODING_BASE64_GOB",
	//		NoWait:   true}, onMsg)
	//}()

	go func() {
		rabbit.Producer(interfaces.RabbitMQOptions{NoWait: true})
	}()

	go func() {
		tss := time.Now()
		time.Sleep(1 * time.Second)
		for range time.NewTicker(1 * time.Second).C {
			if err := rabbit.Push(context.Background(), "", "MESSAGE_ENCODING_BASE64_GOB",
				TestingCode{
					ID: hash.CreateRandomId(16),
				}, nil); err != nil {
				logger.Error(err)
			}

			if time.Since(tss).Seconds() >= 10 {
				logger.Quit()
			}
		}
	}()

	c <- true
}

func onMsg(s interfaces.Messages, cid interfaces.ConsumerCallbackIsDone) {
	var msg TestingCode
	if err := s.Decode(&msg); err == nil {
		logger.Debug("Message Receive %s - %s", s.Exchange(), msg)
	}
}
