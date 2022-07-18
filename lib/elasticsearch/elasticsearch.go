package elasticsearch

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	database "github.com/fajarardiyanto/flt-go-database/interfaces"
	"github.com/fajarardiyanto/flt-go-utils/hash"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ElasticSearch struct {
	tag         string
	id          string
	reconnectAt int
	lastTimeout time.Duration
	config      database.ElasticSearchProviderConfig
	elastic     *elasticsearch.Client
	sync.RWMutex
}

func (c *ElasticSearch) Elastic() *elasticsearch.Client {
	return c.elastic
}

func NewElasticSearch(tag string, config database.ElasticSearchProviderConfig) database.ElasticSearch {
	id := hash.CreateSmallHash(10, config.Host,
		strconv.Itoa(config.Port),
		config.Username,
		config.Password)

	es := ElasticSearch{tag: strings.ToLower(tag), config: config, id: id}

	return &es
}

func (c *ElasticSearch) OnElasticSearchError(e error) (err error) {
	if c.config.MaxError == 0 {
		c.config.MaxError = 5
	}

	if c.reconnectAt >= c.config.MaxError {
		return e
	}

	c.reconnectAt++
	if !c.config.AutoReconnect {
		return e
	}

	var ttm = 2
	if c.lastTimeout.Seconds() == 0 {
		ttm = 1
		c.lastTimeout = time.Duration(c.config.StartInterval) * time.Second
	}

	c.lastTimeout = time.Duration(int(c.lastTimeout.Seconds())*ttm) * time.Second

	time.Sleep(c.lastTimeout)

	return c.ElasticSearch()
}

func (c *ElasticSearch) ElasticSearch() (err error) {
	if !c.config.Enable {
		return fmt.Errorf("aborted, elasticsearch not enable in config, double check configuration again")
	}

	url := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}

	conn, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return c.OnElasticSearchError(fmt.Errorf("[%s] %s", c.id, err.Error()))
	}

	es, errEs := conn.Info()
	if errEs != nil {
		return c.OnElasticSearchError(fmt.Errorf("[%s] %s", c.id, errEs.Error()))
	}
	defer func() {
		err = es.Body.Close()
		return
	}()

	c.elastic = conn

	return nil
}
