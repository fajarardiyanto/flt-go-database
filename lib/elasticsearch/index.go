package elasticsearch

import (
	"fmt"
	"strings"
)

func (c *ElasticSearch) CreateIndex(name string, mapping string) error {
	res, err := c.elastic.Indices.Create(name, c.elastic.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		c.log.Error(err)
		return err
	}
	if res.IsError() {
		c.log.Error(res)
		return fmt.Errorf("error: %s", res)
	}
	return nil
}
