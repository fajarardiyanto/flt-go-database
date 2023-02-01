package rabbitmq

import (
	"context"
	"fmt"
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	gutils "github.com/fajarardiyanto/flt-go-utils/grpc"
	"github.com/fajarardiyanto/flt-go-utils/hash"
	"strconv"
	"strings"
	"sync"
	"time"
)

var storesCallback *Stores

func init() {
	storesCallback = NewStore()
}

type RabbitMQ struct {
	tag      string
	id       string
	log      logger.Logger
	config   database.RabbitMQProviderConfig
	producer *Producer
	consumer map[string]*Consumer
	dialer   *Dialer
	sync.RWMutex
}

func NewRabbitMQ(tag string, lo logger.Logger, config database.RabbitMQProviderConfig) database.RabbitMQ {
	id := hash.CreateSmallHash(10, config.Host,
		strconv.Itoa(config.Port),
		config.Username,
		config.Password)

	if vals, ok := storesCallback.LoadClient(id); ok {
		return vals
	}

	msq := &RabbitMQ{tag: strings.ToLower(tag), log: lo, config: config, id: id,
		consumer: map[string]*Consumer{}}

	if !config.DedicatedConnection && config.Enable {
		msq.dialer = NewDialer(msq.tag, lo)
		msq.dial()
	}

	storesCallback.StoreClient(msq)
	lo.Debug("RabbitMQ Client %s:%d has been registered", config.Host, config.Port)
	return msq
}

func LoadRabbitMQ(tag string) (database.RabbitMQ, bool) {
	return storesCallback.LoadClientByTag(tag)
}

func (c *RabbitMQ) dial() {
	c.dialer.Dial(context.Background(), c.config, c.onDialError, c.onConnected)
}

func (c *RabbitMQ) onConnected() {
	if !c.config.DedicatedConnection {
		time.Sleep(1 * time.Second)
		c.RLock()
		consm := c.consumer
		c.RUnlock()
		for _, val := range consm {
			go val.Init()
		}
	}

	if c.producer != nil {
		c.producer.Init()
	}
}

func (c *RabbitMQ) onDialError(err error) {
	if c.config.ReconnectDuration <= 0 {
		c.config.ReconnectDuration = 5
	}

	if !c.config.DedicatedConnection {
		if err != nil {
			c.log.Trace(err)
		}

		c.RLock()
		consm := c.consumer
		c.RUnlock()
		for _, val := range consm {
			val.onError(nil)
		}

		reconnectDuration := time.Duration(c.config.ReconnectDuration) * time.Second
		c.log.Warning("reconnecting in %s", reconnectDuration)
		time.Sleep(reconnectDuration)
		c.log.Info("reconnecting now ...")
		c.dial()
	}
}

func (c *RabbitMQ) Producer(options database.RabbitMQOptions) {
	if !c.config.Enable {
		c.log.Error("RabbitMQ is disabled").Quit()
	}

	if len(options.ExchangeType) == 0 {
		options.ExchangeType = "direct"
	}

	c.Lock()
	c.producer = NewProducer(c.log, c.dialer,
		c.config, options, storesCallback)
	c.Unlock()

	go c.producer.Init()
}

func (c *RabbitMQ) Consumer(options database.RabbitMQOptions, callback database.ConsumerCallback) {
	if !c.config.Enable {
		c.log.Error("RabbitMQ is disabled").Quit()
	}

	if len(options.Exchange) == 0 {
		c.log.Error("Exchange is required").Quit()
	}

	if len(options.ExchangeType) == 0 {
		options.ExchangeType = "direct"
	}

	consumer := NewConsumer(c.log, c.dialer, c.config, options, callback, storesCallback)
	c.Lock()
	c.consumer[options.Exchange] = consumer
	c.Unlock()

	go consumer.Init()

}

func (c *RabbitMQ) Push(ctx context.Context, id, key string, body interface{}, cb database.ConsumerCallback) error {
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
			producer.SendingData(id, key, body, headers, nil)
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

		producer.SendingData(id, key, body, headers, func(s database.Messages,
			cid database.ConsumerCallbackIsDone) {
			doneCtx = cid
			cb(s, done)
		})

		<-ctx.Done()

		return nil

	}

	return fmt.Errorf("rabbitmq not ready")
}
