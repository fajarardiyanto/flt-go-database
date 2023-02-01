package rabbitmq

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	"github.com/fajarardiyanto/flt-go-utils/hash"
	databaseproto "github.com/fajarardiyanto/module-proto/go/modules/database"
	amqp "github.com/rabbitmq/amqp091-go"
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

type PendingQue struct {
	id     string
	key    string
	body   interface{}
	header map[string]interface{}
	cb     database.ConsumerCallback
}

type Producer struct {
	options     database.RabbitMQOptions
	config      database.RabbitMQProviderConfig
	logger      logger.Logger
	dialer      *Dialer
	store       *Stores
	pending     chan MsgSend
	pendingQue  []PendingQue
	alreadySubs bool
	sync.RWMutex
}

func NewProducer(lg logger.Logger, dialer *Dialer,
	config database.RabbitMQProviderConfig,
	options database.RabbitMQOptions, store *Stores) *Producer {

	pr := &Producer{
		dialer:  dialer,
		options: options,
		config:  config,
		logger:  lg,
		store:   store,
	}
	return pr
}

func (c *Producer) Init() {
	c.RLock()
	alreadySub := c.alreadySubs
	dialer := c.dialer
	c.RUnlock()

	if !alreadySub && dialer.IsConnected() {
		c.Publish(c.dialer.Session)
	}

}

func (c *Producer) Publish(sessions chan chan Session) {
	if len(c.pendingQue) != 0 {
		go func() {
			time.Sleep(1 * time.Second)
			c.RLock()
			pQue := c.pendingQue
			c.RUnlock()
			for _, s := range pQue {
				c.SendingData(s.id, s.key, s.body, s.header, s.cb)
			}

			c.Lock()
			c.pendingQue = []PendingQue{}
			c.Unlock()

		}()
	}

	for session := range sessions {
		c.Lock()
		c.alreadySubs = true
		confirm := make(chan amqp.Confirmation)
		c.pending = make(chan MsgSend, 1)
		pub := <-session
		c.Unlock()

		c.RLock()
		pending := c.pending
		dialer := c.dialer
		c.RUnlock()

		if !dialer.IsConnected() {
			c.Lock()
			c.alreadySubs = false
			c.Unlock()
			return
		}

		if pub.Channel.IsClosed() {
			c.Lock()
			c.alreadySubs = false
			c.Unlock()
			return
		}

		if err := pub.Confirm(false); err != nil {
			c.logger.Warning("publisher confirms not supported")
			close(confirm)
		} else {
			pub.NotifyPublish(confirm)
		}

		for {

			if pub.Channel.IsClosed() {
				c.Lock()
				c.alreadySubs = false
				c.Unlock()
				return
			}

			select {
			case msg := <-pending:
				if pub.Channel.IsClosed() {
					c.Lock()
					c.alreadySubs = false
					c.Unlock()
					return
				}

				if qq, err := pub.Channel.QueueDeclare(msg.Name,
					false, false, false, true, nil); err != nil {
					c.logger.Error("QueueDeclare: %s", err)
				} else {
					if err = pub.Channel.PublishWithContext(context.Background(), "", qq.Name, false, false, amqp.Publishing{
						Headers: msg.Headers,
						Body:    msg.Data,
					}); err != nil {
						c.logger.Error("Publish: %s", err)
					} else {
						c.logger.Trace("Message sending que (%s) ID : %s",
							msg.Name, msg.ID)
					}
				}

			case confirmed := <-confirm:
				if pub.Channel.IsClosed() {
					c.Lock()
					c.alreadySubs = false
					c.Unlock()
					return
				}

				if !confirmed.Ack {
					c.logger.Warning("Failed delivery of delivery tag: %d", confirmed.DeliveryTag)
				}
			}

		}
	}
}

func (c *Producer) SendingData(id string, key string, body interface{}, headers map[string]interface{}, cb database.ConsumerCallback) {

	c.RLock()
	dialer := c.dialer
	c.RUnlock()

	if !dialer.IsConnected() {
		c.logger.Warning("Skip, not connected to rabbitmq server, add to sending que")
		c.Lock()
		c.pendingQue = append(c.pendingQue, PendingQue{
			id:     id,
			key:    key,
			body:   body,
			header: headers,
			cb:     cb,
		})
		c.Unlock()
		return
	}

	if len(key) == 0 {
		key = c.options.Exchange
	}

	if len(key) != 0 {
		var data []byte

		if len(id) == 0 {
			id = hash.CreateRandomId(10)
		}

		if c.options.Encoding == database.EncodingBase64Gob {
			bt := bytes.NewBuffer(nil)
			defer bt.Reset()
			if err := gob.NewEncoder(bt).Encode(body); err == nil {
				data = []byte(base64.StdEncoding.EncodeToString(bt.Bytes()))
			} else {
				c.logger.Error("Failed to encode gob %s", err)
			}
		}

		if c.options.Encoding == database.EncodingProto {
			sendData := &databaseproto.SendData{
				ID: id,
			}
			if val, ok := body.(proto.Message); ok {
				if v, err := anypb.New(val); err == nil {
					sendData.Data = v
				}
			}

			if da, err := proto.Marshal(sendData); err == nil {
				data = da
			}

		}

		if cb != nil {
			if c.store != nil {
				c.store.Put(id, cb)
			}
		}

		c.Lock()
		c.pending <- MsgSend{ID: id, Name: key, Data: data, Headers: headers}
		c.Unlock()
	}

}
