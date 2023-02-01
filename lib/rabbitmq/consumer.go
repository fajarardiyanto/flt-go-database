package rabbitmq

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	"github.com/fajarardiyanto/flt-go-utils/hash"
	"google.golang.org/grpc/metadata"
	"sync"
	"time"
)

type Consumer struct {
	callback    database.ConsumerCallback
	options     database.RabbitMQOptions
	config      database.RabbitMQProviderConfig
	store       *Stores
	logger      logger.Logger
	dialer      *Dialer
	alreadySubs bool
	sync.RWMutex
}

func NewConsumer(lg logger.Logger, dialer *Dialer, config database.RabbitMQProviderConfig,
	options database.RabbitMQOptions,
	callback database.ConsumerCallback,
	store *Stores) *Consumer {
	return &Consumer{
		dialer:   dialer,
		store:    store,
		callback: callback,
		options:  options,
		config:   config,
		logger:   lg,
	}
}

func (c *Consumer) Init() {
	if c.config.DedicatedConnection {
		ctx := context.Background()
		dialer := NewDialer(c.options.Exchange, c.logger)
		dialer.Dial(ctx, c.config, c.onError, nil)
		c.Subscribe(dialer.Session, c.Write())
	} else {

		c.RLock()
		alreadysub := c.alreadySubs
		c.RUnlock()

		if !alreadysub {
			if !c.dialer.connected {
				return
			}
			c.Subscribe(c.dialer.Session, c.Write())
		}
	}
}

func (c *Consumer) onError(err error) {
	if err != nil {
		c.logger.Trace("[%s] %s", c.options.Exchange, err)
	}

	if c.config.DedicatedConnection {
		reconnectDuration := time.Duration(c.config.ReconnectDuration) * time.Second
		c.logger.Warning("[%s] reconnecting in %s", c.options.Exchange, reconnectDuration)
		time.Sleep(reconnectDuration)
		c.logger.Info("[%s] reconnecting now ...", c.options.Exchange)
	}

	c.Lock()
	c.alreadySubs = false
	c.Unlock()

	c.Init()

}

func (c *Consumer) OnSession(queue string, sub Session, messages chan<- Message) {
	c.Lock()
	c.alreadySubs = true
	c.Unlock()

	if _, err := sub.QueueDeclare(
		queue,
		c.options.Durable,
		c.options.AutoDeleted,
		false,
		c.options.NoWait, nil); err != nil {
		c.onError(fmt.Errorf("cannot QueueDeclare from exclusive queue: %q, %v",
			queue, err))
		return
	}

	if len(c.options.RoutingKey) != 0 {
		if err := sub.QueueBind(queue, c.options.RoutingKey, c.options.Exchange, false, nil); err != nil {
			c.onError(fmt.Errorf("cannot QueueBind without a binding to exchange: %q, %v", c.options.Exchange, err))
			return
		}
	}

	deliveries, err := sub.Consume(queue, "", false, false, false, c.options.NoWait, nil)
	if err != nil {
		c.onError(fmt.Errorf("cannot Consume from: %q, %v", queue, err))
		return
	}

	if len(c.options.RoutingKey) != 0 {
		c.logger.Success("Subscribed exchange %s, routing %s", c.options.Exchange, c.options.RoutingKey)
	} else {
		c.logger.Success("Subscribed exchange %s", c.options.Exchange)
	}

	for msg := range deliveries {
		messages <- Message{
			Headers: msg.Headers,
			Body:    msg.Body,
		}

		if err := sub.Ack(msg.DeliveryTag, false); err != nil {
			c.logger.Error("Can't confirm ack delivery %s", err)
		}
	}
}

func (c *Consumer) Subscribe(sessions chan chan Session, messages chan<- Message) {

	queue := hash.CreateSmallHash(10, c.options.Exchange, c.options.RoutingKey)
	if len(c.options.RoutingKey) == 0 {
		queue = c.options.Exchange
	}

	for session := range sessions {
		sub := <-session
		c.OnSession(queue, sub, messages)
	}
}

func (c *Consumer) Write() chan<- Message {
	lines := make(chan Message)
	go func() {
		for line := range lines {
			mdd := make(map[string]string)
			mdd["content-type"] = "application/rabbitmq"
			for kk, ss := range line.Headers {
				if vals, ok := ss.(string); ok {
					mdd[kk] = vals
				}
			}

			md := metadata.New(mdd)
			ctx := metadata.NewIncomingContext(context.Background(), md)

			if c.callback == nil {

			}

			if c.options.Encoding == database.EncodingBase64Gob {
				if data, err := base64.StdEncoding.DecodeString(string(line.Body)); err == nil {
					msg := database.NewEncoder(bytes.NewBuffer(data),
						c.options.Exchange, c.options.RoutingKey,
						database.EncodingBase64Gob)
					msg.SetContext(ctx)

					if c.callback != nil {
						go c.callback(msg, database.ConsumerCallbackIsDone{
							EndRequest: func() {
							},
						})
					}
				} else {
					c.logger.Error(err)
				}
			}

		}
	}()
	return lines
}
