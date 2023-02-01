package rabbitmq

import (
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	"sync"
)

type Stores struct {
	clienttags map[string]string
	client     map[string]*RabbitMQ
	items      map[string]database.ConsumerCallback
	sync.RWMutex
}

func NewStore() *Stores {
	return &Stores{
		clienttags: make(map[string]string),
		client:     make(map[string]*RabbitMQ),
		items:      make(map[string]database.ConsumerCallback),
	}
}

func (c *Stores) StoreClient(client *RabbitMQ) {
	c.Lock()
	c.client[client.id] = client
	c.clienttags[client.tag] = client.id
	c.Unlock()
}

func (c *Stores) LoadClient(id string) (client *RabbitMQ, ok bool) {
	c.RLock()
	client, ok = c.client[id]
	c.RUnlock()
	return client, ok
}

func (c *Stores) LoadClientByTag(ta string) (client *RabbitMQ, ok bool) {
	c.RLock()
	tag, ok := c.clienttags[ta]
	c.RUnlock()
	if ok {
		return c.LoadClient(tag)
	}
	return client, ok
}

func (c *Stores) Put(id string, cb database.ConsumerCallback) {
	c.Lock()
	c.items[id] = cb
	c.Unlock()
}

func (c *Stores) Get(id string) (cb database.ConsumerCallback, ok bool) {
	c.RLock()
	cb, ok = c.items[id]
	c.RUnlock()
	return cb, ok
}

func (c *Stores) Delete(id string) {
	c.Lock()
	delete(c.items, id)
	c.Unlock()
}
