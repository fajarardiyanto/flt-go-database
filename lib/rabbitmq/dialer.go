package rabbitmq

import (
	"context"
	"fmt"
	"github.com/fajarardiyanto/flt-go-database/interfaces"
	logger "github.com/fajarardiyanto/flt-go-logger/interfaces"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type Message struct {
	Headers map[string]interface{}
	Body    []byte
}

type Dialer struct {
	id        string
	logger    logger.Logger
	Session   chan chan Session
	connected bool
	sync.RWMutex
}

func NewDialer(id string, lg logger.Logger) *Dialer {
	return &Dialer{id: id, logger: lg}
}

type Session struct {
	*amqp.Connection
	*amqp.Channel
}

// Close tears the connection down, taking the channel with it.
func (s *Session) Close() error {
	if s.Connection == nil {
		return nil
	}
	return s.Connection.Close()
}

func (c *Dialer) IsConnected() bool {
	c.RLock()
	connected := c.connected
	c.RUnlock()
	return connected
}

type OnError func(error)
type OnConnected func()

func (c *Dialer) Dial(cx context.Context, config interfaces.RabbitMQProviderConfig, onError OnError, onConnected OnConnected) {
	ctx, cancel := context.WithCancel(cx)

	url := fmt.Sprintf("amqp://%s:%s@%s:%d", config.Username,
		config.Password, config.Host, config.Port)

	conn, err := amqp.Dial(url)
	if err != nil {
		c.logger.Error("[%s] %s", c.id, err)
		cancel()
		onError(nil)
		return
	}

	onclose := make(chan *amqp.Error)
	conn.NotifyClose(onclose)

	c.Lock()
	c.connected = true
	c.Session = make(chan chan Session)
	c.Unlock()

	c.logger.Success("[%s] connected to : %s:%d", c.id, config.Host, config.Port)

	go func() {
		sess := make(chan Session)
		defer close(c.Session)

		for {

			select {
			case <-onclose:
				c.logger.Debug("[%s] connection closed", c.id)
				c.Lock()
				c.connected = false
				c.Unlock()
				go onError(nil)
				return
			case c.Session <- sess:
			case <-ctx.Done():
				c.logger.Debug("[%s] shutting down session factory", c.id)
				c.Lock()
				c.connected = false
				c.Unlock()
				go onError(nil)
				return
			}

			if c.connected && conn != nil {
				ch, err := conn.Channel()
				if err != nil {
					c.logger.Error("[%s] %s", c.id, err)
					cancel()
				} else {
					select {
					case sess <- Session{conn, ch}:
					case <-ctx.Done():
						c.logger.Debug("[%s] shutting down new session", c.id)
						return
					}
				}
			}

		}
	}()

	if onConnected != nil {
		go onConnected()
	}

}
