package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	gutils "github.com/fajarardiyanto/flt-go-utils/grpc"
	"github.com/fajarardiyanto/flt-go-utils/hash"
	"github.com/google/uuid"
	"github.com/riferrei/srclient"
	"strings"
	"sync"
)

var storesCallback *Stores

func init() {
	storesCallback = NewStore()
}

type Kafka struct {
	tag      string
	id       string
	log      logger.Logger
	config   database.KafkaProviderConfig
	consumer map[string]*Consumer
	producer *Producer
	sync.RWMutex
}

func init() {

}

func NewKafka(tag string, lo logger.Logger, config database.KafkaProviderConfig) database.Kafka {
	id := hash.CreateSmallHash(10, config.Host, config.Username, config.Password)

	if vals, ok := storesCallback.LoadClient(id); ok {
		return vals
	}

	lo.Debug("Kafka Client %s has been registered", config.Host)

	msq := &Kafka{tag: strings.ToLower(tag), log: lo, config: config, id: id,
		consumer: map[string]*Consumer{}}

	storesCallback.StoreClient(msq)

	return msq
}

func (c *Kafka) Consumer(options database.KafkaOptions, callback database.ConsumerCallback) {
	if !c.config.Enable {
		c.log.Error("Kafka is disabled").Quit()
	}

	if len(options.Topic) == 0 {
		c.log.Error("Topic is required").Quit()
	}

	consumer := NewConsumer(c.log, c.config, options,
		callback, storesCallback)

	c.Lock()
	c.consumer[options.Topic] = consumer
	c.Unlock()

	go consumer.Run()

}

func (c *Kafka) Producer(isReady database.ProducerIsReady) {
	if !c.config.Enable {
		c.log.Error("Kafka is disabled").Quit()
	}

	c.Lock()
	c.producer = NewProducer(c.log, c.config, storesCallback)
	c.Unlock()
	go c.producer.Run(isReady)

}

func (c *Kafka) Push(ctx context.Context, id string, options database.KafkaOptions, body interface{}, cb database.ConsumerCallback) error {
	c.RLock()
	producer := c.producer
	c.RUnlock()

	headers := make(map[string]interface{})

	if ctx != nil {
		meta := gutils.ExtractOutgoing(ctx)
		headers["trace-id"] = meta.Get("trace-id")
		headers["uber-trace-id"] = meta.Get("uber-trace-id")
	}

	if producer != nil {

		if cb == nil {
			producer.SendingData(id, options, body, headers, nil)
			return nil
		}

		ctx, cancel := context.WithCancel(ctx)
		var doneCtx database.ConsumerCallbackIsDone
		defer func() {
			cancel()
			if doneCtx.EndRequest != nil {
				doneCtx.EndRequest()
			}
		}()

		done := database.ConsumerCallbackIsDone{
			Done: cancel,
			EndRequest: func() {
				cancel()
				if doneCtx.EndRequest != nil {
					doneCtx.EndRequest()
				}
			},
		}

		producer.SendingData(id, options, body, headers, func(s database.Messages,
			ccid database.ConsumerCallbackIsDone) {
			doneCtx = ccid
			cb(s, done)
		})

		<-ctx.Done()

		return nil

	}

	return fmt.Errorf("kafka not ready")
}

func createConsumerInit(lo logger.Logger, cfg database.KafkaProviderConfig, options database.KafkaOptions) (config *kafka.ConfigMap) {
	chanLogs := make(chan kafka.LogEvent)
	var tt = "consumer"
	go func() {
		for {
			logEv := <-chanLogs
			switch logEv.Level {
			case 1, 2, 3:
				lo.Error("[%s][%d][%s][%s]%s", tt, logEv.Level, options.Group, options.Topic, logEv.Message)
			case 7:
				lo.Trace("[%s][%d][%s][%s]%s", tt, logEv.Level, options.Group, options.Topic, logEv.Message)
			default:
				lo.Debug("[%s][%d][%s][%s]%s", tt, logEv.Level, options.Group, options.Topic, logEv.Message)
			}
		}
	}()

	config = &kafka.ConfigMap{
		"bootstrap.servers":        cfg.Host,
		"log_level":                7,
		"log.connection.close":     true,
		"go.logs.channel.enable":   true,
		"auto.offset.reset":        "earliest",
		"session.timeout.ms":       120000,
		"heartbeat.interval.ms":    40000,
		"fetch.min.bytes":          1000000,
		"broker.address.family":    "v4",
		"enable.auto.offset.store": true,
		"client.id":                uuid.New().String(),
	}

	if len(cfg.Debug) == 0 {
		cfg.Debug = "consumer,cgrp,topic,fetch"
	}

	config.SetKey("debug", cfg.Debug)

	config.SetKey("go.logs.channel", chanLogs)

	if len(options.Group) != 0 {
		config.SetKey("group.id", options.Group)
	}

	if len(cfg.SecurityProtocol) == 0 {
		cfg.SecurityProtocol = "SASL_SSL"
	}

	if len(cfg.Mechanisms) == 0 {
		cfg.Mechanisms = "PLAIN"
	}

	if len(cfg.Username) != 0 && len(cfg.Password) != 0 {
		config.SetKey("security.protocol", cfg.SecurityProtocol)
		config.SetKey("sasl.mechanisms", cfg.Mechanisms)
		config.SetKey("sasl.username", cfg.Username)
		config.SetKey("sasl.password", cfg.Password)
	}

	return config
}

func createProducerInit(lo logger.Logger, cfg database.KafkaProviderConfig) (config *kafka.ConfigMap) {
	chanLogs := make(chan kafka.LogEvent)
	var tt = "producer"
	go func() {
		for {
			logEv := <-chanLogs
			switch logEv.Level {
			case 1, 2, 3:
				lo.Error("[%s][%d] %s", tt, logEv.Level, logEv.Message)
			case 7:
				lo.Trace("[%s][%d] %s", tt, logEv.Level, logEv.Message)
			default:
				lo.Debug("[%s][%d] %s", tt, logEv.Level, logEv.Message)
			}
		}
	}()

	config = &kafka.ConfigMap{
		"bootstrap.servers":      cfg.Host,
		"log_level":              7,
		"log.connection.close":   true,
		"go.logs.channel.enable": true,
	}

	if len(cfg.Debug) == 0 {
		cfg.Debug = "consumer"
	}

	config.SetKey("debug", cfg.Debug)

	config.SetKey("go.logs.channel", chanLogs)

	if len(cfg.SecurityProtocol) == 0 {
		cfg.SecurityProtocol = "SASL_SSL"
	}

	if len(cfg.Mechanisms) == 0 {
		cfg.Mechanisms = "PLAIN"
	}

	if len(cfg.Username) != 0 && len(cfg.Password) != 0 {
		config.SetKey("security.protocol", cfg.SecurityProtocol)
		config.SetKey("sasl.mechanisms", cfg.Mechanisms)
		config.SetKey("sasl.username", cfg.Username)
		config.SetKey("sasl.password", cfg.Password)
	}

	return config
}

func getLastSchema(cfg database.KafkaProviderConfig, options database.KafkaOptions) (*srclient.SchemaRegistryClient, *srclient.Schema, error) {
	schemaRegistryClient := srclient.CreateSchemaRegistryClient(cfg.Registry)
	if len(cfg.Username) != 0 && len(cfg.Password) != 0 {
		schemaRegistryClient.SetCredentials(cfg.Username, cfg.Password)
	}

	if options.SchemeVersion != 0 {
		schema, errSchema := schemaRegistryClient.GetSchemaByVersion(options.RegistryValue, options.SchemeVersion)
		if errSchema != nil {
			return nil, nil, errSchema
		}
		return schemaRegistryClient, schema, nil
	}

	if options.SchemeID != 0 {
		schema, errSchema := schemaRegistryClient.GetSchema(options.SchemeID)
		if errSchema != nil {
			return nil, nil, errSchema
		}
		return schemaRegistryClient, schema, nil
	}

	schema, errSchema := schemaRegistryClient.GetLatestSchema(options.RegistryValue)
	if errSchema != nil {
		return nil, nil, errSchema
	}

	return schemaRegistryClient, schema, nil
}
