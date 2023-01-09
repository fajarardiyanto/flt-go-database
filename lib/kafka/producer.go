package kafka

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	"github.com/fajarardiyanto/flt-go-utils/hash"
	databaseproto "github.com/fajarardiyanto/module-proto/go/modules/database"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"sync"
	"time"
)

type MsgSend struct {
	ID      string
	Data    []byte
	Name    string
	Headers map[string]interface{}
}

type Producer struct {
	config  database.KafkaProviderConfig
	logger  logger.Logger
	store   *Stores
	pending chan MsgSend
	p       *kafka.Producer
	sync.RWMutex
}

func NewProducer(lg logger.Logger,
	config database.KafkaProviderConfig, store *Stores) *Producer {

	pr := &Producer{
		config:  config,
		logger:  lg,
		store:   store,
		pending: make(chan MsgSend, 1),
	}

	return pr
}

func (c *Producer) Run(isReady database.ProducerIsReady) (err error) {
	c.logger.Debug("Starting kafka producer")
	config := createProducerInit(c.logger, c.config)
	c.p, err = kafka.NewProducer(config)
	if err != nil {
		c.logger.Error(err).Quit()
	}

	go func() {
		time.Sleep(1 * time.Second)
		c.logger.Debug("Starting kafka producer is ready to use")
		isReady()
	}()

	for e := range c.p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				c.logger.Error("Delivery failed: %v", m.TopicPartition.Error)
			} else {
				c.logger.Debug("Delivered message to topic %s [%d] at offset %v",
					*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
			}
		case kafka.Error:
			c.logger.Error("Error: %v", ev)
		default:
			c.logger.Debug("Ignored event: %s", ev)
		}
	}

	return nil

}

func (c *Producer) SendingData(id string, options database.KafkaOptions, body interface{}, headers map[string]interface{}, cb database.ConsumerCallback) {

	var data []byte

	if len(id) == 0 {
		id = hash.CreateRandomId(10)
	}

	if options.Encoding == database.EncodingBase64Gob {
		bt := bytes.NewBuffer(nil)
		defer bt.Reset()
		if err := gob.NewEncoder(bt).Encode(body); err == nil {
			data = []byte(base64.StdEncoding.EncodeToString(bt.Bytes()))
		} else {
			c.logger.Error("Failed to encode gob %s", err)
		}
	} else if options.Encoding == database.EncodingProto {
		sendData := &databaseproto.SendData{
			ID: id,
		}
		if val, ok := body.(proto.Message); ok {
			if any, err := anypb.New(val); err == nil {
				sendData.Data = any
			}
		}

		if da, err := proto.Marshal(sendData); err == nil {
			data = da
		}

	} else if options.Encoding == database.EncodingJSON {
		data, _ = json.Marshal(body)
	} else if options.Encoding == database.EncodingNone {
		if val, ok := body.(string); ok {
			data = []byte(val)
		} else if val, ok := body.([]byte); ok {
			data = val
		}
	}

	if len(data) != 0 {
		c.RLock()
		producer := c.p
		c.RUnlock()

		c.logger.Debug("SENDING DATA => %s", string(data))

		if producer != nil {
			var kheaders []kafka.Header
			for k, v := range headers {
				if val, ok := v.(string); ok {
					kheaders = append(kheaders, kafka.Header{
						Key:   k,
						Value: []byte(val),
					})
				}
			}

			if err := producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &options.Topic, Partition: kafka.PartitionAny},
				Value:          data,
				Headers:        kheaders,
			}, nil); err != nil {
				c.logger.Error("Failed to produce data %s", err)
			}
		}
	}

}
