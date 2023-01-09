package kafka

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	"github.com/fajarardiyanto/flt-go-utils/hash"
	"github.com/linkedin/goavro/v2"
	"github.com/riferrei/srclient"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc/metadata"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Consumer struct {
	callback database.ConsumerCallback
	options  database.KafkaOptions
	config   database.KafkaProviderConfig
	store    *Stores
	logger   logger.Logger
	limit    ratelimit.Limiter
	sync.RWMutex
}

func NewConsumer(lg logger.Logger, config database.KafkaProviderConfig,
	options database.KafkaOptions,
	callback database.ConsumerCallback,
	store *Stores) *Consumer {
	limiter := options.Limiter
	if limiter < 10 {
		limiter = 10
	}
	return &Consumer{
		store:    store,
		callback: callback,
		options:  options,
		config:   config,
		logger:   lg,
		limit:    ratelimit.New(limiter),
	}
}

func (c *Consumer) Run() {
	c.logger.Debug("Starting kafka consumer with topic %s, with group %s", c.options.Topic, c.options.Group)
	config := createConsumerInit(c.logger, c.config, c.options)
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		c.logger.Error(err).Quit()
	}

	var schema *srclient.Schema
	if len(c.config.Registry) != 0 && len(c.options.RegistryValue) != 0 {
		_, schema, err = getLastSchema(c.config, c.options)
		if err != nil {
			c.logger.Error(err).Quit()
		}
	}

	codec, err := goavro.NewCodec(schema.Schema())
	if err != nil {
		c.logger.Error("failed to get schema codec %s", err.Error()).Quit()
	}

	if err := consumer.SubscribeTopics([]string{c.options.Topic}, nil); err != nil {
		c.logger.Error(err).Quit()
	}

	run := true

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for run {
		select {
		case sig := <-sigchan:
			c.logger.Warning("Caught signal %s: terminating", sig.String())
			run = false
		default:
			ev := consumer.Poll(10)
			switch e := ev.(type) {
			case *kafka.Message:
				mdd := make(map[string]string)
				mdd["content-type"] = "application/rabbitmq"
				for _, s := range e.Headers {
					mdd[s.Key] = string(s.Value)
				}
				md := metadata.New(mdd)
				ctx := metadata.NewIncomingContext(context.Background(), md)

				var data []byte
				if schema != nil {
					dataRaw := e.Value[5:]
					//debug results
					os.MkdirAll("/tmp/", 0755)
					if err := os.WriteFile(fmt.Sprintf("/tmp/%s", hash.CreateRandomId(10)), dataRaw, 0755); err != nil {
						log.Println(err)
					}

					if codec == nil {
						err := fmt.Errorf("failed to get codec scheme, codec is nil")
						c.logger.Error(err)
					} else {
						native, _, err := codec.NativeFromBinary(dataRaw)
						if err == nil {
							value, err := codec.TextualFromNative(nil, native)
							if err == nil {
								data = value
							} else {
								c.logger.Error(err)
							}
						} else {
							c.logger.Error(err)
						}
					}

				} else {
					data = e.Value
				}

				switch c.options.Encoding {
				case database.EncodingBase64Gob:
					if data, err := base64.StdEncoding.DecodeString(string(data)); err == nil {
						msg := database.NewEncoder(bytes.NewBuffer(data),
							c.options.Topic, c.options.Group,
							database.EncodingBase64Gob)

						msg.SetContext(ctx)

						if c.callback != nil {
							if c.options.MultipleThread {
								c.limit.Take()

								go c.callback(msg, database.ConsumerCallbackIsDone{
									EndRequest: func() {
									},
								})
							} else {
								c.callback(msg, database.ConsumerCallbackIsDone{
									EndRequest: func() {
									},
								})
							}

						}
					} else {
						c.logger.Error(err)
					}
				default:
					msg := database.NewEncoder(bytes.NewBuffer(data),
						c.options.Topic, c.options.Group,
						c.options.Encoding)

					msg.SetContext(ctx)

					if c.callback != nil {
						if c.options.MultipleThread {
							c.limit.Take()

							go c.callback(msg, database.ConsumerCallbackIsDone{
								EndRequest: func() {
								},
							})
						} else {
							c.callback(msg, database.ConsumerCallbackIsDone{
								EndRequest: func() {
								},
							})
						}

					}
				}
			case kafka.Error:
				c.logger.Error(e.Error())
				run = false

			}
		}
	}

	c.logger.Debug("Closing consumer")
	consumer.Close()
	os.Exit(0)

}
